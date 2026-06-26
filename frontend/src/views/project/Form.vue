<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import { useCariStore } from '@/stores/cari'
import { useProjectCategoryStore } from '@/stores/project_category'
import { useEmployeeStore } from '@/stores/employee'
import { useInvoiceStore } from '@/stores/invoice'
import { useQuoteStore } from '@/stores/quote'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import MultiSelect from 'primevue/multiselect'
import Textarea from 'primevue/textarea'
import Calendar from 'primevue/calendar'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'

const projectStore = useProjectStore()
const cariStore = useCariStore()
const categoryStore = useProjectCategoryStore()
const employeeStore = useEmployeeStore()
const invoiceStore = useInvoiceStore()
const quoteStore = useQuoteStore()
const router = useRouter()
const route = useRoute()
const toast = useToast()

const loading = ref(false)
const submitting = ref(false)
const errorMsg = ref('')
const isEdit = computed(() => route.params.id && route.params.id !== 'new')

const form = ref({
  cari_id: '' as string,
  name: '',
  description: '',
  code: '',
  category_id: null as string | null,
  employee_ids: [] as string[],
  invoice_ids: [] as string[],
  quote_ids: [] as string[],
  status: 'planning',
  start_date: null as any,
  end_date: null as any,
  budget: null as number | null,
  note: '',
})

const statusOptions = [
  { label: 'Planlama', value: 'planning' },
  { label: 'Devam Ediyor', value: 'in_progress' },
  { label: 'Duraklatılıyor', value: 'on_hold' },
  { label: 'Tamamlandı', value: 'completed' },
  { label: 'İptal Edildi', value: 'cancelled' },
]

// Müşteri seçenekleri: "AD (KOD)" etiketiyle, değer = id
const cariOptions = computed(() =>
  cariStore.caris.map((c: any) => ({
    label: c.code ? `${c.name} (${c.code})` : c.name,
    value: c.id,
  }))
)

// Seçili müşteriye ait faturalar (cari_id eşleşmesi)
const cariInvoices = computed(() => {
  if (!form.value.cari_id) return []
  return invoiceStore.invoices
    .filter((i: any) => i.cari_id === form.value.cari_id)
    .map((i: any) => ({
      label: `${i.number} — ${i.total} ${i.currency}`,
      value: i.id,
    }))
})

// Seçili müşteriye ait teklifler
const cariQuotes = computed(() => {
  if (!form.value.cari_id) return []
  return quoteStore.quotes
    .filter((q: any) => q.cari_id === form.value.cari_id)
    .map((q: any) => ({
      label: `${q.number} — ${q.total} ${q.currency}`,
      value: q.id,
    }))
})

// Müşteri değişince seçili fatura/teklifleri temizle (artık o cariye ait değiller)
const onCariChange = () => {
  form.value.invoice_ids = []
  form.value.quote_ids = []
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([
      cariStore.fetchCaris({ page: 1, limit: 1000 }),
      categoryStore.fetchCategories({ page: 1, limit: 1000 }),
      employeeStore.fetchEmployees({ page: 1, limit: 1000 }),
      invoiceStore.fetchInvoices({ page: 1, limit: 1000 }),
      quoteStore.fetchQuotes({ page: 1, limit: 1000 }),
    ])

    if (isEdit.value) {
      const id = route.params.id as string
      await projectStore.fetchProjectByID(id)
      if (projectStore.activeProject) {
        const prj = projectStore.activeProject
        form.value = {
          cari_id: prj.cari_id,
          name: prj.name,
          description: prj.description || '',
          code: prj.code,
          category_id: prj.category_id || null,
          employee_ids: prj.employees?.map((e: any) => e.id) || [],
          invoice_ids: prj.invoices?.map((i: any) => i.id) || [],
          quote_ids: prj.quotes?.map((q: any) => q.id) || [],
          status: prj.status,
          start_date: new Date(prj.start_date),
          end_date: new Date(prj.end_date),
          budget: prj.budget || null,
          note: prj.note || '',
        }
      }
    } else {
      const today = new Date()
      const future = new Date()
      future.setDate(future.getDate() + 30)
      form.value.start_date = today
      form.value.end_date = future
    }
  } finally {
    loading.value = false
  }
})

