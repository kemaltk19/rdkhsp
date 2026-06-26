<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { usePaymentStore } from '@/stores/payment'

import { useCariStore } from '@/stores/cari'
import { useInvoiceStore } from '@/stores/invoice'
import { getCurrentCompanyDatetimeLocal, toBackendDate } from '@/utils/date'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Menu from 'primevue/menu'
import Money from '@/components/Money.vue'
import FormModal from './FormModal.vue'
import { exportToPDF } from '@/utils/pdfExport'

import { usePermission } from '@/composables/usePermission'

const paymentStore = usePaymentStore()
const cariStore = useCariStore()
const invoiceStore = useInvoiceStore()
const toast = useToast()
const { can } = usePermission()

const showPaymentModal = ref(false)
const paymentType = ref<'collection' | 'payment'>('collection')

// Transfer (Virman) state
const showTransferModal = ref(false)
const transferForm = ref({
  from_account_key: '', // e.g. "cash:id"
  to_account_key: '',
  amount: 0,
  date: getCurrentCompanyDatetimeLocal(),
  description: ''
})

// Logs state
const showLogsModal = ref(false)
const activeLogAccount = ref<{ kind: string; name: string } | null>(null)

// Search & filters
const searchQuery = ref('')
const selectedType = ref('')
const selectedStatus = ref('')
const first = ref(0)
const rows = ref(20)
const page = ref(1)
const dt = ref()
const selectedItems = ref([])
const sortField = ref('')
const sortOrder = ref(1)

const typeOptions = ref([
  { label: 'Tüm Türler', value: '' },
  { label: 'Tahsilat (Giriş)', value: 'collection' },
  { label: 'Ödeme (Çıkış)', value: 'payment' },
])

const statusOptions = ref([
  { label: 'Tüm Durumlar', value: '' },
  { label: 'Tamamlandı', value: 'completed' },
  { label: 'İptal Edildi', value: 'canceled' },
])

const loadData = async () => {
  const params: any = {
    page: page.value,
    limit: rows.value,
    q: searchQuery.value,
    type: selectedType.value,
    status: selectedStatus.value,
    sort: sortField.value ? `${sortField.value} ${sortOrder.value === 1 ? 'asc' : 'desc'}` : ''
  }
  await paymentStore.fetchPayments(params)
  await paymentStore.fetchAccounts()
}

const onSort = (event: any) => {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  loadData()
}

onMounted(async () => {
  loadData()
  await cariStore.fetchCaris({ page: 1, limit: 1000 })
  await invoiceStore.fetchInvoices({ page: 1, limit: 1000 })
})

watch([searchQuery, selectedType, selectedStatus], () => {
  page.value = 1
  first.value = 0
  loadData()
})

const onPage = (event: any) => {
  page.value = event.page + 1
  rows.value = event.rows
  first.value = event.first
  loadData()
}

const openCollection = () => {
  paymentType.value = 'collection'
  showPaymentModal.value = true
}

const openPayment = () => {
  paymentType.value = 'payment'
  showPaymentModal.value = true
}

const cancelPaymentItem = async (id: string) => {
  if (confirm('Bu ödeme/tahsilat işlemini iptal etmek istediğinize emin misiniz? Cari bakiye, fatura tahsilat miktarı ve kasa durumları ters hareketle düzeltilecektir.')) {
    try {
      await paymentStore.cancelPayment(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'İşlem iptal edildi', life: 10000 })
      loadData()
    } catch (err: any) {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'İptal işlemi gerçekleştirilemedi', life: 10000 })
    }
  }
}

// Virman (Transfer) Handlers
const openTransfer = () => {
  transferForm.value = {
    from_account_key: '',
    to_account_key: '',
    amount: 0,
    date: getCurrentCompanyDatetimeLocal(),
    description: ''
  }
  showTransferModal.value = true
}

const combinedAccounts = computed(() => {
  const cashList = paymentStore.cashAccounts.map(c => ({
    label: `${c.name} (Kasa - ${c.currency})`,
    value: `cash:${c.id}`,
    kind: 'cash',
    id: c.id,
    currency: c.currency
  }))
  const bankList = paymentStore.bankAccounts.map(b => ({
    label: `${b.name} (Banka - ${b.currency})`,
    value: `bank:${b.id}`,
    kind: 'bank',
    id: b.id,
    currency: b.currency
  }))
  return [...cashList, ...bankList]
})

