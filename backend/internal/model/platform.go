package model

import (
	"encoding/json"
	"time"
)

// Platform 平台注册表模型，存储于 public schema。
// 每接入一个新平台，仅需在此表新增一条记录并创建对应 schema，后端自动路由。
type Platform struct {
	ID        string          `json:"id" gorm:"primaryKey"`                      // 平台唯一标识："aluminum", "pv", "water"
	Name      string          `json:"name"`                                      // 平台中文名："铝厂云控平台"
	Icon      string          `json:"icon"`                                      // 平台图标 emoji："🏭"
	Schema    string          `json:"schema" gorm:"column:schema_name"`          // 平台专属 schema 名："schema_aluminum"
	Config    json.RawMessage `json:"config" gorm:"type:jsonb"`                  // 平台 UI 配置(菜单/页面/字段定义)
	Status    string          `json:"status" gorm:"default:'active'"`            // active/inactive
	SortOrder int             `json:"sortOrder" gorm:"column:sort_order;default:0"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// TableName 指定表名
func (Platform) TableName() string {
	return "platforms"
}

// ---- 平台 UI 配置结构(用于构建/解析 Config JSON) ----
// 注意：这些结构必须与前端 types/platform.ts 中的定义保持一致。

// PlatformConfig 平台完整 UI 描述，包含导航菜单与页面定义
type PlatformConfig struct {
	NavItems []NavItem            `json:"navItems"`
	Pages    map[string]PageDef   `json:"pages"`
}

// NavItem 侧边导航菜单项
type NavItem struct {
	Path     string     `json:"path,omitempty"`     // 路由路径，如 /aluminum/devices（分组标题可省略）
	Label    string     `json:"label"`              // 显示名："设备列表"
	Icon     string     `json:"icon,omitempty"`     // 图标标识："Monitor"
	Children []NavItem  `json:"children,omitempty"` // 子菜单（二级菜单）
}

// PageDef 页面定义，前端据此动态渲染
type PageDef struct {
	Type      string      `json:"type"`              // 页面类型：table / chart / dashboard / custom
	Title     string      `json:"title"`             // 页面标题
	API       string      `json:"api"`               // 数据接口资源名(相对 /api/:platform/)，如 devices、alarms
	Columns   []ColumnDef `json:"columns,omitempty"` // 表格列定义(table 类型时使用)
	Component string      `json:"component,omitempty"` // 自定义组件名(custom 类型时使用)，如 AluminumCellMonitor
}

// ColumnDef 表格列定义
type ColumnDef struct {
	Field   string            `json:"field"`             // 数据字段名
	Label   string            `json:"label"`             // 列标题
	Type    string            `json:"type"`              // 渲染类型：text/tag/dot/temperature/number
	Width   int               `json:"width,omitempty"`   // 列宽(像素)
	Unit    string            `json:"unit,omitempty"`    // 单位(如 ℃、kW)
	Options map[string]string `json:"options,omitempty"` // tag/dot 类型的选项映射
}
