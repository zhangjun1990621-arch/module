/**
 * 光伏平台演示态 mock 数据层
 * ----------------------------------------------------------------
 * 仅在 VITE_SKIP_AUTH==='true'（演示部署、无后端）时由 request.ts 的自定义
 * adapter 调用。所有 matchPvMock 返回的是【业务负载】，adapter 会包成
 * { code:0, data: 负载 } —— 与后端统一响应壳一致，视图层 res.data 即此负载。
 *
 * 数据结构严格对齐视图层消费方式（字段名/数组or对象/嵌套），见探查报告。
 * 后续接真实后端时，删除 request.ts 里的 adapter 注入即可，本文件可保留作 e2e fixture。
 */

const now = Date.now()
const iso = (offsetMin: number) => new Date(now + offsetMin * 60000).toISOString()
const today = (h: number, m: number) => {
  const d = new Date(now)
  d.setHours(h, m, 0, 0)
  return d.toISOString()
}

/* ---------------- 基础数据 ---------------- */

export interface MockStation {
  id: number
  name: string
  code: string
  region: string
  deviceCount: number
}

export const STATIONS: MockStation[] = [
  { id: 1, name: '1号方阵台区', code: 'PV-ST-001', region: '厂区A', deviceCount: 4 },
  { id: 2, name: '2号方阵台区', code: 'PV-ST-002', region: '厂区A', deviceCount: 3 },
  { id: 3, name: '3号方阵台区', code: 'PV-ST-003', region: '厂区B', deviceCount: 3 },
  { id: 4, name: '综合楼台区', code: 'PV-ST-004', region: '办公楼', deviceCount: 2 }
]

export interface MockDevice {
  id: string
  deviceType: string
  model: string
  serialNo: string
  capacity: number
  firmware: string
  software: string
  hardware: string
  stationId: number | null
  status: string
  lastOnline: string | null
  signalStrength: number
}

// 12 台逆变器，分布在 4 个台区
export const DEVICES: MockDevice[] = [
  { id: 'INV-2024-001', deviceType: 'inverter', model: 'SG110CX', serialNo: 'SN240001A', capacity: 110000, firmware: 'V3.2.1', software: 'V3.2.1', hardware: 'V2.0', stationId: 1, status: 'online', lastOnline: iso(-1), signalStrength: 28 },
  { id: 'INV-2024-002', deviceType: 'inverter', model: 'SG110CX', serialNo: 'SN240002A', capacity: 110000, firmware: 'V3.2.1', software: 'V3.2.1', hardware: 'V2.0', stationId: 1, status: 'online', lastOnline: iso(-2), signalStrength: 26 },
  { id: 'INV-2024-003', deviceType: 'inverter', model: 'SG50CX', serialNo: 'SN240003A', capacity: 50000, firmware: 'V3.1.0', software: 'V3.1.0', hardware: 'V1.9', stationId: 1, status: 'online', lastOnline: iso(-3), signalStrength: 22 },
  { id: 'INV-2024-004', deviceType: 'inverter', model: 'SG50CX', serialNo: 'SN240004A', capacity: 50000, firmware: 'V3.1.0', software: 'V3.1.0', hardware: 'V1.9', stationId: 1, status: 'offline', lastOnline: iso(-180), signalStrength: 0 },
  { id: 'INV-2024-005', deviceType: 'inverter', model: 'SG110CX', serialNo: 'SN240005B', capacity: 110000, firmware: 'V3.2.1', software: 'V3.2.1', hardware: 'V2.0', stationId: 2, status: 'online', lastOnline: iso(-1), signalStrength: 27 },
  { id: 'INV-2024-006', deviceType: 'inverter', model: 'SG110CX', serialNo: 'SN240006B', capacity: 110000, firmware: 'V3.2.1', software: 'V3.2.1', hardware: 'V2.0', stationId: 2, status: 'online', lastOnline: iso(-2), signalStrength: 24 },
  { id: 'INV-2024-007', deviceType: 'inverter', model: 'SG30CX', serialNo: 'SN240007B', capacity: 30000, firmware: 'V3.0.5', software: 'V3.0.5', hardware: 'V1.8', stationId: 2, status: 'online', lastOnline: iso(-4), signalStrength: 19 },
  { id: 'INV-2024-008', deviceType: 'inverter', model: 'SG110CX', serialNo: 'SN240008C', capacity: 110000, firmware: 'V3.2.1', software: 'V3.2.1', hardware: 'V2.0', stationId: 3, status: 'online', lastOnline: iso(-1), signalStrength: 29 },
  { id: 'INV-2024-009', deviceType: 'inverter', model: 'SG50CX', serialNo: 'SN240009C', capacity: 50000, firmware: 'V3.1.0', software: 'V3.1.0', hardware: 'V1.9', stationId: 3, status: 'online', lastOnline: iso(-2), signalStrength: 21 },
  { id: 'INV-2024-010', deviceType: 'inverter', model: 'SG30CX', serialNo: 'SN240010C', capacity: 30000, firmware: 'V3.0.5', software: 'V3.0.5', hardware: 'V1.8', stationId: 3, status: 'offline', lastOnline: iso(-320), signalStrength: 0 },
  { id: 'INV-2024-011', deviceType: 'inverter', model: 'SG30CX', serialNo: 'SN240011D', capacity: 30000, firmware: 'V3.0.5', software: 'V3.0.5', hardware: 'V1.8', stationId: 4, status: 'online', lastOnline: iso(-1), signalStrength: 25 },
  { id: 'INV-2024-012', deviceType: 'inverter', model: 'SG30CX', serialNo: 'SN240012D', capacity: 30000, firmware: 'V3.0.5', software: 'V3.0.5', hardware: 'V1.8', stationId: 4, status: 'online', lastOnline: iso(-3), signalStrength: 23 }
]

