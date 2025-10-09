package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/0x113/pocketbase-go"
)

// FileUploadExample demonstrates uploading files to PocketBase collections
func FileUploadExample() {
	fmt.Println("=== File Upload Example ===")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client
	client := SetupDemoClient("http://localhost:8090")

	// Example 1: Create a record with a single file upload
	fmt.Println("\n1. Creating a record with a single file...")
	file1, err := os.Open("example-document.pdf")
	if err != nil {
		log.Printf("Failed to open file: %v (make sure the file exists)", err)
	} else {
		defer file1.Close()

		files := []pocketbase.FileData{
			{Reader: file1, Filename: "example-document.pdf"},
		}

		data := pocketbase.Record{
			"title":       "Important Document",
			"description": "This document contains important information",
		}

		createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
			pocketbase.WithFormData(data),
			pocketbase.WithFileUpload("file", files))

		if err != nil {
			if apiErr, ok := err.(*pocketbase.APIError); ok {
				log.Printf("Failed to create record: %s (Status: %d)", apiErr.Message, apiErr.Status)
			} else {
				log.Printf("Network error: %v", err)
			}
		} else {
			fmt.Printf("✓ Successfully created record with file!\n")
			fmt.Printf("  ID: %s\n", createdRecord["id"])
			fmt.Printf("  Title: %s\n", createdRecord["title"])
			if file, ok := createdRecord["file"]; ok {
				fmt.Printf("  File: %s\n", file)
			}
		}
	}

	// Example 2: Create a record with multiple files
	fmt.Println("\n2. Creating a record with multiple files...")
	file2, err := os.Open("document1.pdf")
	if err != nil {
		log.Printf("Failed to open file: %v", err)
	} else {
		defer file2.Close()

		file3, err := os.Open("document2.pdf")
		if err != nil {
			log.Printf("Failed to open file: %v", err)
			file2.Close()
		} else {
			defer file3.Close()

			files := []pocketbase.FileData{
				{Reader: file2, Filename: "document1.pdf"},
				{Reader: file3, Filename: "document2.pdf"},
			}

			data := pocketbase.Record{
				"title":       "Multiple Documents",
				"description": "A collection with multiple files",
			}

			createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
				pocketbase.WithFormData(data),
				pocketbase.WithFileUpload("files", files))

			if err != nil {
				if apiErr, ok := err.(*pocketbase.APIError); ok {
					log.Printf("Failed to create record: %s (Status: %d)", apiErr.Message, apiErr.Status)
				} else {
					log.Printf("Network error: %v", err)
				}
			} else {
				fmt.Printf("✓ Successfully created record with multiple files!\n")
				fmt.Printf("  ID: %s\n", createdRecord["id"])
				fmt.Printf("  Title: %s\n", createdRecord["title"])
				if files, ok := createdRecord["files"]; ok {
					fmt.Printf("  Files: %v\n", files)
				}
			}
		}
	}

	// Example 3: Update a record by replacing files
	fmt.Println("\n3. Updating a record by replacing a file...")
	newFile, err := os.Open("new-avatar.jpg")
	if err != nil {
		log.Printf("Failed to open file: %v", err)
	} else {
		defer newFile.Close()

		files := []pocketbase.FileData{
			{Reader: newFile, Filename: "new-avatar.jpg"},
		}

		data := pocketbase.Record{
			"name": "Updated User",
		}

		updatedRecord, err := client.UpdateRecordWithFiles(ctx, "users", "RECORD_ID_HERE",
			pocketbase.WithFormData(data),
			pocketbase.WithFileUpload("avatar", files))

		if err != nil {
			if apiErr, ok := err.(*pocketbase.APIError); ok {
				log.Printf("Failed to update record: %s (Status: %d)", apiErr.Message, apiErr.Status)
			} else {
				log.Printf("Network error: %v", err)
			}
		} else {
			fmt.Printf("✓ Successfully updated record with new file!\n")
			fmt.Printf("  ID: %s\n", updatedRecord["id"])
			fmt.Printf("  Name: %s\n", updatedRecord["name"])
		}
	}

	// Example 4: Append files to existing record
	fmt.Println("\n4. Appending files to an existing record...")
	appendFile, err := os.Open("document3.pdf")
	if err != nil {
		log.Printf("Failed to open file: %v", err)
	} else {
		defer appendFile.Close()

		files := []pocketbase.FileData{
			{Reader: appendFile, Filename: "document3.pdf"},
		}

		updatedRecord, err := client.UpdateRecordWithFiles(ctx, "documents", "RECORD_ID_HERE",
			pocketbase.WithFileUpload("files", files, pocketbase.WithAppend()))

		if err != nil {
			if apiErr, ok := err.(*pocketbase.APIError); ok {
				log.Printf("Failed to append file: %s (Status: %d)", apiErr.Message, apiErr.Status)
			} else {
				log.Printf("Network error: %v", err)
			}
		} else {
			fmt.Printf("✓ Successfully appended file to record!\n")
			fmt.Printf("  ID: %s\n", updatedRecord["id"])
			if files, ok := updatedRecord["files"]; ok {
				fmt.Printf("  Files: %v\n", files)
			}
		}
	}

	// Example 5: Delete specific files from a record
	fmt.Println("\n5. Deleting specific files from a record...")
	updatedRecord, err := client.UpdateRecordWithFiles(ctx, "documents", "RECORD_ID_HERE",
		pocketbase.WithFileUpload("files", nil, pocketbase.WithDelete("document1.pdf", "document2.pdf")))

	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to delete files: %s (Status: %d)", apiErr.Message, apiErr.Status)
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully deleted files from record!\n")
		fmt.Printf("  ID: %s\n", updatedRecord["id"])
		if files, ok := updatedRecord["files"]; ok {
			fmt.Printf("  Remaining files: %v\n", files)
		}
	}

	// Example 6: Using helper functions to create FileData
	fmt.Println("\n6. Using helper functions to create FileData...")

	// From file path
	fileData, err := pocketbase.CreateFileDataFromFile("example.txt")
	if err != nil {
		log.Printf("Failed to create file data from path: %v", err)
	} else {
		if fileReader, ok := fileData.Reader.(*os.File); ok {
			defer fileReader.Close()
		}

		data := pocketbase.Record{
			"title": "Example with Helper",
		}

		createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
			pocketbase.WithFormData(data),
			pocketbase.WithFileUpload("file", []pocketbase.FileData{fileData}))

		if err != nil {
			log.Printf("Failed to create record: %v", err)
		} else {
			fmt.Printf("✓ Successfully created record using helper function!\n")
			fmt.Printf("  ID: %s\n", createdRecord["id"])
		}
	}

	// From bytes
	fileBytes := []byte("This is file content from bytes")
	fileDataFromBytes := pocketbase.CreateFileDataFromBytes(fileBytes, "content.txt")

	data := pocketbase.Record{
		"title": "File from Bytes",
	}

	createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
		pocketbase.WithFormData(data),
		pocketbase.WithFileUpload("file", []pocketbase.FileData{fileDataFromBytes}))

	if err != nil {
		log.Printf("Failed to create record: %v", err)
	} else {
		fmt.Printf("✓ Successfully created record from byte data!\n")
		fmt.Printf("  ID: %s\n", createdRecord["id"])
		fmt.Printf("  Title: %s\n", createdRecord["title"])
	}

	fmt.Println("\n=== File Upload Examples Complete ===")
}

