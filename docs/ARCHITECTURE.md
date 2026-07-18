# 架构设计详解

## 1. 整体架构

### 1.1 设计哲学

本平台采用**配置驱动 + 动态注册**的架构理念，核心目标是：

> 新增一个监控平台 = 数据库插入一条记录 + 创建一个 Schema，零代码改动、零停机。

这与传统的"每个平台一套独立系统"或"所有平台共用一套表（通过 platform_id 区分）"方案有本质区别。

### 1.2 三种多租户方案对比

| 方案 | 数据隔离 | 实现复杂度 | 扩展性 | 本平台选择 |
|------|---------|-----------|--------|-----------|
| 独立数据库 | 最强 | 高（每平台一套 DB） | 差 | ❌ |
| 共享数据库 + 共享表 | 弱 | 低 | 中 | ❌ |
| 共享数据库 + 独立 Schema | 强 | 中 | 好 | ✅ |

**选择独立 Schema 的原因**：
1. PostgreSQL 原生支持 Schema 级别隔离，无需额外中间件
2. 同一数据库实例内管理，运维成本低
3. `SET LOCAL search_path` 事务级切换，连接池安全
4. Schema 创建/删除简单（`CREATE SCHEMA` / `DROP SCHEMA CASCADE`），无需跨库操作

---

## 2. 后端架构

### 2.1 分层结构

```
┌─────────────────────────────────┐
│         HTTP Router (Gin)        │  路由注册 + CORS
├─────────────────────────────────┤
│       Middleware Layer           │  JWT Auth → Role Check → Platform Access
├─────────────────────────────────┤
│        Handler Layer             │  请求解析 → 调用 Service/DB → 响应封装
├─────────────────────────────────┤
│        Service Layer             │  业务逻辑（仅 Platform 有复杂逻辑）
├─────────────────────────────────┤
│       Database Layer             │  GORM + 多 Schema 路由
├─────────────────────────────────┤
│      PostgreSQL (多 Schema)      │  public + schema_aluminum + schema_pv + ...
└─────────────────────────────────┘
```

### 2.2 多 Schema 路由核心

`database/db.go` 中的 `GetPlatformDB()` 是整个架构的核心：

```go
func GetPlatformDB(platformID string) (*gorm.DB, error) {
    // 1. 从 public.platforms 查询平台记录
    var platform model.Platform
    DB.Where("id = ? AND status = 'active'", platformID).First(&platform)
    
    // 2. 校验 schema 名（防注入）
    schemaName := platform.Schema
    if !isValidSchemaName(schemaName) {
        return nil, errors.New("invalid schema name")
    }
    
    // 3. 开启事务
    tx := DB.Begin()
    
    // 4. 事务内设置 search_path（SET LOCAL 仅在事务内生效）
    tx.Exec("SET LOCAL search_path TO " + schemaName + ", public")
    
    // 5. 返回事务 DB，后续查询自动命中平台专属表
    return tx, nil
}
```

**关键点**：
- `SET LOCAL` 而非 `SET`：仅当前事务内生效，事务结束后自动还原，连接池中的连接复用时不会残留
- `PreferSimpleProtocol: true`：禁用 pgx 预处理语句缓存，避免 `cached plan must not change result type` 错误
- Schema 名正则校验 `^[a-z_][a-z0-9_]*$`：防止 SQL 注入

### 2.3 通用 Handler 模式

设备/告警/仪表盘的 Handler 都是"空结构体"（无状态），通过 `getPlatformTx(c)` 获取已切换 Schema 的事务 DB：

```go
func (h *DeviceHandler) List(c *gin.Context) {
    pdb, ok := getPlatformTx(c)  // 获取已设 search_path 的事务
    if !ok { return }
    defer pdb.Rollback()          // 确保事务结束
    
    var devices []model.Device
    pdb.Offset(offset).Limit(pageSize).Find(&devices)
    // 查询自动命中 schema_xxx.devices 表
    
    pdb.Commit()                  // 提交事务
    pagedSuccess(c, devices, total, page, pageSize)
}
```

一套 Handler 代码服务所有平台，平台 ID 从 URL 路径参数 `:platform` 获取。

### 2.4 权限模型

三级角色 + 平台访问列表：

```
super_admin  → 可访问所有平台 + 平台管理 + 权限管理
admin        → 可访问 platforms 字段中指定的平台
viewer       → 可访问 platforms 字段中指定的平台（只读）
```

中间件链：
```
AuthMiddleware(jwtSecret)     → 解析 JWT，注入 userID/username/role/platforms
  ↓
RequireRole("super_admin")    → 仅超管可访问（用户管理路由组）
  ↓
PlatformAccess()              → 校验 :platform 参数是否在用户 platforms 列表中
```

### 2.5 配置加载

使用 Viper 从 `config/config.yaml` 加载，支持环境变量覆盖：
- `CONFIG_PATH`：指定配置文件路径
- 环境变量覆盖规则：`database.host` → `DATABASE_HOST`

---

## 3. 前端架构

### 3.1 动态路由注册

```typescript
// router/index.ts
function setupDynamicRoutes(platforms: Platform[]) {
  platforms.forEach(platform => {
    registerNavRoutes(platform.id, platform.config.navItems, platform.config.pages)
  })
}

function registerNavRoutes(platformId, navItems, pages) {
  navItems.forEach(item => {
    if (item.path) {
      // 一级菜单：注册路由
      const pageKey = item.path.split('/').pop()  // '/aluminum/overview' → 'overview'
      router.addRoute('Layout', {
        path: item.path,
        component: DynamicPage,
        props: { platformId, pagePath: pageKey }
      })
    }
    if (item.children) {
      // 二级菜单：递归注册
      registerNavRoutes(platformId, item.children, pages)
    }
  })
}
```

