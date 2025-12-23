package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// S3 Configuration
	S3Bucket     string
	S3Region     string
	S3Key        string
	UseIAMRole   bool
	AWSAccessKey string
	AWSSecretKey string

	// Scan Configuration
	ScanInterval time.Duration
	ScanTimeout  time.Duration

	// Server Configuration
	MetricsPort string
	MetricsPath string

	// Optional Features
	EnableHTML bool
	HTMLPath   string
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	cfg := &Config{
		S3Bucket:     getEnv("S3_BUCKET", ""),
		S3Region:     getEnv("S3_REGION", "us-east-1"),
		S3Key:        getEnv("S3_KEY", "index.yaml"),
		UseIAMRole:   getEnvBool("USE_IAM_ROLE", true),
		AWSAccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		ScanInterval: getEnvDuration("SCAN_INTERVAL", 5*time.Minute),
		ScanTimeout:  getEnvDuration("SCAN_TIMEOUT", 30*time.Second),
		MetricsPort:  getEnv("METRICS_PORT", "9571"),
		MetricsPath:  getEnv("METRICS_PATH", "/metrics"),
		EnableHTML:   getEnvBool("ENABLE_HTML", false),
		HTMLPath:     getEnv("HTML_PATH", "/charts"),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
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