// FileUploadWithQueryOptionsExample demonstrates using query options with file uploads
func FileUploadWithQueryOptionsExample() {
	fmt.Println("\n=== File Upload with Query Options Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := SetupDemoClient("http://localhost:8090")

	// Create a record with files and expand related fields
	file, err := os.Open("profile-picture.jpg")
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return
	}
	defer file.Close()

	files := []pocketbase.FileData{
		{Reader: file, Filename: "profile-picture.jpg"},
	}

	data := pocketbase.Record{
		"name":       "John Doe",
		"email":      "john@example.com",
		"department": "DEPARTMENT_ID",
	}

	// Create with file upload and expand department relation
	createdRecord, err := client.CreateRecordWithFiles(ctx, "employees",
		pocketbase.WithFormData(data),
		pocketbase.WithFileUpload("photo", files),
		// Note: WithExpand and WithFields work as FileUploadOptions
		func(opts *pocketbase.FileUploadOptions) {
			opts.Expand = []string{"department"}
			opts.Fields = []string{"id", "name", "email", "photo", "department"}
		})

	if err != nil {
		if apiErr, ok := err.(*pocketbase.APIError); ok {
			log.Printf("Failed to create record: %s (Status: %d)", apiErr.Message, apiErr.Status)
		} else {
			log.Printf("Network error: %v", err)
		}
	} else {
		fmt.Printf("✓ Successfully created record with expanded relations!\n")
		fmt.Printf("  ID: %s\n", createdRecord["id"])
		fmt.Printf("  Name: %s\n", createdRecord["name"])
		if expand, ok := createdRecord["expand"]; ok {
			fmt.Printf("  Expanded data: %v\n", expand)
		}
	}

	fmt.Println("\n=== File Upload with Query Options Complete ===")
}
