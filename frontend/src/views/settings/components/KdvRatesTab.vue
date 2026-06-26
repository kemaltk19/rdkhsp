<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'

const settingsStore = useSettingsStore()
const toast = useToast()

const rates = ref<number[]>([])
const loading = ref(false)

const showAddDialog = ref(false)
const newRate = ref<number | null>(null)
const editingRate = ref<number | null>(null)
const isEditing = ref(false)

const loadRates = async () => {
  loading.value = true
  try {
    const val = await settingsStore.fetchSetting('kdv_rates')
    if (val) {
      rates.value = JSON.parse(val)
    } else {
      rates.value = [0, 1, 10, 20]
      await settingsStore.saveSetting('kdv_rates', JSON.stringify(rates.value), 'finance')
    }
  } catch {
    rates.value = [0, 1, 10, 20]
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadRates()
})

const saveRates = async () => {
  try {
    await settingsStore.saveSetting('kdv_rates', JSON.stringify(rates.value), 'finance')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'KDV oranları kaydedildi', life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'KDV oranları kaydedilemedi', life: 10000 })
  }
}

const addRate = () => {
  if (newRate.value === null) return
  
  if (isEditing.value && editingRate.value !== null) {
    // Düzenleme modu
    const index = rates.value.indexOf(editingRate.value)
    if (index > -1) {
      rates.value[index] = newRate.value
    } else {
      rates.value.push(newRate.value)
    }
  } else {
    // Ekleme modu
    if (!rates.value.includes(newRate.value)) {
      rates.value.push(newRate.value)
    }
  }
  
  rates.value.sort((a: number, b: number) => a - b)
  saveRates()
  showAddDialog.value = false
  newRate.value = null
  isEditing.value = false
  editingRate.value = null
}

const openAddDialog = () => {
  isEditing.value = false
  editingRate.value = null
  newRate.value = null
  showAddDialog.value = true
}

const openEditDialog = (rate: number) => {
  isEditing.value = true
  editingRate.value = rate
  newRate.value = rate
  showAddDialog.value = true
}

const deleteRate = (rate: number) => {
  rates.value = rates.value.filter((r: number) => r !== rate)
  saveRates()
}
</script>

<template>
  <div>
    <div class="flex justify-end mb-4 max-w-xl">
      <Button label="Ekle" icon="pi pi-plus" class="p-button-sm" @click="openAddDialog" outlined severity="success" />
    </div>

    <DataTable :value="rates.map((r: number) => ({ rate: r }))" :loading="loading" class="p-datatable-sm w-full max-w-xl" responsiveLayout="scroll">
      <Column field="rate" header="KDV Oranı (%)" sortable>
        <template #body="slotProps">
          <span class="font-bold">%{{ slotProps.data.rate }}</span>
        </template>
      </Column>
      <Column header="İşlemler" bodyClass="text-center w-24">
        <template #body="slotProps">
          <Button icon="pi pi-pencil" class="p-button-text p-button-sm p-1! mr-2" @click="openEditDialog(slotProps.data.rate)" severity="warn" />
          <Button icon="pi pi-trash" class="p-button-text p-button-sm p-1!" @click="deleteRate(slotProps.data.rate)" severity="danger" />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="showAddDialog" :header="isEditing ? 'KDV Oranını Düzenle' : 'Yeni KDV Oranı Ekle'" :modal="true" class="w-[400px]">
      <div class="flex flex-col gap-2 pt-2">
        <label>KDV Oranı (%)</label>
        <InputNumber v-model="newRate" :min="0" :max="100" class="w-full" suffix="%" />
      </div>
      <template #footer>
        <Button label="İptal" icon="pi pi-times" class="p-button-text" @click="showAddDialog = false" />
        <Button :label="isEditing ? 'Kaydet' : 'Ekle'" icon="pi pi-check" @click="addRate" severity="warn" />
      </template>
    </Dialog>
  </div>
</template>
