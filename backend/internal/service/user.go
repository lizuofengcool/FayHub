package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct{}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
	Status   *int   `json:"status"`
}

type UserListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	Role     string `json:"role" form:"role"`
	Status   *int   `json:"status" form:"status"`
}

type UserListResponse struct {
	List  []model.User `json:"list"`
	Total int64        `json:"total"`
}

func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var existing model.User
	if err := db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	role := req.Role
	if role == "" {
		role = "tenant_user"
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
		Role:     role,
		Status:   1,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id uint, req UpdateUserRequest) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	db.First(&user, id)
	return &user, nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errors.New("数据库未连接")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	return db.Delete(&user).Error
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetList(ctx context.Context, req UserListRequest) (*UserListResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errors.New("数据库未连接")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := db.Model(&model.User{})

	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []model.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, err
	}

	return &UserListResponse{
		List:  users,
		Total: total,
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errors.New("数据库未连接")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	return db.Model(&user).Update("password", string(hashedPassword)).Error
}
