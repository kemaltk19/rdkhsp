<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useQuoteStore } from '@/stores/quote'
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

const quoteStore = useQuoteStore()
const cariStore = useCariStore()
const router = useRouter()
const toast = useToast()

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

const totalRecords = computed(() => quoteStore.total)

const typeOptions = ref([
  { label: 'Tüm Tipler', value: '' },
  { label: 'Satış Teklifi', value: 'sales' },
  { label: 'Alış Teklifi', value: 'purchase' },
])

const statusOptions = ref([
  { label: 'Tüm Durumlar', value: '' },
  { label: 'Taslak', value: 'draft' },
  { label: 'Teklif Gönderildi', value: 'sent' },
  { label: 'Kabul Edildi', value: 'accepted' },
  { label: 'Reddedildi', value: 'rejected' },
  { label: 'Süresi Doldu', value: 'expired' },
  { label: 'Faturalandı', value: 'converted' },
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
  await quoteStore.fetchQuotes(params)
}

const onSort = (event: any) => {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  loadData()
}

onMounted(async () => {
  loadData()
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
  router.push('/quotes/new')
}

const editQuote = (id: string) => {
  router.push(`/quotes/${id}/edit`)
}

const getCari = (cariId: string) => {
  return cariStore.caris.find(c => c.id === cariId)
}

const viewDetail = (id: string) => {
  router.push(`/quotes/${id}`)
}

const deleteQuoteItem = async (id: string) => {
  if (confirm('Taslak teklifi silmek istediğinize emin misiniz?')) {
    try {
      await quoteStore.deleteQuote(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Taslak teklif silindi', life: 10000 })
      loadData()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Teklif silinemedi', life: 10000 })
      }
    }
  }
}

const sendQuote = async (id: string) => {
  if (confirm('Teklifi müşteriye göndermek istediğinize emin misiniz?')) {
    try {
      await quoteStore.sendQuote(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif e-posta ile gönderildi', life: 10000 })
      loadData()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Teklif gönderilemedi', life: 10000 })
      }
    }
  }
}

const bulkSending = ref(false)

const bulkSendQuotes = async () => {
  const ids = (selectedItems.value as any[]).map(q => q.id)
  if (ids.length === 0) return
  if (!confirm(`Seçilen ${ids.length} teklifi carilere e-posta ile göndermek istediğinize emin misiniz?`)) return

  bulkSending.value = true
  try {
    const result = await quoteStore.bulkSendQuote(ids)
    const sentCount = result.sent.length
    const failedCount = result.failed.length
    if (failedCount === 0) {
      toast.add({ severity: 'success', summary: 'Başarılı', detail: `${sentCount} teklif gönderildi.`, life: 10000 })
    } else {
      toast.add({
        severity: 'warn',
        summary: 'Kısmen tamamlandı',
        detail: `${sentCount} teklif gönderildi, ${failedCount} teklif gönderilemedi.`,
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

const getQuoteTypeLabel = (type: string) => {
  return type === 'sales' ? 'Satış' : 'Alış'
}

const getQuoteTypeSeverity = (type: string) => {
  return type === 'sales' ? 'success' : 'warn'
}

const getStatusLabel = (status: string) => {
  switch (status) {
    case 'draft': return 'Taslak'
    case 'sent': return 'Gönderildi'
    case 'accepted': return 'Kabul Edildi'
    case 'rejected': return 'Reddedildi'
    case 'expired': return 'Süresi Doldu'
    case 'converted': return 'Faturalandı'
    default: return status
  }
}

const getStatusSeverity = (status: string) => {
  switch (status) {
    case 'draft': return 'secondary'
    case 'sent': return 'info'
    case 'accepted': return 'success'
    case 'rejected': return 'danger'
    case 'expired': return 'contrast'
    case 'converted': return 'success'
    default: return 'secondary'
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

const exportData = async (format: string) => {
  if (format === 'excel') {
    const XLSX = await import('xlsx')
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : quoteStore.quotes
    const data = dataToExport.map((inv: any) => ({
      'Teklif No': inv.number || 'Taslak',
      'Tip': getQuoteTypeLabel(inv.type),
      'Cari': getCariName(inv.cari_id),
      'Tarih': formatDate(inv.date),
      'Vade': formatDate(inv.expiry_date),
      'Tutar': inv.total,
      'Döviz': inv.currency,
      'Durum': getStatusLabel(inv.status)
    }))
    
    const ws = XLSX.utils.json_to_sheet(data)
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, "Teklifler")
    XLSX.writeFile(wb, "teklifler.xlsx")
    
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Excel olarak dışa aktarıldı', life: 10000 })
  } else if (format === 'pdf') {
    const { jsPDF } = await import('jspdf')
    const { default: autoTable } = await import('jspdf-autotable')
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : quoteStore.quotes
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
    const head = [['Teklif No', 'Tip', 'Cari', 'Tarih', 'Vade', 'Tutar', 'Doviz', 'Durum']]
    const body = dataToExport.map((inv: any) => [
      normalizeTurkish(inv.number || 'Taslak'),
      normalizeTurkish(getQuoteTypeLabel(inv.type)),
      normalizeTurkish(getCariName(inv.cari_id)),
      normalizeTurkish(formatDate(inv.date)),
      normalizeTurkish(formatDate(inv.expiry_date)),
      inv.total,
      normalizeTurkish(inv.currency),
      normalizeTurkish(getStatusLabel(inv.status))
    ])
    
    autoTable(doc, {
      head: head,
      body: body,
      theme: 'grid',
      styles: { font: 'helvetica', fontSize: 8 },
      headStyles: { fillColor: [6, 182, 212] }
    })
    
    doc.save('teklifler.pdf')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'PDF olarak dışa aktarıldı', life: 10000 })
  }
}

const exportCSV = () => {
  exportData('excel')
}
</script>

<template>
  <div class="quote-list-container">

    <Card class="table-card">
      <template #content>
        <div class="filters-header flex flex-col md:flex-row justify-between items-center gap-4 mb-4">
          <div class="select-filters flex gap-2 w-full md:w-auto">
            <Select
              v-model="selectedType"
              :options="typeOptions"
              optionLabel="label"
              optionValue="value"
              placeholder="Teklif Tipi"
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
              <InputText v-model="searchQuery" placeholder="Teklif numarası veya not ile ara..." class="w-full pl-10" />
            </div>
          </div>

          <div class="flex gap-2 w-full md:w-auto justify-end shrink-0">
            <Button v-if="selectedItems.length > 0"
              :label="`Seçilenleri Gönder (${selectedItems.length})`"
              icon="pi pi-send"
              class="p-button-info"
              :loading="bulkSending"
              @click="bulkSendQuotes"
              severity="warn"
            />
            <Button label="Ekle" icon="pi pi-plus" @click="openNew" severity="success" />
            <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
            <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
          </div>
        </div>

        <DataTable
          ref="dt"
          :value="quoteStore.quotes"
          v-model:selection="selectedItems"
          lazy
          paginator
          :first="first"
          :rows="rows"
          :rowsPerPageOptions="[20, 50, 100]"
          :totalRecords="quoteStore.total"
          :loading="quoteStore.loading"
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
          <Column field="number" header="Teklif No" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              <span 
                class="font-medium text-slate-700 dark:text-slate-200 cursor-pointer hover:text-cyan-600 transition-colors truncate max-w-[140px] inline-block align-middle"
                @click="viewDetail(data.id)"
                :title="data.number || 'Taslak'"
              >
                {{ data.number || 'Taslak' }}
              </span>
              <div v-if="data.type" class="mt-1">
                <Tag :value="getQuoteTypeLabel(data.type)" :severity="getQuoteTypeSeverity(data.type) as any" class="text-xs" />
              </div>
            </template>
          </Column>
          <Column field="cari_id" header="Cari / Müşteri" style="min-width: 200px; width: 25%">
            <template #body="{ data }">
              <div 
                v-if="getCari(data.cari_id)" 
                class="name-cell cursor-pointer hover:underline"
                @click="router.push(`/caris/${data.cari_id}`)"
              >
                <div class="name-text">
                  <div class="font-medium text-slate-700 dark:text-slate-200 truncate max-w-[250px]" :title="getCari(data.cari_id)?.name">
                    {{ getCari(data.cari_id)?.name }}
                  </div>
                  <div 
                    v-if="getCari(data.cari_id)?.contact_name" 
                    class="sub text-xs text-slate-500 dark:text-slate-400 mt-1 truncate max-w-[250px]" 
                    :title="getCari(data.cari_id)?.contact_name"
                  >
                    {{ getCari(data.cari_id)?.contact_name }}
                  </div>
                </div>
              </div>
              <span v-else class="text-slate-400">Yükleniyor...</span>
            </template>
          </Column>
          <Column field="date" header="Tarih / Saat" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              {{ formatDateTime(data.date) }}
            </template>
          </Column>
          <Column field="expiry_date" header="Vade" sortable style="min-width: 120px; width: 12%">
            <template #body="{ data }">
              {{ formatDate(data.expiry_date) }}
            </template>
          </Column>
          <Column field="total" header="Toplam Bakiye" sortable style="min-width: 130px; width: 15%" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <div class="font-medium text-slate-700 dark:text-slate-200">
                <Money :value="data.total" :currency="data.currency" />
              </div>
            </template>
          </Column>
          <Column field="status" header="Durum" sortable style="min-width: 150px; width: 13%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Tag :value="getStatusLabel(data.status)" :severity="getStatusSeverity(data.status)" />
            </template>
          </Column>
          <Column header="Aksiyonlar" style="min-width: 100px; width: 8%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <div class="actions-cell">
                <Button icon="pi pi-eye" class="rounded-md p-button-text p-button-info" @click="viewDetail(data.id)" title="Detay" />
                
                <Button v-if="data.status === 'draft' || data.status === 'rejected'"
                  icon="pi pi-send" 
                  class="rounded-md p-button-text p-button-success" 
                  @click="sendQuote(data.id)" 
                  title="Gönder" 
                />
                
                <Button v-if="data.status === 'draft'" icon="pi pi-pencil" class="rounded-md p-button-text" @click="editQuote(data.id)" title="Düzenle" severity="warn" />
                
                <Button v-if="data.status === 'draft'" icon="pi pi-trash" class="rounded-md p-button-text" @click="deleteQuoteItem(data.id)" title="Sil" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>


  </div>
</template>

<style scoped>
.quote-list-container {
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
</style>

