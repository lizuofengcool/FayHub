package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type PaymentTestSuite struct {
	suite.Suite
	db  *gorm.DB
	ctx context.Context
}

func (s *PaymentTestSuite) SetupSuite() {
	db, err := openTestDB()
	s.Require().NoError(err)

	s.db = db
	utils.SetGlobalDB(db)

	ctx := utils.SkipTenantIsolation(context.Background())
	s.ctx = ctx

	err = db.WithContext(ctx).AutoMigrate(
		&model.PaymentConfig{},
		&model.PaymentOrder{},
	)
	s.Require().NoError(err)
}

func (s *PaymentTestSuite) TearDownSuite() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (s *PaymentTestSuite) TestSaveConfig_CreateNew() {
	svc := &PaymentService{}

	err := svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "save_new_channel",
		Enabled: boolPtr(true),
		MchID:   "mch_001",
		AppID:   "wx_app_001",
		APIKey:  "test_api_key",
	})

	s.Assert().NoError(err)

	var config model.PaymentConfig
	s.db.WithContext(s.ctx).Where("channel = ?", "save_new_channel").First(&config)
	s.Assert().Equal("mch_001", config.MchID)
	s.Assert().Equal("wx_app_001", config.AppID)
	s.Assert().True(config.Enabled)
}

func (s *PaymentTestSuite) TestSaveConfig_UpdateExisting() {
	svc := &PaymentService{}

	err := svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "update_channel",
		Enabled: boolPtr(true),
		AppID:   "app_old",
	})
	s.Require().NoError(err)

	err = svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "update_channel",
		AppID:   "app_new",
	})
	s.Assert().NoError(err)

	var config model.PaymentConfig
	s.db.WithContext(s.ctx).Where("channel = ?", "update_channel").First(&config)
	s.Assert().Equal("app_new", config.AppID)
}

func (s *PaymentTestSuite) TestGetConfig_Success() {
	svc := &PaymentService{}

	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "get_channel",
		Enabled: boolPtr(true),
		MchID:   "mch_get",
	})

	config, err := svc.GetConfig(s.ctx, "get_channel")

	s.Assert().NoError(err)
	s.Assert().NotNil(config)
	s.Assert().Equal("mch_get", config.MchID)
}

func (s *PaymentTestSuite) TestGetConfig_NotFound() {
	svc := &PaymentService{}

	config, err := svc.GetConfig(s.ctx, "nonexistent_channel")

	s.Assert().Error(err)
	s.Assert().Nil(config)
	s.Assert().Contains(err.Error(), "支付配置不存在")
}

func (s *PaymentTestSuite) TestListConfigs() {
	svc := &PaymentService{}

	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "list_ch_1",
		Enabled: boolPtr(true),
	})
	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "list_ch_2",
		Enabled: boolPtr(false),
	})

	configs, err := svc.ListConfigs(s.ctx)

	s.Assert().NoError(err)
	s.Assert().True(len(configs) >= 2)
}

func (s *PaymentTestSuite) TestCreateOrder_InvalidAmount() {
	svc := &PaymentService{}

	resp, err := svc.CreateOrder(s.ctx, 1, CreateOrderRequest{
		Channel: "wechat",
		Amount:  0,
		Subject: "test",
	})

	s.Assert().Error(err)
	s.Assert().Nil(resp)
	s.Assert().Contains(err.Error(), "支付金额必须大于0")
}

func (s *PaymentTestSuite) TestCreateOrder_UnsupportedChannel() {
	svc := &PaymentService{}

	resp, err := svc.CreateOrder(s.ctx, 1, CreateOrderRequest{
		Channel: "bank_transfer",
		Amount:  100,
		Subject: "test",
	})

	s.Assert().Error(err)
	s.Assert().Nil(resp)
	s.Assert().Contains(err.Error(), "不支持的支付渠道")
}

func (s *PaymentTestSuite) TestCreateOrder_ChannelNotEnabled() {
	svc := &PaymentService{}

	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "wechat",
		Enabled: boolPtr(false),
		MchID:   "mch_disabled",
		AppID:   "wx_disabled",
		APIKey:  "disabled_key",
	})

	resp, err := svc.CreateOrder(s.ctx, 1, CreateOrderRequest{
		Channel: "wechat",
		Amount:  100,
		Subject: "test",
	})

	s.Assert().Error(err)
	s.Assert().Nil(resp)
	s.Assert().Contains(err.Error(), "支付渠道未启用")
}

func (s *PaymentTestSuite) TestCreateOrder_ConfigNotFound() {
	svc := &PaymentService{}

	s.db.WithContext(s.ctx).Where("channel = ?", "alipay").Delete(&model.PaymentConfig{})

	resp, err := svc.CreateOrder(s.ctx, 1, CreateOrderRequest{
		Channel: "alipay",
		Amount:  100,
		Subject: "test",
	})

	s.Assert().Error(err)
	s.Assert().Nil(resp)
}

