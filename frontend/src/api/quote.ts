import client from './client'

export const getQuotesApi = (params?: any) => client.get('/quotes', { params })
export const getQuoteByIDApi = (id: string) => client.get(`/quotes/${id}`)
export const createQuoteApi = (data: any) => client.post('/quotes', data)
export const updateQuoteApi = (id: string, data: any) => client.put(`/quotes/${id}`, data)
export const deleteQuoteApi = (id: string) => client.delete(`/quotes/${id}`)
export const updateQuoteStatusApi = (id: string, status: string) => client.put(`/quotes/${id}/status`, { status })
export const convertQuoteApi = (id: string) => client.post(`/quotes/${id}/convert`)
export const sendQuoteApi = (id: string) => client.post(`/quotes/${id}/send`)
export const bulkSendQuoteApi = (ids: string[]) => client.post('/quotes/bulk-send', { ids })

// Public Endpoints
export const getPublicQuoteApi = (token: string) => client.get(`/public/quotes/${token}`)
export const acceptPublicQuoteApi = (token: string) => client.post(`/public/quotes/${token}/accept`)
export const rejectPublicQuoteApi = (token: string, note: string) => client.post(`/public/quotes/${token}/reject`, { note })
