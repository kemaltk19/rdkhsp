<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { usePaymentStore } from '@/stores/payment'
import { useCurrencyStore } from '@/stores/currency'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'

import ProductCategoriesTab from './components/ProductCategoriesTab.vue'
import ProductSettingsTab from './components/ProductSettingsTab.vue'
import WarehousesTab from './components/WarehousesTab.vue'
import ExpenseCategoriesTab from './components/ExpenseCategoriesTab.vue'
import EmployeeSettingsTab from './components/EmployeeSettingsTab.vue'
import CariGroupsTab from './components/CariGroupsTab.vue'
import ProjectCategoriesTab from './components/ProjectCategoriesTab.vue'

const settingsStore = useSettingsStore()
const paymentStore = usePaymentStore()
const currencyStore = useCurrencyStore()
const toast = useToast()

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

const invoiceKdv = ref('20')
const invoicePrefix = ref('INV-S')
const invoiceDueDays = ref('14')
const invoiceFooterNote = ref('')
const kdvOptions = ref<string[]>(['0', '1', '10', '20'])

const quotePrefix = ref('PRO')
const quoteValidityDays = ref('15')
const quoteFooterNote = ref('')

const cariDefaultCurrency = ref('TRY')
const cariPrefix = ref('ACC')
const cariRiskLimit = ref('0')

const expensePrefix = ref('EXP')

const projectCodePrefix = ref('PRJ')
const projectCodeCounter = ref('1000')

const defaultCashAccount = ref('')
const collectionPrefix = ref('TAH')
const paymentPrefix = ref('ODE')
const cashAccountOptions = ref<{ label: string; value: string }[]>([])

const loadSettings = async () => {
  try {
    const kdvVal = await settingsStore.fetchSetting('default_kdv_rate')
    if (kdvVal) invoiceKdv.value = kdvVal

    const kdvRatesStr = await settingsStore.fetchSetting('kdv_rates')
    if (kdvRatesStr) {
      try {
        const parsed = JSON.parse(kdvRatesStr)
        if (Array.isArray(parsed)) kdvOptions.value = parsed.map(String)
      } catch {}
    }

    const prefixVal = await settingsStore.fetchSetting('invoice_prefix')
    if (prefixVal) invoicePrefix.value = prefixVal
    const dueDaysVal = await settingsStore.fetchSetting('invoice_due_days')
    if (dueDaysVal) invoiceDueDays.value = dueDaysVal
    const invFooterVal = await settingsStore.fetchSetting('invoice_footer_note')
    if (invFooterVal) invoiceFooterNote.value = invFooterVal

    const qPrefixVal = await settingsStore.fetchSetting('quote_prefix')
    if (qPrefixVal) quotePrefix.value = qPrefixVal
    const qValidVal = await settingsStore.fetchSetting('quote_validity_days')
    if (qValidVal) quoteValidityDays.value = qValidVal
    const qFooterVal = await settingsStore.fetchSetting('quote_footer_note')
    if (qFooterVal) quoteFooterNote.value = qFooterVal

    const cCurrVal = await settingsStore.fetchSetting('cari_default_currency')
    if (cCurrVal) cariDefaultCurrency.value = cCurrVal
    const cPrefixVal = await settingsStore.fetchSetting('cari_prefix')
    if (cPrefixVal) cariPrefix.value = cPrefixVal
    const cRiskVal = await settingsStore.fetchSetting('cari_risk_limit')
    if (cRiskVal) cariRiskLimit.value = cRiskVal

    const fCashVal = await settingsStore.fetchSetting('default_cash_account')
    if (fCashVal) defaultCashAccount.value = fCashVal
    const collPrefixVal = await settingsStore.fetchSetting('collection_prefix')
    if (collPrefixVal) collectionPrefix.value = collPrefixVal
    const payPrefixVal = await settingsStore.fetchSetting('payment_prefix')
    if (payPrefixVal) paymentPrefix.value = payPrefixVal

    await paymentStore.fetchAccounts()
    cashAccountOptions.value = [
      ...paymentStore.cashAccounts.map(acc => ({ label: `${acc.name} (Kasa)`, value: acc.id })),
      ...paymentStore.bankAccounts.map(acc => ({ label: `${acc.name} (Banka)`, value: acc.id })),
    ]

    const ePrefixVal = await settingsStore.fetchSetting('expense_prefix')
    if (ePrefixVal) expensePrefix.value = ePrefixVal

    const projPrefixVal = await settingsStore.fetchSetting('project_code_prefix')
    if (projPrefixVal) projectCodePrefix.value = projPrefixVal
    const projCounterVal = await settingsStore.fetchSetting('project_code_counter')
    if (projCounterVal) projectCodeCounter.value = projCounterVal

    await currencyStore.fetchCurrencies()
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Ayarlar yüklenemedi.', life: 5000 })
  }
}

