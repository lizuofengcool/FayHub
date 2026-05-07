package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ErrorCodeController struct{}

func (ec *ErrorCodeController) ListErrorCodes(c *gin.Context) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	name := c.Query("name")

	var code *int
	if c := c.Query("code"); c != "" {
		if v, err := strconv.Atoi(c); err == nil {
			code = &v
		}
	}

	var status *int
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = &v
		}
	}

	ctx := c.Request.Context()
	codes, total, err := service.ErrorCodeServiceApp.List(ctx, name, code, status, page, pageSize)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询错误码失败")
		return
	}

	if codes == nil {
		codes = []*model.ErrorCode{}
	}

	response.GinSuccess(c, gin.H{
		"list":      codes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ec *ErrorCodeController) CreateErrorCode(c *gin.Context) {
	var req model.ErrorCode
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ErrorCodeServiceApp.Create(ctx, &req); err != nil {
		response.GinError(c, errs.ErrDatabase, "创建错误码失败")
		return
	}

	response.GinSuccess(c, req)
}

func (ec *ErrorCodeController) UpdateErrorCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	var req model.ErrorCode
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}
	req.ID = id

	ctx := c.Request.Context()
	if err := service.ErrorCodeServiceApp.Update(ctx, &req); err != nil {
		response.GinError(c, errs.ErrDatabase, "更新错误码失败")
		return
	}

	response.GinSuccess(c, req)
}

func (ec *ErrorCodeController) DeleteErrorCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.ErrorCodeServiceApp.Delete(ctx, id); err != nil {
		response.GinError(c, errs.ErrDatabase, "删除错误码失败")
		return
	}

	response.GinSuccess(c, nil)
}

func (ec *ErrorCodeController) RefreshCache(c *gin.Context) {
	ctx := c.Request.Context()
	if err := service.ErrorCodeServiceApp.RefreshCache(ctx); err != nil {
		response.GinError(c, errs.ErrDatabase, "刷新缓存失败")
		return
	}

	response.GinSuccess(c, nil)
}
