package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents error details in API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta represents metadata for paginated responses
type Meta struct {
	Page        int   `json:"page,omitempty"`
	Limit       int   `json:"limit,omitempty"`
	Total       int64 `json:"total,omitempty"`
	TotalPages  int64 `json:"total_pages,omitempty"`
	HasNext     bool  `json:"has_next,omitempty"`
	HasPrevious bool  `json:"has_previous,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message, details string) {
	response := APIResponse{
		Success: false,
		Message: message,
		Error: &APIError{
			Code:    http.StatusText(statusCode),
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

// ErrorResponseWithCode sends an error response with custom error code
func ErrorResponseWithCode(c *gin.Context, statusCode int, errorCode, message, details string) {
	response := APIResponse{
		Success: false,
		Message: message,
		Error: &APIError{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

// PaginatedResponse sends a paginated response
func PaginatedResponse(c *gin.Context, statusCode int, message string, data interface{}, meta *Meta) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, errors map[string]string) {
	response := APIResponse{
		Success: false,
		Message: "Validation failed",
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "The provided data is invalid",
		},
		Data:      errors,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusBadRequest, response)
}

// NotFoundResponse sends a not found response
func NotFoundResponse(c *gin.Context, resource string) {
	ErrorResponseWithCode(c, http.StatusNotFound, "NOT_FOUND",
		resource+" not found", "The requested resource could not be found")
}

// UnauthorizedResponse sends an unauthorized response
func UnauthorizedResponse(c *gin.Context) {
	ErrorResponseWithCode(c, http.StatusUnauthorized, "UNAUTHORIZED",
		"Authentication required", "Please provide valid authentication credentials")
}

// ForbiddenResponse sends a forbidden response
func ForbiddenResponse(c *gin.Context) {
	ErrorResponseWithCode(c, http.StatusForbidden, "FORBIDDEN",
		"Access denied", "You don't have permission to access this resource")
}

// InternalServerErrorResponse sends an internal server error response
func InternalServerErrorResponse(c *gin.Context) {
	ErrorResponseWithCode(c, http.StatusInternalServerError, "INTERNAL_ERROR",
		"Internal server error", "An unexpected error occurred")
}

// BadRequestResponse sends a bad request response
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusBadRequest, "BAD_REQUEST", message, "")
}

// ConflictResponse sends a conflict response
func ConflictResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusConflict, "CONFLICT", message, "")
}

// TooManyRequestsResponse sends a rate limit exceeded response
func TooManyRequestsResponse(c *gin.Context) {
	ErrorResponseWithCode(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
		"Rate limit exceeded", "Too many requests. Please try again later")
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) *Meta {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
	hasPrevious := page > 1

	return &Meta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}
}

// GetPaginationParams extracts pagination parameters from request
func GetPaginationParams(c *gin.Context) PaginationParams {
	var params PaginationParams

	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.Limit = 10
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 10
	}

	return params
}

// EmptyDataResponse sends response with empty data
func EmptyDataResponse(c *gin.Context, message string) {
	SuccessResponse(c, http.StatusOK, message, []interface{}{})
}

// CreatedResponse sends a created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

// UpdatedResponse sends an updated response
func UpdatedResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// DeletedResponse sends a deleted response
func DeletedResponse(c *gin.Context, message string) {
	SuccessResponse(c, http.StatusOK, message, nil)
}

// NoContentResponse sends a no content response
func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
