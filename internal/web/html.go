package web

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/obezpalko/helm-repo-exporter/internal/analyzer"
)

// HTMLGenerator generates HTML dashboard for charts
type HTMLGenerator struct {
	mu           sync.RWMutex
	analysis     *analyzer.ChartAnalysis
	repoAnalyses map[string]*analyzer.ChartAnalysis // Per-repository analysis cache
	template     *template.Template
}

// sanitizeIconURL validates and sanitizes icon URLs to prevent XSS attacks
// It allows:
// - HTTP/HTTPS URLs
// - Data URIs (for inline images with base64 encoding)
// Returns empty string if the URL is potentially malicious
func sanitizeIconURL(iconURL string) string {
	if iconURL == "" {
		return ""
	}

	// Check for data URI (inline images)
	if strings.HasPrefix(iconURL, "data:image/") {
		// Validate data URI format: data:image/<type>;base64,<data>
		// Only allow safe image types and base64 encoding
		parts := strings.SplitN(iconURL, ",", 2)
		if len(parts) != 2 {
			return "" // Invalid data URI format
		}

		header := parts[0]
		// Must be: data:image/<type>;base64
		if !strings.HasPrefix(header, "data:image/") {
			return ""
		}
		if !strings.Contains(header, ";base64") {
			return "" // Only allow base64 encoding to prevent inline scripts
		}

		// Check for allowed image types
		allowedTypes := []string{"svg+xml", "png", "jpeg", "jpg", "gif", "webp"}
		hasValidType := false
		for _, imgType := range allowedTypes {
			if strings.Contains(header, "data:image/"+imgType) {
				hasValidType = true
				break
			}
		}
		if !hasValidType {
			return ""
		}

		// Data URI is valid, return as-is
		return iconURL
	}

	// Parse as regular URL
	parsedURL, err := url.Parse(iconURL)
	if err != nil {
		return "" // Invalid URL
	}

	// Only allow http and https schemes
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return "" // Reject dangerous schemes (javascript:, vbscript:, data:, file:, etc.)
	}

	// URL is valid
	return iconURL
}

// NewHTMLGenerator creates a new HTML generator
func NewHTMLGenerator() (*HTMLGenerator, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"safeIconURL": func(iconURL string) template.URL {
			// Sanitize the URL first to block XSS attacks
			sanitized := sanitizeIconURL(iconURL)
			// Return as template.URL which is safe for src attributes
			// template.URL prevents additional escaping while still being safe
			return template.URL(sanitized)
		},
	}

	tmpl, err := template.New("charts").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}

	return &HTMLGenerator{
		template:     tmpl,
		repoAnalyses: make(map[string]*analyzer.ChartAnalysis),
	}, nil
}

// Update updates the analysis data
// If the analysis has a single repository, it updates that repo's data and merges all repos
// If the analysis has multiple repositories (merged), it replaces the entire view
func (h *HTMLGenerator) Update(analysis *analyzer.ChartAnalysis) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if this is a single-repo update or a merged update
	if len(analysis.ChartsInfo) > 0 {
		// Determine if this is a single repository or merged data
		// Single repo: all charts have the same repository name
		// Merged: charts may have different repository names
		firstRepo := analysis.ChartsInfo[0].Repository
		isSingleRepo := true
		for _, chart := range analysis.ChartsInfo {
			if chart.Repository != firstRepo {
				isSingleRepo = false
				break
			}
		}

		if isSingleRepo && firstRepo != "" {
			// Update the specific repository's data
			h.repoAnalyses[firstRepo] = analysis

			// Merge all repository analyses
			h.analysis = h.mergeAllRepos()
		} else {
			// This is already merged data (from initial scrape or full update)
			h.analysis = analysis

			// Also update the per-repo cache if we can
			for _, chart := range analysis.ChartsInfo {
				if chart.Repository != "" {
					// Note: This is a best-effort cache update for merged data
					// Individual repo data will be properly updated on next scrape
				}
			}
		}
	} else {
		// Empty analysis, just set it
		h.analysis = analysis
	}
}

