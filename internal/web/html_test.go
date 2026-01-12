package web

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/obezpalko/helm-repo-exporter/internal/analyzer"
)

func TestSanitizeIconURL_Internal(t *testing.T) {
	// Test through the template function since sanitizeIconURL is not exported
	gen, err := NewHTMLGenerator()
	if err != nil {
		t.Fatalf("Failed to create HTML generator: %v", err)
	}

	// We'll test the sanitization through actual rendering
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Valid HTTP/HTTPS URLs
		{
			name:     "valid https URL",
			input:    "https://example.com/icon.svg",
			expected: "https://example.com/icon.svg",
		},
		{
			name:     "valid http URL",
			input:    "http://example.com/icon.png",
			expected: "http://example.com/icon.png",
		},

		// Valid data URIs
		{
			name:     "valid data URI - SVG",
			input:    "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjwvc3ZnPg==",
			expected: "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjwvc3ZnPg==",
		},
		{
			name:     "valid data URI - PNG",
			input:    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			expected: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		},
		{
			name:     "valid data URI - JPEG",
			input:    "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			expected: "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
		},

		// XSS attack attempts - should be blocked
		{
			name:     "XSS - javascript protocol",
			input:    "javascript:alert('XSS')",
			expected: "",
		},
		{
			name:     "XSS - data URI with javascript",
			input:    "data:text/html,<script>alert('XSS')</script>",
			expected: "",
		},
		{
			name:     "XSS - quote breakout attempt",
			input:    `" onerror="alert('XSS')`,
			expected: "",
		},
		{
			name:     "XSS - event handler injection",
			input:    `https://example.com/icon.svg" onerror="alert('XSS')`,
			expected: "", // This should be blocked or escaped
		},
		{
			name:     "XSS - data URI without base64",
			input:    "data:image/svg+xml,<svg onload=alert('XSS')></svg>",
			expected: "", // Blocked because not base64
		},
		{
			name:     "XSS - vbscript protocol",
			input:    "vbscript:msgbox('XSS')",
			expected: "",
		},

		// Invalid formats
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid data URI format",
			input:    "data:image/png",
			expected: "",
		},
		{
			name:     "unsupported image type",
			input:    "data:image/bmp;base64,Qk0=",
			expected: "",
		},
		{
			name:     "file protocol",
			input:    "file:///etc/passwd",
			expected: "",
		},
		{
			name:     "ftp protocol",
			input:    "ftp://example.com/icon.png",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test by rendering a chart with this icon
			analysis := &analyzer.ChartAnalysis{
				TotalCharts:   1,
				TotalVersions: 1,
				ChartsInfo: []analyzer.ChartInfo{
					{
						Name:       "test-chart",
						Repository: "test",
						Icon:       tt.input,
						VersionDetails: []analyzer.VersionDetail{
							{Version: "1.0.0", Created: time.Now()},
						},
					},
				},
			}

			gen.Update(analysis)
			req := httptest.NewRequest("GET", "/charts", nil)
			w := httptest.NewRecorder()
			gen.ServeHTTP(w, req)

			body := w.Body.String()

			if tt.expected != "" {
				// Should contain the expected URL (possibly escaped)
				if !strings.Contains(body, tt.expected) && !strings.Contains(body, strings.ReplaceAll(tt.expected, "+", "&#43;")) {
					t.Errorf("Expected icon URL not found in output for input: %q", tt.input)
				}
			} else {
				// Should have empty src or no dangerous content
				if strings.Contains(body, "javascript:") || strings.Contains(body, "vbscript:") {
					t.Errorf("Dangerous protocol found in output for input: %q", tt.input)
				}
			}
		})
	}
}

