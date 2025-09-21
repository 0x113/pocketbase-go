package pocketbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with default options", func(t *testing.T) {
		client := NewClient("http://localhost:8090")

		if client.BaseURL != "http://localhost:8090" {
			t.Errorf("Expected BaseURL to be 'http://localhost:8090', got '%s'", client.BaseURL)
		}
		if client.HTTPClient == nil {
			t.Error("Expected HTTPClient to be set")
		}
		if client.userAgent != "pocketbase-go/1.0" {
			t.Errorf("Expected userAgent to be 'pocketbase-go/1.0', got '%s'", client.userAgent)
		}
	})

	t.Run("creates client with custom options", func(t *testing.T) {
		httpClient := &http.Client{Timeout: 5 * time.Second}
		client := NewClient("http://localhost:8090",
			WithHTTPClient(httpClient),
			WithUserAgent("TestClient/1.0"))

		if client.HTTPClient != httpClient {
			t.Error("Expected custom HTTPClient to be set")
		}
		if client.userAgent != "TestClient/1.0" {
			t.Errorf("Expected userAgent to be 'TestClient/1.0', got '%s'", client.userAgent)
		}
	})

	t.Run("trims trailing slash from baseURL", func(t *testing.T) {
		client := NewClient("http://localhost:8090/")

		if client.BaseURL != "http://localhost:8090" {
			t.Errorf("Expected BaseURL to be 'http://localhost:8090', got '%s'", client.BaseURL)
		}
	})
}

func TestClient_SetToken(t *testing.T) {
	client := NewClient("http://localhost:8090")
	token := "test-token-123"

	client.SetToken(token)

	if client.GetToken() != token {
		t.Errorf("Expected token to be '%s', got '%s'", token, client.GetToken())
	}
}

