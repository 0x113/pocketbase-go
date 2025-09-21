package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x113/pocketbase-go"
)

// UpdateRecordExample demonstrates updating existing records in PocketBase collections
func UpdateRecordExample() {
	fmt.Println("=== Update Record Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with demo token
	client := SetupDemoClient("http://localhost:8090")

	// Example 1: Basic record update
	fmt.Println("1. Updating a post record...")
	updateData := pocketbase.Record{
		"title":   "Updated Post Title",
		"content": "This post has been updated with new content via the Go client.",
		"status":  "published",
		"tags":    []string{"golang", "pocketbase", "tutorial", "updated"},
	}

	updatedPost, err := client.UpdateRecord(ctx, "posts", "EXISTING_POST_ID", updateData)
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to update post: %s (Status: %d)", apiErr.Message, apiErr.Status)
			if apiErr.IsBadRequest() {
				log.Printf("Validation error in the submitted data")
			}
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully updated post!\n")
		fmt.Printf("  ID: %s\n", updatedPost["id"])
		fmt.Printf("  Title: %s\n", updatedPost["title"])
		fmt.Printf("  Status: %s\n", updatedPost["status"])
		if updatedAt, ok := updatedPost["updated"]; ok {
			fmt.Printf("  Last Updated: %s\n", updatedAt)
		}
	}

	// Example 2: Partial update (only update specific fields)
	fmt.Println("\n2. Partially updating a user profile...")
	partialUpdateData := pocketbase.Record{
		"bio":      "Updated bio: Senior software developer with 10+ years of experience",
		"location": "Seattle, WA",
		"website":  "https://updated-profile.dev",
	}

	updatedProfile, err := client.UpdateRecord(ctx, "profiles", "EXISTING_PROFILE_ID", partialUpdateData)
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to update profile: %s (Status: %d)", apiErr.Message, apiErr.Status)
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully updated profile!\n")
		fmt.Printf("  ID: %s\n", updatedProfile["id"])
		fmt.Printf("  Bio: %s\n", updatedProfile["bio"])
		fmt.Printf("  Location: %s\n", updatedProfile["location"])
		fmt.Printf("  Website: %s\n", updatedProfile["website"])
	}

	// Example 3: Update with expanded relations
	fmt.Println("\n3. Updating a post with expanded author information...")
	postUpdateWithExpand := pocketbase.Record{
		"title":    "Advanced Go Patterns",
		"content":  "Exploring advanced Go programming patterns and best practices...",
		"readTime": 8,
		"featured": true,
	}

	updatedPostWithExpand, err := client.UpdateRecord(ctx, "posts", "EXISTING_POST_ID", postUpdateWithExpand,
		pocketbase.WithExpand("author", "category"),
		pocketbase.WithFields("id", "title", "content", "readTime", "featured", "author"))
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to update post with expand: %s (Status: %d)", apiErr.Message, apiErr.Status)
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully updated post with expanded relations!\n")
		fmt.Printf("  ID: %s\n", updatedPostWithExpand["id"])
		fmt.Printf("  Title: %s\n", updatedPostWithExpand["title"])
		fmt.Printf("  Featured: %t\n", updatedPostWithExpand["featured"])

		// Show expanded relations if present
		if expandData, ok := updatedPostWithExpand["expand"]; ok {
			if expandMap, ok := expandData.(pocketbase.Record); ok {
				if author, ok := expandMap["author"]; ok {
					fmt.Printf("  Author: %v\n", author)
				}
				if category, ok := expandMap["category"]; ok {
					fmt.Printf("  Category: %v\n", category)
				}
			}
		}
	}

	// Example 4: Bulk update simulation
	fmt.Println("\n4. Bulk update demonstration...")
	// In a real application, you might update multiple records based on some criteria
	// This example shows the pattern for multiple updates

	updates := []struct {
		id   string
		data pocketbase.Record
	}{
		{
			id: "POST_ID_1",
			data: pocketbase.Record{
				"status": "archived",
			},
		},
		{
			id: "POST_ID_2",
			data: pocketbase.Record{
				"status": "archived",
			},
		},
	}

	for i, update := range updates {
		fmt.Printf("  Updating record %d...\n", i+1)
		_, err := client.UpdateRecord(ctx, "posts", update.id, update.data)
		if err != nil {
			log.Printf("    Failed to update record %s: %v", update.id, err)
		} else {
			fmt.Printf("    ✓ Updated record %s successfully\n", update.id)
		}
	}

	// Example 5: Update with error handling
	fmt.Println("\n5. Update with comprehensive error handling...")
	errorUpdateData := pocketbase.Record{
		// Missing required fields - this will likely cause a validation error
		"content": "Content without title",
	}

	_, err = client.UpdateRecord(ctx, "posts", "EXISTING_POST_ID", errorUpdateData)
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			if apiErr.IsBadRequest() {
				fmt.Printf("  Validation failed as expected: %s\n", apiErr.Message)
				// In a real application, you would show these validation errors to the user
				// and allow them to correct the form data
				fmt.Printf("  Error details: %v\n", apiErr.Data)
			} else if apiErr.IsNotFound() {
				fmt.Printf("  Record not found: %s\n", apiErr.Message)
			} else if apiErr.IsUnauthorized() {
				fmt.Printf("  Unauthorized: %s\n", apiErr.Message)
			} else {
				fmt.Printf("  Unexpected error: %s (Status: %d)\n", apiErr.Message, apiErr.Status)
			}
		} else {
			fmt.Printf("  Network error: %v\n", err)
		}
	} else {
		fmt.Printf("  Unexpected success - this shouldn't happen with invalid data\n")
	}

	fmt.Println("\n=== Record Update Examples Complete ===")
}

