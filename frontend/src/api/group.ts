import request from './pvRequest'

export function getGroups() {
  return request.get('/groups')
}

export function createGroup(data: any) {
  return request.post('/groups', data)
}

export function updateGroup(id: number, data: any) {
  return request.put(`/groups/${id}`, data)
}

export function deleteGroup(id: number) {
  return request.delete(`/groups/${id}`)
}