func (s *PaymentTestSuite) TestCreateOrder_WechatSuccess() {
	svc := &PaymentService{
		wechatUnifiedOrderFn: func(apiURL string, xmlBody string) (string, error) {
			return "weixin://wxpay/bizpayurl?pr=testcode123", nil
		},
	}

	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel:   "wechat",
		Enabled:   boolPtr(true),
		MchID:     "mch_order",
		AppID:     "wx_order",
		APIKey:    "order_key",
		NotifyURL: "https://example.com/notify",
	})

	resp, err := svc.CreateOrder(s.ctx, 1, CreateOrderRequest{
		Channel: "wechat",
		Amount:  9900,
		Subject: "测试商品",
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(resp)
	s.Assert().NotEmpty(resp.OrderNo)
	s.Assert().Equal(int64(9900), resp.Amount)
	s.Assert().Equal("wechat", resp.Channel)
	s.Assert().NotEmpty(resp.ExpiredAt)

	var order model.PaymentOrder
	s.db.WithContext(s.ctx).Where("order_no = ?", resp.OrderNo).First(&order)
	s.Assert().Equal(model.PaymentStatusPending, order.Status)
	s.Assert().Equal(int64(9900), order.Amount)
	s.Assert().Equal("CNY", order.Currency)
}

func (s *PaymentTestSuite) TestCreateOrder_AlipaySuccess() {
	svc := &PaymentService{}

	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel:   "alipay",
		Enabled:   boolPtr(true),
		AppID:     "alipay_order",
		NotifyURL: "https://example.com/notify",
	})

	resp, err := svc.CreateOrder(s.ctx, 1, CreateOrderRequest{
		Channel: "alipay",
		Amount:  5000,
		Subject: "支付宝测试",
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(resp)
	s.Assert().Equal("alipay", resp.Channel)
}

func (s *PaymentTestSuite) TestHandleNotify_MissingOrderNo() {
	svc := &PaymentService{}

	err := svc.HandleNotify(s.ctx, "wechat", map[string]string{})

	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "缺少订单号")
}

func (s *PaymentTestSuite) TestHandleNotify_SignVerificationFails() {
	svc := &PaymentService{}

	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "wechat",
		Enabled: boolPtr(true),
		APIKey:  "real_key",
	})

	err := svc.HandleNotify(s.ctx, "wechat", map[string]string{
		"out_trade_no": "SOME_ORDER",
		"total_fee":    "100",
		"sign":         "wrong_sign",
	})

	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "签名验证失败")
}

func (s *PaymentTestSuite) TestHandleNotify_DuplicatePaidIgnored() {
	svc := &PaymentService{}

	apiKey := "dup_test_key"
	svc.SaveConfig(s.ctx, SaveConfigRequest{
		Channel: "wechat",
		Enabled: boolPtr(true),
		APIKey:  apiKey,
	})

	now := time.Now()
	order := &model.PaymentOrder{
		OrderNo:  "FH_DUP_PAID_001",
		Channel:  "wechat",
		Status:   model.PaymentStatusPaid,
		Amount:   100,
		Currency: "CNY",
		Subject:  "已支付订单",
		PaidAt:   &now,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(order).Error)

	notifyData := map[string]string{
		"out_trade_no": "FH_DUP_PAID_001",
		"total_fee":    "100",
	}
	sign := wechatSign(notifyData, apiKey)
	notifyData["sign"] = sign

	err := svc.HandleNotify(s.ctx, "wechat", notifyData)

	s.Assert().NoError(err)
}

func (s *PaymentTestSuite) TestCloseExpiredOrders() {
	svc := &PaymentService{}

	expiredAt := time.Now().Add(-1 * time.Hour)
	order := &model.PaymentOrder{
		OrderNo:   "FH_EXPIRED_001",
		Channel:   "wechat",
		Status:    model.PaymentStatusPending,
		Amount:    100,
		Currency:  "CNY",
		Subject:   "过期订单",
		ExpiredAt: &expiredAt,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(order).Error)

	err := svc.CloseExpiredOrders(s.ctx)
	s.Assert().NoError(err)

	var updated model.PaymentOrder
	s.db.WithContext(s.ctx).Where("order_no = ?", "FH_EXPIRED_001").First(&updated)
	s.Assert().Equal(model.PaymentStatusClosed, updated.Status)
}

