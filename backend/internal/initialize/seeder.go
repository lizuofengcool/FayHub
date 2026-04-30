package initialize

import (
	"context"
	"errors"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"fmt"
	"log"

	"gorm.io/gorm"
)

func InitDefaultMenus(db *gorm.DB) error {
	if db == nil {
		return errors.New("数据库未连接")
	}

	ctx := utils.SkipTenantIsolation(context.Background())
	db = db.WithContext(ctx)

	var count int64
	db.Model(&model.Menu{}).Count(&count)
	if count > 0 {
		log.Println("菜单数据已存在，跳过初始化")
		return nil
	}

	menus := []model.Menu{
		{
			Title:    "仪表盘",
			Path:     "/dashboard",
			Icon:     "dashboard",
			Sort:     1,
			ParentID: 0,
			Type:     1,
			Status:   1,
		},
		{
			Title:    "系统管理",
			Path:     "/system",
			Icon:     "setting",
			Sort:     2,
			ParentID: 0,
			Type:     1,
			Status:   1,
		},
		{
			Title:    "用户管理",
			Path:     "/system/user",
			Icon:     "user",
			Sort:     1,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "角色管理",
			Path:     "/system/role",
			Icon:     "role",
			Sort:     2,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "菜单管理",
			Path:     "/system/menu",
			Icon:     "menu",
			Sort:     3,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "API管理",
			Path:     "/system/api",
			Icon:     "api",
			Sort:     4,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "租户管理",
			Path:     "/system/tenant",
			Icon:     "tenant",
			Sort:     5,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "插件生态",
			Path:     "/plugins",
			Icon:     "Box",
			Sort:     3,
			ParentID: 0,
			Type:     1,
			Status:   1,
		},
		{
			Title:    "插件管理",
			Path:     "/plugins/installed",
			Icon:     "Box",
			Sort:     1,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "插件引擎",
			Path:     "/plugins/engine",
			Icon:     "Monitor",
			Sort:     2,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},

		{
			Title:    "支付配置",
			Path:     "/payment",
			Icon:     "CreditCard",
			Sort:     4,
			ParentID: 0,
			Type:     1,
			Status:   1,
		},
		{
			Title:    "支付参数配置",
			Path:     "/payment/config",
			Icon:     "Setting",
			Sort:     1,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "交易记录",
			Path:     "/payment/transactions",
			Icon:     "List",
			Sort:     2,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "文件管理",
			Path:     "/system/files",
			Icon:     "Folder",
			Sort:     6,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
		{
			Title:    "部门管理",
			Path:     "/system/department",
			Icon:     "OfficeBuilding",
			Sort:     7,
			ParentID: 0,
			Type:     2,
			Status:   1,
		},
	}

	if err := db.Create(&menus).Error; err != nil {
		return fmt.Errorf("创建默认菜单失败: %v", err)
	}

	var systemMenu model.Menu
	if err := db.Where("path = ?", "/system").First(&systemMenu).Error; err == nil {
		var subMenus []model.Menu
		db.Where("path IN ?", []string{"/system/user", "/system/role", "/system/menu", "/system/api", "/system/tenant", "/system/files", "/system/department"}).Find(&subMenus)
		for i := range subMenus {
			db.Model(&subMenus[i]).Update("parent_id", systemMenu.ID)
		}
	}

	var pluginMenu model.Menu
	if err := db.Where("path = ?", "/plugins").First(&pluginMenu).Error; err == nil {
		var pluginSubMenus []model.Menu
		db.Where("path IN ?", []string{"/plugins/installed", "/plugins/engine"}).Find(&pluginSubMenus)
		for i := range pluginSubMenus {
			db.Model(&pluginSubMenus[i]).Update("parent_id", pluginMenu.ID)
		}
	}

	var paymentMenu model.Menu
	if err := db.Where("path = ?", "/payment").First(&paymentMenu).Error; err == nil {
		var paymentSubMenus []model.Menu
		db.Where("path IN ?", []string{"/payment/config", "/payment/transactions"}).Find(&paymentSubMenus)
		for i := range paymentSubMenus {
			db.Model(&paymentSubMenus[i]).Update("parent_id", paymentMenu.ID)
		}
	}

	log.Printf("默认菜单初始化完成，共创建 %d 条记录", len(menus))
	return nil
}

func InitDefaultAPIs(db *gorm.DB) error {
	if db == nil {
		return errors.New("数据库未连接")
	}

	ctx := utils.SkipTenantIsolation(context.Background())
	db = db.WithContext(ctx)

	var count int64
	db.Model(&model.API{}).Count(&count)
	if count > 0 {
		log.Println("API权限数据已存在，跳过初始化")
		return nil
	}

	apis := []model.API{
		{Path: "/api/auth/login", Method: "POST", Description: "用户登录", Group: "认证管理", Status: 1},
		{Path: "/api/auth/logout", Method: "POST", Description: "用户登出", Group: "认证管理", Status: 1},
		{Path: "/api/auth/refresh", Method: "POST", Description: "刷新Token", Group: "认证管理", Status: 1},
		{Path: "/api/auth/me", Method: "GET", Description: "获取当前用户信息", Group: "认证管理", Status: 1},

		{Path: "/api/users", Method: "POST", Description: "创建用户", Group: "用户管理", Status: 1},
		{Path: "/api/users", Method: "GET", Description: "获取用户列表", Group: "用户管理", Status: 1},
		{Path: "/api/users/profile", Method: "GET", Description: "获取当前用户个人信息", Group: "用户管理", Status: 1},
		{Path: "/api/users/change-password", Method: "PUT", Description: "修改密码", Group: "用户管理", Status: 1},
		{Path: "/api/users/:id", Method: "GET", Description: "获取用户详情", Group: "用户管理", Status: 1},
		{Path: "/api/users/:id", Method: "PUT", Description: "更新用户", Group: "用户管理", Status: 1},
		{Path: "/api/users/:id", Method: "DELETE", Description: "删除用户", Group: "用户管理", Status: 1},
		{Path: "/api/users/:id/reset-password", Method: "PUT", Description: "重置用户密码", Group: "用户管理", Status: 1},

		{Path: "/api/tenants", Method: "POST", Description: "创建租户", Group: "租户管理", Status: 1},
		{Path: "/api/tenants", Method: "GET", Description: "获取租户列表", Group: "租户管理", Status: 1},
		{Path: "/api/tenants/:id", Method: "GET", Description: "获取租户详情", Group: "租户管理", Status: 1},
		{Path: "/api/tenants/:id", Method: "PUT", Description: "更新租户", Group: "租户管理", Status: 1},
		{Path: "/api/tenants/:id", Method: "DELETE", Description: "删除租户", Group: "租户管理", Status: 1},

		{Path: "/api/rbac/roles", Method: "POST", Description: "创建角色", Group: "角色管理", Status: 1},
		{Path: "/api/rbac/roles", Method: "GET", Description: "获取角色列表", Group: "角色管理", Status: 1},
		{Path: "/api/rbac/roles/:roleID", Method: "GET", Description: "获取角色详情", Group: "角色管理", Status: 1},
		{Path: "/api/rbac/roles/:roleID", Method: "PUT", Description: "更新角色", Group: "角色管理", Status: 1},
		{Path: "/api/rbac/roles/:roleID", Method: "DELETE", Description: "删除角色", Group: "角色管理", Status: 1},
		{Path: "/api/rbac/roles/:roleID/users", Method: "GET", Description: "获取角色用户列表", Group: "角色管理", Status: 1},
		{Path: "/api/rbac/check-permission", Method: "POST", Description: "检查权限", Group: "角色管理", Status: 1},

		{Path: "/api/menus", Method: "POST", Description: "创建菜单", Group: "菜单管理", Status: 1},
		{Path: "/api/menus", Method: "GET", Description: "获取菜单列表", Group: "菜单管理", Status: 1},
		{Path: "/api/menus/tree", Method: "GET", Description: "获取菜单树", Group: "菜单管理", Status: 1},
		{Path: "/api/menus/:menuID", Method: "GET", Description: "获取菜单详情", Group: "菜单管理", Status: 1},
		{Path: "/api/menus/:menuID", Method: "PUT", Description: "更新菜单", Group: "菜单管理", Status: 1},
		{Path: "/api/menus/:menuID", Method: "DELETE", Description: "删除菜单", Group: "菜单管理", Status: 1},
		{Path: "/api/menus/assign-role", Method: "POST", Description: "分配角色菜单", Group: "菜单管理", Status: 1},
		{Path: "/api/menus/role/:roleID", Method: "GET", Description: "获取角色菜单", Group: "菜单管理", Status: 1},

		{Path: "/api/apis", Method: "POST", Description: "创建API接口", Group: "API管理", Status: 1},
		{Path: "/api/apis", Method: "GET", Description: "获取API列表", Group: "API管理", Status: 1},
		{Path: "/api/apis/:apiID", Method: "GET", Description: "获取API详情", Group: "API管理", Status: 1},
		{Path: "/api/apis/:apiID", Method: "PUT", Description: "更新API接口", Group: "API管理", Status: 1},
		{Path: "/api/apis/:apiID", Method: "DELETE", Description: "删除API接口", Group: "API管理", Status: 1},
		{Path: "/api/apis/assign-role", Method: "POST", Description: "分配角色API权限", Group: "API管理", Status: 1},
		{Path: "/api/apis/role/:roleID", Method: "GET", Description: "获取角色API权限", Group: "API管理", Status: 1},
	}

	if err := db.Create(&apis).Error; err != nil {
		return fmt.Errorf("创建默认API权限数据失败: %v", err)
	}

	log.Printf("默认API权限初始化完成，共创建 %d 条记录", len(apis))
	return nil
}
