package main

import (
	"context"
	"fayhub/pkg/config"
	"fayhub/internal/initialize"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	log.Println("🚀 开始执行数据库迁移脚本...")

	// 加载配置文件
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置文件失败: %v", err)
	}

	// 初始化数据库连接
	db, err := initialize.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	// 执行数据库迁移
	if err := runMigrations(db); err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}

	// 初始化默认数据
	if err := initDefaultData(db); err != nil {
		log.Fatalf("❌ 默认数据初始化失败: %v", err)
	}

	log.Println("🎉 数据库迁移和初始化完成！")
}

// runMigrations 执行数据库表迁移
func runMigrations(db *gorm.DB) error {
	log.Println("📊 开始数据库表迁移...")

	ctx := utils.SkipTenantIsolation(context.Background())
	migrateDB := db.WithContext(ctx)

	models := []interface{}{
		&model.Tenant{},
		&model.User{},
		&model.TenantUser{},
		&model.Role{},
		&model.Menu{},
		&model.API{},
		&model.RoleMenu{},
		&model.RoleAPI{},
		&model.UserRole{},
		&model.TenantRole{},
	}

	for _, m := range models {
		if err := migrateDB.AutoMigrate(m); err != nil {
			return fmt.Errorf("迁移 %T 失败: %w", m, err)
		}
		log.Printf("✅ 迁移完成: %T", m)
	}

	log.Println("✅ 数据库表迁移完成")
	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData(db *gorm.DB) error {
	log.Println("📝 开始初始化默认数据...")

	ctx := utils.SkipTenantIsolation(context.Background())
	initDB := db.WithContext(ctx)

	var tenantCount int64
	initDB.Model(&model.Tenant{}).Count(&tenantCount)
	if tenantCount > 0 {
		log.Println("ℹ️  数据已存在，跳过默认数据初始化")
		return nil
	}

	defaultTenant := &model.Tenant{
		Name:        "平台租户",
		Domain:      "platform.fayhub.com",
		Description: "系统平台默认租户，用于管理平台级用户",
		Status:      1,
	}
	if err := initDB.Create(defaultTenant).Error; err != nil {
		return fmt.Errorf("创建默认租户失败: %w", err)
	}
	log.Println("✅ 默认租户创建完成")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	superAdmin := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "admin",
		Password:    string(hashedPassword),
		Email:       "admin@fayhub.com",
		Phone:       "13800138000",
		Status:      1,
		Role:        "super_admin",
		RealName:    "超级管理员",
	}
	if err := initDB.Create(superAdmin).Error; err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}
	log.Println("✅ 超级管理员创建完成")

	var superAdminRole model.Role
	if err := initDB.Where("name = ?", "超级管理员").First(&superAdminRole).Error; err != nil {
		return fmt.Errorf("查找超级管理员角色失败: %w", err)
	}

	userRole := model.UserRole{
		UserID: superAdmin.ID,
		RoleID: superAdminRole.ID,
	}
	if err := initDB.Create(&userRole).Error; err != nil {
		return fmt.Errorf("分配超级管理员角色失败: %w", err)
	}
	log.Println("✅ 超级管理员角色分配完成")

	roles := []model.Role{
		{
			Name:        "超级管理员",
			Description: "系统最高权限角色，可管理所有租户和用户",
			Type:        1,
			Status:      1,
		},
		{
			Name:        "平台管理员",
			Description: "平台管理员角色，可管理平台基础功能",
			Type:        1,
			Status:      1,
		},
		{
			Name:        "租户管理员",
			Description: "租户管理员角色，可管理本租户用户",
			Type:        2,
			Status:      1,
		},
		{
			Name:        "普通用户",
			Description: "普通用户角色，基础权限",
			Type:        2,
			Status:      1,
		},
	}

	var createdRoles []model.Role
	for _, role := range roles {
		if err := initDB.Create(&role).Error; err != nil {
			return fmt.Errorf("创建角色 %s 失败: %w", role.Name, err)
		}
		createdRoles = append(createdRoles, role)
		log.Printf("✅ 角色创建完成: %s", role.Name)
	}

	menus := []model.Menu{
		{
			Title:      "系统管理",
			Path:       "/system",
			Component:  "Layout",
			Icon:       "system",
			Sort:       1,
			Type:       1,
			Status:     1,
			Permission: "system:manage",
		},
		{
			Title:      "租户管理",
			Path:       "/tenant",
			Component:  "tenant/index",
			Icon:       "tenant",
			Sort:       2,
			Type:       2,
			Status:     1,
			Permission: "tenant:manage",
		},
		{
			Title:      "用户管理",
			Path:       "/user",
			Component:  "user/index",
			Icon:       "user",
			Sort:       3,
			Type:       2,
			Status:     1,
			Permission: "user:manage",
		},
		{
			Title:      "角色管理",
			Path:       "/role",
			Component:  "role/index",
			Icon:       "role",
			Sort:       4,
			Type:       2,
			Status:     1,
			Permission: "role:manage",
		},
		{
			Title:      "菜单管理",
			Path:       "/menu",
			Component:  "menu/index",
			Icon:       "menu",
			Sort:       5,
			Type:       2,
			Status:     1,
			Permission: "menu:manage",
		},
	}

	var createdMenus []model.Menu
	for _, menu := range menus {
		if err := initDB.Create(&menu).Error; err != nil {
			return fmt.Errorf("创建菜单 %s 失败: %w", menu.Title, err)
		}
		createdMenus = append(createdMenus, menu)
		log.Printf("✅ 菜单创建完成: %s", menu.Title)
	}

	for _, role := range createdRoles {
		switch role.Name {
		case "超级管理员":
			for _, menu := range createdMenus {
				roleMenu := model.RoleMenu{
					RoleID: role.ID,
					MenuID: menu.ID,
				}
				if err := initDB.Create(&roleMenu).Error; err != nil {
					return fmt.Errorf("创建角色菜单关联失败: %w", err)
				}
			}
			log.Printf("✅ 超级管理员菜单权限分配完成")
		case "租户管理员":
			for _, menu := range createdMenus {
				if menu.Title == "租户管理" || menu.Title == "用户管理" {
					roleMenu := model.RoleMenu{
						RoleID: role.ID,
						MenuID: menu.ID,
					}
					if err := initDB.Create(&roleMenu).Error; err != nil {
						return fmt.Errorf("创建角色菜单关联失败: %w", err)
					}
				}
			}
			log.Printf("✅ 租户管理员菜单权限分配完成")
		}
	}

	log.Println("✅ 默认数据初始化完成")
	return nil
}

func createTestTenant(db *gorm.DB) error {
	log.Println("🧪 开始创建测试租户...")

	ctx := utils.SkipTenantIsolation(context.Background())
	testDB := db.WithContext(ctx)

	testTenant := &model.Tenant{
		Name:        "测试租户",
		Domain:      "test.fayhub.com",
		Description: "用于测试的租户",
		Status:      1,
	}
	if err := testDB.Create(testTenant).Error; err != nil {
		return fmt.Errorf("创建测试租户失败: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	testAdmin := &model.User{
		TenantModel: model.TenantModel{TenantID: testTenant.ID},
		Username:    "test_admin",
		Password:    string(hashedPassword),
		Email:       "test_admin@fayhub.com",
		Phone:       "13800138001",
		Status:      1,
		Role:        "tenant_admin",
		RealName:    "测试管理员",
	}
	if err := testDB.Create(testAdmin).Error; err != nil {
		return fmt.Errorf("创建测试管理员失败: %w", err)
	}

	log.Println("✅ 测试租户创建完成")
	return nil
}
