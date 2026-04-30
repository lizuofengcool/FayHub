package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type DataPermissionService struct{}

type DataScopeFilter struct {
	Scope   int    `json:"scope"`
	DeptIDs []uint `json:"dept_ids,omitempty"`
	UserID  uint   `json:"user_id,omitempty"`
	IsAdmin bool   `json:"is_admin"`
}

func (s *DataPermissionService) GetDataScope(ctx context.Context, userID uint) (*DataScopeFilter, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var userRoles []model.UserRole
	if err := queryDB.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询用户角色失败")
	}

	if len(userRoles) == 0 {
		return &DataScopeFilter{
			Scope:   model.DataScopeSelf,
			UserID:  userID,
			IsAdmin: false,
		}, nil
	}

	roleIDs := make([]uint, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	var roles []model.Role
	if err := queryDB.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询角色失败")
	}

	maxScope := model.DataScopeSelf
	var maxScopeRole model.Role
	for _, role := range roles {
		if role.DataScope < maxScope || maxScope == model.DataScopeSelf {
			maxScope = role.DataScope
			maxScopeRole = role
		}
	}

	if maxScope == model.DataScopeAll {
		return &DataScopeFilter{
			Scope:   model.DataScopeAll,
			IsAdmin: true,
		}, nil
	}

	filter := &DataScopeFilter{
		Scope:   maxScope,
		UserID:  userID,
		IsAdmin: false,
	}

	switch maxScope {
	case model.DataScopeDept:
		deptIDs := s.getUserDeptIDs(queryDB, userID)
		filter.DeptIDs = deptIDs

	case model.DataScopeDeptAndSub:
		deptIDs := s.getUserDeptAndSubDeptIDs(queryDB, userID)
		filter.DeptIDs = deptIDs

	case model.DataScopeCustom:
		deptIDs := s.getRoleDeptIDs(queryDB, roleIDs)
		filter.DeptIDs = deptIDs

		if maxScopeRole.DeptID > 0 {
			filter.DeptIDs = append(filter.DeptIDs, maxScopeRole.DeptID)
		}
	}

	return filter, nil
}

func (s *DataPermissionService) ApplyDataScope(db *gorm.DB, filter *DataScopeFilter, deptFieldName string, userFieldName string) *gorm.DB {
	if filter == nil || filter.IsAdmin {
		return db
	}

	switch filter.Scope {
	case model.DataScopeAll:
		return db

	case model.DataScopeDept:
		if len(filter.DeptIDs) > 0 {
			return db.Where(deptFieldName+" IN ?", filter.DeptIDs)
		}
		return db.Where(userFieldName+" = ?", filter.UserID)

	case model.DataScopeDeptAndSub:
		if len(filter.DeptIDs) > 0 {
			return db.Where(deptFieldName+" IN ?", filter.DeptIDs)
		}
		return db.Where(userFieldName+" = ?", filter.UserID)

	case model.DataScopeSelf:
		return db.Where(userFieldName+" = ?", filter.UserID)

	case model.DataScopeCustom:
		if len(filter.DeptIDs) > 0 {
			return db.Where(deptFieldName+" IN ? OR "+userFieldName+" = ?", filter.DeptIDs, filter.UserID)
		}
		return db.Where(userFieldName+" = ?", filter.UserID)

	default:
		return db.Where(userFieldName+" = ?", filter.UserID)
	}
}

func (s *DataPermissionService) getUserDeptIDs(db *gorm.DB, userID uint) []uint {
	var userDepts []model.UserDepartment
	db.Where("user_id = ?", userID).Find(&userDepts)

	deptIDs := make([]uint, 0, len(userDepts))
	for _, ud := range userDepts {
		deptIDs = append(deptIDs, ud.DeptID)
	}
	return deptIDs
}

