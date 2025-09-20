package main

import (
	"context"
	"net/http"
	"time"

	"github.com/0x113/pocketbase-go"
)

// CreateClient creates a new PocketBase client with common configuration
func CreateClient(baseURL string) *pocketbase.Client {
	return pocketbase.NewClient(
		baseURL,
		pocketbase.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		pocketbase.WithUserAgent("PocketBase-Go-Example/1.0"),
	)
}

// CreateContext creates a context with timeout for operations
func CreateContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// SetupDemoClient creates a client and sets a demo token for examples
func SetupDemoClient(baseURL string) *pocketbase.Client {
	client := CreateClient(baseURL)
	client.SetToken("demo-token-123")
	return client
}

// CreateSuperuserClient creates a client specifically for superuser operations
func CreateSuperuserClient(baseURL string) *pocketbase.Client {
	return pocketbase.NewClient(
		baseURL,
		pocketbase.WithTimeout(10*time.Second),
	)
}
