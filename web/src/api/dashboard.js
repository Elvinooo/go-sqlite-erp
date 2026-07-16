import request from './request'

export function getBossDashboard() {
  return request.get('/dashboard/boss')
}

export function getPriceProfitAnalysis(days = 30) {
  return request.get('/dashboard/price-profit', { params: { days } })
}

export function signInDashboard(data) {
  return request.post('/dashboard/sign-in', data)
}

export function listSignInHistory(params) {
  return request.get('/dashboard/sign-ins', { params })
}
