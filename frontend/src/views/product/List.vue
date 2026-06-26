<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useProductStore } from '@/stores/product'
import type { Product } from '@/stores/product'
import { useNotificationStore } from '@/stores/notification'
import { useCurrencyStore } from '@/stores/currency'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Menu from 'primevue/menu'
import Money from '@/components/Money.vue'
import FormModal from './FormModal.vue'
import { exportToPDF } from '@/utils/pdfExport'

import { usePermission } from '@/composables/usePermission'

const productStore = useProductStore()
const notificationStore = useNotificationStore()
const currencyStore = useCurrencyStore()
const settingsStore = useSettingsStore()
const toast = useToast()
const { can } = usePermission()

const defaultCurrencyCode = computed(() => settingsStore.company?.currency || 'TRY')

// Ürünün dövizini, Para Birimleri ekranındaki güncel kur+işaretle varsayılan
// dövize çevirir. Sadece görsel: ürünün kendi kayıtlı değeri/dövizi değişmez.
const convertToDefault = (amount: number, currency: string): number | null => {
  if (!currency || currency === defaultCurrencyCode.value) return null
  const c = currencyStore.getCurrencyByCode(currency)
  if (!c || !c.exchange_rate) return null
  return c.exchange_rate_op === '/' ? amount / c.exchange_rate : amount * c.exchange_rate
}

const showProductModal = ref(false)
const selectedProductId = ref<string | undefined>(undefined)

const showAdjustmentModal = ref(false)
const showMovementsModal = ref(false)

// Active entities for modals
const activeMovementProduct = ref<Product | null>(null)

// Active entities for modals

// Adjustment form state
const adjustmentForm = ref({
  product_id: '',
  warehouse_id: '',
  type: 'in' as 'in' | 'out',
  quantity: 1,
  unit_cost: 0,
  note: '',
  date: new Date()
})

// Search & filters
const searchQuery = ref('')
const selectedType = ref('')
const selectedCategory = ref('')
const first = ref(0)
const rows = ref(20)
const page = ref(1)
const dt = ref()
const selectedItems = ref([])
const sortField = ref('')
const sortOrder = ref(1)

const typeOptions = ref([
  { label: 'Tüm Türler', value: '' },
  { label: 'Stoklu Ürün', value: 'product' },
  { label: 'Hizmet (Stoksuz)', value: 'service' },
])

const loadData = async () => {
  const params: any = {
    page: page.value,
    limit: rows.value,
    q: searchQuery.value,
    type: selectedType.value,
    category_id: selectedCategory.value,
    sort: sortField.value ? `${sortField.value} ${sortOrder.value === 1 ? 'asc' : 'desc'}` : ''
  }
  await productStore.fetchProducts(params)
}

const onSort = (event: any) => {
  sortField.value = event.sortField
  sortOrder.value = event.sortOrder
  loadData()
}

onMounted(async () => {
  await loadData()
  await productStore.fetchCategories()
  await productStore.fetchWarehouses()
  await currencyStore.fetchCurrencies()
  if (!settingsStore.company) await settingsStore.fetchCompanyProfile()
})

let searchDebounceTimeout: any = null
watch(searchQuery, () => {
  if (searchDebounceTimeout) clearTimeout(searchDebounceTimeout)
  searchDebounceTimeout = setTimeout(() => {
    page.value = 1
    first.value = 0
    loadData()
  }, 300)
})

const onFilterChange = () => {
  page.value = 1
  first.value = 0
  loadData()
}

const onPage = (event: any) => {
  page.value = event.page + 1
  rows.value = event.rows
  first.value = event.first
  loadData()
}

// Product Action Handlers
const openNewProduct = () => {
  selectedProductId.value = undefined
  showProductModal.value = true
}

const editProduct = (id: string) => {
  selectedProductId.value = id
  showProductModal.value = true
}

const deleteProductItem = async (id: string) => {
  if (confirm('Bu ürünü/hizmeti silmek istediğinize emin misiniz?')) {
    try {
      await productStore.deleteProduct(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Ürün silindi', life: 10000 })
      await loadData()
      notificationStore.fetchNotifications()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'İşlem gerçekleştirilemedi', life: 10000 })
      }
    }
  }
}

// Product Action Handlers

