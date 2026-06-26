<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useReportStore } from '@/stores/report'
import { useCariStore } from '@/stores/cari'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Money from '@/components/Money.vue'

const reportStore = useReportStore()
const cariStore = useCariStore()
const toast = useToast()

interface TabItem {
  label: string
  value: string
}

const activeTab = ref('sales')
const dateFrom = ref('')
const dateTo = ref('')
const selectedCari = ref('')

const tabs = ref<TabItem[]>([
  { label: 'Satış Raporu', value: 'sales' },
  { label: 'Alış Raporu', value: 'purchases' },
  { label: 'Cari Yaşlandırma', value: 'cari-aging' },
  { label: 'Stok Değeri', value: 'stock' },
  { label: 'Kasa & Banka Akışı', value: 'cash' },
  { label: 'Kâr & Zarar Raporu', value: 'profit' },
])

const loadReport = async () => {
  const params: any = {}
  if (dateFrom.value) params.date_from = new Date(dateFrom.value).toISOString()
  if (dateTo.value) params.date_to = new Date(dateTo.value).toISOString()
  if (selectedCari.value) params.cari_id = selectedCari.value

  try {
    await reportStore.fetchReport(activeTab.value, params)
  } catch (err) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Rapor yüklenemedi', life: 10000 })
  }
}

onMounted(async () => {
  // Set default date range: start of current year to today
  const now = new Date()
  const startOfYear = new Date(now.getFullYear(), 0, 1)
  dateFrom.value = startOfYear.toISOString().substring(0, 16)
  dateTo.value = now.toISOString().substring(0, 16)

  await cariStore.fetchCaris({ page: 1, limit: 1000 })
  loadReport()
})

// Query when tab changes
watch(activeTab, () => {
  // Sekmeler aynı store.data'yı paylaşıyor; eski sekmenin verisi yeni sekmeye
  // sızmasın diye önce temizle, sonra yeni raporu yükle.
  reportStore.data = null
  loadReport()
})

const queryReport = () => {
  loadReport()
}

const clearFilters = () => {
  dateFrom.value = ''
  dateTo.value = ''
  selectedCari.value = ''
  loadReport()
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('tr-TR')
}

