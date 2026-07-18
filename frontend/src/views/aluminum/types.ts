/**
 * 铝厂云控平台 —— 数据模型（前端融合阶段）
 * ----------------------------------------------------------------
 * 以下结构尽量贴近原铝厂项目的业务逻辑（电解槽智能无线测温系统），
 * 字段命名对应原 .NET 后端实体与实时接口，便于后续统一到光伏 MQTT 规约后直接替换数据源。
 *
 * 原铝厂关键实体对照：
 *   T_VirtualDevice   -> Cell        （电解槽，逻辑设备）
 *   T_VirtualPoint    -> CellPoint   （电解槽下的点位，由物理点位 T_DevicePoint 归属而来）
 *   T_DevicePoint     -> CellPoint   （物理传感器，带 Position / Damaged / Cut / 报警字段）
 *   LiveDataForVirtualPoint -> CellPoint 的实时值（Value 温度 / Volt 电压 / Value2 电流）
 *   T_PointError      -> AlarmRecord （报警记录）
 *   T_DeviceData      -> HistoryRow  （历史读数，按天分表）
 */

/** 电解槽整体状态（用于概览与电解槽监测表） */
export type CellStatus = 'normal' | 'warn' | 'alarm' | 'offline'

/** 点位通信状态 */
export type PointStatus = 'online' | 'offline' | 'NC'

/** 报警等级：1 一级 / 2 二级 / 3 三级（原 ErrorValue3/2/1 分级） */
export type AlarmLevel = 1 | 2 | 3

/** 报警处理状态 */
export type AlarmHandleStatus = '未处理' | '处理中' | '已恢复'

/** 标准测温位置（原铝厂 13 个标准位置，A钢棒 → B钢棒） */
export interface PositionDef {
  code: string
  name: string
}

/** 电解槽下挂的一个测温点位（对应原 T_VirtualPoint / T_DevicePoint） */
export interface CellPoint {
  id: string // 点位ID，如 5001-AC1
  cellId: number // 所属电解槽号
  cellName: string // 所属电解槽名称，如 5001#
  position: string // 位置编号 A/AC/AC2/...
  positionName: string // 位置名称 A钢棒/...
  name: string // 点位名称，如 A侧壁1
  temp: number // 温度 ℃（原 Value）
  volt: number // 电压 V（原 Volt，F6 解析 /10）
  current: number // 电流 A（原 Value2，电流分布/新电流）
  yesterdayMax: number // 昨日最高（升温趋势基准，原 Max）
  yesterdayAvg: number // 昨日平均（升温趋势基准，原 Avg）
  warmingLevel: 0 | 1 | 2 | 3 // 升温趋势等级（平均/最大升温趋势分级）
  errorLevel: 0 | 1 | 2 | 3 // 三级阈值报警等级（原 ErrorBaseValue + ErrorValue3/2/1）
  damaged: boolean // 破损高危（原 T_DevicePoint.Damaged，升温超 ContrastBase 基准）
  cut: boolean // 钢棒切削（原 T_DevicePoint.Cut）
  status: PointStatus
  saveTime: string // 最近采集时间
}

/** 电解槽（对应原 T_VirtualDevice） */
export interface Cell {
  id: number
  name: string // 如 5001#
  factory: string // 厂区，如 厂区A
  workzone: string // 工区，如 工区1
  host: string // 设备主机，如 H-A1
  online: boolean
  status: CellStatus
  avgTemp: number // 平均槽温
  maxTemp: number // 最高点位温度
  pointCount: number // 点位总数
  onlinePointCount: number
  warmingLevel: 0 | 1 | 2 | 3 // 槽体最高升温趋势等级
  damagedCount: number // 破损高危点位数量
  cutCount: number // 钢棒切削点位数量
  alarmCount: number // 活跃报警数
}

/** 报警记录（对应原 T_PointError） */
export interface AlarmRecord {
  id: string
  time: string
  cellId: number
  cellName: string
  pointName: string
  level: AlarmLevel
  type: string // 槽温超限 / 升温趋势 / 破损高危 / 钢棒切削
  detail: string // 详情，如 968℃ / 升温 +32℃ / 破损高危
  status: AlarmHandleStatus
}

/** 历史读数（对应原 T_DeviceData，按天分表） */
export interface HistoryRow {
  time: string
  cellId: number
  cellName: string
  pointName: string
  positionName: string
  temp: number
  volt: number
  current: number
}

/** 概览统计 */
export interface OverviewStat {
  cellCount: number
  onlineCount: number
  offlineCount: number
  alarmCount: number
  level1: number
  level2: number
  level3: number
  avgTemp: number
  damagedCount: number
  cutCount: number
}

/** 单槽温度矩阵中的一个格子 */
export interface TempMatrixCell {
  row: string // 'A1' | 'A2' | 'B1' | 'B2'
  col: number // 1~28
  temp: number // 温度值
  errorLevel: 0 | 1 | 2 | 3
}
