<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useEmployeeStore } from '@/stores/employee'
import { useSettingsStore } from '@/stores/settings'
import { useRoleStore } from '@/stores/role'
import { useToast } from 'primevue/usetoast'
import { getCurrentCompanyDatetimeLocal, toBackendDate, toCompanyDatetimeLocal } from '@/utils/date'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import Checkbox from 'primevue/checkbox'
import Menu from 'primevue/menu'
import Select from 'primevue/select'
import PhoneInput from '@/components/PhoneInput.vue'
import { exportToPDF } from '@/utils/pdfExport'

const employeeStore = useEmployeeStore()
const settingsStore = useSettingsStore()
const roleStore = useRoleStore()
const toast = useToast()

const search = ref('')
const page = ref(1)
const limit = ref(20)
const displayDialog = ref(false)
const isEdit = ref(false)
const selectedId = ref('')
const dt = ref()
const selectedItems = ref([])
const sortField = ref('')
const sortOrder = ref(1)

const departments = ref<string[]>([])
const positions = ref<string[]>([])

const form = ref({
  name: '',
  email: '',
  phone: '',
  position: '',
  department: '',
  hire_date: null as string | null,
  give_login_permission: false,
  password: '',
  role_id: null as string | null,
})

const loadEmployees = async () => {
  const params = {
    page: page.value,
    limit: limit.value,
    search: search.value,
    sort: sortField.value ? `${sortField.value} ${sortOrder.value === 1 ? 'asc' : 'desc'}` : ''
  }
  try {
    await employeeStore.fetchEmployees(params)
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Personel listesi yüklenemedi.', life: 10000 })
  }
}

const onSort = (event: any) => {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  loadEmployees()
}

const loadSettings = async () => {
  try {
    const depVal = await settingsStore.fetchSetting('employee_departments')
    if (depVal) departments.value = JSON.parse(depVal)
    
    const posVal = await settingsStore.fetchSetting('employee_positions')
    if (posVal) positions.value = JSON.parse(posVal)
  } catch (e) {
    console.error('Ayarlar yüklenemedi', e)
  }
}

onMounted(() => {
  loadEmployees()
  loadSettings()
  roleStore.fetchRoles()
})

watch(search, () => {
  page.value = 1
  loadEmployees()
})

const openNew = () => {
  isEdit.value = false
  selectedId.value = ''
  form.value = {
    name: '',
    email: '',
    phone: '',
    position: '',
    department: '',
    hire_date: null,
    give_login_permission: false,
    password: '',
    role_id: null,
  }
  displayDialog.value = true
}

const hadLoginPermission = ref(false)

const openEdit = (emp: any) => {
  isEdit.value = true
  selectedId.value = emp.id
  hadLoginPermission.value = !!emp.user_id
  form.value = {
    name: emp.name,
    email: emp.email,
    phone: emp.phone || '',
    position: emp.position || '',
    department: emp.department || '',
    hire_date: emp.hire_date ? toCompanyDatetimeLocal(emp.hire_date) : null,
    give_login_permission: !!emp.user_id,
    password: '',
    role_id: emp.role_id || null,
  }
  displayDialog.value = true
}

const save = async () => {
  if (!form.value.name || !form.value.email) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'İsim ve E-posta girilmesi zorunludur.', life: 10000 })
    return
  }

  try {
    if (isEdit.value) {
      const grantingNewLogin = form.value.give_login_permission && !hadLoginPermission.value
      if (grantingNewLogin && (!form.value.password || form.value.password.length < 8)) {
        toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Giriş yetkisi için en az 8 karakter şifre giriniz.', life: 10000 })
        return
      }
      if (hadLoginPermission.value && form.value.password && form.value.password.length < 8) {
        toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Şifre en az 8 karakter olmalıdır.', life: 10000 })
        return
      }
      await employeeStore.updateEmployee(selectedId.value, {
        name: form.value.name,
        phone: form.value.phone,
        position: form.value.position,
        department: form.value.department,
        hire_date: form.value.hire_date ? toBackendDate(form.value.hire_date) : null,
        is_active: true,
        role_id: form.value.give_login_permission ? form.value.role_id : null,
        give_login_permission: form.value.give_login_permission,
        password: form.value.password || null,
      })
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Personel güncellendi.', life: 10000 })
    } else {
      if (form.value.give_login_permission && (!form.value.password || form.value.password.length < 8)) {
        toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Giriş yetkisi için en az 8 karakter şifre giriniz.', life: 10000 })
        return
      }
      await employeeStore.createEmployee({
        name: form.value.name,
        email: form.value.email,
        phone: form.value.phone,
        position: form.value.position,
        department: form.value.department,
        hire_date: form.value.hire_date ? toBackendDate(form.value.hire_date) : null,
        give_login_permission: form.value.give_login_permission,
        password: form.value.password,
        role_id: form.value.give_login_permission ? form.value.role_id : null,
      })
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Personel oluşturuldu.', life: 10000 })
    }
    displayDialog.value = false
    loadEmployees()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'İşlem başarısız oldu.', life: 10000 })
  }
}

