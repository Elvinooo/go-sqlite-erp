import request from './request'

export function listCustomers(params) {
  return request.get('/customers', { params })
}

export function getCustomer(id) {
  return request.get(`/customers/${id}`)
}

export function createCustomer(data) {
  return request.post('/customers', data)
}

export function updateCustomer(id, data) {
  return request.put(`/customers/${id}`, data)
}

export function deleteCustomer(id) {
  return request.delete(`/customers/${id}`)
}

export function importCustomers(file) {
  const form = new FormData()
  form.append('file', file)
  return request.post('/customers/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function exportCustomers(params) {
  return request.get('/customers/export', { params, responseType: 'blob' })
}

export function getCustomerDebt(id) {
  return request.get(`/customers/${id}/debt`)
}

export function listCustomerOrders(id, params) {
  return request.get(`/customers/${id}/orders`, { params })
}
