package pocketbase

import (
	"net/http"
	"time"
)

// Option represents a functional option for configuring the Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client for the PocketBase client.
// This allows you to configure timeouts, proxies, TLS settings, etc.
//
// Example:
//
//	httpClient := &http.Client{Timeout: 30 * time.Second}
//	client := pocketbase.NewClient("http://localhost:8090", pocketbase.WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithTimeout sets a timeout for HTTP requests by creating a new HTTP client
// with the specified timeout.
//
// Example:
//
//	client := pocketbase.NewClient("http://localhost:8090", pocketbase.WithTimeout(10*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.HTTPClient = &http.Client{Timeout: timeout}
	}
}

// WithUserAgent sets a custom User-Agent header for all requests.
//
// Example:
//
//	client := pocketbase.NewClient("http://localhost:8090", pocketbase.WithUserAgent("MyApp/1.0"))
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}
