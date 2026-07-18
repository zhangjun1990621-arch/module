-- ============================================================
-- 数据迁移脚本：从 pv_platform 迁移到 iot_platform
-- 在 iot_platform 数据库中执行
-- 使用 dblink 跨库读取 pv_platform 数据
-- ============================================================

-- 1. 创建 dblink 扩展
CREATE EXTENSION IF NOT EXISTS dblink;

-- 2. 迁移用户数据
-- 旧表 users(id int, username, password_hash, role, platforms 'al,pv', status, last_login)
-- 新表 users(id varchar, username, password, role, platforms 'aluminum,pv', status, last_login)
INSERT INTO users (id, username, password, role, platforms, status, last_login)
SELECT
    'old-' || id::text,
    username,
    password_hash,
    CASE
        WHEN role = 'ops_aluminum' THEN 'admin'
        WHEN role = 'ops_pv' THEN 'admin'
        WHEN role = 'ops' THEN 'admin'
        WHEN role = 'readonly' THEN 'viewer'
        ELSE role
    END,
    REPLACE(platforms, 'al', 'aluminum'),
    status,
    last_login
FROM dblink(
    'host=127.0.0.1 port=5432 dbname=pv_platform user=postgres password=Qsh@2026#PvSecure',
    'SELECT id, username, password_hash, role, platforms, status, last_login FROM users'
) AS t(
    id int,
    username text,
    password_hash text,
    role text,
    platforms text,
    status text,
    last_login timestamptz
)
ON CONFLICT (username) DO NOTHING;

-- 3. 迁移光伏设备数据
-- 旧表 devices(id, device_type, model, serial_no, capacity, firmware, software, hardware, station_id, status, last_online, signal_strength)
-- 新表 schema_pv.devices(id, platform_id, device_id, name, station_id, status, last_seen, metadata)
INSERT INTO schema_pv.devices (id, platform_id, device_id, name, station_id, status, last_seen, metadata)
SELECT
    'd-pv-' || RIGHT(id, 4),
    'pv',
    id,
    COALESCE(model || ' ' || serial_no, id),
    station_id::text,
    status,
    last_online,
    jsonb_build_object(
        'model', model,
        'serialNo', serial_no,
        'capacity', capacity,
        'firmware', firmware,
        'software', software,
        'hardware', hardware,
        'signalStrength', signal_strength,
        'deviceType', device_type
    )
FROM dblink(
    'host=127.0.0.1 port=5432 dbname=pv_platform user=postgres password=Qsh@2026#PvSecure',
    'SELECT id, device_type, model, serial_no, capacity, firmware, software, hardware, station_id, status, last_online, signal_strength FROM devices'
) AS t(
    id text,
    device_type text,
    model text,
    serial_no text,
    capacity int,
    firmware text,
    software text,
    hardware text,
    station_id int,
    status text,
    last_online timestamptz,
    signal_strength int
)
ON CONFLICT (id) DO NOTHING;

-- 4. 迁移光伏事件/告警数据
-- 旧表 events(id, device_id, event_type, detail, severity, status, occurred_at, recovered_at)
-- 新表 schema_pv.alarms(id, platform_id, device_id, device_name, level, type, detail, status, occurred_at, resolved_at)
INSERT INTO schema_pv.alarms (id, platform_id, device_id, device_name, level, type, detail, status, occurred_at, resolved_at)
SELECT
    'a-pv-' || LPAD(id::text, 4, '0'),
    'pv',
    device_id,
    device_id,
    CASE severity
        WHEN 'danger' THEN 'critical'
        WHEN 'warning' THEN 'warning'
        WHEN 'info' THEN 'info'
        ELSE 'info'
    END,
    event_type,
    detail::text,
    CASE status
        WHEN 'recovered' THEN 'resolved'
        ELSE status
    END,
    occurred_at,
    recovered_at
FROM dblink(
    'host=127.0.0.1 port=5432 dbname=pv_platform user=postgres password=Qsh@2026#PvSecure',
    'SELECT id, device_id, event_type, detail::text, severity, status, occurred_at, recovered_at FROM events'
) AS t(
    id int,
    device_id text,
    event_type text,
    detail text,
    severity text,
    status text,
    occurred_at timestamptz,
    recovered_at timestamptz
)
ON CONFLICT (id) DO NOTHING;

-- 5. 迁移台区数据到光伏设备的 station_id 映射
UPDATE schema_pv.devices d
SET station_id = s.name
FROM dblink(
    'host=127.0.0.1 port=5432 dbname=pv_platform user=postgres password=Qsh@2026#PvSecure',
    'SELECT id, name FROM stations'
) AS t(id int, name text) s
WHERE d.station_id = s.id::text;

-- 6. 验证迁移结果
SELECT 'users' as table_name, COUNT(*) as cnt FROM users WHERE id LIKE 'old-%'
UNION ALL
SELECT 'pv_devices', COUNT(*) FROM schema_pv.devices WHERE id LIKE 'd-pv-%'
UNION ALL
SELECT 'pv_alarms', COUNT(*) FROM schema_pv.alarms WHERE id LIKE 'a-pv-%';
