package tokenblacklist

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("数据库初始化失败: %v", err)
	}
	ctx := utils.SkipTenantIsolation(context.Background())
	if err := db.WithContext(ctx).AutoMigrate(&model.TokenBlacklistEntry{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	utils.SetGlobalDB(db)
}

func TestAddAndCheck(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	token := "test_token_add_check"
	exists := IsBlacklisted(ctx, token)
	if exists {
		t.Fatal("token should not be blacklisted initially")
	}

	err := Add(ctx, token, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	exists = IsBlacklisted(ctx, token)
	if !exists {
		t.Fatal("token should be blacklisted after Add")
	}
}

func TestAddExpiredToken(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	token := "test_token_expired"
	err := Add(ctx, token, time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("Add with expired time should not error: %v", err)
	}

	exists := IsBlacklisted(ctx, token)
	if exists {
		t.Fatal("expired token should not be blacklisted")
	}
}

func TestMultipleTokens(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	token1 := "test_token_multi_1"
	token2 := "test_token_multi_2"

	err := Add(ctx, token1, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Add token1 failed: %v", err)
	}

	err = Add(ctx, token2, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Add token2 failed: %v", err)
	}

	if !IsBlacklisted(ctx, token1) {
		t.Fatal("token1 should be blacklisted")
	}
	if !IsBlacklisted(ctx, token2) {
		t.Fatal("token2 should be blacklisted")
	}
}

func TestNonBlacklistedToken(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	token := "test_token_nonexistent"
	if IsBlacklisted(ctx, token) {
		t.Fatal("non-existent token should not be blacklisted")
	}
}
