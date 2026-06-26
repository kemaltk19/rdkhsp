import client from './client'

export interface RolePermission {
  module: string
  can_create: boolean
  can_read: boolean
  can_update: boolean
  can_delete: boolean
}

export interface Role {
  id: string
  company_id: string
  name: string
  description: string
  permissions: RolePermission[]
}

export const getRolesApi = () => client.get<{ data: Role[] }>('/roles')
export const getRoleByIDApi = (id: string) => client.get<{ data: Role }>(`/roles/${id}`)
export const createRoleApi = (data: Partial<Role>) => client.post<{ data: Role }>('/roles', data)
export const updateRoleApi = (id: string, data: Partial<Role>) => client.put<{ data: Role }>(`/roles/${id}`, data)
export const deleteRoleApi = (id: string) => client.delete(`/roles/${id}`)
