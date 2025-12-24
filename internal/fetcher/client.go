package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/obezpalko/helm-s3-exporter/pkg/config"
)

// Client wraps the HTTP client for fetching index.yaml
type Client struct {
	httpClient *http.Client
	repo       config.Repository
}

// NewClient creates a new HTTP client for fetching index.yaml
func NewClient(repo config.Repository, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		repo: repo,
	}
}

// GetIndexYAML retrieves and returns the index.yaml file from the URL
func (c *Client) GetIndexYAML(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.repo.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if configured
	if c.repo.Auth != nil {
		if err := c.addAuthentication(req); err != nil {
			return nil, fmt.Errorf("failed to add authentication: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", c.repo.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, c.repo.URL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// addAuthentication adds authentication headers to the request
func (c *Client) addAuthentication(req *http.Request) error {
	if c.repo.Auth.Basic != nil {
		req.SetBasicAuth(c.repo.Auth.Basic.Username, c.repo.Auth.Basic.Password)
	}

	if c.repo.Auth.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.repo.Auth.BearerToken)
	}

	for key, value := range c.repo.Auth.Headers {
		req.Header.Set(key, value)
	}

	return nil
}

// RepositoryName returns the name of the repository
func (c *Client) RepositoryName() string {
	return c.repo.Name
}
