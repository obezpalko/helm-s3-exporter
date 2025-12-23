package metrics

import (
	"github.com/obezpalko/helm-s3-exporter/internal/analyzer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	ChartsTotal       prometheus.Gauge
	ChartVersions     *prometheus.GaugeVec
	ChartAgeOldest    *prometheus.GaugeVec
	ChartAgeNewest    *prometheus.GaugeVec
	ChartAgeMedian    *prometheus.GaugeVec
	OverallAgeOldest  prometheus.Gauge
	OverallAgeNewest  prometheus.Gauge
	OverallAgeMedian  prometheus.Gauge
	TotalVersions     prometheus.Gauge
	ScrapeDuration    prometheus.Histogram
	ScrapeErrors      prometheus.Counter
	LastScrapeSuccess prometheus.Gauge
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		ChartsTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "helm_s3_charts_total",
			Help: "Total number of distinct Helm charts in the S3 repository",
		}),
		ChartVersions: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_s3_chart_versions",
			Help: "Number of versions for each Helm chart",
		}, []string{"chart"}),
		ChartAgeOldest: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_s3_chart_age_oldest_seconds",
			Help: "Timestamp of the oldest version of each chart",
		}, []string{"chart"}),
		ChartAgeNewest: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_s3_chart_age_newest_seconds",
			Help: "Timestamp of the newest version of each chart",
		}, []string{"chart"}),
		ChartAgeMedian: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "helm_s3_chart_age_median_seconds",
			Help: "Timestamp of the median version of each chart",
		}, []string{"chart"}),
		OverallAgeOldest: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "helm_s3_overall_age_oldest_seconds",
			Help: "Timestamp of the oldest chart version across all charts",
		}),
		OverallAgeNewest: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "helm_s3_overall_age_newest_seconds",
			Help: "Timestamp of the newest chart version across all charts",
		}),
		OverallAgeMedian: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "helm_s3_overall_age_median_seconds",
			Help: "Timestamp of the median chart version across all charts",
		}),
		TotalVersions: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "helm_s3_versions_total",
			Help: "Total number of chart versions across all charts",
		}),
		ScrapeDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "helm_s3_scrape_duration_seconds",
			Help:    "Duration of the S3 scrape operation in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		ScrapeErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "helm_s3_scrape_errors_total",
			Help: "Total number of scrape errors",
		}),
		LastScrapeSuccess: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "helm_s3_last_scrape_success",
			Help: "Timestamp of the last successful scrape",
		}),
	}
}

// Update updates all metrics based on the chart analysis
func (m *Metrics) Update(analysis *analyzer.ChartAnalysis) {
	// Reset vector metrics to avoid stale data
	m.ChartVersions.Reset()
	m.ChartAgeOldest.Reset()
	m.ChartAgeNewest.Reset()
	m.ChartAgeMedian.Reset()

	// Update scalar metrics
	m.ChartsTotal.Set(float64(analysis.TotalCharts))
	m.TotalVersions.Set(float64(analysis.TotalVersions))

	// Update per-chart metrics
	for _, chart := range analysis.ChartsInfo {
		m.ChartVersions.WithLabelValues(chart.Name).Set(float64(chart.VersionCount))

		if !chart.OldestVersion.IsZero() {
			m.ChartAgeOldest.WithLabelValues(chart.Name).Set(float64(chart.OldestVersion.Unix()))
		}
		if !chart.NewestVersion.IsZero() {
			m.ChartAgeNewest.WithLabelValues(chart.Name).Set(float64(chart.NewestVersion.Unix()))
		}
		if !chart.MedianVersion.IsZero() {
			m.ChartAgeMedian.WithLabelValues(chart.Name).Set(float64(chart.MedianVersion.Unix()))
		}
	}

	// Update overall age metrics
	if !analysis.OldestChartDate.IsZero() {
		m.OverallAgeOldest.Set(float64(analysis.OldestChartDate.Unix()))
	}
	if !analysis.NewestChartDate.IsZero() {
		m.OverallAgeNewest.Set(float64(analysis.NewestChartDate.Unix()))
	}
	if !analysis.MedianChartDate.IsZero() {
		m.OverallAgeMedian.Set(float64(analysis.MedianChartDate.Unix()))
	}
}

// RecordError increments the error counter
func (m *Metrics) RecordError() {
	m.ScrapeErrors.Inc()
}

// RecordSuccess records a successful scrape
func (m *Metrics) RecordSuccess() {
	m.LastScrapeSuccess.SetToCurrentTime()
}
