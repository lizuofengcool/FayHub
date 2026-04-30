package ratelimit

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter("test", 5, time.Minute)
	if rl.key != "test" {
		t.Errorf("expected key 'test', got '%s'", rl.key)
	}
	if rl.maxTokens != 5 {
		t.Errorf("expected maxTokens 5, got %d", rl.maxTokens)
	}
	if rl.window != time.Minute {
		t.Errorf("expected window 1m, got %v", rl.window)
	}
}

func TestLocalRateLimiter_AllowWithinLimit(t *testing.T) {
	rl := NewRateLimiter("test-allow", 3, time.Minute)
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
}

func TestLocalRateLimiter_RejectOverLimit(t *testing.T) {
	rl := NewRateLimiter("test-reject", 2, time.Minute)
	rl.Allow()
	rl.Allow()
	if rl.Allow() {
		t.Error("request over limit should be rejected")
	}
}

func TestGetRateLimitConfig_Default(t *testing.T) {
	cfg := GetRateLimitConfig("nonexistent")
	if cfg.MaxTokens != 60 {
		t.Errorf("expected default maxTokens 60, got %d", cfg.MaxTokens)
	}
}

func TestGetRateLimitConfig_Login(t *testing.T) {
	cfg := GetRateLimitConfig("login")
	if cfg.MaxTokens != 5 {
		t.Errorf("expected login maxTokens 5, got %d", cfg.MaxTokens)
	}
}

func TestGetRateLimitConfig_API(t *testing.T) {
	cfg := GetRateLimitConfig("api")
	if cfg.MaxTokens != 100 {
		t.Errorf("expected api maxTokens 100, got %d", cfg.MaxTokens)
	}
}

func TestCheckRateLimit(t *testing.T) {
	result := CheckRateLimit("test-category", "127.0.0.1")
	if !result {
		t.Error("first request should be allowed")
	}
}
