<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { useCurrencyStore } from '@/stores/currency'
import { useBillingStore } from '@/stores/billing'
import { useAuthStore } from '@/stores/auth'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import PhoneInput from '@/components/PhoneInput.vue'
import { countries } from '@/constants/countries'
import { sectors } from '@/constants/sectors'

import AccountsTab from './components/AccountsTab.vue'
import KdvRatesTab from './components/KdvRatesTab.vue'
import CurrenciesTab from './components/CurrenciesTab.vue'
import ImportTab from './components/ImportTab.vue'
import ModulesTab from './components/ModulesTab.vue'

const settingsStore = useSettingsStore()
const currencyStore = useCurrencyStore()
const billingStore = useBillingStore()
const authStore = useAuthStore()
const toast = useToast()
const route = useRoute()
const router = useRouter()

const activeTab = ref('profile')

const currencies = computed(() => {
  if (currencyStore.currencies.length > 0) {
    return currencyStore.currencies.map(c => ({
      label: `${c.name} (${c.code})`,
      value: c.code,
    }))
  }
  return [
    { label: 'Türk Lirası (TRY)', value: 'TRY' },
    { label: 'Amerikan Doları (USD)', value: 'USD' },
    { label: 'Euro (EUR)', value: 'EUR' },
  ]
})

const locales = ref([
  { label: 'Türkçe', value: 'tr' },
  { label: 'English', value: 'en' },
])

const getOffsetStr = (tz: string) => {
  try {
    const d = new Date()
    const utcDate = new Date(d.toLocaleString('en-US', { timeZone: 'UTC' }))
    const tzDate = new Date(d.toLocaleString('en-US', { timeZone: tz }))
    const diff = (tzDate.getTime() - utcDate.getTime()) / 60000
    const sign = diff >= 0 ? '+' : '-'
    const hours = Math.floor(Math.abs(diff) / 60).toString().padStart(2, '0')
    const mins = (Math.abs(diff) % 60).toString().padStart(2, '0')
    return `UTC${sign}${hours}:${mins}`
  } catch {
    return 'UTC'
  }
}

const timezoneOptions = ref(
  (Intl as any).supportedValuesOf('timeZone')
    .map((tz: string) => ({
      label: `(${getOffsetStr(tz)}) ${tz.replace(/_/g, ' ')}`,
      value: tz,
    }))
    .sort((a: { label: string }, b: { label: string }) => a.label.localeCompare(b.label))
)

const companyForm = ref({
  name: '',
  contact_name: '',
  email: '',
  phone: '',
  landline: '',
  fax: '',
  industry: '',
  country: '',
  city: '',
  district: '',
  tax_office: '',
  tax_number: '',
  address: '',
  logo_url: '',
  currency: 'TRY',
  locale: 'tr',
  timezone: 'Europe/Istanbul',
})

const passwordForm = ref({
  old_password: '',
  new_password: '',
  new_password_confirm: '',
})

// ------- Sidebar nav -------
interface NavItem { id: string; label: string; icon: string }
interface NavGroup { groupLabel: string; groupIcon: string; items: NavItem[] }

const settingsNav: NavGroup[] = [
  {
    groupLabel: 'Firma & Sistem',
    groupIcon: 'pi pi-building',
    items: [
      { id: 'profile',        label: 'Firma Profili',        icon: 'pi pi-id-card' },
      { id: 'modules',        label: 'Modül Yönetimi',       icon: 'pi pi-check-square' },
      { id: 'finance',        label: 'Sistem Parametreleri', icon: 'pi pi-sliders-h' },
      { id: 'modul_ayarlari', label: 'Modül Ayarları',       icon: 'pi pi-th-large' },
    ],
  },
  {
    groupLabel: 'Kasa & Banka',
    groupIcon: 'pi pi-wallet',
    items: [
      { id: 'accounts', label: 'Kasa / Banka Tanımları', icon: 'pi pi-wallet' },
    ],
  },
  {
    groupLabel: 'Veri Yönetimi',
    groupIcon: 'pi pi-database',
    items: [
      { id: 'import', label: 'İçe Aktar', icon: 'pi pi-cloud-upload' },
    ],
  },
  {
    groupLabel: 'Abonelik',
    groupIcon: 'pi pi-crown',
    items: [
      { id: 'billing', label: 'Abonelik / Faturalandırma', icon: 'pi pi-crown' },
    ],
  },
]

const handleNavClick = (id: string) => {
  if (id === 'modul_ayarlari') {
    router.push('/settings/modules')
    return
  }
  activeTab.value = id
}

// ---------------------------

