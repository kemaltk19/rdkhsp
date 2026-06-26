import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getCriticalStockApi } from '@/api/product'
import { getRepeatAnalysisApi } from '@/api/expense'
import { getBillingStatusApi } from '@/api/billing'
import { getNotificationsApi, markAsReadApi, markAllAsReadApi } from '@/api/notification'
import client from '@/api/client'

export const useNotificationStore = defineStore('notification', () => {
  const criticalStocks = ref<any[]>([])
  const repeatExpenses = ref<any[]>([])
  const announcements = ref<any[]>([])
  const billingWarning = ref<{ daysLeft: number; periodEnd: string } | null>(null)
  
  // Database backed notifications
  const dbNotifications = ref<any[]>([])
  const disputedInvoices = ref<any[]>([])
  const acceptedQuotes = ref<any[]>([])
  const rejectedQuotes = ref<any[]>([])
  
  const totalNotifications = ref<number>(0)
  let pollingTimer: ReturnType<typeof setInterval> | null = null

  const fetchNotifications = async () => {
    try {
      const [stockRes, expenseRes, annRes, billingRes, dbNotifRes] = await Promise.all([
        getCriticalStockApi(),
        getRepeatAnalysisApi(),
        client.get('/announcements').catch(() => ({ data: { data: [] } })),
        getBillingStatusApi().catch(() => ({ data: { data: null } })),
        getNotificationsApi({ page: 1, limit: 100 }).catch(() => ({ data: { data: [] } }))
      ])

      const dismissedStockKeys = JSON.parse(localStorage.getItem('dismissedStockIds') || '[]')
      const dismissedExpenseIds = JSON.parse(localStorage.getItem('dismissedExpenseIds') || '[]')
      const dismissedAnnouncementIds = JSON.parse(localStorage.getItem('dismissedAnnouncementIds') || '[]')

      // 1. Dynamic/Virtual notifications (kept in localStorage)
      criticalStocks.value = (stockRes.data.data || []).filter((s: any) => !dismissedStockKeys.includes(`${s.id}:${s.current_stock}`))
      repeatExpenses.value = (expenseRes.data.data || []).filter((e: any) => !dismissedExpenseIds.includes(e.category_name))
      announcements.value = (annRes.data.data || []).filter((a: any) => !dismissedAnnouncementIds.includes(a.id))

      // Billing warning
      const billingStatus = billingRes.data.data
      const today = new Date().toISOString().slice(0, 10)
      const dismissedBillingKey = localStorage.getItem('dismissedBillingWarningKey')
      if (
        billingStatus &&
        billingStatus.subscription_status === 'active' &&
        billingStatus.current_period_end &&
        billingStatus.period_days_left <= 15
      ) {
        const billingKey = `${today}:${billingStatus.period_days_left}`
        billingWarning.value = dismissedBillingKey === billingKey
          ? null
          : { daysLeft: billingStatus.period_days_left, periodEnd: billingStatus.current_period_end }
      } else {
        billingWarning.value = null
      }

      // 2. Database-backed notifications (filtered by is_read == false)
      const allDbNotifs = dbNotifRes.data.data || []
      dbNotifications.value = allDbNotifs.filter((n: any) => !n.is_read)

      disputedInvoices.value = dbNotifications.value.filter((n: any) => n.type === 'invoice_dispute')
      acceptedQuotes.value = dbNotifications.value.filter((n: any) => n.type === 'quote_accepted')
      rejectedQuotes.value = dbNotifications.value.filter((n: any) => n.type === 'quote_rejected')

      recountTotal()
    } catch (error) {
      console.error('Error fetching notifications:', error)
    }
  }

  const recountTotal = () => {
    totalNotifications.value =
      criticalStocks.value.length +
      repeatExpenses.value.length +
      announcements.value.length +
      dbNotifications.value.length +
      (billingWarning.value ? 1 : 0)
  }

  const startPolling = () => {
    fetchNotifications() // İlk yükleme
    // Her 60 saniyede bir otomatik güncelle
    if (!pollingTimer) {
      pollingTimer = setInterval(() => {
        fetchNotifications()
      }, 60_000)
    }
  }

  const stopPolling = () => {
    if (pollingTimer) {
      clearInterval(pollingTimer)
      pollingTimer = null
    }
  }

  // Dismiss virtual notifications
  const dismissStockNotification = () => {
    const keys = criticalStocks.value.map(s => `${s.id}:${s.current_stock}`)
    const existing = JSON.parse(localStorage.getItem('dismissedStockIds') || '[]')
    localStorage.setItem('dismissedStockIds', JSON.stringify([...new Set([...existing, ...keys])]))
    criticalStocks.value = []
    recountTotal()
  }

  const dismissExpenseNotification = () => {
    const ids = repeatExpenses.value.map(e => e.category_name)
    const existing = JSON.parse(localStorage.getItem('dismissedExpenseIds') || '[]')
    localStorage.setItem('dismissedExpenseIds', JSON.stringify([...new Set([...existing, ...ids])]))
    repeatExpenses.value = []
    recountTotal()
  }

  const dismissAnnouncementNotification = () => {
    const ids = announcements.value.map(a => a.id)
    const existing = JSON.parse(localStorage.getItem('dismissedAnnouncementIds') || '[]')
    localStorage.setItem('dismissedAnnouncementIds', JSON.stringify([...new Set([...existing, ...ids])]))
    announcements.value = []
    recountTotal()
  }

  const dismissBillingWarning = () => {
    if (!billingWarning.value) return
    const today = new Date().toISOString().slice(0, 10)
    localStorage.setItem('dismissedBillingWarningKey', `${today}:${billingWarning.value.daysLeft}`)
    billingWarning.value = null
    recountTotal()
  }

  // Database-backed notification actions
  const markNotificationAsRead = async (id: string) => {
    try {
      await markAsReadApi(id)
      dbNotifications.value = dbNotifications.value.filter((n: any) => n.id !== id)
      disputedInvoices.value = disputedInvoices.value.filter((n: any) => n.id !== id)
      acceptedQuotes.value = acceptedQuotes.value.filter((n: any) => n.id !== id)
      rejectedQuotes.value = rejectedQuotes.value.filter((n: any) => n.id !== id)
      recountTotal()
    } catch (error) {
      console.error('Error marking notification as read:', error)
    }
  }

  const markAllNotificationsAsRead = async () => {
    try {
      await markAllAsReadApi()
      dbNotifications.value = []
      disputedInvoices.value = []
      acceptedQuotes.value = []
      rejectedQuotes.value = []
      recountTotal()
    } catch (error) {
      console.error('Error marking all notifications as read:', error)
    }
  }

  return {
    criticalStocks,
    repeatExpenses,
    announcements,
    billingWarning,
    disputedInvoices,
    acceptedQuotes,
    rejectedQuotes,
    dbNotifications,
    totalNotifications,
    fetchNotifications,
    startPolling,
    stopPolling,
    dismissStockNotification,
    dismissExpenseNotification,
    dismissBillingWarning,
    dismissAnnouncementNotification,
    markNotificationAsRead,
    markAllNotificationsAsRead
  }
})
