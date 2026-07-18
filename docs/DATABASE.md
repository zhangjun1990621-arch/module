# 数据库设计文档

## 1. 数据库概览

- **数据库引擎**：PostgreSQL 16
- **数据库名**：`iot_platform`
- **字符编码**：UTF-8
- **时区**：Asia/Shanghai（TIMESTAMPTZ 存储所有时间字段）

### Schema 架构

```
iot_platform (数据库)
├── public                    # 公共 Schema
│   ├── platforms             # 平台注册表
│   ├── users                 # 用户表
│   ├── devices               # 通用设备视图（可选，主要用于自动迁移）
│   └── alarms                # 通用告警视图（可选，主要用于自动迁移）
├── schema_aluminum           # 铝厂专属
│   ├── devices               # 铝厂设备表
│   └── alarms                # 铝厂告警表
├── schema_pv                 # 光伏专属
│   ├── devices
│   └── alarms
├── schema_power              # 电厂专属
│   ├── devices
│   └── alarms
├── schema_water              # 水务专属
│   ├── devices
│   └── alarms
└── schema_farm               # 农场专属
    ├── devices
    └── alarms
```

---

## 2. 公共表结构

### 2.1 platforms（平台注册表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(64) | PRIMARY KEY | 平台唯一标识（如 aluminum） |
| name | VARCHAR(128) | NOT NULL | 平台显示名称 |
| icon | VARCHAR(32) | | Emoji 或图标名 |
| schema_name | VARCHAR(64) | NOT NULL | 对应的 PostgreSQL Schema 名 |
| config | JSONB | DEFAULT '{}' | UI 配置（NavItems + Pages） |
| status | VARCHAR(16) | DEFAULT 'active' | active / inactive |
| sort_order | INT | DEFAULT 0 | 排序权重（升序） |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | 更新时间 |

**config 字段 JSONB 结构**：

```json
{
  "navItems": [
    {
      "path": "/aluminum/overview",
      "label": "铝厂概览",
      "icon": "Odometer",
      "children": [                    // 可选，二级菜单
        {
          "path": "/aluminum/cells",
          "label": "电解槽监控",
          "icon": "Monitor"
        }
      ]
    }
  ],
  "pages": {
    "overview": {
      "type": "custom",                // table / dashboard / chart / custom
      "title": "铝厂概览",
      "component": "AluminumOverview", // type=custom 时必填
      "api": "devices",                // type=table 时必填
      "columns": [                     // type=table 时必填
        {
          "field": "deviceId",
          "label": "设备ID",
          "width": 160,
          "type": "text",              // text / tag / dot / temperature / number
          "options": {}                // type=tag 时提供值映射
        }
      ]
    }
  }
}
```

### 2.2 users（用户表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(64) | PRIMARY KEY | 用户唯一标识（UUID） |
| username | VARCHAR(64) | UNIQUE NOT NULL | 登录用户名 |
| password | VARCHAR(128) | NOT NULL | bcrypt 哈希密码 |
| role | VARCHAR(32) | DEFAULT 'viewer' | super_admin / admin / viewer |
| platforms | VARCHAR(256) | | 可访问平台 ID 列表（逗号分隔） |
| status | VARCHAR(16) | DEFAULT 'active' | active / inactive |
| last_login | TIMESTAMPTZ | | 最后登录时间 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | |

**角色权限**：

| 角色 | 平台访问 | 平台管理 | 权限管理 | 用户管理 |
|------|---------|---------|---------|---------|
| super_admin | 全部 | ✅ | ✅ | ✅ |
| admin | platforms 字段指定 | ❌ | ❌ | ❌ |
| viewer | platforms 字段指定 | ❌ | ❌ | ❌ |

### 2.3 devices / alarms（公共表）

公共 Schema 中的 devices 和 alarms 表主要用于：
1. GORM AutoMigrate 的模板表
2. 创建新平台 Schema 时通过 `LIKE` 语句复制表结构

实际业务数据存储在各平台专属 Schema 中。

---

## 3. 平台专属表结构

### 3.1 devices（设备表）

每个平台 Schema 中的 devices 表结构完全相同：

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(64) | PRIMARY KEY | 设备记录 UUID |
| platform_id | VARCHAR(64) | INDEX | 平台 ID |
| device_id | VARCHAR(64) | INDEX | 业务设备编号 |
| name | VARCHAR(128) | | 设备名称 |
| station_id | VARCHAR(64) | INDEX | 站点/区域 ID |
| status | VARCHAR(16) | DEFAULT 'offline' | online / offline / alarm |
| last_seen | TIMESTAMPTZ | | 最后在线时间 |
| metadata | JSONB | DEFAULT '{}' | 扩展属性（电压、电流、温度等） |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | |

