package controller

import (
	"encoding/csv"
	"encoding/json"
	"fayhub/internal/service"
	"fayhub/pkg/response"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type BackupController struct{}

func (bc *BackupController) CreateBackup(c *gin.Context) {
	ctx := c.Request.Context()
	svc := &service.BackupService{}

	record, err := svc.CreateBackup(ctx)
	if err != nil {
		response.GinError(c, 50000, "创建备份失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "备份创建成功", record)
}

func (bc *BackupController) ListBackups(c *gin.Context) {
	ctx := c.Request.Context()
	svc := &service.BackupService{}

	records, total, err := svc.ListBackups(ctx)
	if err != nil {
		response.GinError(c, 50000, "获取备份列表失败: "+err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"list":  records,
		"total": total,
	})
}

func (bc *BackupController) DeleteBackup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, 40001, "无效的备份ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	if err := svc.DeleteBackup(ctx, id); err != nil {
		response.GinError(c, 50000, "删除备份失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "备份删除成功", nil)
}

func (bc *BackupController) DownloadBackup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, 40001, "无效的备份ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	filePath, err := svc.GetBackupFilePath(ctx, id)
	if err != nil {
		response.GinError(c, 40400, "备份文件不存在")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.GinError(c, 40400, "备份文件已丢失")
		return
	}

	filename := filepath.Base(filePath)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.File(filePath)
	c.Status(http.StatusOK)
}

func (bc *BackupController) RestoreBackup(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.GinError(c, 40001, "请上传备份文件")
		return
	}

	if filepath.Ext(file.Filename) != ".sql" {
		response.GinError(c, 40001, "仅支持.sql文件")
		return
	}

	tmpDir := filepath.Join("data", "backups", "restore_tmp")
	os.MkdirAll(tmpDir, 0755)
	tmpPath := filepath.Join(tmpDir, file.Filename)
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		response.GinError(c, 50000, "保存上传文件失败")
		return
	}
	defer os.Remove(tmpPath)

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	if err := svc.RestoreBackup(ctx, tmpPath); err != nil {
		response.GinError(c, 50000, "恢复数据库失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "数据库恢复成功", nil)
}

func (bc *BackupController) RestoreBackupByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, 40001, "无效的备份ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	if err := svc.RestoreBackupByID(ctx, id); err != nil {
		response.GinError(c, 50000, "恢复备份失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "备份恢复成功", nil)
}

func (bc *BackupController) ListTables(c *gin.Context) {
	ctx := c.Request.Context()
	svc := &service.BackupService{}

	tables, err := svc.ListTables(ctx)
	if err != nil {
		response.GinError(c, 50000, "获取表列表失败: "+err.Error())
		return
	}

	response.GinSuccess(c, tables)
}

func (bc *BackupController) CreateBackupForTables(c *gin.Context) {
	var req struct {
		Tables []string `json:"tables"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "请求参数错误")
		return
	}
	if len(req.Tables) == 0 {
		response.GinError(c, 40001, "请选择要备份的表")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	record, err := svc.CreateBackupForTables(ctx, req.Tables)
	if err != nil {
		response.GinError(c, 50000, "创建备份失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "备份创建成功", record)
}

func (bc *BackupController) ExecuteSQL(c *gin.Context) {
	var req struct {
		SQL        string `json:"sql"`
		ShowErrors bool   `json:"show_errors"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "请求参数错误")
		return
	}
	if req.SQL == "" {
		response.GinError(c, 40001, "SQL语句不能为空")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	result, err := svc.ExecuteSQL(ctx, req.SQL, req.ShowErrors)
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (bc *BackupController) ListProcesses(c *gin.Context) {
	ctx := c.Request.Context()
	svc := &service.BackupService{}

	processes, err := svc.ListProcesses(ctx)
	if err != nil {
		response.GinError(c, 50000, "获取进程列表失败: "+err.Error())
		return
	}

	response.GinSuccess(c, processes)
}

func (bc *BackupController) KillProcess(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		response.GinError(c, 40001, "无效的进程ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	if err := svc.KillProcess(ctx, pid); err != nil {
		response.GinError(c, 50000, "终止进程失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "进程已终止", nil)
}

func (bc *BackupController) VerifyFields(c *gin.Context) {
	ctx := c.Request.Context()
	svc := &service.BackupService{}

	results, err := svc.VerifyFields(ctx)
	if err != nil {
		response.GinError(c, 50000, "字段校验失败: "+err.Error())
		return
	}

	response.GinSuccess(c, results)
}

func (bc *BackupController) ReplaceData(c *gin.Context) {
	var req struct {
		Table   string `json:"table"`
		Find    string `json:"find"`
		Replace string `json:"replace"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "请求参数错误")
		return
	}
	if req.Table == "" || req.Find == "" {
		response.GinError(c, 40001, "表名和查找内容不能为空")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	affected, err := svc.ReplaceData(ctx, req.Table, req.Find, req.Replace)
	if err != nil {
		response.GinError(c, 50000, "替换失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, fmt.Sprintf("替换成功，影响 %d 行", affected), gin.H{"affected": affected})
}

func (bc *BackupController) ExportTable(c *gin.Context) {
	table := c.Query("table")
	format := c.DefaultQuery("format", "csv")
	if table == "" {
		response.GinError(c, 40001, "表名不能为空")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	filename, data, err := svc.ExportTable(ctx, table, format)
	if err != nil {
		response.GinError(c, 50000, "导出失败: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (bc *BackupController) ImportData(c *gin.Context) {
	table := c.PostForm("table")
	if table == "" {
		response.GinError(c, 40001, "表名不能为空")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.GinError(c, 40001, "请上传数据文件")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".csv" && ext != ".json" {
		response.GinError(c, 40001, "仅支持.csv或.json文件")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.GinError(c, 50000, "读取文件失败")
		return
	}
	defer f.Close()

	var records []map[string]interface{}

	if ext == ".csv" {
		reader := csv.NewReader(f)
		headers, err := reader.Read()
		if err != nil {
			response.GinError(c, 40001, "CSV文件格式错误")
			return
		}
		for {
			row, err := reader.Read()
			if err != nil {
				break
			}
			record := make(map[string]interface{})
			for i, h := range headers {
				if i < len(row) {
					record[h] = row[i]
				}
			}
			records = append(records, record)
		}
	} else {
		if err := json.NewDecoder(f).Decode(&records); err != nil {
			response.GinError(c, 40001, "JSON文件格式错误")
			return
		}
	}

	if len(records) == 0 {
		response.GinError(c, 40001, "文件中没有数据")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	affected, err := svc.ImportData(ctx, table, records)
	if err != nil {
		response.GinError(c, 50000, "导入失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, fmt.Sprintf("导入成功，影响 %d 行", affected), gin.H{"affected": affected})
}

func (bc *BackupController) GetTableFields(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		response.GinError(c, 40001, "表名不能为空")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	fields, err := svc.GetTableFields(ctx, tableName)
	if err != nil {
		response.GinError(c, 50000, "获取字段信息失败: "+err.Error())
		return
	}

	response.GinSuccess(c, fields)
}

func (bc *BackupController) GetTableCount(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		response.GinError(c, 40001, "表名不能为空")
		return
	}

	condition := c.Query("condition")

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	total, _, err := svc.GetTableCount(ctx, tableName, condition)
	if err != nil {
		response.GinError(c, 50000, "获取表行数失败: "+err.Error())
		return
	}

	response.GinSuccess(c, gin.H{"total": total})
}

func (bc *BackupController) PreviewTable(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		response.GinError(c, 40001, "表名不能为空")
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	rows, columns, err := svc.PreviewTable(ctx, tableName, limit)
	if err != nil {
		response.GinError(c, 50000, "预览表失败: "+err.Error())
		return
	}

	response.GinSuccess(c, gin.H{"rows": rows, "columns": columns})
}

func (bc *BackupController) AdvancedReplace(c *gin.Context) {
	var req struct {
		Table       string `json:"table"`
		Field       string `json:"field"`
		Find        string `json:"find"`
		Replace     string `json:"replace"`
		ReplaceType int    `json:"replace_type"`
		Condition   string `json:"condition"`
		BatchSize   int    `json:"batch_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "请求参数错误")
		return
	}
	if req.Table == "" || req.Find == "" {
		response.GinError(c, 40001, "表名和查找内容不能为空")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	affected, err := svc.AdvancedReplace(ctx, req.Table, req.Field, req.Find, req.Replace, req.ReplaceType, req.Condition, req.BatchSize)
	if err != nil {
		response.GinError(c, 50000, "高级替换失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, fmt.Sprintf("替换成功，影响 %d 行", affected), gin.H{"affected": affected})
}

func (bc *BackupController) DataTransfer(c *gin.Context) {
	var req struct {
		SourceTable string `json:"source_table"`
		TargetTable string `json:"target_table"`
		Condition   string `json:"condition"`
		DeleteSource bool  `json:"delete_source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "请求参数错误")
		return
	}
	if req.SourceTable == "" || req.TargetTable == "" {
		response.GinError(c, 40001, "源表和目标表不能为空")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	affected, err := svc.DataTransfer(ctx, req.SourceTable, req.TargetTable, req.Condition, req.DeleteSource)
	if err != nil {
		response.GinError(c, 50000, "数据互转失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, fmt.Sprintf("数据互转成功，影响 %d 行", affected), gin.H{"affected": affected})
}

func (bc *BackupController) AdvancedExport(c *gin.Context) {
	table := c.Query("table")
	format := c.DefaultQuery("format", "csv")
	if table == "" {
		response.GinError(c, 40001, "表名不能为空")
		return
	}

	var fields []string
	if f := c.Query("fields"); f != "" {
		fields = strings.Split(f, ",")
	}

	pageSize := 5000
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	filename, data, _, _, err := svc.AdvancedExport(ctx, table, fields,
		c.Query("condition"), c.Query("time_field"),
		c.Query("from_date"), c.Query("to_date"),
		c.Query("order"), format, pageSize, page)
	if err != nil {
		response.GinError(c, 50000, "导出失败: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (bc *BackupController) ExecuteWriteSQL(c *gin.Context) {
	var req struct {
		SQL string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "请求参数错误")
		return
	}
	if req.SQL == "" {
		response.GinError(c, 40001, "SQL语句不能为空")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	affected, opType, err := svc.ExecuteWriteSQL(ctx, req.SQL)
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, fmt.Sprintf("%s成功，影响 %d 行", opType, affected), gin.H{"affected": affected})
}

func (bc *BackupController) UpdateBackupNotes(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, 40001, "无效的备份ID")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "请求参数错误")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	if err := svc.UpdateBackupNotes(ctx, id, req.Notes); err != nil {
		response.GinError(c, 50000, "更新备注失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "备注更新成功", nil)
}
