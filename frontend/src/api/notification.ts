import client from './client'

export const getNotificationsApi = (params?: { page?: number; limit?: number }) => {
  return client.get('/notifications', { params })
}

export const markAsReadApi = (id: string) => {
  return client.post(`/notifications/${id}/read`)
}

export const markAllAsReadApi = () => {
  return client.post('/notifications/read-all')
}