onMounted(loadSettings)

const saveInvoiceSettings = async () => {
  try {
    await settingsStore.saveSetting('default_kdv_rate', invoiceKdv.value, 'invoice')
    await settingsStore.saveSetting('invoice_prefix', invoicePrefix.value, 'invoice')
    await settingsStore.saveSetting('invoice_due_days', invoiceDueDays.value, 'invoice')
    await settingsStore.saveSetting('invoice_footer_note', invoiceFooterNote.value, 'invoice')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Fatura ayarları kaydedildi.', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Fatura ayarları kaydedilemedi.', life: 5000 })
  }
}

const saveQuoteSettings = async () => {
  try {
    await settingsStore.saveSetting('quote_prefix', quotePrefix.value, 'quote')
    await settingsStore.saveSetting('quote_validity_days', quoteValidityDays.value, 'quote')
    await settingsStore.saveSetting('quote_footer_note', quoteFooterNote.value, 'quote')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif ayarları kaydedildi.', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Teklif ayarları kaydedilemedi.', life: 5000 })
  }
}

const saveCariSettings = async () => {
  try {
    await settingsStore.saveSetting('cari_default_currency', cariDefaultCurrency.value, 'cari')
    await settingsStore.saveSetting('cari_prefix', cariPrefix.value, 'cari')
    await settingsStore.saveSetting('cari_risk_limit', cariRiskLimit.value, 'cari')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Cari ayarları kaydedildi.', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Cari ayarları kaydedilemedi.', life: 5000 })
  }
}

const saveExpenseSettings = async () => {
  try {
    await settingsStore.saveSetting('expense_prefix', expensePrefix.value, 'expense')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Gider ayarları kaydedildi.', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Gider ayarları kaydedilemedi.', life: 5000 })
  }
}

const saveProjectSettings = async () => {
  try {
    await settingsStore.saveSetting('project_code_prefix', projectCodePrefix.value, 'project')
    await settingsStore.saveSetting('project_code_counter', projectCodeCounter.value, 'project')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Proje ayarları kaydedildi.', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Proje ayarları kaydedilemedi.', life: 5000 })
  }
}

const savePaymentSettings = async () => {
  try {
    await settingsStore.saveSetting('default_cash_account', defaultCashAccount.value, 'finance')
    await settingsStore.saveSetting('collection_prefix', collectionPrefix.value, 'finance')
    await settingsStore.saveSetting('payment_prefix', paymentPrefix.value, 'finance')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Kasa/Ödeme ayarları kaydedildi.', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Kasa/Ödeme ayarları kaydedilemedi.', life: 5000 })
  }
}
</script>