func TestClient_AuthenticateWithPassword_Success(t *testing.T) {
	// Mock server that returns successful authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		expectedPath := "/api/collections/users/auth-with-password"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json'")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept header to be 'application/json'")
		}

		// Parse and verify request body
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if body["identity"] != "alice@example.com" {
			t.Errorf("Expected identity 'alice@example.com', got '%s'", body["identity"])
		}
		if body["password"] != "password123" {
			t.Errorf("Expected password 'password123', got '%s'", body["password"])
		}

		// Send successful response
		response := authResp{
			Token: "auth-token-12345",
			Record: Record{
				"id":    "user-id-123",
				"email": "alice@example.com",
				"name":  "Alice Johnson",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	record, err := client.AuthenticateWithPassword(context.Background(), "users", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify returned record
	if record["id"] != "user-id-123" {
		t.Errorf("Expected record ID 'user-id-123', got '%v'", record["id"])
	}
	if record["email"] != "alice@example.com" {
		t.Errorf("Expected record email 'alice@example.com', got '%v'", record["email"])
	}

	// Verify token was stored
	if client.GetToken() != "auth-token-12345" {
		t.Errorf("Expected stored token 'auth-token-12345', got '%s'", client.GetToken())
	}
}

func TestClient_AuthenticateWithPassword_Failure(t *testing.T) {
	// Mock server that returns authentication failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := apiErrorResp{
			Status:  400,
			Message: "Failed to authenticate.",
			Data: map[string]any{
				"identity": map[string]string{
					"code":    "validation_invalid_email",
					"message": "Must be a valid email address.",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.AuthenticateWithPassword(context.Background(), "users", "invalid-email", "password")

	// Verify error is APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Status != 400 {
		t.Errorf("Expected error status 400, got %d", apiErr.Status)
	}
	if apiErr.Message != "Failed to authenticate." {
		t.Errorf("Expected error message 'Failed to authenticate.', got '%s'", apiErr.Message)
	}
	if !apiErr.IsBadRequest() {
		t.Error("Expected IsBadRequest() to return true")
	}
}

func TestClient_AuthenticateAsSuperuser_Success(t *testing.T) {
	// Mock server that returns successful superuser authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		expectedPath := "/api/collections/_superusers/auth-with-password"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Parse and verify request body
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if body["identity"] != "admin@example.com" {
			t.Errorf("Expected identity 'admin@example.com', got '%s'", body["identity"])
		}

		// Send successful response
		response := authResp{
			Token: "superuser-token-12345",
			Record: Record{
				"id":    "superuser-id-123",
				"email": "admin@example.com",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	superuser, err := client.AuthenticateAsSuperuser(context.Background(), "admin@example.com", "superuser_password")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify returned record
	if superuser["id"] != "superuser-id-123" {
		t.Errorf("Expected superuser ID 'superuser-id-123', got '%v'", superuser["id"])
	}
	if superuser["email"] != "admin@example.com" {
		t.Errorf("Expected superuser email 'admin@example.com', got '%v'", superuser["email"])
	}

	// Verify token was stored
	if client.GetToken() != "superuser-token-12345" {
		t.Errorf("Expected stored token 'superuser-token-12345', got '%s'", client.GetToken())
	}
}

func TestClient_Impersonate_Success(t *testing.T) {
	// Mock server that returns successful impersonation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		expectedPath := "/api/collections/users/impersonate/user-id-456"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "superuser-token" {
			t.Errorf("Expected Authorization header 'superuser-token', got '%s'", authHeader)
		}

		// Parse request body to check duration
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if duration, ok := body["duration"]; ok {
			if duration != float64(3600) { // JSON unmarshals numbers as float64
				t.Errorf("Expected duration 3600, got %v", duration)
			}
		}

		// Send impersonation response
		response := impersonateResp{
			Token: "impersonate-token-789",
			Record: Record{
				"id":       "user-id-456",
				"email":    "user@example.com",
				"username": "testuser",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("superuser-token")

	result, err := client.Impersonate(context.Background(), "users", "user-id-456", 3600)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify impersonation result
	if result.Token != "impersonate-token-789" {
		t.Errorf("Expected token 'impersonate-token-789', got '%s'", result.Token)
	}
	if result.Record["id"] != "user-id-456" {
		t.Errorf("Expected record ID 'user-id-456', got '%v'", result.Record["id"])
	}
	if result.Record["email"] != "user@example.com" {
		t.Errorf("Expected record email 'user@example.com', got '%v'", result.Record["email"])
	}
}

func TestClient_Impersonate_Unauthorized(t *testing.T) {
	// Mock server that returns 403 for non-superuser
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		response := apiErrorResp{
			Status:  403,
			Message: "The authorized record model is not allowed to perform this action.",
			Data:    map[string]any{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("regular-user-token")

	_, err := client.Impersonate(context.Background(), "users", "user-id-456", 3600)

	// Verify error is APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Status != 403 {
		t.Errorf("Expected error status 403, got %d", apiErr.Status)
	}
	if !apiErr.IsForbidden() {
		t.Error("Expected IsForbidden() to return true")
	}
}

func TestClient_Impersonate_WithOptions(t *testing.T) {
	// Mock server that verifies query parameters
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		expand := r.URL.Query().Get("expand")
		if expand != "profile,settings" {
			t.Errorf("Expected expand parameter 'profile,settings', got '%s'", expand)
		}

		fields := r.URL.Query().Get("fields")
		if fields != "id,email,username" {
			t.Errorf("Expected fields parameter 'id,email,username', got '%s'", fields)
		}

		// Send impersonation response
		response := impersonateResp{
			Token: "impersonate-token-with-options",
			Record: Record{
				"id":       "user-id-789",
				"email":    "user@example.com",
				"username": "testuser",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("superuser-token")

	result, err := client.Impersonate(context.Background(), "users", "user-id-789", 0,
		WithExpand("profile", "settings"),
		WithFields("id", "email", "username"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Token != "impersonate-token-with-options" {
		t.Errorf("Expected token 'impersonate-token-with-options', got '%s'", result.Token)
	}
}

func TestClient_GetRecord_Success(t *testing.T) {
	// Mock server that returns a single record
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		expectedPath := "/api/collections/posts/records/record-id-123"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "test-token" {
			t.Errorf("Expected Authorization header 'test-token', got '%s'", authHeader)
		}

		// Send record response
		record := Record{
			"id":      "record-id-123",
			"title":   "Test Post",
			"content": "This is a test post.",
			"author":  "user-id-456",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	record, err := client.GetRecord(context.Background(), "posts", "record-id-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if record["id"] != "record-id-123" {
		t.Errorf("Expected record ID 'record-id-123', got '%v'", record["id"])
	}
	if record["title"] != "Test Post" {
		t.Errorf("Expected record title 'Test Post', got '%v'", record["title"])
	}
}

func TestClient_GetRecord_NotFound(t *testing.T) {
	// Mock server that returns 404 not found
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		response := apiErrorResp{
			Status:  404,
			Message: "The requested resource wasn't found.",
			Data:    map[string]any{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	_, err := client.GetRecord(context.Background(), "posts", "nonexistent-id")

	// Verify error is APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Status != 404 {
		t.Errorf("Expected error status 404, got %d", apiErr.Status)
	}
	if !apiErr.IsNotFound() {
		t.Error("Expected IsNotFound() to return true")
	}
}

func TestClient_GetAllRecords_SinglePage(t *testing.T) {
	// Mock server that returns a single page of records
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		expectedPath := "/api/collections/posts/records"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Check query parameters
		page := r.URL.Query().Get("page")
		if page != "1" {
			t.Errorf("Expected page parameter '1', got '%s'", page)
		}

		// Send paginated response
		response := listResp{
			Page:       1,
			PerPage:    30,
			TotalItems: 2,
			TotalPages: 1,
			Items: []Record{
				{"id": "record-1", "title": "Post 1"},
				{"id": "record-2", "title": "Post 2"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	records, err := client.GetAllRecords(context.Background(), "posts")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
	if records[0]["id"] != "record-1" {
		t.Errorf("Expected first record ID 'record-1', got '%v'", records[0]["id"])
	}
}

func TestClient_GetAllRecords_MultiplePages(t *testing.T) {
	// Mock server that returns multiple pages
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		page := r.URL.Query().Get("page")

		var response listResp

		switch page {
		case "1":
			response = listResp{
				Page:       1,
				PerPage:    2,
				TotalItems: 3,
				TotalPages: 2,
				Items: []Record{
					{"id": "record-1", "title": "Post 1"},
					{"id": "record-2", "title": "Post 2"},
				},
			}
		case "2":
			response = listResp{
				Page:       2,
				PerPage:    2,
				TotalItems: 3,
				TotalPages: 2,
				Items: []Record{
					{"id": "record-3", "title": "Post 3"},
				},
			}
		default:
			t.Errorf("Unexpected page parameter: %s", page)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	records, err := client.GetAllRecords(context.Background(), "posts")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all records from both pages were retrieved
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
	if requestCount != 2 {
		t.Errorf("Expected 2 requests to be made, got %d", requestCount)
	}

	// Verify records are in correct order
	expectedIDs := []string{"record-1", "record-2", "record-3"}
	for i, expected := range expectedIDs {
		if records[i]["id"] != expected {
			t.Errorf("Expected record %d ID '%s', got '%v'", i, expected, records[i]["id"])
		}
	}
}

func TestClient_GetAllRecords_Error(t *testing.T) {
	// Mock server that returns 403 forbidden
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		response := apiErrorResp{
			Status:  403,
			Message: "You don't have access to this resource.",
			Data:    map[string]any{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("invalid-token")

	_, err := client.GetAllRecords(context.Background(), "posts")

	// Verify error is APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Status != 403 {
		t.Errorf("Expected error status 403, got %d", apiErr.Status)
	}
	if !apiErr.IsForbidden() {
		t.Error("Expected IsForbidden() to return true")
	}
}

func TestClient_GetRecord_WithOptions(t *testing.T) {
	// Mock server that verifies query parameters
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		expand := r.URL.Query().Get("expand")
		if expand != "author,category" {
			t.Errorf("Expected expand parameter 'author,category', got '%s'", expand)
		}

		fields := r.URL.Query().Get("fields")
		if fields != "id,title,content" {
			t.Errorf("Expected fields parameter 'id,title,content', got '%s'", fields)
		}

		// Send record response
		record := Record{
			"id":      "record-id-123",
			"title":   "Test Post",
			"content": "This is a test post.",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	record, err := client.GetRecord(context.Background(), "posts", "record-id-123",
		WithExpand("author", "category"),
		WithFields("id", "title", "content"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if record["title"] != "Test Post" {
		t.Errorf("Expected record title 'Test Post', got '%v'", record["title"])
	}
}

func TestAPIError_Methods(t *testing.T) {
	tests := []struct {
		status   int
		method   string
		expected bool
	}{
		{400, "IsBadRequest", true},
		{401, "IsUnauthorized", true},
		{403, "IsForbidden", true},
		{404, "IsNotFound", true},
		{500, "IsBadRequest", false},
		{200, "IsNotFound", false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%d", test.method, test.status), func(t *testing.T) {
			apiErr := &APIError{
				Status:  test.status,
				Message: "Test error",
				Data:    nil,
			}

			var result bool
			switch test.method {
			case "IsBadRequest":
				result = apiErr.IsBadRequest()
			case "IsUnauthorized":
				result = apiErr.IsUnauthorized()
			case "IsForbidden":
				result = apiErr.IsForbidden()
			case "IsNotFound":
				result = apiErr.IsNotFound()
			}

			if result != test.expected {
				t.Errorf("Expected %s() to return %v, got %v", test.method, test.expected, result)
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	apiErr := &APIError{
		Status:  404,
		Message: "Not found",
		Data:    nil,
	}

	expected := "pocketbase API error: 404 Not found"
	if apiErr.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, apiErr.Error())
	}
}

func TestClient_doRequest_InvalidJSON(t *testing.T) {
	// Mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		// Send invalid JSON to test error handling
		w.Write([]byte("invalid json response"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GetRecord(context.Background(), "posts", "test-id")

	// Should still return APIError even with invalid JSON
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Status != 500 {
		t.Errorf("Expected error status 500, got %d", apiErr.Status)
	}
}

func TestWithTimeout(t *testing.T) {
	timeout := 5 * time.Second
	client := NewClient("http://localhost:8090", WithTimeout(timeout))

	if client.HTTPClient.Timeout != timeout {
		t.Errorf("Expected HTTPClient timeout to be %v, got %v", timeout, client.HTTPClient.Timeout)
	}
}

func TestGetAllRecords_WithListOptions(t *testing.T) {
	// Mock server that verifies query parameters for list options
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		sort := r.URL.Query().Get("sort")
		if sort != "-created" {
			t.Errorf("Expected sort parameter '-created', got '%s'", sort)
		}

		filter := r.URL.Query().Get("filter")
		if filter != "status='published'" {
			t.Errorf("Expected filter parameter 'status='published'', got '%s'", filter)
		}

		expand := r.URL.Query().Get("expand")
		if expand != "author" {
			t.Errorf("Expected expand parameter 'author', got '%s'", expand)
		}

		perPage := r.URL.Query().Get("perPage")
		if perPage != "10" {
			t.Errorf("Expected perPage parameter '10', got '%s'", perPage)
		}

		// Send response
		response := listResp{
			Page:       1,
			PerPage:    10,
			TotalItems: 1,
			TotalPages: 1,
			Items: []Record{
				{"id": "record-1", "title": "Test Post", "status": "published"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	records, err := client.GetAllRecords(context.Background(), "posts",
		WithSort("-created"),
		WithFilter("status='published'"),
		WithListExpand("author"),
		WithPerPage(10))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}
}

func TestClient_CreateRecord_Success(t *testing.T) {
	// Mock server that accepts record creation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		expectedPath := "/api/collections/posts/records"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "test-token" {
			t.Errorf("Expected Authorization header 'test-token', got '%s'", authHeader)
		}

		// Parse and verify request body
		var record Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if record["title"] != "Test Post" {
			t.Errorf("Expected title 'Test Post', got '%v'", record["title"])
		}
		if record["content"] != "This is test content" {
			t.Errorf("Expected content 'This is test content', got '%v'", record["content"])
		}

		// Send created record response
		createdRecord := Record{
			"id":      "created-record-123",
			"title":   "Test Post",
			"content": "This is test content",
			"created": "2023-01-01T12:00:00Z",
			"updated": "2023-01-01T12:00:00Z",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(createdRecord)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	recordData := Record{
		"title":   "Test Post",
		"content": "This is test content",
	}

	createdRecord, err := client.CreateRecord(context.Background(), "posts", recordData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify created record
	if createdRecord["id"] != "created-record-123" {
		t.Errorf("Expected created record ID 'created-record-123', got '%v'", createdRecord["id"])
	}
	if createdRecord["title"] != "Test Post" {
		t.Errorf("Expected created record title 'Test Post', got '%v'", createdRecord["title"])
	}
	if createdRecord["content"] != "This is test content" {
		t.Errorf("Expected created record content 'This is test content', got '%v'", createdRecord["content"])
	}
	if createdRecord["created"] != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected created timestamp, got '%v'", createdRecord["created"])
	}
}

func TestClient_CreateRecord_ValidationError(t *testing.T) {
	// Mock server that returns validation error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := apiErrorResp{
			Status:  400,
			Message: "An error occurred while validating the submitted data.",
			Data: map[string]any{
				"title": map[string]any{
					"code":    "validation_required",
					"message": "Missing required value.",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	recordData := Record{
		"content": "Content without title",
	}

	_, err := client.CreateRecord(context.Background(), "posts", recordData)

	// Verify error is APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Status != 400 {
		t.Errorf("Expected error status 400, got %d", apiErr.Status)
	}
	if apiErr.Message != "An error occurred while validating the submitted data." {
		t.Errorf("Expected error message 'An error occurred while validating the submitted data.', got '%s'", apiErr.Message)
	}
	if !apiErr.IsBadRequest() {
		t.Error("Expected IsBadRequest() to return true")
	}
}

func TestClient_CreateRecord_WithOptions(t *testing.T) {
	// Mock server that verifies query parameters and returns created record
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		expand := r.URL.Query().Get("expand")
		if expand != "author,category" {
			t.Errorf("Expected expand parameter 'author,category', got '%s'", expand)
		}

		fields := r.URL.Query().Get("fields")
		if fields != "id,title,content,author" {
			t.Errorf("Expected fields parameter 'id,title,content,author', got '%s'", fields)
		}

		// Parse request body to verify record data
		var record Record
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send created record response with expanded relations
		createdRecord := Record{
			"id":      "created-with-options-456",
			"title":   "Post with Options",
			"content": "Content with options",
			"expand": Record{
				"author": Record{
					"id":   "author-123",
					"name": "John Doe",
				},
				"category": Record{
					"id":   "category-456",
					"name": "Technology",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(createdRecord)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")

	recordData := Record{
		"title":   "Post with Options",
		"content": "Content with options",
	}

	createdRecord, err := client.CreateRecord(context.Background(), "posts", recordData,
		WithExpand("author", "category"),
		WithFields("id", "title", "content", "author"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if createdRecord["id"] != "created-with-options-456" {
		t.Errorf("Expected created record ID 'created-with-options-456', got '%v'", createdRecord["id"])
	}

	// Verify expanded relations are included
	if expandData, ok := createdRecord["expand"]; ok {
		expandMap, ok := expandData.(map[string]any)
		if !ok {
			t.Error("Expected expand data to be a map")
		} else {
			if author, ok := expandMap["author"]; ok {
				authorMap, ok := author.(map[string]any)
				if !ok {
					t.Error("Expected author data to be a map")
				} else {
					if authorMap["name"] != "John Doe" {
						t.Errorf("Expected expanded author name 'John Doe', got '%v'", authorMap["name"])
					}
				}
			} else {
				t.Error("Expected expanded author data to be present")
			}
		}
	} else {
		t.Error("Expected expand data to be present")
	}
}

func TestClient_CreateRecord_Unauthorized(t *testing.T) {
	// Mock server that returns 401 unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		response := apiErrorResp{
			Status:  401,
			Message: "The request requires valid record authorization token to be set.",
			Data:    map[string]any{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	// Note: No token set for this test

	recordData := Record{
		"title": "Unauthorized Post",
	}

	_, err := client.CreateRecord(context.Background(), "posts", recordData)

	// Verify error is APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Status != 401 {
		t.Errorf("Expected error status 401, got %d", apiErr.Status)
	}
	if !apiErr.IsUnauthorized() {
		t.Error("Expected IsUnauthorized() to return true")
	}
}
