-- ============================================================
-- IoT 多平台后端 - 初始化 SQL
-- 说明：
--   1. public schema 存放公共表(platforms/users/devices/alarms)
--   2. 每个平台拥有独立 schema(如 schema_aluminum)，内含 devices/alarms 表
--   3. 通过 search_path 实现多租户路由
--   4. 默认管理员: admin / admin123 (bcrypt 加密)
-- ============================================================

-- 开启扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============ public schema 公共表 ============

-- platforms 平台注册表
CREATE TABLE IF NOT EXISTS platforms (
    id          VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    icon        VARCHAR(32),
    schema_name VARCHAR(64) NOT NULL,
    config      JSONB DEFAULT '{}'::jsonb,
    status      VARCHAR(16) DEFAULT 'active',
    sort_order  INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- users 用户表
CREATE TABLE IF NOT EXISTS users (
    id         VARCHAR(64) PRIMARY KEY,
    username   VARCHAR(64) UNIQUE NOT NULL,
    password   VARCHAR(128) NOT NULL,
    role       VARCHAR(32) DEFAULT 'viewer',
    platforms  VARCHAR(256),
    status     VARCHAR(16) DEFAULT 'active',
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- devices 通用设备表(public schema 统一视图)
CREATE TABLE IF NOT EXISTS devices (
    id          VARCHAR(64) PRIMARY KEY,
    platform_id VARCHAR(64),
    device_id   VARCHAR(64),
    name        VARCHAR(128),
    station_id  VARCHAR(64),
    status      VARCHAR(16) DEFAULT 'offline',
    last_seen   TIMESTAMPTZ,
    metadata    JSONB DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_devices_platform_id ON devices(platform_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_id ON devices(device_id);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);

-- alarms 通用告警表(public schema 统一视图)
CREATE TABLE IF NOT EXISTS alarms (
    id           VARCHAR(64) PRIMARY KEY,
    platform_id  VARCHAR(64),
    device_id    VARCHAR(64),
    device_name  VARCHAR(128),
    level        VARCHAR(16) DEFAULT 'info',
    type         VARCHAR(64),
    detail       TEXT,
    status       VARCHAR(16) DEFAULT 'active',
    occurred_at  TIMESTAMPTZ DEFAULT NOW(),
    resolved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alarms_platform_id ON alarms(platform_id);
CREATE INDEX IF NOT EXISTS idx_alarms_status ON alarms(status);
CREATE INDEX IF NOT EXISTS idx_alarms_occurred_at ON alarms(occurred_at);

-- ============ 平台专属 schema：铝厂 ============
CREATE SCHEMA IF NOT EXISTS schema_aluminum;

CREATE TABLE IF NOT EXISTS schema_aluminum.devices (LIKE devices INCLUDING ALL);
CREATE TABLE IF NOT EXISTS schema_aluminum.alarms  (LIKE alarms  INCLUDING ALL);

-- ============ 平台专属 schema：光伏 ============
CREATE SCHEMA IF NOT EXISTS schema_pv;

CREATE TABLE IF NOT EXISTS schema_pv.devices (LIKE devices INCLUDING ALL);
CREATE TABLE IF NOT EXISTS schema_pv.alarms  (LIKE alarms  INCLUDING ALL);

-- 光伏平台 OTA 升级相关表
CREATE TABLE IF NOT EXISTS schema_pv.firmwares (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    version      VARCHAR(64),
    file_path    VARCHAR(512),
    file_size    BIGINT DEFAULT 0,
    md5          VARCHAR(64),
    device_type  VARCHAR(64),
    upload_time  TIMESTAMPTZ DEFAULT NOW(),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schema_pv.ota_tasks (
    id             SERIAL PRIMARY KEY,
    firmware_id    INT REFERENCES schema_pv.firmwares(id),
    status         VARCHAR(16) DEFAULT 'pending',
    total_devices  INT DEFAULT 0,
    success_count  INT DEFAULT 0,
    fail_count     INT DEFAULT 0,
    progress       INT DEFAULT 0,
    created_by     VARCHAR(64),
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    completed_at   TIMESTAMPTZ,
    end_reason     VARCHAR(32),
    failed_devices JSONB
);

CREATE TABLE IF NOT EXISTS schema_pv.ota_task_devices (
    id          SERIAL PRIMARY KEY,
    task_id     INT REFERENCES schema_pv.ota_tasks(id) ON DELETE CASCADE,
    device_id   VARCHAR(64) NOT NULL,
    status      VARCHAR(16) DEFAULT 'pending',
    error_msg   TEXT,
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ota_task_devices_task_id ON schema_pv.ota_task_devices(task_id);

-- 插入示例固件数据
INSERT INTO schema_pv.firmwares (name, version, file_size, upload_time) VALUES
('SG110CX_V3.2.1.bin', '3.2.1', 1048576, NOW() - INTERVAL '3 days'),
('SG50CX_V3.1.0.bin', '3.1.0', 851968, NOW() - INTERVAL '10 days'),
('SG30CX_V3.0.5.bin', '3.0.5', 712704, NOW() - INTERVAL '21 days')
ON CONFLICT DO NOTHING;

-- ============ 平台专属 schema：电厂 ============
CREATE SCHEMA IF NOT EXISTS schema_power;

CREATE TABLE IF NOT EXISTS schema_power.devices (LIKE devices INCLUDING ALL);
CREATE TABLE IF NOT EXISTS schema_power.alarms  (LIKE alarms  INCLUDING ALL);

-- ============ 平台专属 schema：水务 ============
CREATE SCHEMA IF NOT EXISTS schema_water;

CREATE TABLE IF NOT EXISTS schema_water.devices (LIKE devices INCLUDING ALL);
CREATE TABLE IF NOT EXISTS schema_water.alarms  (LIKE alarms  INCLUDING ALL);

-- ============ 插入示例平台（铝厂 + 光伏使用自定义组件） ============

INSERT INTO platforms (id, name, icon, schema_name, config, status, sort_order)
VALUES (
    'aluminum',
    '铝厂云控平台',
    '🏭',
    'schema_aluminum',
    $${
      "navItems": [
        {"path": "/aluminum/overview", "label": "铝厂概览", "icon": "Odometer"},
        {"path": "/aluminum/cells", "label": "电解槽监控", "icon": "Monitor"},
        {"path": "/aluminum/points", "label": "点位实时", "icon": "DataLine"},
        {"label": "告警与历史", "icon": "Bell", "children": [
          {"path": "/aluminum/alarms", "label": "告警列表", "icon": "Warning"},
          {"path": "/aluminum/history", "label": "历史数据", "icon": "Clock"}
        ]}
      ],
      "pages": {
        "overview": {"type": "custom", "component": "AluminumOverview", "title": "铝厂概览"},
        "cells": {"type": "custom", "component": "AluminumCellMonitor", "title": "电解槽监控"},
        "points": {"type": "custom", "component": "AluminumPointRealtime", "title": "点位实时"},
        "alarms": {"type": "custom", "component": "AluminumAlarm", "title": "告警列表"},
        "history": {"type": "custom", "component": "AluminumHistory", "title": "历史数据"}
      }
    }$$::jsonb,
    'active',
    1
) ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config, name = EXCLUDED.name, icon = EXCLUDED.icon;

INSERT INTO platforms (id, name, icon, schema_name, config, status, sort_order)
VALUES (
    'pv',
    '光伏运维平台',
    '☀️',
    'schema_pv',
    $${
      "navItems": [
        {"path": "/pv/dashboard", "label": "仪表盘", "icon": "Odometer"},
        {"path": "/pv/devices", "label": "设备管理", "icon": "Cpu"},
        {"path": "/pv/events", "label": "事件告警", "icon": "Warning"},
        {"path": "/pv/ota", "label": "OTA升级", "icon": "Upload"},
        {"path": "/pv/stations", "label": "台区管理", "icon": "Location"}
      ],
      "pages": {
        "dashboard": {"type": "custom", "component": "PvDashboard", "title": "光伏运维仪表盘"},
        "devices": {"type": "custom", "component": "PvDeviceDetail", "title": "设备管理"},
        "events": {"type": "custom", "component": "PvEventList", "title": "事件告警"},
        "ota": {"type": "custom", "component": "PvOTA", "title": "OTA升级"},
        "stations": {"type": "custom", "component": "PvStation", "title": "台区管理"}
      }
    }$$::jsonb,
    'active',
    2
) ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config, name = EXCLUDED.name, icon = EXCLUDED.icon;

-- ============ 插入电厂平台 ============

INSERT INTO platforms (id, name, icon, schema_name, config, status, sort_order)
VALUES (
    'power',
    '电厂监控平台',
    '⚡',
    'schema_power',
    $${
      "navItems": [
        {"path": "/power/dashboard", "label": "仪表盘", "icon": "Odometer"},
        {"label": "设备监控", "icon": "Monitor", "children": [
          {"path": "/power/generators", "label": "发电机组", "icon": "Cpu"},
          {"path": "/power/boilers", "label": "锅炉系统", "icon": "Setting"},
          {"path": "/power/devices", "label": "全部设备", "icon": "Grid"}
        ]},
        {"label": "运行管理", "icon": "Operation", "children": [
          {"path": "/power/alarms", "label": "告警管理", "icon": "Bell"},
          {"path": "/power/performance", "label": "性能分析", "icon": "DataAnalysis"}
        ]}
      ],
      "pages": {
        "dashboard": {"type": "dashboard", "title": "电厂监控仪表盘", "api": "dashboard"},
        "generators": {"type": "table", "title": "发电机组", "api": "devices", "columns": [
          {"field": "deviceId", "label": "设备ID", "width": 160, "type": "text"},
          {"field": "name", "label": "机组名称", "width": 180, "type": "text"},
          {"field": "stationId", "label": "厂区", "width": 120, "type": "text"},
          {"field": "status", "label": "运行状态", "width": 100, "type": "tag", "options": {"online": "运行中", "offline": "停机", "alarm": "故障"}},
          {"field": "lastSeen", "label": "最后数据时间", "width": 180, "type": "text"}
        ]},
        "boilers": {"type": "table", "title": "锅炉系统", "api": "devices", "columns": [
          {"field": "deviceId", "label": "设备ID", "width": 160, "type": "text"},
          {"field": "name", "label": "锅炉名称", "width": 180, "type": "text"},
          {"field": "stationId", "label": "厂区", "width": 120, "type": "text"},
          {"field": "status", "label": "运行状态", "width": 100, "type": "tag", "options": {"online": "运行中", "offline": "停机", "alarm": "故障"}}
        ]},
        "devices": {"type": "table", "title": "全部设备", "api": "devices", "columns": [
          {"field": "deviceId", "label": "设备ID", "width": 160, "type": "text"},
          {"field": "name", "label": "设备名称", "width": 180, "type": "text"},
          {"field": "stationId", "label": "厂区", "width": 120, "type": "text"},
          {"field": "status", "label": "状态", "width": 100, "type": "tag", "options": {"online": "运行中", "offline": "停机", "alarm": "故障"}},
          {"field": "lastSeen", "label": "最后在线", "width": 180, "type": "text"}
        ]},
        "alarms": {"type": "table", "title": "告警管理", "api": "alarms", "columns": [
          {"field": "deviceName", "label": "设备", "width": 160, "type": "text"},
          {"field": "level", "label": "等级", "width": 100, "type": "tag", "options": {"critical": "严重", "warning": "警告", "info": "信息"}},
          {"field": "type", "label": "类型", "width": 140, "type": "text"},
          {"field": "detail", "label": "详情", "type": "text"},
          {"field": "status", "label": "状态", "width": 100, "type": "tag", "options": {"active": "未处理", "resolved": "已处理"}},
          {"field": "occurredAt", "label": "发生时间", "width": 180, "type": "text"}
        ]},
        "performance": {"type": "table", "title": "性能分析", "api": "devices", "columns": [
          {"field": "deviceId", "label": "设备ID", "width": 160, "type": "text"},
          {"field": "name", "label": "设备名称", "width": 180, "type": "text"},
          {"field": "stationId", "label": "厂区", "width": 120, "type": "text"},
          {"field": "status", "label": "状态", "width": 100, "type": "tag", "options": {"online": "运行中", "offline": "停机", "alarm": "故障"}}
        ]}
      }
    }$$::jsonb,
    'active',
    3
) ON CONFLICT (id) DO NOTHING;

-- ============ 插入水务平台 ============

INSERT INTO platforms (id, name, icon, schema_name, config, status, sort_order)
VALUES (
    'water',
    '智慧水务平台',
    '💧',
    'schema_water',
    $${
      "navItems": [
        {"path": "/water/dashboard", "label": "总览看板", "icon": "Odometer"},
        {"label": "管网监控", "icon": "Share", "children": [
          {"path": "/water/pumps", "label": "泵站设备", "icon": "Cpu"},
          {"path": "/water/valves", "label": "阀门控制", "icon": "Setting"},
          {"path": "/water/meters", "label": "流量计", "icon": "DataLine"}
        ]},
        {"label": "水质管理", "icon": "Histogram", "children": [
          {"path": "/water/quality", "label": "水质监测", "icon": "DataAnalysis"},
          {"path": "/water/alarms", "label": "告警中心", "icon": "Warning"}
        ]}
      ],
      "pages": {
        "dashboard": {"type": "dashboard", "title": "智慧水务总览看板", "api": "dashboard"},
        "pumps": {"type": "table", "title": "泵站设备", "api": "devices", "columns": [
          {"field": "deviceId", "label": "设备ID", "width": 160, "type": "text"},
          {"field": "name", "label": "泵站名称", "width": 180, "type": "text"},
          {"field": "stationId", "label": "区域", "width": 120, "type": "text"},
          {"field": "status", "label": "运行状态", "width": 100, "type": "tag", "options": {"online": "运行中", "offline": "停机", "alarm": "故障"}},
          {"field": "lastSeen", "label": "最后数据时间", "width": 180, "type": "text"}
        ]},
        "valves": {"type": "table", "title": "阀门控制", "api": "devices", "columns": [
          {"field": "deviceId", "label": "阀门ID", "width": 160, "type": "text"},
          {"field": "name", "label": "阀门名称", "width": 180, "type": "text"},
          {"field": "stationId", "label": "区域", "width": 120, "type": "text"},
          {"field": "status", "label": "状态", "width": 100, "type": "tag", "options": {"online": "已开启", "offline": "已关闭", "alarm": "故障"}}
        ]},
        "meters": {"type": "table", "title": "流量计", "api": "devices", "columns": [
          {"field": "deviceId", "label": "仪表ID", "width": 160, "type": "text"},
          {"field": "name", "label": "安装位置", "width": 180, "type": "text"},
          {"field": "stationId", "label": "区域", "width": 120, "type": "text"},
          {"field": "status", "label": "状态", "width": 100, "type": "tag", "options": {"online": "正常", "offline": "离线", "alarm": "异常"}}
        ]},
        "quality": {"type": "table", "title": "水质监测", "api": "devices", "columns": [
          {"field": "deviceId", "label": "监测点ID", "width": 160, "type": "text"},
          {"field": "name", "label": "监测点名称", "width": 180, "type": "text"},
          {"field": "stationId", "label": "区域", "width": 120, "type": "text"},
          {"field": "status", "label": "水质状态", "width": 100, "type": "tag", "options": {"online": "达标", "offline": "待检", "alarm": "超标"}}
        ]},
        "alarms": {"type": "table", "title": "告警中心", "api": "alarms", "columns": [
          {"field": "deviceName", "label": "设备", "width": 160, "type": "text"},
          {"field": "level", "label": "等级", "width": 100, "type": "tag", "options": {"critical": "严重", "warning": "警告", "info": "信息"}},
          {"field": "type", "label": "类型", "width": 140, "type": "text"},
          {"field": "detail", "label": "详情", "type": "text"},
          {"field": "status", "label": "状态", "width": 100, "type": "tag", "options": {"active": "未处理", "resolved": "已处理"}},
          {"field": "occurredAt", "label": "发生时间", "width": 180, "type": "text"}
        ]}
      }
    }$$::jsonb,
    'active',
    4
) ON CONFLICT (id) DO NOTHING;
-- ============ 插入默认管理员 ============
-- 用户名: admin  密码: admin123 (bcrypt 加密)
INSERT INTO users (id, username, password, role, platforms, status)
VALUES (
    'admin',
    'admin',
    '$2a$10$nc27zIknVMi3xgCjaVaMMO7M8NROm3x2C3gcx5bsQCiNiktMtz3Gy',
    'super_admin',
    'aluminum,pv,power,water',
    'active'
) ON CONFLICT (username) DO NOTHING;

-- ============ 插入示例设备与告警数据(可选，便于演示) ============

INSERT INTO schema_aluminum.devices (id, platform_id, device_id, name, station_id, status, last_seen, metadata)
VALUES
    ('d-al-001', 'aluminum', 'AL-ELEC-001', '1#电解槽', 'station-01', 'online',  NOW(), '{"voltage": "4.2V", "current": "320kA"}'),
    ('d-al-002', 'aluminum', 'AL-ELEC-002', '2#电解槽', 'station-01', 'online',  NOW(), '{"voltage": "4.3V", "current": "315kA"}'),
    ('d-al-003', 'aluminum', 'AL-FURN-001', '保温炉',   'station-02', 'offline', NOW() - INTERVAL '2 hours', '{"temp": "0"}'),
    ('d-al-004', 'aluminum', 'AL-CRANE-01', '铸造天车', 'station-03', 'alarm',  NOW() - INTERVAL '10 min', '{"load": "95%"}')
ON CONFLICT (id) DO NOTHING;

INSERT INTO schema_aluminum.alarms (id, platform_id, device_id, device_name, level, type, detail, status, occurred_at)
VALUES
    ('a-al-001', 'aluminum', 'AL-CRANE-01', '铸造天车', 'critical', 'overload',  '载荷达到 95%，超过预警阈值', 'active',   NOW() - INTERVAL '10 min'),
    ('a-al-002', 'aluminum', 'AL-FURN-001', '保温炉',   'warning',  'disconnection', '设备离线超过 2 小时', 'active', NOW() - INTERVAL '2 hours'),
    ('a-al-003', 'aluminum', 'AL-ELEC-001', '1#电解槽', 'info',     'voltage',   '电压轻微波动',           'resolved', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

INSERT INTO schema_pv.devices (id, platform_id, device_id, name, station_id, status, last_seen, metadata)
VALUES
    ('d-pv-001', 'pv', 'PV-INV-001', '1#逆变器', 'plant-A', 'online',  NOW(), '{"power": "50kW", "efficiency": "98%"}'),
    ('d-pv-002', 'pv', 'PV-INV-002', '2#逆变器', 'plant-A', 'online',  NOW(), '{"power": "48kW", "efficiency": "97%"}'),
    ('d-pv-003', 'pv', 'PV-MET-001', '气象站',   'plant-B', 'offline', NOW() - INTERVAL '30 min', '{"irradiance": "0"}')
ON CONFLICT (id) DO NOTHING;

INSERT INTO schema_pv.alarms (id, platform_id, device_id, device_name, level, type, detail, status, occurred_at)
VALUES
    ('a-pv-001', 'pv', 'PV-MET-001', '气象站', 'warning', 'disconnection', '气象站离线 30 分钟', 'active', NOW() - INTERVAL '30 min'),
    ('a-pv-002', 'pv', 'PV-INV-001', '1#逆变器', 'info', 'efficiency', '转换效率低于阈值', 'resolved', NOW() - INTERVAL '3 hours')
ON CONFLICT (id) DO NOTHING;

-- ============ 电厂 demo 数据 ============

INSERT INTO schema_power.devices (id, platform_id, device_id, name, station_id, status, last_seen, metadata)
VALUES
    ('d-pw-001', 'power', 'PW-GEN-001', '1#汽轮发电机组', 'main-plant', 'online', NOW(), '{"power": "300MW", "speed": "3000rpm", "temp": "538℃"}'),
    ('d-pw-002', 'power', 'PW-GEN-002', '2#汽轮发电机组', 'main-plant', 'online', NOW(), '{"power": "330MW", "speed": "3000rpm", "temp": "542℃"}'),
    ('d-pw-003', 'power', 'PW-GEN-003', '3#燃气轮机组', 'main-plant', 'online', NOW(), '{"power": "250MW", "speed": "3600rpm", "temp": "620℃"}'),
    ('d-pw-004', 'power', 'PW-BLR-001', '1#煤粉锅炉', 'main-plant', 'online', NOW(), '{"steamFlow": "1025t/h", "pressure": "17.5MPa"}'),
    ('d-pw-005', 'power', 'PW-BLR-002', '2#循环流化床锅炉', 'main-plant', 'online', NOW(), '{"steamFlow": "480t/h", "pressure": "9.8MPa"}'),
    ('d-pw-006', 'power', 'PW-BLR-003', '3#余热锅炉', 'main-plant', 'offline', NOW() - INTERVAL '3 hours', '{"steamFlow": "0"}'),
    ('d-pw-007', 'power', 'PW-COOL-01', '冷却塔A', 'aux-area', 'online', NOW(), '{"flow": "12000m³/h"}'),
    ('d-pw-008', 'power', 'PW-COOL-02', '冷却塔B', 'aux-area', 'alarm', NOW() - INTERVAL '15 min', '{"flow": "8000m³/h"}'),
    ('d-pw-009', 'power', 'PW-ESP-001', '静电除尘器', 'env-area', 'online', NOW(), '{"efficiency": "99.8%"}'),
    ('d-pw-010', 'power', 'PW-FGD-001', '脱硫装置', 'env-area', 'online', NOW(), '{"so2Removal": "95%"}'),
    ('d-pw-011', 'power', 'PW-SCR-001', '脱硝装置', 'env-area', 'offline', NOW() - INTERVAL '6 hours', '{"noxRemoval": "0%"}'),
    ('d-pw-012', 'power', 'PW-PUMP-01', '给水泵A', 'aux-area', 'online', NOW(), '{"flow": "450t/h"}')
ON CONFLICT (id) DO NOTHING;

INSERT INTO schema_power.alarms (id, platform_id, device_id, device_name, level, type, detail, status, occurred_at)
VALUES
    ('a-pw-001', 'power', 'PW-COOL-02', '冷却塔B', 'critical', 'overtemp', '冷却水出水温度 40℃，超过警戒值 38℃', 'active', NOW() - INTERVAL '15 min'),
    ('a-pw-002', 'power', 'PW-BLR-003', '3#余热锅炉', 'warning', 'shutdown', '余热锅炉非计划停机', 'active', NOW() - INTERVAL '3 hours'),
    ('a-pw-003', 'power', 'PW-SCR-001', '脱硝装置', 'warning', 'offline', '脱硝装置离线超过 6 小时', 'active', NOW() - INTERVAL '6 hours'),
    ('a-pw-004', 'power', 'PW-GEN-003', '3#燃气轮机组', 'info', 'vibration', '轴承振动幅值轻微升高', 'resolved', NOW() - INTERVAL '2 days'),
    ('a-pw-005', 'power', 'PW-BLR-001', '1#煤粉锅炉', 'info', 'maintenance', '计划性检修完成', 'resolved', NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

-- ============ 水务 demo 数据 ============

INSERT INTO schema_water.devices (id, platform_id, device_id, name, station_id, status, last_seen, metadata)
VALUES
    ('d-w-001', 'water', 'W-PUMP-001', '一级泵站A', 'north-zone', 'online', NOW(), '{"flow": "1200m³/h", "pressure": "0.45MPa"}'),
    ('d-w-002', 'water', 'W-PUMP-002', '一级泵站B', 'north-zone', 'online', NOW(), '{"flow": "1150m³/h", "pressure": "0.42MPa"}'),
    ('d-w-003', 'water', 'W-PUMP-003', '二级泵站A', 'south-zone', 'online', NOW(), '{"flow": "800m³/h", "pressure": "0.38MPa"}'),
    ('d-w-004', 'water', 'W-PUMP-004', '加压泵站', 'west-zone', 'alarm', NOW() - INTERVAL '20 min', '{"flow": "0"}'),
    ('d-w-005', 'water', 'W-PUMP-005', '排污泵', 'treatment-plant', 'online', NOW(), '{"flow": "200m³/h"}'),
    ('d-w-006', 'water', 'W-VALVE-001', '主进水阀', 'north-zone', 'online', NOW(), '{"openness": "100%"}'),
    ('d-w-007', 'water', 'W-VALVE-002', '南区调节阀', 'south-zone', 'online', NOW(), '{"openness": "65%"}'),
    ('d-w-008', 'water', 'W-VALVE-003', '西区泄压阀', 'west-zone', 'offline', NOW() - INTERVAL '2 hours', '{"openness": "0%"}'),
    ('d-w-009', 'water', 'W-VALVE-004', '回流控制阀', 'treatment-plant', 'online', NOW(), '{"openness": "30%"}'),
    ('d-w-010', 'water', 'W-METER-001', '总出水流量计', 'north-zone', 'online', NOW(), '{"reading": "2350m³/h"}'),
    ('d-w-011', 'water', 'W-METER-002', '南区流量计', 'south-zone', 'online', NOW(), '{"reading": "800m³/h"}'),
    ('d-w-012', 'water', 'W-METER-003', '西区流量计', 'west-zone', 'offline', NOW() - INTERVAL '4 hours', '{"reading": "0"}'),
    ('d-w-013', 'water', 'W-QUAL-001', '水源地监测点', 'intake', 'online', NOW(), '{"pH": "7.2", "turbidity": "0.8NTU"}'),
    ('d-w-014', 'water', 'W-QUAL-002', '出厂水监测点', 'treatment-plant', 'online', NOW(), '{"pH": "7.5", "turbidity": "0.3NTU"}'),
    ('d-w-015', 'water', 'W-QUAL-003', '管网末梢监测点', 'south-zone', 'alarm', NOW() - INTERVAL '30 min', '{"pH": "8.1", "turbidity": "2.5NTU"}')
ON CONFLICT (id) DO NOTHING;

INSERT INTO schema_water.alarms (id, platform_id, device_id, device_name, level, type, detail, status, occurred_at)
VALUES
    ('a-w-001', 'water', 'W-PUMP-004', '加压泵站', 'critical', 'failure', '加压泵站突发停机，西区供水压力下降', 'active', NOW() - INTERVAL '20 min'),
    ('a-w-002', 'water', 'W-QUAL-003', '管网末梢监测点', 'critical', 'quality_exceed', '浊度 2.5NTU 超标（限值 1.0NTU），余氯偏低', 'active', NOW() - INTERVAL '30 min'),
    ('a-w-003', 'water', 'W-METER-003', '西区流量计', 'warning', 'offline', '西区流量计离线超过 4 小时', 'active', NOW() - INTERVAL '4 hours'),
    ('a-w-004', 'water', 'W-VALVE-003', '西区泄压阀', 'warning', 'abnormal', '西区泄压阀异常关闭', 'active', NOW() - INTERVAL '2 hours'),
    ('a-w-005', 'water', 'W-PUMP-001', '一级泵站A', 'info', 'maintenance', '计划性保养完成', 'resolved', NOW() - INTERVAL '3 days'),
    ('a-w-006', 'water', 'W-QUAL-001', '水源地监测点', 'info', 'notice', 'pH 值轻微偏高', 'resolved', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;
