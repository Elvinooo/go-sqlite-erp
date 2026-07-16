import request from './request'

export function login(data) {
  return request.post('/auth/login', data)
}

export function getCurrentUser() {
  return request.get('/auth/me')
}

export function logout() {
  return request.post('/auth/logout')
}

export function changePassword(data) {
  return request.put('/auth/password', data)
}

export function updateProfile(data) {
  return request.put('/auth/profile', data)
}