### 3.2 DynamicPage 分发器

```vue
<template>
  <component :is="renderComponent" :platform="platformId" />
</template>

<script setup>
const pageConfig = platform.config.pages[pagePath]

const renderComponent = computed(() => {
  switch (pageConfig.type) {
    case 'table':     return DynamicTable
    case 'dashboard': return DynamicDashboard
    case 'chart':     return DynamicChart
    case 'custom':    return getCustomComponent(pageConfig.component)
  }
})
</script>
```

### 3.3 自定义组件注册表

```typescript
// customComponentRegistry.ts
const registry = {
  'AluminumOverview':   () => import('@/views/aluminum/OverviewView.vue'),
  'AluminumCellMonitor':() => import('@/views/aluminum/CellMonitorView.vue'),
  'PvDashboard':        () => import('@/views/dashboard/DashboardView.vue'),
  // ...
}
```

异步加载（`defineAsyncComponent`），仅访问该页面时才下载组件代码，减小首屏体积。

### 3.4 状态管理（Pinia）

| Store | 职责 |
|-------|------|
| auth | Token 持久化、用户信息、登录/登出、角色判断 |
| platform | 平台列表加载、当前平台切换、visiblePlatforms 过滤 |
| device | 设备树、选中设备、实时数据（光伏专用） |
| mqtt | MQTT 命令状态追踪（光伏专用） |

### 3.5 双数据源策略

| 数据源 | 用途 | 说明 |
|--------|------|------|
| `request.ts` (`/api`) | 主数据源 | 走真实后端，用于认证/平台/设备/告警/仪表盘 |
| `pvRequest.ts` (mock) | 光伏专用 | 自定义 adapter 全部走 mock，后续可无缝切换真实 API |

光伏组件使用 `pvRequest.ts`，通过 `pvMock.ts` 返回演示数据。后续接入真实后端只需移除 adapter 并添加 `/pv/` 前缀。

---

## 4. 数据流

### 4.1 用户登录流程

```
用户输入用户名密码
  ↓
POST /api/auth/login
  ↓
后端 bcrypt 校验密码
  ↓
签发 JWT（含 userID/username/role/platforms）
  ↓
前端存储 Token 到 localStorage
  ↓
GET /api/platforms（带 Token）
  ↓
前端动态注册路由
  ↓
跳转 /dashboard
```

### 4.2 平台切换流程

```
用户点击侧边栏平台菜单
  ↓
Vue Router 导航到 /aluminum/overview
  ↓
DynamicPage.vue 加载，props: platformId=aluminum, pagePath=overview
  ↓
从 platformStore 查找 aluminum 的 config.pages.overview
  ↓
pageConfig.type = 'custom', component = 'AluminumOverview'
  ↓
customComponentRegistry 异步加载 OverviewView.vue
  ↓
组件渲染，调 GET /api/aluminum/devices 获取数据
```

### 4.3 新增平台流程

```
超管进入「平台管理」页面
  ↓
填写平台 ID/名称/图标/菜单/页面配置
  ↓
POST /api/platforms（含 JSONB config）
  ↓
后端 service.CreatePlatform():
  1. 校验 ID 唯一
  2. CREATE SCHEMA schema_xxx
  3. AutoMigrate(Device, Alarm) 在新 Schema 内建表
  4. INSERT INTO platforms 记录
  ↓
前端刷新平台列表
  ↓
侧边栏自动出现新平台菜单
  ↓
点击进入，DynamicPage 按配置渲染页面
```

---

## 5. 安全设计

### 5.1 认证与授权
- JWT HS256 签名，24 小时过期
- 密码 bcrypt 哈希（cost=10）
- Token 通过 `Authorization: Bearer` 传输
- 401 自动跳转登录页

### 5.2 SQL 注入防护
- Schema 名正则校验 `^[a-z_][a-z0-9_]*$`
- GORM 参数化查询
- `SET LOCAL search_path` 使用拼接但已校验合法性

### 5.3 业务安全
- 不能操作自己的账号（编辑/删除）
- 不能降级/禁用/删除最后一个超级管理员
- 非超管用户必须分配可访问平台
- 用户名唯一约束

### 5.4 CORS
- 开发环境允许所有 Origin（生产建议收紧）
- 支持 Credentials
- OPTIONS 预检自动响应

---

## 6. 扩展性设计

### 6.1 新增通用页面类型

当前支持 4 种页面类型（table/dashboard/chart/custom），如需新增（如 map 地图）：

1. 前端创建 `DynamicMap.vue` 组件
2. `DynamicPage.vue` 的 `switch` 添加 `case 'map'`
3. `types/platform.ts` 的 `PageConfig.type` 联合类型添加 `'map'`
4. 平台配置中使用 `"type": "map"`

### 6.2 新增自定义组件

1. 在 `views/` 下创建组件文件
2. 在 `customComponentRegistry.ts` 注册
3. 平台配置的 `component` 字段填写注册名

### 6.3 新增平台

零代码方式（通过 UI）：
1. 超管进入「平台管理」
2. 填写平台信息 + 菜单 + 页面配置
3. 保存即可

SQL 方式：
```sql
INSERT INTO platforms (id, name, icon, schema_name, config, status, sort_order)
VALUES ('gas', '燃气监控', '🔥', 'schema_gas', '{...}', 'active', 6);
-- Schema 由后端自动创建
```
