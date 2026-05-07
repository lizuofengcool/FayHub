package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TenantPackageController struct{}

var TenantPackageControllerApp = new(TenantPackageController)

func (ctrl *TenantPackageController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")

	var status *int
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		status = &v
	}

	packages, total, err := service.TenantPackageServiceApp.List(c.Request.Context(), name, status, page, pageSize)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询套餐列表失败")
		return
	}

	response.GinSuccess(c, gin.H{
		"list":  packages,
		"total": total,
	})
}

func (ctrl *TenantPackageController) GetAll(c *gin.Context) {
	packages, err := service.TenantPackageServiceApp.GetAll(c.Request.Context())
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询套餐列表失败")
		return
	}

	response.GinSuccess(c, packages)
}

func (ctrl *TenantPackageController) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	pkg, err := service.TenantPackageServiceApp.GetByID(c.Request.Context(), id)
	if err != nil {
		response.GinError(c, errs.ErrDatabase, "查询套餐失败")
		return
	}

	menuIDs, _ := service.TenantPackageServiceApp.GetMenuIDs(c.Request.Context(), id)

	response.GinSuccess(c, gin.H{
		"package":  pkg,
		"menu_ids": menuIDs,
	})
}

func (ctrl *TenantPackageController) Create(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Code         string `json:"code" binding:"required"`
		Status       int    `json:"status"`
		Sort         int    `json:"sort"`
		Remark       string `json:"remark"`
		MaxUsers     int    `json:"max_users"`
		MaxStorageMB int    `json:"max_storage_mb"`
		MaxPlugins   int    `json:"max_plugins"`
		MenuIDs      []int64 `json:"menu_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "参数错误: "+err.Error())
		return
	}

	pkg := &model.TenantPackage{
		Name:         req.Name,
		Code:         req.Code,
		Status:       req.Status,
		Sort:         req.Sort,
		Remark:       req.Remark,
		MaxUsers:     req.MaxUsers,
		MaxStorageMB: req.MaxStorageMB,
		MaxPlugins:   req.MaxPlugins,
	}

	if err := service.TenantPackageServiceApp.Create(c.Request.Context(), pkg, req.MenuIDs); err != nil {
		response.GinError(c, errs.ErrDatabase, "创建套餐失败")
		return
	}

	response.GinSuccess(c, pkg)
}

func (ctrl *TenantPackageController) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	var req struct {
		Name         string `json:"name" binding:"required"`
		Code         string `json:"code" binding:"required"`
		Status       int    `json:"status"`
		Sort         int    `json:"sort"`
		Remark       string `json:"remark"`
		MaxUsers     int    `json:"max_users"`
		MaxStorageMB int    `json:"max_storage_mb"`
		MaxPlugins   int    `json:"max_plugins"`
		MenuIDs      []int64 `json:"menu_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamMissing, "参数错误: "+err.Error())
		return
	}

	pkg := &model.TenantPackage{
		SnowflakeModel: model.SnowflakeModel{ID: id},
		Name:         req.Name,
		Code:         req.Code,
		Status:       req.Status,
		Sort:         req.Sort,
		Remark:       req.Remark,
		MaxUsers:     req.MaxUsers,
		MaxStorageMB: req.MaxStorageMB,
		MaxPlugins:   req.MaxPlugins,
	}

	if err := service.TenantPackageServiceApp.Update(c.Request.Context(), pkg, req.MenuIDs); err != nil {
		response.GinError(c, errs.ErrDatabase, "更新套餐失败")
		return
	}

	response.GinSuccess(c, nil)
}

func (ctrl *TenantPackageController) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamMissing, "无效的ID")
		return
	}

	if err := service.TenantPackageServiceApp.Delete(c.Request.Context(), id); err != nil {
		response.GinError(c, errs.ErrDatabase, "删除套餐失败")
		return
	}

	response.GinSuccess(c, nil)
}
