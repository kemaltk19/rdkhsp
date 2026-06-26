import client from './client'

export const getExpensesApi = (params?: any) => client.get('/expenses', { params })
export const getExpenseByIDApi = (id: string) => client.get(`/expenses/${id}`)
export const createExpenseApi = (data: any) => client.post('/expenses', data)
export const updateExpenseApi = (id: string, data: any) => client.put(`/expenses/${id}`, data)
export const cancelExpenseApi = (id: string) => client.post(`/expenses/${id}/cancel`)
export const getRepeatAnalysisApi = () => client.get('/expenses/repeat-analysis')

// Expense Category endpoints
export const getExpenseCategoriesApi = () => client.get('/expense-categories')
export const createExpenseCategoryApi = (data: any) => client.post('/expense-categories', data)
export const updateExpenseCategoryApi = (id: string, data: any) => client.put(`/expense-categories/${id}`, data)
export const deleteExpenseCategoryApi = (id: string) => client.delete(`/expense-categories/${id}`)