const stationName = (id: number) => STATIONS.find((s) => s.id === id)?.name || '-'
const onlineCount = (sid: number) => DEVICES.filter((d) => d.stationId === sid && d.status === 'online').length

/* ---------------- 设备树（StationTree[]，嵌套 devices） ---------------- */
export const DEVICE_TREE = STATIONS.map((s) => ({
  id: s.id,
  name: s.name,
  code: s.code,
  region: s.region,
  deviceCount: s.deviceCount,
  devices: DEVICES.filter((d) => d.stationId === s.id)
}))

/* ---------------- 概览统计 ---------------- */
export const OVERVIEW = {
  totalDevices: DEVICES.length,
  onlineDevices: DEVICES.filter((d) => d.status === 'online').length,
  offlineDevices: DEVICES.filter((d) => d.status === 'offline').length,
  todayEvents: 17,
  totalStations: STATIONS.length,
  unresolvedEvents: 3
}

/* ---------------- 设备状态分布（按台区） ---------------- */
export const DEVICE_STATUS = STATIONS.map((s) => ({
  stationName: s.name,
  online: onlineCount(s.id),
  total: s.deviceCount
}))

/* ---------------- 过电压 Top ---------------- */
export const OVERVOLTAGE_TOP = [
  { deviceId: 'INV-2024-001', stationName: stationName(1), count: 8 },
  { deviceId: 'INV-2024-005', stationName: stationName(2), count: 6 },
  { deviceId: 'INV-2024-008', stationName: stationName(3), count: 5 },
  { deviceId: 'INV-2024-002', stationName: stationName(1), count: 4 },
  { deviceId: 'INV-2024-006', stationName: stationName(2), count: 3 },
  { deviceId: 'INV-2024-003', stationName: stationName(1), count: 2 },
  { deviceId: 'INV-2024-009', stationName: stationName(3), count: 2 },
  { deviceId: 'INV-2024-011', stationName: stationName(4), count: 1 }
]

