import request from './request'

export function restoreTestData() {
  return request.post('/settings/restore-test-data')
}

export function getMerchantInfo() {
  return request.get('/settings/merchant-info')
}

export function updateMerchantInfo(data) {
  return request.put('/settings/merchant-info', data)
}
