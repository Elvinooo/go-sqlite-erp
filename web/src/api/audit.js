import request from './request'

export function listOperationLogs(params) {
  return request.get('/audit/operation-logs', { params })
}

export function listLoginLogs(params) {
  return request.get('/audit/login-logs', { params })
}
