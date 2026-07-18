package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"iot-platform/internal/model"
)

// DashboardHandler 通用仪表盘处理器
type DashboardHandler struct{}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// dashboardKPI 仪表盘 KPI 统计
type dashboardKPI struct {
	DeviceTotal int   `json:"deviceTotal"`
	Online      int64 `json:"online"`
	Offline     int64 `json:"offline"`
	Alarm       int64 `json:"alarm"`
	ActiveAlarm int64 `json:"activeAlarm"`
}

// trendPoint 趋势数据点
type trendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// dashboardData 仪表盘聚合数据
type dashboardData struct {
	KPI          dashboardKPI  `json:"kpi"`
	RecentAlarms []model.Alarm `json:"recentAlarms"`
	Trend        []trendPoint  `json:"trend"`
}

// Get 返回平台仪表盘数据：KPI 统计、最近告警、告警趋势
func (h *DashboardHandler) Get(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	// 确保最终回滚（读操作回滚也安全）
	defer pdb.Rollback()

	var kpi dashboardKPI

	// 设备总数
	var deviceTotal int64
	if err := pdb.Model(&model.Device{}).Count(&deviceTotal).Error; err == nil {
		kpi.DeviceTotal = int(deviceTotal)
	}

	// 各状态设备数（逐条检查错误，失败时保持默认值 0）
	pdb.Model(&model.Device{}).Where("status = ?", model.DeviceStatusOnline).Count(&kpi.Online)
	pdb.Model(&model.Device{}).Where("status = ?", model.DeviceStatusOffline).Count(&kpi.Offline)
	pdb.Model(&model.Device{}).Where("status = ?", model.DeviceStatusAlarm).Count(&kpi.Alarm)

	// 活跃告警数
	pdb.Model(&model.Alarm{}).Where("status = ?", model.AlarmStatusActive).Count(&kpi.ActiveAlarm)

	// 最近告警(取 10 条) —— 初始化为空切片避免 JSON 序列化为 null
	recentAlarms := make([]model.Alarm, 0, 10)
	pdb.Where("status = ?", model.AlarmStatusActive).
		Order("occurred_at DESC").
		Limit(10).
		Find(&recentAlarms)

	// 告警趋势：最近 7 天每天的告警数量
	trend := h.buildTrend(pdb)

	// 读操作无需提交，直接返回数据
	success(c, dashboardData{
		KPI:          kpi,
		RecentAlarms: recentAlarms,
		Trend:        trend,
	})
}

// buildTrend 构建最近 7 天的告警趋势数据
func (h *DashboardHandler) buildTrend(pdb *gorm.DB) []trendPoint {
	type dailyCount struct {
		Date  string
		Count int64
	}

	var counts []dailyCount
	sevenDaysAgo := time.Now().AddDate(0, 0, -6).Truncate(24 * time.Hour)

	pdb.Model(&model.Alarm{}).
		Select("TO_CHAR(date_trunc('day', occurred_at), 'YYYY-MM-DD') as date, count(*) as count").
		Where("occurred_at >= ?", sevenDaysAgo).
		Group("date").
		Order("date ASC").
		Scan(&counts)

	// 构建连续 7 天的日期，补全无告警的日期
	countMap := make(map[string]int64)
	for _, dc := range counts {
		countMap[dc.Date] = dc.Count
	}

	trend := make([]trendPoint, 0, 7)
	for i := 0; i < 7; i++ {
		date := sevenDaysAgo.AddDate(0, 0, i).Format("2006-01-02")
		trend = append(trend, trendPoint{
			Date:  date,
			Count: countMap[date],
		})
	}
	return trend
}
