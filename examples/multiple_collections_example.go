package main

import (
	"context"
	"fmt"
	"time"
)

// MultipleCollectionsExample demonstrates working with different collections
func MultipleCollectionsExample() {
	fmt.Println("=== Multiple Collections Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with demo token
	client := SetupDemoClient("http://localhost:8090")

	// List of collections to work with
	collections := []string{"users", "posts", "categories"}

	for _, collection := range collections {
		fmt.Printf("Working with collection: %s\n", collection)

		records, err := client.GetAllRecords(ctx, collection)
		if err != nil {
			fmt.Printf("  [ERROR] %s: %v\n", collection, err)
		} else {
			fmt.Printf("  [SUCCESS] %s: %d records found\n", collection, len(records))

			// Show a sample of the records
			if len(records) > 0 {
				fmt.Printf("  Sample record: %v\n", records[0])
			}
		}
		fmt.Println()
	}

	fmt.Println()
}
