/**
 * 铝厂云控平台 —— 状态/等级展示映射
 * 统一把枚举值映射为 .al-tag 的配色类（green/red/orange/blue/gray），
 * 供各表格视图复用，避免重复逻辑。
 */
import type { CellStatus, PointStatus, AlarmHandleStatus } from './types'

export function cellStatusMeta(s: CellStatus): { label: string; cls: string } {
  switch (s) {
    case 'alarm':
      return { label: '告警', cls: 'red' }
    case 'warn':
      return { label: '预警', cls: 'orange' }
    case 'offline':
      return { label: '离线', cls: 'gray' }
    default:
      return { label: '正常', cls: 'green' }
  }
}

export function pointStatusMeta(s: PointStatus): { label: string; cls: string } {
  switch (s) {
    case 'offline':
      return { label: '离线', cls: 'gray' }
    case 'NC':
      return { label: '异常', cls: 'orange' }
    default:
      return { label: '在线', cls: 'green' }
  }
}

/** 等级 → 标签：1 一级 / 2 二级 / 3 三级 */
export function levelMeta(lv: number): { label: string; cls: string } {
  switch (lv) {
    case 3:
      return { label: '三级', cls: 'red' }
    case 2:
      return { label: '二级', cls: 'orange' }
    case 1:
      return { label: '一级', cls: 'blue' }
    default:
      return { label: '—', cls: 'gray' }
  }
}

export function alarmStatusMeta(s: AlarmHandleStatus): { label: string; cls: string } {
  switch (s) {
    case '未处理':
      return { label: '未处理', cls: 'red' }
    case '处理中':
      return { label: '处理中', cls: 'orange' }
    default:
      return { label: '已恢复', cls: 'green' }
  }
}

/** 温度配色：越高越红（槽温越高越危险） */
export function tempMeta(t: number): string {
  if (t >= 975) return 'color:#f6565c;font-weight:700'
  if (t >= 965) return 'color:#f0a030;font-weight:600'
  if (t >= 955) return 'color:#e6c34a'
  if (t <= 0) return 'color:#8d9db8'
  return 'color:#e6ecf5'
}
