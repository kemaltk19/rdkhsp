<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCurrencyStore } from '@/stores/currency'
import { useSettingsStore } from '@/stores/settings'
import { useBillingStore } from '@/stores/billing'
import { useTheme } from '@/composables/useTheme'
import { useI18n } from 'vue-i18n'
import NotificationBell from '@/components/NotificationBell.vue'

const authStore = useAuthStore()
const currencyStore = useCurrencyStore()
const billingStore = useBillingStore()
const route = useRoute()
const { isDark, toggleTheme } = useTheme()
const { t, locale } = useI18n()

const user = computed(() => authStore.user)
const company = computed(() => authStore.company)

const trialDaysLeft = computed(() => {
  if (!company.value || company.value.subscription_status !== 'trial') return 0
  const end = new Date(company.value.trial_ends_at)
  const now = new Date()
  const diffTime = end.getTime() - now.getTime()
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
  return diffDays > 0 ? diffDays : 0
})

const periodEndWarning = computed(() => {
  const status = billingStore.status
  if (!status || status.subscription_status !== 'active' || !status.current_period_end) return null
  if (status.period_days_left > 15) return null
  return {
    daysLeft: status.period_days_left,
    dateStr: new Date(status.current_period_end).toLocaleDateString('tr-TR')
  }
})

const menuItems = computed(() => {
  const role = user.value?.role || 'admin'
  let items: any[] = []

  if (role === 'superadmin') {
    items = [
      { label: t('nav.dashboard'), icon: 'pi pi-chart-bar', path: '/superadmin/dashboard' },
      { label: t('nav.companies'), icon: 'pi pi-building', path: '/superadmin/companies' },
      { label: t('nav.plans'), icon: 'pi pi-box', path: '/superadmin/plans' },
      { label: t('nav.emailSettings'), icon: 'pi pi-envelope', path: '/superadmin/email-settings' },
      { label: t('nav.announcements'), icon: 'pi pi-megaphone', path: '/superadmin/announcements' }
    ]
  } else if (role === 'cari') {
    items = [
      { label: t('nav.dashboard'), icon: 'pi pi-chart-bar', path: '/dashboard' },
      { label: t('nav.invoices'), icon: 'pi pi-file', path: '/invoices' },
      { label: t('nav.payments'), icon: 'pi pi-money-bill', path: '/payments' },
      { label: t('nav.cash'), icon: 'pi pi-wallet', path: '/cash' },
    ]
  } else if (role === 'personel') {
    items = [
      { label: t('nav.dashboard'), icon: 'pi pi-chart-bar', path: '/dashboard' },
      { label: t('nav.caris'), icon: 'pi pi-users', path: '/caris' },
      { label: t('nav.invoices'), icon: 'pi pi-file', path: '/invoices' },
      { label: t('nav.quotes'), icon: 'pi pi-file-edit', path: '/quotes' },
      { label: t('nav.payments'), icon: 'pi pi-money-bill', path: '/payments' },
      { label: t('nav.cash'), icon: 'pi pi-wallet', path: '/cash' },
      { label: t('nav.expenses'), icon: 'pi pi-shopping-bag', path: '/expenses' },
      { label: t('nav.products'), icon: 'pi pi-box', path: '/products' },
    ]
   } else {
     // admin
     items = [
       { label: t('nav.dashboard'), icon: 'pi pi-chart-bar', path: '/dashboard' },
       { label: t('nav.caris'), icon: 'pi pi-users', path: '/caris' },
       { label: t('nav.invoices'), icon: 'pi pi-file', path: '/invoices' },
       { label: t('nav.quotes'), icon: 'pi pi-file-edit', path: '/quotes' },
       { label: t('nav.projects'), icon: 'pi pi-briefcase', path: '/projects' },
       { label: t('nav.payments'), icon: 'pi pi-money-bill', path: '/payments' },
       { label: t('nav.cash'), icon: 'pi pi-wallet', path: '/cash' },
       { label: t('nav.expenses'), icon: 'pi pi-shopping-bag', path: '/expenses' },
       { label: t('nav.products'), icon: 'pi pi-box', path: '/products' },
       { label: t('nav.employees'), icon: 'pi pi-id-card', path: '/employees' },
       { label: t('nav.reports'), icon: 'pi pi-chart-line', path: '/reports' },
       { label: t('nav.settings'), icon: 'pi pi-cog', path: '/settings' },
     ]
   }

  // Filter based on enabled modules for tenant roles (admin, personel, cari)
  if (role !== 'superadmin' && company.value) {
    let enabledModules: string[] = []
    if (company.value.enabled_modules) {
      try {
        enabledModules = JSON.parse(company.value.enabled_modules)
      } catch {
        enabledModules = []
      }
    }

    // Map menu paths to module identifiers.
    // Dashboard ve Ayarlar kasıtlı olarak burada yok: her zaman sidebar'da
    // görünmeli, aksi halde admin modülleri kapatıp kendini kilitleyebilir.
     const pathMap: Record<string, string> = {
       '/caris': 'caris',
       '/invoices': 'invoices',
       '/quotes': 'quotes',
       '/projects': 'projects',
       '/payments': 'payments',
       '/cash': 'cash',
       '/expenses': 'expenses',
       '/products': 'products',
       '/reports': 'reports',
       '/employees': 'employees'
     }

    if (enabledModules.length > 0) {
      items = items.filter(item => {
        const modKey = pathMap[item.path]
        if (modKey) {
          return enabledModules.includes(modKey)
        }
        return true // keep other items (dashboard, settings) — always visible
      })
    }

    // Personel için kişisel okuma iznine göre süz. /quotes fatura, /cash ise
    // ödeme iznine bağlı olduğundan menü-permission eşlemesi ayrı tutulur.
    if (role === 'personel') {
      const permModuleMap: Record<string, string> = {
         '/caris': 'caris',
         '/invoices': 'invoices',
         '/quotes': 'invoices',
         '/projects': 'projects',
         '/payments': 'payments',
         '/cash': 'payments',
         '/expenses': 'expenses',
         '/products': 'products',
         '/reports': 'reports'
       }
      const perms = authStore.permissions || []
      items = items.filter(item => {
        const modKey = permModuleMap[item.path]
        if (!modKey) return true // dashboard, duyuru vb. her zaman görünür
        return perms.some(p => p.module === modKey && p.can_read)
      })
    }
  }

  // Always append Duyuru Panosu for tenant users (except superadmin who has it in their main menu)
  if (role !== 'superadmin') {
    items.push({ label: 'Duyuru Panosu', icon: 'pi pi-megaphone', path: '/announcements' })
  }

  return items
})

