import { defineStore } from 'pinia'
import { getReportApi } from '@/api/report'

export const useReportStore = defineStore('report', {
  state: () => ({
    data: null as any,
    loading: false,
  }),
  actions: {
    async fetchReport(type: string, params?: any) {
      this.loading = true
      try {
        const res = await getReportApi(type, params)
        this.data = res.data.data
      } finally {
        this.loading = false
      }
    },
  },
})
