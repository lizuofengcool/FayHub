package service

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/config"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/eventbus"
	"fayhub/pkg/logger"
	"fayhub/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentService struct {
	wechatUnifiedOrderFn func(apiURL string, xmlBody string) (string, error)
	wechatRefundFn       func(apiURL string, xmlBody string) (string, error)
	alipayRefundFn       func(params map[string]string, payConfig *model.PaymentConfig) (string, error)
}

type CreateOrderRequest struct {
	Channel     string `json:"channel" binding:"required"`
	Amount      int64  `json:"amount" binding:"required"`
	Currency    string `json:"currency"`
	Subject     string `json:"subject" binding:"required"`
	Description string `json:"description"`
	PluginID    string `json:"plugin_id"`
	OutTradeNo  string `json:"out_trade_no"`
}

type CreateOrderResponse struct {
	OrderNo   string `json:"order_no"`
	PayURL    string `json:"pay_url"`
	QRCodeURL string `json:"qr_code_url"`
	Amount    int64  `json:"amount"`
	Channel   string `json:"channel"`
	ExpiredAt string `json:"expired_at"`
}

func (s *PaymentService) GetConfig(ctx context.Context, channel string) (*model.PaymentConfig, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var config model.PaymentConfig
	if err := queryDB.Where("channel = ?", channel).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewServiceError(errs.ErrPaymentConfigNotFound, "")
		}
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询支付配置失败")
	}

	return &config, nil
}

type SaveConfigRequest struct {
	Channel     string `json:"channel" binding:"required"`
	Enabled     *bool  `json:"enabled"`
	MchID       string `json:"mch_id"`
	AppID       string `json:"app_id"`
	APIKey      string `json:"api_key"`
	PrivateKey  string `json:"private_key"`
	PublicKey   string `json:"public_key"`
	SerialNo    string `json:"serial_no"`
	NotifyURL   string `json:"notify_url"`
	Sandbox     *bool  `json:"sandbox"`
	ExtraConfig string `json:"extra_config"`
}

func (s *PaymentService) SaveConfig(ctx context.Context, req SaveConfigRequest) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var config model.PaymentConfig
	err := queryDB.Where("channel = ?", req.Channel).First(&config).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		config = model.PaymentConfig{
			Channel: req.Channel,
		}
		if req.Enabled != nil {
			config.Enabled = *req.Enabled
		}
		config.MchID = req.MchID
		config.AppID = req.AppID
		config.APIKey = req.APIKey
		config.PrivateKey = req.PrivateKey
		config.PublicKey = req.PublicKey
		config.SerialNo = req.SerialNo
		config.NotifyURL = req.NotifyURL
		config.Sandbox = req.Sandbox != nil && *req.Sandbox
		config.ExtraConfig = req.ExtraConfig

		if err := queryDB.Create(&config).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "创建支付配置失败")
		}
		return nil
	}

	if err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "查询支付配置失败")
	}

	updates := map[string]interface{}{}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.MchID != "" {
		updates["mch_id"] = req.MchID
	}
	if req.AppID != "" {
		updates["app_id"] = req.AppID
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.PrivateKey != "" {
		updates["private_key"] = req.PrivateKey
	}
	if req.PublicKey != "" {
		updates["public_key"] = req.PublicKey
	}
	if req.SerialNo != "" {
		updates["serial_no"] = req.SerialNo
	}
	if req.NotifyURL != "" {
		updates["notify_url"] = req.NotifyURL
	}
	if req.Sandbox != nil {
		updates["sandbox"] = *req.Sandbox
	}
	if req.ExtraConfig != "" {
		updates["extra_config"] = req.ExtraConfig
	}

	if len(updates) > 0 {
		if err := queryDB.Model(&config).Updates(updates).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "更新支付配置失败")
		}
	}

	return nil
}

func (s *PaymentService) ListConfigs(ctx context.Context) ([]model.PaymentConfig, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var configs []model.PaymentConfig
	if err := queryDB.Find(&configs).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "查询支付配置失败")
	}

	return configs, nil
}

