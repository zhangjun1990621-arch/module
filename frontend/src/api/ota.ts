import request from './request'

/**
 * OTA 升级 API
 * 使用通用 request 实例（走真实后端），路径前缀 /pv/ota
 */
const BASE = '/pv/ota'

// ================ 固件管理 ================

export function getFirmwares() {
  return request.get(`${BASE}/firmwares`)
}

export function uploadFirmware(formData: FormData) {
  return request.post(`${BASE}/firmwares`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function deleteFirmware(id: number) {
  return request.delete(`${BASE}/firmwares/${id}`)
}

// ================ 升级任务 ================

export function getOTATasks() {
  return request.get(`${BASE}/tasks`)
}

export function getOTATask(id: number) {
  return request.get(`${BASE}/tasks/${id}`)
}

export function createOTATask(data: { firmwareId: number; deviceIds: string[] }) {
  return request.post(`${BASE}/tasks`, data)
}

export function deleteOTATask(id: number) {
  return request.delete(`${BASE}/tasks/${id}`)
}

export function pauseOTATask(id: number) {
  return request.post(`${BASE}/tasks/${id}/pause`)
}

export function resumeOTATask(id: number) {
  return request.post(`${BASE}/tasks/${id}/resume`)
}

export function cancelOTATask(id: number) {
  return request.post(`${BASE}/tasks/${id}/cancel`)
}

/** 手动结束任务（进度停在当前位置） */
export function completeOTATask(id: number) {
  return request.post(`${BASE}/tasks/${id}/complete`)
}

/** 重试失败设备 */
export function retryFailedDevices(id: number) {
  return request.post(`${BASE}/tasks/${id}/retry`)
}
