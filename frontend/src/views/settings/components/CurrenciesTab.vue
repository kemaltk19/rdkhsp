<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h3 class="text-xl font-bold dark:text-white">Para Birimleri</h3>
      <Button label="Yeni Para Birimi" icon="pi pi-plus" size="small" @click="openModal()" severity="success" />
    </div>

    <DataTable :value="currencyStore.currencies" :loading="currencyStore.loading" class="p-datatable-sm">
      <Column field="name" header="Ad" />
      <Column field="code" header="Kod" />
      <Column field="symbol" header="Sembol" />
      <Column header="Varsayılan">
        <template #body="slotProps">
          <Tag v-if="slotProps.data.is_default" severity="success" value="Evet" />
          <span v-else class="text-gray-400">-</span>
        </template>
      </Column>
      <Column header="Kur">
        <template #body="slotProps">
          <span v-if="!slotProps.data.is_default" class="font-mono text-sm">
            {{ slotProps.data.exchange_rate_op === '/' ? '÷' : '×' }} {{ slotProps.data.exchange_rate }}
          </span>
          <span v-else class="text-gray-400">-</span>
        </template>
      </Column>
      <Column header="Format (Örnek)">
        <template #body="slotProps">
          {{ formatPreview(slotProps.data) }}
        </template>
      </Column>
      <Column :exportable="false" style="min-width:8rem" alignFrozen="right" :frozen="true">
        <template #body="slotProps">
          <div class="flex gap-2 justify-end">
            <Button icon="pi pi-pencil" outlined rounded size="small" @click="openModal(slotProps.data)" severity="warn" />
            <Button icon="pi pi-trash" outlined rounded size="small" @click="confirmDelete(slotProps.data)" :disabled="slotProps.data.is_default" severity="danger" />
          </div>
        </template>
      </Column>
      <template #empty>
        <div class="text-center p-4 text-gray-500">
          Kayıtlı para birimi bulunamadı.
        </div>
      </template>
    </DataTable>

    <Dialog v-model:visible="modalVisible" :header="editingId ? 'Para Birimi Düzenle' : 'Yeni Para Birimi Ekle'" :modal="true" :style="{ width: '95vw', maxWidth: '800px' }">
      <div class="flex flex-col gap-4 mt-2">
        <div class="flex flex-col gap-2">
          <label class="font-medium">Para Birimi Adı</label>
          <InputText v-model="form.name" placeholder="Örn: Türk Lirası" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="font-medium">Kod</label>
            <InputText v-model="form.code" placeholder="Örn: TRY" />
          </div>
          <div class="flex flex-col gap-2">
            <label class="font-medium">Sembol</label>
            <InputText v-model="form.symbol" placeholder="Örn: ₺" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="font-medium">Kur (Varsayılan Dövize Karşılık)</label>
            <InputNumber v-model="form.exchange_rate" mode="decimal" :minFractionDigits="2" :maxFractionDigits="6" :disabled="form.is_default" />
          </div>
          <div class="flex flex-col gap-2">
            <label class="font-medium">İşlem İşareti</label>
            <Select
              v-model="form.exchange_rate_op"
              :options="[{ label: 'Çarp (×)', value: '*' }, { label: 'Böl (÷)', value: '/' }]"
              optionLabel="label"
              optionValue="value"
              :disabled="form.is_default"
            />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="font-medium">Binlik Ayracı</label>
            <InputText v-model="form.format_thousand_sep" placeholder="Örn: ." />
          </div>
          <div class="flex flex-col gap-2">
            <label class="font-medium">Ondalık Ayracı</label>
            <InputText v-model="form.format_decimal_sep" placeholder="Örn: ," />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="font-medium">Kuruş Hanesi</label>
            <InputNumber v-model="form.format_decimals" :min="0" :max="8" />
          </div>
          <div class="flex flex-col gap-2">
            <label class="font-medium">Sembol Konumu</label>
            <Select v-model="form.format_position" :options="positionOptions" optionLabel="label" optionValue="value" />
          </div>
        </div>
        <div class="flex items-center gap-2 mt-2">
          <Checkbox v-model="form.is_default" inputId="isDef" :binary="true" />
          <label for="isDef">Varsayılan Para Birimi Yap</label>
        </div>
      </div>
      <template #footer>
        <Button label="İptal" icon="pi pi-times" text @click="modalVisible = false" />
        <Button label="Kaydet" icon="pi pi-check" @click="save" :loading="saving" severity="primary" />
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCurrencyStore, type Currency } from '@/stores/currency'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'