func (s *PaymentService) CreateOrder(ctx context.Context, userID int64, req CreateOrderRequest) (*CreateOrderResponse, error) {
	if req.Amount <= 0 {
		return nil, errs.NewServiceError(errs.ErrPaymentAmountInvalid, "支付金额必须大于0")
	}

	if req.Channel != model.PaymentChannelWechat && req.Channel != model.PaymentChannelAlipay {
		return nil, errs.NewServiceError(errs.ErrPaymentChannelDisabled, "不支持的支付渠道")
	}

	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	payConfig, err := s.GetConfig(ctx, req.Channel)
	if err != nil {
		return nil, err
	}

	if !payConfig.Enabled {
		return nil, errs.NewServiceError(errs.ErrPaymentChannelDisabled, "支付渠道未启用")
	}

	orderNo := generateOrderNo()
	if req.Currency == "" {
		req.Currency = "CNY"
	}

	expireMin := 30
	if cfg := getConfig(); cfg != nil && cfg.Payment.OrderExpireMin > 0 {
		expireMin = cfg.Payment.OrderExpireMin
	}
	expiredAt := time.Now().Add(time.Duration(expireMin) * time.Minute)

	order := model.PaymentOrder{
		OrderNo:     orderNo,
		Channel:     req.Channel,
		Status:      model.PaymentStatusPending,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Subject:     req.Subject,
		Description: req.Description,
		UserID:      userID,
		PluginID:    req.PluginID,
		OutTradeNo:  req.OutTradeNo,
		ExpiredAt:   &expiredAt,
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	if err := queryDB.Create(&order).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建支付订单失败")
	}

	payURL, qrCodeURL := s.buildPayURL(payConfig, &order)
	if payURL == "" {
		return nil, errs.NewServiceError(errs.ErrPaymentChannelDisabled, qrCodeURL)
	}

	return &CreateOrderResponse{
		OrderNo:   orderNo,
		PayURL:    payURL,
		QRCodeURL: qrCodeURL,
		Amount:    req.Amount,
		Channel:   req.Channel,
		ExpiredAt: expiredAt.Format(time.RFC3339),
	}, nil
}

func (s *PaymentService) buildPayURL(config *model.PaymentConfig, order *model.PaymentOrder) (string, string) {
	switch config.Channel {
	case model.PaymentChannelWechat:
		return s.buildWechatPayURL(config, order)
	case model.PaymentChannelAlipay:
		return s.buildAlipayPayURL(config, order)
	default:
		return "", ""
	}
}

func (s *PaymentService) buildWechatPayURL(payConfig *model.PaymentConfig, order *model.PaymentOrder) (string, string) {
	params := map[string]string{
		"appid":            payConfig.AppID,
		"mch_id":           payConfig.MchID,
		"nonce_str":        generateNonceStr(),
		"body":             order.Subject,
		"out_trade_no":     order.OrderNo,
		"total_fee":        strconv.FormatInt(order.Amount, 10),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       payConfig.NotifyURL,
		"trade_type":       "NATIVE",
	}

	sign := wechatSign(params, payConfig.APIKey)
	params["sign"] = sign

	buf := strings.Builder{}
	buf.WriteString("<xml>")
	for k, v := range params {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	buf.WriteString("</xml>")

	wechatGateway := "https://api.mch.weixin.qq.com"
	if cfg := getConfig(); cfg != nil && cfg.Payment.WechatGatewayURL != "" {
		wechatGateway = cfg.Payment.WechatGatewayURL
	}
	unifiedOrderURL := wechatGateway + "/pay/unifiedorder"

	callFn := s.callWechatUnifiedOrder
	if s.wechatUnifiedOrderFn != nil {
		callFn = s.wechatUnifiedOrderFn
	}
	codeURL, err := callFn(unifiedOrderURL, buf.String())
	if err != nil {
		return "", fmt.Sprintf("微信统一下单失败: %v", err)
	}
	if codeURL == "" {
		return "", "微信统一下单成功但未返回支付链接"
	}

	return codeURL, codeURL
}

type wechatUnifiedOrderResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	ResultCode string   `xml:"result_code"`
	PrepayID   string   `xml:"prepay_id"`
	TradeType  string   `xml:"trade_type"`
	CodeURL    string   `xml:"code_url"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
}

func (s *PaymentService) callWechatUnifiedOrder(apiURL string, xmlBody string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(apiURL, "application/xml; charset=utf-8", strings.NewReader(xmlBody))
	if err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "微信统一下单请求失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "读取微信响应失败")
	}

	var result wechatUnifiedOrderResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "解析微信响应失败")
	}

	if result.ReturnCode != "SUCCESS" {
		return "", errs.NewServiceError(errs.ErrPaymentSignFailed, fmt.Sprintf("微信统一下单通信失败: %s", result.ReturnMsg))
	}

	if result.ResultCode != "SUCCESS" {
		return "", errs.NewServiceError(errs.ErrPaymentSignFailed, fmt.Sprintf("微信统一下单业务失败: [%s] %s", result.ErrCode, result.ErrCodeDes))
	}

	return result.CodeURL, nil
}

func (s *PaymentService) buildAlipayPayURL(payConfig *model.PaymentConfig, order *model.PaymentOrder) (string, string) {
	baseURL := "https://openapi.alipay.com/gateway.do"
	sandboxURL := "https://openapi.alipaydev.com/gateway.do"
	if cfg := getConfig(); cfg != nil {
		if cfg.Payment.AlipayGatewayURL != "" {
			baseURL = cfg.Payment.AlipayGatewayURL
		}
		if cfg.Payment.AlipaySandboxURL != "" {
			sandboxURL = cfg.Payment.AlipaySandboxURL
		}
	}
	if payConfig.Sandbox {
		baseURL = sandboxURL
	}

	bizContent := fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"%s","product_code":"FAST_INSTANT_TRADE_PAY"}`,
		order.OrderNo, float64(order.Amount)/100, order.Subject)

	params := map[string]string{
		"app_id":      payConfig.AppID,
		"method":      "alipay.trade.page.pay",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  payConfig.NotifyURL,
		"biz_content": bizContent,
	}

	sign := alipayRsa2Sign(params, payConfig.PrivateKey)
	params["sign"] = sign

	var buf strings.Builder
	buf.WriteString(baseURL)
	buf.WriteString("?")
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteString("=")
		buf.WriteString(url.QueryEscape(params[k]))
	}

	payURL := buf.String()
	return payURL, payURL
}

