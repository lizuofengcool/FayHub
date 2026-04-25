package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

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

func (uc *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "无效的用户ID")
		return
	}

	if err := service.ServiceGroupApp.UserService.Delete(ctx.Request.Context(), uint(id)); err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "删除成功", nil)
}

func (uc *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "无效的用户ID")
		return
	}

	user, err := service.ServiceGroupApp.UserService.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", user)
}

func (uc *UserController) GetUserList(ctx *gin.Context) {
	var req service.UserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.GinError(ctx, errors.ParamValidationError.Code, "请求参数错误: "+err.Error())
		return
	}

	result, err := service.ServiceGroupApp.UserService.GetList(ctx.Request.Context(), req)
	if err != nil {
		response.GinError(ctx, errors.InternalServerError.Code, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "获取成功", result)
}
