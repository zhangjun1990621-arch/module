package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"iot-platform/internal/database"
	"iot-platform/internal/model"
)

// PlatformService 平台服务层，封装平台的创建/查询/更新/删除逻辑
type PlatformService struct {
	DB *gorm.DB
}

// NewPlatformService 创建平台服务实例
func NewPlatformService(db *gorm.DB) *PlatformService {
	return &PlatformService{DB: db}
}

// CreatePlatformInput 创建平台入参
type CreatePlatformInput struct {
	ID        string             `json:"id" binding:"required"`
	Name      string             `json:"name" binding:"required"`
	Icon      string             `json:"icon"`
	Schema    string             `json:"schema" binding:"required"`
	Config    *model.PlatformConfig `json:"config"`
	Status    string             `json:"status"`
	SortOrder int                `json:"sortOrder"`
}

// CreatePlatform 创建平台记录 + 创建专属 schema + 在 schema 内建表。
// 实现"加一条记录即接入新平台"的零代码改动理念。
func (s *PlatformService) CreatePlatform(in CreatePlatformInput) (*model.Platform, error) {
	// 校验平台 ID 唯一
	var count int64
	if err := s.DB.Model(&model.Platform{}).Where("id = ?", in.ID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("查询平台失败: %w", err)
	}
	if count > 0 {
		return nil, errors.New("平台 ID 已存在")
	}

	// 处理状态默认值
	status := in.Status
	if status == "" {
		status = "active"
	}

	// 处理 config JSON
	configJSON, err := marshalConfig(in.Config)
	if err != nil {
		return nil, fmt.Errorf("解析平台配置失败: %w", err)
	}

	platform := &model.Platform{
		ID:        in.ID,
		Name:      in.Name,
		Icon:      in.Icon,
		Schema:    in.Schema,
		Config:    configJSON,
		Status:    status,
		SortOrder: in.SortOrder,
	}

	// 先创建 schema 及其下表
	if err := database.CreatePlatformSchema(in.Schema); err != nil {
		return nil, fmt.Errorf("创建平台 schema 失败: %w", err)
	}

	// 写入平台记录
	if err := s.DB.Create(platform).Error; err != nil {
		// 回滚已创建的 schema
		_ = database.DropPlatformSchema(in.Schema)
		return nil, fmt.Errorf("创建平台记录失败: %w", err)
	}

	return platform, nil
}

// GetActivePlatforms 查询所有 active 平台(按 sortOrder 排序)
func (s *PlatformService) GetActivePlatforms() ([]model.Platform, error) {
	var platforms []model.Platform
	err := s.DB.Where("status = ?", "active").
		Order("sort_order ASC, created_at ASC").
		Find(&platforms).Error
	return platforms, err
}

// GetAllPlatforms 查询所有平台(含 inactive)
func (s *PlatformService) GetAllPlatforms() ([]model.Platform, error) {
	var platforms []model.Platform
	err := s.DB.Order("sort_order ASC, created_at ASC").Find(&platforms).Error
	return platforms, err
}

// GetPlatformByID 查询单个平台
func (s *PlatformService) GetPlatformByID(id string) (*model.Platform, error) {
	var platform model.Platform
	if err := s.DB.Where("id = ?", id).First(&platform).Error; err != nil {
		return nil, err
	}
	return &platform, nil
}

// UpdatePlatformInput 更新平台入参
type UpdatePlatformInput struct {
	Name      *string             `json:"name"`
	Icon      *string             `json:"icon"`
	Config    *model.PlatformConfig `json:"config"`
	Status    *string             `json:"status"`
	SortOrder *int                `json:"sortOrder"`
}

// UpdatePlatform 更新平台信息(不含 schema 名)
func (s *PlatformService) UpdatePlatform(id string, in UpdatePlatformInput) (*model.Platform, error) {
	platform, err := s.GetPlatformByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Icon != nil {
		updates["icon"] = *in.Icon
	}
	if in.Config != nil {
		configJSON, err := marshalConfig(in.Config)
		if err != nil {
			return nil, err
		}
		updates["config"] = configJSON
	}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if in.SortOrder != nil {
		updates["sort_order"] = *in.SortOrder
	}

	if len(updates) > 0 {
		if err := s.DB.Model(platform).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.GetPlatformByID(id)
}

// DeletePlatform 删除平台记录并删除其专属 schema
func (s *PlatformService) DeletePlatform(id string) error {
	platform, err := s.GetPlatformByID(id)
	if err != nil {
		return err
	}

	// 删除平台记录
	if err := s.DB.Delete(platform).Error; err != nil {
		return fmt.Errorf("删除平台记录失败: %w", err)
	}

	// 删除专属 schema(含其下所有表)
	if err := database.DropPlatformSchema(platform.Schema); err != nil {
		return fmt.Errorf("删除平台 schema 失败: %w", err)
	}

	return nil
}

// Count 返回平台总数
func (s *PlatformService) Count() (int64, error) {
	var count int64
	err := s.DB.Model(&model.Platform{}).Count(&count).Error
	return count, err
}

// marshalConfig 将 PlatformConfig 转为 json.RawMessage
func marshalConfig(cfg *model.PlatformConfig) (json.RawMessage, error) {
	if cfg == nil {
		return json.RawMessage("{}"), nil
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}
