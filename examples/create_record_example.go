package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x113/pocketbase-go"
)

// CreateRecordExample demonstrates creating new records in PocketBase collections
func CreateRecordExample() {
	fmt.Println("=== Create Record Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with demo token
	client := SetupDemoClient("http://localhost:8090")

	// Example 1: Basic record creation
	fmt.Println("1. Creating a basic post record...")
	postData := pocketbase.Record{
		"title":   "My First Post",
		"content": "This is the content of my first post created via the Go client.",
		"status":  "published",
		"tags":    []string{"golang", "pocketbase", "tutorial"},
	}

	createdPost, err := client.CreateRecord(ctx, "posts", postData)
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to create post: %s (Status: %d)", apiErr.Message, apiErr.Status)
			if apiErr.IsBadRequest() {
				log.Printf("Validation error in the submitted data")
			}
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully created post!\n")
		fmt.Printf("  ID: %s\n", createdPost["id"])
		fmt.Printf("  Title: %s\n", createdPost["title"])
		fmt.Printf("  Status: %s\n", createdPost["status"])
		if createdAt, ok := createdPost["created"]; ok {
			fmt.Printf("  Created: %s\n", createdAt)
		}
	}

	// Example 2: Creating a user profile with relations
	fmt.Println("\n2. Creating a user profile with expanded relations...")
	profileData := pocketbase.Record{
		"username":  "johndoe",
		"email":     "john.doe@example.com",
		"firstName": "John",
		"lastName":  "Doe",
		"bio":       "Software developer passionate about Go and PocketBase",
		"location":  "San Francisco, CA",
		"website":   "https://johndoe.dev",
		"isPublic":  true,
	}

	createdProfile, err := client.CreateRecord(ctx, "profiles", profileData,
		pocketbase.WithExpand("user", "interests"),
		pocketbase.WithFields("id", "username", "firstName", "lastName", "bio", "user", "interests"))
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to create profile: %s (Status: %d)", apiErr.Message, apiErr.Status)
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully created profile!\n")
		fmt.Printf("  ID: %s\n", createdProfile["id"])
		fmt.Printf("  Username: %s\n", createdProfile["username"])
		fmt.Printf("  Name: %s %s\n", createdProfile["firstName"], createdProfile["lastName"])
		fmt.Printf("  Bio: %s\n", createdProfile["bio"])

		// Show expanded relations if present
		if expandData, ok := createdProfile["expand"]; ok {
			if expandMap, ok := expandData.(pocketbase.Record); ok {
				if user, ok := expandMap["user"]; ok {
					fmt.Printf("  Associated User: %v\n", user)
				}
				if interests, ok := expandMap["interests"]; ok {
					fmt.Printf("  Interests: %v\n", interests)
				}
			}
		}
	}

	// Example 3: Creating a comment with parent relationship
	fmt.Println("\n3. Creating a comment on a post...")
	commentData := pocketbase.Record{
		"content":    "Great post! Thanks for sharing this tutorial.",
		"author":     "user-id-123", // This would typically be the authenticated user's ID
		"parentPost": "post-id-456", // Reference to the post this comments on
		"isApproved": false,
	}

	createdComment, err := client.CreateRecord(ctx, "comments", commentData)
	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to create comment: %s (Status: %d)", apiErr.Message, apiErr.Status)
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully created comment!\n")
		fmt.Printf("  ID: %s\n", createdComment["id"])
		fmt.Printf("  Content: %s\n", createdComment["content"])
		if updatedAt, ok := createdComment["updated"]; ok {
			fmt.Printf("  Last Updated: %s\n", updatedAt)
		}
	}

	// Example 4: Creating multiple related records
	fmt.Println("\n4. Creating related records (category and post)...")
	categoryData := pocketbase.Record{
		"name":        "Technology",
		"description": "Posts about technology and programming",
		"color":       "#3B82F6",
		"isActive":    true,
	}

	createdCategory, err := client.CreateRecord(ctx, "categories", categoryData)
	if err != nil {
		log.Printf("Failed to create category: %v", err)
	} else {
		fmt.Printf("✓ Created category: %s\n", createdCategory["name"])

		// Now create a post in this category
		techPostData := pocketbase.Record{
			"title":    "Building APIs with PocketBase and Go",
			"content":  "In this comprehensive guide, we'll explore how to build robust APIs using PocketBase as the backend and Go as the client...",
			"status":   "published",
			"category": createdCategory["id"], // Link to the created category
			"readTime": 5,
			"featured": true,
		}

		createdTechPost, err := client.CreateRecord(ctx, "posts", techPostData)
		if err != nil {
			log.Printf("Failed to create tech post: %v", err)
		} else {
			fmt.Printf("✓ Created tech post: %s\n", createdTechPost["title"])
			fmt.Printf("  Category: %s\n", createdCategory["name"])
		}
	}

	fmt.Println("\n=== Record Creation Examples Complete ===")
}

// CreateRecordWithValidationExample demonstrates creating records with validation handling
func CreateRecordWithValidationExample() {
	fmt.Println("\n=== Create Record with Validation Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := SetupDemoClient("http://localhost:8090")

	// This example shows how to handle validation errors gracefully
	attempts := []pocketbase.Record{
		// First attempt - missing required fields
		{
			"content": "Content without title",
		},
		// Second attempt - with required fields
		{
			"title":   "Properly Formatted Post",
			"content": "This post has all required fields",
			"status":  "draft",
		},
	}

	for i, recordData := range attempts {
		fmt.Printf("\nAttempt %d:\n", i+1)

		_, err := client.CreateRecord(ctx, "posts", recordData)
		if err != nil {
			if apiErr, ok := err.(*pocketbase.APIError); ok {
				if apiErr.IsBadRequest() {
					fmt.Printf("  Validation failed as expected: %s\n", apiErr.Message)
					// In a real application, you would show these validation errors to the user
					// and allow them to correct the form data
				} else {
					fmt.Printf("  Unexpected error: %s (Status: %d)\n", apiErr.Message, apiErr.Status)
				}
			} else {
				fmt.Printf("  Network error: %v\n", err)
			}
		} else {
			fmt.Printf("  ✓ Record created successfully!\n")
		}
	}
}
