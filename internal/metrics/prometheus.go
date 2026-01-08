package metrics

import (
	"github.com/obezpalko/helm-repo-exporter/internal/analyzer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	ChartsTotal       *prometheus.GaugeVec
	ChartVersions     *prometheus.GaugeVec
	ChartAgeOldest    *prometheus.GaugeVec
	ChartAgeNewest    *prometheus.GaugeVec
	ChartAgeMedian    *prometheus.GaugeVec
	OverallAgeOldest  *prometheus.GaugeVec
	OverallAgeNewest  *prometheus.GaugeVec
	OverallAgeMedian  *prometheus.GaugeVec
	TotalVersions     *prometheus.GaugeVec
	ScrapeDuration    *prometheus.HistogramVec
	ScrapeErrors      *prometheus.CounterVec
	LastScrapeSuccess *prometheus.GaugeVec
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		ChartsTotal: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_charts_total",
			Help: "Total number of distinct Helm charts in the repository",
		}, []string{"repository"}),
		ChartVersions: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_chart_versions",
			Help: "Number of versions for each Helm chart",
		}, []string{"repository", "chart"}),
		ChartAgeOldest: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_chart_age_oldest_seconds",
			Help: "Timestamp of the oldest version of each chart",
		}, []string{"repository", "chart"}),
		ChartAgeNewest: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_chart_age_newest_seconds",
			Help: "Timestamp of the newest version of each chart",
		}, []string{"repository", "chart"}),
		ChartAgeMedian: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_chart_age_median_seconds",
			Help: "Timestamp of the median version of each chart",
		}, []string{"repository", "chart"}),
		OverallAgeOldest: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_overall_age_oldest_seconds",
			Help: "Timestamp of the oldest chart version in the repository",
		}, []string{"repository"}),
		OverallAgeNewest: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_overall_age_newest_seconds",
			Help: "Timestamp of the newest chart version in the repository",
		}, []string{"repository"}),
		OverallAgeMedian: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_overall_age_median_seconds",
			Help: "Timestamp of the median chart version in the repository",
		}, []string{"repository"}),
		TotalVersions: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_versions_total",
			Help: "Total number of chart versions in the repository",
		}, []string{"repository"}),
		ScrapeDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "helm_repo_scrape_duration_seconds",
			Help:    "Duration of the repository scrape operation in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"repository"}),
		ScrapeErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "helm_repo_scrape_errors_total",
			Help: "Total number of scrape errors per repository",
		}, []string{"repository"}),
		LastScrapeSuccess: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_repo_last_scrape_success",
			Help: "Timestamp of the last successful scrape per repository",
		}, []string{"repository"}),
	}
}

// Update updates all metrics based on the chart analysis for a specific repository
func (m *Metrics) Update(repository string, analysis *analyzer.ChartAnalysis) {
	// Update per-repository metrics
	m.ChartsTotal.WithLabelValues(repository).Set(float64(analysis.TotalCharts))
	m.TotalVersions.WithLabelValues(repository).Set(float64(analysis.TotalVersions))

	// Update per-chart metrics
	for _, chart := range analysis.ChartsInfo {
		m.ChartVersions.WithLabelValues(repository, chart.Name).Set(float64(chart.VersionCount))

		if !chart.OldestVersion.IsZero() {
			m.ChartAgeOldest.WithLabelValues(repository, chart.Name).Set(float64(chart.OldestVersion.Unix()))
		}
		if !chart.NewestVersion.IsZero() {
			m.ChartAgeNewest.WithLabelValues(repository, chart.Name).Set(float64(chart.NewestVersion.Unix()))
		}
		if !chart.MedianVersion.IsZero() {
			m.ChartAgeMedian.WithLabelValues(repository, chart.Name).Set(float64(chart.MedianVersion.Unix()))
		}
	}

	// Update overall age metrics for this repository
	if !analysis.OldestChartDate.IsZero() {
		m.OverallAgeOldest.WithLabelValues(repository).Set(float64(analysis.OldestChartDate.Unix()))
	}
	if !analysis.NewestChartDate.IsZero() {
		m.OverallAgeNewest.WithLabelValues(repository).Set(float64(analysis.NewestChartDate.Unix()))
	}
	if !analysis.MedianChartDate.IsZero() {
		m.OverallAgeMedian.WithLabelValues(repository).Set(float64(analysis.MedianChartDate.Unix()))
	}
}

// RecordError increments the error counter for a repository
func (m *Metrics) RecordError(repository string) {
	m.ScrapeErrors.WithLabelValues(repository).Inc()
}

// RecordSuccess records a successful scrape for a repository
func (m *Metrics) RecordSuccess(repository string) {
	m.LastScrapeSuccess.WithLabelValues(repository).SetToCurrentTime()
}