// Stock Adjustment Handlers
const openAdjustment = (product: Product | null = null) => {
  adjustmentForm.value = {
    product_id: product ? product.id : '',
    warehouse_id: productStore.warehouses.find(w => w.is_default)?.id || (productStore.warehouses[0]?.id || ''),
    type: 'in',
    quantity: 1,
    unit_cost: product ? parseFloat(product.average_cost) || 0 : 0,
    note: 'Stok düzeltme',
    date: new Date()
  }
  showAdjustmentModal.value = true
}

const handleCreateAdjustment = async () => {
  if (!adjustmentForm.value.product_id || !adjustmentForm.value.warehouse_id) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Ürün ve depo alanları zorunludur.', life: 10000 })
    return
  }
  try {
    const payload = {
      ...adjustmentForm.value,
      date: adjustmentForm.value.date ? new Date(adjustmentForm.value.date).toISOString() : new Date().toISOString(),
      quantity: adjustmentForm.value.quantity.toString(),
      unit_cost: adjustmentForm.value.unit_cost.toString()
    }
    await productStore.createStockMovement(payload)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Stok hareketi eklendi', life: 10000 })
    showAdjustmentModal.value = false
    await loadData()
    notificationStore.fetchNotifications()
  } catch (err: any) {
    if (err.response && err.response.data && err.response.data.error) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Stok hareketi eklenemedi.', life: 10000 })
    }
  }
}

// Movements Log Dialog
const viewMovements = async (product: Product) => {
  activeMovementProduct.value = product
  try {
    await productStore.fetchProductMovements(product.id)
    showMovementsModal.value = true
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Stok hareketleri yüklenemedi', life: 10000 })
  }
}

// Formatting helpers
const truncate = (val: any) => {
  if (val === undefined || val === null) return ''
  const str = String(val)
  return str.length > 15 ? str.substring(0, 15) + '...' : str
}

const getProductTypeLabel = (type: string) => {
  return type === 'product' ? 'Stoklu Ürün' : 'Hizmet (Stoksuz)'
}

const getStockStatus = (product: Product) => {
  if (!product.track_stock || product.type === 'service') {
    return { label: 'Takip Dışı', severity: 'secondary' }
  }
  const stock = parseFloat(product.current_stock) || 0
  const rawMin = parseFloat(product.min_stock) || 0
  const min = rawMin > 0 ? rawMin : 5 // backend GetCriticalStockProducts ile aynı varsayılan eşik
  if (stock <= 0) {
    return { label: `Tükendi (${stock} ${product.unit})`, severity: 'danger' }
  }
  if (stock <= min) {
    return { label: `Kritik (${stock} ${product.unit})`, severity: 'warn' }
  }
  return { label: `${stock} ${product.unit}`, severity: 'success' }
}

const getMovementSourceLabel = (src: string) => {
  switch (src) {
    case 'invoice': return 'Fatura'
    case 'manual': return 'Manuel Düzeltme'
    case 'transfer': return 'Depo Transferi'
    default: return src
  }
}

// Computed stats
const stats = computed(() => {
  const criticalCount = productStore.products.filter(p => {
    if (p.type === 'service' || !p.track_stock) return false
    const stock = parseFloat(p.current_stock) || 0
    const rawMin = parseFloat(p.min_stock) || 0
    const min = rawMin > 0 ? rawMin : 5
    return stock <= min
  }).length

  return {
    total: productStore.total,
    critical: criticalCount,
    categories: productStore.categories.length,
    warehouses: productStore.warehouses.length
  }
})

const exportMenu = ref()
const exportOptions = [
  { label: 'PDF', icon: 'pi pi-file-pdf', command: () => exportData('pdf') },
  { label: 'Excel', icon: 'pi pi-file-excel', command: () => exportData('excel') }
]

const toggleExportMenu = (event: any) => {
  exportMenu.value.toggle(event)
}

