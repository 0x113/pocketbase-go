package pocketbase

import "io"

// Record represents a generic PocketBase record as a map of field names to values.
// This flexible structure allows handling different collection schemas dynamically.
type Record map[string]any

// FileData represents a file to be uploaded with optional metadata
type FileData struct {
	Reader   io.Reader
	Filename string
	Size     int64
}

// FileUpload represents file upload configuration for a field
type FileUpload struct {
	Field  string
	Files  []FileData
	Append bool     // If true, appends to existing files (fieldname+)
	Delete []string // Filenames to delete (used with fieldname-)
}

// FileUploadOption represents functional options for file upload operations.
type FileUploadOption func(*FileUploadOptions)

// FileUploadOptions holds options for file upload requests.
type FileUploadOptions struct {
	Uploads []FileUpload
	Data    Record // Regular form data to include with the upload
	QueryOptions
}

// WithFileUpload adds a file upload configuration to the request.
func WithFileUpload(field string, files []FileData, options ...FileUploadModifier) FileUploadOption {
	return func(opts *FileUploadOptions) {
		upload := FileUpload{
			Field: field,
			Files: files,
		}

		for _, option := range options {
			option(&upload)
		}

		opts.Uploads = append(opts.Uploads, upload)
	}
}

// WithFormData adds regular form data to include with file uploads.
func WithFormData(data Record) FileUploadOption {
	return func(opts *FileUploadOptions) {
		opts.Data = data
	}
}

// FileUploadModifier represents functional options for individual file uploads.
type FileUploadModifier func(*FileUpload)

// WithAppend sets the file upload to append mode (fieldname+).
func WithAppend() FileUploadModifier {
	return func(upload *FileUpload) {
		upload.Append = true
	}
}

// WithDelete sets filenames to delete from the field (fieldname-).
func WithDelete(filenames ...string) FileUploadModifier {
	return func(upload *FileUpload) {
		upload.Delete = filenames
	}
}

// authResp represents the response structure from the auth-with-password endpoint.
type authResp struct {
	Token  string `json:"token"`
	Record Record `json:"record"`
}

// listResp represents the paginated response structure from the list records endpoint.
type listResp struct {
	Page       int      `json:"page"`
	PerPage    int      `json:"perPage"`
	TotalItems int      `json:"totalItems"`
	TotalPages int      `json:"totalPages"`
	Items      []Record `json:"items"`
}

// apiErrorResp represents the error response structure from PocketBase API.
type apiErrorResp struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

// impersonateResp represents the response structure from the impersonate endpoint.
type impersonateResp struct {
	Token  string `json:"token"`
	Record Record `json:"record"`
}

// ImpersonateResult contains the result of an impersonation request.
type ImpersonateResult struct {
	Token  string
	Record Record
}

// QueryOption represents functional options for single record queries.
type QueryOption func(*QueryOptions)

// QueryOptions holds query parameters for single record requests.
type QueryOptions struct {
	Expand []string
	Fields []string
}

// ListOption represents functional options for list queries.
type ListOption func(*ListOptions)

// ListOptions holds query parameters for list requests.
type ListOptions struct {
	Page    int
	PerPage int
	Sort    string
	Filter  string
	Expand  []string
	Fields  []string
}

// WithExpand adds expand fields to query options.
func WithExpand(fields ...string) QueryOption {
	return func(opts *QueryOptions) {
		opts.Expand = fields
	}
}

// WithFields adds specific fields to query options.
func WithFields(fields ...string) QueryOption {
	return func(opts *QueryOptions) {
		opts.Fields = fields
	}
}

// WithSort adds sorting to list options.
func WithSort(sort string) ListOption {
	return func(opts *ListOptions) {
		opts.Sort = sort
	}
}

// WithFilter adds filtering to list options.
func WithFilter(filter string) ListOption {
	return func(opts *ListOptions) {
		opts.Filter = filter
	}
}

// WithListExpand adds expand fields to list options.
func WithListExpand(fields ...string) ListOption {
	return func(opts *ListOptions) {
		opts.Expand = fields
	}
}

// WithListFields adds specific fields to list options.
func WithListFields(fields ...string) ListOption {
	return func(opts *ListOptions) {
		opts.Fields = fields
	}
}

// WithPage sets the page number for list options.
func WithPage(page int) ListOption {
	return func(opts *ListOptions) {
		opts.Page = page
	}
}

// WithPerPage sets the per page limit for list options.
func WithPerPage(perPage int) ListOption {
	return func(opts *ListOptions) {
		opts.PerPage = perPage
	}
}