func (s *PaymentService) HandleNotify(ctx context.Context, channel string, data map[string]string) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	orderNo, ok := data["out_trade_no"]
	if !ok || orderNo == "" {
		return errs.NewServiceError(errs.ErrPaymentNotifyFailed, "缺少订单号")
	}

	payConfig, err := s.GetConfig(ctx, channel)
	if err != nil {
		return err
	}

	if !verifyNotifySign(data, channel, payConfig) {
		return errs.NewServiceError(errs.ErrPaymentSignFailed, "签名验证失败")
	}

	paidAmountStr, ok := data["total_fee"]
	if !ok {
		paidAmountStr = data["total_amount"]
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	err = queryDB.Transaction(func(tx *gorm.DB) error {
		var order model.PaymentOrder
		if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return errs.NewServiceError(errs.ErrPaymentOrderNotFound, "")
		}

		if order.Status == model.PaymentStatusPaid {
			return nil
		}

		if paidAmountStr != "" {
			paidAmount, _ := strconv.ParseInt(paidAmountStr, 10, 64)
			if paidAmount > 0 && paidAmount != order.Amount {
				return errs.NewServiceError(errs.ErrPaymentAmountInvalid, "支付金额不匹配")
			}
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":      model.PaymentStatusPaid,
			"paid_at":     now,
			"notify_data": fmt.Sprintf("%v", data),
		}

		if tradeNo, ok := data["transaction_id"]; ok {
			updates["out_trade_no"] = tradeNo
		} else if tradeNo, ok := data["trade_no"]; ok {
			updates["out_trade_no"] = tradeNo
		}

		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "更新支付订单状态失败")
		}

		return nil
	})

	return err
}

type ListOrdersRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   *int   `json:"status"`
	Channel  string `json:"channel"`
	OrderNo  string `json:"order_no"`
}

func (s *PaymentService) ListOrders(ctx context.Context, req ListOrdersRequest) ([]model.PaymentOrder, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	query := queryDB.Model(&model.PaymentOrder{})

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Channel != "" {
		query = query.Where("channel = ?", req.Channel)
	}
	if req.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+req.OrderNo+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询订单数量失败")
	}

	var orders []model.PaymentOrder
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&orders).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询订单列表失败")
	}

	return orders, total, nil
}

type PaymentStats struct {
	TotalAmount       string `json:"total_amount"`
	TotalCount        int64  `json:"total_count"`
	PlatformIncome    string `json:"platform_income"`
	PendingSettlement string `json:"pending_settlement"`
	TodayAmount       string `json:"today_amount"`
	TodayCount        int64  `json:"today_count"`
}

