/**
 * 铝厂云控平台 —— 演示数据层（前端融合阶段）
 * ----------------------------------------------------------------
 * 本文件用「结构化 mock 数据」还原原铝厂项目的核心监控逻辑：
 *   1) 每台电解槽沿 13 个标准位置布点测温；
 *   2) 依据「当前温度 - 昨日平均/最高」做升温趋势分级（原 DataProcessor.ErrProcess）；
 *   3) 升温超 ContrastBase 基准判定破损高危（原 T_DevicePoint.Damaged）；
 *   4) 钢棒位置标记钢棒切削（原 T_DevicePoint.Cut）；
 *   5) 绝对温度阈值做三级阈值报警（原 ErrorBaseValue + ErrorValue3/2/1）；
 *   6) 汇总生成报警记录与历史读数。
 *
 * ⚠️ 后续统一到光伏 MQTT 规约（up/r/{deviceId}）后，只需把 aluminumApi 各方法
 *    替换为对后端 /ws 实时消息或 REST 接口的调用，视图层无需改动。
 */
import type {
  Cell,
  CellPoint,
  CellStatus,
  AlarmRecord,
  HistoryRow,
  OverviewStat,
  PointStatus,
  AlarmLevel,
  AlarmHandleStatus,
  PositionDef,
  TempMatrixCell
} from './types'

/* ============================== 基础常量 ============================== */

/** 电解槽 13 个标准测温位置（A钢棒 → B钢棒） */
const POSITIONS: PositionDef[] = [
  { code: 'A', name: 'A钢棒' },
  { code: 'AC', name: 'A侧壁' },
  { code: 'AC2', name: 'A侧壁2' },
  { code: 'AC3', name: 'A侧壁3' },
  { code: 'CL', name: '出铝' },
  { code: 'AD', name: 'A槽底' },
  { code: 'CD', name: '槽底' },
  { code: 'BD', name: 'B槽底' },
  { code: 'YD', name: '烟道' },
  { code: 'BC3', name: 'B侧壁3' },
  { code: 'BC2', name: 'B侧壁2' },
  { code: 'BC', name: 'B侧壁' },
  { code: 'B', name: 'B钢棒' }
]

const CELL_TOTAL = 24
const STEEL_POSITIONS = new Set(['A', 'B']) // 钢棒位置才可能标记「钢棒切削」

/* ============================== 可复现随机 ============================== */
function makeRng(seed: number) {
  let s = seed >>> 0
  return () => {
    s = (s * 1664525 + 1013904223) >>> 0
    return s / 4294967296
  }
}
const rng = makeRng(20250706)
const rand = (min: number, max: number) => min + rng() * (max - min)
const randInt = (min: number, max: number) => Math.floor(rand(min, max + 1))

