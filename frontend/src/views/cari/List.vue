<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCariStore } from '@/stores/cari'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Menu from 'primevue/menu'
import Avatar from 'primevue/avatar'
import Money from '@/components/Money.vue'
import FormModal from './FormModal.vue'
import { exportToPDF } from '@/utils/pdfExport'

import { usePermission } from '@/composables/usePermission'

const cariStore = useCariStore()
const settingsStore = useSettingsStore()
const router = useRouter()
const toast = useToast()
const { can } = usePermission()

const showModal = ref(false)
const selectedCariId = ref<string | undefined>(undefined)
const searchQuery = ref('')
const selectedType = ref('')
const selectedGroup = ref('')
const groupOptions = ref<{ label: string, value: string }[]>([])
const first = ref(0)
const rows = ref(20)
const page = ref(1)
const dt = ref()
const selectedItems = ref([])
const sortField = ref('')
const sortOrder = ref(1)

const typeOptions = ref([
  { label: 'Tümü', value: '' },
  { label: 'Müşteri', value: 'customer' },
  { label: 'Tedarikçi', value: 'supplier' },
  { label: 'Her İkisi', value: 'both' },
])

const loadData = async () => {
  const sortParam = sortField.value ? `${sortField.value} ${sortOrder.value === 1 ? 'asc' : 'desc'}` : ''
  await cariStore.fetchCaris({ page: page.value, limit: rows.value, q: searchQuery.value, type: selectedType.value, group: selectedGroup.value, sort: sortParam })
}

const onSort = (event: any) => {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  loadData()
}

onMounted(async () => {
  await loadData()
  try {
    const gVal = await settingsStore.fetchSetting('cari_groups')
    if (gVal) {
      const parsed = JSON.parse(gVal) as string[]
      groupOptions.value = [
        { label: 'Tüm Gruplar', value: '' },
        ...parsed.map((g: string) => ({ label: g, value: g }))
      ]
    } else {
      groupOptions.value = [
        { label: 'Tüm Gruplar', value: '' },
        ...['Bireysel', 'Kurumsal', 'Kurum', 'Fabrika', 'Esnaf', 'Şirket', 'Diğer'].map((g: string) => ({ label: g, value: g }))
      ]
    }
  } catch {
    groupOptions.value = [
      { label: 'Tüm Gruplar', value: '' }
    ]
  }
})

let searchTimeout: any = null
watch(searchQuery, (newVal) => {
  // If user typed 1 character, do not trigger search yet
  if (newVal.length === 1) return

  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    page.value = 1
    first.value = 0
    loadData()
  }, 300)
})

watch(selectedType, () => {
  page.value = 1
  first.value = 0
  loadData()
})

watch(selectedGroup, () => {
  page.value = 1
  first.value = 0
  loadData()
})

const activeCount = computed(() => cariStore.caris.filter(c => c.is_active).length)
const passiveCount = computed(() => cariStore.caris.filter(c => !c.is_active).length)

const onPage = (event: any) => {
  page.value = event.page + 1
  rows.value = event.rows
  first.value = event.first
  loadData()
}

const openNew = () => {
  selectedCariId.value = undefined
  showModal.value = true
}
const editCari = (id: string) => {
  selectedCariId.value = id
  showModal.value = true
}
const viewDetail = (id: string) => router.push(`/caris/${id}`)

