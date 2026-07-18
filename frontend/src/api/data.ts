import request from './pvRequest'

export function getHistoryData(params: { deviceId: string; start: string; end: string }) {
  return request.get('/data/history', { params })
}

export function exportCSV(params: { deviceId: string; start: string; end: string }) {
  return request.get('/data/export', { params, responseType: 'blob' })
}
