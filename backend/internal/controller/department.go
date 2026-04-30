package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DepartmentController struct{}

func (dc *DepartmentController) GetTree(c *gin.Context) {
	ctx := c.Request.Context()
	svc := &service.DepartmentService{}

	tree, err := svc.GetTree(ctx)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	if tree == nil {
		tree = []model.Department{}
	}

	response.GinSuccess(c, tree)
}

func (dc *DepartmentController) Create(c *gin.Context) {
	var req service.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	svc := &service.DepartmentService{}

	dept, err := svc.Create(ctx, req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, dept)
}

func (dc *DepartmentController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的部门ID")
		return
	}

	var req service.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	svc := &service.DepartmentService{}

	dept, err := svc.Update(ctx, uint(id), req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, dept)
}

func (dc *DepartmentController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的部门ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.DepartmentService{}

	if err := svc.Delete(ctx, uint(id)); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "删除成功", nil)
}

func (dc *DepartmentController) AssignUser(c *gin.Context) {
	deptID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的部门ID")
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的用户ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.DepartmentService{}

	if err := svc.AssignUser(ctx, uint(userID), uint(deptID)); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "分配成功", nil)
}

func (dc *DepartmentController) RemoveUser(c *gin.Context) {
	deptID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的部门ID")
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "无效的用户ID")
		return
	}

	ctx := c.Request.Context()
	svc := &service.DepartmentService{}

	if err := svc.RemoveUser(ctx, uint(userID), uint(deptID)); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "移除成功", nil)
}
