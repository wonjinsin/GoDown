package http

import (
	"cheetah/internal/domain/repository"
	"cheetah/util"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// httpRepository implements the HTTPRepository interface
type httpRepository struct{}

// NewHTTPRepository creates a new HTTP repository
func NewHTTPRepository() repository.HTTPRepository {
	return &httpRepository{}
}

// CreateClient creates a new HTTP client with the given configuration
func (hr *httpRepository) CreateClient(config repository.HTTPClientConfig) repository.HTTPClient {
	return NewHTTPClient(config)
}

// ValidateURL validates if the given string is a valid URL
func (hr *httpRepository) ValidateURL(rawURL string) error {
	if rawURL == "" {
		return util.NewValidationError("url", rawURL, "URL cannot be empty")
	}

	// Check if URL starts with http or https
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return util.NewValidationError("url", rawURL, "URL must start with http:// or https://")
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return util.NewValidationError("url", rawURL, fmt.Sprintf("Invalid URL format: %v", err))
	}

	// Check if host is present
	if parsedURL.Host == "" {
		return util.NewValidationError("url", rawURL, "URL must have a valid host")
	}

	return nil
}

// ExtractDomain extracts the domain (scheme + host) from a URL
func (hr *httpRepository) ExtractDomain(rawURL string) string {
	r := regexp.MustCompile(`^((https?):\/\/([^\/]+))`)
	matches := r.FindStringSubmatch(rawURL)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// BuildHeaders builds a standard set of HTTP headers
func (hr *httpRepository) BuildHeaders(host, origin, userAgent string) map[string]string {
	headers := make(map[string]string)

	if userAgent != "" {
		headers["User-Agent"] = userAgent
	}

	if host != "" {
		headers["Host"] = host
	}

	if origin != "" {
		headers["Referer"] = origin
	}

	// Add common headers
	headers["Connection"] = "keep-alive"
	headers["Accept"] = "*/*"
	headers["Accept-Language"] = "en-US,en;q=0.9"
	headers["Accept-Encoding"] = "gzip, deflate, br"
	headers["sec-ch-ua"] = `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`
	headers["sec-ch-ua-mobile"] = "?0"
	headers["sec-ch-ua-platform"] = `"macOS"`
	headers["Sec-Fetch-Dest"] = "empty"
	headers["Sec-Fetch-Mode"] = "cors"
	headers["Sec-Fetch-Site"] = "same-origin"

	return headers
}
