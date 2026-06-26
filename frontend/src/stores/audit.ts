import { defineStore } from 'pinia'
import { getAuditLogsApi } from '@/api/audit'

export interface AuditLog {
  id: string
  company_id: string
  module: string
  record_id: string
  action: 'create' | 'update' | 'delete' | 'cancel'
  user_id: string | null
  user_name: string
  user_role: string
  summary: string
  created_at: string
}

export const useAuditStore = defineStore('audit', {
  state: () => ({
    logs: [] as AuditLog[],
    loading: false,
  }),
  actions: {
    async fetchLogs(params?: { module?: string; record_id?: string }) {
      this.loading = true
      try {
        const res = await getAuditLogsApi(params)
        this.logs = res.data.data || res.data
      } finally {
        this.loading = false
      }
    },
  },
})
