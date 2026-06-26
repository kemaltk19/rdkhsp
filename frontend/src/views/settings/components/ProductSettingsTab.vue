<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Button from 'primevue/button'

const settingsStore = useSettingsStore()
const toast = useToast()

const productPrefix = ref('SKU')
const productKdv = ref('20')
const kdvOptions = ref<string[]>(['0', '1', '10', '20'])
const criticalStockThreshold = ref<number>(5)

const loadSettings = async () => {
  try {
    const kdvRatesStr = await settingsStore.fetchSetting('kdv_rates')
    if (kdvRatesStr) {
      try {
        const parsed = JSON.parse(kdvRatesStr)
        if (Array.isArray(parsed)) {
          kdvOptions.value = parsed.map(String)
        }
      } catch (e) {}
    }

    const pPrefixVal = await settingsStore.fetchSetting('product_prefix')
    if (pPrefixVal) productPrefix.value = pPrefixVal

    const pKdvVal = await settingsStore.fetchSetting('product_default_kdv')
    if (pKdvVal) productKdv.value = pKdvVal

    const criticalStockVal = await settingsStore.fetchSetting('critical_stock_threshold')
    if (criticalStockVal) {
      criticalStockThreshold.value = parseInt(criticalStockVal, 10)
    }
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Ürün ayarları yüklenemedi.', life: 10000 })
  }
}

onMounted(() => {
  loadSettings()
})

const saveProductSettings = async () => {
  try {
    await settingsStore.saveSetting('product_prefix', productPrefix.value, 'product')
    await settingsStore.saveSetting('product_default_kdv', productKdv.value, 'product')
    await settingsStore.saveSetting('critical_stock_threshold', criticalStockThreshold.value.toString(), 'product')
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Ürün ayarları kaydedildi.', life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Ürün ayarları kaydedilemedi.', life: 10000 })
  }
}
</script>

<template>
  <div class="flex flex-col gap-4 w-full">
    <div class="flex flex-col gap-1">
      <label class="font-bold text-[14px] text-slate-800 dark:text-slate-200">Ürün/Stok Ön Eki</label>
      <InputText v-model="productPrefix" placeholder="örn: SKU" class="w-full" />
    </div>

    <div class="flex flex-col gap-1">
      <label class="font-bold text-[14px] text-slate-800 dark:text-slate-200">Varsayılan KDV Oranı (%)</label>
      <Select
        v-model="productKdv"
        :options="kdvOptions"
        class="w-full"
      />
    </div>

    <div class="flex flex-col gap-1">
      <label class="font-bold text-[14px] text-slate-800 dark:text-slate-200">Kritik Stok Uyarısı Eşiği</label>
      <InputNumber
        v-model="criticalStockThreshold"
        inputId="criticalStock"
        :min="0"
        class="w-full"
      />
      <small class="text-slate-500 text-xs">Stok miktarı bu değerin altına düştüğünde sistem uyarı verecektir.</small>
    </div>

    <div class="flex justify-end mt-6">
      <Button label="Kaydet" icon="pi pi-check" class="w-full sm:w-auto" @click="saveProductSettings" :loading="settingsStore.loading" outlined severity="primary" />
    </div>
  </div>
</template>
