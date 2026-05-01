package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type BackupService struct{}

func (s *BackupService) CreateBackup(ctx context.Context) (*model.BackupRecord, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	backupDir := filepath.Join("data", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("fayhub_backup_%s.sql", timestamp)
	filePath := filepath.Join(backupDir, filename)

	record := &model.BackupRecord{
		Filename: filename,
		Status:   "pending",
	}

	if err := db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("创建备份记录失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		db.Model(record).Update("status", "failed")
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	f, err := os.Create(filePath)
	if err != nil {
		db.Model(record).Update("status", "failed")
		return nil, fmt.Errorf("创建备份文件失败: %w", err)
	}
	defer f.Close()

	if err := sqlDB.Ping(); err != nil {
		db.Model(record).Update("status", "failed")
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	var tables []string
	if err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables).Error; err != nil {
		var sqliteTables []model.Menu
		db.Table("sqlite_master").Where("type = ?", "table").Select("name").Find(&sqliteTables)
	}

	for _, table := range tables {
		var rows []map[string]interface{}
		if err := db.Table(table).Find(&rows).Error; err != nil {
			continue
		}
		if len(rows) > 0 {
			fmt.Fprintf(f, "-- Table: %s (%d rows)\n", table, len(rows))
			for _, row := range rows {
				cols := make([]string, 0, len(row))
				vals := make([]string, 0, len(row))
				for k, v := range row {
					cols = append(cols, k)
					if v == nil {
						vals = append(vals, "NULL")
					} else {
						vals = append(vals, fmt.Sprintf("'%v'", v))
					}
				}
				fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s);\n", table, joinStr(cols), joinStr(vals))
			}
			fmt.Fprintln(f)
		}
	}

	stat, _ := f.Stat()
	db.Model(record).Updates(map[string]interface{}{
		"status":    "completed",
		"file_size": stat.Size(),
	})
	record.Status = "completed"
	record.FileSize = stat.Size()

	return record, nil
}

func (s *BackupService) ListBackups(ctx context.Context) ([]model.BackupRecord, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var records []model.BackupRecord
	var total int64
	db.Model(&model.BackupRecord{}).Count(&total)
	db.Order("id DESC").Limit(50).Find(&records)

	return records, total, nil
}

func (s *BackupService) DeleteBackup(ctx context.Context, id uint) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.BackupRecord
	if err := db.First(&record, id).Error; err != nil {
		return fmt.Errorf("备份记录不存在")
	}

	filePath := filepath.Join("data", "backups", record.Filename)
	os.Remove(filePath)

	return db.Delete(&record).Error
}

func (s *BackupService) GetBackupFilePath(ctx context.Context, id uint) (string, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return "", errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.BackupRecord
	if err := db.First(&record, id).Error; err != nil {
		return "", fmt.Errorf("备份记录不存在")
	}

	return filepath.Join("data", "backups", record.Filename), nil
}

func (s *BackupService) RestoreBackup(ctx context.Context, filePath string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %w", err)
	}

	sqlContent := string(content)
	if sqlContent == "" {
		return fmt.Errorf("备份文件内容为空")
	}

	if err := db.Exec(sqlContent).Error; err != nil {
		return fmt.Errorf("恢复数据库失败: %w", err)
	}

	return nil
}

func joinStr(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}