func (s *PaymentService) GetStats(ctx context.Context) (*PaymentStats, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var totalAmount int64
	var totalCount int64
	queryDB.Model(&model.PaymentOrder{}).Where("status = ?", model.PaymentStatusPaid).Count(&totalCount)
	queryDB.Model(&model.PaymentOrder{}).Where("status = ?", model.PaymentStatusPaid).Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)

	var todayAmount int64
	var todayCount int64
	today := time.Now().Truncate(24 * time.Hour)
	queryDB.Model(&model.PaymentOrder{}).Where("status = ? AND paid_at >= ?", model.PaymentStatusPaid, today).Count(&todayCount)
	queryDB.Model(&model.PaymentOrder{}).Where("status = ? AND paid_at >= ?", model.PaymentStatusPaid, today).Select("COALESCE(SUM(amount), 0)").Scan(&todayAmount)

	platformIncome := totalAmount * 10 / 100
	pendingSettlement := totalAmount - platformIncome

	return &PaymentStats{
		TotalAmount:       formatAmount(totalAmount),
		TotalCount:        totalCount,
		PlatformIncome:    formatAmount(platformIncome),
		PendingSettlement: formatAmount(pendingSettlement),
		TodayAmount:       formatAmount(todayAmount),
		TodayCount:        todayCount,
	}, nil
}

func (s *PaymentService) CloseExpiredOrders(ctx context.Context) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	result := queryDB.Model(&model.PaymentOrder{}).
		Where("status = ? AND expired_at < ?", model.PaymentStatusPending, time.Now()).
		Update("status", model.PaymentStatusClosed)

	if result.Error != nil {
		return errs.NewServiceError(errs.ErrDatabase, "关闭过期订单失败")
	}

	if result.RowsAffected > 0 {
		eventbus.PublishAsync(eventbus.EventOrderExpired, 0, map[string]interface{}{
			"closed_count": result.RowsAffected,
		})
	}

	return nil
}

func (s *PaymentService) StartCloseExpiredOrders() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(context.Background(), "支付过期订单关闭panic", zap.Any("error", r))
			}
		}()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctx := utils.SkipTenantIsolation(context.Background())
			if err := s.CloseExpiredOrders(ctx); err != nil {
				continue
			}
		}
	}()
}

func generateOrderNo() string {
	now := time.Now()
	randBytes := make([]byte, 3)
	if _, err := rand.Read(randBytes); err != nil {
		randBytes = []byte{byte(now.Nanosecond()), byte(now.Nanosecond() >> 8), byte(now.Nanosecond() >> 16)}
	}
	suffix := int(randBytes[0])<<16 | int(randBytes[1])<<8 | int(randBytes[2])
	return fmt.Sprintf("FH%s%06d", now.Format("20060102150405"), suffix%1000000)
}

func generateNonceStr() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		now := time.Now()
		for i := range b {
			b[i] = byte(now.Nanosecond() + i)
		}
	}
	return hex.EncodeToString(b)
}

func wechatSign(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := strings.Builder{}
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}
	buf.WriteString("&key=")
	buf.WriteString(apiKey)

	h := md5.Sum([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

func wechatHMACSHA256Sign(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := strings.Builder{}
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}
	buf.WriteString("&key=")
	buf.WriteString(apiKey)

	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

func alipayRsa2Sign(params map[string]string, privateKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := strings.Builder{}
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}

	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return ""
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return ""
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return ""
	}

	hashed := sha256.Sum256([]byte(buf.String()))
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed[:])
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(signature)
}

func alipayRsa2Verify(params map[string]string, publicKey string) bool {
	signStr, ok := params["sign"]
	if !ok {
		return false
	}

	sign, err := base64.StdEncoding.DecodeString(signStr)
	if err != nil {
		return false
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := strings.Builder{}
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}

	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return false
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return false
	}

	hashed := sha256.Sum256([]byte(buf.String()))
	return rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hashed[:], sign) == nil
}

func verifyNotifySign(data map[string]string, channel string, config *model.PaymentConfig) bool {
	sign, ok := data["sign"]
	if !ok {
		return false
	}

	filtered := make(map[string]string)
	for k, v := range data {
		if k == "sign" || k == "sign_type" || v == "" {
			continue
		}
		filtered[k] = v
	}

	switch channel {
	case model.PaymentChannelWechat:
		signType := data["sign_type"]
		switch signType {
		case "HMAC-SHA256":
			return sign == wechatHMACSHA256Sign(filtered, config.APIKey)
		default:
			return sign == wechatSign(filtered, config.APIKey)
		}
	case model.PaymentChannelAlipay:
		return alipayRsa2Verify(data, config.PublicKey)
	}

	return false
}

