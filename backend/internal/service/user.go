package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
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
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var existing model.User
	if err := db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errs.NewServiceError(errs.ErrUserAlreadyExist, "")
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, errs.NewServiceError(errs.ErrPasswordWeak, err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "密码加密失败")
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
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建用户失败")
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id uint, req UpdateUserRequest) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
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
			return nil, errs.NewServiceError(errs.ErrDatabase, "更新用户失败")
		}
	}

	db.First(&user, id)
	return &user, nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	return db.Delete(&user).Error
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	return &user, nil
}

func (s *UserService) GetList(ctx context.Context, req UserListRequest) (*UserListResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := db.Model(&model.User{})

	if req.Keyword != "" {
		keyword := utils.EscapeLike(req.Keyword)
		query = query.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户总数失败")
	}

	var users []model.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户列表失败")
	}

	return &UserListResponse{
		List:  users,
		Total: total,
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errs.NewServiceError(errs.ErrOldPasswordWrong, "")
	}

	if err := utils.ValidatePassword(newPassword); err != nil {
		return errs.NewServiceError(errs.ErrPasswordWeak, err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, "密码加密失败")
	}

	return db.Model(&user).Update("password", string(hashedPassword)).Error
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (s *UserService) ResetPassword(ctx context.Context, id uint, newPassword string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if err := utils.ValidatePassword(newPassword); err != nil {
		return errs.NewServiceError(errs.ErrPasswordWeak, err.Error())
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewServiceError(errs.ErrInternalServer, "密码加密失败")
	}

	return db.Model(&user).Update("password", string(hashedPassword)).Error
}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrUserNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户失败")
	}

	return &user, nil
}
