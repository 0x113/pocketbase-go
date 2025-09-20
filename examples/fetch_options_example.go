package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x113/pocketbase-go"
)

// FetchOptionsExample demonstrates fetching records with filtering, sorting, and expansion options
func FetchOptionsExample() {
	fmt.Println("=== Fetch Records with Options Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with demo token
	client := SetupDemoClient("http://localhost:8090")

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

		// Display the filtered records
		for i, record := range filteredRecords {
			fmt.Printf("Record %d: %v\n", i+1, record["title"])
			if author, ok := record["author"]; ok {
				fmt.Printf("  Author: %v\n", author)
			}
		}
	}

	fmt.Println()
}
