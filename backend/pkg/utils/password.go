package utils

import (
	"fmt"
	"unicode"
)

type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: true,
}

type PasswordValidationError struct {
	Errors []string
}

func (e *PasswordValidationError) Error() string {
	return fmt.Sprintf("密码强度不足: %v", e.Errors)
}

func ValidatePassword(password string) error {
	return ValidatePasswordWithPolicy(password, DefaultPasswordPolicy)
}

func ValidatePasswordWithPolicy(password string, policy PasswordPolicy) error {
	var errs []string

	if len(password) < policy.MinLength {
		errs = append(errs, fmt.Sprintf("长度不能少于%d个字符", policy.MinLength))
	}
	if len(password) > policy.MaxLength {
		errs = append(errs, fmt.Sprintf("长度不能超过%d个字符", policy.MaxLength))
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if policy.RequireUpper && !hasUpper {
		errs = append(errs, "必须包含至少一个大写字母")
	}
	if policy.RequireLower && !hasLower {
		errs = append(errs, "必须包含至少一个小写字母")
	}
	if policy.RequireDigit && !hasDigit {
		errs = append(errs, "必须包含至少一个数字")
	}
	if policy.RequireSpecial && !hasSpecial {
		errs = append(errs, "必须包含至少一个特殊字符")
	}

	if len(errs) > 0 {
		return &PasswordValidationError{Errors: errs}
	}
	return nil
}
