<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useProductStore } from '@/stores/product'
import { useCurrencyStore } from '@/stores/currency'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Message from 'primevue/message'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
  productId: {
    type: String,
    default: null,
  },
})

const emit = defineEmits(['update:visible', 'saved'])

const productStore = useProductStore()
const currencyStore = useCurrencyStore()
const settingsStore = useSettingsStore()
const toast = useToast()

const loading = ref(false)
const errorMsg = ref('')

// Firmanın varsayılan dövizi (alış fiyatı/maliyet hep bu dövizde tutulur)
const defaultCurrencyCode = computed(() =>
  settingsStore.company?.currency || currencyStore.defaultCurrency?.code || 'TRY'
)

const form = ref({
  code: '',
  name: '',
  brand: '',
  description: '',
  type: 'product' as 'product' | 'service',
  unit: 'Adet',
  barcode: '',
  custom_codes: '',
  serial_numbers: '',
  purchase_price: 0,
  sale_price: 0,
  currency: 'TRY',
  tax_included: false,
  purchase_tax_included: false,
  tax_rate: 20, // default 20% (satış KDV oranı)
  purchase_tax_rate: 20, // alış KDV oranı
  track_stock: true,
  min_stock: 5,
  initial_stock: 0,
  current_stock: 0,
  category_id: null as string | null,
  is_active: true,
})

const typeOptions = ref([
  { label: 'Stoklu Ürün', value: 'product' },
  { label: 'Hizmet (Stoksuz)', value: 'service' },
])

const taxRateOptions = ref([
  { label: '%0', value: 0 },
  { label: '%1', value: 1 },
  { label: '%10', value: 10 },
  { label: '%20', value: 20 },
])

const unitOptions = ref([
  { label: 'Adet', value: 'Adet' },
  { label: 'Kg', value: 'Kg' },
  { label: 'Gram', value: 'Gram' },
  { label: 'Litre', value: 'Litre' },
  { label: 'Mililitre', value: 'Mililitre' },
  { label: 'Kutu', value: 'Kutu' },
  { label: 'Paket', value: 'Paket' },
  { label: 'Çuval', value: 'Çuval' },
  { label: 'Metre', value: 'Metre' },
  { label: 'Santimetre', value: 'Santimetre' },
  { label: 'M2 (Metrekare)', value: 'M2' },
  { label: 'M3 (Metreküp)', value: 'M3' },
  { label: 'Koli', value: 'Koli' },
  { label: 'Palet', value: 'Palet' },
  { label: 'Saat', value: 'Saat' },
  { label: 'Gün', value: 'Gün' },
  { label: 'Ay', value: 'Ay' },
  { label: 'Yıl', value: 'Yıl' },
  { label: 'Seans', value: 'Seans' },
  { label: 'Takım', value: 'Takım' },
])

onMounted(async () => {
  loading.value = true
  try {
    await productStore.fetchCategories()
    if (!settingsStore.company) await settingsStore.fetchCompanyProfile()
    if (currencyStore.currencies.length === 0) await currencyStore.fetchCurrencies()

    const kdvRatesStr = await settingsStore.fetchSetting('kdv_rates')
    if (kdvRatesStr) {
      try {
        const parsed = JSON.parse(kdvRatesStr)
        if (Array.isArray(parsed)) {
          taxRateOptions.value = parsed.map((r: number) => ({ label: `%${r}`, value: r }))
        }
      } catch (e) {}
    }

    if (!props.productId) {
      const pKdvVal = await settingsStore.fetchSetting('product_default_kdv')
      if (pKdvVal) form.value.tax_rate = parseFloat(pKdvVal)

      if (currencyStore.currencies.length === 0) {
        await currencyStore.fetchCurrencies()
      }
      form.value.currency = currencyStore.defaultCurrency?.code || 'TRY'
      // Kod alanı boş bırakılır: gerçek kod kaydetme anında backend'de üretilir.
      // Önceden burada bir "önizleme" kodu çekilip gösteriliyordu; bu önizleme
      // kaydetme anına kadar başka bir ürün tarafından alınabildiğinden
      // "kod zaten kullanımda" çakışmasına yol açıyordu.
    }

    if (props.productId) {
      if (currencyStore.currencies.length === 0) {
        await currencyStore.fetchCurrencies()
      }
      const data = await productStore.getProductByID(props.productId)
      form.value = {
        code: data.code,
        name: data.name,
        brand: data.brand || '',
        description: data.description || '',
        type: data.type,
        unit: data.unit || 'Adet',
        barcode: data.barcode || '',
        custom_codes: data.custom_codes || '',
        serial_numbers: data.serial_numbers || '',
        purchase_price: parseFloat(data.purchase_price) || 0,
        sale_price: parseFloat(data.sale_price) || 0,
        currency: data.currency || 'TRY',
        tax_included: data.tax_included || false,
        purchase_tax_included: data.purchase_tax_included || false,
        tax_rate: parseFloat(data.tax_rate) || 0,
        purchase_tax_rate: parseFloat(data.purchase_tax_rate) || 0,
        track_stock: data.track_stock,
        min_stock: parseFloat(data.min_stock) || 0,
        initial_stock: 0,
        current_stock: parseFloat(data.current_stock) || 0,
        category_id: data.category_id,
        is_active: data.is_active,
      }
    }
  } catch (err) {
    errorMsg.value = 'Ürün verileri yüklenemedi.'
  } finally {
    loading.value = false
  }
})

