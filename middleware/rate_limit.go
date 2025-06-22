package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"ondrive/utils"

	"github.com/gin-gonic/gin"
)

// RateLimiter interface for different rate limiting strategies
type RateLimiter interface {
	Allow(key string) bool
	Reset(key string)
	GetLimit(key string) int
	GetRemaining(key string) int
	GetResetTime(key string) time.Time
}

// TokenBucketLimiter implements token bucket rate limiting
type TokenBucketLimiter struct {
	buckets map[string]*TokenBucket
	mutex   sync.RWMutex
	rate    int           // requests per window
	window  time.Duration // time window
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     int
	capacity   int
	refillRate int
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(rate int, window time.Duration) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets: make(map[string]*TokenBucket),
		rate:    rate,
		window:  window,
	}
}

// Allow checks if a request is allowed
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	tbl.mutex.RLock()
	bucket, exists := tbl.buckets[key]
	tbl.mutex.RUnlock()

	if !exists {
		tbl.mutex.Lock()
		bucket = &TokenBucket{
			tokens:     tbl.rate,
			capacity:   tbl.rate,
			refillRate: tbl.rate,
			lastRefill: time.Now(),
		}
		tbl.buckets[key] = bucket
		tbl.mutex.Unlock()
	}

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Seconds() * float64(bucket.refillRate) / tbl.window.Seconds())

	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		if bucket.tokens > bucket.capacity {
			bucket.tokens = bucket.capacity
		}
		bucket.lastRefill = now
	}

	// Check if request is allowed
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// Reset resets the rate limit for a key
func (tbl *TokenBucketLimiter) Reset(key string) {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()
	delete(tbl.buckets, key)
}

// GetLimit returns the rate limit
func (tbl *TokenBucketLimiter) GetLimit(key string) int {
	return tbl.rate
}

// GetRemaining returns remaining requests
func (tbl *TokenBucketLimiter) GetRemaining(key string) int {
	tbl.mutex.RLock()
	bucket, exists := tbl.buckets[key]
	tbl.mutex.RUnlock()

	if !exists {
		return tbl.rate
	}

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	return bucket.tokens
}

// GetResetTime returns when the bucket will be fully refilled
func (tbl *TokenBucketLimiter) GetResetTime(key string) time.Time {
	tbl.mutex.RLock()
	bucket, exists := tbl.buckets[key]
	tbl.mutex.RUnlock()

	if !exists {
		return time.Now()
	}

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	tokensNeeded := bucket.capacity - bucket.tokens
	if tokensNeeded <= 0 {
		return time.Now()
	}

	timeToRefill := time.Duration(float64(tokensNeeded)*tbl.window.Seconds()/float64(bucket.refillRate)) * time.Second
	return bucket.lastRefill.Add(timeToRefill)
}

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	Limiter      RateLimiter
	KeyGenerator func(*gin.Context) string
	ErrorHandler func(*gin.Context, RateLimiter, string)
	SkipPaths    []string
	SkipFunc     func(*gin.Context) bool
}

// DefaultKeyGenerator generates a key based on client IP
func DefaultKeyGenerator(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyGenerator generates a key based on user ID (for authenticated requests)
func UserKeyGenerator(c *gin.Context) string {
	userID := c.GetString("user_id")
	if userID != "" {
		return "user:" + userID
	}
	return "ip:" + c.ClientIP()
}

// APIKeyGenerator generates a key based on API key
func APIKeyGenerator(c *gin.Context) string {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return "api:" + apiKey
	}
	return DefaultKeyGenerator(c)
}

// DefaultErrorHandler handles rate limit exceeded errors
func DefaultErrorHandler(c *gin.Context, limiter RateLimiter, key string) {
	remaining := limiter.GetRemaining(key)
	limit := limiter.GetLimit(key)
	resetTime := limiter.GetResetTime(key)

	// Set rate limit headers
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
	c.Header("Retry-After", strconv.FormatInt(int64(resetTime.Sub(time.Now()).Seconds()), 10))

	utils.TooManyRequestsResponse(c)
	c.Abort()
}

// RateLimit returns a rate limiting middleware
func RateLimit() gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(100, time.Minute) // 100 requests per minute
	config := RateLimitConfig{
		Limiter:      limiter,
		KeyGenerator: DefaultKeyGenerator,
		ErrorHandler: DefaultErrorHandler,
		SkipPaths: []string{
			"/health",
			"/metrics",
		},
	}
	return RateLimitWithConfig(config)
}