func (s *PaymentTestSuite) TestCloseExpiredOrders_DoesNotCloseActiveOrders() {
	svc := &PaymentService{}

	futureExpiry := time.Now().Add(1 * time.Hour)
	order := &model.PaymentOrder{
		OrderNo:   "FH_ACTIVE_001",
		Channel:   "wechat",
		Status:    model.PaymentStatusPending,
		Amount:    100,
		Currency:  "CNY",
		Subject:   "活跃订单",
		ExpiredAt: &futureExpiry,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(order).Error)

	err := svc.CloseExpiredOrders(s.ctx)
	s.Assert().NoError(err)

	var updated model.PaymentOrder
	s.db.WithContext(s.ctx).Where("order_no = ?", "FH_ACTIVE_001").First(&updated)
	s.Assert().Equal(model.PaymentStatusPending, updated.Status)
}

func (s *PaymentTestSuite) TestListOrders_Pagination() {
	svc := &PaymentService{}

	for i := 0; i < 5; i++ {
		s.db.WithContext(s.ctx).Create(&model.PaymentOrder{
			OrderNo:  "FH_LISTORD_" + string(rune('A'+i)),
			Channel:  "wechat",
			Status:   model.PaymentStatusPending,
			Amount:   100,
			Currency: "CNY",
			Subject:  "列表订单",
		})
	}

	orders, total, err := svc.ListOrders(s.ctx, ListOrdersRequest{
		Page:     1,
		PageSize: 3,
	})

	s.Assert().NoError(err)
	s.Assert().True(total >= 5)
	s.Assert().True(len(orders) <= 3)
}

func (s *PaymentTestSuite) TestListOrders_FilterByStatus() {
	svc := &PaymentService{}

	s.db.WithContext(s.ctx).Create(&model.PaymentOrder{
		OrderNo:  "FH_FILTER_PAID",
		Channel:  "wechat",
		Status:   model.PaymentStatusPaid,
		Amount:   200,
		Currency: "CNY",
		Subject:  "已支付",
	})

	paidStatus := model.PaymentStatusPaid
	orders, _, err := svc.ListOrders(s.ctx, ListOrdersRequest{
		Page:     1,
		PageSize: 10,
		Status:   &paidStatus,
	})

	s.Assert().NoError(err)
	for _, o := range orders {
		s.Assert().Equal(model.PaymentStatusPaid, o.Status)
	}
}

func (s *PaymentTestSuite) TestGetStats() {
	svc := &PaymentService{}

	now := time.Now()
	s.db.WithContext(s.ctx).Create(&model.PaymentOrder{
		OrderNo:  "FH_STATS_002",
		Channel:  "wechat",
		Status:   model.PaymentStatusPaid,
		Amount:   10000,
		Currency: "CNY",
		Subject:  "统计订单",
		PaidAt:   &now,
	})

	stats, err := svc.GetStats(s.ctx)

	s.Assert().NoError(err)
	s.Assert().NotNil(stats)
	s.Assert().True(stats.TotalCount >= 1)
}

func (s *PaymentTestSuite) TestGenerateOrderNo_Unique() {
	orderNos := make(map[string]bool)
	for i := 0; i < 100; i++ {
		no := generateOrderNo()
		s.Assert().NotEmpty(no)
		s.Assert().True(len(no) > 10)
		orderNos[no] = true
	}
	s.Assert().True(len(orderNos) > 90, "order numbers should be mostly unique")
}

func (s *PaymentTestSuite) TestWechatSign() {
	params := map[string]string{
		"appid":     "wx123",
		"mch_id":    "mch456",
		"nonce_str": "abc123",
	}
	sign := wechatSign(params, "test_api_key")
	s.Assert().NotEmpty(sign)
	s.Assert().Len(sign, 32)
}

func (s *PaymentTestSuite) TestWechatSign_Consistency() {
	params := map[string]string{
		"appid":     "wx_consistency",
		"mch_id":    "mch_consistency",
		"nonce_str": "fixed_nonce",
	}
	sign1 := wechatSign(params, "same_key")
	sign2 := wechatSign(params, "same_key")
	s.Assert().Equal(sign1, sign2)
}

func (s *PaymentTestSuite) TestFormatAmount() {
	s.Assert().Equal("1.00", formatAmount(100))
	s.Assert().Equal("0.01", formatAmount(1))
	s.Assert().Equal("99.99", formatAmount(9999))
	s.Assert().Equal("0.00", formatAmount(0))
}

func (s *PaymentTestSuite) TestVerifyNotifySign_WechatMD5() {
	apiKey := "test_api_key_123456"
	params := map[string]string{
		"appid":     "wx123",
		"mch_id":    "mch456",
		"nonce_str": "abc123",
	}
	sign := wechatSign(params, apiKey)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	payConfig := &model.PaymentConfig{APIKey: apiKey}
	result := verifyNotifySign(params, "wechat", payConfig)
	s.Assert().True(result, "微信MD5签名验证应通过")
}

