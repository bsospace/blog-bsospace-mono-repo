package ai

import "testing"

func TestWebSearchRateLimiter(t *testing.T) {
	limiter := newPublicSearchRateLimiter()
	for i := 0; i < publicSearchLimit; i++ {
		if !limiter.allow("client") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if limiter.allow("client") {
		t.Fatal("request over the limit should be rejected")
	}
	if !limiter.allow("another-client") {
		t.Fatal("a different client should have its own window")
	}
}
