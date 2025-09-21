package pocketbase

import "fmt"

// APIError represents an error response from the PocketBase API.
// It implements the error interface and provides structured error information.
type APIError struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

// Error returns a formatted error string implementing the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("pocketbase API error: %d %s", e.Status, e.Message)
}

// IsNotFound returns true if this is a 404 Not Found error.
func (e *APIError) IsNotFound() bool {
	return e.Status == 404
}

// IsUnauthorized returns true if this is a 401 Unauthorized error.
func (e *APIError) IsUnauthorized() bool {
	return e.Status == 401
}

// IsForbidden returns true if this is a 403 Forbidden error.
func (e *APIError) IsForbidden() bool {
	return e.Status == 403
}

// IsBadRequest returns true if this is a 400 Bad Request error.
func (e *APIError) IsBadRequest() bool {
	return e.Status == 400
}