func TestHTMLGenerator_IconRendering(t *testing.T) {
	// Create HTML generator
	gen, err := NewHTMLGenerator()
	if err != nil {
		t.Fatalf("Failed to create HTML generator: %v", err)
	}

	// Test data with different icon types including XSS attempts
	analysis := &analyzer.ChartAnalysis{
		TotalCharts:   4,
		TotalVersions: 4,
		ChartsInfo: []analyzer.ChartInfo{
			{
				Name:       "chart-with-url",
				Repository: "test-repo",
				Icon:       "https://example.com/icon.svg",
				VersionDetails: []analyzer.VersionDetail{
					{Version: "1.0.0", Created: time.Now()},
				},
			},
			{
				Name:       "chart-with-data-uri",
				Repository: "test-repo",
				Icon:       "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjwvc3ZnPg==",
				VersionDetails: []analyzer.VersionDetail{
					{Version: "1.0.0", Created: time.Now()},
				},
			},
			{
				Name:       "chart-with-xss-attempt",
				Repository: "test-repo",
				Icon:       `" onerror="alert('XSS')`,
				VersionDetails: []analyzer.VersionDetail{
					{Version: "1.0.0", Created: time.Now()},
				},
			},
			{
				Name:       "chart-with-javascript",
				Repository: "test-repo",
				Icon:       "javascript:alert('XSS')",
				VersionDetails: []analyzer.VersionDetail{
					{Version: "1.0.0", Created: time.Now()},
				},
			},
		},
	}

	gen.Update(analysis)

	// Create test request
	req := httptest.NewRequest("GET", "/charts", nil)
	w := httptest.NewRecorder()

	// Serve HTML
	gen.ServeHTTP(w, req)

	// Check response
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Verify valid URL icon is present
	if !strings.Contains(body, "https://example.com/icon.svg") {
		t.Error("Valid URL icon not found in HTML output")
	}

	// Verify data URI icon is present (may be URL-escaped by template.URL)
	if !strings.Contains(body, "data:image/svg") {
		t.Error("Data URI icon not found in HTML output")
	}

	// Verify XSS attempts are blocked (should have empty src or sanitized)
	// The XSS attempt should NOT appear as an onerror attribute
	if strings.Contains(body, `onerror="alert('XSS')`) || strings.Contains(body, `onerror=&#34;alert(&#39;XSS&#39;)`) {
		t.Error("XSS attempt was not properly sanitized - onerror handler found in output")
	}

	// Verify javascript: protocol is blocked
	if strings.Contains(body, "javascript:alert") {
		t.Error("JavaScript protocol was not blocked")
	}

	// Count img tags to ensure all charts are rendered (even if icon is blocked)
	imgCount := strings.Count(body, `<img src=`)
	if imgCount < 2 { // At least the 2 valid ones should be present
		t.Errorf("Expected at least 2 img tags, found %d", imgCount)
	}
}

func TestHTMLGenerator_XSSProtection(t *testing.T) {
	gen, err := NewHTMLGenerator()
	if err != nil {
		t.Fatalf("Failed to create HTML generator: %v", err)
	}

	// Test various XSS payloads
	xssPayloads := []string{
		`"><script>alert('XSS')</script>`,
		`' onerror='alert(String.fromCharCode(88,83,83))'`,
		`javascript:alert('XSS')`,
		`data:text/html,<script>alert('XSS')</script>`,
		`vbscript:msgbox('XSS')`,
		`data:image/svg+xml,<svg onload=alert('XSS')></svg>`,
	}

	for _, payload := range xssPayloads {
		analysis := &analyzer.ChartAnalysis{
			TotalCharts:   1,
			TotalVersions: 1,
			ChartsInfo: []analyzer.ChartInfo{
				{
					Name:       "malicious-chart",
					Repository: "test-repo",
					Icon:       payload,
					VersionDetails: []analyzer.VersionDetail{
						{Version: "1.0.0", Created: time.Now()},
					},
				},
			},
		}

		gen.Update(analysis)

		req := httptest.NewRequest("GET", "/charts", nil)
		w := httptest.NewRecorder()
		gen.ServeHTTP(w, req)

		body := w.Body.String()

		// Check that dangerous patterns from the payload are not in the src attribute
		// We look specifically in img src attributes to avoid false positives from our own code
		imgSrcPattern := `<img src="`
		imgStart := strings.Index(body, imgSrcPattern)
		if imgStart != -1 {
			imgStart += len(imgSrcPattern)
			imgEnd := strings.Index(body[imgStart:], `"`)
			if imgEnd != -1 {
				srcValue := body[imgStart : imgStart+imgEnd]

				// Check that the src doesn't contain dangerous patterns
				dangerousPatterns := []string{
					"<script>",
					"javascript:",
					"vbscript:",
					"alert(",
					"onload=",
				}

				for _, pattern := range dangerousPatterns {
					if strings.Contains(srcValue, pattern) {
						t.Errorf("Dangerous pattern %q found in img src for payload: %s", pattern, payload)
					}
				}
			}
		}
	}
}
