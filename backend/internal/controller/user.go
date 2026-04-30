package controller

import (
	"fayhub/internal/middleware"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

// CreateUser godoc
// @Summary 创建用户
// @Description 在当前租户下创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.CreateUserRequest true "创建用户请求参数"
// @Success 200 {object} response.Response "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users [post]
func (uc *UserController) CreateUser(ctx *gin.Context) {
	var req service.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	user, err := service.ServiceGroupApp.UserService.Create(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "创建成功", user)
}

// UpdateUser godoc
// @Summary 更新用户
// @Description 根据用户ID更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body service.UpdateUserRequest true "更新用户请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users/{id} [put]
func (uc *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "无效的用户ID")
		return
	}

	var req service.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	user, err := service.ServiceGroupApp.UserService.Update(ctx.Request.Context(), uint(id), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "更新成功", user)
}

// DeleteUser godoc
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users/{id} [delete]
func (uc *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "无效的用户ID")
		return
	}

	if err := service.ServiceGroupApp.UserService.Delete(ctx.Request.Context(), uint(id)); err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

// GetUser godoc
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users/{id} [get]
func (uc *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "无效的用户ID")
		return
	}

	user, err := service.ServiceGroupApp.UserService.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", user)
}

// GetUserList godoc
// @Summary 获取用户列表
// @Description 分页查询当前租户下的用户列表，支持关键词搜索和角色筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Param role query string false "角色筛选"
// @Param status query int false "状态筛选(0:禁用,1:正常)"
// @Success 200 {object} response.Response "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users [get]
func (uc *UserController) GetUserList(ctx *gin.Context) {
	var req service.UserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	result, err := service.ServiceGroupApp.UserService.GetList(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", result)
}

// ChangePassword godoc
// @Summary 修改密码
// @Description 当前登录用户修改自己的密码，需要提供原密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object{old_password=string,new_password=string} true "修改密码请求参数"
// @Success 200 {object} response.Response "密码修改成功"
// @Failure 400 {object} response.Response "请求参数错误或原密码错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users/change-password [put]
func (uc *UserController) ChangePassword(ctx *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(ctx)
	if !exists {
		response.GinError(ctx, errors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	err := service.ServiceGroupApp.UserService.ChangePassword(ctx.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "原密码错误" {
			response.GinError(ctx, errors.ErrParamValidation, err.Error())
		} else {
			response.GinError(ctx, errors.ErrInternalServer, err.Error())
		}
		return
	}

	response.GinSuccessWithMessage(ctx, "密码修改成功", nil)
}

// ResetPassword godoc
// @Summary 重置密码
// @Description 管理员重置指定用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body service.ResetPasswordRequest true "重置密码请求参数"
// @Success 200 {object} response.Response "密码重置成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users/{id}/reset-password [put]
func (uc *UserController) ResetPassword(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "无效的用户ID")
		return
	}

	var req service.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "请求参数错误: "+err.Error())
		return
	}

	err = service.ServiceGroupApp.UserService.ResetPassword(ctx.Request.Context(), uint(id), req.NewPassword)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "密码重置成功", nil)
}

// GetProfile godoc
// @Summary 获取个人资料
// @Description 获取当前登录用户的个人资料
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/users/profile [get]
func (uc *UserController) GetProfile(ctx *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(ctx)
	if !exists {
		response.GinError(ctx, errors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	user, err := service.ServiceGroupApp.UserService.GetProfile(ctx.Request.Context(), userID)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", user)
}
