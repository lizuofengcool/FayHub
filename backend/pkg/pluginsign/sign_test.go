package pluginsign

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, string, string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成RSA密钥对失败: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("序列化私钥失败: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("序列化公钥失败: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	tmpDir := t.TempDir()
	privPath := filepath.Join(tmpDir, "test_private.pem")
	pubPath := filepath.Join(tmpDir, "test_public.pem")

	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		t.Fatalf("写入私钥文件失败: %v", err)
	}
	if err := os.WriteFile(pubPath, pubPEM, 0600); err != nil {
		t.Fatalf("写入公钥文件失败: %v", err)
	}

	return priv, pubPath, privPath
}

func resetSignState() {
	publicKey = nil
	privateKey = nil
	keyLoaded = false
}

func TestInitPublicKey(t *testing.T) {
	resetSignState()
	defer resetSignState()

	_, pubPath, _ := generateTestKeyPair(t)

	if err := InitPublicKey(pubPath); err != nil {
		t.Fatalf("InitPublicKey失败: %v", err)
	}

	if !IsInitialized() {
		t.Error("InitPublicKey后IsInitialized应返回true")
	}
}

func TestInitPublicKeyEmptyPath(t *testing.T) {
	resetSignState()
	defer resetSignState()

	err := InitPublicKey("")
	if err == nil {
		t.Error("空路径应返回错误")
	}
}

func TestInitPublicKeyInvalidFile(t *testing.T) {
	resetSignState()
	defer resetSignState()

	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.pem")
	os.WriteFile(invalidPath, []byte("not a pem"), 0600)

	err := InitPublicKey(invalidPath)
	if err == nil {
		t.Error("无效PEM文件应返回错误")
	}
}

func TestInitPublicKeyIdempotent(t *testing.T) {
	resetSignState()
	defer resetSignState()

	_, pubPath, _ := generateTestKeyPair(t)

	if err := InitPublicKey(pubPath); err != nil {
		t.Fatalf("第一次InitPublicKey失败: %v", err)
	}

	if err := InitPublicKey(pubPath); err != nil {
		t.Fatalf("重复InitPublicKey应幂等，但返回: %v", err)
	}
}

func TestSignAndVerify(t *testing.T) {
	resetSignState()
	defer resetSignState()

	priv, pubPath, _ := generateTestKeyPair(t)

	if err := InitPublicKey(pubPath); err != nil {
		t.Fatalf("InitPublicKey失败: %v", err)
	}

	privateKey = priv

	data := []byte("test data for signing")

	sig, err := Sign(data)
	if err != nil {
		t.Fatalf("Sign失败: %v", err)
	}

	if err := Verify(data, sig); err != nil {
		t.Fatalf("Verify失败: %v", err)
	}
}

func TestVerifyWithTamperedData(t *testing.T) {
	resetSignState()
	defer resetSignState()

	priv, pubPath, _ := generateTestKeyPair(t)

	if err := InitPublicKey(pubPath); err != nil {
		t.Fatalf("InitPublicKey失败: %v", err)
	}

	privateKey = priv

	data := []byte("original data")
	sig, err := Sign(data)
	if err != nil {
		t.Fatalf("Sign失败: %v", err)
	}

	tamperedData := []byte("tampered data")
	if err := Verify(tamperedData, sig); err == nil {
		t.Error("篡改数据后Verify应失败")
	}
}

func TestVerifyWithInvalidSignature(t *testing.T) {
	resetSignState()
	defer resetSignState()

	_, pubPath, _ := generateTestKeyPair(t)

	if err := InitPublicKey(pubPath); err != nil {
		t.Fatalf("InitPublicKey失败: %v", err)
	}

	err := Verify([]byte("test data"), "invalid-base64!!!")
	if err == nil {
		t.Error("无效签名应导致Verify失败")
	}
}

func TestVerifyWithoutPublicKey(t *testing.T) {
	resetSignState()
	defer resetSignState()

	err := Verify([]byte("test data"), "somesig")
	if err == nil {
		t.Error("公钥未初始化时Verify应返回错误")
	}
}

func TestSignWithoutPrivateKey(t *testing.T) {
	resetSignState()
	defer resetSignState()

	_, err := Sign([]byte("test data"))
	if err == nil {
		t.Error("私钥未初始化时Sign应返回错误")
	}
}

func TestVerifyPlugin(t *testing.T) {
	resetSignState()
	defer resetSignState()

	priv, pubPath, _ := generateTestKeyPair(t)
	InitPublicKey(pubPath)
	privateKey = priv

	wasmBytes := []byte("fake wasm binary content")
	sig, err := Sign(wasmBytes)
	if err != nil {
		t.Fatalf("Sign失败: %v", err)
	}

	if err := VerifyPlugin(wasmBytes, sig); err != nil {
		t.Fatalf("VerifyPlugin失败: %v", err)
	}
}

func TestVerifyPluginNoSignature(t *testing.T) {
	resetSignState()
	defer resetSignState()

	_, pubPath, _ := generateTestKeyPair(t)
	InitPublicKey(pubPath)

	err := VerifyPlugin([]byte("wasm"), "")
	if err == nil {
		t.Error("空签名应返回错误")
	}
}

func TestVerifyPluginNoPublicKey(t *testing.T) {
	resetSignState()
	defer resetSignState()

	err := VerifyPlugin([]byte("wasm"), "somesig")
	if err != nil {
		t.Error("公钥未初始化时应跳过校验（返回nil）")
	}
}

func TestVerifyPluginHash(t *testing.T) {
	data := []byte("wasm binary content")
	hash := sha256.Sum256(data)
	expectedHash := fmt.Sprintf("%x", hash)

	if err := VerifyPluginHash(data, expectedHash); err != nil {
		t.Fatalf("VerifyPluginHash失败: %v", err)
	}
}

func TestVerifyPluginHashMismatch(t *testing.T) {
	if err := VerifyPluginHash([]byte("original"), "abc123"); err == nil {
		t.Error("哈希不匹配应返回错误")
	}
}

func TestVerifyPluginHashEmpty(t *testing.T) {
	if err := VerifyPluginHash([]byte("data"), ""); err != nil {
		t.Error("空期望哈希应跳过校验")
	}
}

func TestManualSignVerify(t *testing.T) {
	resetSignState()
	defer resetSignState()

	priv, pubPath, _ := generateTestKeyPair(t)
	InitPublicKey(pubPath)

	data := []byte("manual test payload")
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("手动签名失败: %v", err)
	}

	sigB64 := base64.StdEncoding.EncodeToString(signature)
	if err := Verify(data, sigB64); err != nil {
		t.Fatalf("手动签名验证失败: %v", err)
	}
}