function fmtTime(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(
    d.getMinutes()
  )}:${p(d.getSeconds())}`
}

/* ============================== 诊断辅助 ============================== */
function calcWarming(temp: number, yAvg: number, yMax: number): 0 | 1 | 2 | 3 {
  const rise = temp - yAvg
  const riseMax = temp - yMax
  if (rise > 20 || riseMax > 20) return 3
  if (rise > 12) return 2
  if (rise > 5) return 1
  return 0
}
function calcErrorLevel(temp: number): 0 | 1 | 2 | 3 {
  if (temp >= 975) return 3
  if (temp >= 965) return 2
  if (temp >= 955) return 1
  return 0
}

/* ============================== 数据生成 ============================== */
interface Built {
  cells: Cell[]
  points: CellPoint[]
  alarms: AlarmRecord[]
  history: HistoryRow[]
}

function buildAll(): Built {
  const cells: Cell[] = []
  const points: CellPoint[] = []
  const now = new Date()

  for (let i = 0; i < CELL_TOTAL; i++) {
    const id = 5001 + i
    const name = `${id}#`
    const factory = i < 12 ? '厂区A' : '厂区B'
    const wzIdx = (i % 4) + 1
    const workzone = `工区${wzIdx}`
    const hostPrefix = i < 12 ? 'A' : 'B'
    const host = `H-${hostPrefix}${wzIdx}`
    const isOfflineCell = i === 13 // 演示：1 台离线（对应 demo 的离线槽）
    // 升温热度：normal 0~6，warn 10~18，alarm 28~40
    let heat = rand(0, 6)
    if (i === 4 || i === 10 || i === 18) heat = rand(10, 18)
    if (i === 5) heat = rand(28, 40)

    const cellPoints: CellPoint[] = []
    for (let pi = 0; pi < POSITIONS.length; pi++) {
      const pos = POSITIONS[pi]
      const suffix = (i % 3) + 1
      const pointName = `${pos.name}${suffix}`
      const pid = `${id}-${pos.code}${suffix}`

      const yAvg = 938 + pi * 0.4 + rand(-2, 2)
      const yMax = yAvg + rand(8, 15)

      let temp: number
      let status: PointStatus
      let volt = +rand(3.9, 4.6).toFixed(2)
      let current = STEEL_POSITIONS.has(pos.code) ? +rand(180, 360).toFixed(1) : +rand(0, 40).toFixed(1)

      if (isOfflineCell) {
        temp = 0
        status = 'offline'
      } else {
        status = 'online'
        temp = +(yAvg + heat + rand(-1.5, 1.5)).toFixed(1)
      }

      const warming = (status === 'online' ? calcWarming(temp, yAvg, yMax) : 0) as 0 | 1 | 2 | 3
      const errorLevel = (status === 'online' ? calcErrorLevel(temp) : 0) as 0 | 1 | 2 | 3
      const rise = temp - yAvg
      const damaged = status === 'online' && warming >= 2 && rise > 25
      const cut = STEEL_POSITIONS.has(pos.code) && status === 'online' && rng() < 0.08

      cellPoints.push({
        id: pid,
        cellId: id,
        cellName: name,
        position: pos.code,
        positionName: pos.name,
        name: pointName,
        temp,
        volt,
        current,
        yesterdayMax: +yMax.toFixed(1),
        yesterdayAvg: +yAvg.toFixed(1),
        warmingLevel: warming,
        errorLevel,
        damaged,
        cut,
        status,
        saveTime: fmtTime(now)
      })
    }

    points.push(...cellPoints)

    // 聚合电解槽状态
    const onlinePts = cellPoints.filter((p) => p.status === 'online')
    const temps = onlinePts.map((p) => p.temp)
    const avgTemp = temps.length ? +(temps.reduce((a, b) => a + b, 0) / temps.length).toFixed(1) : 0
    const maxTemp = temps.length ? Math.max(...temps) : 0
    const warmingLevel = Math.max(0, ...cellPoints.map((p) => p.warmingLevel)) as 0 | 1 | 2 | 3
    const damagedCount = cellPoints.filter((p) => p.damaged).length
    const cutCount = cellPoints.filter((p) => p.cut).length
    const alarmCount = cellPoints.filter(
      (p) => p.errorLevel >= 1 || p.warmingLevel >= 1 || p.damaged || p.cut
    ).length

    let status: CellStatus
    if (isOfflineCell) status = 'offline'
    else if (cellPoints.some((p) => p.errorLevel === 3 || p.damaged)) status = 'alarm'
    else if (cellPoints.some((p) => p.warmingLevel >= 2 || p.errorLevel >= 2)) status = 'warn'
    else status = 'normal'

    cells.push({
      id,
      name,
      factory,
      workzone,
      host,
      online: !isOfflineCell,
      status,
      avgTemp,
      maxTemp,
      pointCount: cellPoints.length,
      onlinePointCount: onlinePts.length,
      warmingLevel,
      damagedCount,
      cutCount,
      alarmCount
    })
  }

  // 报警记录：遍历所有「有意义」的点位
  const alarms: AlarmRecord[] = []
  let alarmIdx = 0
  const handleStatusPool: AlarmHandleStatus[] = ['未处理', '处理中', '已恢复']
  for (const p of points) {
    if (p.status !== 'online') continue
    const rise = +(p.temp - p.yesterdayAvg).toFixed(1)
    let rec: Omit<AlarmRecord, 'id' | 'time'> | null = null

    if (p.damaged) {
      rec = { cellId: p.cellId, cellName: p.cellName, pointName: p.name, level: 3, type: '破损高危', detail: `升温 ${rise}℃`, status: '未处理' }
    } else if (p.cut) {
      rec = { cellId: p.cellId, cellName: p.cellName, pointName: p.name, level: 2, type: '钢棒切削', detail: '钢棒切削异常', status: '未处理' }
    } else if (p.errorLevel >= 1) {
      rec = { cellId: p.cellId, cellName: p.cellName, pointName: p.name, level: p.errorLevel as AlarmLevel, type: '槽温超限', detail: `${p.temp}℃`, status: handleStatusPool[randInt(0, 2)] }
    } else if (p.warmingLevel >= 1) {
      rec = { cellId: p.cellId, cellName: p.cellName, pointName: p.name, level: p.warmingLevel as AlarmLevel, type: '升温趋势', detail: `升温 ${rise}℃`, status: handleStatusPool[randInt(0, 2)] }
    }

    if (rec) {
      alarms.push({
        id: `AL${String(++alarmIdx).padStart(4, '0')}`,
        time: fmtTime(new Date(now.getTime() - randInt(0, 3600 * 6) * 1000)),
        ...rec
      })
    }
  }
  // 确保各级别都有演示数据
  const ensureLevel = (lv: AlarmLevel) => {
    if (!alarms.some((a) => a.level === lv)) {
      const p = points.find((x) => x.status === 'online')
      if (p) {
        alarms.unshift({
          id: `AL${String(++alarmIdx).padStart(4, '0')}`,
          time: fmtTime(now),
          cellId: p.cellId,
          cellName: p.cellName,
          pointName: p.name,
          level: lv,
          type: '演示报警',
          detail: `示例 ${lv} 级`,
          status: '未处理'
        })
      }
    }
  }
  ensureLevel(1)
  ensureLevel(2)
  ensureLevel(3)

  // 历史读数：选取 8 台电解槽，各生成近 12 小时、每小时一条代表性点位读数
  const history: HistoryRow[] = []
  const sampleCells = [0, 1, 2, 4, 5, 8, 10, 13].map((i) => cells[i]).filter(Boolean)
  for (const c of sampleCells) {
    const repr = points.find((p) => p.cellId === c.id && p.status === 'online')
    if (!repr) continue
    for (let h = 12; h >= 0; h--) {
      const t = new Date(now.getTime() - h * 3600 * 1000)
      const base = repr.yesterdayAvg
      const temp = +(base + (c.status === 'alarm' ? rand(20, 38) : c.status === 'warn' ? rand(8, 16) : rand(-3, 5))).toFixed(1)
      history.push({
        time: fmtTime(t),
        cellId: c.id,
        cellName: c.name,
        pointName: repr.name,
        positionName: repr.positionName,
        temp,
        volt: +rand(3.9, 4.6).toFixed(2),
        current: repr.current
      })
    }
  }

  return { cells, points, alarms, history }
}

