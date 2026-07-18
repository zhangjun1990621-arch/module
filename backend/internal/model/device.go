package model

import (
	"time"

	"gorm.io/datatypes"
)

// 设备状态常量
const (
	DeviceStatusOnline  = "online"
	DeviceStatusOffline = "offline"
	DeviceStatusAlarm   = "alarm"
)

// Device 通用设备模型。
// 该模型同时用于 public schema 的统一设备表，以及各平台专属 schema 内的 devices 表。
// 当通过 GetPlatformDB 切换 search_path 后，操作将命中平台专属表。
type Device struct {
	ID         string         `json:"id" gorm:"primaryKey"`
	PlatformID string         `json:"platformId" gorm:"column:platform_id;index"`
	DeviceID   string         `json:"deviceId" gorm:"column:device_id;index"`
	Name       string         `json:"name" gorm:"size:128"`
	StationID  string         `json:"stationId" gorm:"column:station_id;index;size:64"`
	Status     string         `json:"status" gorm:"size:16;default:'offline'"`
	LastSeen   *time.Time     `json:"lastSeen" gorm:"column:last_seen"`
	Metadata   datatypes.JSON `json:"metadata" gorm:"type:jsonb"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}