// mergeAllRepos merges all cached repository analyses into a single view
func (h *HTMLGenerator) mergeAllRepos() *analyzer.ChartAnalysis {
	if len(h.repoAnalyses) == 0 {
		return &analyzer.ChartAnalysis{}
	}

	var merged *analyzer.ChartAnalysis
	for _, repoAnalysis := range h.repoAnalyses {
		if merged == nil {
			// Deep copy the first analysis
			merged = &analyzer.ChartAnalysis{
				TotalCharts:     repoAnalysis.TotalCharts,
				TotalVersions:   repoAnalysis.TotalVersions,
				ChartsInfo:      append([]analyzer.ChartInfo{}, repoAnalysis.ChartsInfo...),
				OldestChartDate: repoAnalysis.OldestChartDate,
				NewestChartDate: repoAnalysis.NewestChartDate,
				MedianChartDate: repoAnalysis.MedianChartDate,
			}
		} else {
			// Merge this repo's data
			merged.TotalCharts += repoAnalysis.TotalCharts
			merged.TotalVersions += repoAnalysis.TotalVersions
			merged.ChartsInfo = append(merged.ChartsInfo, repoAnalysis.ChartsInfo...)

			// Update date statistics
			if !repoAnalysis.OldestChartDate.IsZero() {
				if merged.OldestChartDate.IsZero() || repoAnalysis.OldestChartDate.Before(merged.OldestChartDate) {
					merged.OldestChartDate = repoAnalysis.OldestChartDate
				}
			}

			if !repoAnalysis.NewestChartDate.IsZero() {
				if merged.NewestChartDate.IsZero() || repoAnalysis.NewestChartDate.After(merged.NewestChartDate) {
					merged.NewestChartDate = repoAnalysis.NewestChartDate
				}
			}

			// For median, approximate by averaging (good enough for display)
			if !repoAnalysis.MedianChartDate.IsZero() && !merged.MedianChartDate.IsZero() {
				avg := (merged.MedianChartDate.Unix() + repoAnalysis.MedianChartDate.Unix()) / 2
				merged.MedianChartDate = time.Unix(avg, 0)
			} else if !repoAnalysis.MedianChartDate.IsZero() {
				merged.MedianChartDate = repoAnalysis.MedianChartDate
			}
		}
	}

	return merged
}