func (s *DataPermissionService) getUserDeptAndSubDeptIDs(db *gorm.DB, userID uint) []uint {
	directDeptIDs := s.getUserDeptIDs(db, userID)
	if len(directDeptIDs) == 0 {
		return nil
	}

	allDeptIDs := make([]uint, 0)
	allDeptIDs = append(allDeptIDs, directDeptIDs...)

	var subDepts []model.Department
	for _, deptID := range directDeptIDs {
		s.findSubDepts(db, deptID, &subDepts)
	}

	for _, sd := range subDepts {
		allDeptIDs = append(allDeptIDs, sd.ID)
	}

	return allDeptIDs
}

func (s *DataPermissionService) findSubDepts(db *gorm.DB, parentID uint, result *[]model.Department) {
	var children []model.Department
	db.Where("parent_id = ?", parentID).Find(&children)

	for _, child := range children {
		*result = append(*result, child)
		s.findSubDepts(db, child.ID, result)
	}
}

func (s *DataPermissionService) getRoleDeptIDs(db *gorm.DB, roleIDs []uint) []uint {
	var roleDepts []model.RoleDept
	db.Where("role_id IN ?", roleIDs).Find(&roleDepts)

	deptIDs := make([]uint, 0, len(roleDepts))
	seen := make(map[uint]bool)
	for _, rd := range roleDepts {
		if !seen[rd.DeptID] {
			deptIDs = append(deptIDs, rd.DeptID)
			seen[rd.DeptID] = true
		}
	}
	return deptIDs
}

type DepartmentService struct{}

type CreateDepartmentRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID uint   `json:"parent_id"`
	Sort     int    `json:"sort"`
	LeaderID uint   `json:"leader_id"`
}

type UpdateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"`
	Sort     *int   `json:"sort"`
	Status   *int   `json:"status"`
	LeaderID *uint  `json:"leader_id"`
}

func (ds *DepartmentService) Create(ctx context.Context, req CreateDepartmentRequest) (*model.Department, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	dept := &model.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
		Sort:     req.Sort,
		LeaderID: req.LeaderID,
		Status:   1,
	}

	if err := db.Create(dept).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建部门失败")
	}

	return dept, nil
}

func (ds *DepartmentService) Update(ctx context.Context, id uint, req UpdateDepartmentRequest) (*model.Department, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var dept model.Department
	if err := db.First(&dept, id).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "部门不存在")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.LeaderID != nil {
		updates["leader_id"] = *req.LeaderID
	}

	if len(updates) > 0 {
		if err := db.Model(&dept).Updates(updates).Error; err != nil {
			return nil, errs.NewServiceError(errs.ErrDatabase, "更新部门失败")
		}
	}

	db.First(&dept, id)
	return &dept, nil
}

func (ds *DepartmentService) Delete(ctx context.Context, id uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var count int64
	db.Model(&model.Department{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errs.NewServiceError(errs.ErrDatabase, "存在子部门，无法删除")
	}

	return db.Delete(&model.Department{}, id).Error
}

func (ds *DepartmentService) GetTree(ctx context.Context) ([]model.Department, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var depts []model.Department
	if err := db.Where("status = 1").Order("sort ASC, id ASC").Find(&depts).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询部门失败")
	}

	return ds.buildTree(depts, 0), nil
}

func (ds *DepartmentService) buildTree(depts []model.Department, parentID uint) []model.Department {
	var tree []model.Department
	for _, dept := range depts {
		if dept.ParentID == parentID {
			dept.Children = ds.buildTree(depts, dept.ID)
			tree = append(tree, dept)
		}
	}
	return tree
}

func (ds *DepartmentService) AssignUser(ctx context.Context, userID uint, deptID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var existing model.UserDepartment
	result := db.Where("user_id = ? AND dept_id = ?", userID, deptID).First(&existing)
	if result.Error == nil {
		return nil
	}

	ud := model.UserDepartment{UserID: userID, DeptID: deptID}
	return db.Create(&ud).Error
}

func (ds *DepartmentService) RemoveUser(ctx context.Context, userID uint, deptID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	return db.Where("user_id = ? AND dept_id = ?", userID, deptID).Delete(&model.UserDepartment{}).Error
}
