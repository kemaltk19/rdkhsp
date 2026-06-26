<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import RolesTab from './RolesTab.vue'

const settingsStore = useSettingsStore()
const toast = useToast()

const departments = ref<string[]>([])
const positions = ref<string[]>([])
const loading = ref(false)

const showDeptDialog = ref(false)
const showPosDialog = ref(false)

const newDept = ref('')
const newPos = ref('')

const editingIndex = ref<number | null>(null)

const loadData = async () => {
  loading.value = true
  try {
    const depVal = await settingsStore.fetchSetting('employee_departments')
    if (depVal) {
      departments.value = JSON.parse(depVal)
    } else {
      departments.value = ['Satış', 'Pazarlama', 'Muhasebe', 'İnsan Kaynakları', 'IT', 'Yönetim']
      await settingsStore.saveSetting('employee_departments', JSON.stringify(departments.value), 'employee')
    }

    const posVal = await settingsStore.fetchSetting('employee_positions')
    if (posVal) {
      positions.value = JSON.parse(posVal)
    } else {
      positions.value = ['Müdür', 'Müdür Yardımcısı', 'Uzman', 'Asistan', 'Stajyer']
      await settingsStore.saveSetting('employee_positions', JSON.stringify(positions.value), 'employee')
    }
  } catch {
    departments.value = []
    positions.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})

const saveDepartments = async () => {
  try {
    await settingsStore.saveSetting('employee_departments', JSON.stringify(departments.value), 'employee')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Departmanlar kaydedildi', life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Departmanlar kaydedilemedi', life: 10000 })
  }
}

const savePositions = async () => {
  try {
    await settingsStore.saveSetting('employee_positions', JSON.stringify(positions.value), 'employee')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Pozisyonlar kaydedildi', life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Pozisyonlar kaydedilemedi', life: 10000 })
  }
}

// Department Actions
const openAddDept = () => {
  editingIndex.value = null
  newDept.value = ''
  showDeptDialog.value = true
}

const openEditDept = (index: number) => {
  editingIndex.value = index
  newDept.value = departments.value[index]
  showDeptDialog.value = true
}

const saveDeptDialog = () => {
  if (!newDept.value.trim()) return
  if (editingIndex.value !== null) {
    departments.value[editingIndex.value] = newDept.value.trim()
  } else {
    departments.value.push(newDept.value.trim())
  }
  saveDepartments()
  showDeptDialog.value = false
}

const deleteDept = (index: number) => {
  departments.value.splice(index, 1)
  saveDepartments()
}

// Position Actions
const openAddPos = () => {
  editingIndex.value = null
  newPos.value = ''
  showPosDialog.value = true
}

const openEditPos = (index: number) => {
  editingIndex.value = index
  newPos.value = positions.value[index]
  showPosDialog.value = true
}

const savePosDialog = () => {
  if (!newPos.value.trim()) return
  if (editingIndex.value !== null) {
    positions.value[editingIndex.value] = newPos.value.trim()
  } else {
    positions.value.push(newPos.value.trim())
  }
  savePositions()
  showPosDialog.value = false
}

const deletePos = (index: number) => {
  positions.value.splice(index, 1)
  savePositions()
}

</script>

<template>
  <div class="flex flex-col gap-8">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
      <!-- Departmanlar -->
      <div>
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-medium">Departmanlar</h3>
          <Button label="Ekle" icon="pi pi-plus" class="p-button-sm" @click="openAddDept" outlined severity="success" />
        </div>

        <DataTable :value="departments.map((d, i) => ({ name: d, index: i }))" :loading="loading" class="p-datatable-sm p-datatable-gridlines w-full" responsiveLayout="scroll">
          <Column field="name" header="Departman Adı" sortable></Column>
          <Column header="İşlemler" bodyClass="text-center w-24" headerStyle="text-align: center;">
            <template #body="slotProps">
              <Button icon="pi pi-pencil" class="p-button-text p-button-sm p-1! mr-2" @click="openEditDept(slotProps.data.index)" severity="warn" />
              <Button icon="pi pi-trash" class="p-button-text p-button-sm p-1!" @click="deleteDept(slotProps.data.index)" severity="danger" />
            </template>
          </Column>
        </DataTable>
      </div>

      <!-- Pozisyonlar -->
      <div>
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-medium">Pozisyonlar / Ünvanlar</h3>
          <Button label="Ekle" icon="pi pi-plus" class="p-button-sm" @click="openAddPos" outlined severity="success" />
        </div>

        <DataTable :value="positions.map((p, i) => ({ name: p, index: i }))" :loading="loading" class="p-datatable-sm p-datatable-gridlines w-full" responsiveLayout="scroll">
          <Column field="name" header="Pozisyon Adı" sortable></Column>
          <Column header="İşlemler" bodyClass="text-center w-24" headerStyle="text-align: center;">
            <template #body="slotProps">
              <Button icon="pi pi-pencil" class="p-button-text p-button-sm p-1! mr-2" @click="openEditPos(slotProps.data.index)" severity="warn" />
              <Button icon="pi pi-trash" class="p-button-text p-button-sm p-1!" @click="deletePos(slotProps.data.index)" severity="danger" />
            </template>
          </Column>
        </DataTable>
      </div>

      <!-- Dialogs -->
      <Dialog v-model:visible="showDeptDialog" :header="editingIndex !== null ? 'Departmanı Düzenle' : 'Yeni Departman'" :modal="true" class="w-[400px]">
        <div class="flex flex-col gap-2 pt-2">
          <label>Departman Adı</label>
          <InputText v-model="newDept" class="w-full" @keyup.enter="saveDeptDialog" />
        </div>
        <template #footer>
          <Button label="İptal" icon="pi pi-times" class="p-button-text" @click="showDeptDialog = false" />
          <Button :label="editingIndex !== null ? 'Kaydet' : 'Ekle'" icon="pi pi-check" @click="saveDeptDialog" severity="warn" />
        </template>
      </Dialog>

      <Dialog v-model:visible="showPosDialog" :header="editingIndex !== null ? 'Pozisyonu Düzenle' : 'Yeni Pozisyon'" :modal="true" class="w-[400px]">
        <div class="flex flex-col gap-2 pt-2">
          <label>Pozisyon Adı</label>
          <InputText v-model="newPos" class="w-full" @keyup.enter="savePosDialog" />
        </div>
        <template #footer>
          <Button label="İptal" icon="pi pi-times" class="p-button-text" @click="showPosDialog = false" />
          <Button :label="editingIndex !== null ? 'Kaydet' : 'Ekle'" icon="pi pi-check" @click="savePosDialog" severity="warn" />
        </template>
      </Dialog>
    </div>

    <hr class="border-slate-200 dark:border-slate-700 my-4" />

    <!-- Roller ve Yetkiler -->
    <div>
      <RolesTab />
    </div>
  </div>
</template>

<style scoped>
:deep(.p-datatable .p-datatable-tbody > tr > td) {
  padding: 0.35rem 0.6rem !important;
}
:deep(.p-datatable .p-datatable-thead > tr > th) {
  padding: 0.45rem 0.6rem !important;
}
</style>
