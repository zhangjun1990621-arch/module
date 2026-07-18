export const ROLES = {
  SUPER_ADMIN: 'super_admin',
  ADMIN: 'admin',
  OPS: 'ops',
  READONLY: 'readonly'
} as const

export const ROLE_LABELS: Record<string, string> = {
  super_admin: '超级管理员',
  admin: '管理员',
  ops: '运维人员',
  ops_aluminum: '铝厂运维',
  ops_pv: '光伏运维',
  ops_powerplant: '电厂运维',
  readonly: '只读用户'
}

export const EVENT_TYPES = {
  EOV: 'eov',
  EOVR: 'eov_r',
  EUV: 'euv',
  EUVR: 'euv_r',
  ELC: 'elc',
  ONLINE: 'online'
} as const

export const EVENT_TYPE_LABELS: Record<string, string> = {
  eov: '过电压',
  eov_r: '过电压恢复',
  euv: '低电压',
  euv_r: '低电压恢复',
  elc: '本地调控',
  online: '上线通知'
}

export const UPGRADE_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  PAUSED: 'paused',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled'
} as const

export const UPGRADE_STATUS_LABELS: Record<string, string> = {
  pending: '待执行',
  running: '执行中',
  paused: '已暂停',
  completed: '已完成',
  cancelled: '已取消'
}
