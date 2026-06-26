<template>
  <div class="card">
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-2xl font-bold">{{ $t('superadmin.companies.subtitle') }}</h2>
      <div class="flex gap-2">
        <Button icon="pi pi-plus" :label="$t('superadmin.companies.addNew')" @click="openNew" outlined severity="success" />
        <Button icon="pi pi-refresh" :label="$t('superadmin.plans.refresh')" @click="loadData" :loading="store.loading" outlined />
      </div>
    </div>

    <DataTable :value="store.companies" :loading="store.loading" responsiveLayout="scroll" paginator :rows="10" :rowsPerPageOptions="[10, 20, 50]">
      <Column field="name" :header="$t('superadmin.companies.name')" style="max-width: 220px;">
        <template #body="slotProps">
          <span class="truncate-cell" :title="slotProps.data.name">{{ slotProps.data.name }}</span>
        </template>
      </Column>
      <Column field="country" :header="$t('superadmin.companies.country')"></Column>
      <Column field="admin_name" :header="$t('superadmin.companies.admin')"></Column>
      <Column field="admin_email" :header="$t('superadmin.companies.adminEmail')"></Column>
      <Column field="subscription_status" :header="$t('superadmin.companies.status')">
        <template #body="slotProps">
          <Tag :severity="getStatusSeverity(slotProps.data.subscription_status)" :value="slotProps.data.subscription_status.toUpperCase()" />
        </template>
      </Column>
      <Column field="stats.total_users" :header="$t('superadmin.companies.users')"></Column>
      <Column field="stats.total_invoices" :header="$t('superadmin.companies.invoices')"></Column>
      <Column :header="$t('superadmin.companies.actions')" style="width: 180px;">
        <template #body="slotProps">
          <Button icon="pi pi-pencil" class="rounded-md p-button-sm mr-1" @click="editCompany(slotProps.data)" :title="$t('superadmin.plans.edit')" text severity="warn" />
          <Button icon="pi pi-trash" class="rounded-md p-button-sm mr-1" @click="confirmDelete(slotProps.data)" :title="$t('superadmin.plans.delete')" text severity="danger" />
          <Button v-if="slotProps.data.subscription_status === 'canceled'" icon="pi pi-check" class="rounded-md p-button-sm" @click="toggleStatus(slotProps.data.id, 'activate')" :title="$t('superadmin.companies.activate')" text severity="primary" />
          <Button v-else icon="pi pi-ban" class="rounded-md p-button-danger p-button-sm" @click="toggleStatus(slotProps.data.id, 'suspend')" :title="$t('superadmin.companies.suspend')" text />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" :style="{ width: '95vw', maxWidth: '1000px' }" :header="editMode ? $t('superadmin.companies.edit') : $t('superadmin.companies.addNew')" :modal="true" class="p-fluid">
      <div class="flex flex-col gap-4 bg-slate-50 p-4 -mx-4 -mt-2">
        
        <!-- Genel & Finansal Bilgiler Kartı -->
        <div class="bg-white p-4 rounded-lg shadow-sm border border-slate-200">
          <div class="grid grid-cols-12 gap-x-4 gap-y-4">
            <div class="field col-span-12 md:col-span-8">
              <label for="name" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.name') }} *</label>
              <InputText id="name" v-model.trim="company.name" required="true" maxlength="255" :class="{'p-invalid': submitted && !company.name}" class="w-full" />
              <small class="p-error" v-if="submitted && !company.name">{{ $t('superadmin.companies.nameRequired') }}</small>
            </div>

            <div class="field col-span-12 md:col-span-4">
              <label for="industry" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.industry') }}</label>
              <Select id="industry" v-model="company.industry" :options="sectorOptions" optionLabel="label" optionValue="value" :placeholder="$t('common.selectPlaceholder')" filter showClear class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-4">
              <label for="contact_name" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.contactName') }}</label>
              <InputText id="contact_name" v-model.trim="company.contact_name" maxlength="255" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-4">
              <label for="tax_office" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.taxOffice') }}</label>
              <InputText id="tax_office" v-model.trim="company.tax_office" maxlength="255" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-4">
              <label for="tax_number" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.taxNumber') }}</label>
              <InputText id="tax_number" v-model.trim="company.tax_number" maxlength="11" class="w-full" />
            </div>
          </div>
        </div>

        <!-- İletişim Bilgileri Kartı -->
        <div class="bg-white p-4 rounded-lg shadow-sm border border-slate-200">
          <h3 class="font-bold text-base text-slate-800 mb-4">{{ $t('superadmin.companies.contactInfo') || 'İletişim Bilgileri' }}</h3>
          
          <div class="grid grid-cols-12 gap-x-4 gap-y-4">
            <div class="field col-span-12 md:col-span-2">
              <label for="country" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.country') }}</label>
              <InputText id="country" v-model.trim="company.country" maxlength="255" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-2">
              <label for="city" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.city') }}</label>
              <InputText id="city" v-model.trim="company.city" maxlength="255" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-2">
              <label for="district" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.district') }}</label>
              <InputText id="district" v-model.trim="company.district" maxlength="255" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-6">
              <label for="address" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.address') }}</label>
              <InputText id="address" v-model="company.address" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-3">
              <label for="phone" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.phone') }}</label>
              <PhoneInput id="phone" v-model="company.phone" :maxlength="50" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-3">
              <label for="fax" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.fax') }}</label>
              <InputText id="fax" v-model.trim="company.fax" maxlength="50" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-3">
              <label for="landline" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.landline') }}</label>
              <InputText id="landline" v-model.trim="company.landline" maxlength="50" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-3">
              <label for="email" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.email') }}</label>
              <InputText id="email" v-model.trim="company.email" maxlength="255" class="w-full" />
            </div>
          </div>
        </div>

        <!-- Sistem ve Abonelik Kartı -->
        <div class="bg-white p-4 rounded-lg shadow-sm border border-slate-200">
          <h3 class="font-bold text-base text-slate-800 mb-4">{{ $t('superadmin.companies.systemInfo') || 'Sistem ve Abonelik' }}</h3>
          
          <div class="grid grid-cols-12 gap-x-4 gap-y-4">
            <div class="field col-span-12 md:col-span-3">
              <label for="currency" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.currency') }} *</label>
              <Select id="currency" v-model="company.currency" :options="currencies" optionLabel="label" optionValue="value" :placeholder="$t('common.selectPlaceholder')" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-3">
              <label for="locale" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.locale') }} *</label>
              <Select id="locale" v-model="company.locale" :options="locales" optionLabel="label" optionValue="value" :placeholder="$t('common.selectPlaceholder')" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-3">
              <label for="timezone" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.timezone') }}</label>
              <InputText id="timezone" v-model.trim="company.timezone" maxlength="64" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-3">
              <label for="subscription_status" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.subscriptionStatus') }} *</label>
              <Select id="subscription_status" v-model="company.subscription_status" :options="subscriptionStatuses" optionLabel="label" optionValue="value" :placeholder="$t('common.selectPlaceholder')" class="w-full" />
            </div>

            <div class="field col-span-12 md:col-span-6">
              <label for="plan_id" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.plan') }}</label>
              <Select id="plan_id" v-model="company.plan_id" :options="store.plans" optionLabel="name" optionValue="id" :placeholder="$t('superadmin.companies.selectPlan')" showClear class="w-full" />
            </div>
          </div>
        </div>

        <!-- Initial Admin User Info Kartı (Only for New Company) -->
        <template v-if="!editMode">
          <div class="bg-white p-4 rounded-lg shadow-sm border border-slate-200">
            <h3 class="font-bold text-base text-slate-800 mb-4">{{ $t('superadmin.companies.adminSectionTitle') }}</h3>
            
            <div class="grid grid-cols-12 gap-x-4 gap-y-4">
              <div class="field col-span-12 md:col-span-4">
                <label for="admin_name" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.adminName') }} *</label>
                <InputText id="admin_name" v-model.trim="company.admin_name" required="true" maxlength="255" :class="{'p-invalid': submitted && !company.admin_name}" class="w-full" />
                <small class="p-error" v-if="submitted && !company.admin_name">{{ $t('superadmin.companies.adminNameRequired') }}</small>
              </div>

              <div class="field col-span-12 md:col-span-4">
                <label for="admin_email" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.adminEmail') }} *</label>
                <InputText id="admin_email" type="email" v-model.trim="company.admin_email" required="true" maxlength="255" :class="{'p-invalid': submitted && !company.admin_email}" class="w-full" />
                <small class="p-error" v-if="submitted && !company.admin_email">{{ $t('superadmin.companies.adminEmailRequired') }}</small>
              </div>

              <div class="field col-span-12 md:col-span-4">
                <label for="admin_password" class="font-semibold text-sm block mb-1 text-slate-700">{{ $t('superadmin.companies.adminPassword') }}</label>
                <InputText id="admin_password" type="password" v-model="company.admin_password" required="true" :class="{'p-invalid': submitted && (!company.admin_password || company.admin_password.length < 8)}" class="w-full" />
                <small class="p-error" v-if="submitted && (!company.admin_password || company.admin_password.length < 8)">{{ $t('superadmin.companies.adminPasswordHint') }}</small>
              </div>
            </div>
          </div>
        </template>
      </div>

      <template #footer>
        <Button :label="$t('superadmin.plans.cancel')" icon="pi pi-times" class="p-button-text" @click="hideDialog" outlined />
        <Button :label="$t('superadmin.plans.save')" icon="pi pi-check" @click="saveCompany" outlined severity="primary" />
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSuperadminStore } from '@/stores/superadmin'
import { useToast } from 'primevue/usetoast'
import PhoneInput from '@/components/PhoneInput.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import { sectors } from '@/constants/sectors'

