package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ekf/one-on-one-backend/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// FileMetadata represents file metadata in the database
type FileMetadata struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	StoragePath  string    `json:"storage_path"`
	Bucket       string    `json:"bucket"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	UploadedBy   string    `json:"uploaded_by,omitempty"`
	EntityType   string    `json:"entity_type,omitempty"`
	EntityID     string    `json:"entity_id,omitempty"`
	URL          string    `json:"url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

const defaultBucket = "attachments"

// File upload security constants
const (
	// MaxFileSize is the maximum allowed file size (50 MB)
	MaxFileSize = 50 * 1024 * 1024
	// MaxFileSizeForImages is the maximum allowed image size (10 MB)
	MaxFileSizeForImages = 10 * 1024 * 1024
	// SniffSize is the number of bytes to read for MIME detection
	SniffSize = 512
)

// AllowedFileTypes defines the whitelist of allowed file extensions and their MIME types
var AllowedFileTypes = map[string][]string{
	// Documents
	".pdf":  {"application/pdf"},
	".doc":  {"application/msword"},
	".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	".xls":  {"application/vnd.ms-excel"},
	".xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	".ppt":  {"application/vnd.ms-powerpoint"},
	".pptx": {"application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	".txt":  {"text/plain"},
	".csv":  {"text/csv", "text/plain", "application/csv"},
	".json": {"application/json", "text/plain"},
	".xml":  {"application/xml", "text/xml", "text/plain"},
	// Images
	".png":  {"image/png"},
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".gif":  {"image/gif"},
	".webp": {"image/webp"},
	// Archives (limited support)
	".zip": {"application/zip", "application/x-zip-compressed"},
	// Audio (for meeting recordings)
	".mp3": {"audio/mpeg"},
	".wav": {"audio/wav", "audio/x-wav"},
	".ogg": {"audio/ogg"},
	// Video (for meeting recordings)
	".mp4":  {"video/mp4"},
	".webm": {"video/webm"},
}

// DangerousExtensions are extensions that should never be allowed
var DangerousExtensions = map[string]bool{
	".exe": true, ".bat": true, ".cmd": true, ".sh": true,
	".ps1": true, ".vbs": true, ".js": true, ".msi": true,
	".dll": true, ".com": true, ".scr": true, ".hta": true,
	".jar": true, ".php": true, ".asp": true, ".aspx": true,
	".jsp": true, ".py": true, ".rb": true, ".pl": true,
	".cgi": true, ".htaccess": true, ".svg": true, // SVG can contain scripts
}

// validateFileUpload performs security validation on uploaded file
func validateFileUpload(filename string, size int64, contentBuffer []byte) (string, error) {
	// 1. Check file extension against whitelist
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "", fmt.Errorf("file must have an extension")
	}

	// Check for dangerous extensions first
	if DangerousExtensions[ext] {
		return "", fmt.Errorf("file type not allowed for security reasons: %s", ext)
	}

	allowedMimes, ok := AllowedFileTypes[ext]
	if !ok {
		return "", fmt.Errorf("file type not allowed: %s", ext)
	}

	// 2. Check file size
	maxSize := int64(MaxFileSize)
	if strings.HasPrefix(allowedMimes[0], "image/") {
		maxSize = MaxFileSizeForImages
	}
	if size > maxSize {
		return "", fmt.Errorf("file size exceeds limit: %d bytes (max: %d bytes)", size, maxSize)
	}

	// 3. Detect actual MIME type from file content (magic bytes)
	detectedMime := http.DetectContentType(contentBuffer)
	// Normalize detected MIME (http.DetectContentType may return charset info)
	detectedMime = strings.Split(detectedMime, ";")[0]
	detectedMime = strings.TrimSpace(detectedMime)

	// 4. Validate MIME type matches expected for extension
	mimeValid := false
	for _, allowed := range allowedMimes {
		if detectedMime == allowed {
			mimeValid = true
			break
		}
	}

	// Some files like DOCX/XLSX are detected as application/zip by magic bytes
	// so we need special handling for Office documents
	if !mimeValid {
		// Office Open XML formats are ZIP archives
		if (ext == ".docx" || ext == ".xlsx" || ext == ".pptx") &&
			(detectedMime == "application/zip" || detectedMime == "application/octet-stream") {
			mimeValid = true
		}
		// Text-based files might be detected as text/plain
		if (ext == ".csv" || ext == ".json" || ext == ".xml" || ext == ".txt") &&
			detectedMime == "text/plain" {
			mimeValid = true
		}
		// Binary fallback for some formats
		if detectedMime == "application/octet-stream" {
			// Allow for document types that may not have clear magic bytes
			if ext == ".doc" || ext == ".xls" || ext == ".ppt" {
				mimeValid = true
			}
		}
	}

	if !mimeValid {
		return "", fmt.Errorf("file content does not match extension: detected %s for %s file", detectedMime, ext)
	}

	// 5. Additional checks for dangerous content
	// Check for script content in text files
	if strings.HasPrefix(detectedMime, "text/") {
		content := strings.ToLower(string(contentBuffer))
		dangerousPatterns := []string{
			"<script", "javascript:", "vbscript:", "onload=", "onerror=",
			"onclick=", "onmouseover=", "eval(", "document.cookie",
		}
		for _, pattern := range dangerousPatterns {
			if strings.Contains(content, pattern) {
				return "", fmt.Errorf("file contains potentially dangerous content")
			}
		}
	}

	return detectedMime, nil
}

