package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// Rate limiting strategy: "token_bucket", "sliding_window", "fixed_window"
	Strategy string
	// Maximum requests allowed per time window
	MaxRequests int
	// Time window duration
	WindowSize time.Duration
	// For token bucket: refill rate per second
	RefillRate float64
	// For token bucket: bucket capacity
	BucketCapacity int
	// Whether to use Redis for distributed rate limiting
	UseRedis bool
	// Redis client (required if UseRedis is true)
	RedisClient *redis.Client
	// Key prefix for Redis keys
	RedisKeyPrefix string
	// Logger instance
	Logger *zap.Logger
}

// RateLimiter interface for different rate limiting strategies
type RateLimiter interface {
	Allow(key string) bool
	Reset(key string)
}

// TokenBucketRateLimiter implements token bucket algorithm
type TokenBucketRateLimiter struct {
	capacity    int
	refillRate  float64
	tokens      map[string]float64
	lastRefill  map[string]time.Time
	mutex       sync.RWMutex
	redisClient *redis.Client
	useRedis    bool
	keyPrefix   string
	logger      *zap.Logger
}

// SlidingWindowRateLimiter implements sliding window algorithm
type SlidingWindowRateLimiter struct {
	maxRequests int
	windowSize  time.Duration
	requests    map[string][]time.Time
	mutex       sync.RWMutex
	redisClient *redis.Client
	useRedis    bool
	keyPrefix   string
	logger      *zap.Logger
}

// FixedWindowRateLimiter implements fixed window algorithm
type FixedWindowRateLimiter struct {
	maxRequests int
	windowSize  time.Duration
	requests    map[string]int
	windowStart map[string]time.Time
	mutex       sync.RWMutex
	redisClient *redis.Client
	useRedis    bool
	keyPrefix   string
	logger      *zap.Logger
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter
func NewTokenBucketRateLimiter(config RateLimitConfig) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		capacity:    config.BucketCapacity,
		refillRate:  config.RefillRate,
		tokens:      make(map[string]float64),
		lastRefill:  make(map[string]time.Time),
		redisClient: config.RedisClient,
		useRedis:    config.UseRedis,
		keyPrefix:   config.RedisKeyPrefix,
		logger:      config.Logger,
	}
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter
func NewSlidingWindowRateLimiter(config RateLimitConfig) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		maxRequests: config.MaxRequests,
		windowSize:  config.WindowSize,
		requests:    make(map[string][]time.Time),
		redisClient: config.RedisClient,
		useRedis:    config.UseRedis,
		keyPrefix:   config.RedisKeyPrefix,
		logger:      config.Logger,
	}
}

// NewFixedWindowRateLimiter creates a new fixed window rate limiter
func NewFixedWindowRateLimiter(config RateLimitConfig) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		maxRequests: config.MaxRequests,
		windowSize:  config.WindowSize,
		requests:    make(map[string]int),
		windowStart: make(map[string]time.Time),
		redisClient: config.RedisClient,
		useRedis:    config.UseRedis,
		keyPrefix:   config.RedisKeyPrefix,
		logger:      config.Logger,
	}
}

// Allow checks if a request is allowed for the given key
func (tb *TokenBucketRateLimiter) Allow(key string) bool {
	if tb.useRedis {
		return tb.allowRedis(key)
	}
	return tb.allowLocal(key)
}

// allowLocal implements local token bucket logic
func (tb *TokenBucketRateLimiter) allowLocal(key string) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()

	// Initialize if first time
	if _, exists := tb.tokens[key]; !exists {
		tb.tokens[key] = float64(tb.capacity)
		tb.lastRefill[key] = now
	}

	// Calculate time since last refill
	timePassed := now.Sub(tb.lastRefill[key]).Seconds()

	// Refill tokens
	tokensToAdd := timePassed * tb.refillRate
	tb.tokens[key] = min(float64(tb.capacity), tb.tokens[key]+tokensToAdd)
	tb.lastRefill[key] = now

	// Check if we have enough tokens
	if tb.tokens[key] >= 1.0 {
		tb.tokens[key] -= 1.0
		return true
	}

	return false
}

