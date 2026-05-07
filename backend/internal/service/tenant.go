package service

import (
	"context"
	"errors"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TenantService struct{}

type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
}

type UpdateTenantRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
}

type TenantListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   *int   `json:"status" form:"status"`
}

type TenantListResponse struct {
	List  []model.Tenant `json:"list"`
	Total int64          `json:"total"`
}

type ImpersonateResponse struct {
	Token    string `json:"token"`
	TenantID int64  `json:"tenant_id"`
	Username string `json:"username"`
}

func (s *TenantService) Create(ctx context.Context, req CreateTenantRequest) (*model.Tenant, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	tenant := &model.Tenant{
		Name:        req.Name,
		Domain:      req.Domain,
		Description: req.Description,
		Status:      1,
	}

	ctx = utils.SkipTenantIsolation(ctx)
	db = utils.GetDB(ctx)
	if err := db.Create(tenant).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建租户失败")
	}

	return tenant, nil
}

func (s *TenantService) Update(ctx context.Context, id int64, req UpdateTenantRequest) (*model.Tenant, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	ctx = utils.SkipTenantIsolation(ctx)
	db = utils.GetDB(ctx)

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrTenantNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询租户失败")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Domain != "" {
		updates["domain"] = req.Domain
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := db.Model(&tenant).Updates(updates).Error; err != nil {
			return nil, errs.NewServiceError(errs.ErrDatabase, "更新租户失败")
		}
	}

	db.First(&tenant, id)
	return &tenant, nil
}

func (s *TenantService) SoftDelete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	ctx = utils.SkipTenantIsolation(ctx)
	db = utils.GetDB(ctx)

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrTenantNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询租户失败")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tenant).Update("status", 2).Error; err != nil {
			return err
		}

		cascadeModels := s.getCascadeModels()
		for _, m := range cascadeModels {
			if err := tx.Where("tenant_id = ?", id).Delete(m).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *TenantService) Restore(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	ctx = utils.SkipTenantIsolation(ctx)
	db = utils.GetDB(ctx)

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrTenantNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询租户失败")
	}

	if tenant.Status != 2 {
		return errs.NewServiceError(errs.ErrParamValidation, "该租户不在回收站中")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tenant).Update("status", 1).Error; err != nil {
			return err
		}

		cascadeModels := s.getCascadeModels()
		for _, m := range cascadeModels {
			if err := tx.Unscoped().Model(m).Where("tenant_id = ? AND deleted_at IS NOT NULL", id).Update("deleted_at", nil).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *TenantService) PermanentDelete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	ctx = utils.SkipTenantIsolation(ctx)
	db = utils.GetDB(ctx)

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewServiceError(errs.ErrTenantNotExist, "")
		}
		return errs.NewServiceError(errs.ErrDatabase, "查询租户失败")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		cascadeModels := s.getCascadeModels()
		for _, m := range cascadeModels {
			if err := tx.Unscoped().Where("tenant_id = ?", id).Delete(m).Error; err != nil {
				return err
			}
		}

		if err := tx.Unscoped().Delete(&tenant).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *TenantService) Impersonate(ctx context.Context, tenantID int64, adminID int64) (*ImpersonateResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var tenant model.Tenant
	if err := queryDB.First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrTenantNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询租户失败")
	}

	if tenant.Status != 1 {
		return nil, errs.NewServiceError(errs.ErrParamValidation, "租户已禁用或在回收站中，无法模拟登录")
	}

	var targetUser model.User
	if tenant.OwnerID > 0 {
		if err := queryDB.First(&targetUser, tenant.OwnerID).Error; err != nil {
			return nil, errs.NewServiceError(errs.ErrUserNotExist, "租户主账号不存在")
		}
	} else {
		if err := queryDB.Where("tenant_id = ? AND role = ? AND status = ?", tenantID, "tenant_admin", 1).First(&targetUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				targetUser, err = s.createTenantAdmin(queryDB, tenant)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errs.NewServiceError(errs.ErrDatabase, "查询租户管理员失败")
			}
		}
	}

	token, err := utils.GenerateImpersonateToken(targetUser.ID, targetUser.Username, targetUser.Role, tenantID, adminID)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrInternalServer, "生成模拟登录Token失败")
	}

	return &ImpersonateResponse{
		Token:    token,
		TenantID: tenantID,
		Username: targetUser.Username,
	}, nil
}