// Services don't track stock, auto-toggle track_stock to false if type is service
watch(() => form.value.type, (newType) => {
  if (newType === 'service') {
    form.value.track_stock = false
  } else {
    form.value.track_stock = true
  }
})

let kdvModifiedByUser = false

watch(() => form.value.tax_rate, (newVal, oldVal) => {
  // If user changes tax_rate after initial load, flag it
  if (oldVal !== undefined) {
    kdvModifiedByUser = true
  }
})

watch(() => form.value.category_id, (newCatId) => {
  if (newCatId) {
    const cat = productStore.categories.find(c => c.id === newCatId)
    if (cat && cat.default_kdv_rate) {
      if (!kdvModifiedByUser) {
        // We temporally disable the flag so this programmatic change doesn't count as user modification
        const oldFlag = kdvModifiedByUser
        form.value.tax_rate = parseFloat(cat.default_kdv_rate) || 0
        setTimeout(() => kdvModifiedByUser = oldFlag, 0)
      }
    }
  }
})

const categoryOptions = computed(() => {
  return productStore.categories.map(c => ({
    ...c,
    displayName: `${c.name} (%${parseFloat(c.default_kdv_rate) || 0})`
  }))
})

const close = () => {
  emit('update:visible', false)
}

const generateBarcode = () => {
  // EAN-13 format: 12 digits + 1 checksum digit. Using Turkey prefix (869)
  let base = '869'
  for (let i = 0; i < 9; i++) {
    base += Math.floor(Math.random() * 10).toString()
  }
  
  let sum = 0
  for (let i = 0; i < 12; i++) {
    sum += parseInt(base[i]) * (i % 2 === 0 ? 1 : 3)
  }
  
  const checksum = (10 - (sum % 10)) % 10
  form.value.barcode = base + checksum.toString()
}

