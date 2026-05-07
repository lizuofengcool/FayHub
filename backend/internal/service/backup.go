package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fayhub/internal/model"
	"fayhub/pkg/config"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BackupService struct{}

func (s *BackupService) CreateBackup(ctx context.Context) (*model.BackupRecord, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	cfg := config.GlobalConfig
	backupDir := cfg.Backup.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join("data", "backups")
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, errs.NewServiceError(errs.ErrFileSystem, "创建备份目录失败")
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("fayhub_backup_%s.sql", timestamp)
	filePath := filepath.Join(backupDir, filename)

	record := &model.BackupRecord{
		Filename: filename,
		Status:   "pending",
	}

	if err := db.Create(record).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建备份记录失败")
	}

	if err := s.dumpWithPgDump(ctx, filePath); err != nil {
		s.fallbackInsertDump(ctx, filePath)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		db.Model(record).Update("status", "failed")
		return nil, errs.NewServiceError(errs.ErrFileSystem, "备份文件生成失败")
	}

	db.Model(record).Updates(map[string]interface{}{
		"status":    "completed",
		"file_size": stat.Size(),
	})
	record.Status = "completed"
	record.FileSize = stat.Size()

	go s.cleanupOldBackups(ctx)

	return record, nil
}

func (s *BackupService) dumpWithPgDump(ctx context.Context, filePath string) error {
	dsn := s.getPgDSN()
	if dsn == "" {
		return errs.NewServiceError(errs.ErrDatabase, "非PostgreSQL数据库，无法使用pg_dump")
	}

	pgDumpPath, err := s.findPgDump()
	if err != nil {
		return errs.NewServiceError(errs.ErrFileSystem, "未找到pg_dump")
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
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errs.NewServiceError(errs.ErrFileSystem, "pg_dump执行失败")
	}

	return nil
}

type TableInfo struct {
	Name       string `json:"name"`
	Comment    string `json:"comment"`
	RowCount   int64  `json:"row_count"`
	TotalSize  string `json:"total_size"`
	IndexSize  string `json:"index_size"`
	UpdateTime string `json:"update_time"`
}

func (s *BackupService) ListTables(ctx context.Context) ([]TableInfo, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var tables []TableInfo
	cfg := utils.GetDBConfig()
	if cfg != nil && cfg.Type == "postgresql" {
		rows, err := db.Raw(`
			SELECT
				t.tablename AS name,
				COALESCE(obj_description(c.oid), '') AS comment,
				COALESCE(s.n_live_tup, 0) AS row_count,
				pg_size_pretty(pg_total_relation_size(c.oid)) AS total_size,
				pg_size_pretty(pg_indexes_size(c.oid)) AS index_size,
				COALESCE(to_char(s.last_vacuum, 'YYYY-MM-DD HH24:MI:SS'), '') AS update_time
			FROM pg_tables t
			LEFT JOIN pg_class c ON c.relname = t.tablename
			LEFT JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.schemaname
			LEFT JOIN pg_stat_user_tables s ON s.relname = t.tablename AND s.schemaname = t.schemaname
			WHERE t.schemaname = 'public'
			ORDER BY t.tablename
		`).Rows()
		if err != nil {
			return nil, errs.NewServiceError(errs.ErrDatabase, "查询表信息失败")
		}
		defer rows.Close()
		for rows.Next() {
			var ti TableInfo
			if err := rows.Scan(&ti.Name, &ti.Comment, &ti.RowCount, &ti.TotalSize, &ti.IndexSize, &ti.UpdateTime); err == nil {
				tables = append(tables, ti)
			}
		}
	} else {
		var tableNames []string
		db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tableNames)
		for _, name := range tableNames {
			var count int64
			db.Table(name).Count(&count)
			tables = append(tables, TableInfo{Name: name, RowCount: count})
		}
	}

	return tables, nil
}

