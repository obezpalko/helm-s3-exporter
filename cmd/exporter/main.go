package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/obezpalko/helm-s3-exporter/internal/analyzer"
	"github.com/obezpalko/helm-s3-exporter/internal/metrics"
	"github.com/obezpalko/helm-s3-exporter/internal/s3"
	"github.com/obezpalko/helm-s3-exporter/internal/web"
	"github.com/obezpalko/helm-s3-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Helm S3 Exporter...")

	// Load configuration
	cfg := config.LoadFromEnv()

	// Validate configuration
	if cfg.S3Bucket == "" {
		log.Fatal("S3_BUCKET environment variable is required")
	}

	if !cfg.UseIAMRole && (cfg.AWSAccessKey == "" || cfg.AWSSecretKey == "") {
		log.Println("WARNING: USE_IAM_ROLE is false but credentials are not fully configured")
		log.Println("WARNING: Using static credentials is not recommended for production")
	}

	log.Printf("Configuration loaded:")
	log.Printf("  S3 Bucket: %s", cfg.S3Bucket)
	log.Printf("  S3 Region: %s", cfg.S3Region)
	log.Printf("  S3 Key: %s", cfg.S3Key)
	log.Printf("  Use IAM Role: %v", cfg.UseIAMRole)
	log.Printf("  Scan Interval: %v", cfg.ScanInterval)
	log.Printf("  Scan Timeout: %v", cfg.ScanTimeout)
	log.Printf("  Metrics Port: %s", cfg.MetricsPort)
	log.Printf("  Enable HTML: %v", cfg.EnableHTML)

	// Create S3 client
	ctx := context.Background()
	s3Client, err := s3.NewClient(ctx, cfg.S3Region, cfg.S3Bucket, cfg.UseIAMRole, cfg.AWSAccessKey, cfg.AWSSecretKey)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}
	log.Println("S3 client created successfully")

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
		w.Write([]byte("OK"))
	})

	// Add readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	if cfg.EnableHTML && htmlGenerator != nil {
		mux.Handle(cfg.HTMLPath, htmlGenerator)
		log.Printf("HTML dashboard enabled at %s", cfg.HTMLPath)
	}

	server := &http.Server{
		Addr:    ":" + cfg.MetricsPort,
		Handler: mux,
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on :%s", cfg.MetricsPort)
		log.Printf("Metrics available at http://localhost:%s%s", cfg.MetricsPort, cfg.MetricsPath)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Start scraping loop
	ticker := time.NewTicker(cfg.ScanInterval)
	defer ticker.Stop()

	// Perform initial scrape
	performScrape(ctx, s3Client, cfg, metricsCollector, htmlGenerator)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Main loop
	log.Println("Exporter started successfully")
	for {
		select {
		case <-ticker.C:
			performScrape(ctx, s3Client, cfg, metricsCollector, htmlGenerator)
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

func performScrape(ctx context.Context, s3Client *s3.Client, cfg *config.Config, metricsCollector *metrics.Metrics, htmlGenerator *web.HTMLGenerator) {
	log.Println("Starting scrape...")
	startTime := time.Now()

	// Create context with timeout
	scrapeCtx, cancel := context.WithTimeout(ctx, cfg.ScanTimeout)
	defer cancel()

	// Fetch index.yaml
	data, err := s3Client.GetIndexYAML(scrapeCtx, cfg.S3Key)
	if err != nil {
		log.Printf("ERROR: Failed to fetch index.yaml: %v", err)
		metricsCollector.RecordError()
		metricsCollector.ScrapeDuration.Observe(time.Since(startTime).Seconds())
		return
	}

	// Parse index
	index, err := analyzer.ParseIndex(data)
	if err != nil {
		log.Printf("ERROR: Failed to parse index.yaml: %v", err)
		metricsCollector.RecordError()
		metricsCollector.ScrapeDuration.Observe(time.Since(startTime).Seconds())
		return
	}

	// Analyze charts
	analysis := analyzer.AnalyzeCharts(index)

	// Update metrics
	metricsCollector.Update(analysis)
	metricsCollector.RecordSuccess()

	// Update HTML dashboard if enabled
	if htmlGenerator != nil {
		htmlGenerator.Update(analysis)
	}

	duration := time.Since(startTime)
	metricsCollector.ScrapeDuration.Observe(duration.Seconds())

	log.Printf("Scrape completed in %v", duration)
	log.Printf("  Total charts: %d", analysis.TotalCharts)
	log.Printf("  Total versions: %d", analysis.TotalVersions)
	if !analysis.OldestChartDate.IsZero() {
		log.Printf("  Oldest chart: %s", analysis.OldestChartDate.Format("2006-01-02 15:04:05"))
	}
	if !analysis.NewestChartDate.IsZero() {
		log.Printf("  Newest chart: %s", analysis.NewestChartDate.Format("2006-01-02 15:04:05"))
	}
}
