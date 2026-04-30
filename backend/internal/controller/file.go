package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type FileController struct{}

func (fc *FileController) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "请选择要上传的文件")
		return
	}
	defer file.Close()

	userID, exists := c.Get("user_id")
	if !exists {
		response.GinError(c, errs.ErrUnauthorized, "未获取到用户信息")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		response.GinError(c, errs.ErrUnauthorized, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	result, err := service.ServiceGroupApp.FileService.Upload(ctx, uid, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (fc *FileController) Download(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的文件ID")
		return
	}

	ctx := c.Request.Context()
	reader, record, err := service.ServiceGroupApp.FileService.Download(ctx, uint(id))
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", record.OriginalName))
	c.Header("Content-Type", record.MimeType)
	c.Header("Content-Length", strconv.FormatInt(record.FileSize, 10))
	c.DataFromReader(http.StatusOK, record.FileSize, record.MimeType, reader, nil)
}

func (fc *FileController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的文件ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.FileService.Delete(ctx, uint(id)); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "文件删除成功", nil)
}

func (fc *FileController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	req := service.ListFilesRequest{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
	}

	ctx := c.Request.Context()
	records, total, err := service.ServiceGroupApp.FileService.ListFiles(ctx, req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"list":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
