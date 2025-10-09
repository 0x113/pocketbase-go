# PocketBase Go Client

A simple Go client for [PocketBase](https://pocketbase.io/) that handles the common stuff you need - authentication, fetching records, and working with collections.

[![Go Reference](https://pkg.go.dev/badge/github.com/0x113/pocketbase-go.svg)](https://pkg.go.dev/github.com/0x113/pocketbase-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/0x113/pocketbase-go)](https://goreportcard.com/report/github.com/0x113/pocketbase-go)
[![PocketBase](https://img.shields.io/badge/PocketBase-v0.30+-00d4aa)](https://pocketbase.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## What it does

- User and superuser authentication
- Create new records in collections
- Update existing records in collections
- File uploads with records (single and multiple files)
- Fetch records from collections (with automatic pagination)
- Query single records by ID
- User impersonation for superusers
- Filtering, sorting, and expanding relations
- No external dependencies - just the Go standard library
- Thread-safe token management
- Proper error handling

## Installation

```bash
go get github.com/0x113/pocketbase-go
```

## Getting started

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/0x113/pocketbase-go"
)

func main() {
    client := pocketbase.NewClient("http://localhost:8090")

    // Login
    user, err := client.AuthenticateWithPassword(
        context.Background(),
        "users", 
        "user@example.com", 
        "password123",
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Logged in as: %s\n", user["email"])

    // Get all posts
    posts, err := client.GetAllRecords(context.Background(), "posts")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d posts\n", len(posts))

    // Get one post
    if len(posts) > 0 {
        post, err := client.GetRecord(
            context.Background(),
            "posts", 
            fmt.Sprintf("%v", posts[0]["id"]),
        )
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Post title: %s\n", post["title"])
    }
}
```

## API Reference

### Creating a client

```go
client := pocketbase.NewClient("http://localhost:8090")
```

You can pass options to customize the client:

```go
client := pocketbase.NewClient("http://localhost:8090",
    pocketbase.WithTimeout(30*time.Second),
    pocketbase.WithUserAgent("MyApp/1.0"),
)
```

Available options:
- `WithHTTPClient(client *http.Client)` - Use your own HTTP client
- `WithTimeout(timeout time.Duration)` - Set request timeout
- `WithUserAgent(userAgent string)` - Custom User-Agent header

### Authentication

#### Regular users

```go
user, err := client.AuthenticateWithPassword(ctx, "users", "john@example.com", "secret123")
if err != nil {
    if apiErr, ok := err.(*pocketbase.APIError); ok {
        if apiErr.IsBadRequest() {
            fmt.Println("Wrong email or password")
        }
    }
    return err
}
fmt.Printf("Logged in as: %s\n", user["email"])
```

#### Superusers

```go
superuser, err := client.AuthenticateAsSuperuser(ctx, "admin@example.com", "admin_password")
if err != nil {
    log.Fatal("Failed to authenticate as superuser:", err)
}
fmt.Printf("Superuser: %s\n", superuser["email"])
```

#### User impersonation

Only superusers can impersonate other users. This generates a non-refreshable token for the target user:

```go
// First authenticate as superuser
_, err := client.AuthenticateAsSuperuser(ctx, "admin@example.com", "admin_password")
if err != nil {
    log.Fatal(err)
}

// Then impersonate a user for 1 hour
result, err := client.Impersonate(ctx, "users", "user_record_id", 3600)
if err != nil {
    log.Fatal("Impersonation failed:", err)
}

// Use the impersonation token
impersonatedClient := pocketbase.NewClient("http://localhost:8090")
impersonatedClient.SetToken(result.Token)

// Now make requests as the impersonated user
records, err := impersonatedClient.GetAllRecords(ctx, "user_posts")
```

#### Working with tokens

You can set tokens manually if you have them from somewhere else:

```go
client.SetToken("your-token-here")
token := client.GetToken() // Get current token
```

### Working with records

#### Get all records from a collection

```go
posts, err := client.GetAllRecords(ctx, "posts")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d posts\n", len(posts))
```

The client automatically handles pagination for you. You can also add filters and sorting:

```go
posts, err := client.GetAllRecords(ctx, "posts",
    pocketbase.WithFilter("status='published'"),
    pocketbase.WithSort("-created"),
    pocketbase.WithListExpand("author", "category"),
    pocketbase.WithPerPage(50),
)
```

Available options for `GetAllRecords`:
- `WithSort(sort string)` - Sort records (e.g., "-created", "+title")
- `WithFilter(filter string)` - Filter records (e.g., "status='published'")
- `WithListExpand(fields ...string)` - Expand relation fields
- `WithListFields(fields ...string)` - Select specific fields only
- `WithPerPage(perPage int)` - Records per page
- `WithPage(page int)` - Get specific page only

#### Get a single record

```go
post, err := client.GetRecord(ctx, "posts", "RECORD_ID_HERE")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Post title: %s\n", post["title"])
```

You can also expand relations and select specific fields:

```go
post, err := client.GetRecord(ctx, "posts", "RECORD_ID_HERE",
    pocketbase.WithExpand("author", "comments"),
    pocketbase.WithFields("id", "title", "content", "author"),
)
```

#### Create a new record

```go
// Create a new record in a collection
recordData := pocketbase.Record{
    "title":   "My New Post",
    "content": "This is the content of my new post",
    "status":  "published",
    "tags":    []string{"golang", "tutorial"},
}

createdRecord, err := client.CreateRecord(ctx, "posts", recordData)
if err != nil {
    if apiErr, ok := err.(*pocketbase.APIError); ok {
        if apiErr.IsBadRequest() {
            fmt.Println("Validation error:", apiErr.Message)
        }
    }
    log.Fatal(err)
}
fmt.Printf("Created record with ID: %s\n", createdRecord["id"])
```

You can also expand relations and select specific fields when creating:

```go
createdRecord, err := client.CreateRecord(ctx, "posts", recordData,
    pocketbase.WithExpand("author", "category"),
    pocketbase.WithFields("id", "title", "content", "author"),
)
```

#### Update an existing record

```go
// Update a record by providing only the fields you want to change
updateData := pocketbase.Record{
    "title":   "Updated Post Title",
    "content": "This post has been updated with new content",
    "status":  "published",
    "tags":    []string{"golang", "tutorial", "updated"},
}

updatedRecord, err := client.UpdateRecord(ctx, "posts", "RECORD_ID_HERE", updateData)
if err != nil {
    if apiErr, ok := err.(*pocketbase.APIError); ok {
        if apiErr.IsBadRequest() {
            fmt.Println("Validation error:", apiErr.Message)
        }
    }
    log.Fatal(err)
}
fmt.Printf("Updated record: %s\n", updatedRecord["title"])
```

You can also expand relations and select specific fields when updating:

```go
updatedRecord, err := client.UpdateRecord(ctx, "posts", "RECORD_ID_HERE", updateData,
    pocketbase.WithExpand("author", "category"),
    pocketbase.WithFields("id", "title", "content", "author"),
)
```

### File uploads

The library supports uploading files to PocketBase collections with file fields.

#### Create a record with file uploads

```go
// Open files
file1, err := os.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file1.Close()

file2, err := os.Open("image.jpg")
if err != nil {
    log.Fatal(err)
}
defer file2.Close()

// Prepare files for upload
files := []pocketbase.FileData{
    {Reader: file1, Filename: "document.pdf"},
    {Reader: file2, Filename: "image.jpg"},
}

// Prepare record data
data := pocketbase.Record{
    "title":       "Important Document",
    "description": "This document contains important information",
}

// Create record with files
createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
    pocketbase.WithFormData(data),
    pocketbase.WithFileUpload("files", files))

if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created record with files: %s\n", createdRecord["id"])
```

#### Update a record with file uploads

Replace existing files:

```go
newFile, err := os.Open("new-avatar.jpg")
if err != nil {
    log.Fatal(err)
}
defer newFile.Close()

files := []pocketbase.FileData{
    {Reader: newFile, Filename: "new-avatar.jpg"},
}

data := pocketbase.Record{
    "name": "Updated User",
}

updatedRecord, err := client.UpdateRecordWithFiles(ctx, "users", "RECORD_ID",
    pocketbase.WithFormData(data),
    pocketbase.WithFileUpload("avatar", files))
```

Append files to existing ones:

```go
newFile, err := os.Open("document3.pdf")
if err != nil {
    log.Fatal(err)
}
defer newFile.Close()

files := []pocketbase.FileData{
    {Reader: newFile, Filename: "document3.pdf"},
}

updatedRecord, err := client.UpdateRecordWithFiles(ctx, "documents", "RECORD_ID",
    pocketbase.WithFileUpload("files", files, pocketbase.WithAppend()))
```

Delete specific files:

```go
updatedRecord, err := client.UpdateRecordWithFiles(ctx, "documents", "RECORD_ID",
    pocketbase.WithFileUpload("files", nil, 
        pocketbase.WithDelete("old-file1.pdf", "old-file2.pdf")))
```

#### File upload helper functions

The library provides several helper functions to create `FileData`:

```go
// From an io.Reader
fileData := pocketbase.CreateFileData(reader, "filename.txt")

// From byte data
content := []byte("Hello, World!")
fileData := pocketbase.CreateFileDataFromBytes(content, "hello.txt")

// From file path (caller must close the file)
fileData, err := pocketbase.CreateFileDataFromFile("path/to/file.pdf")
if err != nil {
    log.Fatal(err)
}
// Don't forget to close the file when done
if fileReader, ok := fileData.Reader.(*os.File); ok {
    defer fileReader.Close()
}

// Use in upload
createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
    pocketbase.WithFormData(data),
    pocketbase.WithFileUpload("file", []pocketbase.FileData{fileData}))
```

#### File upload with query options

You can use expand and fields options with file uploads:

```go
// Upload with expanded relations
createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
    pocketbase.WithFormData(data),
    pocketbase.WithFileUpload("files", files),
    func(opts *pocketbase.FileUploadOptions) {
        opts.Expand = []string{"author", "category"}
        opts.Fields = []string{"id", "title", "files", "author"}
    })
```

### Records and errors

Records are returned as `map[string]any`, so you can access any field:

```go
fmt.Printf("Title: %s\n", record["title"])
fmt.Printf("Created: %s\n", record["created"])

// Type assertion for specific types
if id, ok := record["id"].(string); ok {
    fmt.Printf("Record ID: %s\n", id)
}
```

API errors are returned as `*pocketbase.APIError` with useful methods:

```go
record, err := client.GetRecord(ctx, "posts", "invalid-id")
if err != nil {
    if apiErr, ok := err.(*pocketbase.APIError); ok {
        fmt.Printf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.Status)
        
        if apiErr.IsNotFound() {
            fmt.Println("Record not found")
        } else if apiErr.IsUnauthorized() {
            fmt.Println("Need to login")
        }
    } else {
        fmt.Printf("Network error: %v\n", err)
    }
}
```

Available error check methods:
- `IsNotFound()` - 404 errors
- `IsUnauthorized()` - 401 errors  
- `IsForbidden()` - 403 errors
- `IsBadRequest()` - 400 errors

## More examples

### Custom HTTP client

```go
import "crypto/tls"

httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true, // Only for development!
        },
    },
}

