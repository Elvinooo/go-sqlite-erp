import request from './request'

export function listBusiness(module, params) {
  return request.get(`/${module}`, { params })
}

export function getBusiness(module, id) {
  return request.get(`/${module}/${id}`)
}

export function createBusiness(module, data) {
  return request.post(`/${module}`, data)
}

export function updateBusiness(module, id, data) {
  return request.put(`/${module}/${id}`, data)
}

export function deleteBusiness(module, id, data) {
  return request.delete(`/${module}/${id}`, data ? { data } : undefined)
}

export function runBusinessAction(module, action, data = {}) {
  return request.post(`/${module}/actions`, { action, data })
}

export function listBusinessPhotos(module, id) {
  return request.get(`/${module}/${id}/photos`)
}

export function uploadBusinessPhoto(module, id, file, scene = 'general') {
  const form = new FormData()
  form.append('file', file)
  form.append('scene', scene)
  return request.post(`/${module}/${id}/photos`, form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
