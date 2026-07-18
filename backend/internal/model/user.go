package model

import (
	"time"
)

// 用户角色常量
const (
	RoleSuperAdmin = "super_admin" // 超级管理员，可管理所有平台与用户
	RoleAdmin      = "admin"      // 平台管理员
	RoleViewer     = "viewer"      // 只读用户
)

// 用户状态常量
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

// User 用户模型，存储于 public schema
type User struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	Username  string     `json:"username" gorm:"uniqueIndex;size:64"`
	Password  string     `json:"-" gorm:"column:password"` // bcrypt 哈希，不输出到 JSON
	Role      string     `json:"role" gorm:"size:32;default:'viewer'"`
	Platforms string     `json:"platforms" gorm:"size:256"` // 逗号分隔的可访问平台 ID："aluminum,pv"
	Status    string     `json:"status" gorm:"size:16;default:'active'"`
	LastLogin *time.Time `json:"lastLogin" gorm:"column:last_login"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// HasPlatform 判断用户是否有权访问指定平台
func (u *User) HasPlatform(platformID string) bool {
	if u.Role == RoleSuperAdmin {
		return true
	}
	if u.Platforms == "" {
		return false
	}
	for _, p := range splitCSV(u.Platforms) {
		if p == platformID {
			return true
		}
	}
	return false
}

// splitCSV 将逗号分隔字符串切分为切片
func splitCSV(s string) []string {
	var result []string
	current := ""
	for _, r := range s {
		if r == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
