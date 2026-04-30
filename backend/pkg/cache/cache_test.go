package cache

import (
	"context"
	"testing"
	"time"
)

func TestBuildKey(t *testing.T) {
	key := BuildKey("tenant", "1", "config")
	expected := "fayhub:tenant:1:config"
	if key != expected {
		t.Errorf("BuildKey = %q, expected %q", key, expected)
	}
}

func TestCacheManagerNilRedis(t *testing.T) {
	cm := &CacheManager{rdb: nil}

	err := cm.Get(context.Background(), "test", nil)
	if err == nil {
		t.Error("Expected error for nil redis, got nil")
	}

	err = cm.Set(context.Background(), "test", "value", time.Minute)
	if err != nil {
		t.Errorf("Set with nil redis should not error, got: %v", err)
	}

	err = cm.Delete(context.Background(), "test")
	if err != nil {
		t.Errorf("Delete with nil redis should not error, got: %v", err)
	}
}

func TestInitCache(t *testing.T) {
	InitCache(nil)
	if GlobalCache == nil {
		t.Error("GlobalCache should not be nil after InitCache")
	}
}