import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Dialog from 'primevue/dialog'
import Tag from 'primevue/tag'
import Checkbox from 'primevue/checkbox'
import Select from 'primevue/select'

const currencyStore = useCurrencyStore()
const confirm = useConfirm()
const toast = useToast()

const modalVisible = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)

const positionOptions = ref([
  { label: 'Solda (₺ 100)', value: 'LeftSpace' },
  { label: 'Solda Bitişik (₺100)', value: 'Left' },
  { label: 'Sağda (100 ₺)', value: 'RightSpace' },
  { label: 'Sağda Bitişik (100₺)', value: 'Right' },
])

const form = ref<Partial<Currency>>({
  name: '',
  code: '',
  symbol: '',
  exchange_rate: 1,
  exchange_rate_op: '*',
  format_thousand_sep: '.',
  format_decimal_sep: ',',
  format_decimals: 2,
  format_position: 'RightSpace',
  is_default: false
})

const formatPreview = (curr: Currency) => {
  if (!curr) return ''
  const num = 1000
  const decimals = curr.format_decimals !== undefined ? curr.format_decimals : 2
  const decimalSep = curr.format_decimal_sep || ','
  const thousandSep = curr.format_thousand_sep || '.'
  
  const valStr = num.toFixed(decimals).replace('.', decimalSep)
  const parts = valStr.split(decimalSep)
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, thousandSep)
  const formattedVal = parts.join(decimalSep)
  
  const symbol = curr.symbol || ''
  
  switch(curr.format_position) {
    case 'Left': return `${symbol}${formattedVal}`
    case 'LeftSpace': return `${symbol} ${formattedVal}`
    case 'Right': return `${formattedVal}${symbol}`
    case 'RightSpace': return `${formattedVal} ${symbol}`
    default: return `${formattedVal} ${symbol}`
  }
}

const openModal = (curr?: Currency) => {
  if (curr) {
    editingId.value = curr.id
    form.value = { ...curr }
  } else {
    editingId.value = null
    form.value = {
      name: '',
      code: '',
      symbol: '',
      exchange_rate: 1,
      exchange_rate_op: '*',
      format_thousand_sep: '.',
      format_decimal_sep: ',',
      format_decimals: 2,
      format_position: 'RightSpace',
      is_default: false
    }
  }
  modalVisible.value = true
}

const save = async () => {
  if (!form.value.name || !form.value.code) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Ad ve Kod zorunludur.', life: 10000 })
    return
  }
  
  saving.value = true
  try {
    if (editingId.value) {
      await currencyStore.updateCurrency(editingId.value, form.value)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Para birimi güncellendi', life: 10000 })
    } else {
      await currencyStore.createCurrency(form.value)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Para birimi eklendi', life: 10000 })
    }
    modalVisible.value = false
  } catch (e: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: e.message, life: 10000 })
  } finally {
    saving.value = false
  }
}

const confirmDelete = (curr: Currency) => {
  confirm.require({
    message: `${curr.name} para birimini silmek istediğinize emin misiniz?`,
    header: 'Onay',
    icon: 'pi pi-exclamation-triangle',
    accept: async () => {
      try {
        await currencyStore.deleteCurrency(curr.id)
        toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Silindi', life: 10000 })
      } catch (e: any) {
        toast.add({ severity: 'error', summary: 'Hata', detail: e.message, life: 10000 })
      }
    }
  })
}
</script>