func formatAmount(amount int64) string {
	return fmt.Sprintf("%.2f", float64(amount)/100)
}

func getConfig() *config.Config {
	if config.GlobalConfig != nil {
		return config.GlobalConfig
	}
	return nil
}

type RefundRequest struct {
	OrderNo      string `json:"order_no" binding:"required"`
	RefundNo     string `json:"refund_no" binding:"required"`
	TotalAmount  int64  `json:"total_amount" binding:"required"`
	RefundAmount int64  `json:"refund_amount" binding:"required"`
	Reason       string `json:"reason"`
}

type RefundResponse struct {
	OrderNo      string `json:"order_no"`
	RefundNo     string `json:"refund_no"`
	RefundAmount int64  `json:"refund_amount"`
	RefundStatus string `json:"refund_status"`
	Channel      string `json:"channel"`
}

func (s *PaymentService) Refund(ctx context.Context, req RefundRequest) (*RefundResponse, error) {
	if req.RefundAmount <= 0 {
		return nil, errs.NewServiceError(errs.ErrPaymentAmountInvalid, "退款金额必须大于0")
	}
	if req.RefundAmount > req.TotalAmount {
		return nil, errs.NewServiceError(errs.ErrPaymentAmountInvalid, "退款金额不能大于订单金额")
	}

	queryCtx := utils.SkipTenantIsolation(ctx)
	queryDB := utils.GetDB(queryCtx)

	var order model.PaymentOrder
	if err := queryDB.Where("order_no = ?", req.OrderNo).First(&order).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrPaymentOrderNotFound, "")
	}

	if order.Status != model.PaymentStatusPaid {
		return nil, errs.NewServiceError(errs.ErrPaymentRefundFailed, "订单状态不允许退款")
	}

	if order.RefundStatus == model.RefundStatusFull {
		return nil, errs.NewServiceError(errs.ErrPaymentRefundFailed, "订单已全额退款")
	}

	if order.RefundAmount+req.RefundAmount > order.Amount {
		return nil, errs.NewServiceError(errs.ErrPaymentAmountInvalid, "累计退款金额超过订单金额")
	}

	payConfig, err := s.GetConfig(ctx, order.Channel)
	if err != nil {
		return nil, err
	}

	var _ string
	switch order.Channel {
	case model.PaymentChannelWechat:
		_, err = s.wechatRefund(payConfig, &order, req)
	case model.PaymentChannelAlipay:
		_, err = s.alipayRefund(payConfig, &order, req)
	default:
		return nil, errs.NewServiceError(errs.ErrPaymentChannelDisabled, "不支持的退款渠道")
	}

	if err != nil {
		return nil, errs.NewServiceError(errs.ErrPaymentRefundFailed, err.Error())
	}

	err = queryDB.Transaction(func(tx *gorm.DB) error {
		newRefundAmount := order.RefundAmount + req.RefundAmount
		updates := map[string]interface{}{
			"refund_amount": newRefundAmount,
		}

		if newRefundAmount >= order.Amount {
			updates["refund_status"] = model.RefundStatusFull
			updates["status"] = model.PaymentStatusRefunded
		} else {
			updates["refund_status"] = model.RefundStatusPartial
			updates["status"] = model.PaymentStatusRefunding
		}

		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return errs.NewServiceError(errs.ErrDatabase, "更新退款状态失败")
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	refundStatus := "partial"
	if req.RefundAmount >= order.Amount {
		refundStatus = "full"
	}

	return &RefundResponse{
		OrderNo:      req.OrderNo,
		RefundNo:     req.RefundNo,
		RefundAmount: req.RefundAmount,
		RefundStatus: refundStatus,
		Channel:      order.Channel,
	}, nil
}

func (s *PaymentService) wechatRefund(payConfig *model.PaymentConfig, order *model.PaymentOrder, req RefundRequest) (string, error) {
	params := map[string]string{
		"appid":         payConfig.AppID,
		"mch_id":        payConfig.MchID,
		"nonce_str":     generateNonceStr(),
		"out_trade_no":  order.OrderNo,
		"out_refund_no": req.RefundNo,
		"total_fee":     strconv.FormatInt(req.TotalAmount, 10),
		"refund_fee":    strconv.FormatInt(req.RefundAmount, 10),
		"notify_url":    payConfig.NotifyURL,
	}

	if req.Reason != "" {
		params["refund_desc"] = req.Reason
	}

	sign := wechatSign(params, payConfig.APIKey)
	params["sign"] = sign

	buf := strings.Builder{}
	buf.WriteString("<xml>")
	for k, v := range params {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	buf.WriteString("</xml>")

	wechatGateway := "https://api.mch.weixin.qq.com"
	if cfg := getConfig(); cfg != nil && cfg.Payment.WechatGatewayURL != "" {
		wechatGateway = cfg.Payment.WechatGatewayURL
	}
	refundURL := wechatGateway + "/secapi/pay/refund"

	callFn := s.callWechatRefund
	if s.wechatRefundFn != nil {
		callFn = s.wechatRefundFn
	}
	return callFn(refundURL, buf.String())
}

func (s *PaymentService) callWechatRefund(apiURL string, xmlBody string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(apiURL, "application/xml; charset=utf-8", strings.NewReader(xmlBody))
	if err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "微信退款请求失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "读取微信退款响应失败")
	}

	var result wechatRefundResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "解析微信退款响应失败")
	}

	if result.ReturnCode != "SUCCESS" {
		return "", errs.NewServiceError(errs.ErrPaymentRefundFailed, fmt.Sprintf("微信退款通信失败: %s", result.ReturnMsg))
	}

	if result.ResultCode != "SUCCESS" {
		return "", errs.NewServiceError(errs.ErrPaymentRefundFailed, fmt.Sprintf("微信退款业务失败: [%s] %s", result.ErrCode, result.ErrCodeDes))
	}

	return result.RefundID, nil
}

type wechatRefundResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	ResultCode string   `xml:"result_code"`
	RefundID   string   `xml:"refund_id"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
}

func (s *PaymentService) alipayRefund(payConfig *model.PaymentConfig, order *model.PaymentOrder, req RefundRequest) (string, error) {
	bizContent := fmt.Sprintf(`{"out_trade_no":"%s","refund_amount":"%.2f","out_request_no":"%s"}`,
		order.OrderNo, float64(req.RefundAmount)/100, req.RefundNo)

	if req.Reason != "" {
		bizContent = fmt.Sprintf(`{"out_trade_no":"%s","refund_amount":"%.2f","out_request_no":"%s","refund_reason":"%s"}`,
			order.OrderNo, float64(req.RefundAmount)/100, req.RefundNo, req.Reason)
	}

	params := map[string]string{
		"app_id":      payConfig.AppID,
		"method":      "alipay.trade.refund",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}

	sign := alipayRsa2Sign(params, payConfig.PrivateKey)
	params["sign"] = sign

	callFn := s.callAlipayRefund
	if s.alipayRefundFn != nil {
		callFn = s.alipayRefundFn
	}
	return callFn(params, payConfig)
}

func (s *PaymentService) callAlipayRefund(params map[string]string, payConfig *model.PaymentConfig) (string, error) {
	baseURL := "https://openapi.alipay.com/gateway.do"
	if payConfig.Sandbox {
		baseURL = "https://openapi.alipaydev.com/gateway.do"
	}
	if cfg := getConfig(); cfg != nil {
		if cfg.Payment.AlipayGatewayURL != "" {
			baseURL = cfg.Payment.AlipayGatewayURL
		}
	}

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(baseURL, form)
	if err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "支付宝退款请求失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "读取支付宝退款响应失败")
	}

	var result alipayRefundResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", errs.NewServiceError(errs.ErrExternalService, "解析支付宝退款响应失败")
	}

	respData := result.AlipayTradeRefundResponse
	if respData.Code != "10000" {
		return "", errs.NewServiceError(errs.ErrPaymentRefundFailed, fmt.Sprintf("支付宝退款失败: [%s] %s", respData.SubCode, respData.SubMsg))
	}

	return respData.TradeNo, nil
}

type alipayRefundResponse struct {
	AlipayTradeRefundResponse struct {
		Code      string `json:"code"`
		Msg       string `json:"msg"`
		SubCode   string `json:"sub_code"`
		SubMsg    string `json:"sub_msg"`
		TradeNo   string `json:"trade_no"`
		RefundFee string `json:"refund_fee"`
	} `json:"alipay_trade_refund_response"`
}
