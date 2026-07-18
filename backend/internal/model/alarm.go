package model

import (
	"time"
)

// 告警等级常量
const (
	AlarmLevelInfo     = "info"
	AlarmLevelWarning  = "warning"
	AlarmLevelCritical = "critical"
)

// 告警状态常量
const (
	AlarmStatusActive   = "active"
	AlarmStatusResolved = "resolved"
)

// Alarm 通用告警模型。
// 该模型同时用于 public schema 的统一告警表，以及各平台专属 schema 内的 alarms 表。
type Alarm struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	PlatformID string     `json:"platformId" gorm:"column:platform_id;index"`
	DeviceID   string     `json:"deviceId" gorm:"column:device_id;index;size:64"`
	DeviceName string     `json:"deviceName" gorm:"column:device_name;size:128"`
	Level      string     `json:"level" gorm:"size:16;default:'info'"`
	Type       string     `json:"type" gorm:"size:64"`
	Detail      string     `json:"detail" gorm:"type:text"`
	Status     string     `json:"status" gorm:"size:16;default:'active'"`
	OccurredAt time.Time  `json:"occurredAt" gorm:"column:occurred_at;index"`
	ResolvedAt *time.Time `json:"resolvedAt" gorm:"column:resolved_at"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// TableName 指定表名
func (Alarm) TableName() string {
	return "alarms"
}
