package model

import "time"

type OnlineUser struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TenantID  int64     `json:"tenant_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	LoginAt   time.Time `json:"login_at"`
	LastSeen  time.Time `json:"last_seen"`
}