const loadSettings = async () => {
  try {
    const comp = await settingsStore.fetchCompanyProfile()
    if (comp) {
      companyForm.value = {
        name: comp.name || '',
        contact_name: comp.contact_name || '',
        email: comp.email || '',
        phone: comp.phone || '',
        landline: comp.landline || '',
        fax: comp.fax || '',
        industry: comp.industry || '',
        country: comp.country || '',
        city: comp.city || '',
        district: comp.district || '',
        tax_office: comp.tax_office || '',
        tax_number: comp.tax_number || '',
        address: comp.address || '',
        logo_url: comp.logo_url || '',
        currency: comp.currency || 'TRY',
        locale: comp.locale || 'tr',
        timezone: comp.timezone || 'Europe/Istanbul',
      }
    }
    await currencyStore.fetchCurrencies()
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Ayarlar yüklenemedi.', life: 10000 })
  }
}

onMounted(async () => {
  loadSettings()
  if (route.query.tab) activeTab.value = route.query.tab as string
  await loadBillingData()
  await checkUrlParams()
})

const saveCompanyProfile = async () => {
  if (!companyForm.value.name) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Firma adı girmek zorunludur.', life: 10000 })
    return
  }
  try {
    const payload = { ...companyForm.value }
    if (payload.name) payload.name = payload.name.toUpperCase()
    if (payload.contact_name) payload.contact_name = payload.contact_name.toUpperCase()
    if (payload.tax_office) payload.tax_office = payload.tax_office.toUpperCase()
    if (payload.city) payload.city = payload.city.toUpperCase()
    if (payload.district) payload.district = payload.district.toUpperCase()

    const updatedComp = await settingsStore.updateCompanyProfile(payload)
    if (authStore.company) authStore.company = { ...authStore.company, ...updatedComp }
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Firma profili güncellendi.', life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Firma profili güncellenemedi.', life: 10000 })
  }
}

const isPasswordLoading = ref(false)
const changePassword = async () => {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password || !passwordForm.value.new_password_confirm) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Tüm şifre alanlarını doldurunuz.', life: 10000 })
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.new_password_confirm) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Yeni şifreler eşleşmiyor.', life: 10000 })
    return
  }
  isPasswordLoading.value = true
  try {
    await authStore.changePassword({
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password
    })
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Şifreniz başarıyla değiştirildi.', life: 10000 })
    passwordForm.value.old_password = ''
    passwordForm.value.new_password = ''
    passwordForm.value.new_password_confirm = ''
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err?.response?.data?.message || 'Şifre değiştirilemedi.', life: 10000 })
  } finally {
    isPasswordLoading.value = false
  }
}

// --- Billing ---
const displaySuccessDialog = ref(false)
const renewDialog = ref(false)
const selectedPeriod = ref('monthly')

const loadBillingData = async () => {
  try {
    await billingStore.fetchBillingStatus()
    await billingStore.fetchPlans()
    await billingStore.fetchTransactions()
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Abonelik bilgileri yüklenemedi.', life: 10000 })
  }
}

const checkUrlParams = async () => {
  if (route.query.success === 'direct' || route.query.success === 'renew' || route.query.success === 'true') {
    activeTab.value = 'billing'
  }
  if (route.query.success === 'direct' || route.query.success === 'renew') {
    displaySuccessDialog.value = true
    await loadBillingData()
    router.replace({ path: route.path, query: { tab: 'billing' } })
  } else if (route.query.success === 'true' && route.query.session_id && route.query.plan_id) {
    const sessID = route.query.session_id as string
    const planID = route.query.plan_id as string
    try {
      const compID = authStore.company?.id
      if (compID) {
        await billingStore.triggerWebhookSimulation(compID, planID, sessID)
        displaySuccessDialog.value = true
        router.replace({ path: route.path, query: { tab: 'billing' } })
      }
    } catch {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Ödeme simülasyonu başarısız oldu.', life: 10000 })
    }
  }
}

const handleSubscribe = async (planID: string, periodType: string) => {
  try {
    const res = await billingStore.subscribe(planID, periodType)
    if (res && res.checkout_url) window.location.href = res.checkout_url
  } catch (err: any) {
    const code = err?.response?.data?.error?.message
    const detail = code === 'plan_change_not_allowed_yet'
      ? 'Bu pakete geçiş, mevcut paketinizin bitiş tarihine 15 gün kala açılır.'
      : 'Abonelik başlatılamadı.'
    toast.add({ severity: 'error', summary: 'Hata', detail, life: 10000 })
  }
}

const canRenew = computed(() => billingStore.status?.can_renew === true)

const openRenewDialog = () => {
  if (!canRenew.value) return
  selectedPeriod.value = 'monthly'
  renewDialog.value = true
}

