package sanitizer

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	sqlInjectionPattern = regexp.MustCompile(`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|ALTER|CREATE|TRUNCATE|EXEC|EXECUTE|DECLARE|CAST|CONVERT)\b|--|;|/\*|\*/|xp_|sp_)`)
	xssPattern         = regexp.MustCompile(`(?i)<\s*script|<\s*iframe|<\s*object|<\s*embed|<\s*form|javascript:|vbscript:|on(load|click|error|mouseover|focus|blur|submit|change|keydown|keyup|keypress)=`)
	pathTraversalPattern = regexp.MustCompile(`\.\./|\.\.\\`)
	controlCharPattern   = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
)

func SanitizeString(input string) string {
	s := input
	s = controlCharPattern.ReplaceAllString(s, "")
	s = html.EscapeString(s)
	return s
}

func SanitizeHTML(input string) string {
	s := input
	s = xssPattern.ReplaceAllString(s, "")
	s = controlCharPattern.ReplaceAllString(s, "")
	return s
}

func HasSQLInjection(input string) bool {
	cleaned := strings.ToLower(strings.TrimSpace(input))
	return sqlInjectionPattern.MatchString(cleaned)
}

func HasXSS(input string) bool {
	return xssPattern.MatchString(input)
}

func HasPathTraversal(input string) bool {
	return pathTraversalPattern.MatchString(input)
}

func ValidateLength(input string, min, max int) bool {
	length := utf8.RuneCountInString(input)
	return length >= min && length <= max
}

func ValidateAlphanumeric(input string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, input)
	return matched
}

func ValidateEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

func SanitizeMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		switch val := v.(type) {
		case string:
			result[k] = SanitizeString(val)
		case map[string]interface{}:
			result[k] = SanitizeMap(val)
		default:
			result[k] = v
		}
	}
	return result
}

func TruncateString(input string, maxLen int) string {
	if utf8.RuneCountInString(input) <= maxLen {
		return input
	}
	runes := []rune(input)
	return string(runes[:maxLen])
}
