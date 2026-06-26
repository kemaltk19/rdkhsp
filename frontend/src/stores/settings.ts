import { defineStore } from 'pinia'
import {
  getCompanyProfileApi,
  updateCompanyProfileApi,
  getSettingApi,
  saveSettingApi,
  listSettingsApi,
  updateEnabledModulesApi,
} from '@/api/settings'

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    company: null as any,
    settings: {} as Record<string, string>,
    loading: false,
  }),
  actions: {
    async updateEnabledModules(enabledModules: string[]) {
      this.loading = true
      try {
        const res = await updateEnabledModulesApi(enabledModules)
        this.company = res.data.data
        return this.company
      } finally {
        this.loading = false
      }
    },
    async fetchCompanyProfile() {
      this.loading = true
      try {
        const res = await getCompanyProfileApi()
        this.company = res.data.data
        return this.company
      } finally {
        this.loading = false
      }
    },
    async updateCompanyProfile(data: any) {
      this.loading = true
      try {
        const res = await updateCompanyProfileApi(data)
        this.company = res.data.data
        return this.company
      } finally {
        this.loading = false
      }
    },
    async fetchSetting(key: string) {
      try {
        const res = await getSettingApi(key)
        this.settings[key] = res.data.data.value
        return res.data.data.value
      } catch {
        return ''
      }
    },
    async saveSetting(key: string, value: string, category?: string) {
      this.loading = true
      try {
        const res = await saveSettingApi({ key, value, category })
        this.settings[key] = value
        return res.data.data
      } finally {
        this.loading = false
      }
    },
    async fetchSettingsByCategory(category?: string) {
      this.loading = true
      try {
        const res = await listSettingsApi(category)
        const list = res.data.data || []
        for (const item of list) {
          this.settings[item.key] = item.value
        }
        return this.settings
      } finally {
        this.loading = false
      }
    },
  },
})
