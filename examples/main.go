package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0x113/pocketbase-go"
)

func main() {
	// Create a context with timeout for all operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a new PocketBase client
	client := pocketbase.NewClient(
		"http://localhost:8090", // Replace with your PocketBase URL
		pocketbase.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		pocketbase.WithUserAgent("PocketBase-Go-Example/1.0"),
	)

	// Example 1: Authentication
	fmt.Println("=== Authentication Example ===")

	// Authenticate with email and password
	// Replace with actual credentials from your PocketBase instance
	user, err := client.AuthenticateWithPassword(ctx, "users", "alice@example.com", "password123")
	if err != nil {
		// Handle authentication error
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Authentication failed: %s (Status: %d)", apiErr.Message, apiErr.Status)
			if apiErr.IsBadRequest() {
				log.Println("Please check your credentials")
			}
		} else {
			log.Printf("Network error during authentication: %v", err)
		}

		// For demo purposes, continue with a mock token
		// In a real application, you would handle this error appropriately
		client.SetToken("demo-token-123")
		fmt.Println("Using demo token for remaining examples...")
	} else {
		fmt.Printf("Authentication successful!\n")
		fmt.Printf("User ID: %v\n", user["id"])
		fmt.Printf("User Email: %v\n", user["email"])
		fmt.Printf("Auth Token: %s\n", client.GetToken())
	}

	fmt.Println()

	// Example 2: Fetch all records from a collection
	fmt.Println("=== Fetch All Records Example ===")

	// Fetch all records from the "posts" collection
	// Replace "posts" with an actual collection name from your PocketBase
	records, err := client.GetAllRecords(ctx, "posts")
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to fetch records: %s (Status: %d)", apiErr.Message, apiErr.Status)
			if apiErr.IsUnauthorized() {
				log.Println("Authentication required or token expired")
			} else if apiErr.IsForbidden() {
				log.Println("Access denied to this collection")
			} else if apiErr.IsNotFound() {
				log.Println("Collection not found")
			}
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("Found %d records in 'posts' collection\n", len(records))

		// Display first few records
		for i, record := range records {
			if i >= 3 { // Show only first 3 records
				fmt.Printf("... and %d more records\n", len(records)-3)
				break
			}
			fmt.Printf("Record %d: ID=%v, Title=%v\n", i+1, record["id"], record["title"])
		}
	}

	fmt.Println()

	// Example 3: Fetch all records with filtering and sorting
	fmt.Println("=== Fetch Records with Options Example ===")

	filteredRecords, err := client.GetAllRecords(ctx, "posts",
		pocketbase.WithSort("-created"),             // Sort by creation date (newest first)
		pocketbase.WithFilter("status='published'"), // Only published posts
		pocketbase.WithListExpand("author"),         // Expand author relation
		pocketbase.WithPerPage(5),                   // Limit to 5 records per page
	)
	if err != nil {
		log.Printf("Failed to fetch filtered records: %v", err)
	} else {
		fmt.Printf("Found %d published records\n", len(filteredRecords))
	}

	fmt.Println()

	// Example 4: Fetch a single record
	fmt.Println("=== Fetch Single Record Example ===")

	// For demo purposes, try to get the first record from previous results
	if len(records) > 0 {
		recordID := fmt.Sprintf("%v", records[0]["id"])

		singleRecord, err := client.GetRecord(ctx, "posts", recordID,
			pocketbase.WithExpand("author", "category"),                           // Expand relations
			pocketbase.WithFields("id", "title", "content", "author", "category"), // Limit fields
		)
		if err != nil {
			if apiErr, ok := err.(*pocketbase.APIError); ok {
				log.Printf("Failed to fetch record: %s (Status: %d)", apiErr.Message, apiErr.Status)
				if apiErr.IsNotFound() {
					log.Printf("Record with ID '%s' not found", recordID)
				}
			} else {
				log.Printf("Network error: %v", err)
			}
		} else {
			fmt.Printf("Fetched record: %s\n", recordID)
			fmt.Printf("Title: %v\n", singleRecord["title"])
			fmt.Printf("Content preview: %.100v...\n", singleRecord["content"])
			if author, ok := singleRecord["author"]; ok {
				fmt.Printf("Author: %v\n", author)
			}
		}
	} else {
		fmt.Println("No records available to demonstrate single record fetch")
	}

	fmt.Println()

	// Example 5: Error handling demonstration
	fmt.Println("=== Error Handling Example ===")

	// Try to fetch a non-existent record to demonstrate error handling
	_, err = client.GetRecord(ctx, "posts", "non-existent-id-12345")
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			fmt.Printf("Properly handled API error: %s\n", apiErr.Error())
			fmt.Printf("Error Status: %d\n", apiErr.Status)
			fmt.Printf("Error Message: %s\n", apiErr.Message)

			// Demonstrate error type checking
			if apiErr.IsNotFound() {
				fmt.Println("Correctly identified as 404 Not Found error")
			}
		} else {
			fmt.Printf("Network or other error: %v\n", err)
		}
	}

	fmt.Println()

	// Example 6: Working with different collections
	fmt.Println("=== Multiple Collections Example ===")

	collections := []string{"users", "posts", "categories"}
	for _, collection := range collections {
		records, err := client.GetAllRecords(ctx, collection)
		if err != nil {
			fmt.Printf("%s: %v\n", collection, err)
		} else {
			fmt.Printf("%s: %d records\n", collection, len(records))
		}
	}

	fmt.Println()

	// Example 7: Superuser Authentication and Impersonation
	fmt.Println("=== Superuser & Impersonation Example ===")

	// Create a separate client for superuser operations
	superuserClient := pocketbase.NewClient(
		"http://localhost:8090", // Replace with your PocketBase URL
		pocketbase.WithTimeout(10*time.Second),
	)

	// Authenticate as superuser
	// Replace with actual superuser credentials from your PocketBase instance
	superuser, err := superuserClient.AuthenticateAsSuperuser(ctx, "admin@example.com", "admin_password")
	if err != nil {
		fmt.Printf("Superuser authentication failed: %v\n", err)
		fmt.Println("This is expected if you don't have a superuser configured")
	} else {
		fmt.Printf("Authenticated as superuser: %v\n", superuser["email"])

		// Try to impersonate a regular user (requires a valid user record ID)
		// This would typically come from a previous query or known user ID
		if len(records) > 0 {
			userID := fmt.Sprintf("%v", records[0]["id"])

			// Impersonate user for 30 minutes (1800 seconds)
			impersonateResult, err := superuserClient.Impersonate(ctx, "users", userID, 1800,
				pocketbase.WithExpand("profile"),
				pocketbase.WithFields("id", "email", "username", "profile"))

			if err != nil {
				fmt.Printf("Impersonation failed: %v\n", err)
			} else {
				fmt.Printf("Successfully impersonated user: %v\n", impersonateResult.Record["email"])
				fmt.Printf("Impersonation token: %.50s...\n", impersonateResult.Token)

				// Create a new client with the impersonation token
				impersonatedClient := pocketbase.NewClient("http://localhost:8090")
				impersonatedClient.SetToken(impersonateResult.Token)

				// Now make requests as the impersonated user
				userRecords, err := impersonatedClient.GetAllRecords(ctx, "user_data")
				if err != nil {
					fmt.Printf("Request as impersonated user failed: %v\n", err)
				} else {
					fmt.Printf("Fetched %d records as impersonated user\n", len(userRecords))
				}
			}
		} else {
			fmt.Println("No user records available to demonstrate impersonation")
		}
	}

	fmt.Println()
	fmt.Println("=== Example Complete ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("- Authentication with username/email and password")
	fmt.Println("- Superuser authentication")
	fmt.Println("- User impersonation (superuser only)")
	fmt.Println("- Fetching all records from a collection (with pagination)")
	fmt.Println("- Fetching records with filtering, sorting, and expansion")
	fmt.Println("- Fetching a single record by ID")
	fmt.Println("- Proper error handling with typed API errors")
	fmt.Println("- Working with multiple collections")
	fmt.Println("- Using impersonation tokens for restricted operations")
}