// ServeHTTP handles HTTP requests for the charts dashboard
func (h *HTMLGenerator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	analysis := h.analysis
	h.mu.RUnlock()

	if analysis == nil {
		http.Error(w, "No data available yet", http.StatusServiceUnavailable)
		return
	}

	data := struct {
		Analysis  *analyzer.ChartAnalysis
		Generated time.Time
	}{
		Analysis:  analysis,
		Generated: time.Now(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Helm Repository Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        .header {
            background: white;
            border-radius: 10px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2d3748;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #718096;
            font-size: 14px;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .stat-card {
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .stat-value {
            font-size: 36px;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 5px;
        }
        .stat-label {
            color: #718096;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .filters {
            background: white;
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            align-items: center;
        }
        .filter-group {
            flex: 1;
            min-width: 250px;
        }
        .filter-label {
            display: block;
            font-size: 12px;
            font-weight: 600;
            color: #4a5568;
            margin-bottom: 5px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .filter-input {
            width: 100%;
            padding: 10px 15px;
            border: 2px solid #e2e8f0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.2s;
        }
        .filter-input:focus {
            outline: none;
            border-color: #667eea;
        }
        .filter-stats {
            color: #718096;
            font-size: 14px;
            padding: 10px 0;
        }
        .charts-container {
            background: white;
            border-radius: 10px;
            padding: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .charts-grid {
            display: grid;
            gap: 15px;
        }
        .chart-item {
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 20px;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .chart-item:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .chart-item.hidden {
            display: none;
        }
        .chart-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 10px;
        }
        .chart-title-section {
            display: flex;
            align-items: center;
            flex: 1;
        }
        .chart-icon {
            width: 40px;
            height: 40px;
            margin-right: 15px;
            border-radius: 5px;
        }
        .chart-name {
            font-size: 20px;
            font-weight: 600;
            color: #2d3748;
        }
        .chart-description {
            color: #718096;
            font-size: 14px;
            margin-bottom: 10px;
        }
        .chart-meta {
            display: flex;
            flex-wrap: wrap;
            gap: 15px;
            font-size: 13px;
            color: #4a5568;
            margin-bottom: 10px;
        }
        .meta-item {
            display: flex;
            align-items: center;
        }
        .meta-label {
            font-weight: 600;
            margin-right: 5px;
        }
        .badge {
            background: #667eea;
            color: white;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.2s;
        }
        .badge:hover {
            background: #5568d3;
        }
        .date {
            color: #718096;
        }
        .versions-list {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.3s ease-out;
            margin-top: 10px;
        }
        .versions-list.expanded {
            max-height: 500px;
            overflow-y: auto;
        }
        .version-item {
            padding: 8px 12px;
            border-left: 3px solid #667eea;
            background: #f7fafc;
            margin-bottom: 5px;
            border-radius: 4px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .version-number {
            font-weight: 600;
            color: #2d3748;
            font-family: 'Courier New', monospace;
        }
        .version-date {
            color: #718096;
            font-size: 12px;
        }
        .version-link {
            color: #667eea;
            text-decoration: none;
            font-size: 12px;
            font-weight: 600;
            padding: 4px 8px;
            border-radius: 4px;
            transition: background 0.2s;
        }
        .version-link:hover {
            background: #667eea;
            color: white;
        }
        .expand-icon {
            margin-left: 5px;
            font-size: 10px;
            transition: transform 0.2s;
        }
        .expanded-icon {
            transform: rotate(180deg);
        }
        .no-results {
            text-align: center;
            padding: 40px;
            color: #718096;
            font-size: 16px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Helm Repository Dashboard</h1>
            <p class="subtitle">Generated: {{.Generated.Format "2006-01-02 15:04:05 MST"}}</p>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">{{.Analysis.TotalCharts}}</div>
                <div class="stat-label">Total Charts</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{{.Analysis.TotalVersions}}</div>
                <div class="stat-label">Total Versions</div>
            </div>
            {{if not .Analysis.OldestChartDate.IsZero}}
            <div class="stat-card">
                <div class="stat-value">{{.Analysis.OldestChartDate.Format "2006-01-02"}}</div>
                <div class="stat-label">Oldest Chart</div>
            </div>
            {{end}}
            {{if not .Analysis.NewestChartDate.IsZero}}
            <div class="stat-card">
                <div class="stat-value">{{.Analysis.NewestChartDate.Format "2006-01-02"}}</div>
                <div class="stat-label">Newest Chart</div>
            </div>
            {{end}}
        </div>

        <div class="filters">
            <div class="filter-group">
                <label class="filter-label" for="repoFilter">Filter by Repository</label>
                <select id="repoFilter" class="filter-input">
                    <option value="">All Repositories</option>
                </select>
            </div>
            <div class="filter-group">
                <label class="filter-label" for="chartFilter">Filter by Chart Name</label>
                <input type="text" id="chartFilter" class="filter-input" placeholder="Type to filter charts...">
            </div>
            <div class="filter-stats" id="filterStats">
                Showing <span id="visibleCount">{{.Analysis.TotalCharts}}</span> of {{.Analysis.TotalCharts}} charts
            </div>
        </div>

        <div class="charts-container">
            <h2 style="margin-bottom: 20px; color: #2d3748;">📦 Available Charts</h2>
            <div class="charts-grid" id="chartsGrid">
                {{range .Analysis.ChartsInfo}}
                <div class="chart-item" data-chart-name="{{.Name}}" data-repository="{{.Repository}}">
                    <div class="chart-header">
                        <div class="chart-title-section">
                            {{if .Icon}}
                            <img src="{{safeIconURL .Icon}}" alt="{{.Name}}" class="chart-icon" onerror="this.style.display='none'">
                            {{end}}
                            <div>
                                <div class="chart-name">{{.Name}}</div>
                                {{if .Repository}}
                                <div style="font-size: 12px; color: #667eea; font-weight: 600; margin-top: 2px;">📁 {{.Repository}}</div>
                                {{end}}
                            </div>
                        </div>
                    </div>
                    {{if .Description}}
                    <div class="chart-description">{{.Description}}</div>
                    {{end}}
                    <div class="chart-meta">
                        <div class="meta-item">
                            <span class="badge" onclick="toggleVersions(this)">
                                {{.VersionCount}} versions
                                <span class="expand-icon">▼</span>
                            </span>
                        </div>
                        {{if not .OldestVersion.IsZero}}
                        <div class="meta-item">
                            <span class="meta-label">Oldest:</span>
                            <span class="date">{{.OldestVersion.Format "2006-01-02"}}</span>
                        </div>
                        {{end}}
                        {{if not .NewestVersion.IsZero}}
                        <div class="meta-item">
                            <span class="meta-label">Newest:</span>
                            <span class="date">{{.NewestVersion.Format "2006-01-02"}}</span>
                        </div>
                        {{end}}
                    </div>
                    <div class="versions-list">
                        {{range .VersionDetails}}
                        <div class="version-item">
                            <div>
                                <span class="version-number">{{.Version}}</span>
                                {{if not .Created.IsZero}}
                                <span class="version-date"> • {{.Created.Format "2006-01-02"}}</span>
                                {{end}}
                            </div>
                            {{if .URL}}
                            <a href="{{.URL}}" class="version-link" target="_blank">Download</a>
                            {{end}}
                        </div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
            <div class="no-results" id="noResults" style="display: none;">
                No charts match your filter criteria
            </div>
        </div>
    </div>

    <script>
        // Toggle versions list
        function toggleVersions(badge) {
            const chartItem = badge.closest('.chart-item');
            const versionsList = chartItem.querySelector('.versions-list');
            const expandIcon = badge.querySelector('.expand-icon');
            
            versionsList.classList.toggle('expanded');
            expandIcon.classList.toggle('expanded-icon');
        }

        // Filter functionality
        const repoFilter = document.getElementById('repoFilter');
        const chartFilter = document.getElementById('chartFilter');
        const chartsGrid = document.getElementById('chartsGrid');
        const noResults = document.getElementById('noResults');
        const visibleCount = document.getElementById('visibleCount');
        const chartItems = document.querySelectorAll('.chart-item');

        // Populate repository dropdown
        const repositories = new Set();
        chartItems.forEach(item => {
            const repo = item.dataset.repository;
            if (repo) {
                repositories.add(repo);
            }
        });
        
        // Sort repositories alphabetically
        const sortedRepos = Array.from(repositories).sort();
        sortedRepos.forEach(repo => {
            const option = document.createElement('option');
            option.value = repo;
            option.textContent = repo;
            repoFilter.appendChild(option);
        });

        function filterCharts() {
            const repoQuery = repoFilter.value.toLowerCase();
            const chartQuery = chartFilter.value.toLowerCase();
            let visible = 0;

            chartItems.forEach(item => {
                const chartName = item.dataset.chartName.toLowerCase();
                const repository = item.dataset.repository.toLowerCase();
                
                const repoMatches = !repoQuery || repository === repoQuery;
                const chartMatches = !chartQuery || chartName.includes(chartQuery);

                if (repoMatches && chartMatches) {
                    item.classList.remove('hidden');
                    visible++;
                } else {
                    item.classList.add('hidden');
                }
            });

            visibleCount.textContent = visible;
            
            if (visible === 0) {
                chartsGrid.style.display = 'none';
                noResults.style.display = 'block';
            } else {
                chartsGrid.style.display = 'grid';
                noResults.style.display = 'none';
            }
        }

        repoFilter.addEventListener('change', filterCharts);
        chartFilter.addEventListener('input', filterCharts);

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            // Focus search on '/' key
            if (e.key === '/' && document.activeElement !== chartFilter) {
                e.preventDefault();
                chartFilter.focus();
            }
            // Clear search on 'Escape' key
            if (e.key === 'Escape' && document.activeElement === chartFilter) {
                chartFilter.value = '';
                filterCharts();
                chartFilter.blur();
            }
        });
    </script>
</body>
</html>
`
