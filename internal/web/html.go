package web

import (
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/obezpalko/helm-s3-exporter/internal/analyzer"
)

// HTMLGenerator generates HTML dashboard for charts
type HTMLGenerator struct {
	mu       sync.RWMutex
	analysis *analyzer.ChartAnalysis
	template *template.Template
}

// NewHTMLGenerator creates a new HTML generator
func NewHTMLGenerator() (*HTMLGenerator, error) {
	tmpl, err := template.New("charts").Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}

	return &HTMLGenerator{
		template: tmpl,
	}, nil
}

// Update updates the analysis data
func (h *HTMLGenerator) Update(analysis *analyzer.ChartAnalysis) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.analysis = analysis
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
                <div class="chart-item" data-chart-name="{{.Name}}">
                    <div class="chart-header">
                        <div class="chart-title-section">
                            {{if .Icon}}
                            <img src="{{.Icon}}" alt="{{.Name}}" class="chart-icon" onerror="this.style.display='none'">
                            {{end}}
                            <div class="chart-name">{{.Name}}</div>
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
        const chartFilter = document.getElementById('chartFilter');
        const chartsGrid = document.getElementById('chartsGrid');
        const noResults = document.getElementById('noResults');
        const visibleCount = document.getElementById('visibleCount');
        const chartItems = document.querySelectorAll('.chart-item');

        function filterCharts() {
            const chartQuery = chartFilter.value.toLowerCase();
            let visible = 0;

            chartItems.forEach(item => {
                const chartName = item.dataset.chartName.toLowerCase();
                const matches = chartName.includes(chartQuery);

                if (matches) {
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
