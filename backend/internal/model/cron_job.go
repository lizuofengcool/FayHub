package model

import (
	"time"
)

type CronJob struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	TenantID    int64      `gorm:"index;default:0" json:"tenant_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	CronExpr    string    `gorm:"size:50;not null" json:"cron_expr"`
	Command     string    `gorm:"size:500;not null" json:"command"`
	Description string    `gorm:"size:500" json:"description"`
	Status      int       `gorm:"default:1" json:"status"`
	LastRunAt   *time.Time `json:"last_run_at"`
	NextRunAt   *time.Time `json:"next_run_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (CronJob) TableName() string {
	return "cron_jobs"
}

type CronJobLog struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	JobID     int64      `gorm:"index;not null" json:"job_id"`
	TenantID  int64      `gorm:"index;default:0" json:"tenant_id"`
	Status    string    `gorm:"size:20;not null" json:"status"`
	Output    string    `gorm:"type:text" json:"output"`
	Error     string    `gorm:"type:text" json:"error"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (CronJobLog) TableName() string {
	return "cron_job_logs"
}
