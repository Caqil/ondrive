package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
	AllowWildcard    bool
	AllowBrowserExt  bool
	AllowWebSockets  bool
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
			"X-HTTP-Method-Override",
			"User-Agent",
			"Referer",
			"Cache-Control",
			"X-Forwarded-For",
			"X-Real-IP",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 hours
		AllowWildcard:    true,
		AllowBrowserExt:  true,
		AllowWebSockets:  true,
	}
}

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		requestMethod := c.Request.Method

		// Handle preflight OPTIONS request
		if requestMethod == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			
			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			
			// Set allowed origins for preflight
			if isOriginAllowed(origin, config.AllowOrigins, config.AllowWildcard) {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if config.AllowWildcard && len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			}
			
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Set CORS headers for actual requests
		setCORSHeaders(c, origin, config)
		
		c.Next()
	}
}

// setCORSHeaders sets the appropriate CORS headers
func setCORSHeaders(c *gin.Context, origin string, config CORSConfig) {
	// Set Access-Control-Allow-Origin
	if isOriginAllowed(origin, config.AllowOrigins, config.AllowWildcard) {
		if config.AllowCredentials {
			// If credentials are allowed, we must specify the exact origin, not "*"
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}
	} else if config.AllowWildcard && len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
		if !config.AllowCredentials {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			// Don't set wildcard if credentials are allowed
			c.Header("Access-Control-Allow-Origin", origin)
		}
	}

	// Set Access-Control-Allow-Credentials
	if config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// Set Access-Control-Expose-Headers
	if len(config.ExposeHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
	}

	// Set Vary header to indicate that the response varies based on the Origin header
	varyHeader := c.Writer.Header().Get("Vary")
	if varyHeader == "" {
		c.Header("Vary", "Origin")
	} else if !strings.Contains(varyHeader, "Origin") {
		c.Header("Vary", varyHeader+", Origin")
	}
}

// isOriginAllowed checks if the origin is allowed
func isOriginAllowed(origin string, allowedOrigins []string, allowWildcard bool) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" && allowWildcard {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains like *.example.com
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}

// DevelopmentCORS returns a permissive CORS configuration for development
func DevelopmentCORS() gin.HandlerFunc {
	config := CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"http://localhost:8081",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:8081",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"*", // Allow all headers in development
		},
		ExposeHeaders: []string{
			"*", // Expose all headers in development
		},
		AllowCredentials: true,
		MaxAge:           3600,
		AllowWildcard:    true,
		AllowBrowserExt:  true,
		AllowWebSockets:  true,
	}
	return CORSWithConfig(config)
}

// ProductionCORS returns a secure CORS configuration for production
func ProductionCORS(allowedOrigins []string) gin.HandlerFunc {
	config := CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
			"Cache-Control",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
		AllowWildcard:    false,
		AllowBrowserExt:  false,
		AllowWebSockets:  true,
	}
	return CORSWithConfig(config)
}

// AdminCORS returns CORS configuration for admin panel
func AdminCORS() gin.HandlerFunc {
	config := CORSConfig{
		AllowOrigins: []string{
			"http://localhost:8081",
			"https://admin.indrive.com",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           3600,
		AllowWildcard:    false,
		AllowBrowserExt:  false,
		AllowWebSockets:  false,
	}
	return CORSWithConfig(config)
}