const handleRenew = async () => {
  try {
    await billingStore.renew(selectedPeriod.value)
    renewDialog.value = false
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Aboneliğiniz başarıyla uzatıldı.', life: 10000 })
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Abonelik uzatılamadı.', life: 10000 })
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('tr-TR')
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('tr-TR')
}

const statusText = computed(() => {
  if (!billingStore.status) return '-'
  const st = billingStore.status.subscription_status
  if (st === 'trial') return '14 Günlük Deneme Süresi'
  if (st === 'active') return 'Aktif Abonelik'
  if (st === 'canceled') return 'İptal Edilmiş / Süresi Dolmuş'
  return st
})

const activePlan = computed(() => {
  if (!billingStore.status || !billingStore.plans) return null
  return billingStore.plans.find(p => p.code === billingStore.status.plan_code)
})

const activePlanMonthlyPrice = computed(() => {
  if (!activePlan.value) return '-'
  return `${parseFloat(activePlan.value.price_monthly)} ${activePlan.value.currency}`
})

const activePlanYearlyPrice = computed(() => {
  if (!activePlan.value) return '-'
  return `${parseFloat(activePlan.value.price_yearly)} ${activePlan.value.currency}`
})

const isPlanSelectable = (plan: any) => {
  if (!billingStore.status || billingStore.status.subscription_status !== 'active' || !activePlan.value) return true
  if (canRenew.value) return true
  return parseFloat(plan.price_monthly) > parseFloat(activePlan.value.price_monthly)
}

const currencySymbols: Record<string, string> = {
  TRY: '₺', USD: '$', EUR: '€', GBP: '£', RUB: '₽',
}

const formatPrice = (value: string | number, currency: string) => {
  const num = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(num)) return '-'
  const formatted = new Intl.NumberFormat('tr-TR', { maximumFractionDigits: 0 }).format(num)
  const sym = currencySymbols[currency] || ''
  return sym ? `${sym}${formatted}` : `${formatted} ${currency}`
}

const featureLabels: Record<string, string> = {
  dashboard: 'Kontrol Paneli',
  caris: 'Cari Hesaplar',
  invoices: 'Fatura Yönetimi',
  quotes: 'Teklif Yönetimi',
  payments: 'Tahsilat / Ödeme',
  cash: 'Kasa / Banka',
  expenses: 'Gider Takibi',
  products: 'Ürün & Stok',
  reports: 'Raporlar',
  employees: 'Personel Yönetimi',
  projects: 'Proje Yönetimi',
  settings: 'Gelişmiş Ayarlar',
}
const featureLabel = (key: string) => featureLabels[key] || key

