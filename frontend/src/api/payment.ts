import client from './client'

export const getPaymentsApi = (params?: any) => client.get('/payments', { params })
export const getPaymentByIDApi = (id: string) => client.get(`/payments/${id}`)
export const createPaymentApi = (data: any) => client.post('/payments', data)
export const cancelPaymentApi = (id: string) => client.post(`/payments/${id}/cancel`)
export const updatePaymentApi = (id: string, data: any) => client.put(`/payments/${id}`, data)

// Cash & Bank account endpoints
export const createCashAccountApi = (data: any) => client.post('/payments/cash-accounts', data)
export const getCashAccountsApi = () => client.get('/payments/cash-accounts')
export const updateCashAccountApi = (id: string, data: any) => client.put(`/payments/cash-accounts/${id}`, data)
export const deleteCashAccountApi = (id: string) => client.delete(`/payments/cash-accounts/${id}`)

export const createBankAccountApi = (data: any) => client.post('/payments/bank-accounts', data)
export const getBankAccountsApi = () => client.get('/payments/bank-accounts')
export const updateBankAccountApi = (id: string, data: any) => client.put(`/payments/bank-accounts/${id}`, data)
export const deleteBankAccountApi = (id: string) => client.delete(`/payments/bank-accounts/${id}`)

// Cash Transactions (Virman & Logs)
export const transferCashApi = (data: any) => client.post('/cash-transactions/transfer', data)
export const getCashTransactionsApi = (params: any) => client.get('/cash-transactions', { params })
