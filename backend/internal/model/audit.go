package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONRawMessage json.RawMessage

func (j *JSONRawMessage) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.(string)
	if !ok {
		b, ok := value.([]byte)
		if !ok {
			return errors.New("failed to scan JSONRawMessage")
		}
		*j = JSONRawMessage(b)
		return nil
	}
	*j = JSONRawMessage(s)
	return nil
}

func (j JSONRawMessage) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j JSONRawMessage) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONRawMessage) UnmarshalJSON(data []byte) error {
	*j = JSONRawMessage(data)
	return nil
}

type AuditLog struct {
	SnowflakeTenantModel
	UserID      int64           `json:"user_id" gorm:"index"`
	Username    string          `json:"username" gorm:"size:100"`
	Action      string          `json:"action" gorm:"size:50;index;not null"`
	Resource    string          `json:"resource" gorm:"size:100;index"`
	ResourceID  string          `json:"resource_id" gorm="size:100;index"`
	Detail      JSONRawMessage  `json:"detail" gorm:"type:text"`
	IP          string          `json:"ip" gorm="size:45"`
	UserAgent   string          `json:"user_agent" gorm="size:500"`
	RequestID   string          `json:"request_id" gorm="size:50;index"`
	StatusCode  int             `json:"status_code"`
	Success     bool            `json:"success" gorm="default:true"`
	ErrorMsg    string          `json:"error_msg,omitempty" gorm="size:500"`
	Duration    int64           `json:"duration"`
	Method      string          `json:"method" gorm="size:10"`
	Path        string          `json:"path" gorm="size:500;index"`
	CreatedAt   *time.Time      `json:"created_at" gorm="index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type AuditAction string

const (
	AuditActionLogin       AuditAction = "login"
	AuditActionLogout      AuditAction = "logout"
	AuditActionCreate      AuditAction = "create"
	AuditActionUpdate      AuditAction = "update"
	AuditActionDelete      AuditAction = "delete"
	AuditActionEnable      AuditAction = "enable"
	AuditActionDisable     AuditAction = "disable"
	AuditActionUpgrade     AuditAction = "upgrade"
	AuditActionRollback    AuditAction = "rollback"
	AuditActionInstall     AuditAction = "install"
	AuditActionUninstall   AuditAction = "uninstall"
	AuditActionExport      AuditAction = "export"
	AuditActionImport      AuditAction = "import"
	AuditActionConfigChange AuditAction = "config_change"
	AuditActionPermission  AuditAction = "permission_change"
)
