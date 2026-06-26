import { defineStore } from 'pinia'
import {
  getCarisApi,
  getCariSummaryApi,
  getCariByIDApi,
  createCariApi,
  updateCariApi,
  deleteCariApi,
  getCariFinancialSummaryApi,
  getCariTransactionsApi,
  addCariPersonApi,
  updateCariPersonApi,
  removeCariPersonApi
} from '@/api/cari'

export interface CariBalance {
  id: string
  cari_id: string
  currency: string
  balance: string
}

export interface CariPerson {
  id: string
  cari_id: string
  name: string
  title: string
  phone: string
  email: string
  created_at: string
}

export interface Cari {
  id: string
  company_id: string
  code: string
  type: string
  group?: string
  title: string
  name: string
  contact_name: string
  tax_office: string
  tax_number: string
  email: string
  phone: string
  landline: string
  fax: string
  address: string
  city: string
  district: string
  postal_code: string
  country: string
  shipping_address: string
  shipping_city: string
  shipping_district: string
  shipping_postal_code: string
  shipping_country: string
  currency: string
  opening_balance: string
  balances: CariBalance[]
  persons: CariPerson[]
  is_active: boolean
  note: string
}

export interface CariTransaction {
  id: string
  company_id: string
  cari_id: string
  date: string
  type: 'debit' | 'credit'
  source_type: 'invoice' | 'payment' | 'expense' | 'manual'
  source_id: string | null
  description: string
  amount: string
  balance_after: string
}

export const useCariStore = defineStore('cari', {
  state: () => ({
    caris: [] as Cari[],
    total: 0,
    summary: {
      total_receivables: [] as { currency: string; amount: string }[],
      total_payables: [] as { currency: string; amount: string }[],
      net_balance: [] as { currency: string; amount: string }[],
    },
    activeCari: null as Cari | null,
    activeCariFinancialSummary: null as any,
    activeCariTxs: [] as CariTransaction[],
    activeCariTxsTotal: 0,
    loading: false,
  }),
  actions: {
    async fetchCaris(params?: any) {
      this.loading = true
      try {
        const res = await getCarisApi(params)
        this.caris = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchSummary() {
      const res = await getCariSummaryApi()
      this.summary = res.data.data
    },
    async fetchCariByID(id: string) {
      const res = await getCariByIDApi(id)
      this.activeCari = res.data.data
    },
    async fetchFinancialSummary(id: string) {
      const res = await getCariFinancialSummaryApi(id)
      this.activeCariFinancialSummary = res.data.data
    },
    async fetchTransactions(id: string, params?: any) {
      const res = await getCariTransactionsApi(id, params)
      this.activeCariTxs = res.data.data
      this.activeCariTxsTotal = res.data.meta.total
    },
    async createCari(data: any) {
      const res = await createCariApi(data)
      return res.data.data
    },
    async updateCari(id: string, data: any) {
      const res = await updateCariApi(id, data)
      return res.data.data
    },
    async deleteCari(id: string) {
      await deleteCariApi(id)
    },
    async addPerson(cariId: string, data: { name: string; title?: string; phone?: string; email?: string }) {
      const res = await addCariPersonApi(cariId, data)
      if (this.activeCari) {
        if (!this.activeCari.persons) this.activeCari.persons = []
        this.activeCari.persons.push(res.data.data)
      }
      return res.data.data
    },
    async updatePerson(cariId: string, personId: string, data: { name: string; title?: string; phone?: string; email?: string }) {
      const res = await updateCariPersonApi(cariId, personId, data)
      if (this.activeCari && this.activeCari.persons) {
        const idx = this.activeCari.persons.findIndex(p => p.id === personId)
        if (idx !== -1) this.activeCari.persons[idx] = res.data.data
      }
      return res.data.data
    },
    async removePerson(cariId: string, personId: string) {
      await removeCariPersonApi(cariId, personId)
      if (this.activeCari && this.activeCari.persons) {
        this.activeCari.persons = this.activeCari.persons.filter(p => p.id !== personId)
      }
    }
  },
})
