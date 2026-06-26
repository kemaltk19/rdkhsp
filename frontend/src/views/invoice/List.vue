<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useInvoiceStore } from '@/stores/invoice'
import { useCariStore } from '@/stores/cari'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Menu from 'primevue/menu'
import { formatDate, formatDateTime } from '@/utils/date'
import Money from '@/components/Money.vue'
// xlsx / jspdf ana bundle'a girmesin diye dışa aktarım anında dinamik yüklenir.

import { usePermission } from '@/composables/usePermission'

const invoiceStore = useInvoiceStore()
const cariStore = useCariStore()
const router = useRouter()
const toast = useToast()
const { can } = usePermission()

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
  { label: 'Tüm Tipler', value: '' },
  { label: 'Satış Faturası', value: 'sales' },
  { label: 'Alış Faturası', value: 'purchase' },
])

const statusOptions = ref([
  { label: 'Tüm Durumlar', value: '' },
  { label: 'Taslak', value: 'draft' },
  // Liste hem satış hem alışı kapsar; 'sent' her ikisinde ortak (satış=Gönderildi, alış=Alındı).
  { label: 'Gönderildi / Alındı', value: 'sent' },
  { label: 'İtiraz Edildi', value: 'disputed' },
  { label: 'Kısmi Ödendi', value: 'partial' },
  { label: 'Ödendi', value: 'paid' },
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
  await invoiceStore.fetchInvoices(params)
}

const onSort = (event: any) => {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  loadData()
}

onMounted(async () => {
  loadData()
  // Fetch caris for name resolution if needed (or backend can return cari name inside JSON)
  // Let's verify: does backend return cari object? Let's verify how it is returned. 
  // GORM usually preloads or we map, but let's load all caris to resolve if needed or read preloaded.
  // Wait, in Go we had `CariID` but did we preload `Cari` or resolve it? The backend returns `CariID` but in database model we can resolve cari info.
  // Actually, let's load all caris to keep a local map for lookup to be safe!
  await cariStore.fetchCaris({ page: 1, limit: 1000 })
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

const openNew = () => {
  router.push('/invoices/new')
}

const editInvoice = (id: string) => {
  router.push(`/invoices/${id}/edit`)
}

const viewDetail = (id: string) => {
  router.push(`/invoices/${id}`)
}

const getCari = (cariId: string) => {
  return cariStore.caris.find(c => c.id === cariId)
}

const cancelInvoiceItem = async (id: string) => {
  if (confirm('Bu faturayı iptal etmek istediğinize emin misiniz? Cari hesap ekstrelerine ters hareket kaydı atılacaktır.')) {
    try {
      await invoiceStore.cancelInvoice(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Fatura iptal edildi', life: 10000 })
      loadData()
      const { useNotificationStore } = await import('@/stores/notification')
      useNotificationStore().fetchNotifications()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Fatura iptal edilemedi', life: 10000 })
      }
    }
  }
}

const deleteInvoiceItem = async (id: string) => {
  if (confirm('Taslak faturayı silmek istediğinize emin misiniz?')) {
    try {
      await invoiceStore.deleteInvoice(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Taslak fatura silindi', life: 10000 })
      loadData()
      const { useNotificationStore } = await import('@/stores/notification')
      useNotificationStore().fetchNotifications()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Fatura silinemedi', life: 10000 })
      }
    }
  }
}

const sendInvoiceItem = async (id: string) => {
  if (confirm('Faturayı müşteriye (veya satıcıya) e-posta ile göndermek istediğinize emin misiniz?')) {
    try {
      await invoiceStore.sendInvoice(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Fatura gönderimi başlatıldı.', life: 10000 })
      loadData()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Fatura gönderilemedi.', life: 10000 })
      }
    }
  }
}

const bulkSending = ref(false)

const bulkSendInvoices = async () => {
  const ids = (selectedItems.value as any[]).map(inv => inv.id)
  if (ids.length === 0) return
  if (!confirm(`Seçilen ${ids.length} faturayı carilere e-posta ile göndermek istediğinize emin misiniz?`)) return

  bulkSending.value = true
  try {
    const result = await invoiceStore.bulkSendInvoice(ids)
    const sentCount = result.sent.length
    const failedCount = result.failed.length
    if (failedCount === 0) {
      toast.add({ severity: 'success', summary: 'Başarılı', detail: `${sentCount} fatura gönderildi.`, life: 10000 })
    } else {
      toast.add({
        severity: 'warn',
        summary: 'Kısmen tamamlandı',
        detail: `${sentCount} fatura gönderildi, ${failedCount} fatura gönderilemedi.`,
        life: 10000,
      })
    }
    selectedItems.value = []
    loadData()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Toplu gönderim başarısız oldu.', life: 10000 })
  } finally {
    bulkSending.value = false
  }
}

const getCariName = (cariId: string) => {
  const cari = cariStore.caris.find(c => c.id === cariId)
  return cari ? cari.name : 'Yükleniyor...'
}

const getInvoiceTypeLabel = (type: string) => {
  return type === 'sales' ? 'Satış' : 'Alış'
}

const getInvoiceTypeSeverity = (type: string) => {
  return type === 'sales' ? 'success' : 'warn'
}

// 'sent' DB değeri her iki tipte ortaktır; ekran etiketi tipe göre değişir:
// satış müşteriye GÖNDERİLİR, alış tedarikçiden ALINIR.
const getStatusLabel = (status: string, type?: string) => {
  switch (status) {
    case 'draft': return 'Taslak'
    case 'sent': return type === 'purchase' ? 'Alındı' : 'Gönderildi'
    case 'disputed': return 'İtiraz Edildi'
    case 'partial': return 'Kısmi Ödendi'
    case 'paid': return 'Ödendi'
    case 'canceled': return 'İptal Edildi'
    default: return status
  }
}

const getStatusSeverity = (status: string) => {
  switch (status) {
    case 'draft': return 'secondary'
    case 'sent': return 'info'
    case 'disputed': return 'danger'
    case 'partial': return 'warn'
    case 'paid': return 'success'
    case 'canceled': return 'danger'
    default: return 'secondary'
  }
}

// Centralized date format helpers are imported from @/utils/date

const exportMenu = ref()
const exportOptions = [
  { label: 'PDF', icon: 'pi pi-file-pdf', command: () => exportData('pdf') },
  { label: 'Excel', icon: 'pi pi-file-excel', command: () => exportData('excel') }
]

const toggleExportMenu = (event: any) => {
  exportMenu.value.toggle(event)
}

const exportData = async (format: string) => {
  if (format === 'excel') {
    const XLSX = await import('xlsx')
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : invoiceStore.invoices
    const data = dataToExport.map((inv: any) => ({
      'Fatura No': inv.number || 'Taslak',
      'Tip': getInvoiceTypeLabel(inv.type),
      'Cari': getCariName(inv.cari_id),
      'Tarih': formatDate(inv.date),
      'Vade': formatDate(inv.due_date),
      'Tutar': inv.total,
      'Döviz': inv.currency,
      'Durum': getStatusLabel(inv.status, inv.type)
    }))
    
    const ws = XLSX.utils.json_to_sheet(data)
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, "Faturalar")
    XLSX.writeFile(wb, "faturalar.xlsx")
    
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Excel olarak dışa aktarıldı', life: 10000 })
  } else if (format === 'pdf') {
    const { jsPDF } = await import('jspdf')
    const { default: autoTable } = await import('jspdf-autotable')
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : invoiceStore.invoices
    const normalizeTurkish = (text: string) => {
      if (!text) return ''
      return text.toString()
        .replace(/ğ/g, 'g').replace(/Ğ/g, 'G')
        .replace(/ü/g, 'u').replace(/Ü/g, 'U')
        .replace(/ş/g, 's').replace(/Ş/g, 'S')
        .replace(/ı/g, 'i').replace(/İ/g, 'I')
        .replace(/ö/g, 'o').replace(/Ö/g, 'O')
        .replace(/ç/g, 'c').replace(/Ç/g, 'C')
    }

    const doc = new jsPDF()
    const head = [['Fatura No', 'Tip', 'Cari', 'Tarih', 'Vade', 'Tutar', 'Doviz', 'Durum']]
    const body = dataToExport.map((inv: any) => [
      normalizeTurkish(inv.number || 'Taslak'),
      normalizeTurkish(getInvoiceTypeLabel(inv.type)),
      normalizeTurkish(getCariName(inv.cari_id)),
      normalizeTurkish(formatDate(inv.date)),
      normalizeTurkish(formatDate(inv.due_date)),
      inv.total,
      normalizeTurkish(inv.currency),
      normalizeTurkish(getStatusLabel(inv.status, inv.type))
    ])
    
    autoTable(doc, {
      head: head,
      body: body,
      theme: 'grid',
      styles: { font: 'helvetica', fontSize: 8 },
      headStyles: { fillColor: [6, 182, 212] }
    })
    
    doc.save('faturalar.pdf')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'PDF olarak dışa aktarıldı', life: 10000 })
  }
}


