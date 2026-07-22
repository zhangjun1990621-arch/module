import request from './request'

// 获取当前平台 ID（从 URL 路径提取）
function platformPrefix(): string {
  // 从 URL 路径 /pv/xxx 或 /aluminum/xxx 中提取平台 ID
  const m = window.location.pathname.match(/^\/([a-z]+)\//i)
  return m ? m[1] : 'pv'
}

export function getDevices(params?: any) {
  return request.get(`/${platformPrefix()}/devices`, { params })
}

export function getDeviceTree() {
  return request.get(`/${platformPrefix()}/devices/tree`)
}

export function getDevice(id: string) {
  return request.get(`/${platformPrefix()}/devices/${id}`)
}

export function updateDeviceGroup(id: string, groupId: number) {
  return request.patch(`/${platformPrefix()}/devices/${id}/group`, { groupId })
}

export function updateDeviceStation(id: string, stationId: number | null) {
  return request.patch(`/${platformPrefix()}/devices/${id}/station`, { stationId })
}

export function pollDevice(id: string, items: string[]) {
  return request.post(`/${platformPrefix()}/devices/${id}/polling`, { items })
}

export function setDevice(id: string, params: Record<string, any>) {
  return request.post(`/${platformPrefix()}/devices/${id}/set`, params)
}

export function actionDevice(id: string, action: string, value?: any) {
  return request.post(`/${platformPrefix()}/devices/${id}/action`, { action, value })
}

export function getRealtimeData(id: string) {
  return request.get(`/${platformPrefix()}/devices/${id}/realtime`)
}