func (s *PaymentTestSuite) TestVerifyNotifySign_WechatHMACSHA256() {
	apiKey := "test_api_key_123456"
	params := map[string]string{
		"appid":     "wx123",
		"mch_id":    "mch456",
		"nonce_str": "abc123",
	}
	sign := wechatHMACSHA256Sign(params, apiKey)
	params["sign"] = sign
	params["sign_type"] = "HMAC-SHA256"

	payConfig := &model.PaymentConfig{APIKey: apiKey}
	result := verifyNotifySign(params, "wechat", payConfig)
	s.Assert().True(result, "微信HMAC-SHA256签名验证应通过")
}

func (s *PaymentTestSuite) TestVerifyNotifySign_WechatWrongKey() {
	params := map[string]string{
		"appid":     "wx123",
		"mch_id":    "mch456",
		"nonce_str": "abc123",
	}
	sign := wechatSign(params, "correct_key")
	params["sign"] = sign

	payConfig := &model.PaymentConfig{APIKey: "wrong_key"}
	result := verifyNotifySign(params, "wechat", payConfig)
	s.Assert().False(result, "错误API Key签名验证应失败")
}

func (s *PaymentTestSuite) TestVerifyNotifySign_NoSign() {
	params := map[string]string{
		"appid":  "wx123",
		"mch_id": "mch456",
	}
	payConfig := &model.PaymentConfig{APIKey: "key"}
	result := verifyNotifySign(params, "wechat", payConfig)
	s.Assert().False(result, "无签名字段应返回false")
}

func (s *PaymentTestSuite) TestVerifyNotifySign_TamperedData() {
	apiKey := "test_api_key_123456"
	params := map[string]string{
		"appid":     "wx123",
		"mch_id":    "mch456",
		"nonce_str": "abc123",
	}
	sign := wechatSign(params, apiKey)
	params["sign"] = sign
	params["total_fee"] = "9999"

	payConfig := &model.PaymentConfig{APIKey: apiKey}
	result := verifyNotifySign(params, "wechat", payConfig)
	s.Assert().False(result, "篡改数据后签名验证应失败")
}

func (s *PaymentTestSuite) TestWechatHMACSHA256Sign() {
	params := map[string]string{
		"appid":     "wx_hmac",
		"mch_id":    "mch_hmac",
		"nonce_str": "hmac_nonce",
	}
	sign := wechatHMACSHA256Sign(params, "hmac_key")
	s.Assert().NotEmpty(sign)
}

func (s *PaymentTestSuite) TestWechatHMACSHA256Sign_Consistency() {
	params := map[string]string{
		"appid":     "wx_cons",
		"mch_id":    "mch_cons",
		"nonce_str": "fixed",
	}
	sign1 := wechatHMACSHA256Sign(params, "same_key")
	sign2 := wechatHMACSHA256Sign(params, "same_key")
	s.Assert().Equal(sign1, sign2)
}

func (s *PaymentTestSuite) TestAlipayRsa2Verify_InvalidPublicKey() {
	params := map[string]string{
		"app_id":    "alipay123",
		"trade_no":  "T202604300001",
		"sign":      base64.StdEncoding.EncodeToString([]byte("fake_signature")),
		"sign_type": "RSA2",
	}
	result := alipayRsa2Verify(params, "not_a_valid_pem_public_key")
	s.Assert().False(result, "无效公钥应验证失败")
}

func (s *PaymentTestSuite) TestAlipayRsa2Verify_NoSign() {
	params := map[string]string{
		"app_id":   "alipay123",
		"trade_no": "T202604300001",
	}
	result := alipayRsa2Verify(params, "some_key")
	s.Assert().False(result, "无签名字段应返回false")
}

func (s *PaymentTestSuite) TestAlipayRsa2Verify_ValidSignature() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	s.Require().NoError(err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	params := map[string]string{
		"app_id":    "alipay_test",
		"trade_no":  "T202604300001",
		"sign_type": "RSA2",
	}

	keys := []string{"app_id", "trade_no"}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}

	hashed := sha256.Sum256([]byte(buf.String()))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	s.Require().NoError(err)

	params["sign"] = base64.StdEncoding.EncodeToString(sig)

	result := alipayRsa2Verify(params, string(pubPEM))
	s.Assert().True(result, "RSA2签名验证应通过")
}

func (s *PaymentTestSuite) TestAlipayRsa2Verify_TamperedData() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	s.Require().NoError(err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	params := map[string]string{
		"app_id":    "alipay_test",
		"trade_no":  "T202604300001",
		"sign_type": "RSA2",
	}

	keys := []string{"app_id", "trade_no"}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
	}

	hashed := sha256.Sum256([]byte(buf.String()))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	s.Require().NoError(err)

	params["sign"] = base64.StdEncoding.EncodeToString(sig)
	params["total_amount"] = "9999.00"

	result := alipayRsa2Verify(params, string(pubPEM))
	s.Assert().False(result, "篡改数据后RSA2签名验证应失败")
}

func boolPtr(b bool) *bool {
	return &b
}

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}
