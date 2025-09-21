package pocketbase

// Record represents a generic PocketBase record as a map of field names to values.
// This flexible structure allows handling different collection schemas dynamically.
type Record map[string]any

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
