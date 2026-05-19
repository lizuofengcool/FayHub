package initialize

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func InitDefaultAdmin(db *gorm.DB) error {
	if db == nil {
		return errors.New("数据库未连接")
	}

	ctx := utils.SkipTenantIsolation(context.Background())
	db = db.WithContext(ctx)

	var existingAdmin model.User
	if err := db.Where("username = ?", "admin").First(&existingAdmin).Error; err == nil {
		log.Println("默认超级管理员已存在，跳过初始化")
		return nil
	}

	adminPassword := os.Getenv("FAYHUB_DEFAULT_ADMIN_PASSWORD")
	if adminPassword == "" {
		var err error
		adminPassword, err = utils.GenerateRandomPassword(16)
		if err != nil {
			return fmt.Errorf("生成随机密码失败: %v", err)
		}
		log.Printf("⚠️  未设置 FAYHUB_DEFAULT_ADMIN_PASSWORD，已生成随机密码: %s（请妥善保存）", adminPassword)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	admin := model.User{
		SnowflakeTenantModel: model.SnowflakeTenantModel{TenantID: 0},
		Username:             "admin",
		Password:             string(hashedPassword),
		Email:                "admin@fayhub.com",
		Phone:                "13800000000",
		RealName:             "系统管理员",
		Status:               1,
		Role:                 "super_admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("创建超级管理员失败: %v", err)
	}

	superAdminRole := model.Role{
		Name:        "super_admin",
		Description: "超级管理员，拥有系统所有权限",
		Type:        1,
		Status:      1,
	}

	if err := db.Create(&superAdminRole).Error; err != nil {
		return fmt.Errorf("创建超级管理员角色失败: %v", err)
	}

	userRole := model.UserRole{
		UserID: admin.ID,
		RoleID: superAdminRole.ID,
	}

	if err := db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("分配角色失败: %v", err)
	}

	log.Println("默认超级管理员初始化完成 - 用户名: admin")
	return nil
}

func InitDefaultRoles(db *gorm.DB) error {
	if db == nil {
		return errors.New("数据库未连接")
	}

	ctx := utils.SkipTenantIsolation(context.Background())
	db = db.WithContext(ctx)

	defaultRoles := []model.Role{
		{
			Name:        "platform_admin",
			Description: "平台管理员，管理平台基础功能",
			Type:        1,
			Status:      1,
		},
		{
			Name:        "tenant_admin",
			Description: "租户管理员，管理本租户用户和业务",
			Type:        2,
			Status:      1,
		},
		{
			Name:        "tenant_user",
			Description: "租户普通用户，使用租户业务功能",
			Type:        2,
			Status:      1,
		},
	}

	for _, role := range defaultRoles {
		var existingRole model.Role
		if err := db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&role).Error; err != nil {
					return fmt.Errorf("创建角色 %s 失败: %v", role.Name, err)
				}
				log.Printf("创建默认角色: %s", role.Name)
			}
		}
	}

	return nil
}
