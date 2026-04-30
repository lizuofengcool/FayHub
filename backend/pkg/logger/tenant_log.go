package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap/zapcore"

	"fayhub/pkg/redisclient"
)

const (
	logStreamPrefix = "fayhub:logs:tenant:"
	logStreamMaxLen = 10000
)

type TenantLogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	TenantID   uint   `json:"tenant_id"`
	UserID     uint   `json:"user_id,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Message    string `json:"message"`
	Path       string `json:"path,omitempty"`
	Method     string `json:"method,omitempty"`
	IP         string `json:"ip,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Duration   string `json:"duration,omitempty"`
}

func getTenantLogStreamKey(tenantID uint) string {
	return fmt.Sprintf("%s%d", logStreamPrefix, tenantID)
}

func StoreTenantLog(entry *TenantLogEntry) error {
	if entry.TenantID == 0 {
		return nil
	}

	rdb := redisclient.GetRawClient()
	if rdb == nil {
		return nil
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("序列化日志条目失败: %w", err)
	}

	ctx := context.Background()
	key := getTenantLogStreamKey(entry.TenantID)

	return rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{
			"data": string(data),
		},
		MaxLen: int64(logStreamMaxLen),
		Approx: true,
	}).Err()
}

func QueryTenantLogs(tenantID uint, level string, limit int64, startTime, endTime time.Time) ([]*TenantLogEntry, error) {
	rdb := redisclient.GetRawClient()
	if rdb == nil {
		return nil, fmt.Errorf("Redis未连接，无法查询租户日志")
	}

	ctx := context.Background()
	key := getTenantLogStreamKey(tenantID)

	start := "-"
	end := "+"

	if !startTime.IsZero() {
		start = fmt.Sprintf("%d-0", startTime.UnixMilli())
	}
	if !endTime.IsZero() {
		end = fmt.Sprintf("%d-0", endTime.UnixMilli())
	}

	messages, err := rdb.XRange(ctx, key, start, end).Result()
	if err != nil {
		return nil, fmt.Errorf("查询租户日志失败: %w", err)
	}

	if limit > 0 && int64(len(messages)) > limit {
		messages = messages[:limit]
	}

	var entries []*TenantLogEntry
	for _, msg := range messages {
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var entry TenantLogEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}

		if level != "" && entry.Level != level {
			continue
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func GetTenantLogCount(tenantID uint) (int64, error) {
	rdb := redisclient.GetRawClient()
	if rdb == nil {
		return 0, fmt.Errorf("Redis未连接")
	}

	ctx := context.Background()
	key := getTenantLogStreamKey(tenantID)

	return rdb.XLen(ctx, key).Result()
}

type tenantLogCore struct {
	core zapcore.Core
}

func newTenantLogCore(core zapcore.Core) zapcore.Core {
	return &tenantLogCore{core: core}
}

func (c *tenantLogCore) Enabled(level zapcore.Level) bool {
	return c.core.Enabled(level)
}

func (c *tenantLogCore) With(fields []zapcore.Field) zapcore.Core {
	return newTenantLogCore(c.core.With(fields))
}

func (c *tenantLogCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func (c *tenantLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	var tenantID uint
	var userID uint
	var requestID string
	var path string
	var method string
	var ip string

	for _, f := range fields {
		switch f.Key {
		case "tenant_id":
			tenantID = uint(f.Integer)
		case "user_id":
			userID = uint(f.Integer)
		case "request_id":
			requestID = f.String
		case "path":
			path = f.String
		case "method":
			method = f.String
		case "ip":
			ip = f.String
		}
	}

	if tenantID > 0 {
		logEntry := &TenantLogEntry{
			Timestamp: entry.Time.Format(time.RFC3339),
			Level:     entry.Level.String(),
			TenantID:  tenantID,
			UserID:    userID,
			RequestID: requestID,
			Message:   entry.Message,
			Path:      path,
			Method:    method,
			IP:        ip,
		}
		_ = StoreTenantLog(logEntry)
	}

	return c.core.Write(entry, fields)
}

func (c *tenantLogCore) Sync() error {
	return c.core.Sync()
}