func (s *TenantService) createTenantAdmin(db *gorm.DB, tenant model.Tenant) (model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, errs.NewServiceError(errs.ErrInternalServer, "生成密码失败")
	}

	adminUser := model.User{
		SnowflakeTenantModel: model.SnowflakeTenantModel{TenantID: tenant.ID},
		Username:             "admin",
		Password:             string(hashedPassword),
		Email:                "admin@" + tenant.Domain,
		Status:               1,
		Role:                 "tenant_admin",
		RealName:             tenant.Name + "管理员",
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return model.User{}, errs.NewServiceError(errs.ErrDatabase, "创建租户管理员失败")
	}

	platformDB := utils.GetDB(utils.SkipTenantIsolation(context.Background()))
	if platformDB != nil {
		var tenantAdminRole model.Role
		if err := platformDB.Where("name = ?", "tenant_admin").First(&tenantAdminRole).Error; err == nil {
			userRole := model.UserRole{
				SnowflakeTenantModel: model.SnowflakeTenantModel{TenantID: tenant.ID},
				UserID:               adminUser.ID,
				RoleID:               tenantAdminRole.ID,
			}
			if err := db.Create(&userRole).Error; err != nil {
				log.Printf("分配租户管理员角色失败: %v", err)
			}

			superAdminOnlyPaths := map[string]bool{
				"/system/tenant":          true,
				"/system/menu":            true,
				"/system/api":             true,
				"/system/backups":         true,
				"/system/monitor":         true,
				"/system/tenant-packages": true,
				"/system/tenant-channel":  true,
			}

			var allMenus []model.Menu
			if err := platformDB.Where("status = 1").Find(&allMenus).Error; err == nil {
				for _, menu := range allMenus {
					if superAdminOnlyPaths[menu.Path] {
						continue
					}
					roleMenu := model.RoleMenu{
						SnowflakeTenantModel: model.SnowflakeTenantModel{TenantID: tenant.ID},
						RoleID:               tenantAdminRole.ID,
						MenuID:               menu.ID,
					}
					if err := db.Create(&roleMenu).Error; err != nil {
						log.Printf("分配菜单权限失败 [%s]: %v", menu.Path, err)
					}
				}
			}
		}
	}

	return adminUser, nil
}

func (s *TenantService) getCascadeModels() []interface{} {
	return []interface{}{
		&model.User{},
		&model.TenantUser{},
		&model.Role{},
		&model.RoleMenu{},
		&model.RoleAPI{},
		&model.UserRole{},
		&model.TenantRole{},
		&model.TenantQuota{},
		&model.InstalledPlugin{},
		&model.PluginConfig{},
		&model.PluginEventLog{},
		&model.PaymentConfig{},
		&model.PaymentOrder{},
		&model.WebhookSubscription{},
		&model.WebhookDelivery{},
		&model.FileRecord{},
		&model.APIKey{},
		&model.SettlementRecord{},
		&model.SettlementConfig{},
		&model.Department{},
		&model.UserDepartment{},
		&model.RoleDept{},
		&model.LoginLog{},
		&model.AuditLog{},
		&model.Notification{},
		&model.SSOAuthorizationCode{},
		&model.SSOTokenData{},
		&model.PluginVersionHistory{},
		&model.PluginDependency{},
	}
}

func (s *TenantService) GetByID(ctx context.Context, id int64) (*model.Tenant, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	ctx = utils.SkipTenantIsolation(ctx)
	db = utils.GetDB(ctx)

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrTenantNotExist, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询租户失败")
	}

	return &tenant, nil
}

func (s *TenantService) GetList(ctx context.Context, req TenantListRequest) (*TenantListResponse, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	ctx = utils.SkipTenantIsolation(ctx)
	db = utils.GetDB(ctx)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	query := db.Model(&model.Tenant{})

	if req.Keyword != "" {
		keyword := utils.EscapeLike(req.Keyword)
		query = query.Where("name LIKE ? OR domain LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	} else {
		query = query.Where("status != ?", 2)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询租户总数失败")
	}

	var tenants []model.Tenant
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id DESC").Find(&tenants).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询租户列表失败")
	}

	return &TenantListResponse{
		List:  tenants,
		Total: total,
	}, nil
}
