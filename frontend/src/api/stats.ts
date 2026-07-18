import request from './pvRequest'

export function getOverview() {
  return request.get('/stats/overview')
}

export function getDeviceStatus() {
  return request.get('/stats/device-status')
}

export function getOvervoltageTop() {
  return request.get('/stats/overvoltage-top')
}

export function getOvervoltageTrend() {
  return request.get('/stats/overvoltage-trend')
}
