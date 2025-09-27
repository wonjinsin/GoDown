package value

import (
	"cheetah/util"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// URL represents a validated URL value object
type URL struct {
	raw    string
	parsed *url.URL
}

// NewURL creates a new URL value object
func NewURL(rawURL string) (*URL, error) {
	if rawURL == "" {
		return nil, util.NewValidationError("url", rawURL, "URL cannot be empty")
	}

	// Basic validation
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return nil, util.NewValidationError("url", rawURL, "URL must start with http:// or https://")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, util.NewValidationError("url", rawURL, fmt.Sprintf("Invalid URL format: %v", err))
	}

	return &URL{
		raw:    rawURL,
		parsed: parsed,
	}, nil
}

// String returns the raw URL string
func (u *URL) String() string {
	return u.raw
}

// Host returns the host part of the URL
func (u *URL) Host() string {
	return u.parsed.Host
}

// Scheme returns the scheme (http/https)
func (u *URL) Scheme() string {
	return u.parsed.Scheme
}

// Path returns the path part of the URL
func (u *URL) Path() string {
	return u.parsed.Path
}

// Query returns the query parameters
func (u *URL) Query() url.Values {
	return u.parsed.Query()
}

// Domain returns the domain with scheme
func (u *URL) Domain() string {
	return fmt.Sprintf("%s://%s", u.parsed.Scheme, u.parsed.Host)
}

// IsSecure returns true if the URL uses HTTPS
func (u *URL) IsSecure() bool {
	return u.parsed.Scheme == "https"
}

// HasFileExtension checks if the URL path has a file extension
func (u *URL) HasFileExtension() bool {
	path := u.parsed.Path
	return strings.Contains(path, ".") && !strings.HasSuffix(path, "/")
}

// GetFileExtension extracts the file extension from the URL
func (u *URL) GetFileExtension() (string, error) {
	r := regexp.MustCompile(`\.(\w+)$|\.(\w+)\?`)
	matches := r.FindStringSubmatch(u.raw)
	if len(matches) < 2 {
		return "", fmt.Errorf("no file extension found in URL: %s", u.raw)
	}

	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			return matches[i], nil
		}
	}

	return "", fmt.Errorf("no file extension found in URL: %s", u.raw)
}

// WithPath returns a new URL with the given path
func (u *URL) WithPath(newPath string) *URL {
	newURL := *u.parsed
	newURL.Path = newPath

	return &URL{
		raw:    newURL.String(),
		parsed: &newURL,
	}
}

// WithQuery returns a new URL with the given query parameters
func (u *URL) WithQuery(query url.Values) *URL {
	newURL := *u.parsed
	newURL.RawQuery = query.Encode()

	return &URL{
		raw:    newURL.String(),
		parsed: &newURL,
	}
}
