import { defineStore } from 'pinia'
import { getCurrenciesApi, createCurrencyApi, updateCurrencyApi, deleteCurrencyApi } from '@/api/currency'

export interface Currency {
  id: string
  company_id: string
  name: string
  symbol: string
  code: string
  is_crypto: boolean
  exchange_rate: number
  exchange_rate_op: '*' | '/'
  format_position: 'Left' | 'Right' | 'LeftSpace' | 'RightSpace'
  format_thousand_sep: string
  format_decimal_sep: string
  format_decimals: number
  is_default: boolean
}

export const useCurrencyStore = defineStore('currency', {
  state: () => ({
    currencies: [] as Currency[],
    loading: false,
    error: null as string | null
  }),

  getters: {
    getCurrencyByCode: (state) => (code: string) => {
      if (!Array.isArray(state.currencies)) return undefined
      return state.currencies.find(c => c.code === code)
    },
    defaultCurrency: (state) => {
      if (!Array.isArray(state.currencies)) return undefined
      return state.currencies.find(c => c.is_default) || state.currencies[0]
    }
  },

  actions: {
    async fetchCurrencies() {
      this.loading = true
      this.error = null
      try {
        const response = await getCurrenciesApi()
        const list = (response.data as any)?.data ?? response.data
        this.currencies = Array.isArray(list) ? list : []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Para birimleri yüklenemedi'
        this.currencies = []
      } finally {
        this.loading = false
      }
    },

    async createCurrency(payload: Partial<Currency>) {
      this.loading = true
      try {
        const response = await createCurrencyApi(payload)
        const created = (response.data as any)?.data ?? response.data
        this.currencies.push(created)
        return created
      } catch (err: any) {
        throw new Error(err.response?.data?.error || 'Para birimi eklenemedi')
      } finally {
        this.loading = false
      }
    },

    async updateCurrency(id: string, payload: Partial<Currency>) {
      this.loading = true
      try {
        const response = await updateCurrencyApi(id, payload)
        const updated = (response.data as any)?.data ?? response.data
        const index = this.currencies.findIndex(c => c.id === id)
        if (index !== -1) {
          this.currencies[index] = updated
        }
        
        // If this one is set to default, unset others locally to match backend
        if (payload.is_default) {
          this.currencies.forEach(c => {
            if (c.id !== id) c.is_default = false
          })
        }
        return updated
      } catch (err: any) {
        throw new Error(err.response?.data?.error || 'Para birimi güncellenemedi')
      } finally {
        this.loading = false
      }
    },

    async deleteCurrency(id: string) {
      this.loading = true
      try {
        await deleteCurrencyApi(id)
        this.currencies = this.currencies.filter(c => c.id !== id)
      } catch (err: any) {
        throw new Error(err.response?.data?.error || 'Para birimi silinemedi')
      } finally {
        this.loading = false
      }
    }
  }
})
