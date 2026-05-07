package controller

import (
	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WebhookController struct{}

func (wc *WebhookController) CreateSubscription(c *gin.Context) {
	var req struct {
		Name       string   `json:"name" binding:"required"`
		URL        string   `json:"url" binding:"required,url"`
		Secret     string   `json:"secret"`
		Events     []string `json:"events" binding:"required,min=1"`
		RetryCount int      `json:"retry_count"`
		TimeoutSec int      `json:"timeout_sec"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "参数错误: "+err.Error())
		return
	}

	tenantID, _ := c.Get("tenant_id")
	tenantIDUint, _ := tenantID.(int64)

	sub := &model.WebhookSubscription{
		Name:       req.Name,
		URL:        req.URL,
		Secret:     req.Secret,
		Events:     req.Events,
		RetryCount: req.RetryCount,
		TimeoutSec: req.TimeoutSec,
		IsActive:   true,
	}
	sub.TenantID = tenantIDUint

	if sub.RetryCount <= 0 {
		sub.RetryCount = 3
	}
	if sub.TimeoutSec <= 0 {
		sub.TimeoutSec = 10
	}

	ctx := c.Request.Context()
	if err := service.WebhookServiceApp.CreateSubscription(ctx, sub); err != nil {
		response.GinError(c, 50000, "创建订阅失败: "+err.Error())
		return
	}

	response.GinSuccess(c, sub)
}

func (wc *WebhookController) ListSubscriptions(c *gin.Context) {
	event := c.Query("event")
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

	ctx := c.Request.Context()
	subs, total, err := service.WebhookServiceApp.ListSubscriptions(ctx, event, page, pageSize)
	if err != nil {
		response.GinError(c, 50000, "查询订阅列表失败: "+err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"list":  subs,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func (wc *WebhookController) GetSubscription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, 40001, "无效的订阅ID")
		return
	}

	ctx := c.Request.Context()
	sub, err := service.WebhookServiceApp.GetSubscription(ctx, id)
	if err != nil {
		response.GinError(c, 40400, "订阅不存在")
		return
	}

	response.GinSuccess(c, sub)
}

func (wc *WebhookController) UpdateSubscription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, 40001, "无效的订阅ID")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40001, "参数错误")
		return
	}

	allowed := map[string]bool{
		"name": true, "url": true, "secret": true,
		"events": true, "is_active": true,
		"retry_count": true, "timeout_sec": true,
	}

	updates := make(map[string]interface{})
	for k, v := range req {
		if allowed[k] {
			updates[k] = v
		}
	}

	if len(updates) == 0 {
		response.GinError(c, 40001, "没有可更新的字段")
		return
	}

	ctx := c.Request.Context()
	if err := service.WebhookServiceApp.UpdateSubscription(ctx, id, updates); err != nil {
		response.GinError(c, 50000, "更新订阅失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "更新成功", nil)
}

func (wc *WebhookController) DeleteSubscription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, 40001, "无效的订阅ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.WebhookServiceApp.DeleteSubscription(ctx, id); err != nil {
		response.GinError(c, 50000, "删除订阅失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "删除成功", nil)
}

func (wc *WebhookController) ListDeliveries(c *gin.Context) {
	subID, _ := strconv.ParseInt(c.Query("subscription_id"), 10, 32)
	status := c.Query("status")
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

	ctx := c.Request.Context()
	deliveries, total, err := service.WebhookServiceApp.ListDeliveries(ctx, subID, status, page, pageSize)
	if err != nil {
		response.GinError(c, 50000, "查询投递记录失败: "+err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"list":  deliveries,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func (wc *WebhookController) Redeliver(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		response.GinError(c, 40001, "无效的投递ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.WebhookServiceApp.Redeliver(ctx, id); err != nil {
		response.GinError(c, 50000, "重新投递失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "已触发重新投递", nil)
}

func (wc *WebhookController) GetDeliveryStats(c *gin.Context) {
	subID, _ := strconv.ParseInt(c.Query("subscription_id"), 10, 32)

	ctx := c.Request.Context()
	stats, err := service.WebhookServiceApp.GetDeliveryStats(ctx, subID)
	if err != nil {
		response.GinError(c, 50000, "获取统计失败: "+err.Error())
		return
	}

	response.GinSuccess(c, stats)
}

func (wc *WebhookController) TestDelivery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.GinError(c, 40001, "无效的订阅ID")
		return
	}

	ctx := c.Request.Context()
	delivery, err := service.WebhookServiceApp.TestDelivery(ctx, id)
	if err != nil {
		response.GinError(c, 50000, "测试投递失败: "+err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "测试投递已触发", delivery)
}
