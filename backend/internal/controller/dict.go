package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DictController struct{}

func (dc *DictController) ListDictTypes(c *gin.Context) {
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

	dictName := c.Query("dict_name")
	dictType := c.Query("dict_type")

	var status *int
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = &v
		}
	}

	ctx := c.Request.Context()
	types, total, err := service.DictServiceApp.ListTypes(ctx, dictName, dictType, status, page, pageSize)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询字典类型失败")
		return
	}

	if types == nil {
		types = []*model.DictType{}
	}

	response.GinSuccess(c, gin.H{
		"list":      types,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (dc *DictController) GetDictType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	dt, err := service.DictServiceApp.GetTypeByID(ctx, id)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询字典类型失败")
		return
	}

	response.GinSuccess(c, dt)
}

func (dc *DictController) CreateDictType(c *gin.Context) {
	var req model.DictType
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.DictServiceApp.CreateType(ctx, &req); err != nil {
		response.GinError(c, errs.ErrDatabase, "创建字典类型失败")
		return
	}

	response.GinSuccess(c, req)
}

func (dc *DictController) UpdateDictType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	var req model.DictType
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}
	req.ID = id

	ctx := c.Request.Context()
	if err := service.DictServiceApp.UpdateType(ctx, &req); err != nil {
		response.GinError(c, errs.ErrDatabase, "更新字典类型失败")
		return
	}

	response.GinSuccess(c, req)
}

func (dc *DictController) DeleteDictType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.DictServiceApp.DeleteType(ctx, id); err != nil {
		response.GinError(c, errs.ErrDatabase, "删除字典类型失败")
		return
	}

	response.GinSuccess(c, nil)
}

func (dc *DictController) ListDictData(c *gin.Context) {
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

	dictType := c.Query("dict_type")
	dictLabel := c.Query("dict_label")

	var status *int
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = &v
		}
	}

	ctx := c.Request.Context()
	data, total, err := service.DictServiceApp.ListData(ctx, dictType, dictLabel, status, page, pageSize)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询字典数据失败")
		return
	}

	if data == nil {
		data = []*model.DictData{}
	}

	response.GinSuccess(c, gin.H{
		"list":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (dc *DictController) GetDictDataByType(c *gin.Context) {
	dictType := c.Param("dict_type")
	if dictType == "" {
		response.GinError(c, errs.ErrParamMissing, "字典类型不能为空")
		return
	}

	ctx := c.Request.Context()
	data, err := service.DictServiceApp.GetDataByType(ctx, dictType)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询字典数据失败")
		return
	}

	response.GinSuccess(c, data)
}

func (dc *DictController) CreateDictData(c *gin.Context) {
	var req model.DictData
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.DictServiceApp.CreateData(ctx, &req); err != nil {
		response.GinError(c, errs.ErrDatabase, "创建字典数据失败")
		return
	}

	response.GinSuccess(c, req)
}

func (dc *DictController) UpdateDictData(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	var req model.DictData
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "请求参数错误")
		return
	}
	req.ID = id

	ctx := c.Request.Context()
	if err := service.DictServiceApp.UpdateData(ctx, &req); err != nil {
		response.GinError(c, errs.ErrDatabase, "更新字典数据失败")
		return
	}

	response.GinSuccess(c, req)
}

func (dc *DictController) DeleteDictData(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.DictServiceApp.DeleteData(ctx, id); err != nil {
		response.GinError(c, errs.ErrDatabase, "删除字典数据失败")
		return
	}

	response.GinSuccess(c, nil)
}
