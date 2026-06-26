<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useProductStore } from '@/stores/product'
import type { Warehouse } from '@/stores/product'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'

const productStore = useProductStore()
const toast = useToast()

const newWarehouse = ref({ name: '', address: '', is_default: false })
const editingWarehouseId = ref<string | null>(null)
const editingWarehouse = ref({ name: '', address: '', is_default: false })

onMounted(async () => {
  await productStore.fetchWarehouses()
})

const handleCreateWarehouse = async () => {
  if (!newWarehouse.value.name.trim()) return
  try {
    await productStore.createWarehouse(newWarehouse.value)
    newWarehouse.value = { name: '', address: '', is_default: false }
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Depo eklendi', life: 10000 })
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Depo eklenemedi', life: 10000 })
  }
}

const startEditWarehouse = (wh: Warehouse) => {
  editingWarehouseId.value = wh.id
  editingWarehouse.value = {
    name: wh.name,
    address: wh.address || '',
    is_default: wh.is_default
  }
}

const handleUpdateWarehouse = async () => {
  if (!editingWarehouseId.value || !editingWarehouse.value.name.trim()) return
  try {
    await productStore.updateWarehouse(editingWarehouseId.value, editingWarehouse.value)
    editingWarehouseId.value = null
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Depo güncellendi', life: 10000 })
  } catch (err: any) {
    if (err.response?.data?.error?.message) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Depo güncellenemedi', life: 10000 })
    }
  }
}

const handleDeleteWarehouse = async (id: string) => {
  if (confirm('Bu depoyu silmek istediğinize emin misiniz?')) {
    try {
      await productStore.deleteWarehouse(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Depo silindi', life: 10000 })
    } catch (err: any) {
      if (err.response?.data?.error?.message) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Hareket görmüş depo silinemez.', life: 10000 })
      }
    }
  }
}
</script>

<template>
  <div class="warehouses-tab">
    <div class="mb-4">
      <div class="flex flex-col md:flex-row gap-2 items-center mb-6">
        <InputText v-model="newWarehouse.name" placeholder="Depo adı" class="w-full md:w-64" />
        <InputText v-model="newWarehouse.address" placeholder="Adres (opsiyonel)" class="w-full md:w-64" />
        <div class="flex items-center gap-2">
          <Checkbox v-model="newWarehouse.is_default" :binary="true" inputId="wh_default_new" />
          <label for="wh_default_new" class="text-sm text-slate-700 dark:text-slate-300">Varsayılan Depo</label>
        </div>
        <Button label="Ekle" icon="pi pi-plus" @click="handleCreateWarehouse" :disabled="!newWarehouse.name.trim()" outlined severity="success" />
      </div>
    </div>

    <DataTable :value="productStore.warehouses" class="p-datatable-sm w-full" responsiveLayout="scroll">
      <Column field="name" header="Depo Adı">
        <template #body="{ data }">
          <div v-if="editingWarehouseId === data.id">
            <InputText v-model="editingWarehouse.name" class="w-full p-inputtext-sm" />
          </div>
          <span v-else>{{ data.name }}</span>
        </template>
      </Column>
      <Column field="address" header="Adres">
        <template #body="{ data }">
          <div v-if="editingWarehouseId === data.id">
            <InputText v-model="editingWarehouse.address" class="w-full p-inputtext-sm" />
          </div>
          <span v-else>{{ data.address || '-' }}</span>
        </template>
      </Column>
      <Column field="is_default" header="Varsayılan" style="width: 100px; text-align: center;">
        <template #body="{ data }">
          <div v-if="editingWarehouseId === data.id" class="flex justify-center">
            <Checkbox v-model="editingWarehouse.is_default" :binary="true" />
          </div>
          <div v-else class="flex justify-center">
            <i class="pi pi-check-circle text-green-500" v-if="data.is_default"></i>
            <i class="pi pi-minus text-slate-300" v-else></i>
          </div>
        </template>
      </Column>
      <Column header="İşlemler" style="width: 150px">
        <template #body="{ data }">
          <div v-if="editingWarehouseId === data.id" class="flex gap-2">
            <Button icon="pi pi-check" class="p-button-sm rounded-md p-button-text" @click="handleUpdateWarehouse" title="Kaydet" severity="primary" />
            <Button icon="pi pi-times" class="p-button-sm rounded-md p-button-text" @click="editingWarehouseId = null" title="İptal" severity="warn" />
          </div>
          <div v-else class="flex gap-2">
            <Button icon="pi pi-pencil" class="p-button-sm rounded-md p-button-text" @click="startEditWarehouse(data)" title="Düzenle" severity="warn" />
            <Button icon="pi pi-trash" class="p-button-sm rounded-md p-button-text" @click="handleDeleteWarehouse(data.id)" title="Sil" severity="danger" />
          </div>
        </template>
      </Column>
    </DataTable>
  </div>
</template>
