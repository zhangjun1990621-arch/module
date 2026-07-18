# API 接口文档

## 通用约定

### Base URL

```
http://服务器IP:8099/api
```

### 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code: 0` 表示成功，非 0 表示失败
- HTTP 状态码：200 成功，400 参数错误，401 未认证，403 无权限，404 不存在，500 服务器错误

### 分页响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

### 认证方式

除登录接口外，所有接口需在 Header 中携带 JWT Token：

```
Authorization: Bearer <token>
```

---

## 1. 认证接口

### POST /api/auth/login

用户登录，获取 JWT Token。

**请求体**：
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "admin",
      "username": "admin",
      "role": "super_admin",
      "platforms": "aluminum,pv,power,water",
      "status": "active"
    }
  }
}
```

**错误**：
- 401: 用户名或密码错误

---

### GET /api/auth/profile

获取当前登录用户信息。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "admin",
    "username": "admin",
    "role": "super_admin",
    "platforms": "aluminum,pv,power,water",
    "status": "active",
    "lastLogin": "2026-07-16T12:00:00+08:00"
  }
}
```

---

### POST /api/auth/logout

登出（无状态，前端清除 Token 即可）。

---

## 2. 平台管理接口

### GET /api/platforms

获取所有平台列表。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "aluminum",
      "name": "铝厂云控平台",
      "icon": "🏭",
      "schema": "schema_aluminum",
      "config": {
        "navItems": [...],
        "pages": {...}
      },
      "status": "active",
      "sortOrder": 1
    }
  ]
}
```

---

### GET /api/platforms/:id

获取单个平台详情。

---

### POST /api/platforms

创建新平台（仅 super_admin）。

**请求体**：
```json
{
  "id": "gas",
  "name": "燃气监控平台",
  "icon": "🔥",
  "schema": "schema_gas",
  "config": {
    "navItems": [
      {"path": "/gas/dashboard", "label": "仪表盘", "icon": "Odometer"}
    ],
    "pages": {
      "dashboard": {"type": "dashboard", "title": "仪表盘", "api": "dashboard"}
    }
  },
  "status": "active",
  "sortOrder": 6
}
```

**说明**：后端自动创建 Schema 和设备/告警表。

---

### PUT /api/platforms/:id

更新平台信息（修改 status 需 super_admin）。

---

### DELETE /api/platforms/:id

删除平台（仅 super_admin）。

**说明**：后端自动执行 `DROP SCHEMA CASCADE` 删除该平台的所有数据。

---

## 3. 用户管理接口（仅 super_admin）

### GET /api/users

获取用户列表（支持分页、搜索、筛选）。

**查询参数**：
| 参数 | 说明 |
|------|------|
| page | 页码（默认 1） |
| pageSize | 每页条数（默认 20，最大 200） |
| search | 用户名模糊搜索 |
| role | 角色筛选 |
| status | 状态筛选 |

---

### POST /api/users

创建新用户。

**请求体**：
```json
{
  "username": "operator1",
  "password": "secure123",
  "role": "viewer",
  "platforms": "aluminum,pv",
  "status": "active"
}
```

**校验规则**：
- 用户名 2-64 字符，唯一
- 密码 6-128 字符
- 非超管用户必须指定 platforms
- role 必须是 super_admin / admin / viewer 之一

---

### PUT /api/users/:id

更新用户信息（角色、平台权限、状态）。

**请求体**（所有字段可选）：
```json
{
  "role": "admin",
  "platforms": "aluminum,pv,power",
  "status": "active"
}
```

**保护逻辑**：
- 不能修改自己的账号
- 不能降级最后一个超级管理员
- 不能禁用/删除最后一个活跃超级管理员

---

### DELETE /api/users/:id

删除用户。

---

### PUT /api/users/:id/password

重置用户密码。

**请求体**：
```json
{
  "password": "newpassword123"
}
```

---

## 4. 平台业务接口

> 以下接口中 `:platform` 为平台 ID（如 aluminum、pv、power 等）。
> 需通过 `PlatformAccess` 中间件校验用户是否有权访问该平台。

### GET /api/:platform/devices

获取设备列表。

**查询参数**：
| 参数 | 说明 |
|------|------|
| page | 页码 |
| pageSize | 每页条数 |
| search | 搜索（name / deviceId / stationId 模糊匹配） |
| status | 状态筛选（online / offline / alarm） |

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": "d-al-001",
        "platformId": "aluminum",
        "deviceId": "AL-ELEC-001",
        "name": "1#电解槽",
        "stationId": "station-01",
        "status": "online",
        "lastSeen": "2026-07-16T12:00:00+08:00",
        "metadata": {"voltage": "4.2V", "current": "320kA"},
        "createdAt": "...",
        "updatedAt": "..."
      }
    ],
    "total": 8,
    "page": 1,
    "pageSize": 20
  }
}
```

---

### GET /api/:platform/devices/:id

获取单个设备详情。

---

### POST /api/:platform/devices

创建设备。

**请求体**：
```json
{
  "deviceId": "AL-ELEC-005",
  "name": "5#电解槽",
  "stationId": "station-02",
  "status": "online",
  "metadata": {"voltage": "4.2V"}
}
```

---

### PUT /api/:platform/devices/:id

更新设备信息。

---

### DELETE /api/:platform/devices/:id

删除设备。

---

### GET /api/:platform/alarms

获取告警列表。

**查询参数**：
| 参数 | 说明 |
|------|------|
| page | 页码 |
| pageSize | 每页条数 |
| level | 级别筛选（info / warning / critical） |
| status | 状态筛选（active / resolved） |
| deviceId | 设备 ID 筛选 |
| startTime | 开始时间（RFC3339） |
| endTime | 结束时间（RFC3339） |

---

### PUT /api/:platform/alarms/:id/resolve

将告警标记为已恢复。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "a-al-001",
    "status": "resolved",
    "resolvedAt": "2026-07-16T12:30:00+08:00"
  }
}
```

---

### GET /api/:platform/dashboard

获取仪表盘聚合数据。

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "kpi": {
      "deviceTotal": 8,
      "online": 5,
      "offline": 1,
      "alarm": 2,
      "activeAlarm": 3
    },
    "recentAlarms": [
      {
        "id": "a-al-001",
        "deviceName": "铸造天车",
        "level": "critical",
        "type": "overload",
        "detail": "载荷达到 95%，超过预警阈值",
        "status": "active",
        "occurredAt": "2026-07-16T11:50:00+08:00"
      }
    ],
    "trend": [
      {"date": "2026-07-10", "count": 2},
      {"date": "2026-07-11", "count": 5},
      {"date": "2026-07-12", "count": 0},
      {"date": "2026-07-13", "count": 3},
      {"date": "2026-07-14", "count": 1},
      {"date": "2026-07-15", "count": 4},
      {"date": "2026-07-16", "count": 3}
    ]
  }
}
```

---

## 5. 健康检查

### GET /api/health

无需认证。

**响应**：
```json
{"status": "ok"}
```

---

## 6. 错误码说明

| HTTP 状态码 | code | 说明 |
|------------|------|------|
| 200 | 0 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未认证 / Token 过期 |
| 403 | 403 | 无权限 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源冲突（如用户名已存在） |
| 500 | 500 | 服务器内部错误 |
