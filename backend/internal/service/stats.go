package service

import (
	"context"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/metrics"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type StatsService struct{}

type DashboardStats struct {
	TenantCount        int64               `json:"tenant_count"`
	TenantGrowth       float64             `json:"tenant_growth"`
	UserCount          int64               `json:"user_count"`
	UserGrowth         float64             `json:"user_growth"`
	PluginCount        int64               `json:"plugin_count"`
	ActivePluginCount  int64               `json:"active_plugin_count"`
	TotalRequests      int64               `json:"total_requests"`
	ErrorRequests      int64               `json:"error_requests"`
	UptimeSeconds      float64             `json:"uptime_seconds"`
	MemoryAllocMb      float64             `json:"memory_alloc_mb"`
	GoroutineCount     int                 `json:"goroutine_count"`
	PaymentToday       float64             `json:"payment_today"`
	PaymentMonth       float64             `json:"payment_month"`
	OrderToday         int64               `json:"order_today"`
	OrderMonth         int64               `json:"order_month"`
	RequestTrend       []TrendPoint        `json:"request_trend"`
	TenantDistribution []TenantPackageItem `json:"tenant_distribution"`
	RecentActivities   []ActivityItem      `json:"recent_activities"`
}

type TrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TenantPackageItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type ActivityItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Time  string `json:"time"`
	Icon  string `json:"icon"`
}

func (s *StatsService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	// 跳过租户隔离，统计全局数据
	ctx = utils.SkipTenantIsolation(ctx)
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, nil
	}

	stats := &DashboardStats{}

	s.fillSystemMetrics(stats)
	s.fillTenantStats(ctx, db, stats)
	s.fillUserStats(ctx, db, stats)
	s.fillPluginStats(ctx, db, stats)
	s.fillPaymentStats(ctx, db, stats)
	s.fillRequestTrend(ctx, db, stats)
	s.fillTenantDistribution(ctx, db, stats)
	s.fillRecentActivities(ctx, db, stats)

	return stats, nil
}

func (s *StatsService) fillSystemMetrics(stats *DashboardStats) {
	m := metrics.GetMetrics()
	if v, ok := m["uptime_seconds"].(float64); ok {
		stats.UptimeSeconds = v
	}
	if v, ok := m["total_requests"].(int64); ok {
		stats.TotalRequests = v
	}
	if v, ok := m["error_requests"].(int64); ok {
		stats.ErrorRequests = v
	}
	if v, ok := m["goroutine_count"].(int); ok {
		stats.GoroutineCount = v
	}
	if v, ok := m["memory_alloc_mb"].(float64); ok {
		stats.MemoryAllocMb = v
	}
}

func (s *StatsService) fillTenantStats(ctx context.Context, db *gorm.DB, stats *DashboardStats) {
	db.Model(&model.Tenant{}).Count(&stats.TenantCount)

	monthStart := time.Now().AddDate(0, 0, -30)
	var lastMonthCount int64
	db.Model(&model.Tenant{}).Where("created_at < ?", monthStart).Count(&lastMonthCount)

	if lastMonthCount > 0 {
		stats.TenantGrowth = float64(stats.TenantCount-lastMonthCount) / float64(lastMonthCount) * 100
	}
}

func (s *StatsService) fillUserStats(ctx context.Context, db *gorm.DB, stats *DashboardStats) {
	db.Model(&model.User{}).Count(&stats.UserCount)

	monthStart := time.Now().AddDate(0, 0, -30)
	var lastMonthCount int64
	db.Model(&model.User{}).Where("created_at < ?", monthStart).Count(&lastMonthCount)

	if lastMonthCount > 0 {
		stats.UserGrowth = float64(stats.UserCount-lastMonthCount) / float64(lastMonthCount) * 100
	}
}

func (s *StatsService) fillPluginStats(ctx context.Context, db *gorm.DB, stats *DashboardStats) {
	db.Model(&model.InstalledPlugin{}).Count(&stats.PluginCount)
	db.Model(&model.InstalledPlugin{}).Where("status = ?", "active").Count(&stats.ActivePluginCount)
}

