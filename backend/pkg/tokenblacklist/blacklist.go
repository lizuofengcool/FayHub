package tokenblacklist

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/redisclient"
	"fayhub/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const prefix = "token_blacklist:"

type tokenClaims interface {
	GetExpiresAt() (*jwt.NumericDate, bool)
}

func hashToken(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(h[:])
}

func IsBlacklisted(ctx context.Context, tokenString string) bool {
	tokenHash := hashToken(tokenString)

	if redisclient.IsEnabled() {
		key := prefix + tokenHash
		var result string
		err := redisclient.Get(ctx, key, &result)
		if err == nil {
			return true
		}
	}

	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db != nil {
		var count int64
		db.Model(&model.TokenBlacklistEntry{}).
			Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
			Limit(1).
			Count(&count)
		return count > 0
	}

	return false
}

func Add(ctx context.Context, tokenString string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	tokenHash := hashToken(tokenString)

	if redisclient.IsEnabled() {
		key := prefix + tokenHash
		if err := redisclient.Set(ctx, key, "1", ttl); err != nil {
			fmt.Printf("Token黑名单Redis写入失败: %v，降级到数据库\n", err)
		} else {
			return nil
		}
	}

	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return fmt.Errorf("Token黑名单写入失败: Redis和数据库均不可用")
	}

	entry := model.TokenBlacklistEntry{
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	err := db.Where("token_hash = ?", tokenHash).FirstOrCreate(&entry).Error
	if err != nil {
		return fmt.Errorf("Token黑名单数据库写入失败: %w", err)
	}

	return nil
}

func CleanupExpired(ctx context.Context) error {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil
	}

	result := db.Where("expires_at < ?", time.Now()).Delete(&model.TokenBlacklistEntry{})
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	return nil
}
