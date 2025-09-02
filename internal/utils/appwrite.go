package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
)

// UploadFile uploads a file to Appwrite storage using HTTP API
func UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (map[string]interface{}, error) {
	// Validate file type (only PDF allowed)
	if !isValidPDF(fileHeader) {
		return nil, fmt.Errorf("only PDF files are allowed")
	}

	// Validate file size (max 10MB)
	if fileHeader.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}

	// Generate unique filename for storage
	fileName := generateUniqueFileName(fileHeader.Filename, folder)
	
	// Generate simple fileId for Appwrite (max 36 chars, valid chars only)
	fileId := generateSimpleFileId()

	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file content
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Create HTTP request to Appwrite API
	url := fmt.Sprintf("%s/storage/buckets/%s/files", 
		config.AppConfig.Appwrite.Endpoint,
		config.AppConfig.Appwrite.BucketID)

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Add file
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = part.Write(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}
	
	// Add file ID (simple, valid format)
	err = writer.WriteField("fileId", fileId)
	if err != nil {
		return nil, fmt.Errorf("failed to write file ID: %w", err)
	}
	
	writer.Close()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Appwrite-Project", config.AppConfig.Appwrite.ProjectID)
	req.Header.Set("X-Appwrite-Key", config.AppConfig.Appwrite.APIKey)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Generate file URL
	fileURL := fmt.Sprintf("%s/storage/buckets/%s/files/%s/view?project=%s",
		config.AppConfig.Appwrite.Endpoint,
		config.AppConfig.Appwrite.BucketID,
		fileId,
		config.AppConfig.Appwrite.ProjectID,
	)

	return map[string]interface{}{
		"fileId":   fileId,
		"fileName": fileName,
		"fileUrl":  fileURL,
		"fileSize": fileHeader.Size,
		"mimeType": "application/pdf",
	}, nil
}

// DeleteFile deletes a file from Appwrite storage using HTTP API
func DeleteFile(ctx context.Context, fileID string) error {
	// Create HTTP request to Appwrite API
	url := fmt.Sprintf("%s/storage/buckets/%s/files/%s", 
		config.AppConfig.Appwrite.Endpoint,
		config.AppConfig.Appwrite.BucketID,
		fileID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Appwrite-Project", config.AppConfig.Appwrite.ProjectID)
	req.Header.Set("X-Appwrite-Key", config.AppConfig.Appwrite.APIKey)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// isValidPDF checks if the uploaded file is a valid PDF
func isValidPDF(fileHeader *multipart.FileHeader) bool {
	// Check MIME type
	if !strings.EqualFold(fileHeader.Header.Get("Content-Type"), "application/pdf") {
		return false
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	return ext == ".pdf"
}

// generateUniqueFileName generates a unique filename for the uploaded file
func generateUniqueFileName(originalName, folder string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	
	// Sanitize filename (remove special characters)
	name = sanitizeFileName(name)
	
	// Add timestamp for uniqueness
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	
	if folder != "" {
		return fmt.Sprintf("%s/%s_%s%s", folder, name, timestamp, ext)
	}
	return fmt.Sprintf("%s_%s%s", name, timestamp, ext)
}

// sanitizeFileName removes special characters from filename
func sanitizeFileName(name string) string {
	// Replace spaces and special characters with underscores
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		".", "_",
		",", "_",
		"(", "_",
		")", "_",
		"[", "_",
		"]", "_",
		"{", "_",
		"}", "_",
		"<", "_",
		">", "_",
		"&", "_",
		"$", "_",
		"#", "_",
		"@", "_",
		"!", "_",
		"%", "_",
		"^", "_",
		"*", "_",
		"+", "_",
		"=", "_",
		"|", "_",
		"\\", "_",
		"/", "_",
		"`", "_",
		"~", "_",
	)
	
	result := replacer.Replace(name)
	
	// Remove multiple consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	
	// Remove leading/trailing underscores
	result = strings.Trim(result, "_")
	
	return result
}

// generateSimpleFileId generates a simple file ID that meets Appwrite requirements
func generateSimpleFileId() string {
	// Generate a 20-character ID using only valid characters
	// Format: f_<timestamp>_<random>
	timestamp := time.Now().UnixNano() / 1000000 // milliseconds
	random := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	
	// Create ID: f_<timestamp>_<random> (max 36 chars)
	fileId := fmt.Sprintf("f_%d_%s", timestamp, random)
	
	// Ensure it's not longer than 36 characters
	if len(fileId) > 36 {
		fileId = fileId[:36]
	}
	
	return fileId
}