<template>
  <div class="ma-root">
    <div class="ma-header">
      <h1 class="ma-title">Modül Ayarları</h1>
      <p class="ma-desc">Her modüle ait varsayılan değerleri ve yapılandırmayı yönetin.</p>
    </div>

    <Tabs value="invoice">
      <TabList class="ma-tablist">
        <Tab value="invoice"><i class="pi pi-file mr-1"></i>Faturalar</Tab>
        <Tab value="quote"><i class="pi pi-file-edit mr-1"></i>Teklifler</Tab>
        <Tab value="cari"><i class="pi pi-users mr-1"></i>Cariler</Tab>
        <Tab value="product"><i class="pi pi-box mr-1"></i>Ürün & Stok</Tab>
        <Tab value="expense"><i class="pi pi-receipt mr-1"></i>Giderler</Tab>
        <Tab value="employee"><i class="pi pi-id-card mr-1"></i>Personel</Tab>
        <Tab value="project"><i class="pi pi-briefcase mr-1"></i>Projeler</Tab>
        <Tab value="payment"><i class="pi pi-wallet mr-1"></i>Kasa & Ödeme</Tab>
      </TabList>

      <TabPanels class="ma-panels">

        <TabPanel value="invoice">
          <div class="ma-block">
            <div class="ma-block-header">
              <h2 class="ma-block-title">Fatura Ayarları</h2>
              <p class="ma-block-desc">Yeni faturalar için varsayılan değerleri tanımlayın.</p>
            </div>
            <div class="ma-grid ma-grid--2">
              <div class="ma-field">
                <label class="ma-label">Fatura Ön Eki</label>
                <InputText v-model="invoicePrefix" placeholder="örn: FAT" class="w-full p-inputtext-sm" />
              </div>
              <div class="ma-field">
                <label class="ma-label">Varsayılan KDV Oranı (%)</label>
                <Select v-model="invoiceKdv" :options="kdvOptions" class="w-full" size="small" />
              </div>
              <div class="ma-field">
                <label class="ma-label">Vade (Gün)</label>
                <InputText v-model="invoiceDueDays" placeholder="14" class="w-full p-inputtext-sm" type="number" />
              </div>
              <div class="ma-field ma-field--full">
                <label class="ma-label">Alt Not / IBAN</label>
                <Textarea v-model="invoiceFooterNote" rows="3" class="w-full" placeholder="Ödemelerinizi TR12... numaralı hesabımıza yapabilirsiniz." />
              </div>
              <div class="ma-field--full flex justify-end">
                <Button label="Kaydet" icon="pi pi-check" size="small" @click="saveInvoiceSettings" :loading="settingsStore.loading" outlined severity="primary" />
              </div>
            </div>
          </div>
        </TabPanel>

        <TabPanel value="quote">
          <div class="ma-block">
            <div class="ma-block-header">
              <h2 class="ma-block-title">Teklif Ayarları</h2>
              <p class="ma-block-desc">Yeni teklifler için varsayılan değerleri tanımlayın.</p>
            </div>
            <div class="ma-grid ma-grid--2">
              <div class="ma-field">
                <label class="ma-label">Teklif Ön Eki</label>
                <InputText v-model="quotePrefix" placeholder="örn: TKLF" class="w-full p-inputtext-sm" />
              </div>
              <div class="ma-field">
                <label class="ma-label">Geçerlilik Süresi (Gün)</label>
                <InputText v-model="quoteValidityDays" placeholder="15" class="w-full p-inputtext-sm" type="number" />
              </div>
              <div class="ma-field ma-field--full">
                <label class="ma-label">Teklif Şartları ve Alt Not</label>
                <Textarea v-model="quoteFooterNote" rows="3" class="w-full" placeholder="Bu teklif 15 gün süreyle geçerlidir. KDV dahil değildir." />
              </div>
              <div class="ma-field--full flex justify-end">
                <Button label="Kaydet" icon="pi pi-check" size="small" @click="saveQuoteSettings" :loading="settingsStore.loading" outlined severity="primary" />
              </div>
            </div>
          </div>
        </TabPanel>

        <TabPanel value="cari">
          <div class="ma-2col">
            <div class="ma-block">
              <div class="ma-block-header">
                <h2 class="ma-block-title">Genel Cari Ayarları</h2>
                <p class="ma-block-desc">Yeni cariler için varsayılan değerleri belirleyin.</p>
              </div>
              <div class="ma-grid ma-grid--1">
                <div class="ma-field">
                  <label class="ma-label">Varsayılan Para Birimi</label>
                  <Select v-model="cariDefaultCurrency" :options="currencies" optionLabel="label" optionValue="value" class="w-full" size="small" />
                </div>
                <div class="ma-field">
                  <label class="ma-label">Cari Kodu Ön Eki</label>
                  <InputText v-model="cariPrefix" placeholder="örn: CARI" class="w-full p-inputtext-sm" />
                </div>
                <div class="ma-field">
                  <label class="ma-label">Risk Limiti Uyarı (₺)</label>
                  <InputText v-model="cariRiskLimit" placeholder="50000" class="w-full p-inputtext-sm" type="number" />
                </div>
                <div class="flex justify-end">
                  <Button label="Kaydet" icon="pi pi-check" size="small" @click="saveCariSettings" :loading="settingsStore.loading" outlined severity="primary" />
                </div>
              </div>
            </div>
            <div class="ma-block ma-block--flex">
              <div class="ma-block-header">
                <h2 class="ma-block-title">Cari Grupları</h2>
                <p class="ma-block-desc">Müşteri ve tedarikçileri sınıflandırmak için gruplar oluşturun.</p>
              </div>
              <CariGroupsTab />
            </div>
          </div>
        </TabPanel>

        <TabPanel value="product">
          <div class="ma-2col">
            <div class="ma-block">
              <div class="ma-block-header">
                <h2 class="ma-block-title">Genel Ürün Ayarları</h2>
                <p class="ma-block-desc">Ürün ve stok kayıtları için varsayılan değerleri belirleyin.</p>
              </div>
              <ProductSettingsTab />
            </div>
            <div class="ma-block ma-block--flex">
              <div class="ma-block-header">
                <h2 class="ma-block-title">Ürün Kategorileri</h2>
                <p class="ma-block-desc">Ürünlerinizi sınıflandırmak için kategoriler oluşturun.</p>
              </div>
              <ProductCategoriesTab />
            </div>
          </div>
          <div class="ma-block ma-block--flex mt-4">
            <div class="ma-block-header">
              <h2 class="ma-block-title">Depo Tanımları</h2>
              <p class="ma-block-desc">Ürün stoklarınızı yöneteceğiniz depoları tanımlayın.</p>
            </div>
            <WarehousesTab />
          </div>
        </TabPanel>

        <TabPanel value="expense">
          <div class="ma-2col">
            <div class="ma-block">
              <div class="ma-block-header">
                <h2 class="ma-block-title">Gider Ayarları</h2>
                <p class="ma-block-desc">Gider belgeleri için varsayılan değerleri belirleyin.</p>
              </div>
              <div class="ma-grid ma-grid--1">
                <div class="ma-field">
                  <label class="ma-label">Gider Fişi Ön Eki</label>
                  <InputText v-model="expensePrefix" placeholder="örn: EXP" class="w-full p-inputtext-sm" />
                </div>
                <div class="flex justify-end">
                  <Button label="Kaydet" icon="pi pi-check" size="small" @click="saveExpenseSettings" :loading="settingsStore.loading" outlined severity="primary" />
                </div>
              </div>
            </div>
            <div class="ma-block ma-block--flex">
              <div class="ma-block-header">
                <h2 class="ma-block-title">Gider Kategorileri</h2>
                <p class="ma-block-desc">Giderlerinizi raporlarda sınıflandırmak için kategoriler oluşturun.</p>
              </div>
              <ExpenseCategoriesTab />
            </div>
          </div>
        </TabPanel>

        <TabPanel value="employee">
          <div class="ma-block ma-block--flex">
            <div class="ma-block-header">
              <h2 class="ma-block-title">Personel Ayarları</h2>
              <p class="ma-block-desc">İnsan kaynakları ve personel yönetimi yapılandırması.</p>
            </div>
            <EmployeeSettingsTab />
          </div>
        </TabPanel>

        <TabPanel value="project">
          <div class="ma-block">
            <div class="ma-block-header">
              <h2 class="ma-block-title">Proje Ayarları</h2>
              <p class="ma-block-desc">Projeler için kod ön eki ve sayaç değerlerini tanımlayın.</p>
            </div>
            <div class="ma-grid ma-grid--2">
              <div class="ma-field">
                <label class="ma-label">Proje Kodu Ön Eki</label>
                <InputText v-model="projectCodePrefix" placeholder="örn: PRJ" class="w-full p-inputtext-sm" maxlength="10" />
              </div>
              <div class="ma-field">
                <label class="ma-label">Başlangıç Sayacı</label>
                <InputText v-model="projectCodeCounter" placeholder="1000" class="w-full p-inputtext-sm" type="number" />
              </div>
              <div class="ma-field--full flex justify-end">
                <Button label="Kaydet" icon="pi pi-check" size="small" @click="saveProjectSettings" :loading="settingsStore.loading" outlined severity="primary" />
              </div>
            </div>
          </div>
          <div class="ma-block ma-block--flex mt-4">
            <div class="ma-block-header">
              <h2 class="ma-block-title">Proje Kategorileri</h2>
              <p class="ma-block-desc">Projelerinizi sınıflandırmak için kategoriler oluşturun.</p>
            </div>
            <ProjectCategoriesTab />
          </div>
        </TabPanel>

        <TabPanel value="payment">
          <div class="ma-block">
            <div class="ma-block-header">
              <h2 class="ma-block-title">Kasa & Ödeme Ayarları</h2>
              <p class="ma-block-desc">Tahsilat ve ödeme işlemleri için varsayılan değerleri belirleyin.</p>
            </div>
            <div class="ma-grid ma-grid--3">
              <div class="ma-field">
                <label class="ma-label">Tahsilat Makbuzu Ön Eki</label>
                <InputText v-model="collectionPrefix" placeholder="örn: TAH" class="w-full p-inputtext-sm" />
              </div>
              <div class="ma-field">
                <label class="ma-label">Ödeme Makbuzu Ön Eki</label>
                <InputText v-model="paymentPrefix" placeholder="örn: ODE" class="w-full p-inputtext-sm" />
              </div>
              <div class="ma-field">
                <label class="ma-label">Varsayılan Tahsilat Kasası</label>
                <Select v-model="defaultCashAccount" :options="cashAccountOptions" optionLabel="label" optionValue="value" placeholder="Kasa / banka seçin" class="w-full" size="small" />
              </div>
              <div class="ma-field--full flex justify-end">
                <Button label="Kaydet" icon="pi pi-check" size="small" @click="savePaymentSettings" :loading="settingsStore.loading" outlined severity="primary" />
              </div>
            </div>
          </div>
        </TabPanel>

      </TabPanels>
    </Tabs>
  </div>