const planFeatures = (plan: any): string[] => {
  try {
    const arr = JSON.parse(plan.features)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

const activePlanFeatures = computed(() => {
  if (!activePlan.value) return []
  return planFeatures(activePlan.value)
})

const daysLeft = computed(() => {
  if (!billingStore.status) return 0
  const st = billingStore.status.subscription_status
  if (st === 'trial') return billingStore.status.trial_days_left ?? 0
  if (st === 'active') return billingStore.status.period_days_left ?? 0
  return 0
})

const totalDays = computed(() => {
  return billingStore.status?.subscription_status === 'trial' ? 14 : 30
})

const daysProgress = computed(() => {
  const pct = Math.round((daysLeft.value / totalDays.value) * 100)
  return Math.max(0, Math.min(100, pct))
})

const daysBarClass = computed(() => {
  if (daysLeft.value <= 3) return 'bar-danger'
  if (daysLeft.value <= 7) return 'bar-warn'
  return 'bar-ok'
})
</script>

<template>
  <div class="sp-root">
    <div class="sp-layout">

      <!-- ===== SIDEBAR ===== -->
      <aside class="sp-sidebar">
        <nav>
          <template v-for="group in settingsNav" :key="group.groupLabel">
            <div class="sp-group">
              <div class="sp-group-label">{{ group.groupLabel }}</div>
              <ul class="sp-group-items">
                <li
                  v-for="item in group.items"
                  :key="item.id"
                  class="sp-item"
                  :class="{ 'sp-item--active': activeTab === item.id }"
                  @click="handleNavClick(item.id)"
                >
                  <i :class="item.icon" class="sp-item-icon"></i>
                  <span>{{ item.label }}</span>
                </li>
              </ul>
            </div>
          </template>
        </nav>
      </aside>

      <!-- ===== CONTENT ===== -->
      <main class="sp-content">

        <!-- ========== FIRMA PROFİLİ ========== -->
        <div v-if="activeTab === 'profile'" class="sp-panel">
          <div class="sp-block">
            <div class="sp-block-header">
              <h2 class="sp-block-title">Firma Bilgileri</h2>
              <p class="sp-block-desc">Firma kimlik, iletişim ve bölgesel bilgilerini güncelleyin.</p>
            </div>
            <div class="sp-form-grid sp-form-grid--3">
              <div class="sp-field">
                <label class="sp-label">Firma Adı <span class="sp-required">*</span></label>
                <InputText v-model="companyForm.name" class="w-full uppercase-input p-inputtext-sm" maxlength="15" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Yetkili Ad Soyad</label>
                <InputText v-model="companyForm.contact_name" class="w-full uppercase-input p-inputtext-sm" maxlength="15" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Sektör</label>
                <Select v-model="companyForm.industry" :options="sectors" filter placeholder="Seçin" class="w-full" size="small" />
              </div>
              <div class="sp-field">
                <label class="sp-label">E-posta</label>
                <InputText v-model="companyForm.email" class="w-full p-inputtext-sm" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Cep Telefonu</label>
                <PhoneInput v-model="companyForm.phone" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Sabit Tel / Faks</label>
                <div class="flex gap-2">
                  <InputText v-model="companyForm.landline" class="w-full p-inputtext-sm" placeholder="+90 212..." />
                  <InputText v-model="companyForm.fax" class="w-full p-inputtext-sm" placeholder="Faks" />
                </div>
              </div>
              <div class="sp-field">
                <label class="sp-label">Vergi Dairesi</label>
                <InputText v-model="companyForm.tax_office" class="w-full uppercase-input p-inputtext-sm" maxlength="15" />
              </div>
              <div class="sp-field">
                <label class="sp-label">VN / TCKN</label>
                <InputText v-model="companyForm.tax_number" class="w-full p-inputtext-sm" maxlength="11" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Ülke</label>
                <Select v-model="companyForm.country" :options="countries" filter placeholder="Seçin" class="w-full" size="small" />
              </div>
              <div class="sp-field">
                <label class="sp-label">İl</label>
                <InputText v-model="companyForm.city" class="w-full uppercase-input p-inputtext-sm" />
              </div>
              <div class="sp-field">
                <label class="sp-label">İlçe</label>
                <InputText v-model="companyForm.district" class="w-full uppercase-input p-inputtext-sm" />
              </div>
              <div class="sp-field"><!-- spacer --></div>
              <div class="sp-field">
                <label class="sp-label">Para Birimi</label>
                <Select v-model="companyForm.currency" :options="currencies" optionLabel="label" optionValue="value" class="w-full" size="small" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Dil</label>
                <Select v-model="companyForm.locale" :options="locales" optionLabel="label" optionValue="value" class="w-full" size="small" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Saat Dilimi</label>
                <Select v-model="companyForm.timezone" :options="timezoneOptions" optionLabel="label" optionValue="value" class="w-full" size="small" />
              </div>
              <div class="sp-field sp-field--full">
                <label class="sp-label">Adres</label>
                <Textarea v-model="companyForm.address" rows="2" class="w-full p-textarea-sm" />
              </div>
              <div class="sp-field--full flex justify-end mt-2">
                <Button label="Kaydet" icon="pi pi-check" size="small" @click="saveCompanyProfile" :loading="settingsStore.loading" outlined severity="primary" />
              </div>
            </div>
          </div>

          <div class="sp-block sp-block--danger">
            <div class="sp-block-header">
              <h2 class="sp-block-title">Şifre Değiştir</h2>
              <p class="sp-block-desc">Hesap güvenliğiniz için şifrenizi güncelleyin.</p>
            </div>
            <div class="sp-form-grid sp-form-grid--4">
              <div class="sp-field">
                <label class="sp-label">Eski Şifre</label>
                <InputText v-model="passwordForm.old_password" type="password" class="w-full p-inputtext-sm" placeholder="••••••••" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Yeni Şifre</label>
                <InputText v-model="passwordForm.new_password" type="password" class="w-full p-inputtext-sm" placeholder="••••••••" />
              </div>
              <div class="sp-field">
                <label class="sp-label">Yeni Şifre (Tekrar)</label>
                <InputText v-model="passwordForm.new_password_confirm" type="password" class="w-full p-inputtext-sm" placeholder="••••••••" />
              </div>
              <div class="sp-field justify-end items-start">
                <Button label="Güncelle" icon="pi pi-lock" size="small" @click="changePassword" :loading="isPasswordLoading" outlined severity="warn" />
              </div>
            </div>
          </div>
        </div>

        <!-- ========== MODÜL YÖNETİMİ ========== -->
        <div v-else-if="activeTab === 'modules'" class="sp-panel">
          <div class="sp-block sp-block--flex">
            <div class="sp-block-header">
              <h2 class="sp-block-title">Modül Yönetimi</h2>
              <p class="sp-block-desc">Sisteminizde aktif olmasını istediğiniz modülleri yapılandırın.</p>
            </div>
            <ModulesTab />
          </div>
        </div>

        <!-- ========== SİSTEM PARAMETRELERİ ========== -->
        <div v-else-if="activeTab === 'finance'" class="sp-panel sp-panel--2col">
          <div class="sp-block sp-block--flex">
            <div class="sp-block-header">
              <h2 class="sp-block-title">KDV Tanımları</h2>
              <p class="sp-block-desc">Sistem genelinde kullanılacak KDV oranlarını yönetin.</p>
            </div>
            <KdvRatesTab />
          </div>
          <div class="sp-block sp-block--flex">
            <div class="sp-block-header">
              <h2 class="sp-block-title">Para Birimleri</h2>
              <p class="sp-block-desc">Sistem genelinde geçerli para birimlerini tanımlayın.</p>
            </div>
            <CurrenciesTab />
          </div>
        </div>

        <!-- ========== KASA / BANKA ========== -->
        <div v-else-if="activeTab === 'accounts'" class="sp-panel">
          <div class="sp-block sp-block--flex">
            <div class="sp-block-header">
              <h2 class="sp-block-title">Kasa / Banka Tanımları</h2>
              <p class="sp-block-desc">Nakit ve banka hesaplarınızı tanımlayın ve yönetin.</p>
            </div>
            <AccountsTab />
          </div>
        </div>

        <!-- ========== İÇE AKTAR ========== -->
        <div v-else-if="activeTab === 'import'" class="sp-panel">
          <div class="sp-block sp-block--flex">
            <div class="sp-block-header">
              <h2 class="sp-block-title">İçe Aktar</h2>
              <p class="sp-block-desc">Harici kaynaklardan veri aktarımı yapın.</p>
            </div>
            <ImportTab />
          </div>
        </div>

        <!-- ========== ABONELİK ========== -->
        <div v-else-if="activeTab === 'billing'" class="sp-panel">

          <div class="bl-summary" v-if="billingStore.status">
            <div class="bl-summary-main">
              <div class="bl-summary-top">
                <span class="bl-status-dot"
                  :class="{
                    'dot-trial': billingStore.status.subscription_status === 'trial',
                    'dot-active': billingStore.status.subscription_status === 'active',
                    'dot-canceled': billingStore.status.subscription_status === 'canceled'
                  }"></span>
                <span class="bl-plan-name">{{ billingStore.status.plan_name }}</span>
                <span class="bl-status-badge"
                  :class="{
                    'badge-trial': billingStore.status.subscription_status === 'trial',
                    'badge-active': billingStore.status.subscription_status === 'active',
                    'badge-canceled': billingStore.status.subscription_status === 'canceled'
                  }">{{ statusText }}</span>
              </div>

              <div class="bl-progress" v-if="billingStore.status.subscription_status === 'trial' || billingStore.status.subscription_status === 'active'">
                <div class="bl-progress-info">
                  <span class="bl-days">
                    <strong>{{ daysLeft }}</strong> gün kaldı
                  </span>
                  <span class="bl-end-date">
                    {{ billingStore.status.subscription_status === 'trial' ? 'Deneme bitişi' : 'Sonraki ödeme' }}:
                    {{ formatDate(billingStore.status.subscription_status === 'trial' ? billingStore.status.trial_ends_at : billingStore.status.current_period_end) }}
                  </span>
                </div>
                <div class="bl-bar-track">
                  <div class="bl-bar-fill" :class="daysBarClass" :style="{ width: daysProgress + '%' }"></div>
                </div>
              </div>

              <p v-else-if="billingStore.status.subscription_status === 'canceled'" class="bl-canceled-note">
                <i class="pi pi-exclamation-triangle"></i>
                Aboneliğiniz sona erdi. Devam etmek için aşağıdan bir paket seçin.
              </p>

              <div class="bl-feats" v-if="activePlanFeatures.length">
                <span v-for="f in activePlanFeatures" :key="f" class="bl-feat-chip">{{ featureLabel(f) }}</span>
              </div>
            </div>

            <div class="bl-summary-action" v-if="billingStore.status.subscription_status === 'active'">
              <Button label="Aboneliği Uzat" icon="pi pi-refresh" @click="openRenewDialog" :disabled="!canRenew" severity="success" />
              <span v-if="!canRenew" class="bl-action-note">Bitişe 15 gün kala açılır</span>
            </div>
          </div>

          <div class="bl-plans-header">
            <h2 class="sp-block-title">Abonelik Paketleri</h2>
            <p class="sp-block-desc">İhtiyacınıza en uygun paketi seçin. İstediğiniz zaman yükseltebilirsiniz.</p>
          </div>
          <div class="plans-grid">
            <div v-for="plan in billingStore.plans" :key="plan.id"
              class="plan-card"
              :class="{
                'plan-card--pro': plan.code === 'pro',
                'plan-card--current': billingStore.status && billingStore.status.plan_code === plan.code
              }">
              <div v-if="plan.code === 'pro'" class="plan-badge">ÖNERİLEN</div>
              <h3 class="plan-name">{{ plan.name }}</h3>
              <div class="plan-price">
                <div class="plan-price-line">
                  <span class="plan-price-main">{{ formatPrice(plan.price_monthly, plan.currency) }}</span>
                  <span class="plan-price-period">/ ay</span>
                </div>
                <span class="plan-price-sub" v-if="parseFloat(plan.price_yearly) > 0">
                  Yıllık {{ formatPrice(plan.price_yearly, plan.currency) }}
                </span>
              </div>
              <ul class="plan-feats">
                <li v-for="feat in planFeatures(plan)" :key="feat">
                  <i class="pi pi-check"></i><span>{{ featureLabel(feat) }}</span>
                </li>
              </ul>
              <div class="plan-cta">
                <Button v-if="billingStore.status && billingStore.status.plan_code === plan.code"
                  label="Mevcut Paketiniz" icon="pi pi-check-circle" class="w-full" disabled />
                <div v-else-if="!isPlanSelectable(plan)" class="flex flex-col gap-1">
                  <Button label="Şu An Geçilemez" class="w-full" disabled outlined />
                  <span class="plan-note">Bitişe 15 gün kala açılır</span>
                </div>
                <div v-else class="flex flex-col gap-2">
                  <Button :label="parseFloat(plan.price_monthly) === 0 ? 'Ücretsiz Başla' : 'Aylık Al'"
                    class="w-full" severity="primary"
                    @click="handleSubscribe(plan.id, 'monthly')" :loading="billingStore.loading" />
                  <Button v-if="parseFloat(plan.price_monthly) !== 0"
                    label="Yıllık Al" class="w-full" severity="primary"
                    @click="handleSubscribe(plan.id, 'yearly')" :loading="billingStore.loading" outlined />
                </div>
              </div>
            </div>
          </div>

          <div class="sp-block" style="margin-top:1rem">
            <div class="sp-block-header">
              <h2 class="sp-block-title">Ödeme Geçmişi</h2>
              <p class="sp-block-desc">Geçmiş ödeme ve abonelik hareketleri.</p>
            </div>
            <DataTable :value="billingStore.transactions" :loading="billingStore.loading" responsiveLayout="scroll" class="p-datatable-sm w-full">
              <Column field="created_at" header="Tarih">
                <template #body="sp">{{ formatDateTime(sp.data.created_at) }}</template>
              </Column>
              <Column field="action" header="İşlem">
                <template #body="sp"><span class="font-medium">{{ sp.data.action }}</span></template>
              </Column>
              <Column field="amount" header="Tutar">
                <template #body="sp">{{ parseFloat(sp.data.amount) }} {{ sp.data.currency }}</template>
              </Column>
              <Column field="status" header="Durum">
                <template #body="sp">
                  <Tag :severity="sp.data.status === 'success' ? 'success' : 'danger'" :value="sp.data.status" />
                </template>
              </Column>
              <Column field="provider_ref" header="Referans">
                <template #body="sp"><code class="text-xs text-slate-400">{{ sp.data.provider_ref || '-' }}</code></template>
              </Column>
            </DataTable>
          </div>

          <Dialog v-model:visible="displaySuccessDialog" header="Paket Aktif!" modal :style="{ width: '400px' }">
            <div class="text-center p-4">
              <i class="pi pi-check-circle text-emerald-500 text-6xl mb-4"></i>
              <h3 class="text-lg font-bold mb-2">Paketiniz Başarıyla Tanımlandı!</h3>
              <p class="text-sm text-slate-500">Yeni aboneliğiniz hesabınıza tanımlanmıştır.</p>
            </div>
            <template #footer>
              <div class="flex justify-center">
                <Button label="Tamam" severity="primary" @click="displaySuccessDialog = false" outlined />
              </div>
            </template>
          </Dialog>

          <Dialog v-model:visible="renewDialog" header="Aboneliği Uzat" modal :style="{ width: '450px' }">
            <div class="p-4" v-if="billingStore.status">
              <p class="mb-4">Mevcut planınız (<strong>{{ billingStore.status.plan_name }}</strong>) için uzatma periyodu seçin:</p>
              <div class="flex flex-col gap-3">
                <div class="renew-option" :class="{ active: selectedPeriod === 'monthly' }" @click="selectedPeriod = 'monthly'">
                  <div><span class="font-semibold block">Aylık Uzatma</span><span class="text-xs text-slate-400">Her ay yenilenir</span></div>
                  <span class="font-bold text-sky-600">{{ activePlanMonthlyPrice }}</span>
                </div>
                <div class="renew-option" :class="{ active: selectedPeriod === 'yearly' }" @click="selectedPeriod = 'yearly'">
                  <div><span class="font-semibold block">Yıllık Uzatma</span><span class="text-xs text-slate-400">12 ay yenilenir</span></div>
                  <span class="font-bold text-sky-600">{{ activePlanYearlyPrice }}</span>
                </div>
              </div>
            </div>
            <template #footer>
              <div class="flex justify-end gap-2">
                <Button label="İptal" class="p-button-text" @click="renewDialog = false" outlined />
                <Button label="Ödemeyi Yap ve Uzat" severity="primary" @click="handleRenew" :loading="billingStore.loading" outlined />
              </div>
            </template>
          </Dialog>
        </div>

      </main>
    </div>
  </div>