const confirmDelete = async (id: string) => {
  if (confirm('Bu personeli silmek istediğinize emin misiniz?')) {
    try {
      await employeeStore.deleteEmployee(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Personel silindi.', life: 10000 })
      loadEmployees()
    } catch {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Silme işlemi başarısız.', life: 10000 })
    }
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('tr-TR')
}

const getRoleLabel = (emp: any) => {
  if (!emp.user) return 'Giriş Yetkisi Yok'
  if (emp.user.role === 'superadmin') return 'Super Admin'
  if (emp.user.role === 'admin') return 'Yönetici (Admin)'
  if (emp.user.role === 'personel') {
    return emp.user.role_ref?.name || 'Personel'
  }
  return emp.user.role
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
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : employeeStore.employees
    const columns = [
      { header: 'Ad Soyad', dataKey: 'name' },
      { header: 'E-posta', dataKey: 'email' },
      { header: 'Telefon', dataKey: 'phone' },
      { header: 'Pozisyon', dataKey: 'position' },
      { header: 'Departman', dataKey: 'department' },
      { header: 'İşe Giriş', dataKey: 'hire_date' },
      { header: 'Durum', dataKey: 'status' }
    ]
    exportToPDF('Personel_Listesi', columns, dataToExport.map(item => ({
      ...item,
      hire_date: formatDate(item.hire_date),
      status: item.is_active ? 'Aktif' : 'Pasif'
    })))
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'PDF olarak dışa aktarıldı', life: 10000 })
  }
}
</script>

