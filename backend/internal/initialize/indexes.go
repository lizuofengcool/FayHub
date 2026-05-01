package initialize

import (
	"context"
	"fayhub/pkg/utils"
	"fmt"
)

func MigrateCompositeIndexes() error {
	db := utils.GetDB(context.TODO())
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	indexes := []struct {
		Table string
		Name  string
		SQL   string
	}{
		{
			Table: "audit_logs",
			Name:  "idx_audit_tenant_action",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_audit_tenant_action ON audit_logs(tenant_id, action)",
		},
		{
			Table: "audit_logs",
			Name:  "idx_audit_tenant_created",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_audit_tenant_created ON audit_logs(tenant_id, created_at DESC)",
		},
		{
			Table: "audit_logs",
			Name:  "idx_audit_user_created",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_audit_user_created ON audit_logs(user_id, created_at DESC)",
		},
		{
			Table: "notifications",
			Name:  "idx_notify_user_read",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_notify_user_read ON notifications(user_id, is_read, id DESC)",
		},
		{
			Table: "notifications",
			Name:  "idx_notify_user_type",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_notify_user_type ON notifications(user_id, type, id DESC)",
		},
		{
			Table: "webhook_deliveries",
			Name:  "idx_webhook_del_sub_status",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_webhook_del_sub_status ON webhook_deliveries(subscription_id, status)",
		},
		{
			Table: "webhook_deliveries",
			Name:  "idx_webhook_del_created",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_webhook_del_created ON webhook_deliveries(created_at DESC)",
		},
		{
			Table: "payments",
			Name:  "idx_payment_tenant_status",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_payment_tenant_status ON payments(tenant_id, status)",
		},
		{
			Table: "payments",
			Name:  "idx_payment_user_status",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_payment_user_status ON payments(user_id, status)",
		},
		{
			Table: "plugin_version_histories",
			Name:  "idx_plugin_ver_tenant_plugin",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_plugin_ver_tenant_plugin ON plugin_version_histories(tenant_id, plugin_id, version)",
		},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.SQL).Error; err != nil {
			return fmt.Errorf("创建索引 %s 失败: %w", idx.Name, err)
		}
	}

	return nil
}
