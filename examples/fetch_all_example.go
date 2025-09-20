package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x113/pocketbase-go"
)

// FetchAllExample demonstrates fetching all records from a collection
func FetchAllExample() {
	fmt.Println("=== Fetch All Records Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with demo token
	client := SetupDemoClient("http://localhost:8090")

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
}