client := pocketbase.NewClient("https://your-pb-instance.com",
    pocketbase.WithHTTPClient(httpClient),
)
```

### Complex filtering

```go
records, err := client.GetAllRecords(ctx, "posts",
    pocketbase.WithFilter("(status='published' || status='featured') && author.verified=true"),
    pocketbase.WithSort("-featured, -created, +title"),
    pocketbase.WithListExpand("author", "tags", "category"),
)
```

### Pagination

```go
// Get specific page
page2, err := client.GetAllRecords(ctx, "posts",
    pocketbase.WithPage(2),
    pocketbase.WithPerPage(20),
)

// Or get everything (default - handles pagination automatically)
allPosts, err := client.GetAllRecords(ctx, "posts")
```

### Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

records, err := client.GetAllRecords(ctx, "large_collection")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Request timed out")
    }
}
```

## Testing

### Local Testing

Run the tests locally:

```bash
go test ./...
go test -cover ./...  # with coverage
go test -v ./...      # verbose
```

The tests use `httptest.Server` to mock PocketBase responses and cover:
- Authentication (regular users and superusers)
- Record fetching with pagination
- Impersonation functionality
- Error handling
- Query options
- Thread safety

### Continuous Integration

This project uses GitHub Actions for continuous integration:

#### Automated Testing
- **Triggers**: Runs on every push to `main` branch and every pull request targeting `main`
- **Go versions**: Tests against Go 1.21.x and 1.22.x
- **Coverage**: Generates and reports test coverage to Codecov
- **Quality checks**: Includes formatting validation and `go vet`

