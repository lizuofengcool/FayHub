package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"sync"
)

type ErrorCodeService struct{}

var ErrorCodeServiceApp = new(ErrorCodeService)

var (
	errorCodeCache     map[int]string
	errorCodeCacheOnce sync.Once
)

func (s *ErrorCodeService) List(ctx context.Context, name string, code *int, status *int, page, pageSize int) ([]*model.ErrorCode, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	query := db.Model(&model.ErrorCode{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if code != nil {
		query = query.Where("code = ?", *code)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询错误码总数失败")
	}

	var codes []*model.ErrorCode
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("code ASC").Find(&codes).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询错误码失败")
	}

	return codes, total, nil
}

func (s *ErrorCodeService) Create(ctx context.Context, ec *model.ErrorCode) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if err := db.Create(ec).Error; err != nil {
		return err
	}
	s.refreshCache(ctx)
	return nil
}

func (s *ErrorCodeService) Update(ctx context.Context, ec *model.ErrorCode) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if err := db.Save(ec).Error; err != nil {
		return err
	}
	s.refreshCache(ctx)
	return nil
}

func (s *ErrorCodeService) Delete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if err := db.Delete(&model.ErrorCode{}, id).Error; err != nil {
		return err
	}
	s.refreshCache(ctx)
	return nil
}

func (s *ErrorCodeService) GetMsg(ctx context.Context, code int) string {
	s.initCache(ctx)
	if msg, ok := errorCodeCache[code]; ok {
		return msg
	}
	return errs.GetErrorMessage(code)
}

func (s *ErrorCodeService) initCache(ctx context.Context) {
	errorCodeCacheOnce.Do(func() {
		s.refreshCache(ctx)
	})
}

func (s *ErrorCodeService) refreshCache(ctx context.Context) {
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	var codes []model.ErrorCode
	if err := db.Where("status = 1").Find(&codes).Error; err != nil {
		return
	}

	cache := make(map[int]string, len(codes))
	for _, c := range codes {
		cache[c.Code] = c.Msg
	}
	errorCodeCache = cache
	errorCodeCacheOnce = sync.Once{}
}

func (s *ErrorCodeService) RefreshCache(ctx context.Context) error {
	s.refreshCache(ctx)
	return nil
}
