package utils

import (
	"strings"
	"testing"
)

func TestValidatePassword_Valid(t *testing.T) {
	passwords := []string{
		"Test1234!",
		"Password1@",
		"MyP@ss12#",
		"Abcdefg1!",
	}

	for _, pw := range passwords {
		if err := ValidatePassword(pw); err != nil {
			t.Errorf("期望密码'%s'验证通过，但失败: %v", pw, err)
		}
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	err := ValidatePassword("Ab1")
	if err == nil {
		t.Error("期望短密码验证失败")
	}
	if !strings.Contains(err.Error(), "长度不能少于") {
		t.Errorf("期望长度错误提示，实际: %v", err)
	}
}

func TestValidatePassword_TooLong(t *testing.T) {
	longPassword := strings.Repeat("A", 129) + "a1"
	err := ValidatePassword(longPassword)
	if err == nil {
		t.Error("期望超长密码验证失败")
	}
	if !strings.Contains(err.Error(), "长度不能超过") {
		t.Errorf("期望长度错误提示，实际: %v", err)
	}
}

func TestValidatePassword_NoUppercase(t *testing.T) {
	err := ValidatePassword("abcdefg1")
	if err == nil {
		t.Error("期望无大写字母密码验证失败")
	}
	if !strings.Contains(err.Error(), "大写字母") {
		t.Errorf("期望大写字母错误提示，实际: %v", err)
	}
}

func TestValidatePassword_NoLowercase(t *testing.T) {
	err := ValidatePassword("ABCDEFG1")
	if err == nil {
		t.Error("期望无小写字母密码验证失败")
	}
	if !strings.Contains(err.Error(), "小写字母") {
		t.Errorf("期望小写字母错误提示，实际: %v", err)
	}
}

func TestValidatePassword_NoDigit(t *testing.T) {
	err := ValidatePassword("Abcdefgh")
	if err == nil {
		t.Error("期望无数字密码验证失败")
	}
	if !strings.Contains(err.Error(), "数字") {
		t.Errorf("期望数字错误提示，实际: %v", err)
	}
}

func TestValidatePassword_MultipleErrors(t *testing.T) {
	err := ValidatePassword("abc")
	if err == nil {
		t.Error("期望弱密码验证失败")
	}

	validationErr, ok := err.(*PasswordValidationError)
	if !ok {
		t.Errorf("期望PasswordValidationError类型，实际: %T", err)
	}
	if len(validationErr.Errors) < 3 {
		t.Errorf("期望至少3个错误，实际: %d", len(validationErr.Errors))
	}
}

func TestValidatePasswordWithPolicy_NoSpecialRequired(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:      6,
		MaxLength:      128,
		RequireUpper:   false,
		RequireLower:   false,
		RequireDigit:   false,
		RequireSpecial: false,
	}

	if err := ValidatePasswordWithPolicy("abc123", policy); err != nil {
		t.Errorf("期望简单密码在宽松策略下通过: %v", err)
	}
}

func TestValidatePasswordWithPolicy_RequireSpecial(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:      8,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}

	if err := ValidatePasswordWithPolicy("Test1234", policy); err == nil {
		t.Error("期望无特殊字符密码在严格策略下失败")
	}

	if err := ValidatePasswordWithPolicy("Test123!", policy); err != nil {
		t.Errorf("期望含特殊字符密码在严格策略下通过: %v", err)
	}
}

func TestValidatePassword_Empty(t *testing.T) {
	err := ValidatePassword("")
	if err == nil {
		t.Error("期望空密码验证失败")
	}
}

func TestValidatePassword_BoundaryLength(t *testing.T) {
	exactly8 := "Abcd12!@"
	if err := ValidatePassword(exactly8); err != nil {
		t.Errorf("期望8位密码验证通过: %v", err)
	}

	exactly7 := "Abcd12!"
	if err := ValidatePassword(exactly7); err == nil {
		t.Error("期望7位密码验证失败")
	}
}
