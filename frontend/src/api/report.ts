import client from './client'

export const getReportApi = (type: string, params?: any) => client.get(`/reports/${type}`, { params })