func (s *BackupService) CreateBackupForTables(ctx context.Context, tableNames []string) (*model.BackupRecord, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	backupDir := filepath.Join("data", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, errs.NewServiceError(errs.ErrFileSystem, "创建备份目录失败")
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("fayhub_backup_%s.sql", timestamp)
	filePath := filepath.Join(backupDir, filename)

	record := &model.BackupRecord{
		Filename: filename,
		Status:   "pending",
	}
	if err := db.Create(record).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建备份记录失败")
	}

	f, err := os.Create(filePath)
	if err != nil {
		db.Model(record).Update("status", "failed")
		return nil, errs.NewServiceError(errs.ErrFileSystem, "创建备份文件失败")
	}
	defer f.Close()

	fmt.Fprintf(f, "-- FayHub Backup %s\n", timestamp)
	fmt.Fprintf(f, "-- Tables: %s\n\n", strings.Join(tableNames, ", "))

	const batchSize = 500
	for _, table := range tableNames {
		if !utils.ValidateTableName(table) {
			continue
		}
		var count int64
		db.Table(table).Count(&count)
		if count == 0 {
			fmt.Fprintf(f, "-- Table: %s (0 rows, skipped)\n\n", table)
			continue
		}
		fmt.Fprintf(f, "-- Table: %s (%d rows)\n", table, count)
		for offset := 0; offset < int(count); offset += batchSize {
			var rows []map[string]interface{}
			if err := db.Table(table).Offset(offset).Limit(batchSize).Find(&rows).Error; err != nil {
				continue
			}
			for _, row := range rows {
				cols := make([]string, 0, len(row))
				vals := make([]string, 0, len(row))
				for k, v := range row {
					if !utils.ValidateTableName(k) {
						continue
					}
					cols = append(cols, k)
					if v == nil {
						vals = append(vals, "NULL")
					} else {
						escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
						vals = append(vals, fmt.Sprintf("'%s'", escaped))
					}
				}
				if len(cols) > 0 {
					fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s);\n", table, strings.Join(cols, ", "), strings.Join(vals, ", "))
				}
			}
		}
		fmt.Fprintln(f)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		db.Model(record).Update("status", "failed")
		return nil, errs.NewServiceError(errs.ErrFileSystem, "备份文件生成失败")
	}

	db.Model(record).Updates(map[string]interface{}{
		"status":    "completed",
		"file_size": stat.Size(),
	})
	record.Status = "completed"
	record.FileSize = stat.Size()

	return record, nil
}

var sqlSelectOnlyPatterns = []string{
	"DROP ", "INSERT ", "UPDATE ", "DELETE ",
	"ALTER ", "CREATE ", "TRUNCATE ", "GRANT ",
	"REVOKE ", "EXECUTE ", "PREPARE ", "DO $$",
	"DO $func$", "COPY ", "VACUUM ", "REINDEX ",
}

func (s *BackupService) ExecuteSQL(ctx context.Context, sql string, showErrors bool) (interface{}, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	upperSQL := strings.ToUpper(strings.TrimSpace(sql))
	for _, pattern := range sqlSelectOnlyPatterns {
		if strings.HasPrefix(upperSQL, pattern) {
			return nil, errs.NewServiceError(errs.ErrForbidden, fmt.Sprintf("仅允许SELECT查询，不允许执行: %s", strings.TrimSpace(pattern)))
		}
	}

	if !strings.HasPrefix(upperSQL, "SELECT") {
		return nil, errs.NewServiceError(errs.ErrForbidden, "仅允许SELECT查询语句")
	}

	rows, err := db.Raw(sql).Rows()
	if err != nil {
		if showErrors {
			return nil, errs.NewServiceError(errs.ErrDatabase, fmt.Sprintf("SQL执行错误: %v", err))
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "SQL执行失败")
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		rowMap := make(map[string]interface{})
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		result = append(result, rowMap)
	}

	return gin.H{"columns": cols, "rows": result, "count": len(result)}, nil
}

type ProcessInfo struct {
	PID     int    `json:"pid"`
	User    string `json:"user"`
	Host    string `json:"host"`
	DB      string `json:"database"`
	Command string `json:"command"`
	Time    string `json:"time"`
	State   string `json:"state"`
	Query   string `json:"query"`
}

func (s *BackupService) ListProcesses(ctx context.Context) ([]ProcessInfo, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	cfg := utils.GetDBConfig()
	if cfg == nil || cfg.Type != "postgresql" {
		return nil, errs.NewServiceError(errs.ErrDatabase, "仅支持PostgreSQL数据库")
	}

	rows, err := db.Raw(`
		SELECT pid, usename, client_addr::text, datname, query_start::text, state, query
		FROM pg_stat_activity
		WHERE datname = current_database()
		ORDER BY query_start DESC NULLS LAST
	`).Rows()
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询进程失败")
	}
	defer rows.Close()

	var processes []ProcessInfo
	for rows.Next() {
		var pi ProcessInfo
		var clientAddr, queryStart sql.NullString
		if err := rows.Scan(&pi.PID, &pi.User, &clientAddr, &pi.DB, &queryStart, &pi.State, &pi.Query); err == nil {
			pi.Host = clientAddr.String
			pi.Time = queryStart.String
			pi.Command = "Query"
			processes = append(processes, pi)
		}
	}

	return processes, nil
}

