import { defineStore } from 'pinia'
import client from '@/api/client'

export interface Announcement {
  id: string
  title: string
  body: string
  category: string
  target_plan_id: string | null
  created_by: string
  created_at: string
}

export interface AnnouncementCategory {
  id: string
  slug: string
  name: string
  created_at: string
}

export const useAnnouncementStore = defineStore('announcement', {
  state: () => ({
    announcements: [] as Announcement[],
    categories: [] as AnnouncementCategory[],
    loading: false
  }),
  actions: {
    async fetchAll() {
      this.loading = true
      try {
        const res = await client.get('/superadmin/announcements')
        this.announcements = res.data.data || []
        return this.announcements
      } finally {
        this.loading = false
      }
    },
    async fetchForTenant() {
      this.loading = true
      try {
        const res = await client.get('/announcements')
        this.announcements = res.data.data || []
        return this.announcements
      } finally {
        this.loading = false
      }
    },
    async create(data: { title: string; body: string; category?: string; target_plan_id?: string | null }) {
      this.loading = true
      try {
        const res = await client.post('/superadmin/announcements', data)
        this.announcements.unshift(res.data.data)
        return res.data.data
      } finally {
        this.loading = false
      }
    },
    async delete(id: string) {
      this.loading = true
      try {
        await client.delete(`/superadmin/announcements/${id}`)
        this.announcements = this.announcements.filter(x => x.id !== id)
      } finally {
        this.loading = false
      }
    },
    // Superadmin: tüm kategorileri yönetim için çek.
    async fetchCategories() {
      const res = await client.get('/superadmin/announcement-categories')
      this.categories = res.data.data || []
      return this.categories
    },
    // Tenant (admin): filtre için kategorileri çek.
    async fetchCategoriesForTenant() {
      const res = await client.get('/announcement-categories')
      this.categories = res.data.data || []
      return this.categories
    },
    async createCategory(name: string) {
      const res = await client.post('/superadmin/announcement-categories', { name })
      this.categories.push(res.data.data)
      return res.data.data
    },
    async deleteCategory(id: string) {
      await client.delete(`/superadmin/announcement-categories/${id}`)
      this.categories = this.categories.filter(x => x.id !== id)
    }
  }
})
