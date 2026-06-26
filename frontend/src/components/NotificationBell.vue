<script setup lang="ts">
import { onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notification'
import OverlayPanel from 'primevue/overlaypanel'
import { ref } from 'vue'

const router = useRouter()
const notificationStore = useNotificationStore()
const op = ref()

onMounted(() => {
  notificationStore.startPolling()
})

onUnmounted(() => {
  notificationStore.stopPolling()
})

// ⚠️ totalNotifications ref'i yerine direk store array'larından hesapla
// Bu sayede fetch tamamlanınca Vue reaktivitesi anında devreye girer
const total = computed(() =>
  notificationStore.criticalStocks.length +
  notificationStore.repeatExpenses.length +
  notificationStore.announcements.length +
  notificationStore.dbNotifications.length +
  (notificationStore.billingWarning ? 1 : 0)
)

const criticalStocks = computed(() => notificationStore.criticalStocks)
const repeatExpenses = computed(() => notificationStore.repeatExpenses)
const announcements = computed(() => notificationStore.announcements)
const billingWarning = computed(() => notificationStore.billingWarning)
const disputedInvoices = computed(() => notificationStore.disputedInvoices)
const acceptedQuotes = computed(() => notificationStore.acceptedQuotes)
const rejectedQuotes = computed(() => notificationStore.rejectedQuotes)
const dbNotificationsCount = computed(() => notificationStore.dbNotifications.length)

const togglePanel = (event: Event) => {
  notificationStore.fetchNotifications()
  op.value.toggle(event)
}

const goToProducts = () => {
  notificationStore.dismissStockNotification()
  op.value.hide()
  router.push('/products')
}

const goToExpenses = () => {
  notificationStore.dismissExpenseNotification()
  op.value.hide()
  router.push('/expenses')
}

const goToAnnouncements = () => {
  notificationStore.dismissAnnouncementNotification()
  op.value.hide()
  router.push('/announcements')
}

const goToBilling = () => {
  notificationStore.dismissBillingWarning()
  op.value.hide()
  router.push('/settings?tab=billing')
}

const clickNotification = async (notif: any) => {
  await notificationStore.markNotificationAsRead(notif.id)
  op.value.hide()
  if (notif.target_type === 'invoice') {
    router.push(`/invoices/${notif.target_id}`)
  } else if (notif.target_type === 'quote') {
    router.push(`/quotes/${notif.target_id}`)
  }
}

const goToInvoices = async () => {
  const ids = disputedInvoices.value.map(n => n.id)
  for (const id of ids) {
    await notificationStore.markNotificationAsRead(id)
  }
  op.value.hide()
  router.push('/invoices')
}

const goToQuotes = async () => {
  const ids = [...acceptedQuotes.value, ...rejectedQuotes.value].map(n => n.id)
  for (const id of ids) {
    await notificationStore.markNotificationAsRead(id)
  }
  op.value.hide()
  router.push('/quotes')
}

const markAllAsRead = async () => {
  await notificationStore.markAllNotificationsAsRead()
}

const formatPeriodEnd = (dateStr: string) => new Date(dateStr).toLocaleDateString('tr-TR')
</script>

<template>
  <div class="notification-bell" @click="togglePanel">
    <!-- Badge: doğrudan .notification-bell'in çocuğu, overflow:visible ile -->
    <i class="pi pi-bell text-xl"></i>
    <span v-if="total > 0" class="notif-badge">
      {{ total > 99 ? '99+' : total }}
    </span>

    <OverlayPanel ref="op" class="w-80">
      <div class="flex flex-col gap-4">
        <div class="flex justify-between items-center border-b pb-2">
          <span class="font-bold text-lg">Bildirimler</span>
          <button
            v-if="dbNotificationsCount > 0"
            class="text-xs text-blue-500 hover:underline cursor-pointer"
            @click.stop="markAllAsRead"
          >
            Tümünü Okundu İşaretle
          </button>
        </div>

        <div v-if="billingWarning" class="notification-group">
          <div class="text-red-500 font-semibold mb-2 flex justify-between items-center cursor-pointer" @click="goToBilling">
            <span><i class="pi pi-exclamation-circle mr-2"></i>Abonelik Süresi Doluyor</span>
            <i class="pi pi-chevron-right text-xs"></i>
          </div>
          <div class="text-sm text-gray-600 dark:text-gray-400">
            <template v-if="billingWarning.daysLeft > 0">
              Aboneliğinizin bitişine <strong>{{ billingWarning.daysLeft }} gün</strong> kaldı ({{ formatPeriodEnd(billingWarning.periodEnd) }}). Kesintisiz kullanım için yenileyin.
            </template>
            <template v-else>
              Aboneliğinizin süresi doldu ({{ formatPeriodEnd(billingWarning.periodEnd) }}). Hizmetin kesilmemesi için lütfen yenileyin.
            </template>
          </div>
        </div>

        <div v-if="disputedInvoices.length > 0" class="notification-group">
          <div class="text-red-500 font-semibold mb-2 flex justify-between items-center cursor-pointer" @click="goToInvoices">
            <span><i class="pi pi-exclamation-circle mr-2"></i>Fatura İtirazı ({{ disputedInvoices.length }})</span>
            <i class="pi pi-chevron-right text-xs"></i>
          </div>
          <ul class="text-sm mt-2 max-h-32 overflow-y-auto flex flex-col gap-2">
            <li v-for="notif in disputedInvoices" :key="notif.id" class="border-b border-gray-100 dark:border-gray-800 pb-2 flex flex-col cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 p-1 rounded" @click="clickNotification(notif)">
              <span class="font-semibold text-red-500 text-xs">{{ notif.title }}</span>
              <span class="text-xs text-gray-600 dark:text-gray-400 truncate mt-0.5">{{ notif.message }}</span>
            </li>
          </ul>
        </div>

        <div v-if="acceptedQuotes.length > 0" class="notification-group">
          <div class="text-green-500 font-semibold mb-2 flex justify-between items-center cursor-pointer" @click="goToQuotes">
            <span><i class="pi pi-check-circle mr-2"></i>Kabul Edilen Teklifler ({{ acceptedQuotes.length }})</span>
            <i class="pi pi-chevron-right text-xs"></i>
          </div>
          <ul class="text-sm mt-2 max-h-32 overflow-y-auto flex flex-col gap-2">
            <li v-for="notif in acceptedQuotes" :key="notif.id" class="border-b border-gray-100 dark:border-gray-800 pb-2 flex flex-col cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 p-1 rounded" @click="clickNotification(notif)">
              <span class="font-semibold text-green-500 text-xs">{{ notif.title }}</span>
              <span class="text-xs text-gray-600 dark:text-gray-400 truncate mt-0.5">{{ notif.message }}</span>
            </li>
          </ul>
        </div>

        <div v-if="rejectedQuotes.length > 0" class="notification-group">
          <div class="text-red-500 font-semibold mb-2 flex justify-between items-center cursor-pointer" @click="goToQuotes">
            <span><i class="pi pi-times-circle mr-2"></i>Reddedilen Teklifler ({{ rejectedQuotes.length }})</span>
            <i class="pi pi-chevron-right text-xs"></i>
          </div>
          <ul class="text-sm mt-2 max-h-32 overflow-y-auto flex flex-col gap-2">
            <li v-for="notif in rejectedQuotes" :key="notif.id" class="border-b border-gray-100 dark:border-gray-800 pb-2 flex flex-col cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 p-1 rounded" @click="clickNotification(notif)">
              <span class="font-semibold text-red-500 text-xs">{{ notif.title }}</span>
              <span class="text-xs text-gray-600 dark:text-gray-400 truncate mt-0.5">{{ notif.message }}</span>
            </li>
          </ul>
        </div>

        <div v-if="announcements.length > 0" class="notification-group">
          <div class="text-indigo-500 font-semibold mb-2 flex justify-between items-center cursor-pointer" @click="goToAnnouncements">
            <span><i class="pi pi-megaphone mr-2"></i>Yeni Duyurular ({{ announcements.length }})</span>
            <i class="pi pi-chevron-right text-xs"></i>
          </div>
          <ul class="text-sm mt-2 max-h-32 overflow-y-auto">
            <li v-for="ann in announcements.slice(0, 5)" :key="ann.id" class="border-b border-gray-100 py-1 text-slate-600 dark:text-slate-300 truncate">
              {{ ann.title }}
            </li>
          </ul>
        </div>

        <div v-if="criticalStocks.length > 0" class="notification-group">
          <div class="text-red-500 font-semibold mb-2 flex justify-between items-center cursor-pointer" @click="goToProducts">
            <span><i class="pi pi-exclamation-triangle mr-2"></i>Kritik Stok Uyarısı ({{ criticalStocks.length }})</span>
            <i class="pi pi-chevron-right text-xs"></i>
          </div>
          <div class="text-sm text-gray-600 dark:text-gray-400">
            Aşağıdaki ürünlerin stok miktarı kritik seviyenin altında:
          </div>
          <ul class="text-sm mt-2 max-h-32 overflow-y-auto">
            <li v-for="stock in criticalStocks.slice(0, 5)" :key="stock.id" class="border-b border-gray-100 py-1 flex justify-between">
              <span class="truncate pr-2">{{ stock.name }}</span>
              <span class="font-bold text-red-500 shrink-0">{{ stock.current_stock }} / {{ stock.threshold }}</span>
            </li>
            <li v-if="criticalStocks.length > 5" class="text-xs text-center text-gray-500 pt-1">
              ...ve {{ criticalStocks.length - 5 }} ürün daha
            </li>
          </ul>
        </div>

        <div v-if="repeatExpenses.length > 0" class="notification-group">
          <div class="text-orange-500 font-semibold mb-2 flex justify-between items-center cursor-pointer" @click="goToExpenses">
            <span><i class="pi pi-refresh mr-2"></i>Tekrarlayan Fatura ({{ repeatExpenses.length }})</span>
            <i class="pi pi-chevron-right text-xs"></i>
          </div>
          <div class="text-sm text-gray-600 dark:text-gray-400">
            Aynı kategoride birden fazla fatura girişi tespit edildi:
          </div>
          <ul class="text-sm mt-2 max-h-32 overflow-y-auto">
            <li v-for="(expense, idx) in repeatExpenses.slice(0, 5)" :key="idx" class="border-b border-gray-100 py-1">
              <div class="flex justify-between">
                <span class="truncate pr-2">{{ expense.category_name }}</span>
                <span class="font-bold shrink-0">{{ expense.count }} Fatura</span>
              </div>
              <div class="text-xs text-gray-500">
                Aylık Toplam: {{ expense.total_amount }} {{ expense.currency }}
              </div>
            </li>
            <li v-if="repeatExpenses.length > 5" class="text-xs text-center text-gray-500 pt-1">
              ...ve {{ repeatExpenses.length - 5 }} kategori daha
            </li>
          </ul>
        </div>

        <!-- Hiç bildirim yoksa -->
        <div
          v-if="total === 0"
          class="text-center text-sm text-gray-400 py-4"
        >
          <i class="pi pi-check-circle text-2xl text-green-400 block mb-2"></i>
          Tüm bildirimler temiz!
        </div>
      </div>
    </OverlayPanel>
  </div>
</template>

<style scoped>
.notification-bell {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: visible; /* badge kırpılmasın */
  cursor: pointer;
  transition: background-color 0.2s;
}

.notification-bell:hover {
  background-color: var(--surface-hover);
}

/* Badge: .notification-bell'e göre konumlandırılır */
.notif-badge {
  position: absolute;
  top: 2px;
  right: 2px;
  transform: translate(50%, -50%);
  background-color: #ef4444;
  color: #ffffff;
  font-size: 10px;
  font-weight: 800;
  line-height: 1;
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  border: 2px solid #ffffff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
  z-index: 9999;
  pointer-events: none;
  white-space: nowrap;
}

:root.p-dark .notif-badge {
  border-color: #1e293b;
}

.notification-group:not(:last-child) {
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--surface-border);
}
</style>