const currentModule = computed(() => {
  const activeItem = menuItems.value
    .slice()
    .sort((a, b) => b.path.length - a.path.length) // match longest path first
    .find(item => route.path.startsWith(item.path))
  return activeItem ? activeItem.label : ''
})

const currentSubtitle = computed(() => {
  if (route.path.startsWith('/caris')) return 'Müşteri ve tedarikçi hesaplarınızı, bakiyelerini ve hareketlerini yönetin.'
  if (route.path.startsWith('/invoices')) return 'Satış ve alış faturalarınızı oluşturun, takip edin ve iptal/iade süreçlerini yönetin.'
  if (route.path.startsWith('/quotes')) return 'Müşterilerinize teklifler ve proforma faturalar hazırlayın, durumlarını izleyin.'
  if (route.path.startsWith('/projects')) return 'Projelerinizi yönetin, bütçe belirleyin ve ilgili faturaları ve teklifleri takip edin.'
  if (route.path.startsWith('/payments')) return 'Müşterilerinizden tahsilat yapın, tedarikçilerinize ödemelerinizi kaydedin.'
  if (route.path.startsWith('/cash')) return 'Kasa ve banka hesaplarınızı yönetin, virman işlemlerini gerçekleştirin.'
  if (route.path.startsWith('/expenses')) return 'İşletme giderlerinizi detaylı olarak takip edin, masraflarınızı kategorize edin.'
  if (route.path.startsWith('/products')) return 'Hizmet ve stok kartlarınızı tanımlayın, fiyatlarını ve miktarlarını yönetin.'
  if (route.path.startsWith('/employees')) return 'Personel bilgilerinizi yönetin, maaş ve avans ödemelerini takip edin.'
  if (route.path.startsWith('/reports')) return 'İşletmenizin gelir, gider ve genel finansal durumunu analiz edin.'
  if (route.path.startsWith('/settings')) return 'Sistem, şirket, fatura, e-posta ve yetkilendirme ayarlarınızı yapılandırın.'
  if (route.path.startsWith('/billing')) return 'Abonelik planınızı, faturalandırma geçmişinizi ve ödeme bilgilerinizi yönetin.'
  return ''
})

