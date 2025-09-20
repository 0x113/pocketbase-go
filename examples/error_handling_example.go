package main

import (
	"context"
	"fmt"
	"time"

	"github.com/0x113/pocketbase-go"
)

// ErrorHandlingExample demonstrates proper error handling with PocketBase API errors
func ErrorHandlingExample() {
	fmt.Println("=== Error Handling Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with demo token
	client := SetupDemoClient("http://localhost:8090")

	// Try to fetch a non-existent record to demonstrate error handling
	_, err := client.GetRecord(ctx, "posts", "non-existent-id-12345")
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			fmt.Printf("Properly handled API error: %s\n", apiErr.Error())
			fmt.Printf("Error Status: %d\n", apiErr.Status)
			fmt.Printf("Error Message: %s\n", apiErr.Message)

			// Demonstrate error type checking
			if apiErr.IsNotFound() {
				fmt.Println("[OK] Correctly identified as 404 Not Found error")
			}
			if apiErr.IsUnauthorized() {
				fmt.Println("[OK] Correctly identified as 401 Unauthorized error")
			}
			if apiErr.IsForbidden() {
				fmt.Println("[OK] Correctly identified as 403 Forbidden error")
			}
			if apiErr.IsBadRequest() {
				fmt.Println("[OK] Correctly identified as 400 Bad Request error")
			}
		} else {
			fmt.Printf("Network or other error: %v\n", err)
		}
	}

	fmt.Println()
}
