package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type APIKeyService struct{}

type CreateAPIKeyRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Permissions []model.APIKeyPermission `json:"permissions"`
	RateLimit   int                      `json:"rate_limit"`
	ExpiresAt   *time.Time               `json:"expires_at"`
}

type APIKeyResponse struct {
	*model.APIKey
	Secret string `json:"secret"` // 仅在创建时返回
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	userID, ok := ctx.Value("user_id").(uint)
	if !ok {
		return nil, errs.NewServiceError(errs.ErrUnauthorized, "用户未登录")
	}

	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok {
		return nil, errs.NewServiceError(errs.ErrUnauthorized, "租户未识别")
	}

	// 生成随机密钥
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "生成密钥失败")
	}
	apiKey := hex.EncodeToString(keyBytes)

	// 生成密钥哈希
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	// 密钥前缀（用于显示）
	keyPrefix := apiKey[:8]

	// 序列化权限
	var permissionsStr string
	if len(req.Permissions) > 0 {
		permissionsBytes, err := json.Marshal(req.Permissions)
		if err != nil {
			return nil, errs.NewServiceError(errs.ErrParamValidation, "权限格式错误")
		}
		permissionsStr = string(permissionsBytes)
	}

	// 设置默认限流
	if req.RateLimit <= 0 {
		req.RateLimit = 1000
	}

	apiKeyModel := &model.APIKey{
		TenantID:    tenantID,
		UserID:      userID,
		Name:        req.Name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		Permissions: permissionsStr,
		RateLimit:   req.RateLimit,
		ExpiresAt:   req.ExpiresAt,
		Status:      1,
	}

	if err := db.Create(apiKeyModel).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建API密钥失败")
	}

	return &APIKeyResponse{
		APIKey: apiKeyModel,
		Secret: apiKey, // 仅在创建时返回明文密钥
	}, nil
}

func (s *APIKeyService) ValidateAPIKey(ctx context.Context, apiKey string) (*model.APIKey, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	// 计算密钥哈希
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	var key model.APIKey
	if err := db.Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrUnauthorized, "无效的API密钥")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API密钥失败")
	}

	if !key.IsActive() {
		return nil, errs.NewServiceError(errs.ErrUnauthorized, "API密钥已失效")
	}

	// 更新最后使用时间
	now := time.Now()
	db.Model(&key).Update("last_used_at", now)

	return &key, nil
}

func (s *APIKeyService) ListAPIKeys(ctx context.Context) ([]model.APIKey, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	userID, ok := ctx.Value("user_id").(uint)
	if !ok {
		return nil, errs.NewServiceError(errs.ErrUnauthorized, "用户未登录")
	}

	var keys []model.APIKey
	if err := db.Where("user_id = ?", userID).Find(&keys).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询API密钥列表失败")
	}

	return keys, nil
}

func (s *APIKeyService) DeleteAPIKey(ctx context.Context, keyID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	userID, ok := ctx.Value("user_id").(uint)
	if !ok {
		return errs.NewServiceError(errs.ErrUnauthorized, "用户未登录")
	}

	result := db.Where("id = ? AND user_id = ?", keyID, userID).Delete(&model.APIKey{})
	if result.Error != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除API密钥失败")
	}

	if result.RowsAffected == 0 {
		return errs.NewServiceError(errs.ErrResourceNotFound, "API密钥不存在")
	}

	return nil
}

func (s *APIKeyService) CheckPermission(key *model.APIKey, resource, action string) bool {
	if key.Permissions == "" {
		return true // 无权限限制，允许所有操作
	}

	var permissions []model.APIKeyPermission
	if err := json.Unmarshal([]byte(key.Permissions), &permissions); err != nil {
		return false
	}

	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}

	return false
}

func (s *APIKeyService) CheckRateLimit(ctx context.Context, keyID uint, limit int) (bool, error) {
	// 这里可以实现基于Redis的限流逻辑
	// 简化实现：直接返回true
	return true, nil
}
