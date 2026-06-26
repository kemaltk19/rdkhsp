import client from './client'

export const getEmployeesApi = (params?: any) => client.get('/employees', { params })
export const getEmployeeByIDApi = (id: string) => client.get(`/employees/${id}`)
export const createEmployeeApi = (data: any) => client.post('/employees', data)
export const updateEmployeeApi = (id: string, data: any) => client.put(`/employees/${id}`, data)
export const deleteEmployeeApi = (id: string) => client.delete(`/employees/${id}`)
