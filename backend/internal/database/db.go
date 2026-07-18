package database

import (
	"errors"
	"fmt"
	"regexp"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"iot-platform/internal/config"
	"iot-platform/internal/model"
)

// DB 公共 schema 数据库实例，用于操作 platforms/users 等公共表
var DB *gorm.DB

// schemaNameReg 用于校验 schema 名合法性，防止 SQL 注入
var schemaNameReg = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// Init 初始化 PostgreSQL 连接并自动迁移公共 schema 的表
func Init(dbCfg config.DatabaseConfig) error {
	// 使用 PreferSimpleProtocol 禁用 pgx 预处理语句缓存，
	// 避免多 schema 切换（SET LOCAL search_path）时出现
	// "cached plan must not change result type" 错误
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbCfg.DSN(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	DB = db

	// 自动迁移公共 schema 表
	if err := AutoMigratePublic(); err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	return nil
}

// AutoMigratePublic 自动迁移 public schema 下的公共表
// 如果 init.sql 已创建表则跳过，仅在表不存在时执行 GORM AutoMigrate
func AutoMigratePublic() error {
	if err := DB.Exec(`CREATE SCHEMA IF NOT EXISTS public`).Error; err != nil {
		return err
	}
	if err := DB.Exec(`SET search_path TO public`).Error; err != nil {
		return err
	}

	// 检查 platforms 表是否已存在（init.sql 已执行过）
	var count int64
	DB.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='platforms'").Scan(&count)
	if count > 0 {
		// 表已存在，跳过 AutoMigrate 避免与 init.sql 的约束冲突
		return nil
	}

	return DB.AutoMigrate(
		&model.Platform{},
		&model.User{},
		&model.Device{},
		&model.Alarm{},
	)
}

// GetPlatformDB 根据 platformID 查询对应 schema 名，
// 返回一个设置了 search_path 的事务实例，用于查询该平台专属表。
//
// 使用事务 + SET LOCAL search_path 实现：
//   - 事务保证该连接在整个请求期间 search_path 不被其他请求干扰
//   - SET LOCAL 使得事务结束后 search_path 自动还原，连接归还连接池时无副作用
//
// 调用方需在请求结束时调用 Commit() 提交事务(读操作提交也安全)，
// 若中途出错可调用 Rollback() 回滚。
func GetPlatformDB(platformID string) (*gorm.DB, error) {
	if DB == nil {
		return nil, errors.New("数据库未初始化")
	}

	// 从 public schema 查询平台记录获取 schema 名
	var platform model.Platform
	if err := DB.Where("id = ?", platformID).First(&platform).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("平台不存在: %s", platformID)
		}
		return nil, fmt.Errorf("查询平台失败: %w", err)
	}

	if platform.Status != "active" {
		return nil, fmt.Errorf("平台已停用: %s", platformID)
	}

	schemaName := platform.Schema
	if !isValidSchemaName(schemaName) {
		return nil, fmt.Errorf("非法 schema 名: %s", schemaName)
	}

	// 开启事务，在事务内设置 search_path
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("开启事务失败: %w", tx.Error)
	}

	// SET LOCAL 仅在当前事务内生效，事务结束后自动还原
	if err := tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %s, public", schemaName)).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("设置 search_path 失败: %w", err)
	}

	return tx, nil
}

// GetSchemaName 根据 platformID 查询 schema 名(不开启事务)
func GetSchemaName(platformID string) (string, error) {
	var platform model.Platform
	if err := DB.Where("id = ?", platformID).First(&platform).Error; err != nil {
		return "", err
	}
	return platform.Schema, nil
}

// CreatePlatformSchema 创建平台专属 schema 并在其中创建 devices / alarms 表
func CreatePlatformSchema(schemaName string) error {
	if !isValidSchemaName(schemaName) {
		return fmt.Errorf("非法 schema 名: %s", schemaName)
	}

	// 创建 schema
	if err := DB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error; err != nil {
		return fmt.Errorf("创建 schema 失败: %w", err)
	}

	// 在新 schema 中创建 devices 和 alarms 表
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %s, public", schemaName)).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.AutoMigrate(&model.Device{}, &model.Alarm{}); err != nil {
		tx.Rollback()
		return fmt.Errorf("平台表迁移失败: %w", err)
	}
	return tx.Commit().Error
}

// DropPlatformSchema 删除平台专属 schema(含其下所有表)
func DropPlatformSchema(schemaName string) error {
	if !isValidSchemaName(schemaName) {
		return fmt.Errorf("非法 schema 名: %s", schemaName)
	}
	return DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error
}

// isValidSchemaName 校验 schema 名仅包含小写字母、数字、下划线，且不以数字开头
func isValidSchemaName(name string) bool {
	if name == "" {
		return false
	}
	return schemaNameReg.MatchString(name)
}
