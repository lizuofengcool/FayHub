package service

import (
	"context"
	"fmt"
	"io"

	"fayhub/internal/model"
	"fayhub/pkg/config"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/eventbus"
	"fayhub/pkg/storage"
	"fayhub/pkg/utils"
)

type FileService struct{}

type UploadResult struct {
	ID           uint   `json:"id"`
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FileKey      string `json:"file_key"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	URL          string `json:"url"`
}

type ListFilesRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Keyword  string `json:"keyword"`
	MimeType string `json:"mime_type"`
}

type FileStats struct {
	TotalCount int64   `json:"total_count"`
	TotalSize  int64   `json:"total_size"`
	UsedMB     float64 `json:"used_mb"`
}

func (s *FileService) Upload(ctx context.Context, userID uint, originalName string, fileSize int64, mimeType string, reader io.Reader) (*UploadResult, error) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return nil, fmt.Errorf("系统配置未加载")
	}

	if !storage.IsAllowedType(originalName, cfg.Storage.AllowedTypes) {
		return nil, errs.NewServiceError(errs.ErrParamValidation, fmt.Sprintf("不支持的文件类型，允许: %s", cfg.Storage.AllowedTypes))
	}

	maxBytes := int64(cfg.Storage.MaxSizeMB) * 1024 * 1024
	if fileSize > maxBytes {
		return nil, errs.NewServiceError(errs.ErrParamValidation, fmt.Sprintf("文件大小超过限制（最大%dMB）", cfg.Storage.MaxSizeMB))
	}

	driver := storage.GetDriver()
	if driver == nil {
		return nil, fmt.Errorf("存储驱动未初始化")
	}

	fileKey := storage.GenerateFileKey(originalName)
	storedKey, err := driver.Upload(fileKey, reader)
	if err != nil {
		return nil, fmt.Errorf("文件上传失败: %w", err)
	}

	fileURL := driver.GetURL(storedKey)

	fileName := fmt.Sprintf("%d_%s", userID, originalName)
	if len(fileName) > 256 {
		fileName = fileName[:256]
	}

	record := model.FileRecord{
		FileName:      fileName,
		OriginalName:  originalName,
		FileKey:       storedKey,
		FileSize:      fileSize,
		MimeType:      mimeType,
		StorageDriver: cfg.Storage.Driver,
		URL:           fileURL,
		UserID:        userID,
	}

	db := utils.GetDB(ctx)
	if db == nil {
		driver.Delete(storedKey)
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if err := db.Create(&record).Error; err != nil {
		driver.Delete(storedKey)
		return nil, errs.NewServiceError(errs.ErrDatabase, "保存文件记录失败")
	}

	eventbus.PublishAsync(eventbus.EventFileUploaded, 0, map[string]interface{}{
		"file_id":   record.ID,
		"file_name": originalName,
		"file_size": fileSize,
		"user_id":   userID,
	})

	return &UploadResult{
		ID:           record.ID,
		FileName:     record.FileName,
		OriginalName: record.OriginalName,
		FileKey:      record.FileKey,
		FileSize:     record.FileSize,
		MimeType:     record.MimeType,
		URL:          record.URL,
	}, nil
}

func (s *FileService) Download(ctx context.Context, fileID uint) (io.ReadCloser, *model.FileRecord, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.FileRecord
	if err := db.Where("id = ? AND deleted_at IS NULL", fileID).First(&record).Error; err != nil {
		return nil, nil, errs.NewServiceError(errs.ErrDatabase, "文件不存在")
	}

	driver := storage.GetDriver()
	if driver == nil {
		return nil, nil, fmt.Errorf("存储驱动未初始化")
	}

	reader, err := driver.Download(record.FileKey)
	if err != nil {
		return nil, nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return reader, &record, nil
}

func (s *FileService) Delete(ctx context.Context, fileID uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.FileRecord
	if err := db.Where("id = ? AND deleted_at IS NULL", fileID).First(&record).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "文件不存在")
	}

	driver := storage.GetDriver()
	if driver != nil {
		driver.Delete(record.FileKey)
	}

	if err := db.Delete(&record).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除文件记录失败")
	}

	eventbus.PublishAsync(eventbus.EventFileDeleted, 0, map[string]interface{}{
		"file_id":   record.ID,
		"file_name": record.OriginalName,
	})

	return nil
}

func (s *FileService) ListFiles(ctx context.Context, req ListFilesRequest) ([]model.FileRecord, int64, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	query := db.Model(&model.FileRecord{}).Where("deleted_at IS NULL")

	if req.Keyword != "" {
		query = query.Where("original_name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.MimeType != "" {
		query = query.Where("mime_type LIKE ?", req.MimeType+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, errs.NewServiceError(errs.ErrDatabase, "查询文件数量失败")
	}

	var totalSize int64
	if err := query.Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize).Error; err != nil {
		return nil, 0, 0, errs.NewServiceError(errs.ErrDatabase, "查询文件大小失败")
	}

	var records []model.FileRecord
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id DESC").Offset(offset).Limit(req.PageSize).Find(&records).Error; err != nil {
		return nil, 0, 0, errs.NewServiceError(errs.ErrDatabase, "查询文件列表失败")
	}

	return records, total, totalSize, nil
}
