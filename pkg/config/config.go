package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	// Repository Configuration
	Repositories []Repository `yaml:"repositories"`

	// Scan Configuration
	ScanInterval time.Duration `yaml:"scanInterval"`
	ScanTimeout  time.Duration `yaml:"scanTimeout"`

	// Server Configuration
	MetricsPort string `yaml:"metricsPort"`
	MetricsPath string `yaml:"metricsPath"`

	// Optional Features
	EnableHTML bool   `yaml:"enableHTML"`
	HTMLPath   string `yaml:"htmlPath"`
}

// Repository defines a Helm repository source
type Repository struct {
	// Name is a friendly identifier for this repository
	Name string `yaml:"name"`

	// URL is the HTTP/HTTPS URL to the index.yaml file
	URL string `yaml:"url"`

	// ScanInterval overrides the global scan interval for this repository
	// If not set, uses the global scanInterval
	ScanInterval time.Duration `yaml:"scanInterval,omitempty"`

	// Authentication configuration
	Auth *AuthConfig `yaml:"auth,omitempty"`
}

// AuthConfig defines authentication methods
type AuthConfig struct {
	// Basic authentication
	Basic *BasicAuth `yaml:"basic,omitempty"`

	// Bearer token authentication
	BearerToken string `yaml:"bearerToken,omitempty"`

	// Custom headers
	Headers map[string]string `yaml:"headers,omitempty"`
}

// BasicAuth holds username and password for basic authentication
type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.ScanInterval == 0 {
		cfg.ScanInterval = 5 * time.Minute
	}
	if cfg.ScanTimeout == 0 {
		cfg.ScanTimeout = 30 * time.Second
	}
	if cfg.MetricsPort == "" {
		cfg.MetricsPort = "9571"
	}
	if cfg.MetricsPath == "" {
		cfg.MetricsPath = "/metrics"
	}
	if cfg.HTMLPath == "" {
		cfg.HTMLPath = "/charts"
	}

	// Apply default scan interval to repositories that don't have one
	for i := range cfg.Repositories {
		if cfg.Repositories[i].ScanInterval == 0 {
			cfg.Repositories[i].ScanInterval = cfg.ScanInterval
		}
	}

	return &cfg, nil
}

// LoadFromEnv loads configuration from environment variables (backward compatibility)
func LoadFromEnv() (*Config, error) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile != "" {
		return LoadFromFile(configFile)
	}

	// Fallback to single URL from env var
	indexURL := os.Getenv("INDEX_URL")
	if indexURL == "" {
		return nil, fmt.Errorf("either CONFIG_FILE or INDEX_URL environment variable is required")
	}

	scanInterval := getEnvDuration("SCAN_INTERVAL", 5*time.Minute)

	cfg := &Config{
		Repositories: []Repository{
			{
				Name:         "default",
				URL:          indexURL,
				ScanInterval: scanInterval,
			},
		},
		ScanInterval: scanInterval,
		ScanTimeout:  getEnvDuration("SCAN_TIMEOUT", 30*time.Second),
		MetricsPort:  getEnv("METRICS_PORT", "9571"),
		MetricsPath:  getEnv("METRICS_PATH", "/metrics"),
		EnableHTML:   getEnvBool("ENABLE_HTML", false),
		HTMLPath:     getEnv("HTML_PATH", "/charts"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
