package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"iot-platform/internal/model"
)

// AlarmHandler 通用告警处理器，根据 :platform 参数路由到对应 schema 的 alarms 表
type AlarmHandler struct{}

// NewAlarmHandler 创建告警处理器
func NewAlarmHandler() *AlarmHandler {
	return &AlarmHandler{}
}

// List 告警列表(支持分页、等级筛选、状态筛选、时间范围)
func (h *AlarmHandler) List(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	page, pageSize, offset := pagination(c)

	query := pdb.Model(&model.Alarm{})

	// 等级筛选
	if level := c.Query("level"); level != "" {
		query = query.Where("level = ?", level)
	}
	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	// 设备筛选
	if deviceID := c.Query("deviceId"); deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	// 时间范围筛选
	if start := c.Query("startTime"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			query = query.Where("occurred_at >= ?", t)
		}
	}
	if end := c.Query("endTime"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			query = query.Where("occurred_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询告警总数失败: "+err.Error())
		return
	}

	var alarms []model.Alarm
	if err := query.Offset(offset).Limit(pageSize).Order("occurred_at DESC").Find(&alarms).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询告警列表失败: "+err.Error())
		return
	}
	if alarms == nil {
		alarms = make([]model.Alarm, 0)
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败")
		return
	}
	committed = true
	pagedSuccess(c, alarms, total, page, pageSize)
}

// Resolve 处理告警(标记为已解决)
func (h *AlarmHandler) Resolve(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var alarm model.Alarm
	if err := pdb.Where("id = ?", id).First(&alarm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "告警不存在")
			return
		}
		fail(c, http.StatusInternalServerError, "查询告警失败: "+err.Error())
		return
	}

	if alarm.Status == model.AlarmStatusResolved {
		if err := pdb.Commit().Error; err != nil {
			fail(c, http.StatusInternalServerError, "提交事务失败")
			return
		}
		committed = true
		success(c, alarm)
		return
	}

	now := time.Now()
	if err := pdb.Model(&alarm).Updates(map[string]interface{}{
		"status":      model.AlarmStatusResolved,
		"resolved_at": now,
	}).Error; err != nil {
		fail(c, http.StatusInternalServerError, "处理告警失败: "+err.Error())
		return
	}

	alarm.Status = model.AlarmStatusResolved
	alarm.ResolvedAt = &now

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败")
		return
	}
	committed = true
	success(c, alarm)
}