</template>

<style scoped>
.sp-root { min-height: 0; }

.sp-layout {
  display: flex;
  gap: 0;
  align-items: flex-start;
  min-height: 0;
}

.sp-sidebar {
  width: 210px;
  flex-shrink: 0;
  border-right: 1px solid #e8eaf0;
  padding: 0.5rem 0;
  position: sticky;
  top: 0;
  align-self: flex-start;
  background: #fff;
  min-height: calc(100vh - 64px);
}
:root.p-dark .sp-sidebar { background: #1e293b; border-right-color: rgba(255,255,255,0.07); }

.sp-group { padding: 0.75rem 0 0.25rem; }
.sp-group + .sp-group { border-top: 1px solid #f0f2f5; }
:root.p-dark .sp-group + .sp-group { border-top-color: rgba(255,255,255,0.05); }

.sp-group-label {
  padding: 0 1rem 0.3rem;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  color: #a0aec0;
}

.sp-group-items { list-style: none; margin: 0; padding: 0; }

.sp-item {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.55rem 1rem;
  font-size: 0.875rem;
  color: #4a5568;
  cursor: pointer;
  border-left: 2px solid transparent;
  transition: color 0.12s, background 0.12s, border-color 0.12s;
  user-select: none;
}
:root.p-dark .sp-item { color: #94a3b8; }
.sp-item:hover { color: #06b6d4; background: rgba(6,182,212,0.04); }
.sp-item--active { color: #06b6d4 !important; border-left-color: #06b6d4; background: rgba(6,182,212,0.06); font-weight: 600; }
:root.p-dark .sp-item--active { background: rgba(6,182,212,0.1); }
.sp-item-icon { font-size: 0.75rem; flex-shrink: 0; opacity: 0.8; }

.sp-content {
  flex: 1;
  min-width: 0;
  padding: 1.25rem 1.5rem;
  background: #f7f8fb;
  min-height: calc(100vh - 64px);
}
:root.p-dark .sp-content { background: #0f172a; }

.sp-panel { display: flex; flex-direction: column; gap: 1rem; }
.sp-panel--2col { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; align-items: start; }
@media (max-width: 900px) {
  .sp-panel--2col { grid-template-columns: 1fr; }
  .sp-sidebar { width: 100%; min-height: auto; position: static; border-right: none; border-bottom: 1px solid #e8eaf0; }
  .sp-layout { flex-direction: column; }
  .sp-content { min-height: auto; }
}

.sp-block {
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 0.625rem;
  padding: 1.375rem 1.5rem;
}
:root.p-dark .sp-block { background: #1e293b; border-color: rgba(255,255,255,0.07); }
.sp-block--flex { display: flex; flex-direction: column; }
.sp-block--danger { border-color: rgba(239,68,68,0.18); background: rgba(255,245,245,0.6); }
:root.p-dark .sp-block--danger { background: rgba(239,68,68,0.04); border-color: rgba(239,68,68,0.22); }

.sp-block-header { padding-bottom: 1rem; margin-bottom: 1rem; border-bottom: 1px solid #f0f2f5; }
:root.p-dark .sp-block-header { border-bottom-color: rgba(255,255,255,0.05); }
.sp-block-title { font-size: 0.92rem; font-weight: 700; color: #1a202c; margin: 0 0 0.2rem; }
:root.p-dark .sp-block-title { color: #f1f5f9; }
.sp-block-desc { font-size: 0.75rem; color: #718096; margin: 0; }

.sp-form-grid { display: grid; gap: 0.75rem 1rem; }
.sp-form-grid--1 { grid-template-columns: 1fr; }
.sp-form-grid--2 { grid-template-columns: 1fr 1fr; }
.sp-form-grid--3 { grid-template-columns: 1fr 1fr 1fr; }
.sp-form-grid--4 { grid-template-columns: 1fr 1fr 1fr 1fr; }
@media (max-width: 1100px) {
  .sp-form-grid--4 { grid-template-columns: 1fr 1fr; }
  .sp-form-grid--3 { grid-template-columns: 1fr 1fr; }
}
@media (max-width: 680px) {
  .sp-form-grid--2, .sp-form-grid--3, .sp-form-grid--4 { grid-template-columns: 1fr; }
}

.sp-field { display: flex; flex-direction: column; gap: 0.3rem; }
.sp-field--full { grid-column: 1 / -1; }
.sp-label { font-size: 0.72rem; font-weight: 600; color: #4a5568; letter-spacing: 0.015em; }
:root.p-dark .sp-label { color: #94a3b8; }
.sp-required { color: #e53e3e; }

/* Billing */
.bl-summary {
  display: flex; justify-content: space-between; align-items: center; gap: 1.5rem;
  background: #fff; border: 1px solid #e8eaf0; border-radius: 0.75rem; padding: 1.25rem 1.5rem;
}
:root.p-dark .bl-summary { background: #1e293b; border-color: rgba(255,255,255,0.07); }
.bl-summary-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.85rem; }
.bl-summary-top { display: flex; align-items: center; gap: 0.6rem; flex-wrap: wrap; }
.bl-status-dot { width: 9px; height: 9px; border-radius: 9999px; flex-shrink: 0; }
.dot-trial { background: #d69e2e; box-shadow: 0 0 0 3px rgba(214,158,46,.15); }
.dot-active { background: #38a169; box-shadow: 0 0 0 3px rgba(56,161,105,.15); }
.dot-canceled { background: #e53e3e; box-shadow: 0 0 0 3px rgba(229,62,62,.15); }
.bl-plan-name { font-size: 1.05rem; font-weight: 700; color: #1a202c; }
:root.p-dark .bl-plan-name { color: #f1f5f9; }
.bl-status-badge { font-size: 0.7rem; font-weight: 700; padding: 0.15rem 0.55rem; border-radius: 9999px; text-transform: uppercase; letter-spacing: 0.03em; }
.badge-trial { color: #d69e2e; background: rgba(214,158,46,.12); }
.badge-active { color: #38a169; background: rgba(56,161,105,.12); }
.badge-canceled { color: #e53e3e; background: rgba(229,62,62,.12); }
.bl-progress { display: flex; flex-direction: column; gap: 0.4rem; max-width: 420px; }
.bl-progress-info { display: flex; justify-content: space-between; align-items: baseline; gap: 1rem; }
.bl-days { font-size: 0.82rem; color: #4a5568; }
.bl-days strong { font-size: 1.05rem; font-weight: 700; color: #1a202c; }
:root.p-dark .bl-days { color: #94a3b8; }
:root.p-dark .bl-days strong { color: #f1f5f9; }
.bl-end-date { font-size: 0.72rem; color: #a0aec0; }
.bl-bar-track { height: 6px; border-radius: 9999px; background: #edf0f5; overflow: hidden; }
:root.p-dark .bl-bar-track { background: rgba(255,255,255,0.08); }
.bl-bar-fill { height: 100%; border-radius: 9999px; transition: width 0.4s ease; }
.bar-ok { background: #38a169; }
.bar-warn { background: #d69e2e; }
.bar-danger { background: #e53e3e; }
.bl-canceled-note { display: flex; align-items: center; gap: 0.5rem; font-size: 0.82rem; color: #e53e3e; margin: 0; }
.bl-feats { display: flex; flex-wrap: wrap; gap: 0.4rem; }
.bl-feat-chip { font-size: 0.7rem; font-weight: 500; color: #4a5568; background: #f4f6f9; border: 1px solid #e8eaf0; padding: 0.15rem 0.5rem; border-radius: 0.375rem; }
:root.p-dark .bl-feat-chip { color: #94a3b8; background: rgba(255,255,255,0.04); border-color: rgba(255,255,255,0.08); }
.bl-summary-action { display: flex; flex-direction: column; align-items: flex-end; gap: 0.35rem; flex-shrink: 0; }
.bl-action-note { font-size: 0.68rem; color: #a0aec0; }
.bl-plans-header { margin: 1.5rem 0 0.85rem; }

.plans-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); gap: 1rem; }
.plan-card { border: 1px solid #e8eaf0; border-radius: 0.625rem; padding: 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; position: relative; background: #fff; }
:root.p-dark .plan-card { background: #0f172a; border-color: rgba(255,255,255,0.08); }
.plan-card--pro { border-color: #06b6d4; box-shadow: 0 0 0 1px #06b6d4; }
.plan-card--current { border-color: #38a169; background: rgba(56,161,105,0.02); }
.plan-badge { position: absolute; top: 0.75rem; right: 0.75rem; background: #06b6d4; color: #fff; font-size: 0.58rem; font-weight: 800; letter-spacing: 0.07em; padding: 0.18rem 0.45rem; border-radius: 9999px; }
.plan-name { font-size: 1rem; font-weight: 700; color: #1a202c; margin: 0; }
:root.p-dark .plan-name { color: #f1f5f9; }
.plan-price { display: flex; flex-direction: column; gap: 0.15rem; padding-bottom: 0.85rem; border-bottom: 1px solid #f0f2f5; }
:root.p-dark .plan-price { border-bottom-color: rgba(255,255,255,0.06); }
.plan-price-line { display: flex; align-items: baseline; gap: 0.3rem; }
.plan-price-main { font-size: 1.6rem; font-weight: 800; color: #1a202c; letter-spacing: -0.02em; }
:root.p-dark .plan-price-main { color: #f1f5f9; }
.plan-price-period { font-size: 0.8rem; font-weight: 500; color: #a0aec0; }
.plan-price-sub { font-size: 0.74rem; color: #718096; }
.plan-feats { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 0.4rem; flex: 1; }
.plan-feats li { display: flex; align-items: center; gap: 0.4rem; font-size: 0.77rem; color: #4a5568; }
:root.p-dark .plan-feats li { color: #94a3b8; }
.plan-feats .pi-check { color: #38a169; font-size: 0.7rem; }
.plan-cta { margin-top: auto; }
.plan-note { font-size: 0.68rem; color: #a0aec0; text-align: center; display: block; margin-top: 0.25rem; }

.renew-option { display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 1rem; border-radius: 0.5rem; border: 1px solid #e8eaf0; cursor: pointer; transition: border-color 0.15s, background 0.15s; }
.renew-option:hover { border-color: #06b6d4; }
.renew-option.active { border-color: #06b6d4; background: rgba(6,182,212,0.05); }
</style>
