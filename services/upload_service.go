package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UploadService interface for file upload operations
type UploadService interface {
	// Courier-specific uploads
	UploadPackagePhoto(file *multipart.FileHeader, requestID string) (string, error)
	UploadDeliveryProof(file *multipart.FileHeader, requestID string) (string, error)
	UploadSignature(file *multipart.FileHeader, requestID string) (string, error)
	UploadRecipientPhoto(file *multipart.FileHeader, requestID string) (string, error)
	UploadIssuePhoto(file *multipart.FileHeader, issueID string) (string, error)
	UploadClaimEvidence(file *multipart.FileHeader, claimID string) (string, error)

	// User uploads
	UploadAvatar(file *multipart.FileHeader, userID string) (string, error)
	UploadDocument(file *multipart.FileHeader, userID, docType string) (string, error)
	UploadIDPhoto(file *multipart.FileHeader, userID string) (string, error)

	// Vehicle uploads (for couriers/drivers)
	UploadVehiclePhoto(file *multipart.FileHeader, userID string) (string, error)
	UploadLicensePhoto(file *multipart.FileHeader, userID string) (string, error)

	// Chat uploads
	UploadChatImage(file *multipart.FileHeader, conversationID string) (string, error)
	UploadChatAudio(file *multipart.FileHeader, conversationID string) (string, error)
	UploadChatDocument(file *multipart.FileHeader, conversationID string) (string, error)

	// General utilities
	DeleteFile(fileURL string) error
	GetFileInfo(fileURL string) (*FileInfo, error)
	ValidateFile(file *multipart.FileHeader, fileType FileType) error
	GeneratePresignedURL(filePath string, expiration time.Duration) (string, error)

	// Batch operations
	UploadMultipleFiles(files []*multipart.FileHeader, uploadType UploadType, entityID string) ([]string, error)
	DeleteMultipleFiles(fileURLs []string) error

	// Image processing
	ResizeImage(fileURL string, width, height int) (string, error)
	CreateThumbnail(fileURL string) (string, error)
	CompressImage(fileURL string, quality int) (string, error)
}

// Supporting types
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeDocument FileType = "document"
	FileTypeAudio    FileType = "audio"
	FileTypeVideo    FileType = "video"
	FileTypeAny      FileType = "any"
)

type UploadType string

const (
	UploadTypePackagePhoto   UploadType = "package_photo"
	UploadTypeDeliveryProof  UploadType = "delivery_proof"
	UploadTypeSignature      UploadType = "signature"
	UploadTypeRecipientPhoto UploadType = "recipient_photo"
	UploadTypeIssuePhoto     UploadType = "issue_photo"
	UploadTypeClaimEvidence  UploadType = "claim_evidence"
	UploadTypeAvatar         UploadType = "avatar"
	UploadTypeDocument       UploadType = "document"
	UploadTypeIDPhoto        UploadType = "id_photo"
	UploadTypeVehiclePhoto   UploadType = "vehicle_photo"
	UploadTypeLicensePhoto   UploadType = "license_photo"
	UploadTypeChatImage      UploadType = "chat_image"
	UploadTypeChatAudio      UploadType = "chat_audio"
	UploadTypeChatDocument   UploadType = "chat_document"
)

type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
	EntityID     string    `json:"entity_id"`
	UploadType   string    `json:"upload_type"`
}

type UploadConfig struct {
	MaxSize        int64           `json:"max_size"`
	AllowedTypes   map[string]bool `json:"allowed_types"`
	UploadPath     string          `json:"upload_path"`
	GenerateThumb  bool            `json:"generate_thumb"`
	CompressImages bool            `json:"compress_images"`
	Quality        int             `json:"quality"`
	CreateVariants bool            `json:"create_variants"`
	Variants       []ImageVariant  `json:"variants"`
}

type ImageVariant struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Suffix string `json:"suffix"`
}

// Implementation
type uploadService struct {
	logger        utils.Logger
	basePath      string
	baseURL       string
	storageType   string // "local", "s3", "gcs"
	uploadConfigs map[UploadType]UploadConfig
}

