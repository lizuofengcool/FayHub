package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"fayhub/pkg/pluginsign"
)

var (
	testLicensePriv    *rsa.PrivateKey
	testLicensePubPath string
)

func init() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("生成测试RSA密钥对失败: " + err.Error())
	}
	testLicensePriv = priv

	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	tmpDir, _ := os.MkdirTemp("", "license_test")
	testLicensePubPath = tmpDir + "/pub.pem"
	os.WriteFile(testLicensePubPath, pubPEM, 0600)

	pluginsign.InitPublicKey(testLicensePubPath)
}

func generateLicenseRSAKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	return testLicensePriv, testLicensePubPath
}

func buildLicenseKey(priv *rsa.PrivateKey, pluginID, domain, expiry string) string {
	payload := map[string]string{
		"plugin_id": pluginID,
		"domain":    domain,
		"expiry":    expiry,
		"issued_at": time.Now().Format(time.RFC3339),
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	hash := sha256.Sum256([]byte(payloadB64))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return payloadB64 + "." + sigB64
}

func TestValidateLicenseFormat(t *testing.T) {
	s := &LicenseService{}

	if err := s.ValidateLicenseFormat("short"); err == nil {
		t.Error("短License Key应返回错误")
	}

	longKey := ""
	for i := 0; i < 20; i++ {
		longKey += "a"
	}
	if err := s.ValidateLicenseFormat(longKey); err != nil {
		t.Error("长度足够的Key应通过格式校验")
	}
}

func TestDecodeLicenseKeyValid(t *testing.T) {
	priv, _ := generateLicenseRSAKeyPair(t)

	payload := map[string]string{
		"plugin_id": "com.test.plugin",
		"domain":    "example.com",
		"expiry":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"issued_at": time.Now().Format(time.RFC3339),
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	hash := sha256.Sum256([]byte(payloadB64))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	key := payloadB64 + "." + sigB64

	decoded, err := decodeLicenseKey(key)
	if err != nil {
		t.Fatalf("decodeLicenseKey失败: %v", err)
	}

	if decoded.Data.PluginID != "com.test.plugin" {
		t.Errorf("PluginID不匹配: got %s", decoded.Data.PluginID)
	}

	if decoded.Data.Domain != "example.com" {
		t.Errorf("Domain不匹配: got %s", decoded.Data.Domain)
	}
}

func TestDecodeLicenseKeyNoDot(t *testing.T) {
	_, err := decodeLicenseKey("nodotkey")
	if err == nil {
		t.Error("无点分隔的Key应返回错误")
	}
}

func TestDecodeLicenseKeyDotAtStart(t *testing.T) {
	_, err := decodeLicenseKey(".sigpart")
	if err == nil {
		t.Error("点在开头的Key应返回错误")
	}
}

func TestDecodeLicenseKeyDotAtEnd(t *testing.T) {
	_, err := decodeLicenseKey("payload.")
	if err == nil {
		t.Error("点在结尾的Key应返回错误")
	}
}

func TestDecodeLicenseKeyInvalidBase64(t *testing.T) {
	_, err := decodeLicenseKey("not!!!base64.invalid!!!sig")
	if err == nil {
		t.Error("无效Base64应返回错误")
	}
}

func TestVerifyLicenseLocallyExpired(t *testing.T) {
	priv, _ := generateLicenseRSAKeyPair(t)

	s := &LicenseService{}

	expiredTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	key := buildLicenseKey(priv, "com.test.plugin", "example.com", expiredTime)

	resp, err := s.verifyLicenseLocally(&VerifyLicenseRequest{
		LicenseKey: key,
		PluginID:   "com.test.plugin",
		Domain:     "example.com",
	})
	if err != nil {
		t.Fatalf("verifyLicenseLocally失败: %v", err)
	}

	if resp.Valid {
		t.Error("过期License应返回Valid=false")
	}

	if resp.Message != "License已过期" {
		t.Errorf("过期消息不匹配: got %s", resp.Message)
	}
}

func TestVerifyLicenseLocallyValid(t *testing.T) {
	priv, _ := generateLicenseRSAKeyPair(t)

	s := &LicenseService{}

	futureTime := time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)
	key := buildLicenseKey(priv, "com.test.plugin", "example.com", futureTime)

	resp, err := s.verifyLicenseLocally(&VerifyLicenseRequest{
		LicenseKey: key,
		PluginID:   "com.test.plugin",
		Domain:     "example.com",
	})
	if err != nil {
		t.Fatalf("verifyLicenseLocally失败: %v", err)
	}

	if !resp.Valid {
		t.Errorf("有效License应返回Valid=true, message: %s", resp.Message)
	}
}

func TestVerifyLicenseLocallyDomainMismatch(t *testing.T) {
	priv, _ := generateLicenseRSAKeyPair(t)

	s := &LicenseService{}

	futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	key := buildLicenseKey(priv, "com.test.plugin", "licensed.com", futureTime)

	resp, err := s.verifyLicenseLocally(&VerifyLicenseRequest{
		LicenseKey: key,
		PluginID:   "com.test.plugin",
		Domain:     "other.com",
	})
	if err != nil {
		t.Fatalf("verifyLicenseLocally失败: %v", err)
	}

	if resp.Valid {
		t.Error("域名不匹配应返回Valid=false")
	}
}

func TestVerifyLicenseLocallyNoPublicKey(t *testing.T) {
	s := &LicenseService{}

	resp, err := s.verifyLicenseLocally(&VerifyLicenseRequest{
		LicenseKey: "aaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaa",
		PluginID:   "com.test",
		Domain:     "example.com",
	})
	if err != nil {
		t.Fatalf("verifyLicenseLocally失败: %v", err)
	}

	if resp.Valid {
		t.Error("公钥未初始化应返回Valid=false")
	}
}

func TestHashLicenseKey(t *testing.T) {
	key := "test-license-key-12345"
	hash1 := hashLicenseKey(key)
	hash2 := hashLicenseKey(key)

	if hash1 != hash2 {
		t.Error("相同Key的哈希应一致")
	}

	differentKey := "other-key"
	hash3 := hashLicenseKey(differentKey)
	if hash1 == hash3 {
		t.Error("不同Key的哈希应不同")
	}

	if len(hash1) != 64 {
		t.Errorf("SHA256哈希应为64字符: got %d", len(hash1))
	}
}

func TestVerifyLicenseInvalidFormat(t *testing.T) {
	s := &LicenseService{}

	resp, err := s.VerifyLicense(context.Background(), &VerifyLicenseRequest{
		LicenseKey: "short",
		PluginID:   "com.test",
		Domain:     "example.com",
	})
	if err != nil {
		t.Fatalf("VerifyLicense失败: %v", err)
	}

	if resp.Valid {
		t.Error("格式无效的License应返回Valid=false")
	}
}

func TestVerifyLicenseLocallyTamperedSignature(t *testing.T) {
	priv, _ := generateLicenseRSAKeyPair(t)

	s := &LicenseService{}

	futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	key := buildLicenseKey(priv, "com.test.plugin", "example.com", futureTime)

	tamperedKey := key[:len(key)-5] + "XXXXX"

	resp, err := s.verifyLicenseLocally(&VerifyLicenseRequest{
		LicenseKey: tamperedKey,
		PluginID:   "com.test.plugin",
		Domain:     "example.com",
	})
	if err != nil {
		t.Fatalf("verifyLicenseLocally失败: %v", err)
	}

	if resp.Valid {
		t.Error("篡改签名应返回Valid=false")
	}
}
