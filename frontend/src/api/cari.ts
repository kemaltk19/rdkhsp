import client from './client'

export const getCarisApi = (params?: any) => client.get('/caris', { params })
export const getCariSummaryApi = () => client.get('/caris/summary')
export const getCariByIDApi = (id: string) => client.get(`/caris/${id}`)
export const createCariApi = (data: any) => client.post('/caris', data)
export const updateCariApi = (id: string, data: any) => client.put(`/caris/${id}`, data)
export const deleteCariApi = (id: string) => client.delete(`/caris/${id}`)
export const getNextCariCodeApi = () => client.get('/caris/next-code')
export const getCariFinancialSummaryApi = (id: string) => client.get(`/caris/${id}/summary`)
export const getCariTransactionsApi = (id: string, params?: any) => client.get(`/caris/${id}/transactions`, { params })
export const addCariPersonApi = (id: string, data: any) => client.post(`/caris/${id}/persons`, data)
export const updateCariPersonApi = (id: string, personId: string, data: any) => client.put(`/caris/${id}/persons/${personId}`, data)
export const removeCariPersonApi = (id: string, personId: string) => client.delete(`/caris/${id}/persons/${personId}`)
