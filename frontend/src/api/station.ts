import request from './pvRequest'

export function getStations() {
  return request.get('/stations')
}

export function createStation(data: any) {
  return request.post('/stations', data)
}

export function updateStation(id: number, data: any) {
  return request.put(`/stations/${id}`, data)
}

export function deleteStation(id: number) {
  return request.delete(`/stations/${id}`)
}
