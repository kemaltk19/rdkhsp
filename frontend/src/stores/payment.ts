import { defineStore } from 'pinia'
import {
  getPaymentsApi,
  createPaymentApi,
  cancelPaymentApi,
  updatePaymentApi,
  createCashAccountApi,
  getCashAccountsApi,
  updateCashAccountApi,
  deleteCashAccountApi,
  createBankAccountApi,
  getBankAccountsApi,
  updateBankAccountApi,
  deleteBankAccountApi,
  transferCashApi,
  getCashTransactionsApi,
} from '@/api/payment'
import type { UserRef } from './invoice'

export interface CashAccount {
  id: string
  company_id: string
  code: string
  name: string
  account_no: string
  description: string
  currency: string
  balance: string
  is_default: boolean
}

export interface BankAccount {
  id: string
  company_id: string
  code: string
  name: string
  account_no: string
  description: string
  iban: string
  currency: string
  balance: string
}

export interface Payment {
  id: string
  company_id: string
  cari_id: string
  type: 'collection' | 'payment'
  date: string
  method: 'cash' | 'bank' | 'card' | 'check'
  account_kind: 'cash' | 'bank'
  account_id: string
  amount: string
  currency: string
  invoice_id: string | null
  reference: string
  note: string
  status: 'completed' | 'canceled'
  created_at: string
  updated_at?: string
  created_by_user?: UserRef | null
  updated_by_user?: UserRef | null
}

export interface CashTransaction {
  id: string
  company_id: string
  account_kind: 'cash' | 'bank'
  account_id: string
  date: string
  type: 'in' | 'out'
  source_type: 'payment' | 'expense' | 'manual' | 'transfer'
  source_id: string | null
  amount: string
  balance_after: string
  description: string
  created_at: string
}

export const usePaymentStore = defineStore('payment', {
  state: () => ({
    payments: [] as Payment[],
    total: 0,
    cashAccounts: [] as CashAccount[],
    bankAccounts: [] as BankAccount[],
    transactions: [] as CashTransaction[],
    loading: false,
  }),
  actions: {
    async fetchPayments(params?: any) {
      this.loading = true
      try {
        const res = await getPaymentsApi(params)
        this.payments = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchAccounts() {
      const [cashRes, bankRes] = await Promise.all([
        getCashAccountsApi(),
        getBankAccountsApi(),
      ])
      this.cashAccounts = cashRes.data.data
      this.bankAccounts = bankRes.data.data
    },
    async createPayment(data: any) {
      const res = await createPaymentApi(data)
      return res.data.data
    },
    async updatePayment(id: string, data: any) {
      const res = await updatePaymentApi(id, data)
      return res.data.data
    },
    async cancelPayment(id: string) {
      await cancelPaymentApi(id)
    },
    async createCashAccount(data: any) {
      const res = await createCashAccountApi(data)
      await this.fetchAccounts()
      return res.data.data
    },
    async updateCashAccount(id: string, data: any) {
      const res = await updateCashAccountApi(id, data)
      await this.fetchAccounts()
      return res.data.data
    },
    async deleteCashAccount(id: string) {
      await deleteCashAccountApi(id)
      await this.fetchAccounts()
    },
    async createBankAccount(data: any) {
      const res = await createBankAccountApi(data)
      await this.fetchAccounts()
      return res.data.data
    },
    async updateBankAccount(id: string, data: any) {
      const res = await updateBankAccountApi(id, data)
      await this.fetchAccounts()
      return res.data.data
    },
    async deleteBankAccount(id: string) {
      await deleteBankAccountApi(id)
      await this.fetchAccounts()
    },
    async transferCash(data: any) {
      const res = await transferCashApi(data)
      await this.fetchAccounts()
      return res.data.data
    },
    async fetchCashTransactions(params: any) {
      this.loading = true
      try {
        const res = await getCashTransactionsApi(params)
        this.transactions = res.data.data
      } finally {
        this.loading = false
      }
    },
  },
})
