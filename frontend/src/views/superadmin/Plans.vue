<template>
  <div class="card">
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-2xl font-bold">{{ $t('superadmin.plans.title') }}</h2>
      <div class="flex gap-2">
        <Button icon="pi pi-plus" :label="$t('superadmin.plans.addNew')" @click="openNew" outlined severity="success" />
        <Button icon="pi pi-refresh" :label="$t('superadmin.plans.refresh')" @click="loadData" :loading="store.loading" outlined />
      </div>
    </div>

    <div class="flex flex-col gap-6">
      <div v-for="group in groupedPlans" :key="group.code" class="p-6 bg-slate-50 dark:bg-slate-800/40 rounded-xl border border-slate-200 dark:border-slate-700">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-bold text-slate-800 dark:text-white">
            {{ group.name }} <span class="text-sm font-semibold text-slate-400">({{ group.code }})</span>
          </h3>
          <Button icon="pi pi-plus" :label="$t('superadmin.plans.addVariant')" class="p-button-sm" @click="openNewVariant(group.code, group.name)" outlined severity="success" />
        </div>
        
        <DataTable :value="group.variants" :loading="store.loading" responsiveLayout="scroll" class="p-datatable-sm w-full">
          <Column field="currency" :header="$t('superadmin.plans.currency')"></Column>
          <Column field="price_monthly" :header="$t('superadmin.plans.priceMonthly')">
            <template #body="slotProps">
              {{ slotProps.data.price_monthly }} {{ slotProps.data.currency }}
            </template>
          </Column>
          <Column field="price_yearly" :header="$t('superadmin.plans.priceYearly')">
            <template #body="slotProps">
              {{ slotProps.data.price_yearly }} {{ slotProps.data.currency }}
            </template>
          </Column>
          <Column field="is_active" :header="$t('superadmin.plans.status')">
            <template #body="slotProps">
              <Tag :severity="slotProps.data.is_active ? 'success' : 'danger'" :value="slotProps.data.is_active ? $t('superadmin.plans.active') : $t('superadmin.plans.inactive')" />
            </template>
          </Column>
          <Column :header="$t('superadmin.plans.actions')" style="width: 120px;">
            <template #body="slotProps">
              <Button icon="pi pi-pencil" class="rounded-md p-button-sm mr-1" @click="editPlan(slotProps.data)" :title="$t('superadmin.plans.edit')" text severity="warn" />
              <Button icon="pi pi-trash" class="rounded-md p-button-sm" @click="confirmDeletePlan(slotProps.data)" :title="$t('superadmin.plans.delete')" text severity="danger" />
            </template>
          </Column>
        </DataTable>
      </div>
    </div>

    <Dialog v-model:visible="planDialog" :style="{ width: '95vw', maxWidth: '800px' }" :header="$t('superadmin.plans.detailHeader')" :modal="true" class="p-fluid">
      <div class="flex flex-col gap-3 pt-2">
        <div class="field">
          <label for="name" class="font-semibold text-sm block mb-1">{{ $t('superadmin.plans.name') }} *</label>
          <InputText id="name" v-model.trim="plan.name" required="true" autofocus :class="{'p-invalid': submitted && !plan.name}" maxlength="255" />
          <small class="p-error" v-if="submitted && !plan.name">{{ $t('superadmin.plans.nameRequired') }}</small>
        </div>
        <div class="field">
          <label for="code" class="font-semibold text-sm block mb-1">{{ $t('superadmin.plans.code') }} *</label>
          <InputText id="code" v-model.trim="plan.code" required="true" :class="{'p-invalid': submitted && !plan.code}" maxlength="255" />
          <small class="p-error" v-if="submitted && !plan.code">{{ $t('superadmin.plans.codeRequired') }}</small>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="field">
            <label for="price_monthly" class="font-semibold text-sm block mb-1">{{ $t('superadmin.plans.priceMonthly') }} *</label>
            <InputNumber id="price_monthly" v-model="plan.price_monthly" mode="currency" :currency="plan.currency || 'TRY'" locale="tr-TR" />
          </div>
          <div class="field">
            <label for="price_yearly" class="font-semibold text-sm block mb-1">{{ $t('superadmin.plans.priceYearly') }} *</label>
            <InputNumber id="price_yearly" v-model="plan.price_yearly" mode="currency" :currency="plan.currency || 'TRY'" locale="tr-TR" />
          </div>
        </div>
        <div class="field">
          <label for="currency" class="font-semibold text-sm block mb-1">{{ $t('superadmin.plans.currency') }} *</label>
          <Select id="currency" v-model="plan.currency" :options="currencies" optionLabel="label" optionValue="value" :placeholder="$t('common.selectPlaceholder')" />
        </div>
        
        <div class="field">
          <label class="font-semibold text-sm block mb-2">{{ $t('superadmin.plans.activeModules') }}</label>
          <div class="grid grid-cols-2 gap-2 border p-3 rounded-md bg-slate-50 dark:bg-slate-800">
            <div v-for="mod in availableModules" :key="mod.value" class="flex items-center gap-2">
              <Checkbox :id="'mod_' + mod.value" v-model="selectedFeatures" :value="mod.value" />
              <label :for="'mod_' + mod.value" class="text-sm cursor-pointer">{{ mod.label }}</label>
            </div>
          </div>
        </div>

        <div class="field flex items-center gap-2 mt-2">
          <Checkbox id="is_active" v-model="plan.is_active" :binary="true" />
          <label for="is_active" class="font-semibold text-sm cursor-pointer">{{ $t('superadmin.plans.isActive') }}</label>
        </div>
      </div>

      <template #footer>
        <Button :label="$t('superadmin.plans.cancel')" icon="pi pi-times" class="p-button-text" @click="hideDialog" outlined />
        <Button :label="$t('superadmin.plans.save')" icon="pi pi-check" @click="savePlan" outlined severity="primary" />
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSuperadminStore } from '@/stores/superadmin'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Checkbox from 'primevue/checkbox'
import Select from 'primevue/select'