// Constructor
func NewUploadService(logger utils.Logger, basePath, baseURL, storageType string) UploadService {
	service := &uploadService{
		logger:      logger,
		basePath:    basePath,
		baseURL:     baseURL,
		storageType: storageType,
	}

	service.initUploadConfigs()
	return service
}

func (s *uploadService) initUploadConfigs() {
	s.uploadConfigs = map[UploadType]UploadConfig{
		UploadTypePackagePhoto: {
			MaxSize:        5 << 20, // 5MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "courier/packages",
			GenerateThumb:  true,
			CompressImages: true,
			Quality:        85,
			CreateVariants: true,
			Variants: []ImageVariant{
				{Name: "thumbnail", Width: 150, Height: 150, Suffix: "_thumb"},
				{Name: "medium", Width: 500, Height: 500, Suffix: "_med"},
			},
		},
		UploadTypeDeliveryProof: {
			MaxSize:        5 << 20, // 5MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "courier/delivery_proof",
			GenerateThumb:  true,
			CompressImages: true,
			Quality:        90,
			CreateVariants: true,
			Variants: []ImageVariant{
				{Name: "thumbnail", Width: 200, Height: 200, Suffix: "_thumb"},
			},
		},
		UploadTypeSignature: {
			MaxSize:        1 << 20, // 1MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "courier/signatures",
			CompressImages: true,
			Quality:        95,
		},
		UploadTypeRecipientPhoto: {
			MaxSize:        2 << 20, // 2MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "courier/recipients",
			GenerateThumb:  true,
			CompressImages: true,
			Quality:        85,
		},
		UploadTypeIssuePhoto: {
			MaxSize:        5 << 20, // 5MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "support/issues",
			GenerateThumb:  true,
			CompressImages: true,
			Quality:        90,
		},
		UploadTypeClaimEvidence: {
			MaxSize:       10 << 20, // 10MB
			AllowedTypes:  mergeTypes(utils.AllowedImageTypes, utils.AllowedDocumentTypes),
			UploadPath:    "support/claims",
			GenerateThumb: true,
		},
		UploadTypeAvatar: {
			MaxSize:        2 << 20, // 2MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "users/avatars",
			GenerateThumb:  true,
			CompressImages: true,
			Quality:        85,
			CreateVariants: true,
			Variants: []ImageVariant{
				{Name: "small", Width: 64, Height: 64, Suffix: "_sm"},
				{Name: "medium", Width: 128, Height: 128, Suffix: "_md"},
				{Name: "large", Width: 256, Height: 256, Suffix: "_lg"},
			},
		},
		UploadTypeDocument: {
			MaxSize:      10 << 20, // 10MB
			AllowedTypes: utils.AllowedDocumentTypes,
			UploadPath:   "users/documents",
		},
		UploadTypeIDPhoto: {
			MaxSize:        5 << 20, // 5MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "users/identification",
			CompressImages: true,
			Quality:        95,
		},
		UploadTypeVehiclePhoto: {
			MaxSize:        5 << 20, // 5MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "vehicles/photos",
			GenerateThumb:  true,
			CompressImages: true,
			Quality:        85,
		},
		UploadTypeLicensePhoto: {
			MaxSize:        5 << 20, // 5MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "vehicles/licenses",
			CompressImages: true,
			Quality:        95,
		},
		UploadTypeChatImage: {
			MaxSize:        5 << 20, // 5MB
			AllowedTypes:   utils.AllowedImageTypes,
			UploadPath:     "chat/images",
			GenerateThumb:  true,
			CompressImages: true,
			Quality:        80,
		},
		UploadTypeChatAudio: {
			MaxSize:      10 << 20, // 10MB
			AllowedTypes: utils.AllowedAudioTypes,
			UploadPath:   "chat/audio",
		},
		UploadTypeChatDocument: {
			MaxSize:      10 << 20, // 10MB
			AllowedTypes: utils.AllowedDocumentTypes,
			UploadPath:   "chat/documents",
		},
	}
}