const toggleLocale = () => {
  locale.value = locale.value === 'tr' ? 'en' : 'tr'
}

const settingsStore = useSettingsStore()
const timezone = computed(() => {
  return settingsStore.settings['timezone'] || (company.value as any)?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone || 'Europe/Istanbul'
})
const currentDateTimeStr = ref('')

const updateTime = () => {
  try {
    const now = new Date()
    currentDateTimeStr.value = now.toLocaleString('tr-TR', {
      timeZone: timezone.value,
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch (e) {
    // fallback if timezone invalid
    const now = new Date()
    currentDateTimeStr.value = now.toLocaleString('tr-TR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  }
}

let timerId: any = null

onMounted(async () => {
  // Load timezone setting initially to populate the store
  settingsStore.fetchSetting('timezone')

  updateTime()
  timerId = setInterval(updateTime, 1000)
  currencyStore.fetchCurrencies()
  if (user.value?.role !== 'superadmin') {
    billingStore.fetchBillingStatus()
  }
})

onUnmounted(() => {
  if (timerId) clearInterval(timerId)
})

const handleLogout = async () => {
  await authStore.logout()
}
</script>

<template>
  <div class="app-container">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <span class="logo-icon">▲</span>
        <span class="brand-text">Radikal Hesap</span>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="nav-link"
          :class="{ active: route.path.startsWith(item.path) }"
        >
          <i :class="item.icon" class="nav-icon"></i>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer" v-if="user">
        <div v-if="company && company.subscription_status === 'trial'" class="sidebar-trial-wrapper">
          <span class="trial-tag">
            <i class="pi pi-clock mr-1"></i>
            {{ t('common.trialPeriod') }}: {{ t('common.trialDaysLeft', { days: trialDaysLeft }) }}
          </span>
        </div>
        <div class="user-info">
          <div class="user-avatar">{{ user.name.charAt(0) }}</div>
          <div class="user-details">
            <span class="user-name">{{ user.name }}</span>
            <span class="user-role">{{ user.role.toUpperCase() }}</span>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <div class="main-layout">
      <!-- Header -->
      <header class="header">
        <div class="header-left">
          <div class="flex flex-col">
            <h1 class="current-module-title">{{ currentModule }}</h1>
            <span class="current-module-subtitle" v-if="currentSubtitle">{{ currentSubtitle }}</span>
          </div>
        </div>

        <div class="header-right">
          <!-- Company Name -->
          <div class="company-badge-right" v-if="company" :title="company.name">
            <i class="pi pi-building company-icon shrink-0"></i>
            <span class="company-name-text truncate max-w-[180px] block">{{ company.name }}</span>
          </div>

          <!-- Subscription Period End Warning -->
          <router-link
            v-if="periodEndWarning"
            to="/settings?tab=billing"
            class="period-end-warning"
            :title="periodEndWarning.daysLeft > 0 ? `Aboneliğiniz ${periodEndWarning.daysLeft} gün sonra sona eriyor` : 'Aboneliğinizin süresi doldu'"
          >
            <i class="pi pi-clock"></i>
            <span v-if="periodEndWarning.daysLeft > 0">Aboneliğiniz {{ periodEndWarning.daysLeft }} gün sonra sona eriyor ({{ periodEndWarning.dateStr }})</span>
            <span v-else>Aboneliğinizin süresi doldu ({{ periodEndWarning.dateStr }})</span>
          </router-link>

          <!-- Divider -->
          <div class="divider"></div>

          <!-- Date Time Clock -->
          <div class="header-clock hidden md:flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-800/40 text-xs font-semibold text-slate-600 dark:text-slate-300">
            <i class="pi pi-clock text-indigo-500"></i>
            <span>{{ currentDateTimeStr }}</span>
          </div>

          <!-- Language Toggle -->
          <button class="icon-btn" @click="toggleLocale" :title="locale === 'tr' ? 'English' : 'Türkçe'">
            <span class="text-sm font-bold uppercase">{{ locale }}</span>
          </button>

          <!-- Notification Bell -->
          <NotificationBell />

          <!-- Theme Toggle -->
          <button class="icon-btn" @click="toggleTheme" title="Tema Değiştir">
            <i :class="isDark ? 'pi pi-sun' : 'pi pi-moon'"></i>
          </button>

          <!-- Divider -->
          <div class="divider"></div>

          <!-- Logout Button -->
          <button class="logout-btn" @click="handleLogout">
            <i class="pi pi-power-off"></i>
            <span>{{ t('auth.logout') }}</span>
          </button>
        </div>
      </header>

      <!-- Page Content -->
      <main class="content">
        <div class="fade-in-container">
          <router-view />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background-color: var(--p-content-background, #f8fafc);
  color: var(--p-text-color, #0f172a);
}

:root.p-dark .app-container {
  background-color: #0f172a;
}

.sidebar {
  width: 260px;
  background-color: #0f172a;
  color: #f8fafc;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.sidebar-header {
  height: 70px;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.logo-icon {
  font-size: 1.5rem;
  color: #06b6d4;
}

.brand-text {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.025em;
}

.sidebar-nav {
  flex: 1;
  padding: 1rem 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  overflow-y: auto;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.85rem 1.5rem;
  color: #94a3b8;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
  text-decoration: none;
  font-weight: 500;
  transition: all 0.2s;
  position: relative;
}

.nav-link:hover {
  background-color: rgba(255, 255, 255, 0.03);
  color: #f8fafc;
}

.nav-link.active {
  background-color: rgba(6, 182, 212, 0.1);
  color: #06b6d4;
  border-left: 3px solid #06b6d4;
  padding-left: calc(1.5rem - 3px);
}

.nav-icon {
  font-size: 1.1rem;
}

.sidebar-footer {
  padding: 1.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  background-color: rgba(0, 0, 0, 0.15);
}

.sidebar-trial-wrapper {
  margin-bottom: 1rem;
  display: flex;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: #22d3ee;
  color: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.1rem;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: #f8fafc;
}

.user-role {
  font-size: 0.7rem;
  font-weight: 700;
  color: #22d3ee;
}

.main-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 100vh;
}

.header {
  height: 70px;
  background-color: var(--p-content-background, #ffffff);
  border-bottom: 1px solid var(--p-content-border-color, #e2e8f0);
  display: flex;
  align-items: center;
  padding: 0 1.5rem;
  z-index: 10;
  transition: background-color 0.3s, border-color 0.3s;
}

:root.p-dark .header {
  background-color: #1e293b;
  border-color: #334155;
}

.header-left {
  flex: 1;
  display: flex;
  align-items: center;
}

.current-module-title {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--p-text-color, #0f172a);
  margin: 0;
  line-height: 1.2;
}

.current-module-subtitle {
  font-size: 0.75rem;
  color: var(--p-text-muted-color, #64748b);
  margin-top: 0.15rem;
}

:root.p-dark .current-module-title {
  color: #f8fafc;
}

:root.p-dark .current-module-subtitle {
  color: #94a3b8;
}

.header-right {
  flex: 2;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 1rem;
}

.company-badge-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0 0.5rem;
}

.company-icon {
  font-size: 1.2rem;
  color: var(--p-primary-500, #06b6d4);
}

.company-name-text {
  font-size: 1.1rem;
  font-weight: 800;
  color: var(--p-text-color, #0f172a);
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

:root.p-dark .company-name-text {
  color: #f8fafc;
}

.period-end-warning {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #dc2626;
  background-color: rgba(220, 38, 38, 0.08);
  border: 1px solid rgba(220, 38, 38, 0.2);
  text-decoration: none;
  white-space: nowrap;
  transition: background-color 0.2s;
}

.period-end-warning:hover {
  background-color: rgba(220, 38, 38, 0.14);
}

:root.p-dark .period-end-warning {
  color: #f87171;
  background-color: rgba(248, 113, 113, 0.1);
  border-color: rgba(248, 113, 113, 0.25);
}

.trial-tag {
  background-color: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.icon-btn {
  background: none;
  border: none;
  color: var(--p-text-muted-color, #64748b);
  cursor: pointer;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s, color 0.2s;
}

.icon-btn:hover {
  background-color: var(--p-content-hover-background, #f1f5f9);
  color: var(--p-text-color, #0f172a);
}

:root.p-dark .icon-btn:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.divider {
  width: 1px;
  height: 24px;
  background-color: var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .divider {
  background-color: #334155;
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: none;
  border: none;
  color: #ef4444;
  font-weight: 600;
  cursor: pointer;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.logout-btn:hover {
  background-color: rgba(239, 68, 68, 0.08);
}

.content {
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
}

.fade-in-container {
  animation: fadeIn 0.4s ease-out;
  width: 100%;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