const exportData = (format: string) => {
  if (format === 'excel') {
    dt.value.exportCSV({ selectionOnly: selectedItems.value.length > 0 })
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Excel olarak dışa aktarıldı', life: 10000 })
  } else if (format === 'pdf') {
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : productStore.products
    const columns = [
      { header: 'Ürün Kodu', dataKey: 'code' },
      { header: 'Adı', dataKey: 'name' },
      { header: 'Tür', dataKey: 'type' },
      { header: 'Kategori', dataKey: 'category_name' },
      { header: 'Ort. Maliyet', dataKey: 'average_cost' },
      { header: 'Satış Fiyatı', dataKey: 'sale_price' },
      { header: 'Stok', dataKey: 'current_stock' },
      { header: 'Envanter Değeri', dataKey: 'inventory_value' }
    ]
    exportToPDF('Urun_Kartlari_Listesi', columns, dataToExport.map(item => ({
      ...item,
      type: getProductTypeLabel(item.type),
      category_name: item.category?.name || '-',
      current_stock: `${item.current_stock} ${item.unit}`,
      inventory_value: `${((Math.round(parseFloat(item.current_stock) * 10000) / 10000 || 0) * (Math.round(parseFloat(item.average_cost) * 100) / 100 || 0)).toFixed(2)} ${item.currency}`
    })))
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'PDF olarak dışa aktarıldı', life: 10000 })
  }
}
const exportCSV = () => {
  exportData('excel')
}
</script>

