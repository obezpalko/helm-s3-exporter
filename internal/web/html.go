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
    <title>Helm S3 Repository Dashboard</title>
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
        .chart-header {
            display: flex;
            align-items: center;
            margin-bottom: 10px;
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
        }
        .date {
            color: #718096;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Helm S3 Repository Dashboard</h1>
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

        <div class="charts-container">
            <h2 style="margin-bottom: 20px; color: #2d3748;">📦 Available Charts</h2>
            <div class="charts-grid">
                {{range .Analysis.ChartsInfo}}
                <div class="chart-item">
                    <div class="chart-header">
                        {{if .Icon}}
                        <img src="{{.Icon}}" alt="{{.Name}}" class="chart-icon" onerror="this.style.display='none'">
                        {{end}}
                        <div class="chart-name">{{.Name}}</div>
                    </div>
                    {{if .Description}}
                    <div class="chart-description">{{.Description}}</div>
                    {{end}}
                    <div class="chart-meta">
                        <div class="meta-item">
                            <span class="badge">{{.VersionCount}} versions</span>
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
                </div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>
`

