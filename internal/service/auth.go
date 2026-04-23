package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
)

// AuthService 认证服务
// @Summary 认证服务
// @Description 处理用户登录、注册、Token管理等认证相关业务
// @Tags 认证服务
type AuthService struct{}

// LoginRequest 登录请求
// @Summary 登录请求结构
// @Description 用户登录请求参数
// @Tags 认证服务
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// LoginResponse 登录响应
// @Summary 登录响应结构
// @Description 用户登录成功返回的信息
// @Tags 认证服务
type LoginResponse struct {
	UserID   uint   `json:"user_id"`   // 用户ID
	Username string `json:"username"` // 用户名
	Role     string `json:"role"`     // 用户角色
	Token    string `json:"token"`    // JWT Token
}

// Login 用户登录
// @Summary 用户登录
// @Description 验证用户名密码，生成JWT Token
// @Tags 认证服务
// @Param req LoginRequest true "登录请求"
// @Return *LoginResponse "登录响应"
// @Return error "错误信息"
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// 阶段二：这里先模拟登录逻辑，后续集成数据库
	// 模拟用户数据
	mockUsers := map[string]struct {
		ID       uint
		Password string
		Role     string
	}{
		"admin": {ID: 1, Password: "admin123", Role: "super_admin"},
		"user":  {ID: 2, Password: "user123", Role: "platform_admin"},
	}

	// 验证用户是否存在
	user, exists := mockUsers[req.Username]
	if !exists {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码（生产环境应该使用bcrypt等加密算法）
	if req.Password != user.Password {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成JWT Token
	token, err := utils.GenerateToken(user.ID, req.Username, user.Role)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	// 返回登录响应
	return &LoginResponse{
		UserID:   user.ID,
		Username: req.Username,
		Role:     user.Role,
		Token:    token,
	}, nil
}

// Logout 用户登出
// @Summary 用户登出
// @Description 处理用户登出逻辑
// @Tags 认证服务
// @Param ctx context.Context true "上下文"
// @Return error "错误信息"
func (s *AuthService) Logout(ctx context.Context) error {
	// 阶段二：这里可以添加Token黑名单等逻辑
	// 目前简单返回成功
	return nil
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Description 刷新即将过期的JWT Token
// @Tags 认证服务
// @Param token string true "原Token"
// @Return string "新Token"
// @Return error "错误信息"
func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	// 解析原Token
	_, err := utils.ParseToken(token)
	if err != nil {
		return "", errors.New("Token无效")
	}

	// 生成新Token
	newToken, err := utils.RefreshToken(token)
	if err != nil {
		return "", errors.New("刷新Token失败")
	}

	return newToken, nil
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 根据Token获取当前登录用户信息
// @Tags 认证服务
// @Param ctx context.Context true "上下文"
// @Return *model.User "用户信息"
// @Return error "错误信息"
func (s *AuthService) GetCurrentUser(ctx context.Context) (*model.User, error) {
	// 阶段二：这里先返回模拟数据，后续从数据库查询
	return &model.User{
		Username: "admin",
		Email:    "admin@fayhub.com",
		Role:     "super_admin",
		RealName: "系统管理员",
	}, nil
}