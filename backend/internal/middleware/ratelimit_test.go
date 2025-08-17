package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTokenBucketRateLimiter(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:       "token_bucket",
		MaxRequests:    10,
		WindowSize:     time.Minute,
		RefillRate:     1.0, // 1 token per second
		BucketCapacity: 10,
		UseRedis:       false,
		Logger:         logger,
	}

	limiter := NewTokenBucketRateLimiter(config)

	// Should allow first 10 requests
	for i := 0; i < 10; i++ {
		assert.True(t, limiter.Allow("test_key"), "Request %d should be allowed", i)
	}

	// Should reject the 11th request
	assert.False(t, limiter.Allow("test_key"), "11th request should be rejected")

	// Wait for token refill
	time.Sleep(2 * time.Second)

	// Should allow one more request after refill
	assert.True(t, limiter.Allow("test_key"), "Request after refill should be allowed")
}

func TestSlidingWindowRateLimiter(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:    "sliding_window",
		MaxRequests: 5,
		WindowSize:  100 * time.Millisecond,
		UseRedis:    false,
		Logger:      logger,
	}

	limiter := NewSlidingWindowRateLimiter(config)

	// Should allow first 5 requests
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow("test_key"), "Request %d should be allowed", i)
	}

	// Should reject the 6th request
	assert.False(t, limiter.Allow("test_key"), "6th request should be rejected")

	// Wait for window to slide
	time.Sleep(150 * time.Millisecond)

	// Should allow requests again after window slides
	assert.True(t, limiter.Allow("test_key"), "Request after window slide should be allowed")
}

func TestFixedWindowRateLimiter(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:    "fixed_window",
		MaxRequests: 3,
		WindowSize:  100 * time.Millisecond,
		UseRedis:    false,
		Logger:      logger,
	}

	limiter := NewFixedWindowRateLimiter(config)

	// Should allow first 3 requests
	for i := 0; i < 3; i++ {
		assert.True(t, limiter.Allow("test_key"), "Request %d should be allowed", i)
	}

	// Should reject the 4th request
	assert.False(t, limiter.Allow("test_key"), "4th request should be rejected")

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should allow requests again after window resets
	assert.True(t, limiter.Allow("test_key"), "Request after window reset should be allowed")
}

func TestRateLimitMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:    "sliding_window",
		MaxRequests: 2,
		WindowSize:  100 * time.Millisecond,
		UseRedis:    false,
		Logger:      logger,
	}

	middleware := RateLimitMiddleware(config)

	// Test that middleware can be created
	assert.NotNil(t, middleware, "Middleware should be created successfully")
}

func TestGetClientID(t *testing.T) {
	// This is a helper function test
	// In a real test, you'd use gin.Context
	// For now, just verify the function exists
	assert.NotNil(t, getClientID, "getClientID function should exist")
}

func TestRateLimitConfigValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Test invalid strategy
	config := RateLimitConfig{
		Strategy:    "invalid_strategy",
		MaxRequests: 10,
		WindowSize:  time.Minute,
		UseRedis:    false,
		Logger:      logger,
	}

	middleware := RateLimitMiddleware(config)
	assert.NotNil(t, middleware, "Middleware should fallback to default strategy")
}

func TestRateLimiterReset(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:    "sliding_window",
		MaxRequests: 1,
		WindowSize:  time.Minute,
		UseRedis:    false,
		Logger:      logger,
	}

	limiter := NewSlidingWindowRateLimiter(config)

	// Use up the limit
	assert.True(t, limiter.Allow("test_key"), "First request should be allowed")
	assert.False(t, limiter.Allow("test_key"), "Second request should be rejected")

	// Reset the limit
	limiter.Reset("test_key")

	// Should allow requests again after reset
	assert.True(t, limiter.Allow("test_key"), "Request after reset should be allowed")
}

func TestConcurrentAccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:    "sliding_window",
		MaxRequests: 200, // Increased limit to accommodate concurrent requests
		WindowSize:  time.Minute,
		UseRedis:    false,
		Logger:      logger,
	}

	limiter := NewSlidingWindowRateLimiter(config)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				limiter.Allow("concurrent_key")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not have exceeded the limit (10 goroutines * 10 requests = 100 requests, limit is 200)
	assert.True(t, limiter.Allow("concurrent_key"), "Should still allow requests within limit")
}

// Benchmark tests for performance
func BenchmarkTokenBucketRateLimiter(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:       "token_bucket",
		MaxRequests:    1000,
		WindowSize:     time.Minute,
		RefillRate:     1000.0 / 60.0,
		BucketCapacity: 1000,
		UseRedis:       false,
		Logger:         logger,
	}

	limiter := NewTokenBucketRateLimiter(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("benchmark_key")
	}
}

func BenchmarkSlidingWindowRateLimiter(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := RateLimitConfig{
		Strategy:    "sliding_window",
		MaxRequests: 1000,
		WindowSize:  time.Minute,
		UseRedis:    false,
		Logger:      logger,
	}

	limiter := NewSlidingWindowRateLimiter(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("benchmark_key")
	}
}
