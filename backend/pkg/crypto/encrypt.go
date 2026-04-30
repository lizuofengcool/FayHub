package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

const encPrefix = "FAYENC:"

var (
	encKey     []byte
	encKeyOnce sync.Once
)

func InitEncryptionKey(key string) {
	encKeyOnce.Do(func() {
		if key != "" {
			k := []byte(key)
			if len(k) < 32 {
				padded := make([]byte, 32)
				copy(padded, k)
				k = padded
			} else if len(k) > 32 {
				k = k[:32]
			}
			encKey = k
		}
	})
}

func init() {
	if key := os.Getenv("FAYHUB_ENCRYPTION_KEY"); key != "" {
		InitEncryptionKey(key)
	}
}

func Encrypt(plaintext string) (string, error) {
	if encKey == nil {
		return plaintext, nil
	}

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(ciphertext string) (string, error) {
	if encKey == nil {
		return ciphertext, nil
	}

	if !strings.HasPrefix(ciphertext, encPrefix) {
		return ciphertext, nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext[len(encPrefix):])
	if err != nil {
		return ciphertext, nil
	}

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文长度不足")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}

func IsEncryptionEnabled() bool {
	return encKey != nil
}

func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, encPrefix)
}

func EncryptField(plaintext string) string {
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		return plaintext
	}
	return encrypted
}

func DecryptField(ciphertext string) string {
	decrypted, err := Decrypt(ciphertext)
	if err != nil {
		return ciphertext
	}
	return decrypted
}
