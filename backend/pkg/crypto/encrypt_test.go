package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	InitEncryptionKey("test-encryption-key-32-bytes-long!!")

	plaintext := "hello world"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if encrypted == plaintext {
		t.Error("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestEncryptEmptyString(t *testing.T) {
	InitEncryptionKey("another-test-key-32-bytes-long!")

	encrypted, err := Encrypt("")
	if err != nil {
		t.Fatalf("encrypt empty string failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt empty string failed: %v", err)
	}

	if decrypted != "" {
		t.Errorf("expected empty string, got '%s'", decrypted)
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	InitEncryptionKey("test-key-for-invalid-ciphertext-32!")

	_, err := Decrypt("not-valid-base64!!!")
	if err != nil {
		t.Logf("correctly rejected invalid ciphertext: %v", err)
	}
}

func TestIsEncryptionEnabled(t *testing.T) {
	InitEncryptionKey("enable-test-key-32-bytes-long!!!")
	if !IsEncryptionEnabled() {
		t.Error("encryption should be enabled after key init")
	}
}

func TestEncryptField(t *testing.T) {
	InitEncryptionKey("field-test-key-32-bytes-long!!!!!!")
	plaintext := "sensitive-data"
	encrypted := EncryptField(plaintext)
	if encrypted != plaintext {
		decrypted := DecryptField(encrypted)
		if decrypted != plaintext {
			t.Errorf("roundtrip failed: expected '%s', got '%s'", plaintext, decrypted)
		}
	}
}

func TestEncryptDecryptUnicode(t *testing.T) {
	InitEncryptionKey("unicode-test-key-32-bytes-long!!!")
	plaintext := "中文测试数据 🔐"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt unicode failed: %v", err)
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt unicode failed: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestEncryptDecryptLongText(t *testing.T) {
	InitEncryptionKey("long-text-key-32-bytes-long!!!!!!!")
	plaintext := ""
	for i := 0; i < 1000; i++ {
		plaintext += "a"
	}
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt long text failed: %v", err)
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt long text failed: %v", err)
	}
	if decrypted != plaintext {
		t.Error("long text roundtrip failed")
	}
}