const handleTransfer = async () => {
  if (!transferForm.value.from_account_key || !transferForm.value.to_account_key) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Lütfen kaynak ve hedef hesapları seçin', life: 10000 })
    return
  }

  const fromOption = combinedAccounts.value.find(o => o.value === transferForm.value.from_account_key)
  const toOption = combinedAccounts.value.find(o => o.value === transferForm.value.to_account_key)

  if (!fromOption || !toOption) return

  if (fromOption.currency !== toOption.currency) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Para birimleri farklı hesaplar arasında virman yapılamaz.', life: 10000 })
    return
  }

  if (fromOption.value === toOption.value) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Aynı hesaba transfer yapılamaz.', life: 10000 })
    return
  }

  try {
    const payload = {
      from_kind: fromOption.kind,
      from_id: fromOption.id,
      to_kind: toOption.kind,
      to_id: toOption.id,
      amount: transferForm.value.amount.toString(),
      date: toBackendDate(transferForm.value.date),
      description: transferForm.value.description
    }

    await paymentStore.transferCash(payload)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Transfer başarıyla gerçekleştirildi', life: 10000 })
    showTransferModal.value = false
    loadData()
  } catch (err: any) {
    if (err.response && err.response.data && err.response.data.error) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Transfer gerçekleştirilemedi', life: 10000 })
    }
  }
}

// Log history Dialog Handlers
const viewAccountLogs = async (kind: 'cash' | 'bank', acc: any) => {
  activeLogAccount.value = { kind, name: acc.name }
  try {
    await paymentStore.fetchCashTransactions({
      account_kind: kind,
      account_id: acc.id
    })
    showLogsModal.value = true
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Hesap hareketleri yüklenemedi', life: 10000 })
  }
}

// Format helpers
const getCariName = (cariId: string) => {
  const cari = cariStore.caris.find(c => c.id === cariId)
  return cari ? cari.name : 'Yükleniyor...'
}

const getAccountName = (kind: string, id: string) => {
  if (kind === 'cash') {
    const acc = paymentStore.cashAccounts.find(a => a.id === id)
    return acc ? `${acc.name} (Kasa)` : 'Kasa Hesabı'
  } else {
    const acc = paymentStore.bankAccounts.find(a => a.id === id)
    return acc ? `${acc.name} (Banka)` : 'Banka Hesabı'
  }
}

const getPaymentTypeLabel = (type: string) => {
  return type === 'collection' ? 'Tahsilat' : 'Ödeme'
}

const getPaymentTypeSeverity = (type: string) => {
  return type === 'collection' ? 'success' : 'warn'
}

const getStatusLabel = (status: string) => {
  return status === 'completed' ? 'Tamamlandı' : 'İptal Edildi'
}

const getStatusSeverity = (status: string) => {
  return status === 'completed' ? 'success' : 'danger'
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('tr-TR')
}

const getSourceTypeLabel = (src: string) => {
  switch (src) {
    case 'payment': return 'Cari Ödeme'
    case 'expense': return 'Gider Ödemesi'
    case 'transfer': return 'Virman Transfer'
    case 'manual': return 'Manuel Harek'
    default: return src
  }
}

const exportMenu = ref()
const exportOptions = [
  { label: 'PDF', icon: 'pi pi-file-pdf', command: () => exportData('pdf') },
  { label: 'Excel', icon: 'pi pi-file-excel', command: () => exportData('excel') }
]

const toggleExportMenu = (event: any) => {
  exportMenu.value.toggle(event)
}

const exportData = (format: string) => {
  if (format === 'excel') {
    dt.value.exportCSV({ selectionOnly: selectedItems.value.length > 0 })
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Excel olarak dışa aktarıldı', life: 10000 })
  } else if (format === 'pdf') {
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : paymentStore.payments
    const columns = [
      { header: 'Tarih', dataKey: 'date' },
      { header: 'Cari Hesap', dataKey: 'cari_name' },
      { header: 'Tür', dataKey: 'type' },
      { header: 'Kasa/Banka', dataKey: 'account_name' },
      { header: 'Tutar', dataKey: 'amount' },
      { header: 'Döviz', dataKey: 'currency' },
      { header: 'Durum', dataKey: 'status' }
    ]
    exportToPDF('Kasa_Hareketleri_Listesi', columns, dataToExport.map(item => ({
      ...item,
      date: formatDate(item.date),
      cari_name: getCariName(item.cari_id),
      type: getPaymentTypeLabel(item.type),
      account_name: getAccountName(item.account_kind, item.account_id),
      status: getStatusLabel(item.status)
    })))
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'PDF olarak dışa aktarıldı', life: 10000 })
  }
}
const exportCSV = () => {
  exportData('excel')
}
</script>

