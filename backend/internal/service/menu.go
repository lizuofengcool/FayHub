package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"fmt"

	"gorm.io/gorm"
)

type MenuService struct{}

type CreateMenuRequest struct {
	ParentID   uint   `json:"parent_id"`
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
	ParentID   uint   `json:"parent_id"`
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
	RoleID  uint   `json:"role_id" binding:"required"`
	MenuIDs []uint `json:"menu_ids" binding:"required,min=1"`
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

	menuMap := make(map[uint][]model.Menu)
	for i := range menus {
		if menus[i].ParentID != 0 {
			menuMap[menus[i].ParentID] = append(menuMap[menus[i].ParentID], menus[i])
		}
	}

	var tree []model.Menu
	for i := range menus {
		if menus[i].ParentID == 0 {
			if children, ok := menuMap[menus[i].ID]; ok {
				menus[i].Children = children
			} else {
				menus[i].Children = []model.Menu{}
			}
			tree = append(tree, menus[i])
		}
	}

	return tree, nil
}

func (s *MenuService) GetMenuByID(ctx context.Context, menuID uint) (*model.Menu, error) {
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

func (s *MenuService) UpdateMenu(ctx context.Context, menuID uint, req UpdateMenuRequest) (*model.Menu, error) {
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

func (s *MenuService) DeleteMenu(ctx context.Context, menuID uint) error {
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

	for _, menuID := range req.MenuIDs {
		var menu model.Menu
		if err := platformDB.First(&menu, menuID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errs.NewServiceError(errs.ErrMenuNotExist, fmt.Sprintf("菜单不存在: %d", menuID))
			}
			return errs.NewServiceError(errs.ErrDatabase, "查询菜单失败")
		}
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

func (s *MenuService) GetRoleMenus(ctx context.Context, roleID uint) ([]model.Menu, error) {
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

	var menuIDs []uint
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