func (s *BackupService) KillProcess(ctx context.Context, pid int) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if err := db.Exec("SELECT pg_terminate_backend(?)", pid).Error; err != nil {
		return errs.NewServiceError(errs.ErrOperationFailed, "终止进程失败")
	}
	return nil
}

type FieldVerifyResult struct {
	TableName  string   `json:"table_name"`
	FieldCount int      `json:"field_count"`
	RowCount   int64    `json:"row_count"`
	Status     string   `json:"status"`
	Issues     []string `json:"issues"`
}

func (s *BackupService) VerifyFields(ctx context.Context) ([]FieldVerifyResult, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	tables, err := s.ListTables(ctx)
	if err != nil {
		return nil, err
	}

	var results []FieldVerifyResult
	for _, t := range tables {
		if !utils.ValidateTableName(t.Name) {
			continue
		}
		var issues []string

		type ColInfo struct {
			Name     string
			Nullable string
			TypeName string
		}
		var cols []ColInfo
		db.Raw(`
			SELECT column_name, is_nullable, data_type
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = ?
			ORDER BY ordinal_position
		`, t.Name).Scan(&cols)

		for _, col := range cols {
			if col.Nullable == "YES" {
				var nullCount int64
				db.Table(t.Name).Where(fmt.Sprintf("%s IS NULL", col.Name)).Count(&nullCount)
				if nullCount > 0 {
					issues = append(issues, fmt.Sprintf("字段 %s 有 %d 条NULL值", col.Name, nullCount))
				}
			}
		}

		status := "pass"
		if len(issues) > 0 {
			status = "error"
		}

		results = append(results, FieldVerifyResult{
			TableName:  t.Name,
			FieldCount: len(cols),
			RowCount:   t.RowCount,
			Status:     status,
			Issues:     issues,
		})
	}

	return results, nil
}

