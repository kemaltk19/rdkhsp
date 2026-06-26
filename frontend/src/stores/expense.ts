import { defineStore } from 'pinia'
import {
  getExpensesApi,
  getExpenseByIDApi,
  createExpenseApi,
  updateExpenseApi,
  cancelExpenseApi,
  getExpenseCategoriesApi,
  createExpenseCategoryApi,
  updateExpenseCategoryApi,
  deleteExpenseCategoryApi,
} from '@/api/expense'

export interface ExpenseCategory {
  id: string
  company_id: string
  name: string
  has_active_recurring?: boolean
}

import type { UserRef } from './invoice'

export interface Expense {
  id: string
  company_id: string
  category_id: string
  category?: ExpenseCategory
  cari_id: string | null
  cari?: {
    id: string
    name: string
    title?: string
  }
  date: string
  description: string
  amount: string
  tax_rate: string
  tax_amount: string
  total: string
  account_kind: 'cash' | 'bank' | null
  account_id: string | null
  status: 'paid' | 'unpaid' | 'canceled'
  note: string
  created_at: string
  updated_at?: string
  created_by_user?: UserRef | null
  updated_by_user?: UserRef | null
}

export const useExpenseStore = defineStore('expense', {
  state: () => ({
    expenses: [] as Expense[],
    total: 0,
    categories: [] as ExpenseCategory[],
    loading: false,
  }),
  actions: {
    async fetchExpenses(params?: any) {
      this.loading = true
      try {
        const res = await getExpensesApi(params)
        this.expenses = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchCategories() {
      const res = await getExpenseCategoriesApi()
      this.categories = res.data.data
    },
    async getExpenseByID(id: string) {
      const res = await getExpenseByIDApi(id)
      return res.data.data
    },
    async createExpense(data: any) {
      const res = await createExpenseApi(data)
      return res.data.data
    },
    async updateExpense(id: string, data: any) {
      const res = await updateExpenseApi(id, data)
      return res.data.data
    },
    async cancelExpense(id: string) {
      await cancelExpenseApi(id)
    },
    async createCategory(data: any) {
      const res = await createExpenseCategoryApi(data)
      await this.fetchCategories()
      return res.data.data
    },
    async updateCategory(id: string, data: any) {
      const res = await updateExpenseCategoryApi(id, data)
      await this.fetchCategories()
      return res.data.data
    },
    async deleteCategory(id: string) {
      await deleteExpenseCategoryApi(id)
      await this.fetchCategories()
    },
  },
})