// UpdateRecordWithConditionalLogicExample demonstrates updating records with conditional logic
func UpdateRecordWithConditionalLogicExample() {
	fmt.Println("\n=== Update Record with Conditional Logic Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := SetupDemoClient("http://localhost:8090")

	// First, get a record to see its current state
	fmt.Println("1. Fetching current record state...")
	// Note: Replace with actual record ID
	currentRecord, err := client.GetRecord(ctx, "posts", "EXISTING_POST_ID")
	if err != nil {
		log.Printf("Failed to get current record: %v", err)
		return
	}

	fmt.Printf("  Current title: %s\n", currentRecord["title"])
	fmt.Printf("  Current status: %s\n", currentRecord["status"])

	// Example: Only update if certain conditions are met
	fmt.Println("\n2. Conditional update based on current state...")

	// Check if the record needs updating
	if currentRecord["status"] == "draft" {
		fmt.Printf("  Record is in draft status, updating to published...\n")

		conditionalUpdate := pocketbase.Record{
			"status":   "published",
			"featured": true,
		}

		updatedRecord, err := client.UpdateRecord(ctx, "posts", "EXISTING_POST_ID", conditionalUpdate)
		if err != nil {
			log.Printf("Failed to conditionally update record: %v", err)
		} else {
			fmt.Printf("  ✓ Successfully updated record to published status!\n")
			fmt.Printf("    New status: %s\n", updatedRecord["status"])
			fmt.Printf("    Featured: %t\n", updatedRecord["featured"])
		}
	} else {
		fmt.Printf("  Record is already published, no update needed\n")
	}

	// Example: Update based on timestamp comparison
	fmt.Println("\n3. Update based on timestamp comparison...")
	if createdAt, ok := currentRecord["created"].(string); ok {
		fmt.Printf("  Record created at: %s\n", createdAt)

		// In a real application, you might parse the timestamp and compare with current time
		// For this example, we'll just demonstrate the pattern
		timestampUpdate := pocketbase.Record{
			"lastReviewed": time.Now().Format(time.RFC3339),
		}

		_, err = client.UpdateRecord(ctx, "posts", "EXISTING_POST_ID", timestampUpdate)
		if err != nil {
			log.Printf("Failed to update review timestamp: %v", err)
		} else {
			fmt.Printf("  ✓ Updated review timestamp successfully!\n")
		}
	}
}
