import { defineStore } from 'pinia'
import {
  getQuotesApi,
  getQuoteByIDApi,
  createQuoteApi,
  updateQuoteApi,
  deleteQuoteApi,
  updateQuoteStatusApi,
  convertQuoteApi,
  sendQuoteApi,
  bulkSendQuoteApi,
  getPublicQuoteApi,
  acceptPublicQuoteApi,
  rejectPublicQuoteApi,
} from '@/api/quote'

export interface QuoteItem {
  id: string
  company_id: string
  quote_id: string
  product_id: string | null
  description: string
  quantity: string
  unit: string
  unit_price: string
  discount_rate: string
  tax_rate: string
  line_total: string
  currency: string
  exchange_rate: string
  exchange_rate_op: string
}

import type { UserRef } from './invoice'

export interface Quote {
  id: string
  company_id: string
  cari_id: string
  type?: 'sales' | 'purchase'
  number: string
  date: string
  expiry_date: string
  currency: string
  exchange_rate: string
  subtotal: string
  discount_total: string
  tax_total: string
  total: string
  status: 'draft' | 'sent' | 'accepted' | 'rejected' | 'expired' | 'converted'
  note: string
  reject_note?: string
  responded_at?: string
  sent_at?: string
  last_sent_at?: string
  public_token?: string
  converted_invoice_id?: string
  items?: QuoteItem[]
  created_at: string
  updated_at: string
  created_by_user?: UserRef | null
  updated_by_user?: UserRef | null
}

export const useQuoteStore = defineStore('quote', {
  state: () => ({
    quotes: [] as Quote[],
    total: 0,
    activeQuote: null as Quote | null,
    loading: false,
  }),
  actions: {
    async fetchQuotes(params?: any) {
      this.loading = true
      try {
        const res = await getQuotesApi(params)
        this.quotes = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchQuoteByID(id: string) {
      this.loading = true
      try {
        const res = await getQuoteByIDApi(id)
        this.activeQuote = res.data.data
      } finally {
        this.loading = false
      }
    },
    async createQuote(data: any) {
      const res = await createQuoteApi(data)
      return res.data.data
    },
    async updateQuote(id: string, data: any) {
      const res = await updateQuoteApi(id, data)
      return res.data.data
    },
    async deleteQuote(id: string) {
      await deleteQuoteApi(id)
    },
    async updateQuoteStatus(id: string, status: string) {
      await updateQuoteStatusApi(id, status)
    },
    async convertQuote(id: string) {
      const res = await convertQuoteApi(id)
      return res.data.data
    },
    async sendQuote(id: string) {
      const res = await sendQuoteApi(id)
      return res.data
    },
    async bulkSendQuote(ids: string[]) {
      const res = await bulkSendQuoteApi(ids)
      return res.data.data as { sent: string[]; failed: { id: string; error: string }[] }
    },
    async getPublicQuote(token: string) {
      const res = await getPublicQuoteApi(token)
      return res.data.data
    },
    async acceptPublicQuote(token: string) {
      const res = await acceptPublicQuoteApi(token)
      return res.data.data
    },
    async rejectPublicQuote(token: string, note: string) {
      const res = await rejectPublicQuoteApi(token, note)
      return res.data.data
    },
  },
})