<template>
  <div class="product-list-container">

    <!-- Summary Metrics -->
    <div class="summary-cards">
      <Card class="metric-card">
        <template #content>
          <div class="metric-content">
            <span class="metric-label">Toplam Tanımlı Kart</span>
            <span class="metric-value blue">{{ stats.total }} Adet</span>
          </div>
        </template>
      </Card>

      <Card class="metric-card">
        <template #content>
          <div class="metric-content">
            <span class="metric-label">Kritik Stok Uyarıları</span>
            <span class="metric-value" :class="stats.critical > 0 ? 'red' : 'green'">{{ stats.critical }} Kart</span>
          </div>
        </template>
      </Card>

      <Card class="metric-card">
        <template #content>
          <div class="metric-content">
            <span class="metric-label">Aktif Depo Sayısı</span>
            <span class="metric-value text-slate-700 dark:text-slate-200">{{ stats.warehouses }} Depo</span>
          </div>
        </template>
      </Card>

      <Card class="metric-card">
        <template #content>
          <div class="metric-content">
            <span class="metric-label">Ürün Kategorileri</span>
            <span class="metric-value text-slate-700 dark:text-slate-200">{{ stats.categories }} Kategori</span>
          </div>
        </template>
      </Card>
    </div>

    <!-- Table Card -->
    <Card class="table-card">
      <template #content>
        <!-- Filters Header -->
        <div class="filters-header flex flex-col md:flex-row justify-between items-center gap-4 mb-4">
          <!-- Left: Filters -->
          <div class="filters-group flex gap-2 w-full md:w-auto">
            <Select
              v-model="selectedType"
              :options="typeOptions"
              optionLabel="label"
              optionValue="value"
              placeholder="Tür"
              class="filter-select w-full md:w-40"
              @change="onFilterChange"
            />
            <Select
              v-model="selectedCategory"
              :options="[{ id: '', name: 'Tüm Kategoriler' }, ...productStore.categories]"
              optionLabel="name"
              optionValue="id"
              placeholder="Kategori"
              class="filter-select w-full md:w-48"
              filter
              @change="onFilterChange"
            />
          </div>

          <!-- Middle: Search -->
          <div class="flex-1 flex justify-center w-full md:w-auto">
            <div class="search-input w-full max-w-md relative">
              <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"></i>
              <InputText v-model="searchQuery" placeholder="Ürün adı, kod veya barkod ile ara..." class="w-full pl-10" />
            </div>
          </div>

          <!-- Right: Buttons -->
          <div class="flex gap-2 w-full md:w-auto justify-end shrink-0">
            <Button v-if="can('products', 'create')" label="Ekle" icon="pi pi-plus" @click="openNewProduct" severity="success" />
            <Button v-if="can('products', 'update')" label="Düzelt" icon="pi pi-sliders-h" class="p-button-outlined p-button-warning" @click="openAdjustment(null)" />
            <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
            <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
          </div>
        </div>

        <DataTable
          ref="dt"
          :value="productStore.products"
          v-model:selection="selectedItems"
          lazy
          paginator
          :first="first"
          :rows="rows"
          :rowsPerPageOptions="[20, 50, 100]"
          :totalRecords="productStore.total"
          :loading="productStore.loading"
          @page="onPage"
          @sort="onSort"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          dataKey="id"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ (page - 1) * rows + slotProps.index + 1 }}
            </template>
          </Column>
          <Column selectionMode="multiple" headerStyle="width: 3rem"></Column>
          
          <!-- 1. Sütun: Stok Kodu -->
          <Column field="code" header="Stok Kodu" sortable style="min-width: 120px">
            <template #body="{ data }">
              <span class="text-sky-700 dark:text-sky-300 font-mono text-sm bg-sky-50 dark:bg-sky-950/40 px-1.5 py-0.5 rounded border border-sky-200 dark:border-sky-800 whitespace-nowrap" :title="data.code">
                {{ truncate(data.code) }}
              </span>
            </template>
          </Column>

          <!-- 2. Sütun: Ürün/Hizmet Adı ve Marka -->
          <Column field="name" header="Ürün/Hizmet Adı" sortable style="min-width: 200px" :bodyStyle="{ 'max-width': '250px', 'white-space': 'normal' }">
            <template #body="{ data }">
              <div class="flex flex-col gap-1 py-1">
                <span class="font-medium text-slate-700 dark:text-slate-200 text-sm leading-tight" style="word-break: break-word;" :title="data.name">
                  {{ truncate(data.name) }}
                </span>
                <span v-if="data.brand" class="text-xs text-slate-500 flex items-center gap-1" :title="data.brand">
                  <span class="font-medium text-slate-400">Marka:</span> <span class="text-slate-600 dark:text-slate-300 font-medium">{{ truncate(data.brand) }}</span>
                </span>
              </div>
            </template>
          </Column>

          <!-- 3. Sütun: Barkod No -->
          <Column field="barcode" header="Barkod No" sortable style="min-width: 120px">
            <template #body="{ data }">
              <span class="text-slate-600 dark:text-slate-300 text-sm font-medium truncate max-w-[140px] inline-block align-middle" :title="data.barcode">
                {{ data.barcode || '-' }}
              </span>
            </template>
          </Column>

          <!-- 4. Sütun: Özel Kod -->
          <Column header="Özel Kod" style="min-width: 140px" :bodyStyle="{ 'max-width': '200px', 'white-space': 'normal' }">
            <template #body="{ data }">
              <div v-if="data.custom_codes" class="flex flex-wrap items-center gap-1 py-1">
                <span v-for="(code, idx) in String(data.custom_codes).split(',')" :key="idx" class="px-1.5 py-0.5 bg-slate-100 dark:bg-slate-800 rounded text-[10px] border border-slate-200 dark:border-slate-700 font-mono text-slate-600 dark:text-slate-300 truncate max-w-[160px] inline-block align-middle" :title="code.trim()">
                  {{ code.trim() }}
                </span>
              </div>
              <span v-else class="text-slate-400 text-sm">-</span>
            </template>
          </Column>

          <!-- 5. Sütun: Kategori -->
          <Column field="category.name" header="Kategori" sortable style="min-width: 140px">
            <template #body="{ data }">
              <span class="text-slate-700 dark:text-slate-200 font-medium text-sm truncate max-w-[160px] inline-block align-middle" :title="data.category?.name || '-'">
                {{ data.category?.name || '-' }}
              </span>
            </template>
          </Column>

          <!-- 6. Sütun: Fiyat ve Maliyetler -->
          <Column field="average_cost" header="Fiyat & Maliyet" sortable style="min-width: 180px" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <div class="flex flex-col items-end gap-1 text-sm">
                <!-- Ortalama maliyet her zaman firmanın varsayılan dövizinde (örn. TL) -->
                <div class="text-sky-600 dark:text-sky-400 flex items-center gap-1.5 font-semibold" v-tooltip="'Anlık Ağırlıklı Ortalama Maliyet (varsayılan döviz)'">
                  <span class="text-[10px] uppercase font-bold text-sky-500/70">Ort. Maliyet</span>
                  <Money :value="data.average_cost" :currency="defaultCurrencyCode" />
                </div>
                <div class="font-medium text-slate-700 dark:text-slate-200 flex items-center gap-1.5 mt-1 border-t border-slate-100 dark:border-slate-800 pt-1">
                  <span class="text-[10px] uppercase font-bold text-slate-400">Satış</span>
                  <Money :value="data.sale_price" :currency="data.currency" />
                </div>
                <div v-if="convertToDefault(parseFloat(data.sale_price) || 0, data.currency) !== null" class="text-[11px] text-slate-400 italic">
                  ≈ <Money :value="convertToDefault(parseFloat(data.sale_price) || 0, data.currency)!.toString()" :currency="defaultCurrencyCode" />
                </div>
              </div>
            </template>
          </Column>

          <!-- 7. Sütun: Stok Adedi -->
          <Column field="current_stock" header="Stok & Envanter" sortable style="min-width: 150px">
            <template #body="{ data }">
              <div class="flex flex-col gap-1.5">
                <Tag :value="getStockStatus(data).label" :severity="getStockStatus(data).severity as any" class="px-2 py-1 text-xs w-fit" />
                <div v-if="data.track_stock && data.type === 'product'" class="text-[11px] text-slate-500 flex items-center gap-1">
                  <span class="font-medium text-slate-400">Top. Değer:</span>
                  <Money :value="(Math.round(parseFloat(data.current_stock) * 10000) / 10000 || 0) * (Math.round(parseFloat(data.average_cost) * 100) / 100 || 0)" :currency="defaultCurrencyCode" class="font-semibold text-slate-600 dark:text-slate-300" />
                </div>
              </div>
            </template>
          </Column>

          <Column header="Aksiyonlar" style="min-width: 150px" headerClass="justify-content-center">
            <template #body="{ data }">
              <div class="actions-cell flex gap-1 justify-center">
                <Button icon="pi pi-history" class="p-button-rounded p-button-text p-button-info p-1" @click="viewMovements(data)" v-tooltip.top="'Stok Geçmişi'" />
                <Button v-if="data.type === 'product' && can('products', 'update')" icon="pi pi-sliders-h" class="p-button-rounded p-button-text p-button-warning p-1" @click="openAdjustment(data)" v-tooltip.top="'Stok Düzelt'" />
                <Button v-if="can('products', 'update')" icon="pi pi-pencil" class="p-button-rounded p-button-text p-1" @click="editProduct(data.id)" v-tooltip.top="'Düzenle'" severity="warn" />
                <Button v-if="can('products', 'delete')" icon="pi pi-trash" class="p-button-rounded p-button-text p-1" @click="deleteProductItem(data.id)" v-tooltip.top="'Sil'" severity="danger" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Product Form Modal -->
    <FormModal
      v-if="showProductModal"
      v-model:visible="showProductModal"
      :product-id="selectedProductId"
      @saved="() => { loadData(); notificationStore.fetchNotifications(); }"
    />

    <!-- Manual Stock Adjustment Dialog -->
    <Dialog v-model:visible="showAdjustmentModal" header="Stok Giriş / Çıkış Hareketi (Düzeltme)" :modal="true" :style="{ width: '95vw', maxWidth: '800px' }">
      <div class="flex flex-col gap-4 py-2">
        <!-- Product Select -->
        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Ürün *</label>
          <Select
            v-model="adjustmentForm.product_id"
            :options="productStore.products.filter(p => p.type === 'product')"
            optionLabel="name"
            optionValue="id"
            placeholder="Ürün seçin..."
            class="w-full"
            filter
          />
        </div>

        <!-- Warehouse Select -->
        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Depo *</label>
          <Select
            v-model="adjustmentForm.warehouse_id"
            :options="productStore.warehouses"
            optionLabel="name"
            optionValue="id"
            placeholder="Depo seçin..."
            class="w-full"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <!-- Movement Type -->
          <div class="flex flex-col gap-1">
            <label class="text-xs font-semibold">Hareket Yönü</label>
            <Select
              v-model="adjustmentForm.type"
              :options="[{label:'Stok Girişi (+)', value:'in'}, {label:'Stok Çıkışı (-)', value:'out'}]"
              optionLabel="label"
              optionValue="value"
              class="w-full"
            />
          </div>

          <!-- Quantity -->
          <div class="flex flex-col gap-1">
            <label class="text-xs font-semibold">Miktar *</label>
            <InputNumber v-model="adjustmentForm.quantity" :min="1" class="w-full" />
          </div>
        </div>

        <!-- Cost -->
        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Birim Maliyet (Alış)</label>
          <InputNumber v-model="adjustmentForm.unit_cost" mode="decimal" :minFractionDigits="2" class="w-full" />
        </div>

        <!-- Note -->
        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Açıklama / Not</label>
          <InputText v-model="adjustmentForm.note" placeholder="Ör: Sayım eksiği, sarf vb..." class="w-full" />
        </div>

        <div class="flex justify-end gap-2 mt-4">
          <Button label="Vazgeç" class="p-button-text p-button-secondary" @click="showAdjustmentModal = false" outlined />
          <Button label="Kaydet" icon="pi pi-check" @click="handleCreateAdjustment" outlined severity="success" />
        </div>
      </div>
    </Dialog>

    <!-- Stock Movements Log Dialog -->
    <Dialog
      v-model:visible="showMovementsModal"
      :header="activeMovementProduct ? `${activeMovementProduct.name} - Stok Hareket Geçmişi` : 'Stok Hareketi Logları'"
      :modal="true"
      :style="{ width: '90%', maxWidth: '850px' }"
    >
      <div v-if="productStore.loading" class="text-center py-8">
        <i class="pi pi-spin pi-spinner text-2xl mb-2"></i>
        <div>Yükleniyor...</div>
      </div>
      <div v-else>
        <DataTable
          :value="productStore.movements"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          paginator
          :rows="10"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ slotProps.index + 1 }}
            </template>
          </Column>
          <Column field="date" header="Tarih" style="min-width: 150px">
            <template #body="{ data }">
              <span class="text-xs">{{ new Date(data.date).toLocaleDateString('tr-TR') }}</span>
            </template>
          </Column>
          <Column field="warehouse.name" header="Depo" style="min-width: 200px">
            <template #body="{ data }">
              <span class="text-xs font-medium">{{ data.warehouse?.name || 'Depo Belirsiz' }}</span>
            </template>
          </Column>
          <Column field="type" header="Yön" style="min-width: 120px">
            <template #body="{ data }">
              <Tag :value="data.type === 'in' ? 'Giriş (+)' : 'Çıkış (-)'" :severity="data.type === 'in' ? 'success' : 'danger'" class="text-xs" />
            </template>
          </Column>
          <Column field="quantity" header="Miktar" style="min-width: 120px">
            <template #body="{ data }">
              <span class="font-bold text-xs">{{ data.quantity }}</span>
            </template>
          </Column>
          <Column field="unit_cost" header="Birim Değer" style="min-width: 150px">
            <template #body="{ data }">
              <Money :value="data.unit_cost" />
            </template>
          </Column>
          <Column field="source_type" header="Kaynak" style="min-width: 150px">
            <template #body="{ data }">
              <Tag :value="getMovementSourceLabel(data.source_type)" severity="info" class="text-xs" />
            </template>
          </Column>
          <Column field="note" header="Açıklama" style="min-width: 200px">
            <template #body="{ data }">
              <span class="text-xs text-slate-500">{{ data.note }}</span>
            </template>
          </Column>
        </DataTable>
      </div>
    </Dialog>
  </div>
</template>

<style scoped>
.product-list-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin-bottom: 0.25rem;
}

.page-desc {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.95rem;
}

.header-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.summary-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.5rem;
}

.metric-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .metric-card {
  border-color: #334155;
  background-color: #1e293b;
}

.metric-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.metric-label {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.85rem;
  font-weight: 600;
}

.metric-value {
  font-size: 1.75rem;
  font-weight: 700;
  letter-spacing: -0.025em;
}

.metric-value.blue {
  color: #3b82f6;
}

.metric-value.green {
  color: #16a34a;
}

.metric-value.red {
  color: #dc2626;
}

.table-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .table-card {
  border-color: #334155;
  background-color: #1e293b;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.search-input {
  position: relative;
  flex: 1;
  max-width: 400px;
  min-width: 250px;
}

.search-input input {
  padding-left: 2.5rem;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--p-text-muted-color, #94a3b8);
}

.filters-group {
  display: flex;
  gap: 0.5rem;
}

.filter-select {
  width: 160px;
}

.actions-cell {
  display: flex;
  justify-content: center;
  gap: 0.25rem;
}

.line-clamp-1 {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}


</style>

