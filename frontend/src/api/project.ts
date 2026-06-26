import client from './client'

export const getProjectsApi = (params?: any) => client.get('/projects', { params })
export const getProjectByIDApi = (id: string) => client.get(`/projects/${id}`)
export const createProjectApi = (data: any) => client.post('/projects', data)
export const updateProjectApi = (id: string, data: any) => client.put(`/projects/${id}`, data)
export const deleteProjectApi = (id: string) => client.delete(`/projects/${id}`)
export const addInvoiceToProjectApi = (projectId: string, invoiceId: string) => client.post(`/projects/${projectId}/invoices/add`, { invoice_id: invoiceId })
export const removeInvoiceFromProjectApi = (projectId: string, invoiceId: string) => client.post(`/projects/${projectId}/invoices/remove`, { invoice_id: invoiceId })
export const addQuoteToProjectApi = (projectId: string, quoteId: string) => client.post(`/projects/${projectId}/quotes/add`, { quote_id: quoteId })
export const removeQuoteFromProjectApi = (projectId: string, quoteId: string) => client.post(`/projects/${projectId}/quotes/remove`, { quote_id: quoteId })
export const addEmployeeToProjectApi = (projectId: string, employeeId: string) => client.post(`/projects/${projectId}/employees/add`, { employee_id: employeeId })
export const removeEmployeeFromProjectApi = (projectId: string, employeeId: string) => client.post(`/projects/${projectId}/employees/remove`, { employee_id: employeeId })

