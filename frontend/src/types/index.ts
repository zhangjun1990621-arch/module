export interface UserInfo {
  id: number
  username: string
  role: string
  status: string
}

export interface Device {
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

export interface Station {
  id: number
  name: string
  code: string
  region: string
  deviceCount: number
}

export interface Group {
  id: number
  name: string
  parentId: number | null
  region: string
  children?: Group[]
}

export interface Event {
  id: number
  deviceId: string
  eventType: string
  detail: Record<string, any>
  severity: string
  status: string
  occurredAt: string
  recoveredAt: string | null
}

export interface Firmware {
  id: number
  name: string
  version: string
  filePath: string
  fileSize: number
  md5: string
  deviceType: string
  uploadTime: string
}

export interface UpgradeTask {
  id: number
  firmwareId: number
  firmware?: Firmware
  status: string
  totalDevices: number
  successCount: number
  failCount: number
  progress: number
  createdBy: string
  createdAt: string
}

export interface Overview {
  totalDevices: number
  onlineDevices: number
  offlineDevices: number
  todayEvents: number
  totalStations: number
  unresolvedEvents: number
}
