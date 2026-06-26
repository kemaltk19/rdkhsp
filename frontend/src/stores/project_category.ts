import { defineStore } from 'pinia'
import {
  getProjectCategoriesApi,
  getProjectCategoryByIDApi,
  createProjectCategoryApi,
  updateProjectCategoryApi,
  deleteProjectCategoryApi,
} from '@/api/project_category'

export interface ProjectCategory {
  id: string
  company_id: string
  name: string
  code?: string
  color?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export const useProjectCategoryStore = defineStore('projectCategory', {
  state: () => ({
    categories: [] as ProjectCategory[],
    total: 0,
    activeCategory: null as ProjectCategory | null,
    loading: false,
  }),
  actions: {
    async fetchCategories(params?: any) {
      this.loading = true
      try {
        const res = await getProjectCategoriesApi(params)
        this.categories = res.data.data
        this.total = res.data.meta.total
      } finally {
        this.loading = false
      }
    },
    async fetchCategoryByID(id: string) {
      this.loading = true
      try {
        const res = await getProjectCategoryByIDApi(id)
        this.activeCategory = res.data.data
      } finally {
        this.loading = false
      }
    },
    async createCategory(data: any) {
      const res = await createProjectCategoryApi(data)
      return res.data.data
    },
    async updateCategory(id: string, data: any) {
      const res = await updateProjectCategoryApi(id, data)
      return res.data.data
    },
    async deleteCategory(id: string) {
      await deleteProjectCategoryApi(id)
    },
  },
})
