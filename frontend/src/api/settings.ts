import client from './client'

export const getCompanyProfileApi = () => client.get('/company')
export const updateCompanyProfileApi = (data: any) => client.put('/company', data)
export const getSettingApi = (key: string) => client.get(`/settings/${key}`)
export const saveSettingApi = (data: any) => client.post('/settings', data)
export const listSettingsApi = (category?: string) => client.get('/settings', { params: { category } })

export const importCarisApi = (formData: FormData) => client.post('/settings/import/caris', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
export const importProductsApi = (formData: FormData) => client.post('/settings/import/products', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
export const downloadSampleCariApi = () => client.get('/settings/import/sample/caris', { responseType: 'blob' })
export const downloadSampleProductApi = () => client.get('/settings/import/sample/products', { responseType: 'blob' })
export const updateEnabledModulesApi = (enabledModules: string[]) => client.put('/settings/enabled-modules', { enabled_modules: enabledModules })