const DATA = buildAll()

/* ============================== 对外 API（模拟异步，后续替换为真实接口/MQTT） ============================== */
function delay<T>(data: T, ms = 120): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(data), ms))
}

export const aluminumApi = {
  /** 厂区列表 */
  getFactories(): Promise<string[]> {
    return delay(Array.from(new Set(DATA.cells.map((c) => c.factory))))
  },
  /** 工区列表（按厂区筛选） */
  getWorkzones(factory?: string): Promise<string[]> {
    const list = factory ? DATA.cells.filter((c) => c.factory === factory) : DATA.cells
    return delay(Array.from(new Set(list.map((c) => c.workzone))))
  },
  /** 设备主机列表（按厂区+工区筛选） */
  getHosts(factory?: string, workzone?: string): Promise<string[]> {
    const list = DATA.cells.filter(
      (c) => (!factory || c.factory === factory) && (!workzone || c.workzone === workzone)
    )
    return delay(Array.from(new Set(list.map((c) => c.host))))
  },
  /** 概览统计 */
  getOverview(): Promise<OverviewStat> {
    const temps = DATA.cells.filter((c) => c.online).map((c) => c.avgTemp)
    const avgTemp = temps.length ? +(temps.reduce((a, b) => a + b, 0) / temps.length).toFixed(1) : 0
    const level1 = DATA.alarms.filter((a) => a.level === 1).length
    const level2 = DATA.alarms.filter((a) => a.level === 2).length
    const level3 = DATA.alarms.filter((a) => a.level === 3).length
    return delay({
      cellCount: DATA.cells.length,
      onlineCount: DATA.cells.filter((c) => c.online).length,
      offlineCount: DATA.cells.filter((c) => !c.online).length,
      alarmCount: DATA.alarms.length,
      level1,
      level2,
      level3,
      avgTemp,
      damagedCount: DATA.points.filter((p) => p.damaged).length,
      cutCount: DATA.points.filter((p) => p.cut).length
    })
  },
  /** 电解槽列表 */
  getCells(): Promise<Cell[]> {
    return delay(DATA.cells)
  },
  /** 点位实时数据（可按电解槽过滤） */
  getPoints(cellId?: number): Promise<CellPoint[]> {
    const list = cellId ? DATA.points.filter((p) => p.cellId === cellId) : DATA.points
    return delay(list)
  },
  /** 报警记录 */
  getAlarms(): Promise<AlarmRecord[]> {
    return delay(DATA.alarms)
  },
  /** 历史数据 */
  getHistory(): Promise<HistoryRow[]> {
    return delay(DATA.history)
  },
  /** 获取单槽温度矩阵 (4行×28列)，每次调用生成新的随机波动 */
  getCellTempMatrix(cellId: number): Promise<TempMatrixCell[][]> {
    const cell = DATA.cells.find((c) => c.id === cellId)
    const rows = ['A1', 'A2', 'B1', 'B2']
    const cols = 28
    const matrix: TempMatrixCell[][] = []
    // 基准温度：取该槽平均温度，离线槽用 0
    const base = cell && cell.online ? cell.avgTemp : 0
    // 每行有不同的偏移（A1偏钢棒侧温度偏高，B2偏低）
    const rowOffsets: Record<string, number> = { A1: 3, A2: 1, B1: -1, B2: -3 }
    for (const row of rows) {
      const arr: TempMatrixCell[] = []
      for (let col = 1; col <= cols; col++) {
        if (!cell || !cell.online) {
          arr.push({ row, col, temp: 0, errorLevel: 0 })
          continue
        }
        // 随机波动 ±3℃
        const noise = (Math.random() - 0.5) * 6
        const temp = +(base + rowOffsets[row] + noise).toFixed(0)
        let errorLevel: 0 | 1 | 2 | 3 = 0
        if (temp >= 975) errorLevel = 3
        else if (temp >= 965) errorLevel = 2
        else if (temp >= 955) errorLevel = 1
        arr.push({ row, col, temp, errorLevel })
      }
      matrix.push(arr)
    }
    return delay(matrix)
  }
}

export { POSITIONS }
