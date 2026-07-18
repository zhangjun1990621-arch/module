import request from './pvRequest'

export function getOperationLogs(params?: any) {
  return request.get('/logs/operations', { params })
}

export function getMQTTLogs(params?: any) {
  return request.get('/logs/mqtt', { params })
}