const exportCSV = () => {
  exportData('excel')
}
</script>

<template>
  <div class="invoice-list-container">

    <Card class="table-card">
      <template #content>
        <div class="filters-header flex flex-col md:flex-row justify-between items-center gap-4 mb-4">
          <div class="select-filters flex gap-2 w-full md:w-auto">
            <Select
              v-model="selectedType"
              :options="typeOptions"
              optionLabel="label"
              optionValue="value"
              placeholder="Fatura Tipi"
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

          <div class="flex-1 flex justify-center w-full md:w-auto">
            <div class="search-input w-full max-w-md relative">
              <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"></i>
              <InputText v-model="searchQuery" placeholder="Fatura numarası veya not ile ara..." class="w-full pl-10" />
            </div>
          </div>

          <div class="flex gap-2 w-full md:w-auto justify-end shrink-0">
            <Button v-if="can('invoices', 'update') && selectedItems.length > 0"
              :label="`Seçilenleri Gönder (${selectedItems.length})`"
              icon="pi pi-send"
              class="p-button-info"
              :loading="bulkSending"
              @click="bulkSendInvoices"
              severity="warn"
            />
            <Button v-if="can('invoices', 'create')" label="Ekle" icon="pi pi-plus" @click="openNew" severity="success" />
            <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
            <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
          </div>
        </div>

        <DataTable
          ref="dt"
          :value="invoiceStore.invoices"
          v-model:selection="selectedItems"
          lazy
          paginator
          :first="first"
          :rows="rows"
          :rowsPerPageOptions="[20, 50, 100]"
          :totalRecords="invoiceStore.total"
          :loading="invoiceStore.loading"
          @page="onPage"
          @sort="onSort"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          dataKey="id"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-[11px] font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ (page - 1) * rows + slotProps.index + 1 }}
            </template>
          </Column>
          <Column selectionMode="multiple" headerStyle="width: 3rem"></Column>
          <Column field="number" header="Fatura No" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              <div 
                class="flex items-center gap-1.5 cursor-pointer hover:text-cyan-600 transition-colors truncate max-w-[170px]"
                @click="viewDetail(data.id)"
                :title="data.number || 'Taslak'"
              >
                <Tag v-if="data.type" :value="getInvoiceTypeLabel(data.type)" :severity="getInvoiceTypeSeverity(data.type) as any" style="font-size: 10px; padding: 0.15rem 0.3rem; line-height: 1;" />
                <span v-if="data.type" class="text-slate-400 dark:text-slate-500 font-medium text-[12px] opacity-70">/</span>
                <span class="font-semibold text-[13px] text-slate-700 dark:text-slate-200 leading-tight">
                  {{ data.number || 'Taslak' }}
                </span>
              </div>
            </template>
          </Column>
          <Column field="cari_id" header="Cari / Müşteri" style="min-width: 200px; width: 25%">
            <template #body="{ data }">
              <div 
                v-if="getCari(data.cari_id)" 
                class="name-cell cursor-pointer hover:underline flex flex-col justify-center"
                @click="router.push(`/caris/${data.cari_id}`)"
              >
                <div class="text-[13px] font-semibold text-slate-700 dark:text-slate-200 truncate max-w-[250px] leading-tight" :title="getCari(data.cari_id)?.name">
                  {{ getCari(data.cari_id)?.name }}
                </div>
                <div 
                  v-if="getCari(data.cari_id)?.contact_name" 
                  class="text-[12px] text-slate-500 dark:text-slate-400 truncate max-w-[250px] leading-tight mt-0.5" 
                  :title="getCari(data.cari_id)?.contact_name"
                >
                  {{ getCari(data.cari_id)?.contact_name }}
                </div>
              </div>
              <span v-else class="text-[12px] text-slate-400">Yükleniyor...</span>
            </template>
          </Column>
          <Column field="date" header="Tarih" sortable style="min-width: 150px; width: 15%" bodyClass="text-[13px] text-slate-600 dark:text-slate-300">
            <template #body="{ data }">
              {{ formatDate(data.date) }}
            </template>
          </Column>
          <Column field="due_date" header="Vade" sortable style="min-width: 120px; width: 12%" bodyClass="text-[13px] text-slate-600 dark:text-slate-300">
            <template #body="{ data }">
              {{ formatDate(data.due_date) }}
            </template>
          </Column>
          <Column field="total" header="Toplam / Kalan" sortable style="min-width: 130px; width: 15%" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <div class="font-semibold text-[13px] text-slate-700 dark:text-slate-200 leading-tight">
                <Money :value="data.total" :currency="data.currency" />
              </div>
              <div v-if="data.status === 'partial'" class="partial-remaining leading-tight mt-0">
                <i class="pi pi-hourglass text-[10px]"></i>
                <Money :value="(parseFloat(data.total) - parseFloat(data.paid_total || '0')).toString()" :currency="data.currency" />
              </div>
            </template>
          </Column>
          <Column field="status" header="Durum" sortable style="min-width: 150px; width: 13%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Tag :value="getStatusLabel(data.status, data.type)" :severity="getStatusSeverity(data.status)" style="font-size: 11px; padding: 0.2rem 0.5rem; line-height: 1;" />
            </template>
          </Column>
          <Column header="Aksiyonlar" style="min-width: 100px; width: 8%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <div class="actions-cell flex justify-center items-center gap-0.5">
                <Button icon="pi pi-eye" class="w-7! h-7! p-0! rounded-md p-button-text p-button-info" @click="viewDetail(data.id)" title="Detay" />
                
                <Button v-if="data.status === 'draft' && can('invoices', 'update')" icon="pi pi-pencil" class="w-7! h-7! p-0! rounded-md p-button-text" @click="editInvoice(data.id)" title="Düzenle" severity="warn" />
                
                <Button v-if="data.status === 'draft' && can('invoices', 'delete')" icon="pi pi-trash" class="w-7! h-7! p-0! rounded-md p-button-text" @click="deleteInvoiceItem(data.id)" title="Sil" severity="danger" />
 
                <Button v-if="data.status !== 'draft' && data.status !== 'canceled' && can('invoices', 'update')"
                  icon="pi pi-send" 
                  class="w-7! h-7! p-0! rounded-md p-button-text p-button-info" 
                  @click="sendInvoiceItem(data.id)" 
                  :title="data.status === 'sent' || data.status === 'disputed' ? 'Yeniden Gönder' : 'Gönder'" 
                />
 
                <Button v-if="data.status !== 'draft' && data.status !== 'canceled' && can('invoices', 'delete')" icon="pi pi-ban" class="w-7! h-7! p-0! rounded-md p-button-text" @click="cancelInvoiceItem(data.id)" title="İptal Et" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.invoice-list-container {
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

.partial-remaining {
  font-size: 0.72rem;
  color: #d97706;
  margin-top: 3px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 3px;
  font-weight: 600;
}
:root.p-dark .partial-remaining { color: #fbbf24; }
</style>