</template>

<style scoped>
.ma-root {
  padding: 1.25rem 1.5rem;
  background: #f7f8fb;
  min-height: calc(100vh - 64px);
}
:root.p-dark .ma-root { background: #0f172a; }

.ma-header { margin-bottom: 1.25rem; }
.ma-title { font-size: 1.1rem; font-weight: 700; color: #1a202c; margin: 0 0 0.2rem; }
:root.p-dark .ma-title { color: #f1f5f9; }
.ma-desc { font-size: 0.78rem; color: #718096; margin: 0; }

.ma-tablist :deep(.p-tablist-tab-list) {
  border-bottom: 1px solid #e8eaf0;
  gap: 0;
  background: transparent;
}
:root.p-dark .ma-tablist :deep(.p-tablist-tab-list) { border-bottom-color: rgba(255,255,255,0.07); }
.ma-tablist :deep(.p-tab) {
  font-size: 0.82rem;
  padding: 0.6rem 1rem;
  color: #718096;
  border-bottom: 2px solid transparent;
  border-radius: 0;
  background: transparent;
  white-space: nowrap;
}
.ma-tablist :deep(.p-tab:hover) { color: #06b6d4; background: rgba(6,182,212,0.04); }
.ma-tablist :deep(.p-tab.p-tab-active) { color: #06b6d4; border-bottom-color: #06b6d4; font-weight: 600; }

.ma-panels { padding: 1.25rem 0 0; background: transparent; }

.ma-2col { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; align-items: start; }
@media (max-width: 860px) { .ma-2col { grid-template-columns: 1fr; } }

.ma-block {
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 0.625rem;
  padding: 1.25rem 1.375rem;
}
:root.p-dark .ma-block { background: #1e293b; border-color: rgba(255,255,255,0.07); }
.ma-block--flex { display: flex; flex-direction: column; }
.mt-4 { margin-top: 1rem; }

.ma-block-header { padding-bottom: 0.875rem; margin-bottom: 0.875rem; border-bottom: 1px solid #f0f2f5; }
:root.p-dark .ma-block-header { border-bottom-color: rgba(255,255,255,0.05); }
.ma-block-title { font-size: 0.88rem; font-weight: 700; color: #1a202c; margin: 0 0 0.15rem; }
:root.p-dark .ma-block-title { color: #f1f5f9; }
.ma-block-desc { font-size: 0.73rem; color: #718096; margin: 0; }

.ma-grid { display: grid; gap: 0.75rem 1rem; }
.ma-grid--1 { grid-template-columns: 1fr; }
.ma-grid--2 { grid-template-columns: 1fr 1fr; }
.ma-grid--3 { grid-template-columns: 1fr 1fr 1fr; }
@media (max-width: 680px) { .ma-grid--2, .ma-grid--3 { grid-template-columns: 1fr; } }

.ma-field { display: flex; flex-direction: column; gap: 0.28rem; }
.ma-field--full { grid-column: 1 / -1; }
.ma-label { font-size: 0.71rem; font-weight: 600; color: #4a5568; letter-spacing: 0.01em; }
:root.p-dark .ma-label { color: #94a3b8; }
</style>
