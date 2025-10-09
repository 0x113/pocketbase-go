package pocketbase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
)

// CreateRecordWithFiles creates a new record with file uploads in the specified collection.
// The fileUploads parameter should contain the file upload configurations and regular form data.
//
// Example:
//
//	file1, _ := os.Open("document1.pdf")
//	defer file1.Close()
//
//	files := []pocketbase.FileData{
//		{Reader: file1, Filename: "document1.pdf"},
//	}
//
//	data := pocketbase.Record{
//		"title":       "My Document",
//		"description": "A collection of documents",
//	}
//
//	createdRecord, err := client.CreateRecordWithFiles(ctx, "documents",
//		pocketbase.WithFormData(data),
//		pocketbase.WithFileUpload("files", files))
func (c *Client) CreateRecordWithFiles(ctx context.Context, collection string, fileUploads ...FileUploadOption) (Record, error) {
	options := &FileUploadOptions{}
	for _, opt := range fileUploads {
		opt(options)
	}

	endpoint := fmt.Sprintf("/api/collections/%s/records", collection)

	var createdRecord Record
	err := c.doRequest(ctx, "POST", endpoint, options, &createdRecord)
	if err != nil {
		return nil, err
	}

	return createdRecord, nil
}

// UpdateRecordWithFiles updates an existing record with file uploads in the specified collection.
// The fileUploads parameter should contain the file upload configurations and regular form data.
//
// Example:
//
//	// Replace existing files
//	file, _ := os.Open("new-avatar.jpg")
//	defer file.Close()
//
//	files := []pocketbase.FileData{{Reader: file, Filename: "new-avatar.jpg"}}
//	data := pocketbase.Record{"name": "Updated Name"}
//
//	updatedRecord, err := client.UpdateRecordWithFiles(ctx, "users", "user-id",
//		pocketbase.WithFormData(data),
//		pocketbase.WithFileUpload("avatar", files))
//
//	// Append new files to existing ones
//	appendFiles := []pocketbase.FileData{{Reader: fileReader, Filename: "document3.pdf"}}
//	updatedRecord, err := client.UpdateRecordWithFiles(ctx, "documents", "doc-id",
//		pocketbase.WithFileUpload("files", appendFiles, pocketbase.WithAppend()))
//
//	// Delete specific files
//	updatedRecord, err := client.UpdateRecordWithFiles(ctx, "documents", "doc-id",
//		pocketbase.WithFileUpload("files", nil, pocketbase.WithDelete("old-file1.pdf", "old-file2.pdf")))
func (c *Client) UpdateRecordWithFiles(ctx context.Context, collection, recordID string, fileUploads ...FileUploadOption) (Record, error) {
	options := &FileUploadOptions{}
	for _, opt := range fileUploads {
		opt(options)
	}

	endpoint := fmt.Sprintf("/api/collections/%s/records/%s", collection, recordID)

	var updatedRecord Record
	err := c.doRequest(ctx, "PATCH", endpoint, options, &updatedRecord)
	if err != nil {
		return nil, err
	}

	return updatedRecord, nil
}

// CreateFileData creates a FileData struct from an io.Reader
func CreateFileData(reader io.Reader, filename string) FileData {
	return FileData{
		Reader:   reader,
		Filename: filename,
	}
}

// CreateFileDataFromBytes creates a FileData struct from byte data
func CreateFileDataFromBytes(data []byte, filename string) FileData {
	return FileData{
		Reader:   bytes.NewReader(data),
		Filename: filename,
		Size:     int64(len(data)),
	}
}

// CreateFileDataFromFile creates a FileData struct from a file path.
// Note: The caller is responsible for closing the file when done.
func CreateFileDataFromFile(filepath string) (FileData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return FileData{}, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return FileData{}, fmt.Errorf("failed to stat file: %w", err)
	}

	return FileData{
		Reader:   file,
		Filename: stat.Name(),
		Size:     stat.Size(),
	}, nil
}
