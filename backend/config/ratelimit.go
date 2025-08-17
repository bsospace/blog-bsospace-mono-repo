package config

import (
	"time"
)

// RateLimitSettings holds all rate limiting configuration
type RateLimitSettings struct {
	// Global rate limiting
	Global struct {
		Enabled     bool          `json:"enabled"`
		Strategy    string        `json:"strategy"` // "token_bucket", "sliding_window", "fixed_window"
		MaxRequests int           `json:"max_requests"`
		WindowSize  time.Duration `json:"window_size"`
	} `json:"global"`

	// IP-based rate limiting
	IP struct {
		Enabled     bool          `json:"enabled"`
		Strategy    string        `json:"strategy"`
		MaxRequests int           `json:"max_requests"`
		WindowSize  time.Duration `json:"window_size"`
	} `json:"ip"`

	// API key-based rate limiting
	API struct {
		Enabled        bool          `json:"enabled"`
		Strategy       string        `json:"strategy"`
		MaxRequests    int           `json:"max_requests"`
		WindowSize     time.Duration `json:"window_size"`
		RefillRate     float64       `json:"refill_rate"`
		BucketCapacity int           `json:"bucket_capacity"`
	} `json:"api"`

	// Route-specific rate limiting
	Routes map[string]RouteRateLimit `json:"routes"`

	// Redis settings for distributed rate limiting
	Redis struct {
		Enabled    bool   `json:"enabled"`
		KeyPrefix  string `json:"key_prefix"`
		DefaultTTL int    `json:"default_ttl"` // in seconds
	} `json:"redis"`
}

// RouteRateLimit defines rate limiting for specific routes
type RouteRateLimit struct {
	Enabled        bool          `json:"enabled"`
	Strategy       string        `json:"strategy"`
	MaxRequests    int           `json:"max_requests"`
	WindowSize     time.Duration `json:"window_size"`
	RefillRate     float64       `json:"refill_rate,omitempty"`
	BucketCapacity int           `json:"bucket_capacity,omitempty"`
}

// DefaultRateLimitSettings returns default rate limiting configuration
func DefaultRateLimitSettings() *RateLimitSettings {
	settings := &RateLimitSettings{}

	// Global defaults
	settings.Global.Enabled = true
	settings.Global.Strategy = "sliding_window"
	settings.Global.MaxRequests = 1000
	settings.Global.WindowSize = 1 * time.Minute

	// IP-based defaults
	settings.IP.Enabled = true
	settings.IP.Strategy = "sliding_window"
	settings.IP.MaxRequests = 100
	settings.IP.WindowSize = 1 * time.Minute

	// API-based defaults
	settings.API.Enabled = true
	settings.API.Strategy = "token_bucket"
	settings.API.MaxRequests = 1000
	settings.API.WindowSize = 1 * time.Hour
	settings.API.RefillRate = 1000.0 / 3600.0 // 1000 requests per hour
	settings.API.BucketCapacity = 1000

	// Route-specific defaults
	settings.Routes = map[string]RouteRateLimit{
		"/api/v1/auth/login": {
			Enabled:     true,
			Strategy:    "sliding_window",
			MaxRequests: 5,
			WindowSize:  15 * time.Minute,
		},
		"/api/v1/auth/register": {
			Enabled:     true,
			Strategy:    "sliding_window",
			MaxRequests: 3,
			WindowSize:  1 * time.Hour,
		},
		"/api/v1/ai/chat": {
			Enabled:        true,
			Strategy:       "token_bucket",
			MaxRequests:    50,
			WindowSize:     1 * time.Hour,
			RefillRate:     50.0 / 3600.0,
			BucketCapacity: 50,
		},
		"/api/v1/media/upload": {
			Enabled:     true,
			Strategy:    "sliding_window",
			MaxRequests: 10,
			WindowSize:  1 * time.Hour,
		},
	}

	// Redis defaults
	settings.Redis.Enabled = true
	settings.Redis.KeyPrefix = "rate_limit"
	settings.Redis.DefaultTTL = 3600 // 1 hour

	return settings
}

// LoadRateLimitSettings loads rate limiting configuration from environment or uses defaults
func LoadRateLimitSettings() *RateLimitSettings {
	settings := DefaultRateLimitSettings()
	return settings
}
