package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationController struct{}

func (nc *NotificationController) ListNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GinError(c, 40100, "未获取到用户信息")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		response.GinError(c, 40100, "用户ID格式错误")
		return
	}

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

	filters := &service.NotificationFilters{
		Type:     c.Query("type"),
		Category: c.Query("category"),
	}

	if s := c.Query("is_read"); s != "" {
		val := s == "true" || s == "1"
		filters.IsRead = &val
	}

	ctx := c.Request.Context()
	notifications, total, err := service.NotificationServiceApp.ListByUser(ctx, uid, filters, page, pageSize)
	if err != nil {
		response.GinError(c, 50000, "查询通知列表失败: "+err.Error())
		return
	}

	if notifications == nil {
		notifications = []*model.Notification{}
	}

	response.GinSuccess(c, gin.H{
		"list":      notifications,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (nc *NotificationController) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GinError(c, 40100, "未获取到用户信息")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		response.GinError(c, 40100, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	count, err := service.NotificationServiceApp.GetUnreadCount(ctx, uid)
	if err != nil {
		response.GinError(c, 50000, "查询未读数失败: "+err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"unread_count": count,
	})
}

func (nc *NotificationController) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GinError(c, 40100, "未获取到用户信息")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		response.GinError(c, 40100, "用户ID格式错误")
		return
	}

	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "参数错误: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := service.NotificationServiceApp.MarkAsRead(ctx, uid, req.IDs); err != nil {
		response.GinError(c, 50000, "标记已读失败: "+err.Error())
		return
	}

	response.GinSuccess(c, nil)
}

func (nc *NotificationController) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GinError(c, 40100, "未获取到用户信息")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		response.GinError(c, 40100, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.NotificationServiceApp.MarkAllAsRead(ctx, uid); err != nil {
		response.GinError(c, 50000, "标记全部已读失败: "+err.Error())
		return
	}

	response.GinSuccess(c, nil)
}

func (nc *NotificationController) DeleteNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GinError(c, 40100, "未获取到用户信息")
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		response.GinError(c, 40100, "用户ID格式错误")
		return
	}

	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "参数错误: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := service.NotificationServiceApp.Delete(ctx, uid, req.IDs); err != nil {
		response.GinError(c, 50000, "删除通知失败: "+err.Error())
		return
	}

	response.GinSuccess(c, nil)
}

func (nc *NotificationController) SendNotification(c *gin.Context) {
	var req service.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "参数错误: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := service.NotificationServiceApp.Send(ctx, &req); err != nil {
		response.GinError(c, 50000, "发送通知失败: "+err.Error())
		return
	}

	response.GinSuccess(c, nil)
}
