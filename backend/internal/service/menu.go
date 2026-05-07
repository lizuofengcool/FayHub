package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"log"

	"gorm.io/gorm"
)

type MenuService struct{}

type CreateMenuRequest struct {
	ParentID   int64  `json:"parent_id"`
	Title      string `json:"title" binding:"required,min=1,max=100"`
	Path       string `json:"path" binding:"max=200"`
	Component  string `json:"component" binding:"max=200"`
	Icon       string `json:"icon" binding:"max=100"`
	Sort       int    `json:"sort"`
	Type       int    `json:"type" binding:"required,oneof=1 2 3"`
	Status     int    `json:"status" binding:"oneof=0 1"`
	Permission string `json:"permission" binding:"max=200"`
}

type UpdateMenuRequest struct {
	ParentID   int64  `json:"parent_id"`
	Title      string `json:"title" binding:"min=1,max=100"`
	Path       string `json:"path" binding:"max=200"`
	Component  string `json:"component" binding:"max=200"`
	Icon       string `json:"icon" binding:"max=100"`
	Sort       int    `json:"sort"`
	Type       int    `json:"type" binding:"omitempty,oneof=1 2 3"`
	Status     int    `json:"status" binding:"oneof=0 1"`
	Permission string `json:"permission" binding:"max=200"`
}

type GetMenuListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Title    string `form:"title"`
	Type     int    `form:"type" binding:"omitempty,oneof=0 1 2 3"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
}

type GetMenuListResponse struct {
	List  []model.Menu `json:"list"`
	Total int64        `json:"total"`
}

type AssignRoleMenuRequest struct {
	RoleID  int64   `json:"role_id" binding:"required"`
	MenuIDs []int64 `json:"menu_ids" binding:"required,min=1"`
}

func (s *MenuService) CreateMenu(ctx context.Context, req CreateMenuRequest) (*model.Menu, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	menu := model.Menu{
		ParentID:   req.ParentID,
		Title:      req.Title,
		Path:       req.Path,
		Component:  req.Component,
		Icon:       req.Icon,
		Sort:       req.Sort,
		Type:       req.Type,
		Status:     req.Status,
		Permission: req.Permission,
	}

	if err := db.Create(&menu).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建菜单失败")
	}

	return &menu, nil
}

func (s *MenuService) GetMenuList(ctx context.Context, req GetMenuListRequest) (*GetMenuListResponse, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.Menu{})

	if req.Title != "" {
		query = query.Where("title LIKE ?", "%"+req.Title+"%")
	}
	if req.Type > 0 {
		query = query.Where("type = ?", req.Type)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询菜单总数失败")
	}

	var menus []model.Menu
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询菜单列表失败")
	}

	return &GetMenuListResponse{
		List:  menus,
		Total: total,
	}, nil
}

func (s *MenuService) GetMenuTree(ctx context.Context) ([]model.Menu, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var menus []model.Menu
	if err := db.Where("status = 1").Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询菜单树失败")
	}

	userID, hasUser := utils.GetUserIDFromContext(ctx)
	role, hasRole := utils.GetRoleFromContext(ctx)
	isSuperAdmin := hasRole && (role == "super_admin" || role == "platform_admin")

	if hasUser && !isSuperAdmin {
		tenantDB := utils.GetDB(ctx)
		if tenantDB != nil {
			var roleIDs []int64
			var userRoles []model.UserRole
			if err := tenantDB.Where("user_id = ?", userID).Find(&userRoles).Error; err == nil && len(userRoles) > 0 {
				for _, ur := range userRoles {
					roleIDs = append(roleIDs, ur.RoleID)
				}
			} else if hasRole && role != "" {
				platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))
				if platformDB != nil {
					var roleModel model.Role
					if err := platformDB.Where("name = ?", role).First(&roleModel).Error; err == nil {
						roleIDs = append(roleIDs, roleModel.ID)
					}
				}
			}

			if len(roleIDs) > 0 {
				var roleMenus []model.RoleMenu
				if err := tenantDB.Where("role_id IN ?", roleIDs).Find(&roleMenus).Error; err == nil {
					if len(roleMenus) == 0 {
						roleMenus = s.autoAssignDefaultMenus(ctx, roleIDs, menus)
					}

					allowedMenuIDs := make(map[int64]bool)
					for _, rm := range roleMenus {
						allowedMenuIDs[rm.MenuID] = true
					}

					var filtered []model.Menu
					for _, m := range menus {
						if allowedMenuIDs[m.ID] {
							filtered = append(filtered, m)
						}
					}
					menus = filtered
				}
			}
		}
	}

	activePluginIDs := make(map[string]bool)
	tenantDB := utils.GetDB(ctx)
	if tenantDB != nil {
		var plugins []model.InstalledPlugin
		if err := tenantDB.Where("status = ?", "active").Select("plugin_id").Find(&plugins).Error; err == nil {
			for _, p := range plugins {
				activePluginIDs[p.PluginID] = true
			}
		}
	}

	var filtered []model.Menu
	for _, m := range menus {
		if m.Component != "" && !activePluginIDs[m.Component] {
			continue
		}
		filtered = append(filtered, m)
	}

	tenantID, _ := ctx.Value("tenant_id").(int64)
	if tenantID > 0 {
		allowedMenuIDs, err := TenantPackageServiceApp.GetTenantMenuIDs(ctx, tenantID)
		if err == nil && allowedMenuIDs != nil {
			allowedMenuMap := make(map[int64]bool)
			for _, id := range allowedMenuIDs {
				allowedMenuMap[id] = true
			}

			var filteredByPackage []model.Menu
			for _, m := range filtered {
				if allowedMenuMap[m.ID] {
					filteredByPackage = append(filteredByPackage, m)
				}
			}
			filtered = filteredByPackage
		}
	}

	menuMap := make(map[int64][]model.Menu)
	for i := range filtered {
		if filtered[i].ParentID != 0 {
			menuMap[filtered[i].ParentID] = append(menuMap[filtered[i].ParentID], filtered[i])
		}
	}

	var tree []model.Menu
	for i := range filtered {
		if filtered[i].ParentID == 0 {
			if children, ok := menuMap[filtered[i].ID]; ok {
				filtered[i].Children = children
			} else {
				filtered[i].Children = []model.Menu{}
			}
			tree = append(tree, filtered[i])
		}
	}

	return tree, nil
}

func (s *MenuService) GetMenuByID(ctx context.Context, menuID int64) (*model.Menu, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var menu model.Menu
	if err := db.First(&menu, menuID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrMenuNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询菜单失败")
	}

	return &menu, nil
}

func (s *MenuService) UpdateMenu(ctx context.Context, menuID int64, req UpdateMenuRequest) (*model.Menu, error) {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var menu model.Menu
	if err := db.First(&menu, menuID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrMenuNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询菜单失败")
	}

	menu.ParentID = req.ParentID
	if req.Title != "" {
		menu.Title = req.Title
	}
	menu.Path = req.Path
	menu.Component = req.Component
	menu.Icon = req.Icon
	menu.Sort = req.Sort
	if req.Type > 0 {
		menu.Type = req.Type
	}
	menu.Status = req.Status
	menu.Permission = req.Permission

	if err := db.Save(&menu).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "更新菜单失败")
	}

	return &menu, nil
}

func (s *MenuService) DeleteMenu(ctx context.Context, menuID int64) error {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var menu model.Menu
	if err := db.First(&menu, menuID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrMenuNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询菜单失败")
	}

	var childCount int64
	db.Model(&model.Menu{}).Where("parent_id = ?", menuID).Count(&childCount)
	if childCount > 0 {
		return errs.NewServiceError(errs.ErrForbidden, "存在子菜单，不可删除")
	}

	if err := db.Where("menu_id = ?", menuID).Delete(&model.RoleMenu{}).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除菜单角色关联失败")
	}

	if err := db.Delete(&menu).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除菜单失败")
	}

	return nil
}

func (s *MenuService) AssignRoleMenus(ctx context.Context, req AssignRoleMenuRequest) error {
	tenantDB := utils.GetDB(ctx)
	if tenantDB == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var role model.Role
	if err := tenantDB.First(&role, req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrRoleNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询角色失败")
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var existingMenus []model.Menu
	if err := platformDB.Where("id IN ?", req.MenuIDs).Find(&existingMenus).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "查询菜单失败")
	}
	if len(existingMenus) != len(req.MenuIDs) {
		return errs.NewServiceError(errs.ErrMenuNotExist, "部分菜单不存在")
	}

	if err := tenantDB.Where("role_id = ?", req.RoleID).Delete(&model.RoleMenu{}).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "清除角色菜单关联失败")
	}

	for _, menuID := range req.MenuIDs {
		roleMenu := model.RoleMenu{
			RoleID: req.RoleID,
			MenuID: menuID,
		}
		if err := tenantDB.Create(&roleMenu).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "分配菜单权限失败")
		}
	}

	return nil
}

func (s *MenuService) GetRoleMenus(ctx context.Context, roleID int64) ([]model.Menu, error) {
	tenantDB := utils.GetDB(ctx)
	if tenantDB == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var roleMenus []model.RoleMenu
	if err := tenantDB.Where("role_id = ?", roleID).Find(&roleMenus).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询角色菜单关联失败")
	}

	if len(roleMenus) == 0 {
		return []model.Menu{}, nil
	}

	var menuIDs []int64
	for _, rm := range roleMenus {
		menuIDs = append(menuIDs, rm.MenuID)
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(ctx))

	var menus []model.Menu
	if err := platformDB.Where("id IN ? AND status = 1", menuIDs).Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询菜单失败")
	}

	return menus, nil
}

func (s *MenuService) autoAssignDefaultMenus(ctx context.Context, roleIDs []int64, allMenus []model.Menu) []model.RoleMenu {
	superAdminOnlyPaths := map[string]bool{
		"/system/tenant":                true,
		"/system/menu":                  true,
		"/system/api":                   true,
		"/system/settings":              true,
		"/system/backups":               true,
		"/system/monitor":               true,
		"/system/tenant-packages":       true,
		"/system/tenant-channel":        true,
		"/system/error-codes":           true,
		"/system/sensitive-words":       true,
		"/system/online-users":          true,
		"/system/cron-jobs":             true,
		"/system/subscriptions":         true,
		"/system/notification-channels": true,
		"/system/plugin-monitor":        true,
		"/system/api-keys":              true,
		"/payment/settlement":           true,
		"/payment/config":               true,
		"/plugins/engine":               true,
	}

	tenantDB := utils.GetDB(ctx)
	if tenantDB == nil {
		return nil
	}

	var roleMenus []model.RoleMenu
	for _, roleID := range roleIDs {
		for _, menu := range allMenus {
			if superAdminOnlyPaths[menu.Path] {
				continue
			}
			rm := model.RoleMenu{
				RoleID: roleID,
				MenuID: menu.ID,
			}
			if err := tenantDB.Where("role_id = ? AND menu_id = ?", roleID, menu.ID).FirstOrCreate(&rm).Error; err != nil {
				log.Printf("自动分配菜单权限失败 [role=%d, menu=%d]: %v", roleID, menu.ID, err)
			}
			roleMenus = append(roleMenus, rm)
		}
	}

	return roleMenus
}
