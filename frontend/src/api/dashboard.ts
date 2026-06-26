import client from './client'

export const getDashboardStatsApi = () => client.get('/dashboard')
