<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoleStore } from '@/stores/role'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Checkbox from 'primevue/checkbox'

const roleStore = useRoleStore()
const toast = useToast()

const MODULES = [
  { key: 'caris', label: 'Cariler' },
  { key: 'invoices', label: 'Faturalar' },
  { key: 'payments', label: 'Tahsilat/Ödeme' },
  { key: 'expenses', label: 'Giderler' },
  { key: 'products', label: 'Ürünler' },
  { key: 'reports', label: 'Raporlar' },
]

const showDialog = ref(false)
const editingId = ref<string | null>(null)

const emptyPermissions = () => MODULES.map(m => ({
  module: m.key, can_create: false, can_read: false, can_update: false, can_delete: false
}))

const form = ref({
  name: '',
  description: '',
  permissions: emptyPermissions(),
})

const loadRoles = () => roleStore.fetchRoles()
onMounted(loadRoles)

const openNew = () => {
  editingId.value = null
  form.value = { name: '', description: '', permissions: emptyPermissions() }
  showDialog.value = true
}

const openEdit = (role: any) => {
  editingId.value = role.id
  form.value = {
    name: role.name,
    description: role.description || '',
    permissions: MODULES.map(m => {
      const existing = role.permissions?.find((p: any) => p.module === m.key)
      return existing
        ? { module: m.key, can_create: existing.can_create, can_read: existing.can_read, can_update: existing.can_update, can_delete: existing.can_delete }
        : { module: m.key, can_create: false, can_read: false, can_update: false, can_delete: false }
    }),
  }
  showDialog.value = true
}

const save = async () => {
  if (!form.value.name.trim()) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Rol adı zorunludur.', life: 10000 })
    return
  }
  try {
    if (editingId.value) {
      await roleStore.updateRole(editingId.value, form.value)
    } else {
      await roleStore.createRole(form.value)
    }
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Rol kaydedildi.', life: 10000 })
    showDialog.value = false
    loadRoles()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'İşlem başarısız.', life: 10000 })
  }
}

const confirmDelete = async (id: string) => {
  if (confirm('Bu rolü silmek istediğinize emin misiniz?')) {
    try {
      await roleStore.deleteRole(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Rol silindi.', life: 10000 })
      loadRoles()
    } catch (err: any) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Silme başarısız.', life: 10000 })
    }
  }
}
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-lg font-medium">Roller ve Yetkiler</h3>
      <Button label="Yeni Rol" icon="pi pi-plus" class="p-button-sm" @click="openNew" outlined severity="success" />
    </div>

    <DataTable :value="roleStore.roles" :loading="roleStore.loading" class="p-datatable-sm w-full">
      <Column field="name" header="Rol Adı" />
      <Column field="description" header="Açıklama" />
      <Column header="İşlemler" bodyClass="text-center w-24">
        <template #body="{ data }">
          <Button icon="pi pi-pencil" class="p-button-text p-button-sm" @click="openEdit(data)" severity="warn" />
          <Button icon="pi pi-trash" class="p-button-text p-button-sm" @click="confirmDelete(data.id)" severity="danger" />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="showDialog" :header="editingId ? 'Rolü Düzenle' : 'Yeni Rol'" modal style="width: 600px">
      <div class="flex flex-col gap-4">
        <div class="flex flex-col gap-1">
          <label class="font-semibold text-sm">Rol Adı</label>
          <InputText v-model="form.name" placeholder="örn: Satış Personeli" class="w-full" />
        </div>
        <div class="flex flex-col gap-1">
          <label class="font-semibold text-sm">Açıklama (opsiyonel)</label>
          <InputText v-model="form.description" class="w-full" />
        </div>

        <table class="permission-matrix w-full text-sm">
          <thead>
            <tr>
              <th class="text-left">Modül</th>
              <th>Oluştur</th>
              <th>Görüntüle</th>
              <th>Düzenle</th>
              <th>Sil</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(perm, idx) in form.permissions" :key="perm.module">
              <td>{{ MODULES.find(m => m.key === perm.module)?.label }}</td>
              <td class="text-center"><Checkbox v-model="form.permissions[idx].can_create" binary /></td>
              <td class="text-center"><Checkbox v-model="form.permissions[idx].can_read" binary /></td>
              <td class="text-center"><Checkbox v-model="form.permissions[idx].can_update" binary /></td>
              <td class="text-center"><Checkbox v-model="form.permissions[idx].can_delete" binary /></td>
            </tr>
          </tbody>
        </table>
      </div>
      <template #footer>
        <Button label="İptal" class="p-button-text" @click="showDialog = false" />
        <Button label="Kaydet" @click="save" :loading="roleStore.loading" severity="primary" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.permission-matrix th, .permission-matrix td {
  padding: 0.5rem;
  border-bottom: 1px solid var(--p-content-border-color, #e2e8f0);
}
</style>