// allowRedis implements Redis-based token bucket logic
func (tb *TokenBucketRateLimiter) allowRedis(key string) bool {
	redisKey := tb.keyPrefix + ":token_bucket:" + key
	now := time.Now()

	// Use Lua script for atomic operations
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local current = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(current[1]) or capacity
		local last_refill = tonumber(current[2]) or now
		
		local time_passed = now - last_refill
		local tokens_to_add = time_passed * refill_rate
		tokens = math.min(capacity, tokens + tokens_to_add)
		
		if tokens >= 1.0 then
			tokens = tokens - 1.0
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 3600) -- 1 hour TTL
			return 1
		end
		
		return 0
	`

	result, err := tb.redisClient.Eval(context.Background(), script, []string{redisKey},
		tb.capacity, tb.refillRate, now.Unix()).Result()

	if err != nil {
		tb.logger.Error("Redis rate limit error", zap.Error(err))
		return true // Allow on error
	}

	return result.(int64) == 1
}

// Allow checks if a request is allowed for the given key
func (sw *SlidingWindowRateLimiter) Allow(key string) bool {
	if sw.useRedis {
		return sw.allowRedis(key)
	}
	return sw.allowLocal(key)
}

// allowLocal implements local sliding window logic
func (sw *SlidingWindowRateLimiter) allowLocal(key string) bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.windowSize)

	// Initialize if first time
	if _, exists := sw.requests[key]; !exists {
		sw.requests[key] = make([]time.Time, 0)
	}

	// Remove old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range sw.requests[key] {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	sw.requests[key] = validRequests

	// Check if we can add another request
	if len(sw.requests[key]) < sw.maxRequests {
		sw.requests[key] = append(sw.requests[key], now)
		return true
	}

	return false
}

// allowRedis implements Redis-based sliding window logic
func (sw *SlidingWindowRateLimiter) allowRedis(key string) bool {
	redisKey := sw.keyPrefix + ":sliding_window:" + key
	now := time.Now()
	cutoff := now.Add(-sw.windowSize).Unix()

	// Use Redis sorted set to track requests
	script := `
		local key = KEYS[1]
		local cutoff = tonumber(ARGV[1])
		local max_requests = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		-- Remove old entries
		redis.call('ZREMRANGEBYSCORE', key, 0, cutoff)
		
		-- Count current requests
		local count = redis.call('ZCARD', key)
		
		if count < max_requests then
			-- Add new request
			redis.call('ZADD', key, now, now .. ':' .. math.random())
			redis.call('EXPIRE', key, 3600) -- 1 hour TTL
			return 1
		end
		
		return 0
	`

	result, err := sw.redisClient.Eval(context.Background(), script, []string{redisKey},
		cutoff, sw.maxRequests, now.Unix()).Result()

	if err != nil {
		sw.logger.Error("Redis rate limit error", zap.Error(err))
		return true // Allow on error
	}

	return result.(int64) == 1
}

// Allow checks if a request is allowed for the given key
func (fw *FixedWindowRateLimiter) Allow(key string) bool {
	if fw.useRedis {
		return fw.allowRedis(key)
	}
	return fw.allowLocal(key)
}

// allowLocal implements local fixed window logic
func (fw *FixedWindowRateLimiter) allowLocal(key string) bool {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	now := time.Now()

	// Initialize if first time
	if _, exists := fw.windowStart[key]; !exists {
		fw.windowStart[key] = now
		fw.requests[key] = 0
	}

	// Check if we need to reset the window
	if now.Sub(fw.windowStart[key]) >= fw.windowSize {
		fw.windowStart[key] = now
		fw.requests[key] = 0
	}

	// Check if we can add another request
	if fw.requests[key] < fw.maxRequests {
		fw.requests[key]++
		return true
	}

	return false
}

// allowRedis implements Redis-based fixed window logic
func (fw *FixedWindowRateLimiter) allowRedis(key string) bool {
	redisKey := fw.keyPrefix + ":fixed_window:" + key
	now := time.Now()

	script := `
		local key = KEYS[1]
		local window_size = tonumber(ARGV[1])
		local max_requests = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local current = redis.call('HMGET', key, 'count', 'window_start')
		local count = tonumber(current[1]) or 0
		local window_start = tonumber(current[2]) or now
		
		-- Check if we need to reset the window
		if (now - window_start) >= window_size then
			count = 0
			window_start = now
		end
		
		-- Check if we can add another request
		if count < max_requests then
			count = count + 1
			redis.call('HMSET', key, 'count', count, 'window_start', window_start)
			redis.call('EXPIRE', key, 3600) -- 1 hour TTL
			return 1
		end
		
		return 0
	`

	result, err := fw.redisClient.Eval(context.Background(), script, []string{redisKey},
		int(fw.windowSize.Seconds()), fw.maxRequests, now.Unix()).Result()

	if err != nil {
		fw.logger.Error("Redis rate limit error", zap.Error(err))
		return true // Allow on error
	}

	return result.(int64) == 1
}

// Reset resets the rate limit for the given key
func (tb *TokenBucketRateLimiter) Reset(key string) {
	if tb.useRedis {
		tb.resetRedis(key)
	} else {
		tb.resetLocal(key)
	}
}

func (tb *TokenBucketRateLimiter) resetLocal(key string) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	delete(tb.tokens, key)
	delete(tb.lastRefill, key)
}

func (tb *TokenBucketRateLimiter) resetRedis(key string) {
	redisKey := tb.keyPrefix + ":token_bucket:" + key
	tb.redisClient.Del(context.Background(), redisKey)
}

func (sw *SlidingWindowRateLimiter) Reset(key string) {
	if sw.useRedis {
		sw.resetRedis(key)
	} else {
		sw.resetLocal(key)
	}
}

func (sw *SlidingWindowRateLimiter) resetLocal(key string) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	delete(sw.requests, key)
}

func (sw *SlidingWindowRateLimiter) resetRedis(key string) {
	redisKey := sw.keyPrefix + ":sliding_window:" + key
	sw.redisClient.Del(context.Background(), redisKey)
}

func (fw *FixedWindowRateLimiter) Reset(key string) {
	if fw.useRedis {
		fw.resetRedis(key)
	} else {
		fw.resetLocal(key)
	}
}

func (fw *FixedWindowRateLimiter) resetLocal(key string) {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()
	delete(fw.requests, key)
	delete(fw.windowStart, key)
}

func (fw *FixedWindowRateLimiter) resetRedis(key string) {
	redisKey := fw.keyPrefix + ":fixed_window:" + key
	fw.redisClient.Del(context.Background(), redisKey)
}

// Helper function
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	var limiter RateLimiter

	switch config.Strategy {
	case "token_bucket":
		limiter = NewTokenBucketRateLimiter(config)
	case "sliding_window":
		limiter = NewSlidingWindowRateLimiter(config)
	case "fixed_window":
		limiter = NewFixedWindowRateLimiter(config)
	default:
		config.Logger.Error("Invalid rate limiting strategy", zap.String("strategy", config.Strategy))
		// Default to sliding window
		limiter = NewSlidingWindowRateLimiter(config)
	}

	return func(c *gin.Context) {
		// Get client identifier (IP address or custom header)
		clientID := getClientID(c)

		if !limiter.Allow(clientID) {
			config.Logger.Warn("Rate limit exceeded",
				zap.String("client_id", clientID),
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests, please try again later",
				"retry_after": int(config.WindowSize.Seconds()),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(config.MaxRequests)))
		c.Header("X-RateLimit-Remaining", "1") // Simplified for now
		c.Header("X-RateLimit-Reset", time.Now().Add(config.WindowSize).Format(time.RFC1123))

		c.Next()
	}
}

// getClientID extracts client identifier from request
func getClientID(c *gin.Context) string {
	// Try to get from custom header first (for API keys)
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return "api:" + apiKey
	}

	// Fall back to IP address
	return c.ClientIP()
}

// GlobalRateLimit applies rate limiting to all routes
func GlobalRateLimit(config RateLimitConfig) gin.HandlerFunc {
	return RateLimitMiddleware(config)
}

// RouteRateLimit applies rate limiting to specific routes
func RouteRateLimit(config RateLimitConfig) gin.HandlerFunc {
	return RateLimitMiddleware(config)
}

// IPRateLimit applies IP-based rate limiting
func IPRateLimit(maxRequests int, windowSize time.Duration, redisClient *redis.Client, logger *zap.Logger) gin.HandlerFunc {
	config := RateLimitConfig{
		Strategy:       "sliding_window",
		MaxRequests:    maxRequests,
		WindowSize:     windowSize,
		UseRedis:       redisClient != nil,
		RedisClient:    redisClient,
		RedisKeyPrefix: "rate_limit:ip",
		Logger:         logger,
	}

	return RateLimitMiddleware(config)
}

// APIRateLimit applies API key-based rate limiting
func APIRateLimit(maxRequests int, windowSize time.Duration, redisClient *redis.Client, logger *zap.Logger) gin.HandlerFunc {
	config := RateLimitConfig{
		Strategy:       "token_bucket",
		MaxRequests:    maxRequests,
		WindowSize:     windowSize,
		RefillRate:     float64(maxRequests) / windowSize.Seconds(),
		BucketCapacity: maxRequests,
		UseRedis:       redisClient != nil,
		RedisClient:    redisClient,
		RedisKeyPrefix: "rate_limit:api",
		Logger:         logger,
	}

	return RateLimitMiddleware(config)
}
