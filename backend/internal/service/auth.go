package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var user model.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, errors.New("查询用户失败")
	}

	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	return &LoginResponse{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Token:    token,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context) error {
	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	claims, err := utils.ParseToken(token)
	if err != nil {
		return "", errors.New("Token无效")
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return "", errors.New("数据库未连接")
	}

	var user model.User
	if err := db.First(&user, claims.UserID).Error; err != nil {
		return "", errors.New("用户不存在")
	}

	if user.Status != 1 {
		return "", errors.New("用户已被禁用")
	}

	newToken, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", errors.New("刷新Token失败")
	}

	return newToken, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uint) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("查询用户失败")
	}

	return &user, nil
}