// Courier-specific uploads

func (s *uploadService) UploadPackagePhoto(file *multipart.FileHeader, requestID string) (string, error) {
	return s.uploadFile(file, UploadTypePackagePhoto, requestID)
}

func (s *uploadService) UploadDeliveryProof(file *multipart.FileHeader, requestID string) (string, error) {
	return s.uploadFile(file, UploadTypeDeliveryProof, requestID)
}

func (s *uploadService) UploadSignature(file *multipart.FileHeader, requestID string) (string, error) {
	return s.uploadFile(file, UploadTypeSignature, requestID)
}

func (s *uploadService) UploadRecipientPhoto(file *multipart.FileHeader, requestID string) (string, error) {
	return s.uploadFile(file, UploadTypeRecipientPhoto, requestID)
}

func (s *uploadService) UploadIssuePhoto(file *multipart.FileHeader, issueID string) (string, error) {
	return s.uploadFile(file, UploadTypeIssuePhoto, issueID)
}

func (s *uploadService) UploadClaimEvidence(file *multipart.FileHeader, claimID string) (string, error) {
	return s.uploadFile(file, UploadTypeClaimEvidence, claimID)
}

// User uploads

func (s *uploadService) UploadAvatar(file *multipart.FileHeader, userID string) (string, error) {
	return s.uploadFile(file, UploadTypeAvatar, userID)
}

func (s *uploadService) UploadDocument(file *multipart.FileHeader, userID, docType string) (string, error) {
	return s.uploadFile(file, UploadTypeDocument, userID)
}

func (s *uploadService) UploadIDPhoto(file *multipart.FileHeader, userID string) (string, error) {
	return s.uploadFile(file, UploadTypeIDPhoto, userID)
}

// Vehicle uploads

func (s *uploadService) UploadVehiclePhoto(file *multipart.FileHeader, userID string) (string, error) {
	return s.uploadFile(file, UploadTypeVehiclePhoto, userID)
}

func (s *uploadService) UploadLicensePhoto(file *multipart.FileHeader, userID string) (string, error) {
	return s.uploadFile(file, UploadTypeLicensePhoto, userID)
}

// Chat uploads

func (s *uploadService) UploadChatImage(file *multipart.FileHeader, conversationID string) (string, error) {
	return s.uploadFile(file, UploadTypeChatImage, conversationID)
}

func (s *uploadService) UploadChatAudio(file *multipart.FileHeader, conversationID string) (string, error) {
	return s.uploadFile(file, UploadTypeChatAudio, conversationID)
}

func (s *uploadService) UploadChatDocument(file *multipart.FileHeader, conversationID string) (string, error) {
	return s.uploadFile(file, UploadTypeChatDocument, conversationID)
}

// Core upload method

func (s *uploadService) uploadFile(file *multipart.FileHeader, uploadType UploadType, entityID string) (string, error) {
	config, exists := s.uploadConfigs[uploadType]
	if !exists {
		return "", fmt.Errorf("unsupported upload type: %s", uploadType)
	}

	// Validate file
	if err := s.validateFileWithConfig(file, config); err != nil {
		return "", err
	}

	// Generate unique filename
	fileName := s.generateFileName(file.Filename, entityID)
	filePath := filepath.Join(config.UploadPath, fileName)

	var fileURL string
	var err error

	// Upload based on storage type
	switch s.storageType {
	case "local":
		fileURL, err = s.uploadToLocal(file, filePath)
	case "s3":
		fileURL, err = s.uploadToS3(file, filePath)
	case "gcs":
		fileURL, err = s.uploadToGCS(file, filePath)
	default:
		return "", fmt.Errorf("unsupported storage type: %s", s.storageType)
	}

	if err != nil {
		s.logger.Error().Err(err).Str("file", file.Filename).Msg("Failed to upload file")
		return "", err
	}

	// Post-processing
	if config.CompressImages && utils.IsImageType(utils.DetectContentType(nil, file.Filename)) {
		go s.compressImageAsync(fileURL, config.Quality)
	}

	if config.GenerateThumb && utils.IsImageType(utils.DetectContentType(nil, file.Filename)) {
		go s.createThumbnailAsync(fileURL)
	}

	if config.CreateVariants {
		go s.createVariantsAsync(fileURL, config.Variants)
	}

	s.logger.Info().
		Str("file", file.Filename).
		Str("url", fileURL).
		Str("type", string(uploadType)).
		Str("entity_id", entityID).
		Msg("File uploaded successfully")

	return fileURL, nil
}

