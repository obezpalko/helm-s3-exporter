package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/obezpalko/helm-repo-exporter/internal/analyzer"
	"github.com/obezpalko/helm-repo-exporter/internal/fetcher"
	"github.com/obezpalko/helm-repo-exporter/internal/metrics"
	"github.com/obezpalko/helm-repo-exporter/internal/web"
	"github.com/obezpalko/helm-repo-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Helm Repository Exporter...")

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Repositories: %d", len(cfg.Repositories))
	for _, repo := range cfg.Repositories {
		log.Printf("    - %s: %s (interval: %v)", repo.Name, repo.URL, repo.ScanInterval)
		if repo.Auth != nil {
			switch {
			case repo.Auth.Basic != nil:
				log.Printf("      Auth: Basic (username: %s)", repo.Auth.Basic.Username)
			case repo.Auth.BearerToken != "":
				log.Printf("      Auth: Bearer Token")
			case len(repo.Auth.Headers) > 0:
				log.Printf("      Auth: Custom Headers (%d)", len(repo.Auth.Headers))
			}
		}
	}
	log.Printf("  Default Scan Interval: %v", cfg.ScanInterval)
	log.Printf("  Scan Timeout: %v", cfg.ScanTimeout)
	log.Printf("  Metrics Port: %s", cfg.MetricsPort)
	log.Printf("  Enable HTML: %v", cfg.EnableHTML)

	// Create HTTP clients for each repository
	ctx := context.Background()
	type repoClient struct {
		client   *fetcher.Client
		interval time.Duration
	}
	var repoClients []repoClient
	for _, repo := range cfg.Repositories {
		client := fetcher.NewClient(repo, cfg.ScanTimeout)
		repoClients = append(repoClients, repoClient{
			client:   client,
			interval: repo.ScanInterval,
		})
	}
	log.Printf("Created %d HTTP client(s)", len(repoClients))

	// Initialize metrics
	metricsCollector := metrics.NewMetrics()
	log.Println("Prometheus metrics initialized")

	// Initialize HTML generator if enabled
	var htmlGenerator *web.HTMLGenerator
	if cfg.EnableHTML {
		htmlGenerator, err = web.NewHTMLGenerator()
		if err != nil {
			log.Fatalf("Failed to create HTML generator: %v", err)
		}
		log.Println("HTML generator initialized")
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.Handler())

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing health response: %v", err)
		}
	})

	// Add readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Ready")); err != nil {
			log.Printf("Error writing readiness response: %v", err)
		}
	})

	if cfg.EnableHTML && htmlGenerator != nil {
		mux.Handle(cfg.HTMLPath, htmlGenerator)
		log.Printf("HTML dashboard enabled at %s", cfg.HTMLPath)
	}

	server := &http.Server{
		Addr:              ":" + cfg.MetricsPort,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on :%s", cfg.MetricsPort)
		log.Printf("Metrics available at http://localhost:%s%s", cfg.MetricsPort, cfg.MetricsPath)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Perform initial scrape for all repositories
	var allClients []*fetcher.Client
	for _, rc := range repoClients {
		allClients = append(allClients, rc.client)
	}
	performScrape(ctx, allClients, metricsCollector, htmlGenerator)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create a channel for scrape triggers
	scrapeChan := make(chan *fetcher.Client, 100)

	// Start per-repository scraping goroutines
	for _, rc := range repoClients {
		go func(client *fetcher.Client, interval time.Duration) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					scrapeChan <- client
				case <-sigChan:
					return
				}
			}
		}(rc.client, rc.interval)
		log.Printf("Started scraper for %s with interval %v", rc.client.RepositoryName(), rc.interval)
	}

	// Main loop - handle scrapes and signals
	log.Println("Exporter started successfully")
	for {
		select {
		case client := <-scrapeChan:
			// Scrape single repository
			performSingleRepoScrape(ctx, client, metricsCollector, htmlGenerator)
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down...", sig)
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Printf("Error during shutdown: %v", err)
			}
			log.Println("Exporter stopped")
			return
		}
	}
}

