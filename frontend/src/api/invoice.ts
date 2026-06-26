import client from './client'

export const getInvoicesApi = (params?: any) => client.get('/invoices', { params })
export const getInvoiceByIDApi = (id: string) => client.get(`/invoices/${id}`)
export const createInvoiceApi = (data: any) => client.post('/invoices', data)
export const updateInvoiceApi = (id: string, data: any) => client.put(`/invoices/${id}`, data)
export const updateInvoiceStatusApi = (id: string, status: string, paid_total: number) => client.put(`/invoices/${id}/status`, { status, paid_total })
export const deleteInvoiceApi = (id: string) => client.delete(`/invoices/${id}`)
export const cancelInvoiceApi = (id: string) => client.post(`/invoices/${id}/cancel`)
export const sendInvoiceApi = (id: string) => client.post(`/invoices/${id}/send`)
export const bulkSendInvoiceApi = (ids: string[]) => client.post('/invoices/bulk-send', { ids })

// Public Endpoints
export const getPublicInvoiceApi = (token: string) => client.get(`/public/invoices/${token}`)
export const disputePublicInvoiceApi = (token: string, note: string) => client.post(`/public/invoices/${token}/dispute`, { note })
export const payPublicInvoiceApi = (token: string) => client.post(`/public/invoices/${token}/pay`)
