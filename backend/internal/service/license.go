package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	errs "fayhub/pkg/errors"
	"fayhub/pkg/market"
	"fayhub/pkg/pluginsign"
)

type LicenseService struct{}

type VerifyLicenseRequest struct {
	LicenseKey string `json:"license_key"`
	PluginID   string `json:"plugin_id"`
	Domain     string `json:"domain"`
}

type VerifyLicenseResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Expiry  string `json:"expiry,omitempty"`
}

type licensePayload struct {
	PluginID string `json:"plugin_id"`
	Domain   string `json:"domain"`
	Expiry   string `json:"expiry"`
	IssuedAt string `json:"issued_at"`
}

func (s *LicenseService) VerifyLicense(ctx context.Context, req *VerifyLicenseRequest) (*VerifyLicenseResponse, error) {
	if err := s.ValidateLicenseFormat(req.LicenseKey); err != nil {
		return &VerifyLicenseResponse{
			Valid:   false,
			Message: "License Key 格式无效",
		}, nil
	}

	client := market.GetClient()
	if client != nil {
		result, err := client.VerifyLicense(ctx, req.LicenseKey, req.Domain)
		if err == nil {
			if !result.Valid {
				return &VerifyLicenseResponse{
					Valid:   false,
					Message: result.Message,
				}, nil
			}

			var expiry string
			if result.ExpiresAt != "" {
				expiry = result.ExpiresAt
			}

			return &VerifyLicenseResponse{
				Valid:   true,
				Message: "License验证成功",
				Expiry:  expiry,
			}, nil
		}
	}

	return s.verifyLicenseLocally(req)
}

func (s *LicenseService) verifyLicenseLocally(req *VerifyLicenseRequest) (*VerifyLicenseResponse, error) {
	if !pluginsign.IsInitialized() {
		return &VerifyLicenseResponse{
			Valid:   false,
			Message: "Market API不可用且本地公钥未配置，无法验证License",
		}, nil
	}

	payload, err := decodeLicenseKey(req.LicenseKey)
	if err != nil {
		return &VerifyLicenseResponse{
			Valid:   false,
			Message: fmt.Sprintf("License解码失败: %v", err),
		}, nil
	}

	if err := verifyLicenseSignature(payload); err != nil {
		return &VerifyLicenseResponse{
			Valid:   false,
			Message: "License签名验证失败",
		}, nil
	}

	if payload.Data.Expiry != "" {
		expiry, err := time.Parse(time.RFC3339, payload.Data.Expiry)
		if err == nil && time.Now().After(expiry) {
			return &VerifyLicenseResponse{
				Valid:   false,
				Message: "License已过期",
			}, nil
		}
	}

	if payload.Data.Domain != "" && req.Domain != "" && payload.Data.Domain != req.Domain {
		return &VerifyLicenseResponse{
			Valid:   false,
			Message: "License域名不匹配",
		}, nil
	}

	return &VerifyLicenseResponse{
		Valid:   true,
		Message: "License本地验证成功",
		Expiry:  payload.Data.Expiry,
	}, nil
}

type decodedLicense struct {
	RawPayload   string
	SignatureB64 string
	Data         licensePayload
}

func decodeLicenseKey(licenseKey string) (*decodedLicense, error) {
	dotIdx := -1
	for i := len(licenseKey) - 1; i >= 0; i-- {
		if licenseKey[i] == '.' {
			dotIdx = i
			break
		}
	}

	if dotIdx <= 0 || dotIdx >= len(licenseKey)-1 {
		return nil, fmt.Errorf("License格式错误：缺少签名段")
	}

	rawPayload := licenseKey[:dotIdx]
	sigB64 := licenseKey[dotIdx+1:]

	_, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, fmt.Errorf("签名解码失败: %w", err)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(rawPayload)
	if err != nil {
		return nil, fmt.Errorf("载荷解码失败: %w", err)
	}

	var data licensePayload
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return nil, fmt.Errorf("载荷解析失败: %w", err)
	}

	return &decodedLicense{
		RawPayload:   rawPayload,
		SignatureB64: sigB64,
		Data:         data,
	}, nil
}

func (s *LicenseService) ValidateLicenseFormat(licenseKey string) error {
	if len(licenseKey) < 16 {
		return errs.NewServiceError(errs.ErrLicenseInvalid, "")
	}
	return nil
}

func hashLicenseKey(licenseKey string) string {
	h := sha256.Sum256([]byte(licenseKey))
	return fmt.Sprintf("%x", h)
}

func verifyLicenseSignature(payload *decodedLicense) error {
	sigBytes, err := base64.RawURLEncoding.DecodeString(payload.SignatureB64)
	if err != nil {
		return fmt.Errorf("签名RawURL解码失败: %w", err)
	}

	sigStdB64 := base64.StdEncoding.EncodeToString(sigBytes)
	return pluginsign.Verify([]byte(payload.RawPayload), sigStdB64)
}
