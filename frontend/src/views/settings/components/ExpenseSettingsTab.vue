<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'

const settingsStore = useSettingsStore()
const toast = useToast()

const expensePrefix = ref('EXP')

const loadSettings = async () => {
  try {
    const ePrefixVal = await settingsStore.fetchSetting('expense_prefix')
    if (ePrefixVal) expensePrefix.value = ePrefixVal
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Gider ayarları yüklenemedi.', life: 10000 })
  }
}

onMounted(() => {
  loadSettings()
})

const saveExpenseSettings = async () => {
  try {
    await settingsStore.saveSetting('expense_prefix', expensePrefix.value, 'expense')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Gider ayarları kaydedildi.', life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Gider ayarları kaydedilemedi.', life: 10000 })
  }
}
</script>

<template>
  <div class="grid grid-cols-2 gap-6 max-w-4xl pt-4">
    <div class="flex flex-col gap-1">
      <label class="font-semibold text-sm">Gider Fişi Ön Eki</label>
      <InputText v-model="expensePrefix" placeholder="örn: EXP" class="w-full" />
    </div>

    <div class="col-span-2 flex justify-end mt-4">
      <Button label="Kaydet" icon="pi pi-check" @click="saveExpenseSettings" :loading="settingsStore.loading" outlined severity="primary" />
    </div>
  </div>
</template>