const submit = async () => {
  if (!form.value.cari_id) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Müşteri seçiniz', life: 10000 })
    return
  }
  if (!form.value.name) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Proje adı gereklidir', life: 10000 })
    return
  }

  submitting.value = true
  try {
    const startDate = form.value.start_date instanceof Date ? form.value.start_date.toISOString() : form.value.start_date
    const endDate = form.value.end_date instanceof Date ? form.value.end_date.toISOString() : form.value.end_date

    const payload = {
      cari_id: form.value.cari_id,
      name: form.value.name,
      description: form.value.description,
      code: form.value.code,
      category_id: form.value.category_id,
      status: form.value.status,
      start_date: startDate,
      end_date: endDate,
      budget: form.value.budget,
      note: form.value.note,
    }

    let projectId: string
    if (isEdit.value) {
      const id = route.params.id as string
      await projectStore.updateProject(id, payload)
      projectId = id

      const current = projectStore.activeProject

      // Personel senkronizasyonu
      const curEmp = current?.employees?.map((e: any) => e.id) || []
      for (const eid of curEmp) {
        if (!form.value.employee_ids.includes(eid)) await projectStore.removeEmployeeFromProject(id, eid)
      }
      for (const eid of form.value.employee_ids) {
        if (!curEmp.includes(eid)) await projectStore.addEmployeeToProject(id, eid)
      }

      // Fatura senkronizasyonu
      const curInv = current?.invoices?.map((i: any) => i.id) || []
      for (const iid of curInv) {
        if (!form.value.invoice_ids.includes(iid)) await projectStore.removeInvoiceFromProject(id, iid)
      }
      for (const iid of form.value.invoice_ids) {
        if (!curInv.includes(iid)) await projectStore.addInvoiceToProject(id, iid)
      }

      // Teklif senkronizasyonu
      const curQuo = current?.quotes?.map((q: any) => q.id) || []
      for (const qid of curQuo) {
        if (!form.value.quote_ids.includes(qid)) await projectStore.removeQuoteFromProject(id, qid)
      }
      for (const qid of form.value.quote_ids) {
        if (!curQuo.includes(qid)) await projectStore.addQuoteToProject(id, qid)
      }

      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Proje güncellendi', life: 10000 })
    } else {
      const created = await projectStore.createProject(payload)
      projectId = created.id

      for (const eid of form.value.employee_ids) {
        await projectStore.addEmployeeToProject(projectId, eid)
      }
      for (const iid of form.value.invoice_ids) {
        await projectStore.addInvoiceToProject(projectId, iid)
      }
      for (const qid of form.value.quote_ids) {
        await projectStore.addQuoteToProject(projectId, qid)
      }

      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Proje oluşturuldu', life: 10000 })
    }

    setTimeout(() => {
      router.push('/projects')
    }, 1200)
  } catch (err: any) {
    const detail = err.response?.data?.error?.message || 'İşlem başarısız oldu'
    toast.add({ severity: 'error', summary: 'Hata', detail, life: 10000 })
  } finally {
    submitting.value = false
  }
}

const cancel = () => {
  router.push('/projects')
}
</script>