// performScrape scrapes all repositories (used for initial scrape)
func performScrape(ctx context.Context, clients []*fetcher.Client, metricsCollector *metrics.Metrics, htmlGenerator *web.HTMLGenerator) {
	log.Println("Starting initial scrape...")
	overallStartTime := time.Now()

	// Aggregate analysis from all repositories for HTML dashboard
	var totalAnalysis *analyzer.ChartAnalysis

	for _, client := range clients {
		repoName := client.RepositoryName()
		log.Printf("  Fetching from repository: %s", repoName)
		startTime := time.Now()

		// Fetch index.yaml
		data, err := client.GetIndexYAML(ctx)
		if err != nil {
			log.Printf("  ERROR: Failed to fetch index.yaml from %s: %v", repoName, err)
			metricsCollector.RecordError(repoName)
			continue
		}

		// Parse index
		index, err := analyzer.ParseIndex(data)
		if err != nil {
			log.Printf("  ERROR: Failed to parse index.yaml from %s: %v", repoName, err)
			metricsCollector.RecordError(repoName)
			continue
		}

		// Analyze charts with repository name
		analysis := analyzer.AnalyzeChartsWithRepo(index, repoName)
		duration := time.Since(startTime)
		log.Printf("  Repository %s: %d charts, %d versions (in %v)", repoName, analysis.TotalCharts, analysis.TotalVersions, duration)

		// Update per-repository metrics
		metricsCollector.Update(repoName, analysis)
		metricsCollector.RecordSuccess(repoName)
		metricsCollector.ScrapeDuration.WithLabelValues(repoName).Observe(duration.Seconds())

		// Merge with total analysis for HTML dashboard
		if totalAnalysis == nil {
			totalAnalysis = analysis
		} else {
			totalAnalysis = mergeAnalysis(totalAnalysis, analysis)
		}
	}

	if totalAnalysis == nil {
		log.Println("ERROR: No successful scrapes")
		return
	}

	// Update HTML dashboard with aggregated data if enabled
	if htmlGenerator != nil {
		htmlGenerator.Update(totalAnalysis)
	}

	overallDuration := time.Since(overallStartTime)
	log.Printf("Initial scrape completed in %v", overallDuration)
	log.Printf("  Total charts: %d", totalAnalysis.TotalCharts)
	log.Printf("  Total versions: %d", totalAnalysis.TotalVersions)
	if !totalAnalysis.OldestChartDate.IsZero() {
		log.Printf("  Oldest chart: %s", totalAnalysis.OldestChartDate.Format("2006-01-02 15:04:05"))
	}
	if !totalAnalysis.NewestChartDate.IsZero() {
		log.Printf("  Newest chart: %s", totalAnalysis.NewestChartDate.Format("2006-01-02 15:04:05"))
	}
}

// performSingleRepoScrape scrapes a single repository and updates its metrics
func performSingleRepoScrape(ctx context.Context, client *fetcher.Client, metricsCollector *metrics.Metrics, htmlGenerator *web.HTMLGenerator) {
	repoName := client.RepositoryName()
	log.Printf("Scraping repository: %s", repoName)
	startTime := time.Now()

	// Fetch index.yaml
	data, err := client.GetIndexYAML(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to fetch index.yaml from %s: %v", repoName, err)
		metricsCollector.RecordError(repoName)
		return
	}

	// Parse index
	index, err := analyzer.ParseIndex(data)
	if err != nil {
		log.Printf("ERROR: Failed to parse index.yaml from %s: %v", repoName, err)
		metricsCollector.RecordError(repoName)
		return
	}

	// Analyze charts with repository name
	analysis := analyzer.AnalyzeChartsWithRepo(index, repoName)
	duration := time.Since(startTime)
	log.Printf("Repository %s scraped in %v: %d charts, %d versions", repoName, duration, analysis.TotalCharts, analysis.TotalVersions)

	// Update per-repository metrics
	metricsCollector.Update(repoName, analysis)
	metricsCollector.RecordSuccess(repoName)
	metricsCollector.ScrapeDuration.WithLabelValues(repoName).Observe(duration.Seconds())

	// Update HTML dashboard with this repo's data if enabled
	// Note: This will overwrite previous data in the HTML dashboard
	// For multiple repos, the dashboard will show the last scraped repo
	if htmlGenerator != nil {
		htmlGenerator.Update(analysis)
	}
}

// mergeAnalysis combines two chart analyses
func mergeAnalysis(a1, a2 *analyzer.ChartAnalysis) *analyzer.ChartAnalysis {
	merged := &analyzer.ChartAnalysis{
		TotalCharts:   a1.TotalCharts + a2.TotalCharts,
		TotalVersions: a1.TotalVersions + a2.TotalVersions,
		ChartsInfo:    append(a1.ChartsInfo, a2.ChartsInfo...),
	}

	// Merge date statistics
	switch {
	case !a1.OldestChartDate.IsZero() && !a2.OldestChartDate.IsZero():
		if a1.OldestChartDate.Before(a2.OldestChartDate) {
			merged.OldestChartDate = a1.OldestChartDate
		} else {
			merged.OldestChartDate = a2.OldestChartDate
		}
	case !a1.OldestChartDate.IsZero():
		merged.OldestChartDate = a1.OldestChartDate
	default:
		merged.OldestChartDate = a2.OldestChartDate
	}

	switch {
	case !a1.NewestChartDate.IsZero() && !a2.NewestChartDate.IsZero():
		if a1.NewestChartDate.After(a2.NewestChartDate) {
			merged.NewestChartDate = a1.NewestChartDate
		} else {
			merged.NewestChartDate = a2.NewestChartDate
		}
	case !a1.NewestChartDate.IsZero():
		merged.NewestChartDate = a1.NewestChartDate
	default:
		merged.NewestChartDate = a2.NewestChartDate
	}

	// For median, we'll just use the median of the two medians (approximation)
	switch {
	case !a1.MedianChartDate.IsZero() && !a2.MedianChartDate.IsZero():
		avg := (a1.MedianChartDate.Unix() + a2.MedianChartDate.Unix()) / 2
		merged.MedianChartDate = time.Unix(avg, 0)
	case !a1.MedianChartDate.IsZero():
		merged.MedianChartDate = a1.MedianChartDate
	default:
		merged.MedianChartDate = a2.MedianChartDate
	}

	return merged
}