**索引**：
- `idx_devices_platform_id` (platform_id)
- `idx_devices_device_id` (device_id)
- `idx_devices_status` (status)

### 3.2 alarms（告警表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(64) | PRIMARY KEY | 告警记录 UUID |
| platform_id | VARCHAR(64) | INDEX | 平台 ID |
| device_id | VARCHAR(64) | INDEX | 关联设备 ID |
| device_name | VARCHAR(128) | | 设备名称（冗余） |
| level | VARCHAR(16) | DEFAULT 'info' | info / warning / critical |
| type | VARCHAR(64) | | 告警类型（overload/voltage/...） |
| detail | TEXT | | 告警详情描述 |
| status | VARCHAR(16) | DEFAULT 'active' | active / resolved |
| occurred_at | TIMESTAMPTZ | INDEX | 告警发生时间 |
| resolved_at | TIMESTAMPTZ | | 告警恢复时间 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | |

**索引**：
- `idx_alarms_platform_id` (platform_id)
- `idx_alarms_status` (status)
- `idx_alarms_occurred_at` (occurred_at)

---

## 4. 多 Schema 路由机制

### 4.1 SET LOCAL search_path

```sql
-- 开启事务
BEGIN;

-- 设置当前事务的 search_path（仅事务内生效）
SET LOCAL search_path TO schema_aluminum, public;

-- 此事务内所有查询自动命中 schema_aluminum.devices 和 schema_aluminum.alarms
SELECT * FROM devices WHERE status = 'online';  -- 命中 schema_aluminum.devices
SELECT * FROM alarms WHERE level = 'critical';  -- 命中 schema_aluminum.alarms

-- 提交事务（search_path 自动还原）
COMMIT;
```

### 4.2 为什么用 SET LOCAL 而非 SET

| 方式 | 生效范围 | 连接池安全 |
|------|---------|-----------|
| `SET search_path` | 当前连接（持久） | ❌ 连接复用时残留 |
| `SET LOCAL search_path` | 当前事务（自动还原） | ✅ 事务结束后自动还原 |

### 4.3 PreferSimpleProtocol

GORM + pgx 默认使用预处理语句（Prepared Statements），连接池会缓存执行计划。当 `search_path` 变化时，缓存的执行计划与新的表结构不匹配，导致 `cached plan must not change result type` 错误。

解决方案：`PreferSimpleProtocol: true` 禁用预处理语句缓存。

```go
gorm.Open(postgres.New(postgres.Config{
    DSN:                  dbCfg.DSN(),
    PreferSimpleProtocol: true,  // 禁用预处理语句缓存
}), &gorm.Config{...})
```

---

## 5. 创建新平台的 SQL 流程

```sql
-- 1. 创建 Schema
CREATE SCHEMA IF NOT EXISTS schema_gas;

-- 2. 创建设备表（从模板复制结构）
CREATE TABLE IF NOT EXISTS schema_gas.devices (LIKE devices INCLUDING ALL);

-- 3. 创建告警表
CREATE TABLE IF NOT EXISTS schema_gas.alarms (LIKE alarms INCLUDING ALL);

-- 4. 注册平台
INSERT INTO platforms (id, name, icon, schema_name, config, status, sort_order)
VALUES (
    'gas',
    '燃气监控平台',
    '🔥',
    'schema_gas',
    '{"navItems": [...], "pages": {...}}'::jsonb,
    'active',
    6
);
```

---

## 6. 数据迁移

### 6.1 从旧项目（pv_platform）迁移

使用 `migrations/migrate_from_pv_platform.sql` 脚本：
- 通过 `dblink` 跨库读取 pv_platform 数据
- 迁移用户数据（角色映射：ops_aluminum → admin, readonly → viewer）
- 迁移光伏设备数据（metadata 字段重组）
- 迁移光伏事件/告警数据（severity 映射：danger → critical）

### 6.2 备份与恢复

```bash
# 备份全库
pg_dump -h 127.0.0.1 -p 5433 -U postgres iot_platform > backup.sql

# 恢复
psql -h 127.0.0.1 -p 5433 -U postgres -d iot_platform < backup.sql

# Docker 环境
docker exec module-postgres pg_dump -U postgres iot_platform > backup.sql
cat backup.sql | docker exec -i module-postgres psql -U postgres -d iot_platform
```
