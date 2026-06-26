import client from './client'

export const getCompaniesApi = () => {
  return client.get('/superadmin/companies')
}

export const getDashboardStatsApi = () => {
  return client.get('/superadmin/dashboard')
}

export const createCompanyApi = (data: any) => {
  return client.post('/superadmin/companies', data)
}

export const updateCompanyApi = (id: string, data: any) => {
  return client.put(`/superadmin/companies/${id}`, data)
}

export const deleteCompanyApi = (id: string) => {
  return client.delete(`/superadmin/companies/${id}`)
}

export const toggleCompanyStatusApi = (id: string, action: 'suspend' | 'activate') => {
  return client.put(`/superadmin/companies/${id}/status`, { action })
}

export const getPlansApi = () => {
  return client.get('/superadmin/plans')
}

export const createPlanApi = (data: any) => {
  return client.post('/superadmin/plans', data)
}

export const updatePlanApi = (id: string, data: any) => {
  return client.put(`/superadmin/plans/${id}`, data)
}

export const deletePlanApi = (id: string) => {
  return client.delete(`/superadmin/plans/${id}`)
}

// Email (SMTP) settings — platform-wide
export const getEmailSettingsApi = () => {
  return client.get('/superadmin/email-settings')
}

export const updateEmailSettingsApi = (data: any) => {
  return client.put('/superadmin/email-settings', data)
}

export const testEmailApi = (to: string) => {
  return client.post('/superadmin/email-settings/test', { to })
}