// UploadFile handles file uploads with security validation
func (h *Handler) UploadFile(c *fiber.Ctx) error {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file provided"})
	}

	// 1. Early size check before reading file
	if file.Size > MaxFileSize {
		return c.Status(413).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large: %d bytes (max: %d bytes)", file.Size, MaxFileSize),
		})
	}

	// Get optional metadata
	entityType := c.FormValue("entity_type", "")
	entityID := c.FormValue("entity_id", "")
	uploadedBy := c.FormValue("uploaded_by", "")
	bucket := c.FormValue("bucket", defaultBucket)

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file"})
	}
	defer fileContent.Close()

	// 2. Read first 512 bytes for MIME detection (magic bytes)
	sniffBuffer := make([]byte, SniffSize)
	n, err := fileContent.Read(sniffBuffer)
	if err != nil && err != io.EOF {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file header"})
	}
	sniffBuffer = sniffBuffer[:n]

	// 3. Validate file (extension, size, MIME, dangerous content)
	validatedMime, err := validateFileUpload(file.Filename, file.Size, sniffBuffer)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 4. Reset file reader and read remaining content
	_, err = fileContent.Seek(0, io.SeekStart)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file"})
	}

	// Read file content with size limit
	limitedReader := io.LimitReader(fileContent, MaxFileSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file content"})
	}
	if int64(len(data)) > MaxFileSize {
		return c.Status(413).JSON(fiber.Map{"error": "File too large"})
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(file.Filename))
	fileID := uuid.New().String()
	storagePath := storage.GeneratePath(entityType, entityID, fileID, ext)

	// Use validated MIME type (from magic bytes detection)
	contentType := validatedMime

	// Upload to MinIO Storage
	var publicURL string
	if h.Storage != nil {
		ctx := context.Background()
		publicURL, err = h.Storage.Upload(ctx, storagePath, bytes.NewReader(data), int64(len(data)), contentType)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Storage upload failed: %v", err)})
		}
	} else {
		return c.Status(500).JSON(fiber.Map{"error": "Storage not configured"})
	}

	// Save metadata to database
	metadata := map[string]interface{}{
		"id":            fileID,
		"name":          fileID + ext,
		"original_name": file.Filename,
		"storage_path":  storagePath,
		"bucket":        bucket,
		"content_type":  contentType,
		"size_bytes":    file.Size,
	}

	if uploadedBy != "" {
		metadata["uploaded_by"] = uploadedBy
	}
	if entityType != "" {
		metadata["entity_type"] = entityType
	}
	if entityID != "" {
		metadata["entity_id"] = entityID
	}

	_, err = h.DB.Insert("files", metadata)
	if err != nil {
		// Try to clean up uploaded file
		if h.Storage != nil {
			_ = h.Storage.Delete(context.Background(), storagePath)
		}
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to save file metadata: %v", err)})
	}

	return c.JSON(fiber.Map{
		"id":           fileID,
		"name":         file.Filename,
		"url":          publicURL,
		"content_type": contentType,
		"size_bytes":   file.Size,
		"storage_path": storagePath,
	})
}

