package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	errs "fayhub/pkg/errors"
	"fayhub/pkg/redisclient"
)

type CaptchaService struct{}

const captchaPrefix = "captcha:"

func (s *CaptchaService) Generate(ctx context.Context) (string, string, error) {
	codeBytes := make([]byte, 2)
	if _, err := rand.Read(codeBytes); err != nil {
		return "", "", errs.NewServiceError(errs.ErrCaptchaFailed, "")
	}

	codeNum := int(codeBytes[0])<<8 | int(codeBytes[1])
	code := fmt.Sprintf("%04d", codeNum%10000)

	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", errs.NewServiceError(errs.ErrCaptchaFailed, "")
	}
	key := hex.EncodeToString(keyBytes)

	redisKey := captchaPrefix + key
	if err := redisclient.Set(ctx, redisKey, code, 5*time.Minute); err != nil {
		return key, code, nil
	}

	return key, code, nil
}

func (s *CaptchaService) Verify(ctx context.Context, key string, code string) (bool, error) {
	redisKey := captchaPrefix + key

	var storedCode string
	err := redisclient.Get(ctx, redisKey, &storedCode)
	if err != nil {
		return false, nil
	}

	redisclient.Del(ctx, redisKey)

	if storedCode != code {
		return false, nil
	}

	return true, nil
}
