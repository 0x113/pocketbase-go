package pocketbase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// Client represents a PocketBase API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	userAgent  string

	// Thread-safe token storage
	tokenMu sync.RWMutex
	token   string
}

// NewClient creates a new PocketBase client with the given base URL and options.
//
// Example:
//
//	client := pocketbase.NewClient("http://localhost:8090")
//	// or with options:
//	client := pocketbase.NewClient("http://localhost:8090",
//		pocketbase.WithTimeout(10*time.Second),
//		pocketbase.WithUserAgent("MyApp/1.0"))
func NewClient(baseURL string, opts ...Option) *Client {
	client := &Client{
		BaseURL:    strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{},
		userAgent:  "pocketbase-go/1.0",
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// SetToken manually sets the authentication token for API requests.
// This is useful when you have a token from previous authentication
// or from another source.
func (c *Client) SetToken(token string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = token
}

// GetToken returns the current authentication token.
func (c *Client) GetToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token
}

// AuthenticateWithPassword authenticates with PocketBase using username/email and password.
// On success, it stores the authentication token for subsequent requests and returns the user record.
//
// Example:
//
//	record, err := client.AuthenticateWithPassword(ctx, "users", "user@example.com", "password123")
//	if err != nil {
//		// Handle error
//		return err
//	}
//	fmt.Printf("Authenticated user: %s", record["email"])
func (c *Client) AuthenticateWithPassword(ctx context.Context, collection, identity, password string) (Record, error) {
	endpoint := fmt.Sprintf("/api/collections/%s/auth-with-password", collection)

	body := map[string]string{
		"identity": identity,
		"password": password,
	}

	var resp authResp
	err := c.doRequest(ctx, "POST", endpoint, body, &resp)
	if err != nil {
		return nil, err
	}

	// Store the token for future requests
	c.SetToken(resp.Token)

	return resp.Record, nil
}

// AuthenticateAsSuperuser authenticates as a PocketBase superuser using email and password.
// This is a convenience method that calls AuthenticateWithPassword with the "_superusers" collection.
// On success, it stores the superuser authentication token for subsequent requests.
//
// Example:
//
//	superuser, err := client.AuthenticateAsSuperuser(ctx, "admin@example.com", "superuser_password")
//	if err != nil {
//		// Handle error
//		return err
//	}
//	fmt.Printf("Authenticated superuser: %s", superuser["email"])
func (c *Client) AuthenticateAsSuperuser(ctx context.Context, email, password string) (Record, error) {
	return c.AuthenticateWithPassword(ctx, "_superusers", email, password)
}

// Impersonate allows superusers to impersonate another user by generating a non-refreshable auth token.
// This method requires superuser authentication. The generated token has a custom duration (in seconds)
// or falls back to the default collection auth token duration if duration is 0 or not provided.
//
// Example:
//
//	// First authenticate as superuser
//	_, err := client.AuthenticateAsSuperuser(ctx, "admin@example.com", "admin_password")
//	if err != nil {
//		return err
//	}
//
//	// Then impersonate a user for 1 hour (3600 seconds)
//	result, err := client.Impersonate(ctx, "users", "user_record_id", 3600)
//	if err != nil {
//		return err
//	}
//
//	// The result contains the impersonation token and user record
//	fmt.Printf("Impersonation token: %s\n", result.Token)
//	fmt.Printf("Impersonated user: %s\n", result.Record["email"])
func (c *Client) Impersonate(ctx context.Context, collection, recordID string, duration int, opts ...QueryOption) (*ImpersonateResult, error) {
	options := &QueryOptions{}
	for _, opt := range opts {
		opt(options)
	}

	endpoint := fmt.Sprintf("/api/collections/%s/impersonate/%s", collection, recordID)

	// Build query parameters
	params := url.Values{}
	if len(options.Expand) > 0 {
		params.Set("expand", strings.Join(options.Expand, ","))
	}
	if len(options.Fields) > 0 {
		params.Set("fields", strings.Join(options.Fields, ","))
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	// Prepare request body with optional duration
	body := make(map[string]any)
	if duration > 0 {
		body["duration"] = duration
	}

	var bodyToSend any
	if len(body) > 0 {
		bodyToSend = body
	}

	var resp impersonateResp
	err := c.doRequest(ctx, "POST", endpoint, bodyToSend, &resp)
	if err != nil {
		return nil, err
	}

	return &ImpersonateResult{
		Token:  resp.Token,
		Record: resp.Record,
	}, nil
}

// GetRecord fetches a single record from a collection by its ID.
//
// Example:
//
//	record, err := client.GetRecord(ctx, "posts", "RECORD_ID_HERE")
//	if err != nil {
//		// Handle error
//		return err
//	}
//	fmt.Printf("Post title: %s", record["title"])
func (c *Client) GetRecord(ctx context.Context, collection, recordID string, opts ...QueryOption) (Record, error) {
	options := &QueryOptions{}
	for _, opt := range opts {
		opt(options)
	}

	endpoint := fmt.Sprintf("/api/collections/%s/records/%s", collection, recordID)

	// Build query parameters
	params := url.Values{}
	if len(options.Expand) > 0 {
		params.Set("expand", strings.Join(options.Expand, ","))
	}
	if len(options.Fields) > 0 {
		params.Set("fields", strings.Join(options.Fields, ","))
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	var record Record
	err := c.doRequest(ctx, "GET", endpoint, nil, &record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// GetAllRecords fetches all records from a collection, automatically handling pagination.
// It continues fetching pages until all records are retrieved.
//
// Example:
//
//	records, err := client.GetAllRecords(ctx, "posts")
//	if err != nil {
//		// Handle error
//		return err
//	}
//	fmt.Printf("Found %d posts", len(records))
func (c *Client) GetAllRecords(ctx context.Context, collection string, opts ...ListOption) ([]Record, error) {
	options := &ListOptions{
		Page:    1,
		PerPage: 30, // PocketBase default
	}
	for _, opt := range opts {
		opt(options)
	}

	var allRecords []Record
	page := 1

	// If a specific page was requested, fetch only that page
	if options.Page > 1 {
		page = options.Page
		records, err := c.getRecordPage(ctx, collection, options, page)
		if err != nil {
			return nil, err
		}
		return records.Items, nil
	}

	// Fetch all pages
	for {
		options.Page = page
		resp, err := c.getRecordPage(ctx, collection, options, page)
		if err != nil {
			return nil, err
		}

		allRecords = append(allRecords, resp.Items...)

		// Check if we've reached the last page
		if page >= resp.TotalPages {
			break
		}
		page++
	}

	return allRecords, nil
}

// getRecordPage fetches a single page of records from a collection.
func (c *Client) getRecordPage(ctx context.Context, collection string, options *ListOptions, page int) (*listResp, error) {
	endpoint := fmt.Sprintf("/api/collections/%s/records", collection)

	// Build query parameters
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	if options.PerPage > 0 {
		params.Set("perPage", strconv.Itoa(options.PerPage))
	}
	if options.Sort != "" {
		params.Set("sort", options.Sort)
	}
	if options.Filter != "" {
		params.Set("filter", options.Filter)
	}
	if len(options.Expand) > 0 {
		params.Set("expand", strings.Join(options.Expand, ","))
	}
	if len(options.Fields) > 0 {
		params.Set("fields", strings.Join(options.Fields, ","))
	}

	endpoint += "?" + params.Encode()

	var resp listResp
	err := c.doRequest(ctx, "GET", endpoint, nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateRecord creates a new record in the specified collection.
// The record parameter should contain the field values for the new record.
// Fields like 'id', 'created', and 'updated' are automatically generated by PocketBase.
//
// Example:
//
//	recordData := map[string]any{
//		"title":   "My New Post",
//		"content": "This is the content of my post",
//		"status":  "draft",
//		"author":  "user-id-123",
//	}
//
//	createdRecord, err := client.CreateRecord(ctx, "posts", recordData)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created record with ID: %s", createdRecord["id"])
func (c *Client) CreateRecord(ctx context.Context, collection string, record Record, opts ...QueryOption) (Record, error) {
	options := &QueryOptions{}
	for _, opt := range opts {
		opt(options)
	}

	endpoint := fmt.Sprintf("/api/collections/%s/records", collection)

	// Build query parameters
	params := url.Values{}
	if len(options.Expand) > 0 {
		params.Set("expand", strings.Join(options.Expand, ","))
	}
	if len(options.Fields) > 0 {
		params.Set("fields", strings.Join(options.Fields, ","))
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	var createdRecord Record
	err := c.doRequest(ctx, "POST", endpoint, record, &createdRecord)
	if err != nil {
		return nil, err
	}

	return createdRecord, nil
}

// doRequest is a helper method that handles HTTP requests to the PocketBase API.
// It manages request construction, authentication headers, JSON encoding/decoding,
// and error handling.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body any, out any) error {
	url := c.BaseURL + endpoint

	var reqBody []byte
	var err error

	// Encode request body as JSON if provided
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Add authorization header if token is available
	if token := c.GetToken(); token != "" {
		req.Header.Set("Authorization", token)
	}

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr apiErrorResp
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			// If we can't decode the error response, create a generic API error
			return &APIError{
				Status:  resp.StatusCode,
				Message: resp.Status,
				Data:    nil,
			}
		}
		return &APIError{
			Status:  apiErr.Status,
			Message: apiErr.Message,
			Data:    apiErr.Data,
		}
	}

	// Decode successful response
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
