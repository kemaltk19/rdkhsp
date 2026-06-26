import { defineStore } from 'pinia'
import {
  getEmployeesApi,
  createEmployeeApi,
  updateEmployeeApi,
  deleteEmployeeApi,
} from '@/api/employee'

export const useEmployeeStore = defineStore('employee', {
  state: () => ({
    employees: [] as any[],
    total: 0,
    loading: false,
  }),
  actions: {
    async fetchEmployees(params?: any) {
      this.loading = true
      try {
        const res = await getEmployeesApi(params)
        this.employees = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async createEmployee(data: any) {
      this.loading = true
      try {
        const res = await createEmployeeApi(data)
        return res.data.data
      } finally {
        this.loading = false
      }
    },
    async updateEmployee(id: string, data: any) {
      this.loading = true
      try {
        const res = await updateEmployeeApi(id, data)
        return res.data.data
      } finally {
        this.loading = false
      }
    },
    async deleteEmployee(id: string) {
      this.loading = true
      try {
        await deleteEmployeeApi(id)
      } finally {
        this.loading = false
      }
    },
  },
})
