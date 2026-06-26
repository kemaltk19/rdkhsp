import { defineStore } from 'pinia'
import {
  getBillingStatusApi,
  getPlansApi,
  subscribeApi,
  simulateWebhookApi,
  getTransactionsApi,
  renewApi,
} from '@/api/billing'

export const useBillingStore = defineStore('billing', {
  state: () => ({
    status: null as any,
    plans: [] as any[],
    transactions: [] as any[],
    loading: false,
  }),
  actions: {
    async fetchBillingStatus() {
      this.loading = true
      try {
        const res = await getBillingStatusApi()
        this.status = res.data.data
        return this.status
      } finally {
        this.loading = false
      }
    },
    async fetchPlans() {
      this.loading = true
      try {
        const res = await getPlansApi()
        this.plans = res.data.data
        return this.plans
      } finally {
        this.loading = false
      }
    },
    async subscribe(planID: string, periodType: string) {
      this.loading = true
      try {
        const res = await subscribeApi(planID, periodType)
        return res.data.data // contains checkout_url
      } finally {
        this.loading = false
      }
    },
    async renew(periodType: string) {
      this.loading = true
      try {
        const res = await renewApi(periodType)
        await this.fetchBillingStatus()
        await this.fetchTransactions()
        return res.data.data
      } finally {
        this.loading = false
      }
    },
    async fetchTransactions() {
      this.loading = true
      try {
        const res = await getTransactionsApi()
        this.transactions = res.data.data
        return this.transactions
      } finally {
        this.loading = false
      }
    },
    async triggerWebhookSimulation(companyID: string, planID: string, sessionID: string) {
      this.loading = true
      try {
        await simulateWebhookApi({
          event_id: 'evt_mock_checkout_' + Math.random().toString(36).substring(7),
          type: 'checkout.session.completed',
          company_id: companyID,
          plan_id: planID,
          session_id: sessionID,
        })
        await this.fetchBillingStatus()
        await this.fetchTransactions()
      } finally {
        this.loading = false
      }
    },
    async cancelSubscriptionSimulation(companyID: string) {
      this.loading = true
      try {
        await simulateWebhookApi({
          event_id: 'evt_mock_cancel_' + Math.random().toString(36).substring(7),
          type: 'customer.subscription.deleted',
          company_id: companyID,
        })
        await this.fetchBillingStatus()
        await this.fetchTransactions()
      } finally {
        this.loading = false
      }
    },
  },
})