const store = useSuperadminStore()
const toast = useToast()
const { t } = useI18n()

const dialogVisible = ref(false)
const editMode = ref(false)
const company = ref<any>({})
const submitted = ref(false)

const subscriptionStatuses = computed(() => [
  { label: t('superadmin.companies.statusTrial'), value: 'trial' },
  { label: t('superadmin.companies.statusActive'), value: 'active' },
  { label: t('superadmin.companies.statusPastDue'), value: 'past_due' },
  { label: t('superadmin.companies.statusCanceled'), value: 'canceled' },
])

const currencies = ref([
  { label: '₺ Türk Lirası (TRY)', value: 'TRY' },
  { label: '$ Amerikan Doları (USD)', value: 'USD' },
  { label: '€ Euro (EUR)', value: 'EUR' },
  { label: '£ İngiliz Sterlini (GBP)', value: 'GBP' },
  { label: '₽ Rus Rublesi (RUB)', value: 'RUB' },
])

const locales = ref([
  { label: 'Türkçe', value: 'tr' },
  { label: 'English', value: 'en' },
])

const sectorOptions = ref(sectors.map((s) => ({ label: s, value: s })))

const loadData = async () => {
  try {
    await store.fetchCompanies()
    await store.fetchPlans()
  } catch (e: any) {
    toast.add({ severity: 'error', summary: t('common.error'), detail: t('superadmin.companies.loadError') })
  }
}

