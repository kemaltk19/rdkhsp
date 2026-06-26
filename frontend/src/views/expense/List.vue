<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useExpenseStore } from '@/stores/expense'
import { useCariStore } from '@/stores/cari'
import { usePaymentStore } from '@/stores/payment'
import { useNotificationStore } from '@/stores/notification'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Menu from 'primevue/menu'

import Money from '@/components/Money.vue'
import FormModal from './FormModal.vue'
import { exportToPDF } from '@/utils/pdfExport'

import { usePermission } from '@/composables/usePermission'

const expenseStore = useExpenseStore()
const cariStore = useCariStore()
const paymentStore = usePaymentStore()
const notificationStore = useNotificationStore()
const toast = useToast()
const { can } = usePermission()

const showExpenseModal = ref(false)
const selectedExpenseId = ref<string | undefined>(undefined)

const searchQuery = ref('')
const selectedCategory = ref('')
const selectedStatus = ref('')
const first = ref(0)
const rows = ref(20)
const page = ref(1)
const dt = ref()
const selectedItems = ref([])
const sortField = ref('')
const sortOrder = ref(1)

const statusOptions = ref([
  { label: 'Tüm Durumlar', value: '' },
  { label: 'Ödendi', value: 'paid' },
  { label: 'Ödenmedi', value: 'unpaid' },
  { label: 'İptal Edildi', value: 'canceled' },
])

const loadData = async () => {
  const params: any = {
    page: page.value,
    limit: rows.value,
    q: searchQuery.value,
    category_id: selectedCategory.value,
    status: selectedStatus.value,
    sort: sortField.value ? `${sortField.value} ${sortOrder.value === 1 ? 'asc' : 'desc'}` : ''
  }
  await expenseStore.fetchExpenses(params)
  await expenseStore.fetchCategories()
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
})