<template>
  <div class="project-form-container">
    <Card class="form-card">
      <template #header>
        <div class="card-header">
          <h2 class="page-title">{{ isEdit ? 'Proje Düzenle' : 'Yeni Proje' }}</h2>
        </div>
      </template>
      <template #content>
        <div v-if="errorMsg" class="mb-4">
          <Message severity="error" :text="errorMsg" />
        </div>

        <div v-if="loading" class="text-center py-10">
          <i class="pi pi-spin pi-spinner" style="font-size: 2rem"></i>
          <p class="mt-3">Yükleniyor...</p>
        </div>

        <form v-else @submit.prevent="submit" class="space-y-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="form-group">
              <label for="cari" class="block font-medium mb-2">Müşteri *</label>
              <Select
                id="cari"
                v-model="form.cari_id"
                :options="cariOptions"
                optionLabel="label"
                optionValue="value"
                placeholder="Müşteri seçiniz"
                :disabled="isEdit"
                filter
                class="w-full"
                @change="onCariChange"
              />
            </div>

            <div class="form-group">
              <label for="status" class="block font-medium mb-2">Durum</label>
              <Select
                id="status"
                v-model="form.status"
                :options="statusOptions"
                optionLabel="label"
                optionValue="value"
                placeholder="Durum seçiniz"
                class="w-full"
              />
            </div>

            <div class="form-group md:col-span-2">
              <label for="name" class="block font-medium mb-2">Proje Adı *</label>
              <InputText
                id="name"
                v-model="form.name"
                placeholder="Proje adı"
                class="w-full"
              />
            </div>

            <div class="form-group">
              <label for="category" class="block font-medium mb-2">Proje Kategorisi</label>
              <Select
                id="category"
                v-model="form.category_id"
                :options="categoryStore.categories"
                optionLabel="name"
                optionValue="id"
                placeholder="Kategori seçiniz"
                showClear
                class="w-full"
              />
            </div>

            <div class="form-group">
              <label for="budget" class="block font-medium mb-2">Bütçe</label>
              <InputNumber
                id="budget"
                v-model="form.budget"
                placeholder="Bütçe"
                :useGrouping="true"
                class="w-full"
              />
            </div>

            <div class="form-group md:col-span-2">
              <label for="employees" class="block font-medium mb-2">Proje Personeli</label>
              <MultiSelect
                id="employees"
                v-model="form.employee_ids"
                :options="employeeStore.employees"
                optionLabel="name"
                optionValue="id"
                placeholder="Personel seçiniz"
                display="chip"
                class="w-full"
              />
            </div>

            <!-- Faturalar — seçili müşteriye ait -->
            <div class="form-group">
              <label for="invoices" class="block font-medium mb-2">İlişkili Faturalar</label>
              <MultiSelect
                id="invoices"
                v-model="form.invoice_ids"
                :options="cariInvoices"
                optionLabel="label"
                optionValue="value"
                :placeholder="form.cari_id ? 'Fatura seçiniz' : 'Önce müşteri seçin'"
                :disabled="!form.cari_id"
                display="chip"
                filter
                class="w-full"
              />
              <small v-if="form.cari_id && cariInvoices.length === 0" class="text-slate-400 mt-1">
                Bu müşteriye ait fatura yok.
              </small>
            </div>

            <!-- Teklifler — seçili müşteriye ait -->
            <div class="form-group">
              <label for="quotes" class="block font-medium mb-2">İlişkili Teklifler</label>
              <MultiSelect
                id="quotes"
                v-model="form.quote_ids"
                :options="cariQuotes"
                optionLabel="label"
                optionValue="value"
                :placeholder="form.cari_id ? 'Teklif seçiniz' : 'Önce müşteri seçin'"
                :disabled="!form.cari_id"
                display="chip"
                filter
                class="w-full"
              />
              <small v-if="form.cari_id && cariQuotes.length === 0" class="text-slate-400 mt-1">
                Bu müşteriye ait teklif yok.
              </small>
            </div>

            <div class="form-group">
              <label for="startDate" class="block font-medium mb-2">Başlama Tarihi</label>
              <Calendar
                id="startDate"
                v-model="form.start_date"
                date-format="dd.mm.yy"
                placeholder="Başlama tarihi"
                :show-time="false"
                class="w-full"
              />
            </div>

            <div class="form-group">
              <label for="endDate" class="block font-medium mb-2">Bitiş Tarihi</label>
              <Calendar
                id="endDate"
                v-model="form.end_date"
                date-format="dd.mm.yy"
                placeholder="Bitiş tarihi"
                :show-time="false"
                class="w-full"
              />
            </div>

            <div class="form-group md:col-span-2">
              <label for="description" class="block font-medium mb-2">Açıklama</label>
              <Textarea
                id="description"
                v-model="form.description"
                placeholder="Proje açıklaması"
                rows="3"
                class="w-full"
              />
            </div>

            <div class="form-group md:col-span-2">
              <label for="note" class="block font-medium mb-2">Notlar</label>
              <Textarea
                id="note"
                v-model="form.note"
                placeholder="Ek notlar"
                rows="3"
                class="w-full"
              />
            </div>
          </div>

          <div class="form-actions flex gap-2 pt-6 border-t">
            <Button
              type="submit"
              :label="isEdit ? 'Güncelle' : 'Oluştur'"
              icon="pi pi-check"
              :loading="submitting"
              severity="success"
            />
            <Button
              type="button"
              label="İptal"
              icon="pi pi-times"
              severity="secondary"
              @click="cancel"
            />
          </div>
        </form>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.project-form-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .form-card {
  border-color: #334155;
  background-color: #1e293b;
}

.card-header {
  padding: 1.5rem;
  border-bottom: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .card-header {
  border-bottom-color: #334155;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin: 0;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  color: var(--p-text-color, #000);
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-start;
}
</style>