<template>
  <div class="employee-view">
    <!-- Filters and Table -->
    <Card class="table-card mt-4">
      <template #content>
        <div class="filters-header flex flex-col md:flex-row justify-between items-center gap-4 mb-4">
          <!-- Left: Filters (Empty for now to match layout) -->
          <div class="select-filters flex gap-2 w-full md:w-auto">
          </div>

          <!-- Middle: Search -->
          <div class="flex-1 flex justify-center w-full md:w-auto">
            <div class="search-input w-full max-w-md relative">
              <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"></i>
              <InputText v-model="search" placeholder="İsim veya e-posta ile ara..." class="w-full pl-10" />
            </div>
          </div>

          <!-- Right: Buttons -->
          <div class="flex gap-2 w-full md:w-auto justify-end shrink-0">
            <Button label="Ekle" icon="pi pi-plus" @click="openNew" severity="success" />
            <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
            <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
          </div>
        </div>
        <DataTable
          ref="dt"
          :value="employeeStore.employees"
          v-model:selection="selectedItems"
          :loading="employeeStore.loading"
          @sort="onSort"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          dataKey="id"
          paginator
          :rows="limit"
          :rowsPerPageOptions="[20, 50, 100]"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ (page - 1) * limit + slotProps.index + 1 }}
            </template>
          </Column>
          <Column selectionMode="multiple" headerStyle="width: 3rem"></Column>
          <Column field="name" header="Ad Soyad" sortable style="min-width: 200px">
            <template #body="{ data }">
              <span class="truncate max-w-[220px] inline-block" :title="data.name">{{ data.name }}</span>
            </template>
          </Column>
          <Column field="email" header="E-posta" sortable style="min-width: 200px">
            <template #body="{ data }">
              <span class="truncate max-w-[220px] inline-block" :title="data.email">{{ data.email }}</span>
            </template>
          </Column>
          <Column field="phone" header="Telefon" sortable style="min-width: 130px" />
          <Column field="position" header="Görev / Pozisyon" sortable style="min-width: 150px">
            <template #body="{ data }">
              <span class="truncate max-w-[160px] inline-block" :title="data.position">{{ data.position }}</span>
            </template>
          </Column>
          <Column field="department" header="Departman" sortable style="min-width: 150px">
            <template #body="{ data }">
              <span class="truncate max-w-[160px] inline-block" :title="data.department">{{ data.department }}</span>
            </template>
          </Column>
          <Column field="user.role" header="Rol" style="min-width: 150px">
            <template #body="{ data }">
              <span class="text-xs font-medium text-slate-600 dark:text-slate-400">{{ getRoleLabel(data) }}</span>
            </template>
          </Column>
          <Column field="hire_date" header="İşe Giriş" sortable style="min-width: 120px">
            <template #body="{ data }">{{ formatDate(data.hire_date) }}</template>
          </Column>
          <Column field="is_active" header="Durum" sortable style="min-width: 100px">
            <template #body="{ data }">
              <span :class="data.is_active ? 'badge-active' : 'badge-inactive'">
                {{ data.is_active ? 'Aktif' : 'Pasif' }}
              </span>
            </template>
          </Column>
          <Column header="İşlemler" style="min-width: 100px; text-align: right">
            <template #body="{ data }">
              <div class="flex gap-2 justify-end">
                <Button icon="pi pi-pencil" class="p-button-text p-button-sm" @click="openEdit(data)" severity="warn" />
                <Button icon="pi pi-trash" class="p-button-text p-button-sm" @click="confirmDelete(data.id)" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Create/Edit Form Dialog -->
    <Dialog
      v-model:visible="displayDialog"
      :header="isEdit ? 'Personel Düzenle' : 'Yeni Personel'"
      modal
      :style="{ width: '450px' }"
    >
      <div class="flex flex-col gap-4 mt-2">
        <div class="flex flex-col gap-1">
          <label for="name" class="font-semibold text-sm">Ad Soyad</label>
          <InputText id="name" v-model="form.name" class="w-full" maxlength="255" />
        </div>

        <div class="flex flex-col gap-1">
          <label for="email" class="font-semibold text-sm">E-posta</label>
          <InputText id="email" v-model="form.email" :disabled="isEdit" class="w-full" maxlength="255" />
        </div>

        <div class="flex flex-col gap-1">
          <label for="phone" class="font-semibold text-sm">Telefon</label>
          <PhoneInput id="phone" v-model="form.phone" />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-1">
            <label for="position" class="font-semibold text-sm">Pozisyon</label>
            <Select id="position" v-model="form.position" :options="positions" placeholder="Seçiniz" class="w-full" :editable="true" maxlength="100" />
          </div>
          <div class="flex flex-col gap-1">
            <label for="department" class="font-semibold text-sm">Departman</label>
            <Select id="department" v-model="form.department" :options="departments" placeholder="Seçiniz" class="w-full" :editable="true" maxlength="100" />
          </div>
        </div>

        <div class="flex flex-col gap-1">
          <label for="hire_date" class="font-semibold text-sm">İşe Başlama Tarihi</label>
          <input id="hire_date" type="datetime-local" v-model="form.hire_date" class="p-inputtext w-full" />
        </div>

        <div class="flex items-center gap-2 mt-2" v-if="!isEdit || !hadLoginPermission">
          <Checkbox id="login" v-model="form.give_login_permission" binary />
          <label for="login" class="font-semibold text-sm cursor-pointer">Sisteme Giriş Yetkisi Ver</label>
        </div>

        <div class="flex flex-col gap-1" v-if="form.give_login_permission">
          <label for="password" class="font-semibold text-sm">{{ isEdit ? 'Şifre (değiştirmek için doldurun)' : 'Şifre' }}</label>
          <InputText id="password" type="password" v-model="form.password" class="w-full" maxlength="50" :placeholder="isEdit ? 'Değiştirmek istemiyorsanız boş bırakın' : ''" />
        </div>

        <div class="flex flex-col gap-1" v-if="form.give_login_permission">
          <label for="role_id" class="font-semibold text-sm">Rol (Yetki Şablonu)</label>
          <Select id="role_id" v-model="form.role_id" :options="roleStore.roles" optionLabel="name" optionValue="id" placeholder="Rol seçin" class="w-full" />
        </div>
      </div>

      <template #footer>
        <div class="flex gap-2 justify-end">
          <Button label="İptal" class="p-button-text" @click="displayDialog = false" outlined />
          <Button label="Kaydet" @click="save" :loading="employeeStore.loading" outlined severity="primary" />
        </div>
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.employee-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.filter-card, .table-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .filter-card, :root.p-dark .table-card {
  border-color: #334155;
  background-color: #1e293b;
}

.badge-active {
  background-color: rgba(34, 197, 94, 0.1);
  color: #16a34a;
  padding: 0.25rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.8rem;
  font-weight: 600;
}

.badge-inactive {
  background-color: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  padding: 0.25rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.8rem;
  font-weight: 600;
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
</style>

