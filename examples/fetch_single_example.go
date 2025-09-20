package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x113/pocketbase-go"
)

// FetchSingleExample demonstrates fetching a single record by ID with field selection and expansion
func FetchSingleExample() {
	fmt.Println("=== Fetch Single Record Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with demo token
	client := SetupDemoClient("http://localhost:8090")

	// First, get some records to find a valid ID
	records, err := client.GetAllRecords(ctx, "posts")
	if err != nil {
		log.Printf("Failed to fetch records for demo: %v", err)
		fmt.Println("No records available to demonstrate single record fetch")
		return
	}

	if len(records) == 0 {
		fmt.Println("No records available to demonstrate single record fetch")
		return
	}

	// Get the first record ID
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
		if category, ok := singleRecord["category"]; ok {
			fmt.Printf("Category: %v\n", category)
		}
	}

	fmt.Println()
}
