package middleware

import (
	"fmt"
	"time"

	"ondrive/utils"

	"github.com/gin-gonic/gin"
)

type LoggerConfig struct {
	Logger        utils.Logger
	SkipPaths     []string
	LoggerFunc    gin.LoggerConfig
	EnableConsole bool
}

// Logger returns a gin.HandlerFunc (middleware) that logs requests using the provided logger
func Logger(logger utils.Logger) gin.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{
		Logger: logger,
		SkipPaths: []string{
			"/health",
			"/metrics",
		},
		EnableConsole: false,
	})
}

// LoggerWithConfig returns a gin.HandlerFunc using the specified logger config
func LoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// Don't log to console if we're using structured logging
			if !config.EnableConsole {
				return ""
			}
			// Use the default log format from gin
			return fmt.Sprintf("[GIN] %v | %3d | %13v | %15s |%-7s %#v\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
				param.ErrorMessage,
			)
		},
		Output:    gin.DefaultWriter,
		SkipPaths: config.SkipPaths,
	})
}

// StructuredLogger provides structured logging for HTTP requests
func StructuredLogger(logger utils.Logger) gin.HandlerFunc {
	return StructuredLoggerWithConfig(LoggerConfig{
		Logger: logger,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
	})
}

// StructuredLoggerWithConfig provides structured logging with custom configuration
func StructuredLoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Skip logging for specified paths
		if skipPaths[path] {
			return
		}

		// Calculate latency
		latency := time.Since(start)

		// Get request information
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()

		if raw != "" {
			path = path + "?" + raw
		}

		// Get user information if available
		userID := c.GetString("user_id")
		userRole := c.GetString("user_role")

		// Determine log level based on status code
		logEvent := config.Logger.Info()
		if statusCode >= 500 {
			logEvent = config.Logger.Error()
		} else if statusCode >= 400 {
			logEvent = config.Logger.Warn()
		}

		// Add structured fields
		logEvent = logEvent.
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("latency_human", latency.String()).
			Int("body_size", bodySize).
			Str("client_ip", clientIP).
			Str("user_agent", userAgent).
			Str("referer", referer)

		// Add user context if available
		if userID != "" {
			logEvent = logEvent.Str("user_id", userID)
		}
		if userRole != "" {
			logEvent = logEvent.Str("user_role", userRole)
		}

		// Add error information if present
		errors := c.Errors.ByType(gin.ErrorTypePrivate)
		if len(errors) > 0 {
			logEvent = logEvent.Str("errors", errors.String())
		}

		// Add custom request ID if present
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			logEvent = logEvent.Str("request_id", requestID)
		}

		// Log the request
		logEvent.Msg("HTTP Request")
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateID()
		}

		// Set the request ID in response headers
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Set the request ID in context for use by other middleware/handlers
		c.Set("request_id", requestID)

		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'")

		// Remove server information
		c.Header("Server", "")

		c.Next()
	}
}

// AdminSecurityHeaders adds additional security headers for admin panel
func AdminSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Additional security for admin panel
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "same-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Admin-specific headers
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	}
}
