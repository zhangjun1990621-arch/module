/**
 * 平台配置类型定义
 * ------------------------------------------------
 * 前端是通用"壳"，所有平台配置由后端 GET /api/platforms 返回。
 * 前端通过这些类型定义来解析配置，动态注册路由和渲染页面。
 */

/** 导航菜单项 */
export interface NavItem {
  /** 路由路径，如 /aluminum/overview（一级菜单有 path，二级菜单也直接用完整路径） */
  path?: string
  /** 菜单显示名称 */
  label: string
  /** Element Plus 图标组件名，如 Odometer、Grid、Bell */
  icon?: string
  /** 子菜单（二级菜单），有 children 时 path 可省略（作为分组标题不跳转） */
  children?: NavItem[]
}

/** 表格列定义 */
export interface ColumnDef {
  /** 数据字段名 */
  field: string
  /** 列标题 */
  label: string
  /** 列宽度 */
  width?: number
  /** 对齐方式 */
  align?: 'left' | 'center' | 'right'
  /** 列渲染类型 */
  type?: 'text' | 'tag' | 'dot' | 'temperature' | 'number'
  /** 单位（如 ℃、kW、V） */
  unit?: string
  /** tag/dot 类型的选项映射，key 为原始值，value 为显示文本 */
  options?: Record<string, string>
}

/** 页面配置 */
export interface PageConfig {
  /** 页面类型：表格 / 仪表盘 / 图表 / 自定义组件 */
  type: 'table' | 'dashboard' | 'chart' | 'custom'
  /** 页面标题 */
  title: string
  /** table 类型的列定义 */
  columns?: ColumnDef[]
  /** 数据接口路径（相对于 /api/:platform/），如 devices、alarms */
  api?: string
  /** chart 类型的图表类型：line / bar / pie */
  chartType?: string
  /** 仪表盘/图表的布局配置 */
  layout?: any
  /** custom 类型：自定义组件名称（对应组件注册表中的 key） */
  component?: string
}

/** 平台配置（内嵌在 Platform 的 config 字段中） */
export interface PlatformConfig {
  /** 导航菜单项列表 */
  navItems: NavItem[]
  /** 页面配置映射，key 为页面路径标识（如 overview、devices） */
  pages: Record<string, PageConfig>
}

/** 平台定义 —— 后端返回的完整平台对象 */
export interface Platform {
  /** 平台唯一标识，如 aluminum、pv */
  id: string
  /** 平台名称，如 铝厂云控平台 */
  name: string
  /** 平台图标（Emoji 或 Element Plus 图标名） */
  icon: string
  /** 平台数据库名/Schema */
  schema: string
  /** 平台配置（含菜单和页面定义） */
  config: PlatformConfig
  /** 平台状态：active / disabled */
  status: string
  /** 排序序号 */
  sortOrder: number
}

/** API 统一响应结构 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

/** 分页响应结构 */
export interface PageResult<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

/** 用户信息 */
export interface UserInfo {
  id: number
  username: string
  role: string
  platforms: string
  status: string
}
