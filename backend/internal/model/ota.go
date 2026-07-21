package model

import (
	"time"

	"gorm.io/datatypes"
)

// OTA 任务状态常量
const (
	OTATaskPending   = "pending"
	OTATaskRunning   = "running"
	OTATaskPaused    = "paused"
	OTATaskCompleted = "completed"
	OTATaskCancelled = "cancelled"
)

// OTA 任务结束原因
const (
	OTAEndAllSuccess   = "all_success"
	OTAEndPartialFail  = "partial_fail"
	OTAEndAllFailed    = "all_failed"
	OTAEndManualStop   = "manual_stop"
	OTAEndCancelled    = "cancelled"
)

// Firmware 固件模型
type Firmware struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string    `json:"name" gorm:"size:255;not null"`
	Version    string    `json:"version" gorm:"size:64"`
	FilePath   string    `json:"filePath" gorm:"column:file_path;size:512"`
	FileSize   int64     `json:"fileSize" gorm:"column:file_size;default:0"`
	MD5        string    `json:"md5" gorm:"size:64"`
	DeviceType string    `json:"deviceType" gorm:"column:device_type;size:64"`
	UploadTime time.Time `json:"uploadTime" gorm:"column:upload_time;autoCreateTime"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Firmware) TableName() string {
	return "firmwares"
}

// OTATask 升级任务模型
type OTATask struct {
	ID            uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	FirmwareID    uint           `json:"firmwareId" gorm:"column:firmware_id;not null"`
	Firmware      *Firmware      `json:"firmware,omitempty" gorm:"foreignKey:FirmwareID;references:ID"`
	Status        string         `json:"status" gorm:"size:16;default:'pending'"`
	TotalDevices  int            `json:"totalDevices" gorm:"column:total_devices;default:0"`
	SuccessCount  int            `json:"successCount" gorm:"column:success_count;default:0"`
	FailCount     int            `json:"failCount" gorm:"column:fail_count;default:0"`
	Progress      int            `json:"progress" gorm:"default:0"`
	CreatedBy     string         `json:"createdBy" gorm:"column:created_by;size:64"`
	CreatedAt     time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	CompletedAt   *time.Time     `json:"completedAt" gorm:"column:completed_at"`
	EndReason     string         `json:"endReason" gorm:"column:end_reason;size:32"`
	FailedDevices datatypes.JSON `json:"failedDevices,omitempty" gorm:"column:failed_devices;type:jsonb"`
	TaskDevices   []OTATaskDevice `json:"taskDevices,omitempty" gorm:"foreignKey:TaskID;references:ID"`
}

func (OTATask) TableName() string {
	return "ota_tasks"
}

// OTATaskDevice 任务设备明细
type OTATaskDevice struct {
	ID        uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    uint       `json:"taskId" gorm:"column:task_id;index;not null"`
	DeviceID  string     `json:"deviceId" gorm:"column:device_id;size:64;not null"`
	Status    string     `json:"status" gorm:"size:16;default:'pending'"` // pending / upgrading / success / failed
	ErrorMsg  string     `json:"errorMsg" gorm:"column:error_msg;type:text"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (OTATaskDevice) TableName() string {
	return "ota_task_devices"
}

// OTA 设备明细状态
const (
	OTADevicePending   = "pending"
	OTADeviceUpgrading = "upgrading"
	OTADeviceSuccess   = "success"
	OTADeviceFailed    = "failed"
)
