import { defineStore } from 'pinia'
import {
  getProjectsApi,
  getProjectByIDApi,
  createProjectApi,
  updateProjectApi,
  deleteProjectApi,
  addInvoiceToProjectApi,
  removeInvoiceFromProjectApi,
  addQuoteToProjectApi,
  removeQuoteFromProjectApi,
  addEmployeeToProjectApi,
  removeEmployeeFromProjectApi,
} from '@/api/project'

import type { UserRef } from './invoice'

export interface Project {
  id: string
  company_id: string
  cari_id: string
  name: string
  description?: string
  code: string
  category_id?: string | null
  status: 'planning' | 'in_progress' | 'on_hold' | 'completed' | 'cancelled'
  start_date: string
  end_date: string
  budget?: number
  note?: string
  employees?: any[]
  invoices?: any[]
  quotes?: any[]
  category?: any
  created_at: string
  updated_at: string
  created_by_user?: UserRef | null
  updated_by_user?: UserRef | null
}

export const useProjectStore = defineStore('project', {
  state: () => ({
    projects: [] as Project[],
    total: 0,
    activeProject: null as Project | null,
    loading: false,
  }),
  actions: {
    async fetchProjects(params?: any) {
      this.loading = true
      try {
        const res = await getProjectsApi(params)
        this.projects = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchProjectByID(id: string) {
      this.loading = true
      try {
        const res = await getProjectByIDApi(id)
        this.activeProject = res.data.data
      } finally {
        this.loading = false
      }
    },
    async createProject(data: any) {
      const res = await createProjectApi(data)
      return res.data.data
    },
    async updateProject(id: string, data: any) {
      const res = await updateProjectApi(id, data)
      return res.data.data
    },
    async deleteProject(id: string) {
      await deleteProjectApi(id)
    },
    async addInvoiceToProject(projectId: string, invoiceId: string) {
      const res = await addInvoiceToProjectApi(projectId, invoiceId)
      return res.data.data
    },
    async removeInvoiceFromProject(projectId: string, invoiceId: string) {
      const res = await removeInvoiceFromProjectApi(projectId, invoiceId)
      return res.data.data
    },
    async addQuoteToProject(projectId: string, quoteId: string) {
      const res = await addQuoteToProjectApi(projectId, quoteId)
      return res.data.data
    },
    async removeQuoteFromProject(projectId: string, quoteId: string) {
      const res = await removeQuoteFromProjectApi(projectId, quoteId)
      return res.data.data
    },
    async addEmployeeToProject(projectId: string, employeeId: string) {
      const res = await addEmployeeToProjectApi(projectId, employeeId)
      return res.data.data
    },
    async removeEmployeeFromProject(projectId: string, employeeId: string) {
      const res = await removeEmployeeFromProjectApi(projectId, employeeId)
      return res.data.data
    },
  },
})