/* ---------------- 事件列表 ---------------- */
const EVENT_TYPES = ['eov', 'eov_r', 'euv', 'euv_r', 'elc', 'online']
const STATUS_POOL = ['active', 'recovered']
function buildEvents() {
  const list: any[] = []
  const n = 24
  for (let i = 0; i < n; i++) {
    const dev = DEVICES[i % DEVICES.length]
    const et = EVENT_TYPES[i % EVENT_TYPES.length]
    const active = et === 'eov' || et === 'euv'
    list.push({
      occurredAt: today(8 + (i % 12), (i * 7) % 60),
      deviceId: dev.id,
      eventType: et,
      detail: et.startsWith('eov')
        ? { voltageHigh: 263.4 + (i % 5) * 0.3, voltageLow: 220.0, cycle: 10 + (i % 3) }
        : et.startsWith('euv')
        ? { voltageHigh: 220.0, voltageLow: 188.5 + (i % 4) * 0.4, cycle: 10 + (i % 3) }
        : { data: { reason: et === 'online' ? '设备上线' : '本地调控触发' } },
      status: active ? (i % 5 === 0 ? 'active' : 'recovered') : 'recovered'
    })
  }
  // 保证有未恢复事件
  list[0].status = 'active'
  list[1].status = 'active'
  list[2].status = 'active'
  return list
}
export const EVENTS = buildEvents()

/* ---------------- OTA 固件 ---------------- */
export const FIRMWARES = [
  { id: 1, name: 'SG110CX_V3.2.1.bin', version: '3.2.1', fileSize: 1048576, uploadTime: iso(-60 * 24 * 3) },
  { id: 2, name: 'SG50CX_V3.1.0.bin', version: '3.1.0', fileSize: 851968, uploadTime: iso(-60 * 24 * 10) },
  { id: 3, name: 'SG30CX_V3.0.5.bin', version: '3.0.5', fileSize: 712704, uploadTime: iso(-60 * 24 * 21) },
  { id: 4, name: 'SG110CX_V3.2.0.bin', version: '3.2.0', fileSize: 1024000, uploadTime: iso(-60 * 24 * 35) }
]

/* ---------------- OTA 任务（嵌套 firmware） ---------------- */
export const OTA_TASKS = [
  { id: 101, firmware: { id: 1, name: 'SG110CX_V3.2.1.bin' }, status: 'completed', progress: 100, successCount: 4, failCount: 0 },
  { id: 102, firmware: { id: 2, name: 'SG50CX_V3.1.0.bin' }, status: 'running', progress: 67, successCount: 2, failCount: 0 },
  { id: 103, firmware: { id: 3, name: 'SG30CX_V3.0.5.bin' }, status: 'paused', progress: 33, successCount: 1, failCount: 1 },
  { id: 104, firmware: { id: 4, name: 'SG110CX_V3.2.0.bin' }, status: 'pending', progress: 0, successCount: 0, failCount: 0 }
]

/* ---------------- 账号 ---------------- */
export const USERS = [
  { id: 1, username: 'admin', role: 'super_admin', status: 'active', lastLogin: iso(-60 * 2) },
  { id: 2, username: 'ops_wang', role: 'ops', status: 'active', lastLogin: iso(-60 * 26) },
  { id: 3, username: 'ops_li', role: 'ops', status: 'active', lastLogin: iso(-60 * 72) },
  { id: 4, username: 'viewer', role: 'readonly', status: 'active', lastLogin: iso(-60 * 5) },
  { id: 5, username: 'guest', role: 'readonly', status: 'disabled', lastLogin: null }
]

/* ---------------- 设备详情：历史数据 ---------------- */
export const HISTORY = Array.from({ length: 12 }, (_, i) => ({
  timestamp: new Date(now - (11 - i) * 3600000).toISOString(),
  phaseAV: 231.2 + (i % 5) * 0.4,
  phaseBV: 229.8 + (i % 4) * 0.3,
  phaseCV: 230.5 + (i % 6) * 0.2,
  activePower: 8500 + (i % 7) * 120,
  reactivePower: 320 + (i % 5) * 18,
  powerFactor: 0.98 + (i % 3) * 0.005,
  frequency: 50.01 + (i % 4) * 0.01
}))