const handleSubmit = async () => {
  if (!form.value.name) {
    errorMsg.value = 'Lütfen ürün adını girin.'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const payload = {
      ...form.value,
      // Yeni üründe kod alanı boş gönderilir; backend kaydetme anında üretir.
      code: props.productId ? form.value.code : '',
      purchase_price: form.value.purchase_price.toString(),
      sale_price: form.value.sale_price.toString(),
      tax_rate: form.value.tax_rate.toString(),
      purchase_tax_rate: form.value.purchase_tax_rate.toString(),
      min_stock: form.value.min_stock.toString(),
      initial_stock: form.value.initial_stock.toString(),
    }

    if (props.productId) {
      await productStore.updateProduct(props.productId, payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Ürün/Hizmet güncellendi', life: 10000 })
    } else {
      await productStore.createProduct(payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Ürün/Hizmet oluşturuldu', life: 10000 })
    }

    emit('saved')
    close()
  } catch (err: any) {
    if (err.response && err.response.data && err.response.data.error) {
      errorMsg.value = err.response.data.error.message
    } else {
      errorMsg.value = 'İşlem gerçekleştirilemedi.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Dialog
    :visible="props.visible"
    @update:visible="close"
    :modal="true"
    :style="{ width: '95%', maxWidth: '920px' }"
  >
    <template #header>
      <div class="flex justify-between items-center w-full pr-8">
        <span class="text-lg font-bold text-slate-800 dark:text-white">{{ props.productId ? 'Ürün/Hizmet Düzenle' : 'Yeni Ürün/Hizmet Tanımla' }}</span>
        <div class="flex items-center gap-2">
          <span class="text-sm font-semibold text-slate-600 dark:text-slate-400">Ürün Kodu:</span>
          <InputText v-model="form.code" placeholder="Otomatik Üretilecek" disabled class="w-44 font-semibold bg-slate-50 dark:bg-slate-800 text-left" style="padding: 0.35rem 0.75rem; font-size: 0.875rem;" />
        </div>
      </div>
    </template>
    
    <Message v-if="errorMsg" severity="error" class="mb-4">{{ errorMsg }}</Message>

    <form @submit.prevent="handleSubmit">
      <div class="cari-form grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Left Column: Genel Bilgiler -->
        <div class="form-col flex flex-col gap-3">
          <div class="section-title">Genel Bilgiler</div>
          
          <div class="grid grid-cols-2 gap-3">
            <div class="frow-vert">
              <label for="type">Tür *</label>
              <Select
                id="type"
                v-model="form.type"
                :options="typeOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                :disabled="loading"
                size="small"
              />
            </div>
            <div class="frow-vert">
              <label for="category">Kategori</label>
              <Select
                id="category"
                v-model="form.category_id"
                :options="productStore.categories"
                optionLabel="name"
                optionValue="id"
                placeholder="Kategori seçin..."
                class="w-full"
                filter
                showClear
                :disabled="loading"
                size="small"
              />
            </div>
          </div>

          <div class="frow-vert">
            <label for="name">Ürün/Hizmet Adı *</label>
            <InputText
              id="name"
              v-model="form.name"
              placeholder="Ör: Asus Laptop, Danışmanlık Hizmeti"
              class="w-full p-inputtext-sm"
              maxlength="255"
              required
              :disabled="loading"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="frow-vert">
              <label for="brand">Marka</label>
              <InputText
                id="brand"
                v-model="form.brand"
                placeholder="Ör: Apple, Samsung"
                class="w-full p-inputtext-sm"
                maxlength="100"
                :disabled="loading"
              />
            </div>
            <div class="frow-vert">
              <label for="unit">Ölçü Birimi</label>
              <Select
                id="unit"
                v-model="form.unit"
                :options="unitOptions"
                optionLabel="label"
                optionValue="value"
                placeholder="Seçiniz..."
                class="w-full"
                filter
                :disabled="loading"
                size="small"
              />
            </div>
          </div>

          <div class="frow-vert">
            <label for="barcode">Barkod No</label>
            <div class="flex">
              <InputText
                id="barcode"
                v-model="form.barcode"
                placeholder="Ör: 8691234567890"
                class="w-full p-inputtext-sm"
                style="border-top-right-radius: 0; border-bottom-right-radius: 0;"
                maxlength="100"
                :disabled="loading"
              />
              <Button icon="pi pi-plus" @click="generateBarcode" :disabled="loading" style="border-top-left-radius: 0; border-bottom-left-radius: 0; padding-left: 0.75rem; padding-right: 0.75rem;" title="Otomatik Barkod (EAN-13) Üret" outlined severity="success" />
            </div>
          </div>

          <div class="frow-vert">
            <label for="description">Açıklama</label>
            <Textarea
              id="description"
              v-model="form.description"
              placeholder="Ürün veya hizmetin detayı"
              class="w-full p-textarea-sm"
              rows="1"
              maxlength="2000"
              :disabled="loading"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="frow-vert">
              <label for="custom_codes">Özel Kodlar</label>
              <InputText
                id="custom_codes"
                v-model="form.custom_codes"
                placeholder="Ör: 125468, 55465"
                class="w-full p-inputtext-sm"
                maxlength="255"
                :disabled="loading"
              />
            </div>
            <div class="frow-vert">
              <label for="serial_numbers">Seri No</label>
              <InputText
                id="serial_numbers"
                v-model="form.serial_numbers"
                placeholder="Ör: SN123, SN456"
                class="w-full p-inputtext-sm"
                maxlength="2000"
                :disabled="loading"
              />
            </div>
          </div>
        </div>

        <!-- Right Column: Fiyat, Vergi & Stok Bilgileri -->
        <div class="form-col flex flex-col gap-3">
          <div class="section-title">Fiyat & Vergi Bilgileri</div>

          <!-- Alış Fiyatı her zaman firmanın varsayılan dövizinde girilir
               (ürünü USD alıp TL ödesen bile ödediğin TL tutarını yazarsın);
               böylece ortalama maliyet doğrudan varsayılan dövizde (TL) oluşur. -->
          <div class="grid grid-cols-3 gap-3">
            <div class="frow-vert">
              <label for="purchase_price">Alış Fiyatı ({{ defaultCurrencyCode }})</label>
              <InputNumber
                id="purchase_price"
                v-model="form.purchase_price"
                class="w-full"
                mode="decimal"
                :minFractionDigits="2"
                :maxFractionDigits="4"
                :disabled="loading"
                size="small"
              />
            </div>
            <div class="frow-vert">
              <label for="sale_price">Satış Fiyatı</label>
              <InputNumber
                id="sale_price"
                v-model="form.sale_price"
                class="w-full"
                mode="decimal"
                :minFractionDigits="2"
                :maxFractionDigits="4"
                :disabled="loading"
                size="small"
              />
            </div>
            <div class="frow-vert">
              <label for="currency">Satış Para Birimi</label>
              <Select
                id="currency"
                v-model="form.currency"
                :options="currencyStore.currencies"
                optionLabel="code"
                optionValue="code"
                class="w-full"
                :disabled="loading"
                size="small"
              />
            </div>
          </div>

          <!-- Alış / Satış KDV dahil mi seçimleri, ilgili fiyatın altında hizalı -->
          <div class="grid grid-cols-3 gap-3">
            <div class="flex items-center gap-2">
              <input type="checkbox" id="purchase_tax_included" v-model="form.purchase_tax_included" class="w-4 h-4 cursor-pointer" :disabled="loading" />
              <label for="purchase_tax_included" class="text-xs font-semibold cursor-pointer text-slate-600 dark:text-slate-400">Alış Fiyatı KDV Dahil</label>
            </div>
            <div class="flex items-center gap-2">
              <input type="checkbox" id="tax_included" v-model="form.tax_included" class="w-4 h-4 cursor-pointer" :disabled="loading" />
              <label for="tax_included" class="text-xs font-semibold cursor-pointer text-slate-600 dark:text-slate-400">Satış Fiyatı KDV Dahil</label>
            </div>
          </div>

          <div class="grid grid-cols-3 gap-3 items-end">
            <div class="frow-vert">
              <label for="purchase_tax_rate">Alış KDV Oranı</label>
              <Select
                id="purchase_tax_rate"
                v-model="form.purchase_tax_rate"
                :options="taxRateOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                :disabled="loading"
                size="small"
              />
            </div>
            <div class="frow-vert">
              <label for="tax_rate">Satış KDV Oranı *</label>
              <Select
                id="tax_rate"
                v-model="form.tax_rate"
                :options="taxRateOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                :disabled="loading"
                size="small"
              />
            </div>
          </div>

          <!-- STOK BİLGİLERİ -->
          <template v-if="form.type === 'product'">
            <div class="section-title mt-2">Stok Bilgileri</div>
            
            <div class="grid grid-cols-2 gap-3">
              <div class="frow-vert">
                <label for="min_stock">Asgari Stok Limiti</label>
                <InputNumber
                  id="min_stock"
                  v-model="form.min_stock"
                  class="w-full"
                  mode="decimal"
                  :minFractionDigits="0"
                  :maxFractionDigits="4"
                  :disabled="loading"
                  size="small"
                />
              </div>
              
              <div class="frow-vert" v-if="!props.productId">
                <label for="initial_stock">Açılış Stoğu</label>
                <InputNumber
                  id="initial_stock"
                  v-model="form.initial_stock"
                  class="w-full"
                  mode="decimal"
                  :minFractionDigits="0"
                  :maxFractionDigits="4"
                  :disabled="loading"
                  size="small"
                />
              </div>
              <div class="frow-vert" v-if="props.productId">
                <label for="current_stock">Güncel Stok Adedi</label>
                <InputNumber
                  id="current_stock"
                  v-model="form.current_stock"
                  class="w-full"
                  mode="decimal"
                  disabled
                  size="small"
                />
              </div>
            </div>

            <div class="flex items-center gap-2 mt-1">
              <input type="checkbox" id="track_stock" v-model="form.track_stock" class="w-4 h-4 cursor-pointer" :disabled="loading" />
              <label for="track_stock" class="text-xs font-semibold cursor-pointer text-slate-600 dark:text-slate-400">Stok Takibi Yapılsın</label>
            </div>
          </template>

          <!-- AKTİF Mİ -->
          <div class="pt-3 mt-auto border-t border-slate-200 dark:border-slate-800 flex items-center gap-3">
            <input type="checkbox" id="is_active" v-model="form.is_active" class="w-4 h-4 cursor-pointer" :disabled="loading" />
            <label for="is_active" class="text-sm font-semibold cursor-pointer text-green-600 dark:text-green-400">Bu Ürün/Hizmet Aktif</label>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="footer-buttons">
        <Button label="İptal" icon="pi pi-times" class="p-button-text" @click="close" :disabled="loading" outlined />
        <Button label="Kaydet" icon="pi pi-check" @click="handleSubmit" :loading="loading" outlined severity="primary" />
      </div>
    </template>
  </Dialog>
</template>

<style scoped>
.cari-form {
  padding: 0.5rem 0;
}

.section-title {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--p-primary-500, #0891b2);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--p-content-border-color, #e2e8f0);
  padding-bottom: 0.25rem;
  margin-bottom: 0.25rem;
}

:root.p-dark .section-title {
  border-color: #334155;
}

.frow-vert {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.frow-vert label {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--p-text-muted-color, #64748b);
}

.frow-vert label em {
  color: #ef4444;
  font-style: normal;
  margin-left: 2px;
}

.footer-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
