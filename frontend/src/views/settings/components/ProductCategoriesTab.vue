<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useProductStore } from '@/stores/product'
import type { ProductCategory } from '@/stores/product'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Select from 'primevue/select'

const productStore = useProductStore()
const toast = useToast()

const newCategoryName = ref('')
const newCategoryKdv = ref('20')

const editingCategoryId = ref<string | null>(null)
const editingCategoryName = ref('')
const editingCategoryKdv = ref('20')

const kdvOptions = [
  { label: '%0', value: '0' },
  { label: '%1', value: '1' },
  { label: '%10', value: '10' },
  { label: '%20', value: '20' }
]

onMounted(async () => {
  await productStore.fetchCategories()
})

const handleCreateCategory = async () => {
  if (!newCategoryName.value.trim()) return
  try {
    await productStore.createCategory({ name: newCategoryName.value, default_kdv_rate: newCategoryKdv.value })
    newCategoryName.value = ''
    newCategoryKdv.value = '20'
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Kategori eklendi', life: 10000 })
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Kategori eklenemedi', life: 10000 })
  }
}

const startEditCategory = (cat: ProductCategory) => {
  editingCategoryId.value = cat.id
  editingCategoryName.value = cat.name
  editingCategoryKdv.value = cat.default_kdv_rate || '20'
}

const handleUpdateCategory = async () => {
  if (!editingCategoryId.value || !editingCategoryName.value.trim()) return
  try {
    await productStore.updateCategory(editingCategoryId.value, { name: editingCategoryName.value, default_kdv_rate: editingCategoryKdv.value })
    editingCategoryId.value = null
    editingCategoryName.value = ''
    editingCategoryKdv.value = '20'
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Kategori güncellendi', life: 10000 })
  } catch (err: any) {
    if (err.response?.data?.error?.message) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Kategori güncellenemedi', life: 10000 })
    }
  }
}

const handleDeleteCategory = async (id: string) => {
  if (confirm('Bu kategoriyi silmek istediğinize emin misiniz?')) {
    try {
      await productStore.deleteCategory(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Kategori silindi', life: 10000 })
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Bu kategori kullanımdadır ve silinemez.', life: 10000 })
      }
    }
  }
}
</script>

<template>
  <div class="product-categories-tab">
    <div class="mb-4">
      <div class="flex gap-2 items-center mb-6">
        <InputText v-model="newCategoryName" placeholder="Yeni kategori adı" class="w-full md:w-64" @keyup.enter="handleCreateCategory" />
        <Select v-model="newCategoryKdv" :options="kdvOptions" optionLabel="label" optionValue="value" placeholder="KDV" class="w-24" />
        <Button label="Ekle" icon="pi pi-plus" @click="handleCreateCategory" :disabled="!newCategoryName.trim()" outlined severity="success" />
      </div>
    </div>

    <DataTable :value="productStore.categories" class="p-datatable-sm w-full" responsiveLayout="scroll">
      <Column field="name" header="Kategori Adı" sortable>
        <template #body="{ data }">
          <div v-if="editingCategoryId === data.id" class="flex gap-2">
            <InputText v-model="editingCategoryName" class="w-full p-inputtext-sm" @keyup.enter="handleUpdateCategory" />
            <Select v-model="editingCategoryKdv" :options="kdvOptions" optionLabel="label" optionValue="value" class="w-24 p-inputtext-sm" />
          </div>
          <span v-else>{{ data.name }}</span>
        </template>
      </Column>
      <Column field="default_kdv_rate" header="KDV" style="width: 100px">
        <template #body="{ data }">
          <span v-if="editingCategoryId !== data.id" class="text-slate-500">%{{ parseFloat(data.default_kdv_rate) || 0 }}</span>
        </template>
      </Column>
      <Column header="İşlemler" bodyClass="text-center w-32" headerStyle="text-align: center;">
        <template #body="{ data }">
          <div v-if="editingCategoryId === data.id" class="flex gap-2">
            <Button icon="pi pi-check" class="p-button-sm rounded-md p-button-text" @click="handleUpdateCategory" title="Kaydet" severity="primary" />
            <Button icon="pi pi-times" class="p-button-sm rounded-md p-button-text" @click="editingCategoryId = null" title="İptal" severity="warn" />
          </div>
          <div v-else class="flex gap-2">
            <Button icon="pi pi-pencil" class="p-button-sm rounded-md p-button-text" @click="startEditCategory(data)" title="Düzenle" severity="warn" />
            <Button icon="pi pi-trash" class="p-button-sm rounded-md p-button-text" @click="handleDeleteCategory(data.id)" title="Sil" severity="danger" />
          </div>
        </template>
      </Column>
    </DataTable>
  </div>
</template>