const toggleStatus = async (id: string, action: 'suspend' | 'activate') => {
  try {
    await store.toggleCompanyStatus(id, action)
    toast.add({ severity: 'success', summary: t('common.success'), detail: t('superadmin.companies.statusUpdateSuccess') })
  } catch (e: any) {
    toast.add({ severity: 'error', summary: t('common.error'), detail: t('superadmin.companies.statusUpdateError') })
  }
}

const getStatusSeverity = (status: string) => {
  switch (status) {
    case 'active': return 'success'
    case 'trial': return 'info'
    case 'past_due': return 'warning'
    case 'canceled': return 'danger'
    default: return 'info'
  }
}

const openNew = () => {
  company.value = {
    name: '',
    email: '',
    phone: '',
    tax_office: '',
    tax_number: '',
    address: '',
    currency: 'TRY',
    locale: 'tr',
    subscription_status: 'trial',
    plan_id: null,
    contact_name: '',
    landline: '',
    fax: '',
    industry: '',
    country: 'Türkiye',
    city: '',
    district: '',
    timezone: 'Europe/Istanbul',
    admin_name: '',
    admin_email: '',
    admin_password: '',
  }
  editMode.value = false
  submitted.value = false
  dialogVisible.value = true
}

const editCompany = (data: any) => {
  company.value = { ...data }
  editMode.value = true
  submitted.value = false
  dialogVisible.value = true
}

const hideDialog = () => {
  dialogVisible.value = false
  submitted.value = false
}

const saveCompany = async () => {
  submitted.value = true
  if (!company.value.name?.trim()) return
  if (!company.value.currency || !company.value.locale || !company.value.subscription_status) return

  if (!editMode.value) {
    if (!company.value.admin_name?.trim() || !company.value.admin_email?.trim() || !company.value.admin_password || company.value.admin_password.length < 8) {
      return
    }
  }

  try {
    if (editMode.value) {
      await store.updateCompany(company.value.id, {
        name: company.value.name,
        email: company.value.email,
        phone: company.value.phone,
        tax_office: company.value.tax_office,
        tax_number: company.value.tax_number,
        address: company.value.address,
        currency: company.value.currency,
        locale: company.value.locale,
        subscription_status: company.value.subscription_status,
        plan_id: company.value.plan_id,
        contact_name: company.value.contact_name,
        landline: company.value.landline,
        fax: company.value.fax,
        industry: company.value.industry,
        country: company.value.country,
        city: company.value.city,
        district: company.value.district,
        timezone: company.value.timezone,
      })
      toast.add({ severity: 'success', summary: t('common.success'), detail: t('superadmin.companies.profileUpdateSuccess'), life: 10000 })
    } else {
      await store.createCompany(company.value)
      toast.add({ severity: 'success', summary: t('common.success'), detail: t('superadmin.companies.createSuccess'), life: 10000 })
    }
    dialogVisible.value = false
  } catch (e: any) {
    const detail = e.response?.data?.error?.message || t('superadmin.companies.saveErrorGeneric')
    toast.add({ severity: 'error', summary: t('common.error'), detail: detail, life: 10000 })
  }
}

const confirmDelete = async (data: any) => {
  if (confirm(t('superadmin.companies.confirmDelete', { name: data.name }))) {
    try {
      await store.deleteCompany(data.id)
      toast.add({ severity: 'success', summary: t('common.success'), detail: t('superadmin.companies.deleteSuccess'), life: 10000 })
    } catch (e: any) {
      toast.add({ severity: 'error', summary: t('common.error'), detail: t('superadmin.companies.deleteError'), life: 10000 })
    }
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.truncate-cell {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}
</style>