watch([searchQuery, selectedCategory, selectedStatus], () => {
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

const openNewExpense = () => {
  selectedExpenseId.value = undefined
  showExpenseModal.value = true
}

const editExpenseItem = (id: string) => {
  selectedExpenseId.value = id
  showExpenseModal.value = true
}

const cancelExpenseItem = async (id: string) => {
  if (confirm('Bu gider kaydını iptal etmek istediğinize emin misiniz? Yapılan ödeme ve cari bakiye etkileri ters hareketle düzeltilecektir.')) {
    try {
      await expenseStore.cancelExpense(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Gider kaydı iptal edildi', life: 10000 })
      loadData()
      notificationStore.fetchNotifications()
    } catch (err: any) {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'İptal işlemi gerçekleştirilemedi', life: 10000 })
    }
  }
}



const getCariName = (cariId: string | null) => {
  if (!cariId) return '-'
  const cari = cariStore.caris.find(c => c.id === cariId)
  return cari ? cari.name : '-'
}

const getAccountName = (kind: string | null, id: string | null) => {
  if (!kind || !id) return '-'
  if (kind === 'cash') {
    const acc = paymentStore.cashAccounts.find(a => a.id === id)
    return acc ? `${acc.name} (Kasa)` : 'Kasa Hesabı'
  } else {
    const acc = paymentStore.bankAccounts.find(a => a.id === id)
    return acc ? `${acc.name} (Banka)` : 'Banka Hesabı'
  }
}

const getStatusLabel = (status: string) => {
  switch (status) {
    case 'paid': return 'Ödendi'
    case 'unpaid': return 'Ödenmedi'
    case 'canceled': return 'İptal Edildi'
    default: return status
  }
}

const getStatusSeverity = (status: string) => {
  switch (status) {
    case 'paid': return 'success'
    case 'unpaid': return 'warn'
    case 'canceled': return 'danger'
    default: return 'info'
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('tr-TR')
}

// Category filter list
const categoryFilterOptions = computed(() => {
  return [
    { name: 'Tüm Kategoriler', id: '' },
    ...expenseStore.categories,
  ]
})

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
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : expenseStore.expenses
    const columns = [
      { header: 'Tarih', dataKey: 'date' },
      { header: 'Kategori', dataKey: 'category_name' },
      { header: 'Açıklama', dataKey: 'description' },
      { header: 'Tedarikçi', dataKey: 'cari_name' },
      { header: 'Kasa/Banka', dataKey: 'account_name' },
      { header: 'Tutar', dataKey: 'total' },
      { header: 'Durum', dataKey: 'status' }
    ]
    exportToPDF('Gider_Listesi', columns, dataToExport.map(item => ({
      ...item,
      date: formatDate(item.date),
      category_name: item.category ? item.category.name : 'Genel Gider',
      cari_name: getCariName(item.cari_id),
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
  <div class="expenses-list-container">



    <!-- Table Card -->
    <Card class="table-card">
      <template #content>
        <!-- Filters Header -->
        <div class="filters-header flex flex-col md:flex-row justify-between items-center gap-4 mb-4">
          <!-- Left: Filters -->
          <div class="select-filters flex gap-2 w-full md:w-auto">
            <Select
              v-model="selectedCategory"
              :options="categoryFilterOptions"
              optionLabel="name"
              optionValue="id"
              placeholder="Kategori Filtresi"
              class="category-select w-full md:w-48"
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
              <InputText v-model="searchQuery" placeholder="Açıklama veya not ile ara..." class="w-full pl-10" />
            </div>
          </div>

          <!-- Right: Buttons -->
          <div class="flex gap-2 w-full md:w-auto justify-end shrink-0">
            <Button v-if="can('expenses', 'create')" label="Ekle" icon="pi pi-plus" @click="openNewExpense" severity="success" />
            <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
            <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
          </div>
        </div>

        <DataTable
          ref="dt"
          :value="expenseStore.expenses"
          v-model:selection="selectedItems"
          lazy
          paginator
          :first="first"
          :rows="rows"
          :rowsPerPageOptions="[20, 50, 100]"
          :totalRecords="expenseStore.total"
          :loading="expenseStore.loading"
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
          <Column field="date" header="Tarih" sortable style="min-width: 100px; width: 10%">
            <template #body="{ data }">
              <span>{{ formatDate(data.date) }}</span>
            </template>
          </Column>
          <Column field="category_id" header="Kategori" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200">{{ data.category ? data.category.name : 'Genel Gider' }}</span>
            </template>
          </Column>
          <Column field="description" header="Açıklama" sortable style="min-width: 200px; width: 20%">
            <template #body="{ data }">
              <span class="truncate max-w-[280px] inline-block" :title="data.description">{{ data.description }}</span>
            </template>
          </Column>
          <Column field="cari_id" header="Tedarikçi" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              <span class="truncate max-w-[250px] block" :title="getCariName(data.cari_id)">{{ getCariName(data.cari_id) }}</span>
            </template>
          </Column>
          <Column field="account_id" header="Kasa/Banka" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              <span class="text-xs">{{ getAccountName(data.account_kind, data.account_id) }}</span>
            </template>
          </Column>
          <Column field="amount" header="Tutar" sortable style="min-width: 120px; width: 13%" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200">
                <Money :value="data.total" />
              </span>
            </template>
          </Column>
          <Column field="status" header="Durum" sortable style="min-width: 100px; width: 6%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Tag :value="getStatusLabel(data.status)" :severity="getStatusSeverity(data.status)" />
            </template>
          </Column>
          <Column field="created_by_user" header="Oluşturan" style="min-width: 130px; width: 12%">
            <template #body="{ data }">
              <span class="text-xs text-slate-600 dark:text-slate-400">{{ data.created_by_user?.name || '-' }}</span>
            </template>
          </Column>
          <Column field="is_recurring" header="Tekrarlayan" sortable style="min-width: 110px; width: 7%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Tag v-if="data.is_recurring" value="Tekrarlayan" severity="info" icon="pi pi-refresh" />
              <span v-else class="text-xs text-slate-400">-</span>
            </template>
          </Column>
          <Column header="Aksiyonlar" style="min-width: 100px; width: 6%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <div class="actions-cell">
                <Button v-if="data.status !== 'canceled' && can('expenses', 'update')" icon="pi pi-pencil" class="rounded-md p-button-text" @click="editExpenseItem(data.id)" title="Düzenle" severity="warn" />
                <Button v-if="data.status !== 'canceled' && can('expenses', 'delete')" icon="pi pi-ban" class="rounded-md p-button-text" @click="cancelExpenseItem(data.id)" title="İptal Et" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Expense Modal Component -->
    <FormModal
      v-if="showExpenseModal"
      v-model:visible="showExpenseModal"
      :expenseId="selectedExpenseId"
      @saved="() => { loadData(); notificationStore.fetchNotifications(); }"
    />
  </div>
</template>

<style scoped>
.expenses-list-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.table-card {
  border-radius: 12px;
}

.filters-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.25rem;
  flex-wrap: wrap;
}

.select-filters {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}

.category-select {
  width: 180px;
}

.status-select {
  width: 150px;
}

.search-input {
  position: relative;
  width: 100%;
  max-width: 420px;
  display: block;
}

.search-input i {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
}

.search-input :deep(input) {
  width: 100%;
  padding-left: 2.4rem;
}

.actions-cell {
  display: flex;
  justify-content: center;
  gap: 0.15rem;
}

.stat-item {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 0.5rem 0;
}
</style>

