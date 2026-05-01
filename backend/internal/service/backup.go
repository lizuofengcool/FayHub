package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

	if err := s.dumpWithPgDump(ctx, filePath); err != nil {
		s.fallbackInsertDump(ctx, filePath)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		db.Model(record).Update("status", "failed")
		return nil, fmt.Errorf("备份文件生成失败")
	}

	db.Model(record).Updates(map[string]interface{}{
		"status":    "completed",
		"file_size": stat.Size(),
	})
	record.Status = "completed"
	record.FileSize = stat.Size()

	return record, nil
}

func (s *BackupService) dumpWithPgDump(ctx context.Context, filePath string) error {
	dsn := s.getPgDSN()
	if dsn == "" {
		return fmt.Errorf("非PostgreSQL数据库，无法使用pg_dump")
	}

	pgDumpPath, err := s.findPgDump()
	if err != nil {
		return fmt.Errorf("未找到pg_dump: %w", err)
	}

	args := []string{
		"--no-owner",
		"--no-privileges",
		"--no-comments",
		"--inserts",
		"--column-inserts",
		"-d", dsn,
		"-f", filePath,
	}

	cmd := exec.CommandContext(ctx, pgDumpPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump执行失败: %s: %w", string(output), err)
	}

	return nil
}

func (s *BackupService) getPgDSN() string {
	cfg := utils.GetDBConfig()
	if cfg == nil || cfg.Type != "postgresql" {
		return ""
	}

	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

func (s *BackupService) findPgDump() (string, error) {
	if path, err := exec.LookPath("pg_dump"); err == nil {
		return path, nil
	}

	searchPaths := []string{}
	if runtime.GOOS == "windows" {
		for _, base := range []string{
			`C:\Program Files\PostgreSQL`,
			`D:\Program Files\PostgreSQL`,
		} {
			entries, err := os.ReadDir(base)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					p := filepath.Join(base, e.Name(), "bin", "pg_dump.exe")
					if _, err := os.Stat(p); err == nil {
						searchPaths = append(searchPaths, p)
					}
				}
			}
		}
	} else {
		searchPaths = []string{
			"/usr/bin/pg_dump",
			"/usr/local/bin/pg_dump",
			"/usr/lib/postgresql/17/bin/pg_dump",
			"/usr/lib/postgresql/16/bin/pg_dump",
			"/usr/lib/postgresql/15/bin/pg_dump",
		}
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("pg_dump未安装")
}

func (s *BackupService) fallbackInsertDump(ctx context.Context, filePath string) {
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	f, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	var tables []string
	db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)

	for _, table := range tables {
		var count int64
		db.Table(table).Count(&count)
		if count == 0 {
			continue
		}
		fmt.Fprintf(f, "-- Table: %s (%d rows)\n", table, count)

		var rows []map[string]interface{}
		if err := db.Table(table).Find(&rows).Error; err != nil {
			continue
		}
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
			fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s);\n", table, strings.Join(cols, ", "), strings.Join(vals, ", "))
		}
		fmt.Fprintln(f)
	}
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

	safetyRecord, safetyErr := s.CreateBackup(ctx)
	if safetyErr == nil && safetyRecord != nil {
		oldFilename := safetyRecord.Filename
		newFilename := fmt.Sprintf("pre_restore_%s", oldFilename)
		oldPath := filepath.Join("data", "backups", oldFilename)
		newPath := filepath.Join("data", "backups", newFilename)
		if err := os.Rename(oldPath, newPath); err == nil {
			safetyRecord.Filename = newFilename
			db.Model(safetyRecord).Update("filename", newFilename)
		}
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
