package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/response"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, 40001, "无效的备份ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	if err := svc.DeleteBackup(ctx, uint(id)); err != nil {
		response.GinError(c, 50000, "删除备份失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "备份删除成功", nil)
}

func (bc *BackupController) DownloadBackup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, 40001, "无效的备份ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.BackupService{}
	filePath, err := svc.GetBackupFilePath(ctx, uint(id))
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
