<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import { useInvoiceStore } from '@/stores/invoice'
import { useQuoteStore } from '@/stores/quote'
import { useEmployeeStore } from '@/stores/employee'
import { useCariStore } from '@/stores/cari'
import { useToast } from 'primevue/usetoast'
import { formatDate } from '@/utils/date'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'

const projectStore = useProjectStore()
const invoiceStore = useInvoiceStore()
const quoteStore = useQuoteStore()
const employeeStore = useEmployeeStore()
const cariStore = useCariStore()
const router = useRouter()
const route = useRoute()
const toast = useToast()

const loading = ref(false)
const project = computed(() => projectStore.activeProject)
const activeTab = ref<'info' | 'invoices' | 'quotes' | 'employees'>('info')

const showAddInvoiceDialog = ref(false)
const showAddQuoteDialog = ref(false)
const showAddEmployeeDialog = ref(false)
const selectedInvoiceId = ref('')
const selectedQuoteId = ref('')
const selectedEmployeeId = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const id = route.params.id as string
    await projectStore.fetchProjectByID(id)
    // Hata olursa sessizce geç (modül kapalı olabilir)
    await Promise.allSettled([
      invoiceStore.fetchInvoices({ page: 1, limit: 1000 }),
      quoteStore.fetchQuotes({ page: 1, limit: 1000 }),
      cariStore.fetchCaris({ page: 1, limit: 1000 }),
      employeeStore.fetchEmployees({ page: 1, limit: 1000 }),
    ])
  } finally {
    loading.value = false
  }
})

const statusMap: Record<string, { label: string; severity: string }> = {
  planning:    { label: 'Planlama',       severity: 'secondary' },
  in_progress: { label: 'Devam Ediyor',   severity: 'info' },
  on_hold:     { label: 'Durduruldu',     severity: 'warn' },
  completed:   { label: 'Tamamlandı',     severity: 'success' },
  cancelled:   { label: 'İptal',          severity: 'danger' },
}
const statusLabel    = (s: string) => statusMap[s]?.label    ?? s
const statusSeverity = (s: string) => statusMap[s]?.severity ?? 'secondary'

const getCariName = (cariId: string) => {
  const c = cariStore.caris.find(x => x.id === cariId)
  return c ? c.name : '-'
}

const availableInvoices = computed(() => {
  if (!project.value) return []
  const linked = new Set((project.value.invoices || []).map((i: any) => i.id))
  return invoiceStore.invoices.filter(i => !linked.has(i.id))
})

const availableQuotes = computed(() => {
  if (!project.value) return []
  const linked = new Set((project.value.quotes || []).map((q: any) => q.id))
  return quoteStore.quotes.filter(q => !linked.has(q.id))
})

const availableEmployees = computed(() => {
  if (!project.value) return []
  const linked = new Set((project.value.employees || []).map((e: any) => e.id))
  return employeeStore.employees.filter(e => !linked.has(e.id))
})

const reload = () => projectStore.fetchProjectByID(project.value!.id)

const addInvoice = async () => {
  if (!selectedInvoiceId.value) return
  try {
    await projectStore.addInvoiceToProject(project.value!.id, selectedInvoiceId.value)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Fatura eklendi', life: 3000 })
    selectedInvoiceId.value = ''
    showAddInvoiceDialog.value = false
    await reload()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'İşlem başarısız', life: 10000 })
  }
}

const removeInvoice = async (invoiceId: string) => {
  if (!confirm('Faturayı projeden kaldır?')) return
  try {
    await projectStore.removeInvoiceFromProject(project.value!.id, invoiceId)
    toast.add({ severity: 'success', summary: 'Kaldırıldı', detail: 'Fatura kaldırıldı', life: 3000 })
    await reload()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Hata', life: 10000 })
  }
}

const addQuote = async () => {
  if (!selectedQuoteId.value) return
  try {
    await projectStore.addQuoteToProject(project.value!.id, selectedQuoteId.value)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif eklendi', life: 3000 })
    selectedQuoteId.value = ''
    showAddQuoteDialog.value = false
    await reload()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Hata', life: 10000 })
  }
}

const removeQuote = async (quoteId: string) => {
  if (!confirm('Teklifi projeden kaldır?')) return
  try {
    await projectStore.removeQuoteFromProject(project.value!.id, quoteId)
    toast.add({ severity: 'success', summary: 'Kaldırıldı', detail: 'Teklif kaldırıldı', life: 3000 })
    await reload()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Hata', life: 10000 })
  }
}