// File validation

func (s *uploadService) ValidateFile(file *multipart.FileHeader, fileType FileType) error {
	// Size validation
	maxSize := int64(10 << 20) // Default 10MB
	if file.Size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", file.Size, maxSize)
	}

	// Type validation
	if fileType != FileTypeAny {
		contentType := utils.DetectContentType(nil, file.Filename)
		valid := false

		switch fileType {
		case FileTypeImage:
			valid = utils.IsImageType(contentType)
		case FileTypeDocument:
			valid = utils.IsDocumentType(contentType)
		case FileTypeAudio:
			valid = utils.IsAudioType(contentType)
		}

		if !valid {
			return fmt.Errorf("file type %s is not allowed for %s files", contentType, fileType)
		}
	}

	return nil
}

func (s *uploadService) validateFileWithConfig(file *multipart.FileHeader, config UploadConfig) error {
	// Size validation
	if file.Size > config.MaxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", file.Size, config.MaxSize)
	}

	// Type validation
	contentType := utils.DetectContentType(nil, file.Filename)
	if !config.AllowedTypes[contentType] {
		return fmt.Errorf("file type %s is not allowed", contentType)
	}

	return nil
}

// Storage implementations

func (s *uploadService) uploadToLocal(file *multipart.FileHeader, filePath string) (string, error) {
	fullPath := filepath.Join(s.basePath, filePath)

	// Use existing utility function
	_, err := utils.UploadFile(file, utils.FileUploadConfig{
		MaxSize:      file.Size + 1, // Allow current file size
		AllowedTypes: map[string]bool{utils.DetectContentType(nil, file.Filename): true},
		UploadPath:   filepath.Dir(fullPath),
	})

	if err != nil {
		return "", err
	}

	return s.baseURL + "/" + filePath, nil
}

func (s *uploadService) uploadToS3(file *multipart.FileHeader, filePath string) (string, error) {
	// TODO: Implement S3 upload
	// This would use AWS SDK to upload to S3 bucket
	return "", fmt.Errorf("S3 upload not implemented")
}

func (s *uploadService) uploadToGCS(file *multipart.FileHeader, filePath string) (string, error) {
	// TODO: Implement Google Cloud Storage upload
	// This would use Google Cloud SDK to upload to GCS bucket
	return "", fmt.Errorf("GCS upload not implemented")
}

// Utility methods

func (s *uploadService) DeleteFile(fileURL string) error {
	switch s.storageType {
	case "local":
		return s.deleteFromLocal(fileURL)
	case "s3":
		return s.deleteFromS3(fileURL)
	case "gcs":
		return s.deleteFromGCS(fileURL)
	default:
		return fmt.Errorf("unsupported storage type: %s", s.storageType)
	}
}

func (s *uploadService) deleteFromLocal(fileURL string) error {
	// Extract file path from URL
	filePath := strings.TrimPrefix(fileURL, s.baseURL+"/")
	fullPath := filepath.Join(s.basePath, filePath)

	return utils.DeleteFile(fullPath)
}

func (s *uploadService) deleteFromS3(fileURL string) error {
	// TODO: Implement S3 delete
	return fmt.Errorf("S3 delete not implemented")
}

func (s *uploadService) deleteFromGCS(fileURL string) error {
	// TODO: Implement GCS delete
	return fmt.Errorf("GCS delete not implemented")
}

