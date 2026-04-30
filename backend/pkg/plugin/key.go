package plugin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	allowedTablePattern  = regexp.MustCompile(`(?i)\b(?:plugin_configs|plugin_data|plugin_events)\b`)
	dangerousSQLPattern  = regexp.MustCompile(`(?i)\b(?:DROP|ALTER|TRUNCATE|GRANT|REVOKE|CREATE\s+USER|ATTACH|DETACH)\b`)
	dangerousTablePattern = regexp.MustCompile(`(?i)\b(?:users|roles|menus|apis|tenants|tenant_users|user_roles|role_menus|role_apis|tenant_roles|payment_configs|payment_orders|sso_auth_codes|sso_tokens|installed_plugins|plugin_version_histories|token_blacklist_entries)\b`)
	blockCommentPattern  = regexp.MustCompile(`(?s)/\*.*?\*/`)
	lineCommentPattern   = regexp.MustCompile(`--[^\n]*`)
	multiSpacePattern    = regexp.MustCompile(`\s+`)
)

func stripSQLComments(query string) string {
	cleaned := blockCommentPattern.ReplaceAllString(query, " ")
	cleaned = lineCommentPattern.ReplaceAllString(cleaned, " ")
	cleaned = multiSpacePattern.ReplaceAllString(cleaned, " ")
	return cleaned
}

func ValidatePluginSQL(query string) error {
	trimmed := strings.TrimSpace(query)
	if len(trimmed) == 0 {
		return fmt.Errorf("SQL语句不能为空")
	}

	cleaned := stripSQLComments(trimmed)

	if dangerousSQLPattern.MatchString(cleaned) {
		return fmt.Errorf("禁止执行危险SQL操作")
	}

	if dangerousTablePattern.MatchString(cleaned) {
		return fmt.Errorf("禁止访问系统表")
	}

	return nil
}

func pluginKey(tenantID uint, pluginID string) string {
	return fmt.Sprintf("t%d_p%s", tenantID, pluginID)
}

func splitPluginKey(key string) []string {
	parts := strings.SplitN(key, "_p", 2)
	if len(parts) == 2 && len(parts[0]) > 1 {
		tenantStr := parts[0][1:]
		if _, err := strconv.ParseUint(tenantStr, 10, 32); err == nil {
			return parts
		}
	}
	return nil
}
