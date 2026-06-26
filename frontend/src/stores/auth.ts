import { defineStore } from 'pinia'
import { registerApi, loginApi, logoutApi, meApi, changePasswordApi } from '@/api/auth'

export interface User {
  id: string
  company_id: string
  name: string
  email: string
  role: string
  role_id: string | null
  locale: string
  is_active: boolean
}

export interface RolePermission {
  module: string
  can_create: boolean
  can_read: boolean
  can_update: boolean
  can_delete: boolean
}


export interface Company {
  id: string
  name: string
  slug: string
  currency: string
  locale: string
  subscription_status: string
  trial_ends_at: string
  timezone?: string
  enabled_modules?: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    company: null as Company | null,
    permissions: [] as RolePermission[],
    ready: false,
  }),
  getters: {
    isAuthenticated: (state) => !!state.user,
    role: (state) => state.user?.role || null,
  },
  actions: {
    async fetchMe() {
      try {
        const res = await meApi()
        this.user = res.data.data.user
        this.company = res.data.data.company
        this.permissions = res.data.data.permissions || []
      } catch (err) {
        this.clearSession()
      } finally {
        this.ready = true
      }
    },
    async login(data: any) {
      const res = await loginApi(data)
      this.user = res.data.data.user
      this.company = res.data.data.company
      this.permissions = res.data.data.permissions || []
      this.ready = true
    },
    async register(data: any) {
      const res = await registerApi(data)
      this.user = res.data.data.user
      this.company = res.data.data.company
      this.permissions = res.data.data.permissions || []
      this.ready = true
    },
    async logout() {
      try {
        await logoutApi()
      } finally {
        this.clearSession()
        window.location.href = '/login'
      }
    },
    async changePassword(data: any) {
      await changePasswordApi(data)
    },
    clearSession() {
      this.user = null
      this.company = null
      this.permissions = []
    },
  },
})
