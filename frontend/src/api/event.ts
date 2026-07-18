import request from './pvRequest'

export function getEvents(params?: any) {
  return request.get('/events', { params })
}

export function getEventStats() {
  return request.get('/events/stats')
}
