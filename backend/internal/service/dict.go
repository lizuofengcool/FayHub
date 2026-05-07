package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type DictService struct{}

var DictServiceApp = new(DictService)

func (s *DictService) ListTypes(ctx context.Context, dictName, dictType string, status *int, page, pageSize int) ([]*model.DictType, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.DictType{})

	if dictName != "" {
		query = query.Where("dict_name LIKE ?", "%"+dictName+"%")
	}
	if dictType != "" {
		query = query.Where("dict_type LIKE ?", "%"+dictType+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询字典类型总数失败")
	}

	var types []*model.DictType
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&types).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询字典类型失败")
	}

	return types, total, nil
}

func (s *DictService) GetTypeByID(ctx context.Context, id int64) (*model.DictType, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var dt model.DictType
	if err := db.First(&dt, id).Error; err != nil {
		return nil, err
	}
	return &dt, nil
}

func (s *DictService) CreateType(ctx context.Context, dt *model.DictType) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Create(dt).Error
}

func (s *DictService) UpdateType(ctx context.Context, dt *model.DictType) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Save(dt).Error
}

func (s *DictService) DeleteType(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var dt model.DictType
	if err := db.First(&dt, id).Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dict_type = ?", dt.DictType).Delete(&model.DictData{}).Error; err != nil {
			return err
		}
		return tx.Delete(&dt).Error
	})
}

func (s *DictService) ListData(ctx context.Context, dictType, dictLabel string, status *int, page, pageSize int) ([]*model.DictData, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.DictData{})

	if dictType != "" {
		query = query.Where("dict_type = ?", dictType)
	}
	if dictLabel != "" {
		query = query.Where("dict_label LIKE ?", "%"+dictLabel+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询字典数据总数失败")
	}

	var data []*model.DictData
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&data).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询字典数据失败")
	}

	return data, total, nil
}

func (s *DictService) GetDataByType(ctx context.Context, dictType string) ([]*model.DictData, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var data []*model.DictData
	if err := db.Where("dict_type = ? AND status = 1", dictType).Order("sort ASC, id ASC").Find(&data).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询字典数据失败")
	}

	return data, nil
}

func (s *DictService) CreateData(ctx context.Context, dd *model.DictData) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Create(dd).Error
}

func (s *DictService) UpdateData(ctx context.Context, dd *model.DictData) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Save(dd).Error
}

func (s *DictService) DeleteData(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	return db.Delete(&model.DictData{}, id).Error
}