const deleteCariItem = async (id: string) => {
  if (confirm('Bu cari kartı silmek istediğinize emin misiniz?')) {
    try {
      await cariStore.deleteCari(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Cari kart silindi', life: 10000 })
      loadData()
    } catch (err: any) {
      const msg = err.response?.data?.error?.message || 'İşlem gerçekleştirilemedi'
      toast.add({ severity: 'error', summary: 'Hata', detail: msg, life: 10000 })
    }
  }
}

const getCariTypeLabel = (type: string) => ({ customer: 'Müşteri', supplier: 'Tedarikçi', both: 'Her İkisi' } as any)[type] || type
const getCariTypeSeverity = (type: string) => ({ customer: 'info', supplier: 'warn', both: 'success' } as any)[type] || 'secondary'

const getBalanceColor = (balance: string) => {
  const bal = parseFloat(balance)
  if (bal > 0) return 'pos'
  if (bal < 0) return 'neg'
  return 'zero'
}

const getAvatarColor = (name: string) => {
  const colors = ['#0d9488', '#0ea5e9', '#6366f1', '#0891b2', '#14b8a6', '#64748b', '#f59e0b', '#10b981']
  if (!name) return colors[0]
  return colors[name.charCodeAt(0) % colors.length]
}

const exportMenu = ref()
const exportOptions = [
  { label: 'PDF', icon: 'pi pi-file-pdf', command: () => exportData('pdf') },
  { label: 'Excel', icon: 'pi pi-file-excel', command: () => exportData('excel') },
]
const toggleExportMenu = (event: any) => exportMenu.value.toggle(event)

const exportData = (format: string) => {
  if (format === 'excel') {
    dt.value.exportCSV({ selectionOnly: selectedItems.value.length > 0 })
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Excel olarak dışa aktarıldı', life: 10000 })
  } else if (format === 'pdf') {
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : cariStore.caris
    const columns = [
      { header: 'Kod', dataKey: 'code' },
      { header: 'Cari Adı', dataKey: 'name' },
      { header: 'Tip', dataKey: 'type' },
      { header: 'Telefon', dataKey: 'phone' },
      { header: 'Bakiye', dataKey: 'balance' },
    ]
    exportToPDF('Cari_Listesi', columns, dataToExport)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'PDF olarak dışa aktarıldı', life: 10000 })
  }
}
</script>