<template>
  <div class="payments-list-container">
    <!-- Accounts Balance Summary Cards Grid Removed -->

    <!-- Table Card -->
    <Card class="table-card">
      <template #content>
        <!-- Filters Header -->
        <div class="filters-header flex flex-col md:flex-row justify-between items-center gap-4 mb-4">
          <!-- Left: Filters -->
          <div class="select-filters flex gap-2 w-full md:w-auto">
            <Select
              v-model="selectedType"
              :options="typeOptions"
              optionLabel="label"
              optionValue="value"
              placeholder="İşlem Türü"
              class="type-select w-full md:w-40"
            />
            <Select
              v-model="selectedStatus"
              :options="statusOptions"
              optionLabel="label"
              optionValue="value"
              placeholder="Durum"
              class="status-select w-full md:w-40"
            />
          </div>

          <!-- Middle: Search -->
          <div class="flex-1 flex justify-center w-full md:w-auto">
            <div class="search-input w-full max-w-md relative">
              <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"></i>
              <InputText v-model="searchQuery" placeholder="Referans veya açıklama ile ara..." class="w-full pl-10" />
            </div>
          </div>

          <!-- Right: Buttons -->
          <div class="flex gap-2 w-full md:w-auto justify-end shrink-0">
            <Button v-if="can('payments', 'create')" label="Tahsilat" icon="pi pi-plus" @click="openCollection" severity="success" />
            <Button v-if="can('payments', 'create')" label="Ödeme" icon="pi pi-minus" @click="openPayment" severity="success" />
            <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
            <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
          </div>
        </div>

        <DataTable
          ref="dt"
          :value="paymentStore.payments"
          v-model:selection="selectedItems"
          lazy
          paginator
          :first="first"
          :rows="rows"
          :rowsPerPageOptions="[20, 50, 100]"
          :totalRecords="paymentStore.total"
          :loading="paymentStore.loading"
          @page="onPage"
          @sort="onSort"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          dataKey="id"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ (page - 1) * rows + slotProps.index + 1 }}
            </template>
          </Column>
          <Column selectionMode="multiple" headerStyle="width: 3rem"></Column>
          <Column field="reference" header="Fiş/Ref No" sortable style="min-width: 120px; width: 12%">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200 text-sm">{{ data.reference || '-' }}</span>
            </template>
          </Column>
          <Column field="date" header="Tarih" sortable style="min-width: 120px; width: 12%">
            <template #body="{ data }">
              <span>{{ formatDate(data.date) }}</span>
            </template>
          </Column>
          <Column field="cari_id" header="Cari Hesap" sortable style="min-width: 200px; width: 25%">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200 truncate max-w-[250px] block" :title="getCariName(data.cari_id)">{{ getCariName(data.cari_id) }}</span>
            </template>
          </Column>
          <Column field="type" header="Tür" sortable style="min-width: 120px; width: 12%">
            <template #body="{ data }">
              <Tag :value="getPaymentTypeLabel(data.type)" :severity="getPaymentTypeSeverity(data.type)" />
            </template>
          </Column>
          <Column field="account_id" header="Kasa/Banka" sortable style="min-width: 200px; width: 20%">
            <template #body="{ data }">
              <span class="truncate max-w-[180px] inline-block text-xs" :title="getAccountName(data.account_kind, data.account_id)">{{ getAccountName(data.account_kind, data.account_id) }}</span>
            </template>
          </Column>
          <Column field="amount" header="Tutar" sortable style="min-width: 120px; width: 15%" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200">
                <Money :value="data.amount" :currency="data.currency" />
              </span>
            </template>
          </Column>
          <Column field="status" header="Durum" sortable style="min-width: 130px; width: 10%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Tag :value="getStatusLabel(data.status)" :severity="getStatusSeverity(data.status)" />
            </template>
          </Column>
          <Column field="created_by_user" header="Oluşturan" style="min-width: 130px; width: 12%">
            <template #body="{ data }">
              <span class="truncate max-w-[160px] inline-block text-xs text-slate-600 dark:text-slate-400" :title="data.created_by_user?.name">{{ data.created_by_user?.name || '-' }}</span>
            </template>
          </Column>
          <Column header="Aksiyonlar" style="min-width: 100px; width: 6%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <div class="actions-cell">
                <Button v-if="data.status !== 'canceled' && can('payments', 'delete')" icon="pi pi-ban" class="rounded-md p-button-text" @click="cancelPaymentItem(data.id)" title="İşlemi İptal Et" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Transfer (Virman) Dialog -->
    <Dialog v-model:visible="showTransferModal" header="Hesaplar Arası Transfer (Virman)" :modal="true" :style="{ width: '95vw', maxWidth: '800px' }">
      <div class="flex flex-col gap-4 py-2">
        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Kaynak Hesap *</label>
          <Select
            v-model="transferForm.from_account_key"
            :options="combinedAccounts"
            optionLabel="label"
            optionValue="value"
            placeholder="Gönderen hesabı seçin..."
            class="w-full"
            filter
          />
        </div>

        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Hedef Hesap *</label>
          <Select
            v-model="transferForm.to_account_key"
            :options="combinedAccounts"
            optionLabel="label"
            optionValue="value"
            placeholder="Alıcı hesabı seçin..."
            class="w-full"
            filter
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-semibold">Miktar *</label>
            <InputNumber v-model="transferForm.amount" mode="decimal" :minFractionDigits="2" class="w-full" />
          </div>

          <div class="flex flex-col gap-1">
            <label class="text-xs font-semibold">Tarih *</label>
            <InputText type="datetime-local" v-model="transferForm.date" class="w-full" />
          </div>
        </div>

        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Açıklama</label>
          <InputText v-model="transferForm.description" placeholder="Ör: Günlük kasa devri, virman..." class="w-full" />
        </div>

        <div class="flex justify-end gap-2 mt-4">
          <Button label="İptal" class="p-button-text p-button-secondary" @click="showTransferModal = false" outlined />
          <Button label="Transferi Gerçekleştir" icon="pi pi-check" @click="handleTransfer" outlined severity="primary" />
        </div>
      </div>
    </Dialog>

    <!-- Cash Transaction Logs history Dialog -->
    <Dialog
      v-model:visible="showLogsModal"
      :header="activeLogAccount ? `${activeLogAccount.name} - Hesap Hareketleri Logu` : 'Kasa/Banka Hesap Hareketleri'"
      :modal="true"
      :style="{ width: '90%', maxWidth: '850px' }"
    >
      <div v-if="paymentStore.loading" class="text-center py-8">
        <i class="pi pi-spin pi-spinner text-2xl mb-2"></i>
        <div>Yükleniyor...</div>
      </div>
      <div v-else>
        <DataTable
          :value="paymentStore.transactions"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          paginator
          :rows="10"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ slotProps.index + 1 }}
            </template>
          </Column>
          <Column field="date" header="Tarih" style="min-width: 150px">
            <template #body="{ data }">
              <span class="text-xs">{{ formatDate(data.date) }}</span>
            </template>
          </Column>
          <Column field="type" header="Tür" style="min-width: 120px">
            <template #body="{ data }">
              <Tag :value="data.type === 'in' ? 'Giriş (+)' : 'Çıkış (-)'" :severity="data.type === 'in' ? 'success' : 'danger'" class="text-xs" />
            </template>
          </Column>
          <Column field="amount" header="Tutar" style="min-width: 150px">
            <template #body="{ data }">
              <span class="font-bold text-xs">
                <Money :value="data.amount" />
              </span>
            </template>
          </Column>
          <Column field="balance_after" header="Bakiye Sonrası" style="min-width: 150px">
            <template #body="{ data }">
              <span class="font-bold text-xs text-slate-700 dark:text-slate-300">
                <Money :value="data.balance_after" />
              </span>
            </template>
          </Column>
          <Column field="source_type" header="Kaynak" style="min-width: 150px">
            <template #body="{ data }">
              <Tag :value="getSourceTypeLabel(data.source_type)" severity="info" class="text-xs" />
            </template>
          </Column>
          <Column field="description" header="Açıklama" style="min-width: 280px">
            <template #body="{ data }">
              <span class="text-xs text-slate-500">{{ data.description }}</span>
            </template>
          </Column>
        </DataTable>
      </div>
    </Dialog>

    <!-- Collection/Payment Creation Form Modal -->
    <FormModal
      v-if="showPaymentModal"
      v-model:visible="showPaymentModal"
      :type="paymentType"
      @saved="loadData"
    />
  </div>
</template>

<style scoped>
.payments-list-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
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

.account-summary-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .account-summary-card {
  border-color: #334155;
  background-color: #1e293b;
}

.table-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .table-card {
  border-color: #334155;
  background-color: #1e293b;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.search-input {
  position: relative;
  flex: 1;
  min-width: 280px;
  max-width: 400px;
}

.search-input input {
  padding-left: 2.5rem;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--p-text-muted-color, #94a3b8);
}

.select-filters {
  display: flex;
  gap: 0.5rem;
}

.type-select, .status-select {
  width: 160px;
}

.actions-cell {
  display: flex;
  justify-content: center;
  gap: 0.25rem;
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

.grid {
  display: grid;
}

.grid-cols-1 {
  grid-template-columns: repeat(1, 1fr);
}

.gap-6 {
  gap: 1.5rem;
}

.space-y-3 > * + * {
  margin-top: 0.75rem;
}


/* Helper to override PrimeVue Button paddings when nested */
.p-button-sm {
  font-size: 0.8rem;
}
</style>

