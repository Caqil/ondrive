package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MaxFileSize defines maximum file size (10MB)
	MaxFileSize = 10 << 20 // 10MB
	// MaxImageSize defines maximum image size (5MB)
	MaxImageSize = 5 << 20 // 5MB
	// MaxDocumentSize defines maximum document size (10MB)
	MaxDocumentSize = 10 << 20 // 10MB
)

var (
	// AllowedImageTypes defines allowed image MIME types
	AllowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	// AllowedDocumentTypes defines allowed document MIME types
	AllowedDocumentTypes = map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain": true,
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
	}

	// AllowedAudioTypes defines allowed audio MIME types
	AllowedAudioTypes = map[string]bool{
		"audio/mpeg": true,
		"audio/mp3":  true,
		"audio/wav":  true,
		"audio/ogg":  true,
		"audio/m4a":  true,
		"audio/aac":  true,
	}
)

// FileUploadConfig holds configuration for file uploads
type FileUploadConfig struct {
	MaxSize       int64
	AllowedTypes  map[string]bool
	UploadPath    string
	GenerateThumb bool
}

// UploadResult contains information about uploaded file
type UploadResult struct {
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// UploadFile handles file upload with validation
func UploadFile(file *multipart.FileHeader, config FileUploadConfig) (*UploadResult, error) {
	// Validate file size
	if file.Size > config.MaxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.MaxSize)
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Detect content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Reset file pointer
	src.Seek(0, 0)

	contentType := DetectContentType(buffer, file.Filename)

	// Validate file type
	if !config.AllowedTypes[contentType] {
		return nil, fmt.Errorf("file type %s is not allowed", contentType)
	}

	// Generate unique filename
	fileName := GenerateFileName(file.Filename)
	filePath := filepath.Join(config.UploadPath, fileName)

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(config.UploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	result := &UploadResult{
		FileName:     fileName,
		OriginalName: file.Filename,
		Size:         file.Size,
		MimeType:     contentType,
		URL:          "/uploads/" + fileName,
	}

	// Generate thumbnail for images if requested
	if config.GenerateThumb && IsImageType(contentType) {
		thumbPath, err := GenerateThumbnail(filePath, config.UploadPath)
		if err == nil {
			result.ThumbnailURL = "/uploads/" + thumbPath
		}
	}

	return result, nil
}

// UploadImage uploads and validates image files
func UploadImage(file *multipart.FileHeader, uploadPath string) (*UploadResult, error) {
	config := FileUploadConfig{
		MaxSize:       MaxImageSize,
		AllowedTypes:  AllowedImageTypes,
		UploadPath:    uploadPath,
		GenerateThumb: true,
	}
	return UploadFile(file, config)
}

// UploadDocument uploads and validates document files
func UploadDocument(file *multipart.FileHeader, uploadPath string) (*UploadResult, error) {
	config := FileUploadConfig{
		MaxSize:      MaxDocumentSize,
		AllowedTypes: AllowedDocumentTypes,
		UploadPath:   uploadPath,
	}
	return UploadFile(file, config)
}

// UploadAudio uploads and validates audio files
func UploadAudio(file *multipart.FileHeader, uploadPath string) (*UploadResult, error) {
	config := FileUploadConfig{
		MaxSize:      MaxFileSize,
		AllowedTypes: AllowedAudioTypes,
		UploadPath:   uploadPath,
	}
	return UploadFile(file, config)
}

// GenerateFileName generates a unique filename while preserving extension
func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	randomStr, _ := GenerateRandomString(8)
	return fmt.Sprintf("%d_%s%s", timestamp, randomStr, ext)
}

// DetectContentType detects file content type
func DetectContentType(buffer []byte, filename string) string {
	// Basic MIME type detection based on file extension
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt":
		return "text/plain"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/m4a"
	case ".aac":
		return "audio/aac"
	default:
		return "application/octet-stream"
	}
}

// IsImageType checks if content type is an image
func IsImageType(contentType string) bool {
	return AllowedImageTypes[contentType]
}

// IsDocumentType checks if content type is a document
func IsDocumentType(contentType string) bool {
	return AllowedDocumentTypes[contentType]
}

// IsAudioType checks if content type is audio
func IsAudioType(contentType string) bool {
	return AllowedAudioTypes[contentType]
}

// GenerateThumbnail generates thumbnail for images (placeholder implementation)
func GenerateThumbnail(imagePath, uploadPath string) (string, error) {
	// This is a placeholder implementation
	// In a real application, you would use an image processing library
	// like "github.com/disintegration/imaging" to generate thumbnails

	fileName := filepath.Base(imagePath)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)
	thumbName := nameWithoutExt + "_thumb" + ext

	// For now, just copy the original file as thumbnail
	// In production, you should resize the image
	src, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	thumbPath := filepath.Join(uploadPath, thumbName)
	dst, err := os.Create(thumbPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return thumbName, nil
}

// DeleteFile deletes a file from the filesystem
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(filePath)
}

// GetFileSize returns the size of a file
func GetFileSize(filePath string) (int64, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