const addEmployee = async () => {
  if (!selectedEmployeeId.value) return
  try {
    await projectStore.addEmployeeToProject(project.value!.id, selectedEmployeeId.value)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Personel eklendi', life: 3000 })
    selectedEmployeeId.value = ''
    showAddEmployeeDialog.value = false
    await reload()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Hata', life: 10000 })
  }
}

const removeEmployee = async (empId: string) => {
  if (!confirm('Personeli projeden kaldır?')) return
  try {
    await projectStore.removeEmployeeFromProject(project.value!.id, empId)
    toast.add({ severity: 'success', summary: 'Kaldırıldı', detail: 'Personel kaldırıldı', life: 3000 })
    await reload()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Hata', life: 10000 })
  }
}

const editProject = () => router.push(`/projects/${project.value?.id}/edit`)
const deleteProject = async () => {
  if (!confirm('Projeyi silmek istediğinize emin misiniz?')) return
  try {
    await projectStore.deleteProject(project.value!.id)
    toast.add({ severity: 'success', summary: 'Silindi', detail: 'Proje silindi', life: 3000 })
    router.push('/projects')
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Hata', life: 10000 })
  }
}
</script>

<template>
  <div v-if="loading" class="flex justify-center items-center py-16">
    <i class="pi pi-spin pi-spinner text-3xl text-cyan-500"></i>
  </div>

  <template v-else-if="project">
    <!-- ===== HEADER ===== -->
    <div class="prj-header">
      <div class="prj-header-left">
        <div class="flex items-center gap-2 mb-1">
          <span class="prj-code">{{ project.code }}</span>
          <Tag :value="statusLabel(project.status)" :severity="statusSeverity(project.status)" />
        </div>
        <h1 class="prj-title">{{ project.name }}</h1>
        <div class="prj-meta">
          <span>
            <i class="pi pi-user mr-1"></i>
            <span
              class="cursor-pointer text-cyan-600 hover:underline"
              @click="router.push(`/caris/${project.cari_id}`)"
            >{{ getCariName(project.cari_id) }}</span>
          </span>
          <span><i class="pi pi-calendar mr-1"></i>{{ formatDate(project.start_date) }} — {{ formatDate(project.end_date) }}</span>
          <span v-if="project.budget"><i class="pi pi-wallet mr-1"></i>{{ project.budget }}</span>
        </div>
      </div>
      <div class="flex gap-2">
        <Button icon="pi pi-pencil" size="small" severity="warn" outlined @click="editProject" />
        <Button icon="pi pi-trash" size="small" severity="danger" outlined @click="deleteProject" />
      </div>
    </div>

    <!-- ===== TABS ===== -->
    <div class="prj-tabs">
      <button
        v-for="t in [
          { key: 'info',      label: 'Bilgiler',  icon: 'pi-info-circle' },
          { key: 'invoices',  label: 'Faturalar', icon: 'pi-file-invoice', count: project.invoices?.length },
          { key: 'quotes',    label: 'Teklifler', icon: 'pi-file-edit',    count: project.quotes?.length },
          { key: 'employees', label: 'Personel',  icon: 'pi-users',        count: project.employees?.length },
        ]"
        :key="t.key"
        class="prj-tab"
        :class="{ 'prj-tab--active': activeTab === t.key }"
        @click="activeTab = (t.key as any)"
      >
        <i :class="['pi', t.icon]"></i>
        {{ t.label }}
        <span v-if="t.count" class="prj-tab-badge">{{ t.count }}</span>
      </button>
    </div>

    <!-- ===== TAB CONTENT ===== -->
    <div class="prj-panel">

      <!-- BİLGİLER -->
      <div v-if="activeTab === 'info'">
        <div v-if="project.description" class="prj-info-row">
          <div class="prj-info-label">Açıklama</div>
          <div class="prj-info-value">{{ project.description }}</div>
        </div>
        <div v-if="project.note" class="prj-info-row">
          <div class="prj-info-label">Notlar</div>
          <div class="prj-info-value">{{ project.note }}</div>
        </div>
        <div v-if="project.category" class="prj-info-row">
          <div class="prj-info-label">Kategori</div>
          <div class="prj-info-value">{{ project.category.name }}</div>
        </div>
        <div v-if="!project.description && !project.note && !project.category" class="text-sm text-slate-400 py-4 text-center">
          Ek bilgi girilmemiş.
        </div>
      </div>

      <!-- FATURALAR -->
      <div v-else-if="activeTab === 'invoices'">
        <div class="flex justify-end mb-3">
          <Button
            label="Fatura Ekle"
            icon="pi pi-plus"
            size="small"
            severity="success"
            outlined
            @click="showAddInvoiceDialog = true"
          />
        </div>
        <DataTable
          v-if="project.invoices?.length"
          :value="project.invoices"
          class="p-datatable-sm"
          responsiveLayout="scroll"
        >
          <Column field="number" header="Fatura No">
            <template #body="{ data }">
              <span class="cursor-pointer text-cyan-600 hover:underline" @click="router.push(`/invoices/${data.id}`)">
                {{ data.number }}
              </span>
            </template>
          </Column>
          <Column field="date" header="Tarih">
            <template #body="{ data }">{{ formatDate(data.date) }}</template>
          </Column>
          <Column header="Tutar">
            <template #body="{ data }">{{ data.total }} {{ data.currency }}</template>
          </Column>
          <Column field="status" header="Durum">
            <template #body="{ data }"><Tag :value="data.status" /></template>
          </Column>
          <Column style="width:60px; text-align:center">
            <template #body="{ data }">
              <Button icon="pi pi-times" text severity="danger" size="small" @click="removeInvoice(data.id)" />
            </template>
          </Column>
        </DataTable>
        <div v-else class="text-sm text-slate-400 text-center py-6">Bağlı fatura yok.</div>
      </div>

      <!-- TEKLİFLER -->
      <div v-else-if="activeTab === 'quotes'">
        <div class="flex justify-end mb-3">
          <Button
            label="Teklif Ekle"
            icon="pi pi-plus"
            size="small"
            severity="success"
            outlined
            @click="showAddQuoteDialog = true"
          />
        </div>
        <DataTable
          v-if="project.quotes?.length"
          :value="project.quotes"
          class="p-datatable-sm"
          responsiveLayout="scroll"
        >
          <Column field="number" header="Teklif No">
            <template #body="{ data }">
              <span class="cursor-pointer text-cyan-600 hover:underline" @click="router.push(`/quotes/${data.id}`)">
                {{ data.number }}
              </span>
            </template>
          </Column>
          <Column field="date" header="Tarih">
            <template #body="{ data }">{{ formatDate(data.date) }}</template>
          </Column>
          <Column header="Tutar">
            <template #body="{ data }">{{ data.total }} {{ data.currency }}</template>
          </Column>
          <Column field="status" header="Durum">
            <template #body="{ data }"><Tag :value="data.status" /></template>
          </Column>
          <Column style="width:60px; text-align:center">
            <template #body="{ data }">
              <Button icon="pi pi-times" text severity="danger" size="small" @click="removeQuote(data.id)" />
            </template>
          </Column>
        </DataTable>
        <div v-else class="text-sm text-slate-400 text-center py-6">Bağlı teklif yok.</div>
      </div>

      <!-- PERSONEL -->
      <div v-else-if="activeTab === 'employees'">
        <div class="flex justify-end mb-3">
          <Button
            label="Personel Ekle"
            icon="pi pi-plus"
            size="small"
            severity="success"
            outlined
            @click="showAddEmployeeDialog = true"
          />
        </div>
        <DataTable
          v-if="project.employees?.length"
          :value="project.employees"
          class="p-datatable-sm"
          responsiveLayout="scroll"
        >
          <Column field="name" header="Ad Soyad" />
          <Column field="title" header="Ünvan" />
          <Column field="department" header="Departman" />
          <Column style="width:60px; text-align:center">
            <template #body="{ data }">
              <Button icon="pi pi-times" text severity="danger" size="small" @click="removeEmployee(data.id)" />
            </template>
          </Column>
        </DataTable>
        <div v-else class="text-sm text-slate-400 text-center py-6">Atanmış personel yok.</div>
      </div>

    </div>
  </template>

  <div v-else class="text-center py-10 text-slate-400">Proje bulunamadı.</div>

  <!-- ===== DIALOGS ===== -->
  <Dialog v-model:visible="showAddInvoiceDialog" header="Fatura Ekle" modal :style="{ width: '420px' }">
    <div class="flex flex-col gap-4 pt-2">
      <div v-if="availableInvoices.length === 0" class="text-sm text-slate-400 text-center py-4">
        Eklenebilecek fatura bulunamadı.
      </div>
      <Select
        v-else
        v-model="selectedInvoiceId"
        :options="availableInvoices"
        optionLabel="number"
        optionValue="id"
        placeholder="Fatura seçiniz"
        class="w-full"
        filter
      />
      <div class="flex justify-end gap-2">
        <Button label="İptal" severity="secondary" outlined size="small" @click="showAddInvoiceDialog = false" />
        <Button v-if="availableInvoices.length > 0" label="Ekle" severity="success" size="small" @click="addInvoice" :disabled="!selectedInvoiceId" />
      </div>
    </div>
  </Dialog>

  <Dialog v-model:visible="showAddQuoteDialog" header="Teklif Ekle" modal :style="{ width: '420px' }">
    <div class="flex flex-col gap-4 pt-2">
      <div v-if="availableQuotes.length === 0" class="text-sm text-slate-400 text-center py-4">
        Eklenebilecek teklif bulunamadı.
      </div>
      <Select
        v-else
        v-model="selectedQuoteId"
        :options="availableQuotes"
        optionLabel="number"
        optionValue="id"
        placeholder="Teklif seçiniz"
        class="w-full"
        filter
      />
      <div class="flex justify-end gap-2">
        <Button label="İptal" severity="secondary" outlined size="small" @click="showAddQuoteDialog = false" />
        <Button v-if="availableQuotes.length > 0" label="Ekle" severity="success" size="small" @click="addQuote" :disabled="!selectedQuoteId" />
      </div>
    </div>
  </Dialog>

  <Dialog v-model:visible="showAddEmployeeDialog" header="Personel Ekle" modal :style="{ width: '420px' }">
    <div class="flex flex-col gap-4 pt-2">
      <Select
        v-model="selectedEmployeeId"
        :options="availableEmployees"
        optionLabel="name"
        optionValue="id"
        placeholder="Personel seçiniz"
        class="w-full"
      />
      <div class="flex justify-end gap-2">
        <Button label="İptal" severity="secondary" outlined size="small" @click="showAddEmployeeDialog = false" />
        <Button label="Ekle" severity="success" size="small" @click="addEmployee" :disabled="!selectedEmployeeId" />
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
/* ===== HEADER ===== */
.prj-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 0.625rem;
  padding: 1.1rem 1.25rem;
  margin-bottom: 0;
}
:root.p-dark .prj-header { background: #1e293b; border-color: rgba(255,255,255,0.07); }

.prj-header-left { flex: 1; min-width: 0; }
.prj-code {
  font-size: 0.7rem; font-weight: 700; letter-spacing: 0.06em;
  color: #06b6d4; background: rgba(6,182,212,0.08);
  padding: 0.1rem 0.45rem; border-radius: 0.3rem;
}
.prj-title { font-size: 1.15rem; font-weight: 700; color: #1a202c; margin: 0.2rem 0 0.4rem; }
:root.p-dark .prj-title { color: #f1f5f9; }
.prj-meta { display: flex; flex-wrap: wrap; gap: 1rem; font-size: 0.78rem; color: #718096; }
:root.p-dark .prj-meta { color: #94a3b8; }
.prj-meta .pi { font-size: 0.7rem; opacity: 0.7; }

/* ===== TABS ===== */
.prj-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid #e8eaf0;
  background: #fff;
  border-radius: 0.625rem 0.625rem 0 0;
  overflow: hidden;
  margin-top: 0.75rem;
}
:root.p-dark .prj-tabs { background: #1e293b; border-bottom-color: rgba(255,255,255,0.07); }

.prj-tab {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.65rem 1rem;
  font-size: 0.8rem;
  font-weight: 500;
  color: #718096;
  border: none;
  background: transparent;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: color 0.12s, border-color 0.12s;
  white-space: nowrap;
}
:root.p-dark .prj-tab { color: #94a3b8; }
.prj-tab:hover { color: #06b6d4; }
.prj-tab--active {
  color: #06b6d4 !important;
  border-bottom-color: #06b6d4;
  font-weight: 700;
}
.prj-tab .pi { font-size: 0.72rem; }
.prj-tab-badge {
  font-size: 0.65rem; font-weight: 700;
  background: #06b6d4; color: #fff;
  padding: 0.05rem 0.35rem; border-radius: 9999px;
  min-width: 1.1rem; text-align: center;
}

/* ===== PANEL ===== */
.prj-panel {
  background: #fff;
  border: 1px solid #e8eaf0;
  border-top: none;
  border-radius: 0 0 0.625rem 0.625rem;
  padding: 1rem 1.25rem;
  min-height: 180px;
}
:root.p-dark .prj-panel { background: #1e293b; border-color: rgba(255,255,255,0.07); }

.prj-info-row {
  display: flex;
  gap: 1rem;
  padding: 0.6rem 0;
  border-bottom: 1px solid #f0f2f5;
  font-size: 0.83rem;
}
:root.p-dark .prj-info-row { border-bottom-color: rgba(255,255,255,0.05); }
.prj-info-row:last-child { border-bottom: none; }
.prj-info-label { width: 100px; flex-shrink: 0; font-weight: 600; color: #4a5568; font-size: 0.75rem; }
:root.p-dark .prj-info-label { color: #94a3b8; }
.prj-info-value { color: #1a202c; }
:root.p-dark .prj-info-value { color: #e2e8f0; }
</style>
