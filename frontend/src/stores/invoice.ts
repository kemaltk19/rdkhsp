import { defineStore } from 'pinia'
import {
  getInvoicesApi,
  getInvoiceByIDApi,
  createInvoiceApi,
  updateInvoiceApi,
  updateInvoiceStatusApi,
  deleteInvoiceApi,
  cancelInvoiceApi,
  sendInvoiceApi,
  bulkSendInvoiceApi,
} from '@/api/invoice'

export interface InvoiceItem {
  id: string
  company_id: string
  invoice_id: string
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
  exchange_rate_op: '*' | '/'
}

export interface UserRef {
  id: string
  name: string
  role: string
}

export interface Invoice {
  id: string
  company_id: string
  cari_id: string
  type: 'sales' | 'purchase'
  number: string
  date: string
  due_date: string
  currency: string
  exchange_rate: string
  subtotal: string
  discount_total: string
  tax_total: string
  total: string
  paid_total: string
  status: 'draft' | 'sent' | 'disputed' | 'partial' | 'paid' | 'canceled'
  process_stock?: boolean
  note: string
  dispute_note?: string
  disputed_at?: string
  sent_at?: string
  last_sent_at?: string
  items?: InvoiceItem[]
  created_at: string
  updated_at: string
  created_by_user?: UserRef | null
  updated_by_user?: UserRef | null
}

export const useInvoiceStore = defineStore('invoice', {
  state: () => ({
    invoices: [] as Invoice[],
    total: 0,
    activeInvoice: null as Invoice | null,
    loading: false,
  }),
  actions: {
    async fetchInvoices(params?: any) {
      this.loading = true
      try {
        const res = await getInvoicesApi(params)
        this.invoices = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchInvoiceByID(id: string) {
      this.loading = true
      try {
        const res = await getInvoiceByIDApi(id)
        this.activeInvoice = res.data.data
      } finally {
        this.loading = false
      }
    },
    async createInvoice(data: any) {
      const res = await createInvoiceApi(data)
      return res.data.data
    },
    async updateInvoice(id: string, data: any) {
      const res = await updateInvoiceApi(id, data)
      return res.data.data
    },
    async updateInvoiceStatus(id: string, status: string, paidTotal: number) {
      const res = await updateInvoiceStatusApi(id, status, paidTotal)
      return res.data.data
    },
    async deleteInvoice(id: string) {
      await deleteInvoiceApi(id)
    },
    async cancelInvoice(id: string) {
      await cancelInvoiceApi(id)
    },
    async sendInvoice(id: string) {
      await sendInvoiceApi(id)
    },
    async bulkSendInvoice(ids: string[]) {
      const res = await bulkSendInvoiceApi(ids)
      return res.data.data as { sent: string[]; failed: { id: string; error: string }[] }
    },
  },
})