// RateLimitWithConfig returns a rate limiting middleware with custom configuration
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	// Create skip paths map for faster lookup
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Skip rate limiting for specified paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Skip rate limiting if skip function returns true
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// Generate rate limit key
		key := config.KeyGenerator(c)

		// Check rate limit
		if !config.Limiter.Allow(key) {
			config.ErrorHandler(c, config.Limiter, key)
			return
		}

		// Set rate limit headers for successful requests
		remaining := config.Limiter.GetRemaining(key)
		limit := config.Limiter.GetLimit(key)
		resetTime := config.Limiter.GetResetTime(key)

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		c.Next()
	}
}

// APIRateLimit returns a rate limiting middleware for API endpoints
func APIRateLimit() gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(1000, time.Hour) // 1000 requests per hour for API
	config := RateLimitConfig{
		Limiter:      limiter,
		KeyGenerator: UserKeyGenerator,
		ErrorHandler: DefaultErrorHandler,
		SkipPaths: []string{
			"/health",
			"/metrics",
		},
	}
	return RateLimitWithConfig(config)
}

// AuthRateLimit returns a rate limiting middleware for authentication endpoints
func AuthRateLimit() gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(10, time.Minute) // 10 auth attempts per minute
	config := RateLimitConfig{
		Limiter:      limiter,
		KeyGenerator: DefaultKeyGenerator,
		ErrorHandler: func(c *gin.Context, limiter RateLimiter, key string) {
			remaining := limiter.GetRemaining(key)
			limit := limiter.GetLimit(key)
			resetTime := limiter.GetResetTime(key)

			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(resetTime.Sub(time.Now()).Seconds()), 10))

			utils.ErrorResponseWithCode(c, http.StatusTooManyRequests, "AUTH_RATE_LIMIT",
				"Too many authentication attempts",
				fmt.Sprintf("Please try again in %d seconds", int64(resetTime.Sub(time.Now()).Seconds())))
			c.Abort()
		},
	}
	return RateLimitWithConfig(config)
}

// AdminRateLimit returns a rate limiting middleware for admin endpoints
func AdminRateLimit() gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(500, time.Hour) // 500 requests per hour for admin
	config := RateLimitConfig{
		Limiter: limiter,
		KeyGenerator: func(c *gin.Context) string {
			adminID := c.GetString("admin_id")
			if adminID != "" {
				return "admin:" + adminID
			}
			return "ip:" + c.ClientIP()
		},
		ErrorHandler: DefaultErrorHandler,
		SkipPaths: []string{
			"/admin/health",
			"/admin/static",
		},
	}
	return RateLimitWithConfig(config)
}

// WebSocketRateLimit returns a rate limiting middleware for WebSocket connections
func WebSocketRateLimit() gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(50, time.Minute) // 50 WebSocket connections per minute
	config := RateLimitConfig{
		Limiter:      limiter,
		KeyGenerator: UserKeyGenerator,
		ErrorHandler: DefaultErrorHandler,
	}
	return RateLimitWithConfig(config)
}

// FileUploadRateLimit returns a rate limiting middleware for file upload endpoints
func FileUploadRateLimit() gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(20, time.Minute) // 20 file uploads per minute
	config := RateLimitConfig{
		Limiter:      limiter,
		KeyGenerator: UserKeyGenerator,
		ErrorHandler: func(c *gin.Context, limiter RateLimiter, key string) {
			remaining := limiter.GetRemaining(key)
			limit := limiter.GetLimit(key)
			resetTime := limiter.GetResetTime(key)

			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			utils.ErrorResponseWithCode(c, http.StatusTooManyRequests, "UPLOAD_RATE_LIMIT",
				"Too many file upload attempts",
				"Please wait before uploading more files")
			c.Abort()
		},
	}
	return RateLimitWithConfig(config)
}

// Recovery middleware with rate limiting for error recovery
func Recovery(logger utils.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error().
				Str("error", err).
				Str("path", c.Request.URL.Path).
				Str("method", c.Request.Method).
				Str("client_ip", c.ClientIP()).
				Msg("Panic recovered")
		}

		utils.InternalServerErrorResponse(c)
		c.Abort()
	})
}