func (s *uploadService) GetFileInfo(fileURL string) (*FileInfo, error) {
	// Extract file info from URL and storage metadata
	fileName := filepath.Base(fileURL)

	var size int64
	var err error

	switch s.storageType {
	case "local":
		filePath := strings.TrimPrefix(fileURL, s.baseURL+"/")
		fullPath := filepath.Join(s.basePath, filePath)
		size, err = utils.GetFileSize(fullPath)
		if err != nil {
			return nil, err
		}
	default:
		// For cloud storage, this would query the storage service
		size = 0
	}

	info := &FileInfo{
		Name:       fileName,
		Size:       size,
		MimeType:   utils.DetectContentType(nil, fileName),
		URL:        fileURL,
		UploadedAt: time.Now(), // Would be retrieved from storage metadata
	}

	return info, nil
}

func (s *uploadService) GeneratePresignedURL(filePath string, expiration time.Duration) (string, error) {
	switch s.storageType {
	case "local":
		// For local storage, just return the direct URL
		return s.baseURL + "/" + filePath, nil
	case "s3":
		// TODO: Generate S3 presigned URL
		return "", fmt.Errorf("S3 presigned URLs not implemented")
	case "gcs":
		// TODO: Generate GCS signed URL
		return "", fmt.Errorf("GCS signed URLs not implemented")
	default:
		return "", fmt.Errorf("unsupported storage type: %s", s.storageType)
	}
}

// Batch operations

func (s *uploadService) UploadMultipleFiles(files []*multipart.FileHeader, uploadType UploadType, entityID string) ([]string, error) {
	var urls []string
	var errors []error

	for _, file := range files {
		url, err := s.uploadFile(file, uploadType, entityID)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		urls = append(urls, url)
	}

	if len(errors) > 0 {
		return urls, fmt.Errorf("failed to upload %d out of %d files", len(errors), len(files))
	}

	return urls, nil
}

func (s *uploadService) DeleteMultipleFiles(fileURLs []string) error {
	var errors []error

	for _, url := range fileURLs {
		if err := s.DeleteFile(url); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete %d out of %d files", len(errors), len(fileURLs))
	}

	return nil
}

// Image processing

func (s *uploadService) ResizeImage(fileURL string, width, height int) (string, error) {
	// TODO: Implement image resizing
	// This would use an image processing library like imaging or imagemagick
	return "", fmt.Errorf("image resizing not implemented")
}

func (s *uploadService) CreateThumbnail(fileURL string) (string, error) {
	// TODO: Implement thumbnail creation
	// This would create a smaller version of the image
	return "", fmt.Errorf("thumbnail creation not implemented")
}

func (s *uploadService) CompressImage(fileURL string, quality int) (string, error) {
	// TODO: Implement image compression
	// This would compress the image to reduce file size
	return "", fmt.Errorf("image compression not implemented")
}

// Async processing methods

func (s *uploadService) compressImageAsync(fileURL string, quality int) {
	if _, err := s.CompressImage(fileURL, quality); err != nil {
		s.logger.Error().Err(err).Str("url", fileURL).Msg("Failed to compress image")
	}
}

func (s *uploadService) createThumbnailAsync(fileURL string) {
	if _, err := s.CreateThumbnail(fileURL); err != nil {
		s.logger.Error().Err(err).Str("url", fileURL).Msg("Failed to create thumbnail")
	}
}

func (s *uploadService) createVariantsAsync(fileURL string, variants []ImageVariant) {
	for _, variant := range variants {
		if _, err := s.ResizeImage(fileURL, variant.Width, variant.Height); err != nil {
			s.logger.Error().Err(err).
				Str("url", fileURL).
				Str("variant", variant.Name).
				Msg("Failed to create image variant")
		}
	}
}

// Helper functions

func (s *uploadService) generateFileName(originalName, entityID string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	randomID := primitive.NewObjectID().Hex()[:8]

	return fmt.Sprintf("%s_%d_%s%s", entityID, timestamp, randomID, ext)
}

func mergeTypes(types1, types2 map[string]bool) map[string]bool {
	merged := make(map[string]bool)

	for k, v := range types1 {
		merged[k] = v
	}

	for k, v := range types2 {
		merged[k] = v
	}

	return merged
}
