package service

import (
	"fayhub/internal/model"
	"testing"
)

func TestAuditActionConstants(t *testing.T) {
	actions := []model.AuditAction{
		model.AuditActionLogin,
		model.AuditActionLogout,
		model.AuditActionCreate,
		model.AuditActionUpdate,
		model.AuditActionDelete,
		model.AuditActionEnable,
		model.AuditActionDisable,
		model.AuditActionUpgrade,
		model.AuditActionRollback,
		model.AuditActionInstall,
		model.AuditActionUninstall,
		model.AuditActionExport,
		model.AuditActionImport,
		model.AuditActionConfigChange,
		model.AuditActionPermission,
	}

	expected := []string{
		"login", "logout", "create", "update", "delete",
		"enable", "disable", "upgrade", "rollback",
		"install", "uninstall", "export", "import",
		"config_change", "permission_change",
	}

	for i, action := range actions {
		if string(action) != expected[i] {
			t.Errorf("action[%d]: expected '%s', got '%s'", i, expected[i], string(action))
		}
	}
}

func TestAuditLogModel(t *testing.T) {
	log := &model.AuditLog{
		UserID:     1,
		Username:   "admin",
		Action:     "login",
		Resource:   "auth",
		ResourceID: "",
		IP:         "127.0.0.1",
		UserAgent:  "test-agent",
		StatusCode: 200,
		Success:    true,
		Method:     "POST",
		Path:       "/api/auth/login",
		Duration:   50,
	}

	if log.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", log.UserID)
	}
	if log.Action != "login" {
		t.Errorf("expected Action 'login', got '%s'", log.Action)
	}
	if !log.Success {
		t.Error("expected Success true")
	}
	if log.Duration != 50 {
		t.Errorf("expected Duration 50, got %d", log.Duration)
	}
}
