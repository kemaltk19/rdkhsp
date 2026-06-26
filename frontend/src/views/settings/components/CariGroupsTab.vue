<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'

const settingsStore = useSettingsStore()
const toast = useToast()

const groups = ref<string[]>([])
const loading = ref(false)

const showAddDialog = ref(false)
const newGroup = ref('')
const editingGroup = ref('')
const isEditing = ref(false)

const loadGroups = async () => {
  loading.value = true
  try {
    const val = await settingsStore.fetchSetting('cari_groups')
    if (val) {
      groups.value = JSON.parse(val)
    } else {
      groups.value = ['Bireysel', 'Kurumsal', 'Kurum', 'Fabrika', 'Esnaf', 'Şirket', 'Diğer']
      await settingsStore.saveSetting('cari_groups', JSON.stringify(groups.value), 'cari')
    }
  } catch {
    groups.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadGroups()
})

const saveGroups = async () => {
  try {
    await settingsStore.saveSetting('cari_groups', JSON.stringify(groups.value), 'cari')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Cari grupları kaydedildi', life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Cari grupları kaydedilemedi', life: 10000 })
  }
}

const addGroup = () => {
  const trimmed = newGroup.value.trim()
  if (!trimmed) return
  
  if (isEditing.value && editingGroup.value) {
    const index = groups.value.indexOf(editingGroup.value)
    if (index > -1) {
      groups.value[index] = trimmed
    } else {
      groups.value.push(trimmed)
    }
  } else {
    if (!groups.value.includes(trimmed)) {
      groups.value.push(trimmed)
    }
  }
  
  saveGroups()
  showAddDialog.value = false
  newGroup.value = ''
  isEditing.value = false
  editingGroup.value = ''
}

const openAddDialog = () => {
  isEditing.value = false
  editingGroup.value = ''
  newGroup.value = ''
  showAddDialog.value = true
}

const openEditDialog = (group: string) => {
  isEditing.value = true
  editingGroup.value = group
  newGroup.value = group
  showAddDialog.value = true
}

const deleteGroup = (group: string) => {
  if (confirm(`"${group}" grubunu silmek istediğinize emin misiniz?`)) {
    groups.value = groups.value.filter((g: string) => g !== group)
    saveGroups()
  }
}
</script>

<template>
  <div class="max-w-xl">
    <div class="flex justify-between items-center mb-4">
      <span class="text-sm font-semibold text-slate-500 dark:text-slate-400">{{ groups.length }} Grup Kayıtlı</span>
      <Button label="Yeni Grup Ekle" icon="pi pi-plus" class="p-button-sm" @click="openAddDialog" outlined severity="success" />
    </div>

    <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden bg-white dark:bg-[#0f172a] shadow-sm">
      <DataTable :value="groups.map(g => ({ name: g }))" :loading="loading" class="p-datatable-sm w-full" responsiveLayout="scroll">
        <Column field="name" header="Grup Adı" sortable>
          <template #body="slotProps">
            <span class="font-semibold text-slate-700 dark:text-slate-200 text-sm">{{ slotProps.data.name }}</span>
          </template>
        </Column>
        <Column header="İşlemler" bodyClass="text-right pr-4 w-32" headerStyle="text-align: right; padding-right: 1.5rem;">
          <template #body="slotProps">
            <Button icon="pi pi-pencil" class="p-button-text p-button-sm p-1! mr-2" @click="openEditDialog(slotProps.data.name)" severity="warn" />
            <Button icon="pi pi-trash" class="p-button-text p-button-sm p-1!" @click="deleteGroup(slotProps.data.name)" severity="danger" />
          </template>
        </Column>
      </DataTable>
    </div>

    <Dialog v-model:visible="showAddDialog" :header="isEditing ? 'Cari Grubunu Düzenle' : 'Yeni Cari Grubu Ekle'" :modal="true" class="w-[400px]">
      <div class="flex flex-col gap-2 pt-2">
        <label class="text-sm font-bold text-slate-700 dark:text-slate-200">Grup Adı</label>
        <InputText v-model="newGroup" class="w-full" placeholder="Örn: VIP Müşteriler" @keyup.enter="addGroup" />
      </div>
      <template #footer>
        <Button label="İptal" icon="pi pi-times" class="p-button-text" @click="showAddDialog = false" />
        <Button :label="isEditing ? 'Kaydet' : 'Ekle'" icon="pi pi-check" @click="addGroup" outlined severity="warn" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
:deep(.p-datatable .p-datatable-tbody > tr > td) {
  padding: 0.5rem 1rem;
}
</style>
