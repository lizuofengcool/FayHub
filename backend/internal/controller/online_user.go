package controller

import (
	"net/http"

	"fayhub/internal/middleware"
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type OnlineUserController struct{}

func (ctrl *OnlineUserController) GetOnlineUsers(ctx *gin.Context) {
	users, err := service.ServiceGroupApp.OnlineUserService.GetOnlineUsers(ctx.Request.Context())
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, "获取在线用户失败")
		return
	}

	response.GinSuccess(ctx, users)
}

func (ctrl *OnlineUserController) ForceLogout(ctx *gin.Context) {
	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.GinError(ctx, errors.ErrParamValidation, "参数错误")
		return
	}

	tokenString, _ := middleware.GetTokenString(ctx)

	err := service.ServiceGroupApp.OnlineUserService.ForceLogoutWithToken(
		ctx.Request.Context(),
		req.UserID,
		tokenString,
	)
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(ctx, "强制下线成功", nil)
}

func (ctrl *OnlineUserController) GetOnlineCount(ctx *gin.Context) {
	count, err := service.ServiceGroupApp.OnlineUserService.GetOnlineCount(ctx.Request.Context())
	if err != nil {
		response.GinError(ctx, errors.ErrInternalServer, "获取在线人数失败")
		return
	}

	response.GinSuccess(ctx, gin.H{"count": count})
}

func (ctrl *OnlineUserController) RecordActivity(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	username, _ := ctx.Get("username")
	nickname, _ := ctx.Get("nickname")
	email, _ := ctx.Get("email")
	role, _ := ctx.Get("role")
	tenantID, _ := ctx.Get("tenant_id")

	uid, _ := userID.(int64)
	tid, _ := tenantID.(int64)

	user := model.OnlineUser{
		UserID:    uid,
		Username:  username.(string),
		Nickname:  getString(nickname),
		Email:     getString(email),
		Role:      getString(role),
		TenantID:  tid,
		IP:        ctx.ClientIP(),
		UserAgent: ctx.GetHeader("User-Agent"),
	}

	_ = service.ServiceGroupApp.OnlineUserService.RecordActivity(ctx.Request.Context(), user)

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
