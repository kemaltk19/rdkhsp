import client from './client'

export const getProjectCategoriesApi = (params?: any) => client.get('/project-categories', { params })
export const getProjectCategoryByIDApi = (id: string) => client.get(`/project-categories/${id}`)
export const createProjectCategoryApi = (data: any) => client.post('/project-categories', data)
export const updateProjectCategoryApi = (id: string, data: any) => client.put(`/project-categories/${id}`, data)
export const deleteProjectCategoryApi = (id: string) => client.delete(`/project-categories/${id}`)