// CSV Export Helper
const exportCSV = () => {
  const data = reportStore.data
  if (!data) return

  let rows: any[] = []
  let headers: string[] = []
  let filename = `${activeTab.value}_raporu.csv`

  if (activeTab.value === 'sales' || activeTab.value === 'purchases') {
    headers = ['Fatura No', 'Cari Hesap', 'Tarih', 'Vade Tarihi', 'Ara Toplam', 'Indirim', 'KDV', 'Toplam', 'Odenen', 'Durum']
    rows = (data.rows || []).map((r: any) => [
      r.number,
      r.cari_name,
      formatDate(r.date),
      formatDate(r.due_date),
      r.subtotal,
      r.discount_total,
      r.tax_total,
      r.total,
      r.paid_total,
      r.status
    ])
  } else if (activeTab.value === 'cari-aging') {
    headers = ['Cari Adı', 'Kalan Borç', 'Vadesi Gelmemiş', '1-30 Gün Gecikmiş', '31-60 Gün Gecikmiş', '61-90 Gün Gecikmiş', '90+ Gün Gecikmiş']
    rows = (data.rows || []).map((r: any) => [
      r.cari_name,
      r.total_unpaid,
      r.not_overdue,
      r.overdue_1_30,
      r.overdue_31_60,
      r.overdue_61_90,
      r.overdue_90_plus
    ])
  } else if (activeTab.value === 'stock') {
    headers = ['Kodu', 'Urun Adı', 'Kategori', 'Stok', 'Birim', 'Toplam Deger']
    rows = (data.rows || []).map((r: any) => [
      r.product_code,
      r.product_name,
      r.category_name,
      r.current_stock,
      r.unit,
      r.valuation
    ])
  } else if (activeTab.value === 'cash') {
    headers = ['Hesap Adı', 'Tipi', 'Doviz', 'Acılıs Bakiyesi', 'Girisler', 'Cıkıslar', 'Net Degisim', 'Kapanıs Bakiyesi']
    rows = (data.rows || []).map((r: any) => [
      r.account_name,
      r.account_kind === 'cash' ? 'Kasa' : 'Banka',
      r.currency,
      r.opening_balance,
      r.inflow,
      r.outflow,
      r.net_change,
      r.ending_balance
    ])
  } else if (activeTab.value === 'profit') {
    headers = ['Ay', 'Satıs Ciro', 'Giderler', 'Net Kar/Zarar']
    rows = (data.rows || []).map((r: any) => [
      r.month,
      r.income,
      r.expenses,
      r.net_profit
    ])
  }

  if (rows.length === 0) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Aktarılacak veri bulunmamaktadır.', life: 10000 })
    return
  }

  // Convert array to CSV string
  let csvContent = '\uFEFF' // BOM for Excel encoding support
  csvContent += headers.join(';') + '\n'
  rows.forEach(row => {
    csvContent += row.map((v: any) => `"${(v ?? '').toString().replace(/"/g, '""')}"`).join(';') + '\n'
  })

  // Trigger browser download
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.setAttribute('href', url)
  link.setAttribute('download', filename)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<template>
  <div class="reports-container">

    <!-- Filter Card -->
    <Card class="filter-card">
      <template #content>
        <div class="filters-grid">
          <div class="field-group" v-if="activeTab !== 'stock' && activeTab !== 'cari-aging'">
            <label for="dateFrom">Başlangıç Tarihi</label>
            <input id="dateFrom" type="datetime-local" v-model="dateFrom" class="p-inputtext w-full" />
          </div>

          <div class="field-group" v-if="activeTab !== 'stock' && activeTab !== 'cari-aging'">
            <label for="dateTo">Bitiş Tarihi</label>
            <input id="dateTo" type="datetime-local" v-model="dateTo" class="p-inputtext w-full" />
          </div>

          <div class="field-group" v-if="activeTab === 'sales' || activeTab === 'purchases'">
            <label for="cari">Cari Hesap Filtresi</label>
            <Select
              id="cari"
              v-model="selectedCari"
              :options="cariStore.caris"
              optionLabel="name"
              optionValue="id"
              placeholder="Tüm Cariler"
              class="w-full"
              filter
              showClear
            />
          </div>

          <div class="filter-actions self-end flex gap-2">
            <Button label="Sorgula" icon="pi pi-search" class="p-button-primary flex-1" @click="queryReport" :loading="reportStore.loading" outlined />
            <Button label="Temizle" icon="pi pi-refresh" class="p-button-text p-button-secondary" @click="clearFilters" :disabled="reportStore.loading" outlined />
            <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="exportCSV" :disabled="reportStore.loading" severity="contrast" />
          </div>
        </div>
      </template>
    </Card>

    <!-- Tabs Container -->
    <Card class="table-card mt-4">
      <template #content>
        <Tabs v-model:value="activeTab">
          <TabList>
            <Tab v-for="tab in tabs" :key="tab.value" :value="tab.value">{{ tab.label }}</Tab>
          </TabList>

          <TabPanels class="pt-4" v-if="reportStore.data">
            <!-- SALES & PURCHASES REPORT -->
            <TabPanel value="sales">
              <DataTable :value="reportStore.data.rows" :loading="reportStore.loading" class="p-datatable-sm" responsiveLayout="scroll">
                <Column field="number" header="Fatura No" style="min-width: 150px" />
                <Column field="cari_name" header="Cari Hesap" style="min-width: 200px">
                  <template #body="{ data }">
                    <span class="truncate max-w-[220px] inline-block" :title="data.cari_name">{{ data.cari_name }}</span>
                  </template>
                </Column>
                <Column field="date" header="Tarih" style="min-width: 120px">
                  <template #body="{ data }">{{ formatDate(data.date) }}</template>
                </Column>
                <Column field="due_date" header="Vade Tarihi" style="min-width: 120px">
                  <template #body="{ data }">{{ formatDate(data.due_date) }}</template>
                </Column>
                <Column field="subtotal" header="Ara Toplam" style="min-width: 120px" headerClass="col-right" bodyClass="col-right">
                  <template #body="{ data }"><Money :value="data.subtotal" /></template>
                </Column>
                <Column field="discount_total" header="İndirim" style="min-width: 100px">
                  <template #body="{ data }"><Money :value="data.discount_total" /></template>
                </Column>
                <Column field="tax_total" header="KDV" style="min-width: 100px">
                  <template #body="{ data }"><Money :value="data.tax_total" /></template>
                </Column>
                <Column field="total" header="Genel Toplam" style="min-width: 150px" headerClass="col-right" bodyClass="col-right">
                  <template #body="{ data }"><Money :value="data.total" class="font-medium text-slate-700 dark:text-slate-200" /></template>
                </Column>
              </DataTable>

              <!-- Summary Card -->
              <div class="summary-total-box mt-4 p-4 rounded-lg bg-slate-50 dark:bg-slate-800 flex justify-end gap-8" v-if="reportStore.data.summary">
                <div>Ara Toplam: <strong><Money :value="reportStore.data.summary.subtotal" /></strong></div>
                <div>İndirim: <strong class="text-red-500"><Money :value="reportStore.data.summary.discount_total" /></strong></div>
                <div>KDV Toplam: <strong><Money :value="reportStore.data.summary.tax_total" /></strong></div>
                <div>Genel Toplam: <strong class="text-lg text-primary"><Money :value="reportStore.data.summary.total" /></strong></div>
              </div>
            </TabPanel>

            <TabPanel value="purchases">
              <DataTable :value="reportStore.data.rows" :loading="reportStore.loading" class="p-datatable-sm" responsiveLayout="scroll">
                <Column field="number" header="Fatura No" style="min-width: 150px" />
                <Column field="cari_name" header="Cari Hesap" style="min-width: 200px">
                  <template #body="{ data }">
                    <span class="truncate max-w-[220px] inline-block" :title="data.cari_name">{{ data.cari_name }}</span>
                  </template>
                </Column>
                <Column field="date" header="Tarih" style="min-width: 120px">
                  <template #body="{ data }">{{ formatDate(data.date) }}</template>
                </Column>
                <Column field="due_date" header="Vade Tarihi" style="min-width: 120px">
                  <template #body="{ data }">{{ formatDate(data.due_date) }}</template>
                </Column>
                <Column field="subtotal" header="Ara Toplam" style="min-width: 120px" headerClass="col-right" bodyClass="col-right">
                  <template #body="{ data }"><Money :value="data.subtotal" /></template>
                </Column>
                <Column field="discount_total" header="İndirim" style="min-width: 100px">
                  <template #body="{ data }"><Money :value="data.discount_total" /></template>
                </Column>
                <Column field="tax_total" header="KDV" style="min-width: 100px">
                  <template #body="{ data }"><Money :value="data.tax_total" /></template>
                </Column>
                <Column field="total" header="Genel Toplam" style="min-width: 150px" headerClass="col-right" bodyClass="col-right">
                  <template #body="{ data }"><Money :value="data.total" class="font-medium text-slate-700 dark:text-slate-200" /></template>
                </Column>
              </DataTable>

              <div class="summary-total-box mt-4 p-4 rounded-lg bg-slate-50 dark:bg-slate-800 flex justify-end gap-8" v-if="reportStore.data.summary">
                <div>Ara Toplam: <strong><Money :value="reportStore.data.summary.subtotal" /></strong></div>
                <div>İndirim: <strong class="text-red-500"><Money :value="reportStore.data.summary.discount_total" /></strong></div>
                <div>KDV Toplam: <strong><Money :value="reportStore.data.summary.tax_total" /></strong></div>
                <div>Genel Toplam: <strong class="text-lg text-primary"><Money :value="reportStore.data.summary.total" /></strong></div>
              </div>
            </TabPanel>

            <!-- CARI AGING REPORT -->
            <TabPanel value="cari-aging">
              <DataTable :value="reportStore.data.rows" :loading="reportStore.loading" class="p-datatable-sm" responsiveLayout="scroll">
                <Column field="cari_name" header="Müşteri Adı" style="min-width: 200px">
                  <template #body="{ data }">
                    <span class="truncate max-w-[220px] inline-block" :title="data.cari_name">{{ data.cari_name }}</span>
                  </template>
                </Column>
                <Column field="total_unpaid" header="Toplam Alacak" style="min-width: 150px" headerClass="col-right" bodyClass="col-right">
                  <template #body="{ data }"><Money :value="data.total_unpaid" class="font-medium text-slate-700 dark:text-slate-200" /></template>
                </Column>
                <Column field="not_overdue" header="Vadesi Gelmemiş" style="min-width: 120px">
                  <template #body="{ data }"><Money :value="data.not_overdue" /></template>
                </Column>
                <Column field="overdue_1_30" header="1-30 Gün Gecikmiş" style="min-width: 120px">
                  <template #body="{ data }"><Money :value="data.overdue_1_30" :class="{ 'text-red-500 font-semibold': parseFloat(data.overdue_1_30) > 0 }" /></template>
                </Column>
                <Column field="overdue_31_60" header="31-60 Gün Gecikmiş" style="min-width: 120px">
                  <template #body="{ data }"><Money :value="data.overdue_31_60" :class="{ 'text-red-500 font-semibold': parseFloat(data.overdue_31_60) > 0 }" /></template>
                </Column>
                <Column field="overdue_61_90" header="61-90 Gün Gecikmiş" style="min-width: 120px">
                  <template #body="{ data }"><Money :value="data.overdue_61_90" :class="{ 'text-red-600 font-bold': parseFloat(data.overdue_61_90) > 0 }" /></template>
                </Column>
                <Column field="overdue_90_plus" header="90+ Gün Gecikmiş" style="min-width: 120px">
                  <template #body="{ data }"><Money :value="data.overdue_90_plus" :class="{ 'text-red-700 font-black': parseFloat(data.overdue_90_plus) > 0 }" /></template>
                </Column>
              </DataTable>
            </TabPanel>

            <!-- STOCK VALUATION REPORT -->
            <TabPanel value="stock">
              <DataTable :value="reportStore.data.rows" :loading="reportStore.loading" class="p-datatable-sm" responsiveLayout="scroll">
                <Column field="product_code" header="Ürün Kodu" style="min-width: 150px" />
                <Column field="product_name" header="Ürün Adı" style="min-width: 200px">
                  <template #body="{ data }">
                    <span class="truncate max-w-[220px] inline-block" :title="data.product_name">{{ data.product_name }}</span>
                  </template>
                </Column>
                <Column field="category_name" header="Kategori" style="min-width: 150px" />
                <Column field="current_stock" header="Stok Miktarı" style="min-width: 150px">
                  <template #body="{ data }">{{ parseFloat(data.current_stock) }} {{ data.unit || 'Adet' }}</template>
                </Column>
                <Column field="valuation" header="Stok Değeri" style="min-width: 150px">
                  <template #body="{ data }"><Money :value="data.valuation" class="font-medium text-slate-700 dark:text-slate-200" /></template>
                </Column>
              </DataTable>

              <div class="summary-total-box mt-4 p-4 rounded-lg bg-slate-50 dark:bg-slate-800 flex justify-end" v-if="reportStore.data.total_valuation !== undefined">
                <div>Toplam Stok Değeri: <strong class="text-lg text-primary"><Money :value="reportStore.data.total_valuation" /></strong></div>
              </div>
            </TabPanel>

            <!-- CASH FLOW REPORT -->
            <TabPanel value="cash">
              <DataTable :value="reportStore.data.rows" :loading="reportStore.loading" class="p-datatable-sm" responsiveLayout="scroll">
                <Column field="account_name" header="Kasa/Banka Adı" style="min-width: 200px">
                  <template #body="{ data }">
                    <span class="truncate max-w-[180px] inline-block" :title="data.account_name">{{ data.account_name }}</span>
                  </template>
                </Column>
                <Column field="account_kind" header="Tür" style="min-width: 100px">
                  <template #body="{ data }">{{ data.account_kind === 'cash' ? 'Kasa' : 'Banka' }}</template>
                </Column>
                <Column field="currency" header="Para Birimi" style="min-width: 100px" />
                <Column field="opening_balance" header="Açılış Bakiyesi" style="min-width: 150px">
                  <template #body="{ data }"><Money :value="data.opening_balance" :currency="data.currency" /></template>
                </Column>
                <Column field="inflow" header="Giren (+)" style="min-width: 130px">
                  <template #body="{ data }"><span class="text-green-600"><Money :value="data.inflow" :currency="data.currency" /></span></template>
                </Column>
                <Column field="outflow" header="Çıkan (-)" style="min-width: 130px">
                  <template #body="{ data }"><span class="text-red-500"><Money :value="data.outflow" :currency="data.currency" /></span></template>
                </Column>
                <Column field="ending_balance" header="Dönem Sonu Bakiyesi" style="min-width: 140px">
                  <template #body="{ data }"><Money :value="data.ending_balance" :currency="data.currency" class="font-medium text-slate-700 dark:text-slate-200" /></template>
                </Column>
              </DataTable>
            </TabPanel>

            <!-- PROFIT & LOSS REPORT -->
            <TabPanel value="profit">
              <DataTable :value="reportStore.data.rows" :loading="reportStore.loading" class="p-datatable-sm" responsiveLayout="scroll">
                <Column field="month" header="Dönem (Ay)" style="min-width: 200px" />
                <Column field="income" header="Gelirler (Satış Ciro)" style="min-width: 200px">
                  <template #body="{ data }"><span class="text-green-600"><Money :value="data.income" /></span></template>
                </Column>
                <Column field="expenses" header="Giderler" style="min-width: 200px">
                  <template #body="{ data }"><span class="text-red-500"><Money :value="data.expenses" /></span></template>
                </Column>
                <Column field="net_profit" header="Net Kâr / Zarar" style="min-width: 200px">
                  <template #body="{ data }">
                    <span :class="parseFloat(data.net_profit) >= 0 ? 'text-green-600 font-medium' : 'text-red-600 font-medium'">
                      <Money :value="data.net_profit" />
                    </span>
                  </template>
                </Column>
              </DataTable>

              <div class="summary-total-box mt-4 p-4 rounded-lg bg-slate-50 dark:bg-slate-800 flex justify-end gap-8" v-if="reportStore.data.total_income">
                <div>Toplam Gelirler: <strong class="text-green-600"><Money :value="reportStore.data.total_income" /></strong></div>
                <div>Toplam Giderler: <strong class="text-red-500"><Money :value="reportStore.data.total_expense" /></strong></div>
                <div>Toplam Kâr / Zarar: 
                  <strong :class="parseFloat(reportStore.data.total_profit) >= 0 ? 'text-lg text-green-600' : 'text-lg text-red-600'">
                    <Money :value="reportStore.data.total_profit" />
                  </strong>
                </div>
              </div>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.reports-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin-bottom: 0.25rem;
}

.page-desc {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.95rem;
}

.filter-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .filter-card {
  border-color: #334155;
  background-color: #1e293b;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.field-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--p-text-color, #475569);
}

:root.p-dark .field-group label {
  color: #cbd5e1;
}

.table-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .table-card {
  border-color: #334155;
  background-color: #1e293b;
}

.summary-total-box {
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .summary-total-box {
  border-color: #334155;
}

.text-green-600 {
  color: #16a34a;
}

.text-red-500 {
  color: #ef4444;
}

.text-red-600 {
  color: #dc2626;
}

.text-red-700 {
  color: #b91c1c;
}

.font-black {
  font-weight: 900;
}

.self-end {
  align-self: flex-end;
}

.flex {
  display: flex;
}

.flex-1 {
  flex: 1;
}

.gap-2 {
  gap: 0.5rem;
}

.gap-8 {
  gap: 2rem;
}

.mt-4 {
  margin-top: 1rem;
}

.pt-4 {
  padding-top: 1rem;
}
</style>

