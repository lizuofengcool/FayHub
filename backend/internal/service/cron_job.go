package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/utils"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type CronJobService struct {
	cron    *cron.Cron
	entries map[int64]cron.EntryID
	mu      sync.RWMutex
}

func NewCronJobService() *CronJobService {
	svc := &CronJobService{
		cron:    cron.New(cron.WithSeconds(), cron.WithLocation(time.Local)),
		entries: make(map[int64]cron.EntryID),
	}
	svc.cron.Start()
	return svc
}

func (s *CronJobService) Create(ctx context.Context, job *model.CronJob) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	if err := db.Create(job).Error; err != nil {
		return err
	}

	if job.Status == 1 {
		s.addJob(job)
	}

	return nil
}

func (s *CronJobService) Update(ctx context.Context, job *model.CronJob) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	s.removeJob(job.ID)

	if err := db.Save(job).Error; err != nil {
		return err
	}

	if job.Status == 1 {
		s.addJob(job)
	}

	return nil
}

func (s *CronJobService) Delete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	s.removeJob(id)

	return db.Delete(&model.CronJob{}, id).Error
}

func (s *CronJobService) GetByID(ctx context.Context, id int64) (*model.CronJob, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	var job model.CronJob
	if err := db.First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *CronJobService) List(ctx context.Context, page, pageSize int) ([]model.CronJob, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	var total int64
	var jobs []model.CronJob

	query := db.Model(&model.CronJob{})
	query.Count(&total)

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

func (s *CronJobService) ToggleStatus(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	var job model.CronJob
	if err := db.First(&job, id).Error; err != nil {
		return err
	}

	s.removeJob(id)

	if job.Status == 1 {
		job.Status = 0
	} else {
		job.Status = 1
	}

	if err := db.Save(&job).Error; err != nil {
		return err
	}

	if job.Status == 1 {
		s.addJob(&job)
	}

	return nil
}

func (s *CronJobService) ExecuteOnce(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	var job model.CronJob
	if err := db.First(&job, id).Error; err != nil {
		return err
	}

	go s.runJob(&job)

	return nil
}

func (s *CronJobService) addJob(job *model.CronJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[job.ID]; exists {
		return
	}

	entryID, err := s.cron.AddFunc(job.CronExpr, func() {
		s.runJob(job)
	})
	if err != nil {
		fmt.Printf("添加定时任务失败 [%s]: %v\n", job.Name, err)
		return
	}

	s.entries[job.ID] = entryID
}

func (s *CronJobService) removeJob(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entries[id]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, id)
	}
}

func (s *CronJobService) runJob(job *model.CronJob) {
	ctx := utils.SkipTenantIsolation(context.Background())
	db := utils.GetDB(ctx)

	now := time.Now()
	log := model.CronJobLog{
		JobID:     job.ID,
		TenantID:  job.TenantID,
		Status:    "running",
		StartedAt: now,
	}

	if db != nil {
		db.Create(&log)
		db.Model(&model.CronJob{}).Where("id = ?", job.ID).Update("last_run_at", now)
	}

	var output string
	var jobErr error

	switch job.Command {
	case "backup":
		backupSvc := &BackupService{}
		record, err := backupSvc.CreateBackup(ctx)
		if err != nil {
			jobErr = err
			output = fmt.Sprintf("备份失败: %v", err)
		} else {
			output = fmt.Sprintf("备份成功: %s (大小: %d bytes)", record.Filename, record.FileSize)
		}
	case "cleanup_old_backups":
		count, err := s.cleanupOldBackups(ctx)
		if err != nil {
			jobErr = err
			output = fmt.Sprintf("清理旧备份失败: %v", err)
		} else {
			output = fmt.Sprintf("清理完成，删除了 %d 个旧备份", count)
		}
	default:
		output = fmt.Sprintf("定时任务 [%s] 执行成功 (命令: %s)", job.Name, job.Command)
	}

	if db != nil {
		endedAt := time.Now()
		status := "success"
		if jobErr != nil {
			status = "failed"
		}
		db.Model(&model.CronJobLog{}).Where("id = ?", log.ID).Updates(map[string]interface{}{
			"status":   status,
			"output":   output,
			"ended_at": endedAt,
		})
	}
}

func (s *CronJobService) cleanupOldBackups(ctx context.Context) (int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	cutoff := time.Now().AddDate(0, 0, -30)
	result := db.Where("created_at < ?", cutoff).Delete(&model.BackupRecord{})
	if result.Error != nil {
		return 0, result.Error
	}

	backupDir := filepath.Join("data", "backups")
	var oldRecords []model.BackupRecord
	db.Unscoped().Where("created_at < ?", cutoff).Find(&oldRecords)
	for _, r := range oldRecords {
		os.Remove(filepath.Join(backupDir, r.Filename))
	}

	return result.RowsAffected, nil
}

func (s *CronJobService) GetLogs(ctx context.Context, jobID int64, page, pageSize int) ([]model.CronJobLog, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, fmt.Errorf("数据库未连接")
	}

	var total int64
	var logs []model.CronJobLog

	query := db.Model(&model.CronJobLog{}).Where("job_id = ?", jobID)
	query.Count(&total)

	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (s *CronJobService) LoadJobsFromDB(ctx context.Context) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return fmt.Errorf("数据库未连接")
	}

	var jobs []model.CronJob
	if err := db.Where("status = ?", 1).Find(&jobs).Error; err != nil {
		return err
	}

	for i := range jobs {
		s.addJob(&jobs[i])
	}

	return nil
}

func (s *CronJobService) Stop() {
	s.cron.Stop()
}

func (s *CronJobService) GetNextRunTime(job *model.CronJob) *time.Time {
	schedule, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor).Parse(job.CronExpr)
	if err != nil {
		return nil
	}
	next := schedule.Next(time.Now())
	return &next
}

func (s *CronJobService) GetEntries() []cron.Entry {
	return s.cron.Entries()
}

func (s *CronJobService) GetEntryID(jobID int64) (cron.EntryID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.entries[jobID]
	return id, ok
}

func (s *CronJobService) GetDB() *gorm.DB {
	return utils.GetDB(context.Background())
}
