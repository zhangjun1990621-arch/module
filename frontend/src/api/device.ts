import request from './pvRequest'

export function getDevices(params?: any) {
  return request.get('/devices', { params })
}

export function getDeviceTree() {
  return request.get('/devices/tree')
}

export function getDevice(id: string) {
  return request.get(`/devices/${id}`)
}

export function updateDeviceGroup(id: string, groupId: number) {
  return request.patch(`/devices/${id}/group`, { groupId })
}

export function updateDeviceStation(id: string, stationId: number | null) {
  return request.patch(`/devices/${id}/station`, { stationId })
}

export function pollDevice(id: string, items: string[]) {
  return request.post(`/devices/${id}/polling`, { items })
}

export function setDevice(id: string, params: Record<string, any>) {
  return request.post(`/devices/${id}/set`, params)
}

export function actionDevice(id: string, action: string, value?: any) {
  return request.post(`/devices/${id}/action`, { action, value })
}

export function getRealtimeData(id: string) {
  return request.get(`/devices/${id}/realtime`)
}
