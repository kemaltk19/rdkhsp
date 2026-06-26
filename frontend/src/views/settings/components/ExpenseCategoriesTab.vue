<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useExpenseStore } from '@/stores/expense'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'

const expenseStore = useExpenseStore()
const toast = useToast()

const newCategoryName = ref('')
const editingCategoryId = ref<string | null>(null)
const editingCategoryName = ref('')

onMounted(async () => {
  await expenseStore.fetchCategories()
})

const handleCreateCategory = async () => {
  if (!newCategoryName.value.trim()) return
  try {
    await expenseStore.createCategory({ name: newCategoryName.value })
    newCategoryName.value = ''
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Kategori eklendi', life: 10000 })
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Kategori eklenemedi', life: 10000 })
  }
}

const startEditCategory = (cat: any) => {
  editingCategoryId.value = cat.id
  editingCategoryName.value = cat.name
}

const handleUpdateCategory = async () => {
  if (!editingCategoryId.value || !editingCategoryName.value.trim()) return
  try {
    await expenseStore.updateCategory(editingCategoryId.value, { name: editingCategoryName.value })
    editingCategoryId.value = null
    editingCategoryName.value = ''
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
      await expenseStore.deleteCategory(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Kategori silindi', life: 10000 })
    } catch (err: any) {
      if (err.response?.data?.error?.message) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Kategori silinemedi. Kullanımda olabilir.', life: 10000 })
      }
    }
  }
}
</script>

<template>
  <div class="expense-categories-tab">
    <div class="mb-4">
      <div class="flex gap-2 items-center mb-6">
        <InputText v-model="newCategoryName" placeholder="Yeni kategori adı (Örn: Yemek Giderleri)" class="w-full md:w-64" @keyup.enter="handleCreateCategory" />
        <Button label="Ekle" icon="pi pi-plus" @click="handleCreateCategory" :disabled="!newCategoryName.trim()" outlined severity="success" />
      </div>
    </div>

    <DataTable :value="expenseStore.categories" class="p-datatable-sm w-full" responsiveLayout="scroll">
      <Column field="name" header="Kategori Adı" sortable>
        <template #body="{ data }">
          <div v-if="editingCategoryId === data.id">
            <InputText v-model="editingCategoryName" class="w-full p-inputtext-sm" @keyup.enter="handleUpdateCategory" />
          </div>
          <span v-else>{{ data.name }}</span>
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
