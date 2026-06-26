import { defineStore } from 'pinia'
import {
  getCompaniesApi,
  createCompanyApi,
  updateCompanyApi,
  deleteCompanyApi,
  toggleCompanyStatusApi,
  getPlansApi,
  createPlanApi,
  updatePlanApi,
  deletePlanApi,
} from '../api/superadmin'

export const useSuperadminStore = defineStore('superadmin', {
  state: () => ({
    companies: [] as any[],
    plans: [] as any[],
    loading: false
  }),
  actions: {
    async fetchCompanies() {
      this.loading = true
      try {
        const response = await getCompaniesApi()
        this.companies = response.data.data
      } finally {
        this.loading = false
      }
    },
    async createCompany(data: any) {
      this.loading = true
      try {
        await createCompanyApi(data)
        await this.fetchCompanies()
      } finally {
        this.loading = false
      }
    },
    async updateCompany(id: string, data: any) {
      this.loading = true
      try {
        await updateCompanyApi(id, data)
        await this.fetchCompanies()
      } finally {
        this.loading = false
      }
    },
    async deleteCompany(id: string) {
      this.loading = true
      try {
        await deleteCompanyApi(id)
        await this.fetchCompanies()
      } finally {
        this.loading = false
      }
    },
    async toggleCompanyStatus(id: string, action: 'suspend' | 'activate') {
      await toggleCompanyStatusApi(id, action)
      await this.fetchCompanies()
    },
    async fetchPlans() {
      this.loading = true
      try {
        const response = await getPlansApi()
        this.plans = response.data.data
      } finally {
        this.loading = false
      }
    },
    async createPlan(data: any) {
      await createPlanApi(data)
      await this.fetchPlans()
    },
    async updatePlan(id: string, data: any) {
      await updatePlanApi(id, data)
      await this.fetchPlans()
    },
    async deletePlan(id: string) {
      await deletePlanApi(id)
      await this.fetchPlans()
    }
  }
})
