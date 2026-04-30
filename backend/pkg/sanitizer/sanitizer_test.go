package sanitizer

import (
	"testing"
)

func TestHasSQLInjection(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"normal text", false},
		{"SELECT * FROM users", true},
		{"DROP TABLE users", true},
		{"1; DROP TABLE users", true},
		{"' OR 1=1 --", true},
		{"hello world", false},
		{"UNION SELECT", true},
		{"insert into", true},
	}

	for _, tt := range tests {
		result := HasSQLInjection(tt.input)
		if result != tt.expected {
			t.Errorf("HasSQLInjection(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestHasXSS(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"normal text", false},
		{"<script>alert(1)</script>", true},
		{"<iframe src='evil'>", true},
		{"javascript:alert(1)", true},
		{"onclick=alert(1)", true},
		{"<b>bold</b>", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		result := HasXSS(tt.input)
		if result != tt.expected {
			t.Errorf("HasXSS(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestHasPathTraversal(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/api/users", false},
		{"../../../etc/passwd", true},
		{"..\\windows\\system32", true},
		{"/normal/path", false},
	}

	for _, tt := range tests {
		result := HasPathTraversal(tt.input)
		if result != tt.expected {
			t.Errorf("HasPathTraversal(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeString(t *testing.T) {
	input := "<script>alert('xss')</script>"
	result := SanitizeString(input)
	if containsScript(result) {
		t.Errorf("SanitizeString(%q) = %q, still contains script tag", input, result)
	}
}

func TestValidateLength(t *testing.T) {
	tests := []struct {
		input    string
		min      int
		max      int
		expected bool
	}{
		{"hello", 1, 10, true},
		{"", 1, 10, false},
		{"a very long string", 1, 5, false},
		{"ok", 2, 2, true},
	}

	for _, tt := range tests {
		result := ValidateLength(tt.input, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("ValidateLength(%q, %d, %d) = %v, expected %v", tt.input, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestValidateAlphanumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello123", true},
		{"hello_world", true},
		{"hello-world", true},
		{"hello world", false},
		{"hello@world", false},
	}

	for _, tt := range tests {
		result := ValidateAlphanumeric(tt.input)
		if result != tt.expected {
			t.Errorf("ValidateAlphanumeric(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"user@example.com", true},
		{"invalid", false},
		{"@example.com", false},
		{"user@", false},
	}

	for _, tt := range tests {
		result := ValidateEmail(tt.input)
		if result != tt.expected {
			t.Errorf("ValidateEmail(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeMap(t *testing.T) {
	data := map[string]interface{}{
		"name":   "<script>alert(1)</script>",
		"nested": map[string]interface{}{"key": "<b>bold</b>"},
		"number": 42,
	}

	result := SanitizeMap(data)
	nameStr, _ := result["name"].(string)
	if containsScript(nameStr) {
		t.Errorf("SanitizeMap did not sanitize name field: %s", nameStr)
	}
}

func TestTruncateString(t *testing.T) {
	result := TruncateString("hello world", 5)
	if result != "hello" {
		t.Errorf("TruncateString = %q, expected %q", result, "hello")
	}

	result = TruncateString("hi", 5)
	if result != "hi" {
		t.Errorf("TruncateString = %q, expected %q", result, "hi")
	}
}

func containsScript(s string) bool {
	for i := 0; i <= len(s)-7; i++ {
		if s[i:i+7] == "<script" {
			return true
		}
	}
	return false
}
