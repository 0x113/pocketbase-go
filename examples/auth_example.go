package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x113/pocketbase-go"
)

// AuthExample demonstrates user authentication with PocketBase
func AuthExample() {
	fmt.Println("=== Authentication Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client
	client := CreateClient("http://localhost:8090")

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
}
