package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type TenantPackageService struct{}

var TenantPackageServiceApp = new(TenantPackageService)

func (s *TenantPackageService) List(ctx context.Context, name string, status *int, page, pageSize int) ([]*model.TenantPackage, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.TenantPackage{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询套餐总数失败")
	}

	var packages []*model.TenantPackage
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&packages).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询套餐列表失败")
	}

	return packages, total, nil
}

func (s *TenantPackageService) GetAll(ctx context.Context) ([]*model.TenantPackage, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var packages []*model.TenantPackage
	if err := db.Where("status = 1").Order("sort ASC, id ASC").Find(&packages).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询套餐列表失败")
	}

	return packages, nil
}

func (s *TenantPackageService) GetByID(ctx context.Context, id int64) (*model.TenantPackage, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var pkg model.TenantPackage
	if err := db.First(&pkg, id).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (s *TenantPackageService) Create(ctx context.Context, pkg *model.TenantPackage, menuIDs []int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pkg).Error; err != nil {
			return err
		}

		for _, menuID := range menuIDs {
			pm := &model.TenantPackageMenu{
				PackageID: pkg.ID,
				MenuID:    menuID,
			}
			if err := tx.Create(pm).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *TenantPackageService) Update(ctx context.Context, pkg *model.TenantPackage, menuIDs []int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(pkg).Error; err != nil {
			return err
		}

		if err := tx.Where("package_id = ?", pkg.ID).Delete(&model.TenantPackageMenu{}).Error; err != nil {
			return err
		}

		for _, menuID := range menuIDs {
			pm := &model.TenantPackageMenu{
				PackageID: pkg.ID,
				MenuID:    menuID,
			}
			if err := tx.Create(pm).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *TenantPackageService) Delete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("package_id = ?", id).Delete(&model.TenantPackageMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.TenantPackage{}, id).Error
	})
}

func (s *TenantPackageService) GetMenuIDs(ctx context.Context, packageID int64) ([]int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var pms []model.TenantPackageMenu
	if err := db.Where("package_id = ?", packageID).Find(&pms).Error; err != nil {
		return nil, err
	}

	ids := make([]int64, len(pms))
	for i, pm := range pms {
		ids[i] = pm.MenuID
	}
	return ids, nil
}

func (s *TenantPackageService) GetTenantMenuIDs(ctx context.Context, tenantID int64) ([]int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var tenant model.Tenant
	if err := queryDB.First(&tenant, tenantID).Error; err != nil {
		return nil, nil
	}

	if tenant.PackageID == nil {
		return nil, nil
	}

	return s.GetMenuIDs(ctx, *tenant.PackageID)
}
