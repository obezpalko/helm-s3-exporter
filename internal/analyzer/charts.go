package analyzer

import (
	"fmt"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// HelmIndex represents the structure of a Helm repository index.yaml
type HelmIndex struct {
	APIVersion string                        `yaml:"apiVersion"`
	Entries    map[string][]ChartVersionInfo `yaml:"entries"`
	Generated  time.Time                     `yaml:"generated"`
}

// ChartVersionInfo represents a single chart version in the index
type ChartVersionInfo struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	Description string    `yaml:"description"`
	Icon        string    `yaml:"icon"`
	Created     time.Time `yaml:"created"`
}

// ChartAnalysis contains analyzed information about charts
type ChartAnalysis struct {
	TotalCharts     int
	TotalVersions   int
	ChartsInfo      []ChartInfo
	OldestChartDate time.Time
	NewestChartDate time.Time
	MedianChartDate time.Time
}

// ChartInfo contains information about a single chart
type ChartInfo struct {
	Name          string
	VersionCount  int
	Versions      []string
	OldestVersion time.Time
	NewestVersion time.Time
	MedianVersion time.Time
	Icon          string
	Description   string
}

// ParseIndex parses the Helm index.yaml content
func ParseIndex(data []byte) (*HelmIndex, error) {
	var index HelmIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index.yaml: %w", err)
	}
	return &index, nil
}

// AnalyzeCharts performs analysis on the Helm index
func AnalyzeCharts(index *HelmIndex) *ChartAnalysis {
	analysis := &ChartAnalysis{
		TotalCharts: len(index.Entries),
		ChartsInfo:  make([]ChartInfo, 0, len(index.Entries)),
	}

	var allDates []time.Time

	for chartName, versions := range index.Entries {
		if len(versions) == 0 {
			continue
		}

		chartInfo := ChartInfo{
			Name:         chartName,
			VersionCount: len(versions),
			Versions:     make([]string, 0, len(versions)),
		}

		var dates []time.Time
		for _, version := range versions {
			analysis.TotalVersions++
			chartInfo.Versions = append(chartInfo.Versions, version.Version)

			if !version.Created.IsZero() {
				dates = append(dates, version.Created)
				allDates = append(allDates, version.Created)
			}

			// Capture icon and description from the latest version
			if chartInfo.Icon == "" && version.Icon != "" {
				chartInfo.Icon = version.Icon
			}
			if chartInfo.Description == "" && version.Description != "" {
				chartInfo.Description = version.Description
			}
		}

		if len(dates) > 0 {
			sort.Slice(dates, func(i, j int) bool {
				return dates[i].Before(dates[j])
			})
			chartInfo.OldestVersion = dates[0]
			chartInfo.NewestVersion = dates[len(dates)-1]
			chartInfo.MedianVersion = dates[len(dates)/2]
		}

		analysis.ChartsInfo = append(analysis.ChartsInfo, chartInfo)
	}

	// Calculate overall dates
	if len(allDates) > 0 {
		sort.Slice(allDates, func(i, j int) bool {
			return allDates[i].Before(allDates[j])
		})
		analysis.OldestChartDate = allDates[0]
		analysis.NewestChartDate = allDates[len(allDates)-1]
		analysis.MedianChartDate = allDates[len(allDates)/2]
	}

	// Sort charts by name for consistent output
	sort.Slice(analysis.ChartsInfo, func(i, j int) bool {
		return analysis.ChartsInfo[i].Name < analysis.ChartsInfo[j].Name
	})

	return analysis
}
