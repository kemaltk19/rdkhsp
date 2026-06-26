import client from './client'

export const getAuditLogsApi = (params?: { module?: string; record_id?: string }) =>
  client.get('/audit-logs', { params })