func (s *BackupService) ReplaceData(ctx context.Context, tableName, findStr, replaceStr string) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if !utils.ValidateTableName(tableName) {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}

	type ColInfo struct {
		Name     string
		TypeName string
	}
	var cols []ColInfo
	db.Raw(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = ?
		AND data_type IN ('character varying', 'text', 'character', 'varchar', 'char')
		ORDER BY ordinal_position
	`, tableName).Scan(&cols)

	if len(cols) == 0 {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "该表没有可替换的文本字段")
	}

	var totalAffected int64
	for _, col := range cols {
		if !utils.ValidateTableName(col.Name) {
			continue
		}
		result := db.Exec(
			fmt.Sprintf("UPDATE %s SET %s = REPLACE(%s, ?, ?) WHERE %s LIKE ?", tableName, col.Name, col.Name, col.Name),
			replaceStr, findStr, "%"+findStr+"%",
		)
		if result.Error == nil {
			totalAffected += result.RowsAffected
		}
	}

	return totalAffected, nil
}

func (s *BackupService) ExportTable(ctx context.Context, tableName string, format string) (string, []byte, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return "", nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if !utils.ValidateTableName(tableName) {
		return "", nil, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}

	var rows []map[string]interface{}
	if err := db.Table(tableName).Find(&rows).Error; err != nil {
		return "", nil, errs.NewServiceError(errs.ErrDatabase, "查询数据失败")
	}

	if len(rows) == 0 {
		return "", nil, errs.NewServiceError(errs.ErrResourceNotFound, "表中没有数据")
	}

	filename := fmt.Sprintf("%s_%s.%s", tableName, time.Now().Format("20060102_150405"), format)

	if format == "csv" {
		var buf strings.Builder
		cols := make([]string, 0)
		for k := range rows[0] {
			cols = append(cols, k)
		}
		sort.Strings(cols)
		buf.WriteString(strings.Join(cols, ",") + "\n")
		for _, row := range rows {
			vals := make([]string, 0, len(cols))
			for _, col := range cols {
				v := row[col]
				if v == nil {
					vals = append(vals, "")
				} else {
					s := fmt.Sprintf("%v", v)
					s = strings.ReplaceAll(s, "\"", "\"\"")
					vals = append(vals, "\""+s+"\"")
				}
			}
			buf.WriteString(strings.Join(vals, ",") + "\n")
		}
		return filename, []byte(buf.String()), nil
	}

	return "", nil, errs.NewServiceError(errs.ErrParamValidation, fmt.Sprintf("不支持的导出格式: %s", format))
}

func (s *BackupService) ImportData(ctx context.Context, tableName string, records []map[string]interface{}) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if !utils.ValidateTableName(tableName) {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}

	var totalAffected int64
	const batchSize = 100
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]
		result := db.Table(tableName).Create(batch)
		if result.Error != nil {
			continue
		}
		totalAffected += result.RowsAffected
	}

	return totalAffected, nil
}
func stripSQLCommentsForRestore(sql string) string {
	result := ""
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		if idx := strings.Index(line, "--"); idx > 0 {
			inQuote := false
			quoteChar := byte(0)
			for i := 0; i < idx; i++ {
				if line[i] == '\'' || line[i] == '"' {
					if inQuote && line[i] == quoteChar {
						inQuote = false
					} else if !inQuote {
						inQuote = true
						quoteChar = line[i]
					}
				}
			}
			if !inQuote {
				line = line[:idx]
			}
		}
		result += line + "\n"
	}
	blockCommentRe := regexp.MustCompile(`(?s)/\*.*?\*/`)
	result = blockCommentRe.ReplaceAllString(result, " ")
	return result
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

	return "", errs.NewServiceError(errs.ErrFileSystem, "pg_dump未安装")
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

	const batchSize = 500
	for _, table := range tables {
		if !utils.ValidateTableName(table) {
			continue
		}
		var count int64
		db.Table(table).Count(&count)
		if count == 0 {
			continue
		}
		fmt.Fprintf(f, "-- Table: %s (%d rows)\n", table, count)

		for offset := 0; offset < int(count); offset += batchSize {
			var rows []map[string]interface{}
			if err := db.Table(table).Offset(offset).Limit(batchSize).Find(&rows).Error; err != nil {
				continue
			}
			for _, row := range rows {
				cols := make([]string, 0, len(row))
				vals := make([]string, 0, len(row))
				for k, v := range row {
					if !utils.ValidateTableName(k) {
						continue
					}
					cols = append(cols, k)
					if v == nil {
						vals = append(vals, "NULL")
					} else {
						escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
						vals = append(vals, fmt.Sprintf("'%s'", escaped))
					}
				}
				if len(cols) > 0 {
					fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s);\n", table, strings.Join(cols, ", "), strings.Join(vals, ", "))
				}
			}
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

func (s *BackupService) DeleteBackup(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.BackupRecord
	if err := db.First(&record, id).Error; err != nil {
		return errs.NewServiceError(errs.ErrResourceNotFound, "备份记录不存在")
	}

	filePath := filepath.Join("data", "backups", record.Filename)
	os.Remove(filePath)

	return db.Delete(&record).Error
}

func (s *BackupService) GetBackupFilePath(ctx context.Context, id int64) (string, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return "", errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.BackupRecord
	if err := db.First(&record, id).Error; err != nil {
		return "", errs.NewServiceError(errs.ErrResourceNotFound, "备份记录不存在")
	}

	return filepath.Join("data", "backups", record.Filename), nil
}

var sqlDangerousPatterns = []string{
	"DROP DATABASE",
	"DROP SCHEMA",
	"ALTER SYSTEM",
	"COPY ",
	"COPY(",
	"pg_read_file",
	"pg_execute_server_program",
	"lo_import",
	"lo_export",
	"dblink",
	"pg_sleep",
	"CREATE EXTENSION",
	"CREATE FUNCTION",
	"CREATE LANGUAGE",
	"CREATE RULE",
	"CREATE TRIGGER",
	"CREATE PROCEDURE",
	"ALTER ROLE",
	"ALTER USER",
	"GRANT ALL",
	"SET SESSION",
	"RESET ALL",
	"LOAD ",
	"EXECUTE ",
	"PREPARE ",
	"DO $$",
	"DO $func$",
}

var sqlAllowedPrefixes = []string{
	"INSERT ",
	"INSERT\n",
	"INSERT\t",
}

func (s *BackupService) RestoreBackupByID(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.BackupRecord
	if err := db.First(&record, id).Error; err != nil {
		return errs.NewServiceError(errs.ErrResourceNotFound, "备份记录不存在")
	}

	if record.Status != "completed" {
		return errs.NewServiceError(errs.ErrOperationFailed, "只能恢复已完成的备份")
	}

	filePath := filepath.Join("data", "backups", record.Filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errs.NewServiceError(errs.ErrFileSystem, "备份文件已丢失")
	}

	preRestoreFilename := fmt.Sprintf("pre_restore_%s", record.Filename)
	preRestorePath := filepath.Join("data", "backups", preRestoreFilename)
	if err := s.CreatePreRestoreBackup(ctx, preRestorePath); err != nil {
		return errs.NewServiceError(errs.ErrFileSystem, "创建恢复前备份失败")
	}

	preRecord := &model.BackupRecord{
		Filename: preRestoreFilename,
		Status:   "completed",
	}
	if stat, err := os.Stat(preRestorePath); err == nil {
		preRecord.FileSize = stat.Size()
	}
	db.Create(preRecord)

	return s.RestoreBackup(ctx, filePath)
}

func (s *BackupService) CreatePreRestoreBackup(ctx context.Context, filePath string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	backupDir := filepath.Dir(filePath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return errs.NewServiceError(errs.ErrFileSystem, "创建备份目录失败")
	}

	if err := s.dumpWithPgDump(ctx, filePath); err != nil {
		s.fallbackInsertDump(ctx, filePath)
	}

	if _, err := os.Stat(filePath); err != nil {
		return errs.NewServiceError(errs.ErrFileSystem, "预恢复备份文件生成失败")
	}

	return nil
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
		return errs.NewServiceError(errs.ErrFileSystem, "读取备份文件失败")
	}

	sqlContent := string(content)
	if sqlContent == "" {
		return errs.NewServiceError(errs.ErrParamValidation, "备份文件内容为空")
	}

	cleanedContent := stripSQLCommentsForRestore(sqlContent)
	upperContent := strings.ToUpper(cleanedContent)
	for _, pattern := range sqlDangerousPatterns {
		if strings.Contains(upperContent, pattern) {
			return errs.NewServiceError(errs.ErrParamValidation, fmt.Sprintf("备份文件包含不允许的语句: %s", pattern))
		}
	}

	statements := strings.Split(sqlContent, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		trimmedUpper := strings.ToUpper(strings.TrimSpace(stripSQLCommentsForRestore(stmt)))
		if trimmedUpper == "" {
			continue
		}
		allowed := false
		for _, prefix := range sqlAllowedPrefixes {
			if strings.HasPrefix(trimmedUpper, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errs.NewServiceError(errs.ErrParamValidation, fmt.Sprintf("备份文件包含不允许的语句类型，仅允许INSERT: %s", truncateStr(trimmedUpper, 50)))
		}
		if err := db.Exec(stmt).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "恢复数据库失败")
		}
	}

	return nil
}

type FieldInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable string `json:"nullable"`
	Default  string `json:"default"`
	Comment  string `json:"comment"`
}

func (s *BackupService) GetTableFields(ctx context.Context, tableName string) ([]FieldInfo, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if !utils.ValidateTableName(tableName) {
		return nil, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}
	var fields []FieldInfo
	rows, err := db.Raw(`
		SELECT c.column_name, c.data_type, c.is_nullable,
			COALESCE(c.column_default, ''),
			COALESCE(col_description(t.oid, a.attnum), '')
		FROM information_schema.columns c
		LEFT JOIN pg_class t ON t.relname = c.table_name
		LEFT JOIN pg_namespace n ON n.nspname = c.table_schema AND n.oid = t.relnamespace
		LEFT JOIN pg_attribute a ON a.attrelid = t.oid AND a.attname = c.column_name
		WHERE c.table_schema = 'public' AND c.table_name = ?
		ORDER BY c.ordinal_position
	`, tableName).Rows()
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询字段信息失败")
	}
	defer rows.Close()
	for rows.Next() {
		var f FieldInfo
		if err := rows.Scan(&f.Name, &f.Type, &f.Nullable, &f.Default, &f.Comment); err == nil {
			fields = append(fields, f)
		}
	}
	return fields, nil
}

func (s *BackupService) AdvancedReplace(ctx context.Context, tableName, fieldName, findStr, replaceStr string, replaceType int, condition string, batchSize int) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if !utils.ValidateTableName(tableName) {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}
	if batchSize <= 0 {
		batchSize = 1000
	}
	if batchSize > 5000 {
		batchSize = 5000
	}

	var textFields []string
	if fieldName != "" {
		if !utils.ValidateTableName(fieldName) {
			return 0, errs.NewServiceError(errs.ErrParamValidation, "无效的字段名")
		}
		textFields = []string{fieldName}
	} else {
		type ColInfo struct {
			Name     string
			TypeName string
		}
		var cols []ColInfo
		db.Raw(`
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = ?
			AND data_type IN ('character varying', 'text', 'character', 'varchar', 'char')
			ORDER BY ordinal_position
		`, tableName).Scan(&cols)
		for _, c := range cols {
			textFields = append(textFields, c.Name)
		}
	}

	if len(textFields) == 0 {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "没有可替换的文本字段")
	}

	var totalAffected int64
	for _, col := range textFields {
		var result *gorm.DB
		switch replaceType {
		case 2:
			result = db.Exec(
				fmt.Sprintf("UPDATE %s SET %s = ? || %s WHERE %s LIKE ?", tableName, col, col, col),
				findStr, "%"+findStr+"%",
			)
		case 3:
			result = db.Exec(
				fmt.Sprintf("UPDATE %s SET %s = %s || ? WHERE %s LIKE ?", tableName, col, col, col),
				findStr, "%"+findStr+"%",
			)
		default:
			result = db.Exec(
				fmt.Sprintf("UPDATE %s SET %s = REPLACE(%s, ?, ?) WHERE %s LIKE ?", tableName, col, col, col),
				findStr, replaceStr, "%"+findStr+"%",
			)
		}
		if result.Error == nil {
			totalAffected += result.RowsAffected
		}
	}

	return totalAffected, nil
}

func (s *BackupService) DataTransfer(ctx context.Context, sourceTable, targetTable string, condition string, deleteSource bool) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if !utils.ValidateTableName(sourceTable) || !utils.ValidateTableName(targetTable) {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}
	if sourceTable == targetTable {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "源表和目标表不能相同")
	}

	var sourceRows []map[string]interface{}
	query := db.Table(sourceTable)
	if condition != "" {
		condition = strings.TrimSpace(condition)
		if strings.HasPrefix(strings.ToUpper(condition), "AND ") {
			condition = condition[4:]
		}
		query = query.Where(condition)
	}
	if err := query.Find(&sourceRows).Error; err != nil {
		return 0, errs.NewServiceError(errs.ErrDatabase, "查询源表数据失败")
	}

	if len(sourceRows) == 0 {
		return 0, errs.NewServiceError(errs.ErrResourceNotFound, "源表没有符合条件的数据")
	}

	var targetFields []FieldInfo
	targetFields, _ = s.GetTableFields(ctx, targetTable)
	targetFieldMap := make(map[string]bool)
	for _, f := range targetFields {
		targetFieldMap[f.Name] = true
	}

	var inserted int64
	const batchSize = 100
	for i := 0; i < len(sourceRows); i += batchSize {
		end := i + batchSize
		if end > len(sourceRows) {
			end = len(sourceRows)
		}
		var batch []map[string]interface{}
		for _, row := range sourceRows[i:end] {
			filtered := make(map[string]interface{})
			for k, v := range row {
				if targetFieldMap[k] && k != "id" {
					filtered[k] = v
				}
			}
			if len(filtered) > 0 {
				batch = append(batch, filtered)
			}
		}
		if len(batch) > 0 {
			result := db.Table(targetTable).Create(batch)
			if result.Error == nil {
				inserted += result.RowsAffected
			}
		}
	}

	if deleteSource && inserted > 0 {
		delQuery := db.Table(sourceTable)
		if condition != "" {
			delQuery = delQuery.Where(condition)
		}
		delQuery.Delete(nil)
	}

	return inserted, nil
}

func (s *BackupService) AdvancedExport(ctx context.Context, tableName string, fields []string, condition string, timeField string, fromDate string, toDate string, orderStr string, format string, pageSize int, page int) (string, []byte, int, int, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return "", nil, 0, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if !utils.ValidateTableName(tableName) {
		return "", nil, 0, 0, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}

	query := db.Table(tableName)

	if len(fields) > 0 {
		validFields := make([]string, 0, len(fields))
		for _, f := range fields {
			if utils.ValidateTableName(f) {
				validFields = append(validFields, f)
			}
		}
		if len(validFields) > 0 {
			query = query.Select(strings.Join(validFields, ", "))
		}
	}

	if condition != "" {
		condition = strings.TrimSpace(condition)
		if strings.HasPrefix(strings.ToUpper(condition), "AND ") {
			condition = condition[4:]
		}
		query = query.Where(condition)
	}

	if timeField != "" && utils.ValidateTableName(timeField) {
		if fromDate != "" {
			query = query.Where(fmt.Sprintf("%s >= ?", timeField), fromDate)
		}
		if toDate != "" {
			query = query.Where(fmt.Sprintf("%s <= ?", timeField), toDate+" 23:59:59")
		}
	}

	if orderStr != "" {
		query = query.Order(orderStr)
	}

	var total int64
	query.Count(&total)

	if pageSize <= 0 {
		pageSize = 5000
	}
	if pageSize > 10000 {
		pageSize = 10000
	}
	if page <= 0 {
		page = 1
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	offset := (page - 1) * pageSize
	var rows []map[string]interface{}
	if err := query.Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return "", nil, 0, 0, errs.NewServiceError(errs.ErrDatabase, "查询数据失败")
	}

	if len(rows) == 0 {
		return "", nil, 0, 0, errs.NewServiceError(errs.ErrResourceNotFound, "没有符合条件的数据")
	}

	filename := fmt.Sprintf("%s_%s.%s", tableName, time.Now().Format("20060102_150405"), format)

	var data []byte
	var err error
	switch format {
	case "csv":
		data, err = s.exportCSV(rows)
	case "json":
		data, err = s.exportJSON(rows)
	case "sql":
		data, err = s.exportSQL(rows, tableName)
	case "xml":
		data, err = s.exportXML(rows, tableName)
	default:
		return "", nil, 0, 0, errs.NewServiceError(errs.ErrParamValidation, "不支持的导出格式")
	}

	if err != nil {
		return "", nil, 0, 0, errs.NewServiceError(errs.ErrInternalServer, "生成导出文件失败")
	}

	return filename, data, totalPages, int(total), nil
}

func (s *BackupService) exportCSV(rows []map[string]interface{}) ([]byte, error) {
	var buf strings.Builder
	cols := make([]string, 0)
	for k := range rows[0] {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	buf.WriteString(strings.Join(cols, ",") + "\n")
	for _, row := range rows {
		vals := make([]string, 0, len(cols))
		for _, col := range cols {
			v := row[col]
			if v == nil {
				vals = append(vals, "")
			} else {
				str := strings.ReplaceAll(fmt.Sprintf("%v", v), "\"", "\"\"")
				vals = append(vals, "\""+str+"\"")
			}
		}
		buf.WriteString(strings.Join(vals, ",") + "\n")
	}
	return []byte(buf.String()), nil
}

func (s *BackupService) exportJSON(rows []map[string]interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *BackupService) exportSQL(rows []map[string]interface{}, tableName string) ([]byte, error) {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("-- FayHub Export %s\n", time.Now().Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("-- Table: %s (%d rows)\n\n", tableName, len(rows)))
	for _, row := range rows {
		cols := make([]string, 0, len(row))
		vals := make([]string, 0, len(row))
		for k, v := range row {
			if k == "id" {
				continue
			}
			if !utils.ValidateTableName(k) {
				continue
			}
			cols = append(cols, k)
			if v == nil {
				vals = append(vals, "NULL")
			} else {
				escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
				vals = append(vals, fmt.Sprintf("'%s'", escaped))
			}
		}
		if len(cols) > 0 {
			buf.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n", tableName, strings.Join(cols, ", "), strings.Join(vals, ", ")))
		}
	}
	return []byte(buf.String()), nil
}

func (s *BackupService) exportXML(rows []map[string]interface{}, tableName string) ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString(fmt.Sprintf("<table name=\"%s\">\n", tableName))
	for _, row := range rows {
		buf.WriteString("  <row>\n")
		for k, v := range row {
			if v == nil {
				buf.WriteString(fmt.Sprintf("    <%s>null</%s>\n", k, k))
			} else {
				escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "&", "&amp;")
				escaped = strings.ReplaceAll(escaped, "<", "&lt;")
				escaped = strings.ReplaceAll(escaped, ">", "&gt;")
				buf.WriteString(fmt.Sprintf("    <%s>%s</%s>\n", k, escaped, k))
			}
		}
		buf.WriteString("  </row>\n")
	}
	buf.WriteString("</table>")
	return []byte(buf.String()), nil
}

func (s *BackupService) PreviewTable(ctx context.Context, tableName string, limit int) ([]map[string]interface{}, []string, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if !utils.ValidateTableName(tableName) {
		return nil, nil, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var rows []map[string]interface{}
	if err := db.Table(tableName).Limit(limit).Find(&rows).Error; err != nil {
		return nil, nil, errs.NewServiceError(errs.ErrDatabase, "查询数据失败")
	}

	var cols []string
	if len(rows) > 0 {
		for k := range rows[0] {
			cols = append(cols, k)
		}
		sort.Strings(cols)
	}

	return rows, cols, nil
}

func (s *BackupService) UpdateBackupNotes(ctx context.Context, id int64, notes string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if len(notes) > 500 {
		notes = notes[:500]
	}
	return db.Model(&model.BackupRecord{}).Where("id = ?", id).Update("notes", notes).Error
}

func (s *BackupService) ExecuteWriteSQL(ctx context.Context, sqlStr string) (int64, string, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, "", errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	upperSQL := strings.ToUpper(strings.TrimSpace(sqlStr))

	for _, pattern := range sqlDangerousPatterns {
		if strings.Contains(upperSQL, pattern) {
			return 0, "", errs.NewServiceError(errs.ErrForbidden, fmt.Sprintf("禁止执行危险SQL: %s", pattern))
		}
	}

	allowed := false
	allowedPrefixes := []string{"INSERT ", "UPDATE ", "DELETE "}
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(upperSQL, prefix) {
			allowed = true
			break
		}
	}
	if !allowed {
		return 0, "", errs.NewServiceError(errs.ErrForbidden, "仅允许INSERT/UPDATE/DELETE语句")
	}

	if strings.Contains(upperSQL, ";") && !strings.HasSuffix(strings.TrimSpace(upperSQL), ";") {
		return 0, "", errs.NewServiceError(errs.ErrForbidden, "禁止执行多条SQL语句")
	}

	result := db.Exec(sqlStr)
	if result.Error != nil {
		return 0, "", errs.NewServiceError(errs.ErrDatabase, fmt.Sprintf("SQL执行失败: %v", result.Error))
	}

	opType := "操作"
	if strings.HasPrefix(upperSQL, "INSERT") {
		opType = "插入"
	} else if strings.HasPrefix(upperSQL, "UPDATE") {
		opType = "更新"
	} else if strings.HasPrefix(upperSQL, "DELETE") {
		opType = "删除"
	}

	return result.RowsAffected, opType, nil
}

func (s *BackupService) GetTableCount(ctx context.Context, tableName string, condition string) (int64, int, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}
	if !utils.ValidateTableName(tableName) {
		return 0, 0, errs.NewServiceError(errs.ErrParamValidation, "无效的表名")
	}

	query := db.Table(tableName)
	if condition != "" {
		condition = strings.TrimSpace(condition)
		if strings.HasPrefix(strings.ToUpper(condition), "AND ") {
			condition = condition[4:]
		}
		query = query.Where(condition)
	}

	var total int64
	query.Count(&total)
	return total, 0, nil
}

func (s *BackupService) cleanupOldBackups(ctx context.Context) {
	db := utils.GetDB(ctx)
	if db == nil {
		return
	}

	cfg := config.GlobalConfig
	if cfg == nil {
		return
	}

	retentionDays := cfg.Backup.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 30
	}

	maxBackups := cfg.Backup.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 100
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	var recordsToDelete []model.BackupRecord
	db.Where("created_at < ?", cutoffDate).Find(&recordsToDelete)

	var allRecords []model.BackupRecord
	db.Order("created_at DESC").Find(&allRecords)
	if len(allRecords) > maxBackups {
		recordsToDelete = append(recordsToDelete, allRecords[maxBackups:]...)
	}

	for _, record := range recordsToDelete {
		backupDir := cfg.Backup.BackupDir
		if backupDir == "" {
			backupDir = filepath.Join("data", "backups")
		}
		filePath := filepath.Join(backupDir, record.Filename)
		os.Remove(filePath)
		db.Delete(&record)
	}
}

func (s *BackupService) AutoBackup(ctx context.Context) error {
	cfg := config.GlobalConfig
	if cfg == nil || !cfg.Backup.Enabled {
		return nil
	}

	_, err := s.CreateBackup(ctx)
	return err
}
