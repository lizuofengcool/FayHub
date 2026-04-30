// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package plugin

import (
	"encoding/json"
	"fmt"
	"time"
)

type SandboxConfig struct {
	MemoryLimitPages uint32        `json:"memory_limit_pages"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
	MaxHTTPRequests  int           `json:"max_http_requests"`
	MaxDBQueries     int           `json:"max_db_queries"`
	AllowNetwork     bool          `json:"allow_network"`
	AllowFileAccess  bool          `json:"allow_file_access"`
}

func DefaultSandboxConfig() *SandboxConfig {
	return &SandboxConfig{
		MemoryLimitPages: 256,
		ExecutionTimeout: 30 * time.Second,
		MaxHTTPRequests:  100,
		MaxDBQueries:     500,
		AllowNetwork:     false,
		AllowFileAccess:  false,
	}
}

func ParseSandboxConfig(jsonStr string) (*SandboxConfig, error) {
	if jsonStr == "" {
		return DefaultSandboxConfig(), nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return DefaultSandboxConfig(), nil
	}

	cfg := DefaultSandboxConfig()

	if v, ok := raw["memory_limit_pages"]; ok {
		if f, ok := v.(float64); ok && f > 0 {
			cfg.MemoryLimitPages = uint32(f)
		}
	}

	if v, ok := raw["execution_timeout_seconds"]; ok {
		if f, ok := v.(float64); ok && f > 0 {
			cfg.ExecutionTimeout = time.Duration(f) * time.Second
		}
	}

	if v, ok := raw["max_http_requests"]; ok {
		if f, ok := v.(float64); ok && f > 0 {
			cfg.MaxHTTPRequests = int(f)
		}
	}

	if v, ok := raw["max_db_queries"]; ok {
		if f, ok := v.(float64); ok && f > 0 {
			cfg.MaxDBQueries = int(f)
		}
	}

	if v, ok := raw["allow_network"]; ok {
		if b, ok := v.(bool); ok {
			cfg.AllowNetwork = b
		}
	}

	if v, ok := raw["allow_file_access"]; ok {
		if b, ok := v.(bool); ok {
			cfg.AllowFileAccess = b
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *SandboxConfig) Validate() error {
	if c.MemoryLimitPages == 0 {
		return fmt.Errorf("内存限制页数不能为0")
	}
	if c.MemoryLimitPages > 512 {
		return fmt.Errorf("内存限制页数不能超过512(32MB)")
	}
	if c.ExecutionTimeout < time.Second {
		return fmt.Errorf("执行超时不能小于1秒")
	}
	if c.ExecutionTimeout > 5*time.Minute {
		return fmt.Errorf("执行超时不能超过5分钟")
	}
	return nil
}

func (c *SandboxConfig) MemoryLimitBytes() uint32 {
	return c.MemoryLimitPages * 65536
}
