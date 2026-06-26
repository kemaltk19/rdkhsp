import { defineStore } from 'pinia'
import { getRolesApi, createRoleApi, updateRoleApi, deleteRoleApi } from '@/api/role'

export const useRoleStore = defineStore('role', {
  state: () => ({
    roles: [] as any[],
    loading: false,
  }),
  actions: {
    async fetchRoles() {
      this.loading = true
      try {
        const res = await getRolesApi()
        this.roles = res.data.data
      } finally {
        this.loading = false
      }
    },
    async createRole(data: any) {
      this.loading = true
      try {
        const res = await createRoleApi(data)
        return res.data.data
      } finally {
        this.loading = false
      }
    },
    async updateRole(id: string, data: any) {
      this.loading = true
      try {
        const res = await updateRoleApi(id, data)
        return res.data.data
      } finally {
        this.loading = false
      }
    },
    async deleteRole(id: string) {
      this.loading = true
      try {
        await deleteRoleApi(id)
      } finally {
        this.loading = false
      }
    },
  },
})
