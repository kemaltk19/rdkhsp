import client from './client'

export const registerApi = (data: any) => client.post('/auth/register', data)
export const loginApi = (data: any) => client.post('/auth/login', data)
export const logoutApi = () => client.post('/auth/logout')
export const meApi = () => client.get('/auth/me')
export const changePasswordApi = (data: any) => client.put('/auth/password', data)
