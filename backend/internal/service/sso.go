package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/config"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/logger"
	redisclient "fayhub/pkg/redisclient"
	"fayhub/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SSOService struct{}

type SSOTokenExchangeRequest struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func getSSOClients() map[string]string {
	if config.GlobalConfig != nil && len(config.GlobalConfig.SSO.Clients) > 0 {
		return config.GlobalConfig.SSO.Clients
	}
	return map[string]string{}
}

type SSOTokenExchangeResponse struct {
	UserID   int64  `json:"user_id"`
	TenantID int64  `json:"tenant_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

const (
	ssoAuthCodePrefix = "sso:auth_code:"
	ssoTokenPrefix    = "sso:token:"
)

func (s *SSOService) GenerateAuthorizationCode(ctx context.Context) (string, error) {
	userID, _ := utils.GetUserIDFromContext(ctx)
	tenantID, _ := utils.GetTenantIDFromCtx(ctx)
	username, _ := utils.GetUsernameFromContext(ctx)
	role, _ := utils.GetRoleFromContext(ctx)

	if userID == 0 {
		return "", errs.NewServiceError(errs.ErrSSONotLoggedIn, "")
	}

	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return "", errs.NewServiceError(errs.ErrSSOCodeGenerate, "")
	}
	code := hex.EncodeToString(codeBytes)

	authCodeData := &model.SSOAuthorizationCode{
		Code:      code,
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	authCodeData.TenantID = tenantID

	if redisclient.IsEnabled() {
		if err := redisclient.Set(ctx, ssoAuthCodePrefix+code, authCodeData, 5*time.Minute); err != nil {
			return "", errs.NewServiceError(errs.ErrInternalServer, "存储授权码失败")
		}
	} else {
		db := utils.GetDB(utils.SkipTenantIsolation(ctx))
		if db == nil {
			return "", errs.NewServiceError(errs.ErrDBNotConnected, "")
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(authCodeData).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return "", errs.NewServiceError(errs.ErrInternalServer, "存储授权码失败")
		}
	}

	return code, nil
}

func (s *SSOService) ExchangeToken(ctx context.Context, code string, clientID string, clientSecret string) (*SSOTokenExchangeResponse, error) {
	// 校验 Client ID / Client Secret
	if clientID == "" || clientSecret == "" {
		return nil, errs.NewServiceError(errs.ErrParamValidation, "缺少 client_id 或 client_secret")
	}
	expectedSecret, ok := getSSOClients()[clientID]
	if !ok || expectedSecret != clientSecret {
		return nil, errs.NewServiceError(errs.ErrSSOCodeInvalid, "client_id 或 client_secret 无效")
	}

	var authCode model.SSOAuthorizationCode
	var found bool

	if redisclient.IsEnabled() {
		var cached model.SSOAuthorizationCode
		err := redisclient.Get(ctx, ssoAuthCodePrefix+code, &cached)
		if err == nil {
			authCode = cached
			found = true
			redisclient.Del(ctx, ssoAuthCodePrefix+code)
		}
	}

	if !found {
		db := utils.GetDB(utils.SkipTenantIsolation(ctx))
		if db == nil {
			return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
		}
		if err := db.Where("code = ? AND expires_at > ?", code, time.Now()).First(&authCode).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errs.NewServiceError(errs.ErrSSOCodeInvalid, "")
			}
			return nil, errs.NewServiceError(errs.ErrInternalServer, "查询授权码失败")
		}
		db.Delete(&authCode)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "生成SSO令牌失败")
	}
	ssoToken := hex.EncodeToString(tokenBytes)

	now := time.Now()
	tokenData := &model.SSOTokenData{
		Token:     ssoToken,
		UserID:    authCode.UserID,
		Username:  authCode.Username,
		Role:      authCode.Role,
		ExpiresAt: now.Add(30 * time.Minute),
	}
	tokenData.TenantID = authCode.TenantID

	if redisclient.IsEnabled() {
		if err := redisclient.Set(ctx, ssoTokenPrefix+ssoToken, tokenData, 30*time.Minute); err != nil {
			return nil, errs.NewServiceError(errs.ErrInternalServer, "存储SSO令牌失败")
		}
	} else {
		db := utils.GetDB(utils.SkipTenantIsolation(ctx))
		if db == nil {
			return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
		}
		if err := db.Create(tokenData).Error; err != nil {
			return nil, errs.NewServiceError(errs.ErrInternalServer, "存储SSO令牌失败")
		}
	}

	return &SSOTokenExchangeResponse{
		UserID:   authCode.UserID,
		TenantID: authCode.TenantID,
		Username: authCode.Username,
		Role:     authCode.Role,
	}, nil
}

func (s *SSOService) VerifyToken(ctx context.Context, token string, clientID string, clientSecret string) (bool, error) {
	if clientID == "" || clientSecret == "" {
		return false, errs.NewServiceError(errs.ErrParamValidation, "缺少 client_id 或 client_secret")
	}
	expectedSecret, ok := getSSOClients()[clientID]
	if !ok || expectedSecret != clientSecret {
		return false, errs.NewServiceError(errs.ErrSSOCodeInvalid, "client_id 或 client_secret 无效")
	}

	if redisclient.IsEnabled() {
		var tokenData model.SSOTokenData
		err := redisclient.Get(ctx, ssoTokenPrefix+token, &tokenData)
		if err != nil {
			return false, nil
		}
		return true, nil
	}

	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return false, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var tokenData model.SSOTokenData
	if err := db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&tokenData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, errs.NewServiceError(errs.ErrInternalServer, "查询SSO令牌失败")
	}

	return true, nil
}

func cleanupExpiredSSOData(db *gorm.DB) {
	db.Where("expires_at < ?", time.Now()).Delete(&model.SSOAuthorizationCode{})
	db.Where("expires_at < ?", time.Now()).Delete(&model.SSOTokenData{})
}

func (s *SSOService) StartCleanup() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(context.Background(), "SSO清理任务panic", zap.Any("error", r))
			}
		}()
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if redisclient.IsEnabled() {
				continue
			}
			db := utils.GetDB(utils.SkipTenantIsolation(context.Background()))
			if db != nil {
				cleanupExpiredSSOData(db)
			}
		}
	}()
}