<template>
  <div class="cari-list-container">
    <!-- Table Card -->
    <Card class="table-card">
      <template #content>
        <div class="filters-header">
          <div class="f-left flex gap-2">
            <Select v-model="selectedType" :options="typeOptions" optionLabel="label" optionValue="value" placeholder="Cari Tipi" class="type-select" />
            <Select v-model="selectedGroup" :options="groupOptions" optionLabel="label" optionValue="value" placeholder="Cari Grubu" class="group-select" showClear />
          </div>
          <div class="f-mid">
            <span class="search-input">
              <i class="pi pi-search"></i>
              <InputText v-model="searchQuery" placeholder="İsim, unvan veya kod ile ara..." />
            </span>
          </div>
          <div class="f-stats">
            Cari : {{ cariStore.total }} / aktif: {{ activeCount }} / Pasif : {{ passiveCount }}
          </div>
          <div class="f-right">
            <Button v-if="can('caris', 'create')" label="Ekle" icon="pi pi-plus" @click="openNew" severity="success" />
            <Button label="Aktar" icon="pi pi-upload" outlined @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
            <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
          </div>
        </div>

        <DataTable
          ref="dt"
          :value="cariStore.caris"
          v-model:selection="selectedItems"
          lazy paginator
          :first="first"
          :rows="rows"
          :rowsPerPageOptions="[20, 50, 100]"
          :totalRecords="cariStore.total"
          :loading="cariStore.loading"
          @page="onPage"
          @sort="onSort"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          dataKey="id"
          :rowClass="() => 'cursor-pointer'"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ (page - 1) * rows + slotProps.index + 1 }}
            </template>
          </Column>
          <Column selectionMode="multiple" headerStyle="width: 3rem"></Column>
          <Column field="code" header="Kod" sortable style="min-width: 100px; width: 10%">
            <template #body="{ data }"><span class="font-medium text-slate-700 dark:text-slate-200 cursor-pointer hover:text-cyan-600 transition-colors" @click="viewDetail(data.id)">{{ data.code }}</span></template>
          </Column>
          <Column header="Logo" style="width: 5%">
            <template #body="{ data }">
              <Avatar :label="data.name ? data.name.charAt(0).toUpperCase() : '?'" shape="circle" :style="{ backgroundColor: getAvatarColor(data.name), color: '#ffffff' }" @click="viewDetail(data.id)" class="cursor-pointer" />
            </template>
          </Column>
          <Column field="name" header="Cari" sortable style="min-width: 200px; width: 20%">
            <template #body="{ data }">
              <div class="name-cell cursor-pointer" @click="viewDetail(data.id)">
                <div class="name-text">
                  <div class="font-medium text-slate-700 dark:text-slate-200 truncate max-w-[250px]" :title="data.name">{{ data.name }}</div>
                  <div class="text-xs text-slate-500 dark:text-slate-400 mt-0.5 truncate max-w-[250px]" v-if="data.contact_name" :title="data.contact_name">{{ data.contact_name }}</div>
                </div>
              </div>
            </template>
          </Column>
          <Column field="type" header="Tip" sortable style="min-width: 100px; width: 10%">
            <template #body="{ data }"><Tag :value="getCariTypeLabel(data.type)" :severity="getCariTypeSeverity(data.type)" /></template>
          </Column>
          <Column field="group" header="Grup" sortable style="min-width: 100px; width: 10%">
            <template #body="{ data }">
              <span v-if="data.group" class="text-sm">{{ data.group }}</span>
              <span v-else class="text-xs text-gray-400">-</span>
            </template>
          </Column>
          <Column field="phone" header="İletişim" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              <div class="flex flex-col gap-1 text-sm mt-1 mb-1">
                <span v-if="data.phone">{{ data.phone }}</span>
                <span v-if="data.landline">{{ data.landline }}</span>
                <span v-if="!data.phone && !data.landline" class="text-xs text-gray-400">-</span>
              </div>
            </template>
          </Column>
          <Column header="Bakiye" style="min-width: 160px; width: 15%" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <div v-if="data.balances && data.balances.length > 0" class="bal-cell">
                <span v-for="b in data.balances" :key="b.currency" :class="getBalanceColor(b.balance)">
                  <Money :value="b.balance" :currency="b.currency" />
                </span>
              </div>
              <div v-else class="bal-cell zero">-</div>
            </template>
          </Column>
          <Column field="is_active" header="Durum" sortable style="min-width: 100px; width: 8%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Tag :value="data.is_active ? 'Aktif' : 'Pasif'" :severity="data.is_active ? 'success' : 'danger'" />
            </template>
          </Column>
          <Column header="Aksiyonlar" style="min-width: 120px; width: 7%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <div class="actions-cell">
                <Button icon="pi pi-eye" text rounded severity="info" @click="viewDetail(data.id)" v-tooltip.top="'Detay'" />
                <Button v-if="can('caris', 'update')" icon="pi pi-pencil" text rounded @click="editCari(data.id)" v-tooltip.top="'Düzenle'" severity="warn" />
                <Button v-if="can('caris', 'delete')" icon="pi pi-trash" text rounded @click="deleteCariItem(data.id)" v-tooltip.top="'Sil'" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <FormModal v-if="showModal" v-model:visible="showModal" :cari-id="selectedCariId" @saved="loadData" />
  </div>
</template>

<style scoped>
.cari-list-container { display: flex; flex-direction: column; gap: 1.5rem; }

.table-card { border-radius: 12px; }

.filters-header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1.25rem; flex-wrap: wrap; }
.f-left { flex-shrink: 0; }
.f-mid { flex: 1; display: flex; justify-content: center; min-width: 200px; }
.f-stats { font-size: 0.95rem; font-weight: 600; color: #64748b; white-space: nowrap; padding: 0 1rem; }
.f-right { display: flex; gap: 0.5rem; flex-shrink: 0; }
.type-select, .group-select { width: 180px; }
.search-input { position: relative; width: 100%; max-width: 420px; display: block; }
.search-input i { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #94a3b8; }
.search-input :deep(input) { width: 100%; padding-left: 2.4rem; }

.name-cell { display: flex; align-items: center; gap: 10px; }

.bal-cell { display: flex; flex-direction: column; align-items: flex-end; gap: 1px; text-align: right; font-variant-numeric: tabular-nums; }
.bal-cell .pos { color: #16a34a; font-weight: 600; }
.bal-cell .neg { color: #dc2626; font-weight: 600; }
.bal-cell .zero { color: #94a3b8; }

.actions-cell { display: flex; justify-content: center; gap: 0.15rem; }

:root.p-dark .metric-label { color: #94a3b8; }
</style>