// GetFile downloads a file by ID
func (h *Handler) GetFile(c *fiber.Ctx) error {
	fileID := c.Params("id")

	// Get file metadata
	var files []FileMetadata
	err := h.DB.From("files").Select("*").Eq("id", fileID).Execute(&files)
	if err != nil || len(files) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	file := files[0]

	// Download from MinIO storage
	if h.Storage == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Storage not configured"})
	}

	ctx := context.Background()
	data, contentType, err := h.Storage.Download(ctx, file.StoragePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to download file: %v", err)})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalName))
	return c.Send(data)
}

// DeleteFile deletes a file by ID
func (h *Handler) DeleteFile(c *fiber.Ctx) error {
	fileID := c.Params("id")

	// Get file metadata
	var files []FileMetadata
	err := h.DB.From("files").Select("*").Eq("id", fileID).Execute(&files)
	if err != nil || len(files) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	file := files[0]

	// Delete from MinIO storage
	if h.Storage != nil {
		ctx := context.Background()
		err = h.Storage.Delete(ctx, file.StoragePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to delete file from storage: %v", err)})
		}
	}

	// Delete metadata
	err = h.DB.Delete("files", "id", fileID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to delete file metadata: %v", err)})
	}

	return c.JSON(fiber.Map{"success": true})
}

// ListFiles lists files for an entity
func (h *Handler) ListFiles(c *fiber.Ctx) error {
	entityType := c.Query("entity_type")
	entityID := c.Query("entity_id")
	uploadedBy := c.Query("uploaded_by")

	query := h.DB.From("files").Select("*")

	if entityType != "" {
		query = query.Eq("entity_type", entityType)
	}
	if entityID != "" {
		query = query.Eq("entity_id", entityID)
	}
	if uploadedBy != "" {
		query = query.Eq("uploaded_by", uploadedBy)
	}

	query = query.Order("created_at", true).Limit(100)

	var files []FileMetadata
	err := query.Execute(&files)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Add URLs from MinIO
	if h.Storage != nil {
		for i := range files {
			files[i].URL = h.Storage.GetPublicURL(files[i].StoragePath)
		}
	}

	return c.JSON(files)
}

// GetFileURL returns the public URL for a file
func (h *Handler) GetFileURL(c *fiber.Ctx) error {
	fileID := c.Params("id")

	var files []FileMetadata
	err := h.DB.From("files").Select("*").Eq("id", fileID).Execute(&files)
	if err != nil || len(files) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	file := files[0]
	var url string
	if h.Storage != nil {
		url = h.Storage.GetPublicURL(file.StoragePath)
	}

	return c.JSON(fiber.Map{
		"id":           file.ID,
		"name":         file.OriginalName,
		"url":          url,
		"content_type": file.ContentType,
	})
}

// AttachFileToEntity attaches an existing file to an entity
func (h *Handler) AttachFileToEntity(c *fiber.Ctx) error {
	var req struct {
		FileID     string `json:"file_id"`
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update file metadata
	_, err := h.DB.Update("files", "id", req.FileID, map[string]interface{}{
		"entity_type": req.EntityType,
		"entity_id":   req.EntityID,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Also update the entity's attachments JSONB field
	if req.EntityType == "task" || req.EntityType == "message" {
		table := req.EntityType + "s" // tasks or messages

		// Get current attachments
		var entities []map[string]interface{}
		err := h.DB.From(table).Select("attachments").Eq("id", req.EntityID).Execute(&entities)
		if err == nil && len(entities) > 0 {
			attachments := []string{}
			if att, ok := entities[0]["attachments"].([]interface{}); ok {
				for _, a := range att {
					if s, ok := a.(string); ok {
						attachments = append(attachments, s)
					}
				}
			}
			attachments = append(attachments, req.FileID)
			attJSON, _ := json.Marshal(attachments)
			h.DB.Update(table, "id", req.EntityID, map[string]interface{}{
				"attachments": string(attJSON),
			})
		}
	}

	return c.JSON(fiber.Map{"success": true})
}

// generateStoragePath is deprecated - use storage.GeneratePath instead
// Kept for backwards compatibility
func generateStoragePath(entityType, entityID, fileID, ext string) string {
	return storage.GeneratePath(entityType, entityID, fileID, ext)
}

// detectContentType detects content type from file extension
func detectContentType(ext string) string {
	ext = strings.ToLower(ext)
	contentTypes := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".json": "application/json",
		".xml":  "application/xml",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}
