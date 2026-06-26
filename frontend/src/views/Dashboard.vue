<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useDashboardStore } from '@/stores/dashboard'
import { useI18n } from 'vue-i18n'
import Card from 'primevue/card'
import Tag from 'primevue/tag'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import Money from '@/components/Money.vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'

const authStore = useAuthStore()
const dashboardStore = useDashboardStore()
const { t } = useI18n()

const user = computed(() => authStore.user)
const company = computed(() => authStore.company)
const stats = computed(() => dashboardStore.stats)

const formattedTrialEndDate = computed(() => {
  if (!company.value || !company.value.trial_ends_at) return ''
  return new Date(company.value.trial_ends_at).toLocaleDateString('tr-TR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
})

const showWelcome = ref(false)

onMounted(async () => {
  await dashboardStore.fetchStats()
  if (sessionStorage.getItem('show_welcome') === '1') {
    showWelcome.value = true
    sessionStorage.removeItem('show_welcome')
  }
})

const getMonthName = (monthStr: string) => {
  const [year, month] = monthStr.split('-')
  const date = new Date(parseInt(year), parseInt(month) - 1, 1)
  return date.toLocaleDateString('tr-TR', { month: 'short' })
}

// Custom SVG Chart calculations
const maxValue = computed(() => {
  if (!stats.value || !stats.value.chart_series || stats.value.chart_series.length === 0) return 100
  let max = 0
  stats.value.chart_series.forEach(series => {
    series.data.forEach(d => {
      const val = parseFloat(d.total) || 0
      if (val > max) max = val
    })
  })
  return max > 0 ? max : 100
})

const getCurrencyColor = (currency: string) => {
  const colors: Record<string, string> = {
    'TRY': '#0ea5e9', // sky-500
    'USD': '#10b981', // emerald-500
    'EUR': '#f59e0b', // amber-500
    'GBP': '#8b5cf6', // violet-500
  }
  return colors[currency] || '#64748b' // slate-500 fallback
}

const seriesPoints = computed(() => {
  if (!stats.value || !stats.value.chart_series || stats.value.chart_series.length === 0) return []
  const max = maxValue.value

  return stats.value.chart_series.map(series => {
    const points = series.data.map((d, index) => {
      const val = parseFloat(d.total) || 0
      const x = (index / 5) * 440 + 60 // range 60 to 500
      const y = 200 - (val / max) * 150 // range 50 to 200
      return {
        x,
        y,
        val,
        month: getMonthName(d.month),
        fullName: d.month
      }
    })
    
    const linePath = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
    const fillPath = points.length > 0 ? `${linePath} L ${points[points.length - 1].x} 200 L ${points[0].x} 200 Z` : ''

    return {
      currency: series.currency,
      color: getCurrencyColor(series.currency),
      points,
      linePath,
      fillPath
    }
  })
})

const yGridLines = computed(() => {
  const max = maxValue.value
  return [
    { label: (max).toFixed(0), y: 50 },
    { label: (max * 0.75).toFixed(0), y: 87.5 },
    { label: (max * 0.5).toFixed(0), y: 125 },
    { label: (max * 0.25).toFixed(0), y: 162.5 },
    { label: '0', y: 200 }
  ]
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('tr-TR')
}

const getCariTxSeverity = (type: string) => {
  return type === 'debit' ? 'success' : 'danger'
}

const getCariTxLabel = (type: string) => {
  return type === 'debit' ? 'Borç' : 'Alacak'
}

const getCashTxSeverity = (type: string) => {
  return type === 'in' ? 'success' : 'danger'
}

const getCashTxLabel = (type: string) => {
  return type === 'in' ? 'Giriş' : 'Çıkış'
}

const truncateString = (str: string, len: number = 10) => {
  if (!str) return ''
  if (str.length <= len) return str
  return str.substring(0, len) + '...'
}
</script>

<template>
  <div class="dashboard-container">

    <!-- Karsilama Pop-up (yeni kayit sonrasi) -->
    <Dialog v-model:visible="showWelcome" modal :closable="true" :draggable="false"
      class="welcome-dialog" :style="{ width: '480px' }" header=" ">
      <div class="welcome-content">
        <div class="welcome-icon"><i class="pi pi-check-circle"></i></div>
        <h2 class="welcome-h">Hos geldiniz{{ user ? ', ' + user.name : '' }}!</h2>
        <p class="welcome-p">
          <strong>{{ company ? company.name : 'Firmaniz' }}</strong> basariyla olusturuldu.
          Firma bilgilerinizi, logonuzu, vergi dairesi ve para birimi gibi ayarlarinizi
          dilediginiz zaman <strong>Ayarlar</strong> menusunden duzenleyebilirsiniz.
        </p>
        <div class="welcome-actions">
          <Button label="Ayarlara Git" icon="pi pi-cog" @click="$router.push('/settings'); showWelcome = false" />
          <Button label="Panele Devam Et" icon="pi pi-arrow-right" text @click="showWelcome = false" />
        </div>
      </div>
    </Dialog>


    <!-- Stats Grid -->
    <div class="stats-grid" v-if="stats">
      <Card class="stat-card">
        <template #content>
          <div class="card-inner">
            <div class="icon-wrapper green">
              <i class="pi pi-chart-line"></i>
            </div>
            <div class="stat-info">
              <span class="stat-label">Bu Ayki Ciro</span>
              <div v-if="stats.ciro && stats.ciro.length > 0">
                <div v-for="item in stats.ciro" :key="item.currency" class="stat-value text-green-600 dark:text-green-400">
                  <Money :value="item.amount" :currency="item.currency" />
                </div>
              </div>
              <div v-else class="stat-value text-green-600 dark:text-green-400">
                <Money :value="0" currency="TRY" />
              </div>
            </div>
          </div>
        </template>
      </Card>

      <Card class="stat-card">
        <template #content>
          <div class="card-inner">
            <div class="icon-wrapper blue">
              <i class="pi pi-inbox"></i>
            </div>
            <div class="stat-info">
              <span class="stat-label">Tahsil Edilecek</span>
              <div v-if="stats.to_collect && stats.to_collect.length > 0">
                <div v-for="item in stats.to_collect" :key="item.currency" class="stat-value text-sky-600 dark:text-sky-400">
                  <Money :value="item.amount" :currency="item.currency" />
                </div>
              </div>
              <div v-else class="stat-value text-sky-600 dark:text-sky-400">
                <Money :value="0" currency="TRY" />
              </div>
            </div>
          </div>
        </template>
      </Card>

      <Card class="stat-card">
        <template #content>
          <div class="card-inner">
            <div class="icon-wrapper orange">
              <i class="pi pi-wallet"></i>
            </div>
            <div class="stat-info">
              <span class="stat-label">Kasa & Banka Toplam</span>
              <div v-if="stats.cash_bank_total && stats.cash_bank_total.length > 0">
                <div v-for="item in stats.cash_bank_total" :key="item.currency" class="stat-value text-amber-600 dark:text-amber-400">
                  <Money :value="item.amount" :currency="item.currency" />
                </div>
              </div>
              <div v-else class="stat-value text-amber-600 dark:text-amber-400">
                <Money :value="0" currency="TRY" />
              </div>
            </div>
          </div>
        </template>
      </Card>

      <Card class="stat-card">
        <template #content>
          <div class="card-inner">
            <div class="icon-wrapper red">
              <i class="pi pi-exclamation-circle"></i>
            </div>
            <div class="stat-info">
              <span class="stat-label">Vadesi Geçen Alacak</span>
              <div v-if="stats.overdue_total && stats.overdue_total.length > 0">
                <div v-for="item in stats.overdue_total" :key="item.currency" class="stat-value text-red-600 dark:text-red-400">
                  <Money :value="item.amount" :currency="item.currency" />
                </div>
              </div>
              <div v-else class="stat-value text-red-600 dark:text-red-400">
                <Money :value="0" currency="TRY" />
              </div>
            </div>
          </div>
        </template>
      </Card>
    </div>

    <!-- Loading State -->
    <div v-if="dashboardStore.loading" class="text-center py-8">
      <i class="pi pi-spin pi-spinner text-3rem text-primary"></i>
    </div>

    <!-- Main Content Panel (Graph & Activities) -->
    <div class="content-grid mt-4" v-else-if="stats">
      <!-- Sales Chart -->
      <Card class="chart-card col-span-12 lg:col-span-7">
        <template #title>
          <div class="chart-header">
            <h2 class="card-title font-bold text-slate-800 dark:text-slate-100">Satış Performansı (Son 6 Ay)</h2>
          </div>
        </template>
        <template #content>
          <div class="chart-wrapper">
            <svg viewBox="0 0 540 240" class="w-full h-full custom-chart-svg">
              <!-- Definitions for Gradients -->
              <defs>
                <linearGradient id="chartGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#0284c7" stop-opacity="0.25" />
                  <stop offset="100%" stop-color="#0284c7" stop-opacity="0.0" />
                </linearGradient>
              </defs>

              <!-- Grid Lines -->
              <g v-for="grid in yGridLines" :key="grid.y">
                <line x1="50" :y1="grid.y" x2="510" :y2="grid.y" stroke="currentColor" stroke-dasharray="4" class="text-slate-200 dark:text-slate-700/50" />
                <text x="40" :y="grid.y + 4" fill="currentColor" font-size="10" text-anchor="end" class="text-slate-400 font-medium">{{ grid.label }}</text>
              </g>

              <g v-for="series in seriesPoints" :key="series.currency">
                <!-- Area Fill -->
                <path :d="series.fillPath" :fill="series.color" fill-opacity="0.1" />

                <!-- Line -->
                <path :d="series.linePath" fill="none" :stroke="series.color" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="chart-line" />

                <!-- Points -->
                <g v-for="(point, i) in series.points" :key="'pt-' + i" class="chart-point-group">
                  <circle :cx="point.x" :cy="point.y" r="4" :fill="series.color" stroke="white" stroke-width="2" class="chart-point" />
                  
                  <!-- Tooltip (CSS Hover) -->
                  <g class="chart-tooltip">
                    <rect :x="point.x - 40" :y="point.y - 35" width="80" height="24" rx="4" fill="#1e293b" />
                    <text :x="point.x" :y="point.y - 18" fill="white" font-size="11" font-weight="bold" text-anchor="middle">
                      {{ point.val.toLocaleString() }} {{ series.currency }}
                    </text>
                  </g>
                </g>
              </g>

              <!-- X Axis Labels (just use first series for X axis) -->
              <g v-if="seriesPoints.length > 0">
                <text v-for="(point, i) in seriesPoints[0].points" :key="'label-' + i" :x="point.x" y="220" fill="currentColor" font-size="11" text-anchor="middle" class="text-slate-400 font-medium">
                  {{ point.month }}
                </text>
              </g>
            </svg>
            
            <!-- Legend -->
            <div class="flex gap-4 justify-center mt-4">
              <div v-for="series in seriesPoints" :key="series.currency" class="flex items-center gap-2">
                <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: series.color }"></div>
                <span class="text-sm font-medium text-slate-600 dark:text-slate-300">{{ series.currency }}</span>
              </div>
            </div>
          </div>
        </template>
      </Card>

      <!-- Recent Transactions Tabs -->
      <Card class="activities-card col-span-12 lg:col-span-5">
        <template #title>
          <h2 class="card-title font-bold text-slate-800 dark:text-slate-100">Son İşlemler</h2>
        </template>
        <template #content>
          <Tabs value="0">
            <TabList>
              <Tab value="0">Cari Hareketler</Tab>
              <Tab value="1">Kasa & Banka</Tab>
              <Tab value="2">Giderler</Tab>
            </TabList>
            <TabPanels class="pt-3">
              <!-- Cari Tab -->
              <TabPanel value="0">
                <div class="activity-list" v-if="stats.recent_cari_tx.length > 0">
                  <div v-for="tx in stats.recent_cari_tx" :key="tx.id" class="activity-item">
                    <div class="item-main">
                      <div class="item-title font-bold text-slate-800 dark:text-slate-100 truncate" :title="tx.cari_name">{{ truncateString(tx.cari_name, 10) }}</div>
                      <div class="item-desc text-xs text-slate-500 truncate" :title="tx.description">{{ tx.description }}</div>
                      <div class="item-date text-xxs text-slate-400 mt-0.5">{{ formatDate(tx.date) }}</div>
                    </div>
                    <div class="item-side text-right">
                      <div class="item-amount font-bold">
                        <Money :value="tx.amount" />
                      </div>
                      <Tag :value="getCariTxLabel(tx.type)" :severity="getCariTxSeverity(tx.type)" class="text-xxs mt-0.5" />
                    </div>
                  </div>
                </div>
                <div v-else class="empty-list-state">Cari hareket bulunamadı.</div>
              </TabPanel>

              <!-- Cash Tab -->
              <TabPanel value="1">
                <div class="activity-list" v-if="stats.recent_cash_tx.length > 0">
                  <div v-for="tx in stats.recent_cash_tx" :key="tx.id" class="activity-item">
                    <div class="item-main">
                      <div class="item-title font-bold text-slate-800 dark:text-slate-100 truncate" :title="tx.account_name">{{ truncateString(tx.account_name, 10) }}</div>
                      <div class="item-desc text-xs text-slate-500 truncate" :title="tx.description">{{ tx.description }}</div>
                      <div class="item-date text-xxs text-slate-400 mt-0.5">{{ formatDate(tx.date) }}</div>
                    </div>
                    <div class="item-side text-right">
                      <div class="item-amount font-bold">
                        <Money :value="tx.amount" />
                      </div>
                      <Tag :value="getCashTxLabel(tx.type)" :severity="getCashTxSeverity(tx.type)" class="text-xxs mt-0.5" />
                    </div>
                  </div>
                </div>
                <div v-else class="empty-list-state">Kasa/Banka hareketi bulunamadı.</div>
              </TabPanel>

              <!-- Expenses Tab -->
              <TabPanel value="2">
                <div class="activity-list" v-if="stats.recent_expenses && stats.recent_expenses.length > 0">
                  <div v-for="tx in stats.recent_expenses" :key="tx.id" class="activity-item">
                    <div class="item-main">
                      <div class="item-title font-bold text-slate-800 dark:text-slate-100 truncate" :title="tx.category_name">{{ truncateString(tx.category_name, 10) }}</div>
                      <div class="item-desc text-xs text-slate-500 truncate" :title="tx.description">{{ tx.description }}</div>
                      <div class="item-date text-xxs text-slate-400 mt-0.5">{{ formatDate(tx.date) }}</div>
                    </div>
                    <div class="item-side text-right">
                      <div class="item-amount font-bold">
                        <Money :value="tx.amount" :currency="tx.currency" />
                      </div>
                      <Tag value="Gider" severity="warning" class="text-xxs mt-0.5" />
                    </div>
                  </div>
                </div>
                <div v-else class="empty-list-state">Gider hareketi bulunamadı.</div>
              </TabPanel>
            </TabPanels>
          </Tabs>
        </template>
      </Card>
    </div>
  </div>
</template>

<style scoped>
.dashboard-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.welcome-banner {
  margin-bottom: 0.5rem;
}

.welcome-title {
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin-bottom: 0.25rem;
}

.welcome-desc {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.95rem;
}

.trial-banner {
  background-color: rgba(245, 158, 11, 0.08);
  border-left: 4px solid #f59e0b;
  color: #b45309;
  padding: 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
}

.trial-banner-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.banner-icon {
  font-size: 1.25rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.5rem;
}

.stat-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  background: var(--p-content-background, #ffffff);
  transition: transform 0.2s, box-shadow 0.2s, border-color 0.3s, background-color 0.3s;
}

:root.p-dark .stat-card {
  background: #1e293b;
  border-color: #334155;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05);
}

.card-inner {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.icon-wrapper.blue {
  background-color: rgba(14, 165, 233, 0.1);
  color: #0ea5e9;
}

.icon-wrapper.green {
  background-color: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.icon-wrapper.orange {
  background-color: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.icon-wrapper.red {
  background-color: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-label {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.85rem;
  font-weight: 500;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin-bottom: 0.2rem;
}

.content-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 1.5rem;
}

.chart-card, .activities-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .chart-card, :root.p-dark .activities-card {
  background: #1e293b;
  border-color: #334155;
}

.card-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--p-text-color, #1e293b);
  padding: 0.25rem 0.5rem;
}

:root.p-dark .card-title {
  color: #f8fafc;
}

.chart-wrapper {
  position: relative;
  height: 280px;
  padding: 1rem 0;
}

.custom-chart-svg {
  display: block;
}

.chart-dot {
  cursor: pointer;
  transition: r 0.2s, stroke-width 0.2s;
}

.chart-dot:hover {
  r: 8px;
  stroke-width: 3px;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.activity-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.75rem;
  background-color: rgba(148, 163, 184, 0.05);
  border-radius: 8px;
}

.item-main {
  flex: 1;
  min-width: 0;
  margin-right: 1.5rem;
}

.item-side {
  flex-shrink: 0;
}

.item-title {
  font-size: 0.875rem;
}

.item-date {
  font-size: 0.75rem;
}

.text-xxs {
  font-size: 0.65rem;
}

.text-right {
  text-align: right;
}

.empty-list-state {
  text-align: center;
  color: var(--p-text-muted-color, #94a3b8);
  font-size: 0.875rem;
  padding: 2rem 1rem;
}

.mt-4 {
  margin-top: 1rem;
}

.pt-3 {
  padding-top: 0.75rem;
}

.col-span-12 {
  grid-column: span 12;
}

@media (min-width: 1024px) {
  .lg\:col-span-7 {
    grid-column: span 7;
  }
  .lg\:col-span-5 {
    grid-column: span 5;
  }
}

.welcome-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 0.5rem 1rem 1rem;
  gap: 0.75rem;
}
.welcome-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: rgba(6, 182, 212, 0.12);
  color: #06b6d4;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
}
.welcome-h {
  font-size: 1.4rem;
  font-weight: 700;
  margin: 0;
}
.welcome-p {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.95rem;
  line-height: 1.55;
  margin: 0;
}
.welcome-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.5rem;
  flex-wrap: wrap;
  justify-content: center;
}
</style>
