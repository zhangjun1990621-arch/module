# 综合能源云控平台（IoT Multi-Platform）

> 基于 Go + Vue 3 的动态多平台物联网监控架构，运行时动态注册新平台，前后端零代码改动、零停机接入。

---

## 目录

- [项目简介](#项目简介)
- [核心特性](#核心特性)
- [技术栈](#技术栈)
- [系统架构](#系统架构)
- [目录结构](#目录结构)
- [快速开始](#快速开始)
- [默认账号](#默认账号)
- [平台列表](#平台列表)
- [模块说明](#模块说明)
- [数据库设计](#数据库设计)
- [API 接口](#api-接口)
- [部署指南](#部署指南)
- [运维手册](#运维手册)
- [与原项目隔离方案](#与原项目隔离方案)
- [常见问题](#常见问题)

---

## 项目简介

综合能源云控平台是一个面向多行业（铝厂、光伏、电厂、水务、农场等）的统一物联网监控平台。采用**一平台一 Schema**的多租户架构，通过 PostgreSQL 的 `search_path` 机制实现平台间数据完全隔离。新增平台只需在数据库 `platforms` 表插入一条记录并创建对应 Schema，前端自动出现新平台菜单，无需任何代码改动或重启服务。

### 为什么这样设计

1. **多租户隔离**：每个平台拥有独立的 PostgreSQL Schema（`schema_aluminum`、`schema_pv` 等），设备表和告警表物理隔离，避免跨平台数据污染。
2. **零代码接入**：平台配置（菜单结构、页面类型、列定义）以 JSONB 存储在 `platforms.config` 字段，前端通过动态路由注册机制自动渲染，新增平台无需修改前端代码。
3. **配置驱动 UI**：页面类型（table/dashboard/chart/custom）由后端配置决定，前端 `DynamicPage.vue` 作为统一入口分发渲染，实现"一次开发，多平台复用"。
4. **自定义组件扩展**：对于复杂业务页面（如铝厂电解槽温度矩阵、光伏设备详情），通过 `customComponentRegistry.ts` 注册异步组件，在不破坏通用架构的前提下支持深度定制。

---

## 核心特性

| 特性 | 说明 |
|------|------|
| 动态多平台 | 运行时注册新平台，自动创建 Schema、菜单、路由 |
| 多租户隔离 | 每平台独立 Schema，`SET LOCAL search_path` 事务级切换 |
| 配置驱动 UI | JSONB 存储 NavItems + Pages 配置，前端动态渲染 |
| 三级权限模型 | super_admin / admin / viewer，平台级访问控制 |
| 通用 CRUD | 设备/告警/仪表盘通用 Handler，一套代码服务所有平台 |
| 自定义组件 | 注册表模式支持平台专属深度定制页面 |
| JWT 认证 | 无状态 Token，24 小时过期，Bearer 方式传输 |
| Docker 部署 | 一键 `docker compose up`，含 PostgreSQL + 后端 + Nginx |

---

## 技术栈

### 后端

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.24 | 编程语言 |
| Gin | 1.10.0 | Web 框架 |
| GORM | 1.25.10 | ORM |
| PostgreSQL | 16 | 数据库 |
| pgx | 5.5.5 | PostgreSQL 驱动 |
| Viper | 1.18.2 | 配置管理 |
| golang-jwt | 5.2.1 | JWT 认证 |
| bcrypt | - | 密码哈希 |
| UUID | 1.6.0 | 主键生成 |

### 前端

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.5 | 前端框架 |
| Vite | 8.0 | 构建工具 |
| TypeScript | 6.0 | 类型系统 |
| Pinia | 3.0 | 状态管理 |
| Element Plus | 2.14 | UI 组件库 |
| ECharts | 6.1 | 图表库 |
| Vue Router | 4.6 | 路由 |
| Axios | 1.18 | HTTP 客户端 |

### 部署

| 技术 | 用途 |
|------|------|
| Docker | 容器化 |
| Docker Compose | 多容器编排 |
| Nginx | 前端静态资源 + API 反代 |

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      浏览器 (用户)                            │
│                   http://服务器:8099                         │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP
┌────────────────────────▼────────────────────────────────────┐
│                   Nginx (容器: module-frontend)              │
│  ┌──────────────┐  ┌──────────────────┐                     │
│  │ 静态资源托管  │  │ /api/ → 反代后端  │                     │
│  │ dist/ (SPA)  │  │ proxy_pass       │                     │
│  └──────────────┘  └────────┬─────────┘                     │
└─────────────────────────────┼───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│              Go Backend (容器: module-backend)                │
│              Gin + GORM  (端口: 8090 容器内部)                 │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐   │
│  │ Auth    │ │ Platform │ │ Device   │ │ Alarm/Dashboard│   │
│  │ Handler │ │ Handler  │ │ Handler  │ │ Handler        │   │
│  └─────────┘ └──────────┘ └──────────┘ └────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Middleware: JWT Auth + Role Check + Platform Access  │   │
│  └──────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Database Layer: GetPlatformDB()                      │   │
│  │  BEGIN → SET LOCAL search_path → Query → COMMIT       │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│           PostgreSQL (容器: module-postgres)                  │
│              数据库: iot_platform  (端口: 5432 容器内部)       │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  public schema (公共表)                               │    │
│  │  ├── platforms (平台注册表)                           │    │
│  │  ├── users (用户表)                                   │    │
│  │  ├── devices (通用设备视图)                           │    │
│  │  └── alarms (通用告警视图)                            │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  schema_aluminum (铝厂专属)                          │    │
│  │  ├── devices (铝厂设备)                              │    │
│  │  └── alarms  (铝厂告警)                              │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  schema_pv (光伏专属)                                │    │
│  │  ├── devices / alarms                                │    │
│  ├─────────────────────────────────────────────────────┤    │
│  │  schema_power / schema_water / schema_farm ...       │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 请求流转

1. 浏览器访问 `http://服务器:8099`，Nginx 返回 Vue SPA
2. 用户登录 → `POST /api/auth/login` → Nginx 反代 → Go 后端签发 JWT
3. 前端获取平台列表 → `GET /api/platforms` → 动态注册路由
4. 用户点击平台菜单 → `DynamicPage.vue` 按 `pageConfig.type` 分发渲染
5. 数据请求 → `GET /api/:platform/devices` → 中间件校验 JWT + 平台权限
6. `GetPlatformDB(platformID)` → `BEGIN → SET LOCAL search_path TO schema_xxx, public → 查询 → COMMIT`

---

## 目录结构

```
module/
├── README.md                        # 项目总文档（本文件）
├── start.bat                        # Windows 一键启动（后端+前端）
├── start-backend.bat                # 后端启动脚本（含 PG 连通性检测）
├── start-frontend.bat               # 前端启动脚本
├── .dockerignore
│
├── backend/                         # Go 后端
│   ├── go.mod / go.sum              # 依赖管理
│   ├── Dockerfile                   # 后端多阶段构建
│   ├── config/
│   │   └── config.yaml              # 配置文件（端口/数据库/JWT）
│   ├── cmd/
│   │   └── server/main.go           # 程序入口 + 路由注册
│   ├── migrations/
│   │   ├── init.sql                 # 数据库初始化（建表+示例数据）
│   │   └── migrate_from_pv_platform.sql  # 旧项目数据迁移
│   └── internal/
│       ├── config/config.go         # 配置加载（Viper）
│       ├── database/db.go           # DB 初始化 + 多 Schema 路由
│       ├── model/                   # 数据模型
│       │   ├── platform.go          # 平台模型 + UI 配置结构
│       │   ├── user.go             # 用户模型 + 角色常量
│       │   ├── device.go           # 设备模型
│       │   └── alarm.go            # 告警模型
│       ├── middleware/auth.go       # JWT 鉴权 + 角色/平台权限
│       ├── handler/                 # HTTP 处理器
│       │   ├── common.go           # 统一响应 + 分页工具
│       │   ├── auth.go             # 登录/Profile/登出
│       │   ├── platform.go         # 平台 CRUD
│       │   ├── user.go             # 用户管理（仅超管）
│       │   ├── device.go           # 设备 CRUD
│       │   ├── alarm.go            # 告警查询/处理
│       │   └── dashboard.go        # 仪表盘聚合
│       └── service/platform.go      # 平台业务逻辑
│
├── frontend/                        # Vue 3 前端
│   ├── package.json
│   ├── vite.config.ts               # Vite 配置（代理 /api → :8090）
│   ├── tsconfig.json
│   ├── index.html
│   └── src/
│       ├── main.ts                  # 应用入口
│       ├── App.vue                  # 根组件
│       ├── router/index.ts          # 路由 + 动态注册
│       ├── stores/                  # Pinia 状态管理
│       │   ├── auth.ts             # 认证状态
│       │   ├── platform.ts         # 平台列表状态
│       │   ├── device.ts           # 设备状态
│       │   └── mqtt.ts             # MQTT 命令状态
│       ├── api/                     # API 调用层
│       │   ├── request.ts          # 主 Axios 实例
│       │   ├── auth.ts             # 认证 API
│       │   ├── platform.ts         # 平台 API
│       │   ├── resource.ts         # 通用 RESTful API
│       │   └── ...                  # 光伏专用 API
│       ├── components/              # 通用组件
│       │   ├── MainLayout.vue      # 主框架（侧边栏+顶栏）
│       │   ├── DynamicPage.vue     # 动态页面分发器
│       │   ├── DynamicTable.vue    # 通用表格
│       │   ├── DynamicDashboard.vue# 通用仪表盘
│       │   ├── DynamicChart.vue    # 通用图表
│       │   ├── IconPicker.vue      # 图标选择器
│       │   └── customComponentRegistry.ts  # 自定义组件注册表
│       ├── views/                   # 页面
│       │   ├── login/LoginView.vue
│       │   ├── DashboardView.vue   # 首页平台概览
│       │   ├── PlatformManagement.vue    # 平台管理
│       │   ├── PermissionManagement.vue  # 权限管理
│       │   ├── aluminum/           # 铝厂自定义页面
│       │   ├── dashboard/          # 光伏仪表盘
│       │   ├── device/             # 光伏设备详情
│       │   ├── event/              # 光伏事件列表
│       │   ├── ota/                # 光伏 OTA 升级
│       │   └── station/            # 光伏台区管理
│       ├── types/                   # TypeScript 类型定义
│       │   ├── platform.ts         # 平台配置类型
│       │   └── index.ts            # 业务实体类型
│       ├── composables/             # Vue 组合式函数
│       │   ├── useWebSocket.ts
│       │   └── useCurveChart.ts
│       ├── utils/constants.ts
│       └── styles/main.scss
│
├── deploy/                          # Docker 部署配置
│   ├── docker-compose.yml           # 一键编排
│   ├── Dockerfile.frontend          # 前端构建镜像
│   └── nginx.conf                   # Nginx 配置
│
└── docs/                            # 详细文档
    ├── ARCHITECTURE.md              # 架构设计详解
    ├── DEPLOYMENT.md                # 部署指南
    ├── DATABASE.md                  # 数据库设计
    ├── API.md                       # API 接口文档
    └── OPERATIONS.md                # 运维手册
```

---

## 快速开始

### 方式一：Docker 一键部署（推荐生产环境）

```bash
# 克隆仓库
git clone https://github.com/zhangjun1990621-arch/module.git
cd module/deploy

# 一键启动（首次会自动构建镜像）
docker compose up -d --build

# 查看服务状态
docker compose ps

# 访问
# 前端: http://服务器IP:8099
# 后端: http://服务器IP:8090/api/health
# 默认账号: admin / admin123
```

### 方式二：本地开发

```bash
# 1. 启动 PostgreSQL（确保 5435 端口可用）
# 2. 初始化数据库
psql -h 127.0.0.1 -p 5435 -U postgres -d iot_platform -f backend/migrations/init.sql

# 3. 启动后端
cd backend
go run ./cmd/server          # 监听 :8090

# 4. 启动前端
cd frontend
npm install
npm run dev                  # 监听 :3000, 代理 /api → :8090

# 5. Windows 一键启动
# 双击 start.bat
```

---

## 默认账号

| 用户名 | 密码 | 角色 | 可访问平台 |
|--------|------|------|------------|
| admin | admin123 | super_admin | 全部平台 |

> 首次启动后请立即通过「权限管理」页面修改密码并创建其他用户。

---

## 平台列表

系统预置以下平台（可通过「平台管理」页面增删）：

| ID | 名称 | 图标 | Schema | 页面类型 | 说明 |
|----|------|------|--------|----------|------|
| aluminum | 铝厂云控平台 | 🏭 | schema_aluminum | custom | 电解槽温度矩阵、点位实时、告警历史 |
| pv | 光伏运维平台 | ☀️ | schema_pv | custom | 仪表盘、设备详情、OTA、台区管理 |
| power | 电厂监控平台 | ⚡ | schema_power | table+dashboard | 发电机组、锅炉、环保设备 |
| water | 智慧水务平台 | 💧 | schema_water | table+dashboard | 泵站、阀门、水质监测 |
| farm | 农场监控平台 | 🌾 | schema_farm | dashboard | 农场设备与环境监控 |

---

## 模块说明

### 后端模块

#### 认证模块 (`handler/auth.go`)
- JWT Token 签发（HS256），24 小时过期
- bcrypt 密码校验
- 登录/登出/获取 Profile

#### 平台模块 (`handler/platform.go` + `service/platform.go`)
- 平台 CRUD（创建时自动建 Schema + 建表）
- JSONB 配置存储（NavItems + Pages）
- 创建平台时自动执行 `CREATE SCHEMA` + `AutoMigrate(Device, Alarm)`

#### 用户管理模块 (`handler/user.go`)
- 仅超级管理员可访问
- 用户 CRUD + 重置密码
- 保护逻辑：不能操作自己、不能降级/删除最后一个超管

#### 设备模块 (`handler/device.go`)
- 通用设备 CRUD，通过 `getPlatformTx()` 切换 Schema
- 支持分页、搜索（name/deviceId/stationId）、状态筛选

#### 告警模块 (`handler/alarm.go`)
- 告警列表查询（支持级别/状态/设备/时间范围筛选）
- 告警处理（标记为已恢复）

#### 仪表盘模块 (`handler/dashboard.go`)
- KPI 聚合（设备总数/在线/离线/告警/活跃告警）
- 最近 10 条活跃告警
- 7 天告警趋势（按天聚合，自动补零）

#### 中间件 (`middleware/auth.go`)
- `AuthMiddleware`：JWT 解析与注入
- `RequireRole`：角色校验（super_admin / admin / viewer）
- `PlatformAccess`：平台访问权限校验（超管直通，其他角色检查 platforms CSV）

### 前端模块

#### 动态路由 (`router/index.ts`)
- 静态路由：`/login`、`/dashboard`、`/platform-management`、`/permission-management`
- 动态路由：登录后从 `GET /api/platforms` 获取配置，遍历 `navItems` 调用 `router.addRoute()` 注册
- 路由守卫：Token 校验、超管权限校验、平台加载

#### 动态页面分发 (`DynamicPage.vue`)
- 接收 `platformId` + `pagePath` props
- 从平台配置中查找 `pages[pagePath]`
- 按 `type` 分发：`table` → DynamicTable / `dashboard` → DynamicDashboard / `chart` → DynamicChart / `custom` → 注册表查找

#### 自定义组件注册表 (`customComponentRegistry.ts`)
- key-value 映射，key 为组件名，value 为 `defineAsyncComponent` 工厂函数
- 铝厂：AluminumOverview / AluminumCellMonitor / AluminumPointRealtime / AluminumAlarm / AluminumHistory
- 光伏：PvDashboard / PvDeviceDetail / PvEventList / PvOTA / PvStation

#### 权限管理 (`PermissionManagement.vue`)
- 用户列表（搜索/角色筛选/分页）
- 新增/编辑用户（角色选择 + 平台权限多选）
- 重置密码（二次确认）
- 删除用户（弹窗确认）

---

## 数据库设计

### 公共表（public schema）

| 表名 | 说明 | 主键 |
|------|------|------|
| platforms | 平台注册表 | id (VARCHAR) |
| users | 用户表 | id (VARCHAR) |
| devices | 通用设备视图 | id (VARCHAR) |
| alarms | 通用告警视图 | id (VARCHAR) |

### 平台专属表（各 schema_xxx）

每个平台 Schema 包含相同的两张表：

| 表名 | 说明 |
|------|------|
| devices | 该平台的设备表（结构同 public.devices） |
| alarms | 该平台的告警表（结构同 public.alarms） |

### 关键字段

详见 [docs/DATABASE.md](docs/DATABASE.md)

---

## API 接口

### 认证

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/auth/login | 登录 | 公开 |
| GET | /api/auth/profile | 获取当前用户 | 已认证 |
| POST | /api/auth/logout | 登出 | 已认证 |

### 平台管理

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/platforms | 获取平台列表 | 已认证 |
| GET | /api/platforms/:id | 获取单个平台 | 已认证 |
| POST | /api/platforms | 创建平台 | super_admin |
| PUT | /api/platforms/:id | 更新平台 | 已认证 |
| DELETE | /api/platforms/:id | 删除平台 | super_admin |

### 用户管理

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/users | 用户列表 | super_admin |
| POST | /api/users | 创建用户 | super_admin |
| PUT | /api/users/:id | 更新用户 | super_admin |
| DELETE | /api/users/:id | 删除用户 | super_admin |
| PUT | /api/users/:id/password | 重置密码 | super_admin |

### 平台业务（:platform 为平台 ID）

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/:platform/devices | 设备列表 | 已认证 + 平台权限 |
| GET | /api/:platform/devices/:id | 设备详情 | 已认证 + 平台权限 |
| POST | /api/:platform/devices | 创建设备 | 已认证 + 平台权限 |
| PUT | /api/:platform/devices/:id | 更新设备 | 已认证 + 平台权限 |
| DELETE | /api/:platform/devices/:id | 删除设备 | 已认证 + 平台权限 |
| GET | /api/:platform/alarms | 告警列表 | 已认证 + 平台权限 |
| PUT | /api/:platform/alarms/:id/resolve | 处理告警 | 已认证 + 平台权限 |
| GET | /api/:platform/dashboard | 仪表盘数据 | 已认证 + 平台权限 |

> 详见 [docs/API.md](docs/API.md)

---

## 部署指南

详见 [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

### 快速部署（Docker）

```bash
cd deploy
docker compose up -d --build
# 访问 http://服务器IP:8099
```

### 端口规划

| 服务 | 容器内部端口 | 宿主机映射端口 | 说明 |
|------|------------|--------------|------|
| PostgreSQL | 5432 | 5433 | 数据库 |
| Go 后端 | 8090 | 8090 | API 服务 |
| Nginx 前端 | 80 | 8099 | Web 访问入口 |

---

## 运维手册

详见 [docs/OPERATIONS.md](docs/OPERATIONS.md)

### 常用命令

```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f postgres

# 重启服务
docker compose restart backend

# 更新代码后重新构建
docker compose up -d --build backend frontend

# 进入数据库
docker exec -it module-postgres psql -U postgres -d iot_platform

# 备份数据库
docker exec module-postgres pg_dump -U postgres iot_platform > backup_$(date +%Y%m%d).sql
```

---

## 与原项目隔离方案

本项目与原 `zonghe` 项目（综合能源云控平台旧版）完全隔离，互不影响：

| 维度 | 原项目 (zonghe) | 本项目 (module) |
|------|-----------------|-----------------|
| GitHub 仓库 | zhangjun1990621-arch/zonghe.git | zhangjun1990621-arch/module.git |
| 部署目录 | /data/zonghe | /data/module |
| Web 端口 | 8089 | 8099 |
| 数据库名 | pv_platform | iot_platform |
| PostgreSQL 容器 | zonghe-postgres | module-postgres |
| 后端容器 | zonghe-backend | module-backend |
| 前端容器 | zonghe-frontend | module-frontend |
| Docker 网络 | zonghe_net | module_net |
| 数据卷 | pg_data / redis_data / firmware_data | module_pg_data |
| Nginx 配置 | /etc/nginx/conf.d/zonghe.conf | /etc/nginx/conf.d/module.conf |

---

## 常见问题

### Q: 新增平台后前端菜单没出现？
A: 刷新浏览器页面。前端在登录时加载平台列表，新增平台后需要重新加载。

### Q: 切换平台菜单报 500 错误？
A: 检查该平台的 Schema 是否已创建（`SELECT * FROM platforms WHERE id = 'xxx'`），以及 Schema 内的 devices/alarms 表是否存在。

### Q: PostgreSQL 报 `cached plan must not change result type`？
A: 已在 `db.go` 中通过 `PreferSimpleProtocol: true` 解决。如果仍出现，检查是否使用了旧的 GORM 配置。

### Q: 如何迁移旧项目数据？
A: 使用 `backend/migrations/migrate_from_pv_platform.sql` 脚本，通过 dblink 跨库迁移。

### Q: 忘记 admin 密码？
A: 直接在数据库中重置：
```sql
UPDATE users SET password = '$2a$10$nc27zIknVMi3xgCjaVaMMO7M8NROm3x2C3gcx5bsQCiNiktMtz3Gy' WHERE username = 'admin';
-- 重置为 admin123
```

---

## License

MIT
