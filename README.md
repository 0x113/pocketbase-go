# PocketBase Go Client

A simple Go client for [PocketBase](https://pocketbase.io/) that handles the common stuff you need - authentication, fetching records, and working with collections.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-007d9c)](https://golang.org/)
[![PocketBase](https://img.shields.io/badge/PocketBase-v0.20+-00d4aa)](https://pocketbase.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## What it does

- User and superuser authentication
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

### Records and errors

Records are returned as `map[string]interface{}`, so you can access any field:

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

Run the tests:

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

## Examples

The [examples](examples/) directory contains well-documented code examples that demonstrate different features:

- `common.go` - Shared utilities and client setup
- `auth_example.go` - User authentication
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

This covers the basic read operations. Future versions might add:

- Creating, updating, and deleting records
- File uploads
- Real-time subscriptions  
- Admin API
- OAuth2 login

## Contributing

Pull requests welcome! This is a simple library so let's keep it that way.

## License

MIT - see [LICENSE](LICENSE) file.
