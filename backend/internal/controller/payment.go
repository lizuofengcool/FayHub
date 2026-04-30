package controller

import (
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"
	"fayhub/pkg/utils"
	"strconv"

	"fayhub/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentController struct{}

func (pc *PaymentController) GetConfig(c *gin.Context) {
	ctx := c.Request.Context()

	channel := c.Query("channel")
	if channel != "" {
		config, err := service.ServiceGroupApp.PaymentService.GetConfig(ctx, channel)
		if err != nil {
			response.GinError(c, errs.ErrInternalServer, err.Error())
			return
		}
		safeConfig := gin.H{
			"channel":    config.Channel,
			"enabled":    config.Enabled,
			"mch_id":     config.MchID,
			"app_id":     config.AppID,
			"serial_no":  config.SerialNo,
			"notify_url": config.NotifyURL,
			"sandbox":    config.Sandbox,
		}
		response.GinSuccess(c, safeConfig)
		return
	}

	configs, err := service.ServiceGroupApp.PaymentService.ListConfigs(ctx)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	result := gin.H{}
	for _, cfg := range configs {
		safeCfg := gin.H{
			"enabled":    cfg.Enabled,
			"mch_id":     cfg.MchID,
			"app_id":     cfg.AppID,
			"serial_no":  cfg.SerialNo,
			"notify_url": cfg.NotifyURL,
			"sandbox":    cfg.Sandbox,
		}
		result[cfg.Channel] = safeCfg
	}

	if _, ok := result["wechat"]; !ok {
		result["wechat"] = gin.H{
			"enabled":    false,
			"mch_id":     "",
			"app_id":     "",
			"serial_no":  "",
			"notify_url": "",
			"sandbox":    false,
		}
	}
	if _, ok := result["alipay"]; !ok {
		result["alipay"] = gin.H{
			"enabled":    false,
			"app_id":     "",
			"public_key": "",
			"notify_url": "",
			"sandbox":    false,
		}
	}

	response.GinSuccess(c, result)
}

func (pc *PaymentController) UpdateConfig(c *gin.Context) {
	var req service.SaveConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PaymentService.SaveConfig(ctx, req); err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "支付配置保存成功", nil)
}

func (pc *PaymentController) CreateOrder(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	userID, _ := utils.GetUserIDFromContext(ctx)

	result, err := service.ServiceGroupApp.PaymentService.CreateOrder(ctx, userID, req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (pc *PaymentController) ListTransactions(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := service.ListOrdersRequest{
		Page:     page,
		PageSize: pageSize,
		Channel:  c.Query("channel"),
		OrderNo:  c.Query("order_no"),
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status, err := strconv.Atoi(statusStr)
		if err == nil {
			req.Status = &status
		}
	}

	orders, total, err := service.ServiceGroupApp.PaymentService.ListOrders(ctx, req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	stats, _ := service.ServiceGroupApp.PaymentService.GetStats(ctx)
	if stats == nil {
		stats = &service.PaymentStats{}
	}

	response.GinSuccess(c, gin.H{
		"list":      orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"stats":     stats,
	})
}

func (pc *PaymentController) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := service.ServiceGroupApp.PaymentService.GetStats(ctx)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, stats)
}

func (pc *PaymentController) WechatNotify(c *gin.Context) {
	ctx := c.Request.Context()

	data := make(map[string]string)
	if err := c.ShouldBindXML(&data); err != nil {
		c.XML(200, gin.H{
			"return_code": "FAIL",
			"return_msg":  "参数错误",
		})
		return
	}

	if err := service.ServiceGroupApp.PaymentService.HandleNotify(ctx, "wechat", data); err != nil {
		c.XML(200, gin.H{
			"return_code": "FAIL",
			"return_msg":  err.Error(),
		})
		return
	}

	c.XML(200, gin.H{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	})
}

func (pc *PaymentController) AlipayNotify(c *gin.Context) {
	ctx := c.Request.Context()

	data := make(map[string]string)
	c.Request.ParseForm()
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}

	if err := service.ServiceGroupApp.PaymentService.HandleNotify(ctx, "alipay", data); err != nil {
		c.String(200, "fail")
		return
	}

	c.String(200, "success")
}

func (pc *PaymentController) Refund(c *gin.Context) {
	var req service.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	result, err := service.ServiceGroupApp.PaymentService.Refund(ctx, req)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}