/* ---------------- 设备详情：MQTT 日志 ---------------- */
export const MQTT_LOGS = Array.from({ length: 8 }, (_, i) => ({
  createdAt: iso(-i * 5),
  direction: i % 2 === 0 ? 'up' : 'down',
  topic: `up/r/INV-2024-001`,
  payload: { deviceId: 'INV-2024-001', ts: iso(-i * 5), ac: { v: [231, 230, 230.5], p: 8500 } }
}))

/* ---------------- 设备详情：操作日志 ---------------- */
export const OP_LOGS = [
  { createdAt: iso(-3), operator: 'admin', actionType: '参数设置', content: '设置过电压保护阈值 265V', result: 'success' },
  { createdAt: iso(-30), operator: 'ops_wang', actionType: '召测', content: '召测 AC/DC/SW/HW', result: 'success' },
  { createdAt: iso(-120), operator: 'ops_li', actionType: '重启', content: '远程重启设备', result: 'success' },
  { createdAt: iso(-240), operator: 'admin', actionType: '固件升级', content: '升级至 V3.2.1', result: 'success' },
  { createdAt: iso(-300), operator: 'ops_wang', actionType: '恢复出厂', content: '恢复出厂设置', result: 'fail' }
]

/* ---------------- 登录 ---------------- */
const DEMO_USER = { id: 0, username: 'demo', role: 'super_admin', status: 'active' }

/* ---------------- 实时数据（详情页实时面板走 WS，此处仅作 /devices/:id/realtime 兜底） ---------------- */
function realtimeFor(devId: string) {
  return {
    ac: { ph: 3, v: [231.2, 229.8, 230.5], c: [12.3, 12.1, 12.4], p: 8500, q: 320, pf: 0.985, f: 50.02 },
    dc: { v: [620, 618, 615], c: [4.5, 4.4, 4.3], p: [2790, 2719, 2644] },
    cs: 1
  }
}

/**
 * 匹配光伏 mock 端点。命中返回业务负载；未命中返回 undefined（adapter 兜底空对象）。
 */
export function matchPvMock(method: string, url: string, params?: any): any {
  const m = (method || 'get').toLowerCase()
  const u = (url || '').split('?')[0]

  if (m === 'get') {
    if (u === '/stats/overview') return OVERVIEW
    if (u === '/stats/device-status') return DEVICE_STATUS
    if (u === '/stats/overvoltage-top') return OVERVOLTAGE_TOP
    if (u === '/devices/tree') return DEVICE_TREE
    if (u === '/devices') return DEVICES
    if (u === '/ota/firmwares') return FIRMWARES
    if (u === '/ota/tasks') return OTA_TASKS
    if (u === '/users') return USERS
    if (u === '/stations') return STATIONS
    if (u === '/groups') return []
    if (u === '/auth/profile') return DEMO_USER

    // /devices/:id/realtime
    let mt = u.match(/^\/devices\/([^/]+)\/realtime$/)
    if (mt) return realtimeFor(mt[1])

    // /data/history
    if (u === '/data/history') return HISTORY

    // /logs/mqtt
    if (u === '/logs/mqtt') return { items: MQTT_LOGS, total: MQTT_LOGS.length }

    // /logs/operations
    if (u === '/logs/operations') return { items: OP_LOGS, total: OP_LOGS.length }

    // /events（分页）
    if (u === '/events') {
      const page = Number(params?.page || 1)
      const pageSize = Number(params?.pageSize || 20)
      const start = (page - 1) * pageSize
      return { items: EVENTS.slice(start, start + pageSize), total: EVENTS.length }
    }

    // /devices/:id （放最后，避免吃掉 /devices/tree 等）
    mt = u.match(/^\/devices\/([^/]+)$/)
    if (mt) return DEVICES.find((d) => d.id === mt![1]) || DEVICES[0]

    // /ota/tasks/:id
    mt = u.match(/^\/ota\/tasks\/([^/]+)$/)
    if (mt) return OTA_TASKS.find((t) => String(t.id) === mt![1]) || OTA_TASKS[0]
  }

  if (m === 'post') {
    if (u === '/auth/login') return { token: 'demo-token-' + Date.now(), user: DEMO_USER }
  }

  // 其余写操作（POST/PUT/PATCH/DELETE）演示态统一成功
  return null
}