func (s *StatsService) fillPaymentStats(ctx context.Context, db *gorm.DB, stats *DashboardStats) {
	today := time.Now().Truncate(24 * time.Hour)
	monthStart := time.Now().AddDate(0, 0, -30)

	db.Model(&model.PaymentOrder{}).
		Where("status = ? AND paid_at >= ?", model.PaymentStatusPaid, today).
		Count(&stats.OrderToday)

	db.Model(&model.PaymentOrder{}).
		Where("status = ? AND paid_at >= ?", model.PaymentStatusPaid, monthStart).
		Count(&stats.OrderMonth)

	var todayAmount, monthAmount int64
	db.Model(&model.PaymentOrder{}).
		Where("status = ? AND paid_at >= ?", model.PaymentStatusPaid, today).
		Select("COALESCE(SUM(amount), 0)").Scan(&todayAmount)
	db.Model(&model.PaymentOrder{}).
		Where("status = ? AND paid_at >= ?", model.PaymentStatusPaid, monthStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&monthAmount)

	stats.PaymentToday = float64(todayAmount) / 100.0
	stats.PaymentMonth = float64(monthAmount) / 100.0
}

func (s *StatsService) fillRequestTrend(ctx context.Context, db *gorm.DB, stats *DashboardStats) {
	trend := make([]TrendPoint, 0)
	now := time.Now()

	for i := 6; i >= 0; i-- {
		dayStart := now.AddDate(0, 0, -i).Truncate(24 * time.Hour)
		dayEnd := dayStart.Add(24 * time.Hour)

		var count int64
		db.Model(&model.AuditLog{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&count)

		trend = append(trend, TrendPoint{
			Date:  dayStart.Format("01-02"),
			Count: count,
		})
	}

	stats.RequestTrend = trend
}

func (s *StatsService) fillTenantDistribution(ctx context.Context, db *gorm.DB, stats *DashboardStats) {
	type packageCount struct {
		PackageID int64
		Count     int64
	}

	var results []packageCount
	db.Model(&model.Tenant{}).
		Select("package_id, COUNT(*) as count").
		Group("package_id").
		Scan(&results)

	distribution := make([]TenantPackageItem, 0)
	for _, r := range results {
		name := "默认套餐"
		if r.PackageID > 0 {
			var pkg model.TenantPackage
			if err := db.Model(&model.TenantPackage{}).Where("id = ?", r.PackageID).Select("name").Scan(&pkg).Error; err == nil {
				name = pkg.Name
			}
		}
		distribution = append(distribution, TenantPackageItem{
			Name:  name,
			Value: r.Count,
		})
	}

	if len(distribution) == 0 {
		distribution = append(distribution, TenantPackageItem{Name: "暂无数据", Value: 0})
	}

	stats.TenantDistribution = distribution
}

func (s *StatsService) fillRecentActivities(ctx context.Context, db *gorm.DB, stats *DashboardStats) {
	var logs []model.AuditLog
	db.Order("created_at DESC").Limit(8).Find(&logs)

	activities := make([]ActivityItem, 0, len(logs))
	for _, log := range logs {
		activities = append(activities, ActivityItem{
			ID:    log.ID,
			Title: actionToTitle(log.Action, log.Resource, log.Username),
			Time:  log.CreatedAt.Format("2006-01-02 15:04"),
			Icon:  actionToIcon(log.Action),
		})
	}

	stats.RecentActivities = activities
}

func actionToTitle(action, resource, username string) string {
	user := username
	if user == "" {
		user = "系统"
	}

	label := actionLabel(action)
	if resource != "" {
		return user + " " + label + " " + resource
	}
	return user + " " + label
}

func actionLabel(action string) string {
	labels := map[string]string{
		"login":     "登录了系统",
		"logout":    "退出了系统",
		"create":    "创建了",
		"update":    "更新了",
		"delete":    "删除了",
		"enable":    "启用了",
		"disable":   "禁用了",
		"install":   "安装了",
		"uninstall": "卸载了",
		"pay":       "支付了",
		"refund":    "退款了",
	}
	if l, ok := labels[action]; ok {
		return l
	}
	return action
}

func actionToIcon(action string) string {
	icons := map[string]string{
		"login":     "UserFilled",
		"logout":    "UserFilled",
		"create":    "Plus",
		"update":    "Setting",
		"delete":    "Delete",
		"install":   "Download",
		"uninstall": "Delete",
		"pay":       "Money",
		"refund":    "Money",
	}
	if i, ok := icons[action]; ok {
		return i
	}
	return "Setting"
}
