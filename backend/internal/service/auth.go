package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/config"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/eventbus"
	"fayhub/pkg/tokenblacklist"
	"fayhub/pkg/utils"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	UserID   int64  `json:"user_id,string"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID int64  `json:"tenant_id,string"`
	Token    string `json:"token"`
}

type RegisterRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	RealName    string `json:"real_name"`
	CaptchaKey  string `json:"captcha_key" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

type RegisterResponse struct {
	UserID   int64  `json:"user_id,string"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID int64  `json:"tenant_id,string"`
	Token    string `json:"token"`
}

type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	loginCtx := utils.SkipTenantIsolation(ctx)
	loginDB := utils.GetDB(loginCtx)

	var user model.User
	if err := loginDB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrLoginFailed, "用户名或密码错误")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	if user.Status != 1 {
		return nil, errs.NewServiceError(errs.ErrAccountDisabled, "")
	}

	now := time.Now().Unix()
	if user.LockedUntil > 0 && now < user.LockedUntil {
		remaining := user.LockedUntil - now
		minutes := remaining / 60
		if minutes < 1 {
			minutes = 1
		}
		return nil, errs.NewServiceError(errs.ErrAccountLocked, fmt.Sprintf("账户已锁定，请在%d分钟后重试", minutes))
	}

	if user.LockedUntil > 0 && now >= user.LockedUntil {
		loginDB.Model(&user).Updates(map[string]interface{}{
			"login_fail_count": 0,
			"locked_until":     0,
		})
		user.LoginFailCount = 0
		user.LockedUntil = 0
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		eventbus.PublishAsync(eventbus.EventLoginFailed, user.TenantID, map[string]interface{}{
			"username": req.Username,
			"user_id":  user.ID,
		})

		maxAttempts := getMaxLoginAttempts()
		lockDuration := getLockDurationMin()

		newFailCount := user.LoginFailCount + 1
		updates := map[string]interface{}{
			"login_fail_count": newFailCount,
		}

		if newFailCount >= maxAttempts {
			updates["locked_until"] = now + int64(lockDuration*60)
			updates["login_fail_count"] = 0
			loginDB.Model(&user).Updates(updates)
			return nil, errs.NewServiceError(errs.ErrAccountLocked, fmt.Sprintf("密码错误次数过多，账户已锁定%d分钟", lockDuration))
		}

		loginDB.Model(&user).Updates(updates)
		remaining := maxAttempts - newFailCount
		return nil, errs.NewServiceError(errs.ErrLoginFailed, fmt.Sprintf("用户名或密码错误，还剩%d次尝试机会", remaining))
	}

	loginDB.Model(&user).Updates(map[string]interface{}{
		"login_fail_count": 0,
		"locked_until":     0,
		"last_login_at":    now,
	})

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, user.TenantID)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "生成Token失败")
	}

	eventbus.PublishAsync(eventbus.EventLoginSuccess, user.TenantID, map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
	})

	return &LoginResponse{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		TenantID: user.TenantID,
		Token:    token,
	}, nil
}

func getMaxLoginAttempts() int {
	if config.GlobalConfig != nil && config.GlobalConfig.Security.MaxLoginAttempts > 0 {
		return config.GlobalConfig.Security.MaxLoginAttempts
	}
	return 5
}

func getLockDurationMin() int {
	if config.GlobalConfig != nil && config.GlobalConfig.Security.LockDurationMin > 0 {
		return config.GlobalConfig.Security.LockDurationMin
	}
	return 15
}

func (s *AuthService) Logout(ctx context.Context) error {
	tokenString, ok := ctx.Value("token_string").(string)
	if !ok || tokenString == "" {
		return nil
	}

	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		return nil
	}

	return tokenblacklist.Add(ctx, tokenString, claims.ExpiresAt.Time)
}

func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	claims, err := utils.ParseToken(token)
	if err != nil {
		return "", errs.NewServiceError(errs.ErrTokenInvalid, "")
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return "", errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	refreshCtx := utils.SkipTenantIsolation(ctx)
	refreshDB := utils.GetDB(refreshCtx)

	var user model.User
	if err := refreshDB.First(&user, claims.UserID).Error; err != nil {
		return "", errs.NewServiceError(errs.ErrUserNotExist, "")
	}

	if user.Status != 1 {
		return "", errs.NewServiceError(errs.ErrAccountDisabled, "")
	}

	newToken, err := utils.GenerateToken(user.ID, user.Username, user.Role, user.TenantID)
	if err != nil {
		return "", errs.NewServiceError(errs.ErrInternalServer, "刷新Token失败")
	}

	_ = tokenblacklist.Add(ctx, token, claims.ExpiresAt.Time)

	return newToken, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var user model.User
	if err := queryDB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	return &user, nil
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	valid, _ := ServiceGroupApp.CaptchaService.Verify(ctx, req.CaptchaKey, req.CaptchaCode)
	if !valid {
		return nil, errs.NewServiceError(errs.ErrCaptchaFailed, "验证码错误或已过期")
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	registerCtx := utils.SkipTenantIsolation(ctx)
	registerDB := utils.GetDB(registerCtx)

	var existing model.User
	if err := registerDB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errs.NewServiceError(errs.ErrUserAlreadyExist, "")
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, errs.NewServiceError(errs.ErrPasswordWeak, err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "密码加密失败")
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
		Role:     "tenant_user",
		Status:   1,
	}

	if err := registerDB.Create(user).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "注册失败")
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, user.TenantID)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "生成Token失败")
	}

	return &RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		TenantID: user.TenantID,
		Token:    token,
	}, nil
}