const store = useSuperadminStore()
const toast = useToast()
const { t } = useI18n()

const planDialog = ref(false)
const plan = ref<any>({})
const submitted = ref(false)
const selectedFeatures = ref<string[]>([])

const groupedPlans = computed(() => {
  const groups: Record<string, any[]> = {}
  const list = store.plans || []
  for (const p of list) {
    if (!groups[p.code]) groups[p.code] = []
    groups[p.code].push(p)
  }
  return Object.entries(groups).map(([code, plans]) => ({
    code,
    name: plans[0].name,
    variants: plans,
  }))
})

const currencies = ref([
  { label: 'Türk Lirası (TRY)', value: 'TRY' },
  { label: 'Amerikan Doları (USD)', value: 'USD' },
  { label: 'Euro (EUR)', value: 'EUR' },
  { label: 'İngiliz Sterlini (GBP)', value: 'GBP' },
])

const availableModules = ref([
  { label: 'Panel (Dashboard)', value: 'dashboard' },
  { label: 'Firmalar / Cariler', value: 'caris' },
  { label: 'Faturalar', value: 'invoices' },
  { label: 'Teklifler', value: 'quotes' },
  { label: 'Ödemeler', value: 'payments' },
  { label: 'Kasa / Banka', value: 'cash' },
  { label: 'Giderler', value: 'expenses' },
  { label: 'Ürün / Stok Yönetimi', value: 'products' },
  { label: 'Personel Yönetimi', value: 'employees' },
  { label: 'Proje Yönetimi', value: 'projects' },
  { label: 'Raporlar', value: 'reports' },
  { label: 'Şirket Ayarları', value: 'settings' },
])

const loadData = async () => {
  try {
    await store.fetchPlans()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: t('common.error'), detail: t('superadmin.plans.loadError') })
  }
}

const openNew = () => {
  plan.value = {
    name: '',
    code: '',
    price_monthly: 0,
    price_yearly: 0,
    currency: 'TRY',
    is_active: true
  }
  selectedFeatures.value = ['dashboard', 'caris', 'invoices', 'payments', 'expenses', 'products', 'employees', 'settings']
  submitted.value = false
  planDialog.value = true
}

const openNewVariant = (code: string, name: string) => {
  openNew()
  plan.value.code = code
  plan.value.name = name
}

const hideDialog = () => {
  planDialog.value = false
  submitted.value = false
}

const savePlan = async () => {
  submitted.value = true
  if (plan.value.name?.trim() && plan.value.code?.trim()) {
    const payload = {
      ...plan.value,
      features: JSON.stringify(selectedFeatures.value),
      price_monthly: parseFloat(plan.value.price_monthly) || 0,
      price_yearly: parseFloat(plan.value.price_yearly) || 0
    }

    try {
      if (plan.value.id) {
        await store.updatePlan(plan.value.id, payload)
        toast.add({ severity: 'success', summary: t('common.success'), detail: t('superadmin.plans.updateSuccess'), life: 10000 })
      } else {
        await store.createPlan(payload)
        toast.add({ severity: 'success', summary: t('common.success'), detail: t('superadmin.plans.createSuccess'), life: 10000 })
      }
      planDialog.value = false
      plan.value = {}
    } catch (e: any) {
      const detail = e.response?.data?.error?.message || t('superadmin.plans.saveError')
      toast.add({ severity: 'error', summary: t('common.error'), detail: detail, life: 10000 })
    }
  }
}

const editPlan = (p: any) => {
  plan.value = { ...p }
  
  let parsed: string[] = []
  if (typeof plan.value.features === 'string') {
    try {
      parsed = JSON.parse(plan.value.features)
    } catch (e) {
      parsed = []
    }
  } else if (Array.isArray(plan.value.features)) {
    parsed = plan.value.features
  }
  selectedFeatures.value = parsed
  planDialog.value = true
}

const confirmDeletePlan = async (data: any) => {
  if (confirm(t('superadmin.plans.confirmDelete', { name: data.name }))) {
    try {
      await store.deletePlan(data.id)
      toast.add({ severity: 'success', summary: t('common.success'), detail: t('superadmin.plans.deleteSuccess'), life: 10000 })
    } catch (e: any) {
      toast.add({ severity: 'error', summary: t('common.error'), detail: t('superadmin.plans.deleteError'), life: 10000 })
    }
  }
}

onMounted(() => {
  loadData()
})
</script>