#### Manual Testing
- **Trigger**: Can be manually triggered via GitHub Actions interface
- **Custom Go version**: Allows specifying a custom Go version for testing
- **Purpose**: Useful for testing other branches or specific Go versions

The CI pipeline ensures code quality and compatibility across supported Go versions.

## Examples

The [examples](examples/) directory contains well-documented code examples that demonstrate different features:

- `common.go` - Shared utilities and client setup
- `auth_example.go` - User authentication
- `create_record_example.go` - **Creating new records** in collections
- `update_record_example.go` - **Updating existing records** in collections
- `file_upload_example.go` - **Uploading files** with records
- `fetch_all_example.go` - Fetching all records from collections
- `fetch_options_example.go` - Filtering, sorting, and expanding records
- `fetch_single_example.go` - Fetching individual records
- `error_handling_example.go` - Proper error handling
- `multiple_collections_example.go` - Working with different collections
- `superuser_example.go` - Superuser authentication and impersonation

### Learning from Examples

Each example file is self-contained and includes detailed comments explaining the functionality. To learn how to use the library:

1. **Read the example files** - Each file demonstrates a specific aspect of the PocketBase Go client
2. **Study the comments** - Detailed explanations are provided inline
3. **Understand the patterns** - See how to handle authentication, errors, and data fetching
4. **Adapt to your needs** - Use the patterns as templates for your own code

The examples show real-world usage patterns including proper error handling, context management, and best practices for working with PocketBase collections.

## Requirements

- Go 1.21+
- PocketBase 0.20+  
- No external dependencies

## What's missing

This covers the basic read and write operations. Future versions might add:

- Deleting records
- Real-time subscriptions
- Admin API
- OAuth2 login

## Contributing

Pull requests welcome! This is a simple library so let's keep it that way.

## License

MIT - see [LICENSE](LICENSE) file.
