import { defineStore } from 'pinia'
import { getDashboardStatsApi } from '@/api/dashboard'

export interface RecentCariTx {
  id: string
  cari_name: string
  date: string
  type: 'debit' | 'credit'
  source_type: 'invoice' | 'payment' | 'expense' | 'manual'
  description: string
  amount: string
}

export interface RecentCashTx {
  id: string
  account_name: string
  account_kind: 'cash' | 'bank'
  date: string
  type: 'in' | 'out'
  amount: string
  description: string
}

export interface RecentExpenseTx {
  id: string
  category_name: string
  date: string
  amount: string
  currency: string
  description: string
}

export interface ChartDataPoint {
  month: string
  total: string
}

export interface CurrencyAmount {
  currency: string
  amount: string
}

export interface ChartSeriesData {
  currency: string
  data: ChartDataPoint[]
}

export interface DashboardStats {
  ciro: CurrencyAmount[]
  to_collect: CurrencyAmount[]
  cash_bank_total: CurrencyAmount[]
  overdue_total: CurrencyAmount[]
  recent_cari_tx: RecentCariTx[]
  recent_cash_tx: RecentCashTx[]
  recent_expenses: RecentExpenseTx[]
  chart_data: ChartDataPoint[]
  chart_series: ChartSeriesData[]
}

export const useDashboardStore = defineStore('dashboard', {
  state: () => ({
    stats: null as DashboardStats | null,
    loading: false,
  }),
  actions: {
    async fetchStats() {
      this.loading = true
      try {
        const res = await getDashboardStatsApi()
        this.stats = res.data.data
      } finally {
        this.loading = false
      }
    },
  },
})
