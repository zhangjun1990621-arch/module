import request from './pvRequest'

export function getFirmwares() {
  return request.get('/ota/firmwares')
}

export function uploadFirmware(formData: FormData) {
  return request.post('/ota/firmwares', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function getOTATasks() {
  return request.get('/ota/tasks')
}

export function getOTATask(id: number) {
  return request.get(`/ota/tasks/${id}`)
}

export function createOTATask(data: { firmwareId: number; deviceIds: string[] }) {
  return request.post('/ota/tasks', data)
}

export function pauseOTATask(id: number) {
  return request.post(`/ota/tasks/${id}/pause`)
}

export function resumeOTATask(id: number) {
  return request.post(`/ota/tasks/${id}/resume`)
}

export function cancelOTATask(id: number) {
  return request.post(`/ota/tasks/${id}/cancel`)
}
