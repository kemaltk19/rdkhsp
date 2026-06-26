<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import { useCariStore } from '@/stores/cari'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { formatDate, formatDateTime } from '@/utils/date'

const projectStore = useProjectStore()
const cariStore = useCariStore()
const router = useRouter()
const toast = useToast()

const searchQuery = ref('')
const selectedStatus = ref('')
const first = ref(0)
const rows = ref(20)
const page = ref(1)
const dt = ref()
const sortField = ref('')
const sortOrder = ref(1)

const totalRecords = computed(() => projectStore.total)

const statusOptions = ref([
  { label: 'Tüm Durumlar', value: '' },
  { label: 'Planlama', value: 'planning' },
  { label: 'Devam Ediyor', value: 'in_progress' },
  { label: 'Duruturuluyor', value: 'on_hold' },
  { label: 'Tamamlandı', value: 'completed' },
  { label: 'İptal Edildi', value: 'cancelled' },
])

const loadData = async () => {
  const params: any = {
    page: page.value,
    limit: rows.value,
    q: searchQuery.value,
    status: selectedStatus.value,
    sort: sortField.value ? `${sortField.value} ${sortOrder.value === 1 ? 'asc' : 'desc'}` : ''
  }
  await projectStore.fetchProjects(params)
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

watch([searchQuery, selectedStatus], () => {
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
  router.push('/projects/new')
}

const editProject = (id: string) => {
  router.push(`/projects/${id}/edit`)
}

const getCari = (cariId: string) => {
  return cariStore.caris.find(c => c.id === cariId)
}

const viewDetail = (id: string) => {
  router.push(`/projects/${id}`)
}

const deleteProject = async (id: string) => {
  if (confirm('Proje silmek istediğinize emin misiniz?')) {
    try {
      await projectStore.deleteProject(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Proje silindi', life: 10000 })
      loadData()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Proje silinemedi', life: 10000 })
      }
    }
  }
}

const getCariName = (cariId: string) => {
  const cari = cariStore.caris.find(c => c.id === cariId)
  return cari ? cari.name : 'Yükleniyor...'
}

const getStatusLabel = (status: string) => {
  switch (status) {
    case 'planning': return 'Planlama'
    case 'in_progress': return 'Devam Ediyor'
    case 'on_hold': return 'Duruturuluyor'
    case 'completed': return 'Tamamlandı'
    case 'cancelled': return 'İptal Edildi'
    default: return status
  }
}

const getStatusSeverity = (status: string) => {
  switch (status) {
    case 'planning': return 'secondary'
    case 'in_progress': return 'info'
    case 'on_hold': return 'warn'
    case 'completed': return 'success'
    case 'cancelled': return 'danger'
    default: return 'secondary'
  }
}
</script>

<template>
  <div class="project-list-container">
    <Card class="table-card">
      <template #content>
        <div class="filters-header flex flex-col md:flex-row justify-between items-center gap-4 mb-4">
          <div class="select-filters flex gap-2 w-full md:w-auto">
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
              <InputText v-model="searchQuery" placeholder="Proje adı veya kodu ile ara..." class="w-full pl-10" />
            </div>
          </div>

          <div class="flex gap-2 w-full md:w-auto justify-end shrink-0">
            <Button label="Yeni Proje" icon="pi pi-plus" @click="openNew" severity="success" />
          </div>
        </div>

        <DataTable
          ref="dt"
          :value="projectStore.projects"
          lazy
          paginator
          :first="first"
          :rows="rows"
          :rowsPerPageOptions="[20, 50, 100]"
          :totalRecords="projectStore.total"
          :loading="projectStore.loading"
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
          <Column field="code" header="Proje Kodu" sortable style="min-width: 120px; width: 15%">
            <template #body="{ data }">
              <span 
                class="font-medium text-slate-700 dark:text-slate-200 cursor-pointer hover:text-cyan-600 transition-colors"
                @click="viewDetail(data.id)"
              >
                {{ data.code }}
              </span>
            </template>
          </Column>
          <Column field="name" header="Proje Adı" sortable style="min-width: 200px; width: 25%">
            <template #body="{ data }">
              <div 
                class="name-cell cursor-pointer hover:underline"
                @click="viewDetail(data.id)"
              >
                <div class="font-medium text-slate-700 dark:text-slate-200 truncate max-w-[250px]" :title="data.name">
                  {{ data.name }}
                </div>
              </div>
            </template>
          </Column>
          <Column field="cari_id" header="Müşteri" style="min-width: 150px; width: 20%">
            <template #body="{ data }">
              <div 
                v-if="getCari(data.cari_id)" 
                class="name-cell cursor-pointer hover:underline"
                @click="router.push(`/caris/${data.cari_id}`)"
              >
                <div class="font-medium text-slate-700 dark:text-slate-200 truncate max-w-[200px]" :title="getCari(data.cari_id)?.name">
                  {{ getCari(data.cari_id)?.name }}
                </div>
              </div>
              <span v-else class="text-slate-400">Yükleniyor...</span>
            </template>
          </Column>
          <Column field="start_date" header="Başlama" sortable style="min-width: 120px; width: 12%">
            <template #body="{ data }">
              {{ formatDate(data.start_date) }}
            </template>
          </Column>
          <Column field="end_date" header="Bitiş" sortable style="min-width: 120px; width: 12%">
            <template #body="{ data }">
              {{ formatDate(data.end_date) }}
            </template>
          </Column>
          <Column field="status" header="Durum" sortable style="min-width: 120px; width: 12%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Tag :value="getStatusLabel(data.status)" :severity="getStatusSeverity(data.status)" />
            </template>
          </Column>
          <Column header="Aksiyonlar" style="min-width: 100px; width: 8%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <div class="actions-cell">
                <Button icon="pi pi-eye" class="rounded-md p-button-text p-button-info" @click="viewDetail(data.id)" title="Detay" />
                <Button icon="pi pi-pencil" class="rounded-md p-button-text" @click="editProject(data.id)" title="Düzenle" severity="warn" />
                <Button icon="pi pi-trash" class="rounded-md p-button-text" @click="deleteProject(data.id)" title="Sil" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.project-list-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
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

.select-filters {
  display: flex;
  gap: 0.5rem;
}

.status-select {
  width: 160px;
}

.actions-cell {
  display: flex;
  justify-content: center;
  gap: 0.25rem;
}
</style>
