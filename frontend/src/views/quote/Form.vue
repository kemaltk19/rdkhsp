<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useQuoteStore } from '@/stores/quote'
import { useCariStore } from '@/stores/cari'
import { useSettingsStore } from '@/stores/settings'
import { useCurrencyStore } from '@/stores/currency'
import { useProductStore } from '@/stores/product'
import { getCurrentCompanyDatetimeLocal, toBackendDate, toCompanyDatetimeLocal } from '@/utils/date'
import { useToast } from 'primevue/usetoast'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import AutoComplete from 'primevue/autocomplete'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import Money from '@/components/Money.vue'
import Decimal from 'decimal.js'

const quoteStore = useQuoteStore()
const cariStore = useCariStore()
const settingsStore = useSettingsStore()
const currencyStore = useCurrencyStore()
const productStore = useProductStore()
const router = useRouter()
const route = useRoute()
const toast = useToast()

const loading = ref(false)
const errorMsg = ref('')
const isEdit = computed(() => route.params.id && route.params.id !== 'new')
const defaultKdvRate = ref(20)

const form = ref({
  cari_id: '',
  type: 'sales',
  number: '',
  date: '',
  expiry_date: '',
  currency: 'TRY', // şirketin varsayılan dövizi — kullanıcı tarafından seçilmez
  note: '',
  status: 'draft',
  discount_total: '0',
  items: [] as Array<{
    product_id: string | null
    selected_product?: any
    description: string
    quantity: string
    unit: string
    unit_price: string
    discount_rate: string
    tax_rate: number
    currency: string
    exchange_rate: string
    exchange_rate_op: '*' | '/'
  }>,
})

const typeOptions = [
  { label: 'Satış Teklifi', value: 'sales' },
  { label: 'Alış Teklifi', value: 'purchase' },
]

const statusOptions = [
  { label: 'Taslak', value: 'draft' },
  { label: 'Gönderildi', value: 'sent' },
]

const taxRateOptions = ref([
  { label: '%0', value: 0 }, { label: '%1', value: 1 },
  { label: '%10', value: 10 }, { label: '%20', value: 20 },
])

const unitOptions = [
  { label: 'Adet', value: 'Adet' }, { label: 'Kg', value: 'Kg' },
  { label: 'Gram', value: 'Gram' }, { label: 'Litre', value: 'Litre' },
  { label: 'Kutu', value: 'Kutu' }, { label: 'Paket', value: 'Paket' },
  { label: 'Metre', value: 'Metre' }, { label: 'M2', value: 'M2' },
  { label: 'M3', value: 'M3' }, { label: 'Koli', value: 'Koli' },
  { label: 'Palet', value: 'Palet' }, { label: 'Saat', value: 'Saat' },
  { label: 'Gün', value: 'Gün' }, { label: 'Ay', value: 'Ay' },
  { label: 'Takım', value: 'Takım' },
]

const loadTaxRates = async () => {
  try {
    const val = await settingsStore.fetchSetting('kdv_rates')
    if (val) {
      const parsed = JSON.parse(val)
      if (Array.isArray(parsed) && parsed.length > 0)
        taxRateOptions.value = parsed
          .map((r: any) => Number(r)).sort((a: number, b: number) => a - b)
          .map((r: number) => ({ label: `%${r}`, value: r }))
    }
  } catch {}
}

const filteredCaris = computed(() => cariStore.caris.filter(c => {
  if (form.value.type === 'sales') return c.type === 'customer' || c.type === 'both'
  return c.type === 'supplier' || c.type === 'both'
}))

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([
      cariStore.fetchCaris({ page: 1, limit: 1000 }),
      loadTaxRates(), currencyStore.fetchCurrencies(),
      settingsStore.fetchCompanyProfile(),
      productStore.fetchProducts({ page: 1, limit: 1000 }),
    ])
    try {
      const companyKdv = await settingsStore.fetchSetting('default_kdv_rate')
      if (companyKdv) defaultKdvRate.value = Number(companyKdv)
    } catch {}
    if (!isEdit.value) {
      form.value.currency = settingsStore.company?.currency || 'TRY'
      form.value.date = getCurrentCompanyDatetimeLocal()
      form.value.expiry_date = getCurrentCompanyDatetimeLocal(15)
      addRow()
      try {
        const n = await settingsStore.fetchSetting('quote_footer_note')
        if (n) form.value.note = n
      } catch {}
    } else {
      const id = route.params.id as string
      await quoteStore.fetchQuoteByID(id)
      if (quoteStore.activeQuote) {
        const inv = quoteStore.activeQuote
        if (inv.status !== 'draft') {
          toast.add({ severity: 'error', summary: 'Hata', detail: 'Sadece taslak teklifler düzenlenebilir.', life: 10000 })
          router.push('/quotes'); return
        }
        form.value = {
          cari_id: inv.cari_id, type: inv.type || 'sales', number: inv.number,
          date: inv.date ? toCompanyDatetimeLocal(inv.date) : '',
          expiry_date: inv.expiry_date ? toCompanyDatetimeLocal(inv.expiry_date) : '',
          currency: inv.currency, note: inv.note || '', status: inv.status,
          discount_total: (inv.discount_total || '0').toString(),
          items: (inv.items || []).map(item => {
            const p = item.product_id ? productStore.products.find(p => p.id === item.product_id) : null
            return {
              product_id: item.product_id,
              selected_product: p ? { ...p, displayName: p.brand ? `${p.brand} ${p.name}` : p.name } : null,
              description: item.description,
              quantity: (item.quantity || '0').toString(), unit: item.unit || 'Adet',
              unit_price: (item.unit_price || '0').toString(),
              discount_rate: '0',
              tax_rate: Number(item.tax_rate || 0),
              currency: item.currency || inv.currency,
              exchange_rate: toKurDisplay((item.exchange_rate || '1').toString(), item.currency || inv.currency),
              exchange_rate_op: (item.exchange_rate_op === '/' ? '/' : '*') as '*' | '/',
            }
          }),
        }
      }
    }
  } catch (err: any) {
    errorMsg.value = err?.response?.status === 403
      ? 'Bu işlem için yetkiniz bulunmamaktadır. Lütfen yöneticinize başvurun.'
      : 'Veriler yüklenirken hata oluştu.'
  }
  finally { loading.value = false }
})

const selectedCari = computed(() => cariStore.caris.find(c => c.id === form.value.cari_id))
const addRow = () => {
  form.value.items.push({
    product_id: null,
    selected_product: null,
    description: '',
    quantity: '1',
    unit: 'Adet',
    unit_price: '0',
    discount_rate: '0',
    tax_rate: defaultKdvRate.value,
    currency: form.value.currency,
    exchange_rate: '1',
    exchange_rate_op: '*',
  })
}
const removeRow = (i: number) => { if (form.value.items.length > 1) form.value.items.splice(i, 1) }

const productSuggestions = ref<any[]>([])

const searchProducts = async (event: any) => {
  await productStore.fetchProducts({ q: event.query, limit: 50 })
  productSuggestions.value = productStore.products.map(p => ({
    ...p,
    displayName: p.brand ? `${p.brand} ${p.name}` : p.name
  }))
}

const itemProductLabel = (item: any) => {
  const p = item.product_id ? productStore.products.find(p => p.id === item.product_id) : null
  if (!p) return ''
  return p.brand ? `${p.brand} ${p.name}` : p.name
}

const handleProductSelect = (item: any, event: any) => {
  const p = event.value
  if (p) {
    const isFirstTimeSelection = !item.product_id

    item.product_id = p.id
    item.description = p.description || ''
    item.unit = p.unit || 'Adet'
    // Ürünün satış fiyatı KDV dahilse, teklif birim fiyatı KDV hariç olmalı:
    // backend birim fiyatın üzerine KDV ekler, çift KDV'yi önlemek için burada ayrıştırılır.
    const rawPrice = parseDec(p.sale_price || '0')
    const tRate = parseDec(p.tax_rate || '0')
    const netPrice = p.tax_included
      ? rawPrice.div(new Decimal(1).add(tRate.div(100))) // tRate=0 → /1, güvenli
      : rawPrice
    item.unit_price = netPrice.toDecimalPlaces(4).toString()

    // Satır, ürünün kendi dövizinde kalır; varsayılan dövizden farklıysa
    // Para Birimleri ekranındaki güncel kur+işaret otomatik önerilir.
    item.currency = p.currency || form.value.currency
    if (item.currency !== form.value.currency) {
      const c = currencyStore.getCurrencyByCode(item.currency)
      // Önerilen kuru satırın para birimi ayracıyla göster (TRY=virgül, USD=nokta).
      item.exchange_rate = c ? toKurDisplay(c.exchange_rate.toString(), item.currency) : '1'
      item.exchange_rate_op = c?.exchange_rate_op === '/' ? '/' : '*'
    } else {
      item.exchange_rate = '1'
      item.exchange_rate_op = '*'
    }

    if (isFirstTimeSelection || item.tax_rate === defaultKdvRate.value) {
      // Öncelik: ürünün kendi KDV'si > ürün kategorisinin varsayılan KDV'si > genel varsayılan KDV.
      const productRate = parseDec(p.tax_rate)
      const categoryRate = p.category ? parseDec(p.category.default_kdv_rate) : null
      if (productRate.gt(0)) {
        item.tax_rate = productRate.toNumber()
      } else if (categoryRate && categoryRate.gt(0)) {
        item.tax_rate = categoryRate.toNumber()
      } else {
        item.tax_rate = defaultKdvRate.value
      }
    }
  } else {
    item.product_id = null
  }
}

const handleProductChange = (item: any) => {
  // Keeping this for compatibility or direct selection fallback
  if (!item.selected_product) {
    item.product_id = null
    return
  }
}

// Kullanıcı virgül veya nokta ondalık ayracıyla yazabilir; her ikisini de güvenle
// Decimal'e çevirir. Son görülen ',' veya '.' ondalık ayraç kabul edilir; öncesindeki
// tüm ',' '.' ve boşluklar binlik ayraç sayılıp temizlenir.
const normalizeNum = (val: any): string => {
  if (typeof val === 'number') return val.toString()
  let s = String(val ?? '').trim()
  if (!s) return '0'
  s = s.replace(/\s/g, '')
  const lastComma = s.lastIndexOf(',')
  const lastDot = s.lastIndexOf('.')
  const decPos = Math.max(lastComma, lastDot)
  if (decPos === -1) return s.replace(/[^\d-]/g, '')
  const intPart = s.slice(0, decPos).replace(/[^\d-]/g, '')
  const fracPart = s.slice(decPos + 1).replace(/[^\d]/g, '')
  return `${intPart}.${fracPart}`
}
const parseDec = (val: any) => { try { return new Decimal(normalizeNum(val)) } catch(e) { return new Decimal(0) } }

const lSub   = (i: any) => parseDec(i.quantity).mul(parseDec(i.unit_price))
const lTax   = (i: any) => lSub(i).mul(parseDec(i.tax_rate).div(100))
const lTot   = (i: any) => lSub(i).add(lTax(i))

// Satırın para biriminin ondalık ayracı (Para Birimleri ekranı ayarı; TRY=virgül, USD=nokta vb.)
const decSepFor = (code: string) => {
  const c = currencyStore.getCurrencyByCode(code)
  return c?.format_decimal_sep || '.'
}
// Backend'den gelen ham kuru (her zaman '.') satırın para birimi ayracıyla gösterime çevirir.
const toKurDisplay = (raw: any, code: string) => {
  const s = String(raw ?? '')
  if (!s) return ''
  return decSepFor(code) === ',' ? s.replace('.', ',') : s
}

// Satırı kendi dövizinden, satırın kendi kuru/işaretiyle varsayılan dövize çevirir.
// Backend quote_service.go'daki convertToDefaultCurrency ile aynı mantık.
const convertItem = (i: any, amount: any) => {
  if (!i.currency || i.currency === form.value.currency) return amount
  const rate = parseDec(i.exchange_rate)
  if (rate.lte(0)) return amount
  return i.exchange_rate_op === '/' ? amount.div(rate) : amount.mul(rate)
}

const subtotal      = computed(() => form.value.items.reduce((s, i) => s.add(convertItem(i, lSub(i))), new Decimal(0)))
const taxTotal      = computed(() => form.value.items.reduce((s, i) => s.add(convertItem(i, lTax(i))), new Decimal(0)))
const grandTotal    = computed(() => subtotal.value.add(taxTotal.value).add(parseDec(form.value.discount_total)))

const cancel = () => router.push('/quotes')

const handleSubmit = async () => {
  if (!form.value.cari_id) { errorMsg.value = 'Lütfen bir Cari Hesap seçin.'; return }
  if (form.value.items.some(i => !i.product_id && !i.description)) { errorMsg.value = 'Tüm kalemler için ürün veya açıklama zorunludur.'; return }
  if (form.value.items.some(i => parseDec(i.quantity).lte(0))) { errorMsg.value = 'Miktar sıfırdan büyük olmalıdır.'; return }
  loading.value = true; errorMsg.value = ''
  try {
    const payload = {
      ...form.value,
      subtotal: subtotal.value.toString(),
      tax_total: taxTotal.value.toString(),
      total: grandTotal.value.toString(),
      date: toBackendDate(form.value.date),
      expiry_date: toBackendDate(form.value.expiry_date),
      items: form.value.items.map(item => {
        const p = item.product_id ? productStore.products.find(p => p.id === item.product_id) : null
        return {
          ...item,
          description: p
            ? `${p.brand ? p.brand + ' ' : ''}${p.name}${item.description ? ' - ' + item.description : ''}`
            : item.description,
          quantity: parseDec(item.quantity).toString(), unit_price: parseDec(item.unit_price).toString(),
          discount_rate: '0', tax_rate: item.tax_rate.toString(),
          line_total: lSub(item).toString(),
          currency: item.currency,
          exchange_rate: parseDec(item.exchange_rate).toString(),
          exchange_rate_op: item.exchange_rate_op,
        }
      }),
    }
    let savedQuote
    if (isEdit.value) {
      savedQuote = await quoteStore.updateQuote(route.params.id as string, payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif güncellendi', life: 10000 })
    } else {
      savedQuote = await quoteStore.createQuote(payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif oluşturuldu', life: 10000 })
    }

    // Backend, mail gönderimi başarısız olursa hata döner ve teklif "sent" olamaz;
    // bu yüzden bu mesaj sadece backend'den dönen gerçek statü "sent" ise gösterilir.
    if (savedQuote?.status === 'sent') {
      toast.add({ severity: 'info', summary: 'E-posta', detail: 'Teklif müşteriye e-posta ile gönderildi.', life: 10000 })
    }
    router.push('/quotes')
  } catch (err: any) {
    errorMsg.value = err?.response?.data?.error?.message || 'İşlem sırasında bir hata oluştu.'
  } finally { loading.value = false }
}
</script>

<template>
  <div class="fp-page">

    <!-- Action Bar -->
    <div class="fp-bar">
      <div class="fp-bar-l">
        <Button class="fp-back" @click="cancel" :disabled="loading" severity="secondary">
          <i class="pi pi-arrow-left"></i> Geri
        </button>
        <span class="fp-bar-title">{{ isEdit ? 'Teklifi Düzenle' : 'Yeni Teklif' }}</span>
      </div>
      <div class="fp-bar-r">
        <Select v-model="form.type" :options="typeOptions" optionLabel="label" optionValue="value" class="fp-bar-sel" :disabled="isEdit || loading" />
        <Select v-model="form.status" :options="statusOptions" optionLabel="label" optionValue="value" class="fp-bar-sel" :disabled="loading" />
        <Button class="fp-save" type="submit" form="fp-form" :disabled="loading" severity="primary">
          <i class="pi pi-check"></i>{{ loading ? 'Kaydediliyor...' : (isEdit ? 'Güncelle' : 'Kaydet') }}
        </button>
      </div>
    </div>

    <Message v-if="errorMsg" severity="error" class="fp-err">{{ errorMsg }}</Message>

    <!-- ═══ DOCUMENT ═══ -->
    <form id="fp-form" @submit.prevent="handleSubmit" class="fp-doc">

      <!-- ── Header: split panel ── -->
      <div class="fp-head">
        <div class="fp-head-l">
          <div class="fp-logo-row">
            <div class="fp-logo"><i class="pi pi-chart-line"></i></div>
            <div>
              <div class="fp-co-name">{{ settingsStore.company?.name || 'FİRMA ADINIZ' }}</div>
              <div class="fp-co-sub">{{ settingsStore.company?.title || '' }}</div>
            </div>
          </div>
          <div class="fp-divider"></div>
          <div class="fp-to-label">TEKLİF KESİLEN</div>
          <Select v-model="form.cari_id" :options="filteredCaris" optionLabel="name" optionValue="id"
            placeholder="Cari hesap seçin..." class="fp-cari-sel" filter :disabled="isEdit || loading" />
          <template v-if="selectedCari">
            <div class="fp-cari-name">{{ selectedCari.name }}</div>
            <div v-if="selectedCari.address" class="fp-cari-row"><i class="pi pi-map-marker"></i>{{ selectedCari.address }}</div>
            <div v-if="selectedCari.phone" class="fp-cari-row"><i class="pi pi-phone"></i>{{ selectedCari.phone }}</div>
            <div v-if="selectedCari.email" class="fp-cari-row"><i class="pi pi-envelope"></i>{{ selectedCari.email }}</div>
            <div v-if="selectedCari.tax_number" class="fp-cari-row"><i class="pi pi-id-card"></i>V.D.: {{ selectedCari.tax_office }} | V.N.: {{ selectedCari.tax_number }}</div>
          </template>
          <div v-else class="fp-cari-empty">Listeden müşteri / tedarikçi seçin</div>
        </div>

        <div class="fp-head-r">
          <div v-if="settingsStore.company?.phone" class="fp-cr"><i class="pi pi-phone"></i><span>{{ settingsStore.company.phone }}</span></div>
          <div v-if="settingsStore.company?.website" class="fp-cr"><i class="pi pi-globe"></i><span>{{ settingsStore.company.website }}</span></div>
          <div v-if="settingsStore.company?.email" class="fp-cr"><i class="pi pi-envelope"></i><span>{{ settingsStore.company.email }}</span></div>
          <div v-if="settingsStore.company?.address" class="fp-cr"><i class="pi pi-map-marker"></i><span>{{ settingsStore.company.address }}</span></div>
          <div v-if="settingsStore.company?.tax_number" class="fp-cr"><i class="pi pi-id-card"></i><span>V.N.: {{ settingsStore.company.tax_number }}</span></div>
        </div>
      </div>

      <!-- ── Meta bar ── -->
      <div class="fp-meta">
        <div class="fp-mc">
          <div class="fp-ml">GENEL TOPLAM</div>
          <div class="fp-mv fp-mv-total"><Money :value="grandTotal.toString()" :currency="form.currency" /></div>
        </div>
        <div class="fp-ms"></div>
        <div class="fp-mc">
          <div class="fp-ml">TEKLİF NO</div>
          <InputText v-model="form.number" placeholder="Otomatik (boş bırakın)" class="fp-mi uppercase-input" :disabled="isEdit || loading" maxlength="100" />
        </div>
        <div class="fp-ms"></div>
        <div class="fp-mc">
          <div class="fp-ml">TEKLİF TARİHİ</div>
          <input type="datetime-local" v-model="form.date" class="fp-md" required :disabled="loading" />
        </div>
        <div class="fp-ms"></div>
        <div class="fp-mc">
          <div class="fp-ml">GEÇERLİLİK TARİHİ</div>
          <input type="datetime-local" v-model="form.expiry_date" class="fp-md" required :disabled="loading" />
        </div>
        <div class="fp-ms"></div>
        <div class="fp-mc">
          <div class="fp-ml">VARSAYILAN DÖVİZ</div>
          <div class="fp-mv">{{ form.currency }}</div>
        </div>
      </div>

      <!-- ── Items Table ── -->
      <div class="fp-tbl-wrap">
        <table class="fp-tbl">
          <colgroup>
            <col style="width:30px">
            <col style="width:auto">
            <col style="width:80px">
            <col style="width:70px">
            <col style="width:130px">
            <col style="width:90px">
            <col style="width:130px">
            <col style="width:36px">
          </colgroup>
          <thead>
            <tr>
              <th class="fth c">#</th>
              <th class="fth">ÜRÜN / HİZMET AÇIKLAMASI</th>
              <th class="fth c">BİRİM</th>
              <th class="fth c">MİKT.</th>
              <th class="fth r">BİRİM FİYAT</th>
              <th class="fth c">KDV%</th>
              <th class="fth r">KDV'SİZ TUTAR</th>
              <th class="fth"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, idx) in form.items" :key="idx" class="ftr">
              <td class="ftd c ftd-no">{{ idx + 1 }}</td>
              <td class="ftd ftd-desc">
                <AutoComplete v-model="item.selected_product" :suggestions="productSuggestions"
                  @complete="searchProducts" optionLabel="displayName"
                  placeholder="Ürün veya Kod Ara..." dropdown class="fi-sel w-full"
                  @item-select="handleProductSelect(item, $event)"
                  @change="handleProductChange(item)" :disabled="loading">
                  <template #option="slotProps">
                    <div class="flex flex-col">
                      <span class="font-bold">{{ slotProps.option.brand ? slotProps.option.brand + ' ' : '' }}{{ slotProps.option.name }}</span>
                      <span class="text-xs text-slate-500" v-if="slotProps.option.custom_codes">Özel Kod: {{ slotProps.option.custom_codes }}</span>
                      <span class="text-xs text-slate-500" v-if="slotProps.option.barcode">Barkod: {{ slotProps.option.barcode }}</span>
                    </div>
                  </template>
                </AutoComplete>
                <div v-if="item.product_id" class="fi-product-label">{{ itemProductLabel(item) }}</div>
                <InputText v-model="item.description"
                  :placeholder="item.product_id ? 'Ek açıklama (opsiyonel)...' : 'Stokta yoksa: ürün/hizmet adını buraya yazın...'"
                  class="fi-txt fi-desc" :disabled="loading" maxlength="2000" />
              </td>
              <td class="ftd">
                <Select v-model="item.unit" :options="unitOptions" optionLabel="label"
                  optionValue="value" class="fi-sel" filter :disabled="loading" />
              </td>
              <td class="ftd c">
                <InputText v-model="item.quantity" class="fi-num text-right w-full p-1 border rounded" :disabled="loading" />
              </td>
              <td class="ftd r">
                <InputText v-model="item.unit_price" class="fi-num text-right w-full p-1 border rounded" :disabled="loading" />
                <div v-if="item.currency && item.currency !== form.currency" class="fi-kur-row">
                  <span class="fi-kur-cur">{{ item.currency }}</span>
                  <Button type="button" class="fi-kur-op"
                    @click="item.exchange_rate_op = item.exchange_rate_op === '/' ? '*' : '/'"
                    :disabled="loading" :title="item.exchange_rate_op === '/' ? 'Böl' : 'Çarp'">
                    {{ item.exchange_rate_op === '/' ? '÷' : '×' }}
                  </button>
                  <InputText v-model="item.exchange_rate" class="fi-kur-rate" :disabled="loading" inputmode="decimal" />
                </div>
              </td>
              <td class="ftd">
                <Select v-model="item.tax_rate" :options="taxRateOptions" optionLabel="label"
                  optionValue="value" class="fi-sel" :disabled="loading" />
              </td>
              <td class="ftd r ftd-linetot">
                <Money :value="lSub(item).toString()" :currency="item.currency || form.currency" />
              </td>
              <td class="ftd c">
                <Button type="button" class="fi-del" @click="removeRow(idx)"
                  :disabled="form.items.length <= 1 || loading">
                  <i class="pi pi-times"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <Button type="button" class="fp-addrow" @click="addRow" :disabled="loading">
          <i class="pi pi-plus"></i> Satır Ekle
        </button>
      </div>

      <!-- ── Bottom: Notes + Totals ── -->
      <div class="fp-bot">
        <div class="fp-bot-l">
          <div class="fp-sec-lbl">NOTLAR / ÖDEME BİLGİLERİ</div>
          <Textarea v-model="form.note" rows="4"
            placeholder="IBAN, banka bilgileri, ödeme koşulları..."
            class="fp-notes" :disabled="loading" maxlength="2000" />
          <div class="fp-sig">
            <div class="fp-sig-name">{{ settingsStore.company?.name || '' }}</div>
            <div class="fp-sig-line"></div>
            <div class="fp-sig-role">Yetkili / Authorized</div>
          </div>
        </div>

        <div class="fp-bot-r">
          <div class="fp-tots">
            <div class="fp-trow"><span>Ara Toplam</span><Money :value="subtotal.toString()" :currency="form.currency" /></div>
            <div class="fp-trow">
              <span class="whitespace-nowrap">İndirim / Ek (-/+)</span>
              <InputText v-model="form.discount_total" class="w-24 text-right p-1 border rounded text-xs" :disabled="loading" />
            </div>
            <div class="fp-trow"><span>KDV</span><Money :value="taxTotal.toString()" :currency="form.currency" /></div>
            <div class="fp-tsep"></div>
            <div class="fp-tgrand">
              <span>GENEL TOPLAM</span>
              <span class="fp-tgval"><Money :value="grandTotal.toString()" :currency="form.currency" /></span>
            </div>
          </div>
          <div class="fp-deco">{{ form.type === 'sales' ? 'TEKLİF' : 'ALIŞ' }}</div>
        </div>
      </div>

      <div class="fp-terms">
        <span>Teklif bedeli vade tarihinde ödenecektir. Türk Ticaret Kanunu hükümlerine uygun düzenlenmiştir.</span>
        <span class="fp-thanks">İş birliğiniz için teşekkürler!</span>
      </div>

    </form>
  </div>
</template>

<style scoped>
.fp-page { min-height: 100vh; background: #f1f5f9; padding-bottom: 60px; }

/* Action Bar */
.fp-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 20px; background: #fff; border-bottom: 1px solid #e2e8f0;
  gap: 10px; flex-wrap: wrap; position: sticky; top: 0; z-index: 20;
  box-shadow: 0 1px 3px rgba(15,23,42,.06);
}
.fp-bar-l { display: flex; align-items: center; gap: 10px; }
.fp-bar-r { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.fp-back {
  display: flex; align-items: center; gap: 5px; padding: 6px 12px;
  border: 1px solid #cbd5e1; border-radius: 7px; background: transparent;
  cursor: pointer; font-size: 12.5px; color: #64748b; transition: all .15s;
}
.fp-back:hover { border-color: #06b6d4; color: #06b6d4; }
.fp-bar-title { font-size: 14px; font-weight: 700; color: #0f172a; }
.fp-bar-sel { font-size: 12.5px; height: 32px; min-width: 130px; }
.fp-save {
  display: flex; align-items: center; gap: 6px; padding: 6px 18px;
  border: 1.5px solid #06b6d4; border-radius: 7px; background: transparent;
  color: #06b6d4; font-size: 13px; font-weight: 600; cursor: pointer; transition: all .15s;
}
.fp-save:hover { background: #ecfeff; }
.fp-save:disabled { opacity: .5; cursor: not-allowed; }
.fp-err { margin: 8px 20px 0; font-size: 1.05rem; font-weight: 600; }

/* Document Layout Match - A4 proportions */
.fp-doc {
  width: 96%;
  max-width: 1150px;
  min-height: 600px;
  margin: 20px auto;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #d1d9e0;
  box-shadow: 0 10px 25px rgba(0,0,0,0.1);
  overflow: hidden;
}

@media print {
  .fp-page { background: #fff; padding: 0; }
  .fp-bar { display: none; }
  .fp-doc {
    width: 210mm;
    min-height: 297mm;
    margin: 0;
    box-shadow: none;
    border: none;
    border-radius: 0;
  }
}

/* Header */
.fp-head { display: grid; grid-template-columns: 1fr 220px; border-bottom: 1px solid #d1d9e0; }
.fp-head-l { padding: 22px 24px; display: flex; flex-direction: column; gap: 6px; border-right: 1px solid #d1d9e0; }
.fp-logo-row { display: flex; align-items: center; gap: 12px; margin-bottom: 4px; }
.fp-logo {
  width: 42px; height: 42px; border-radius: 9px; background: #06b6d4;
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 18px; flex-shrink: 0;
}
.fp-co-name { font-size: 15px; font-weight: 800; color: #0f172a; text-transform: uppercase; letter-spacing: .3px; }
.fp-co-sub { font-size: 11px; color: #94a3b8; }
.fp-divider { border-top: 1px solid #e8edf5; margin: 4px 0; }
.fp-to-label { font-size: 9.5px; font-weight: 700; letter-spacing: 1.5px; color: #06b6d4; text-transform: uppercase; }
.fp-cari-sel { width: 100%; font-size: 12.5px; }
.fp-cari-name { font-size: 13px; font-weight: 700; color: #0f172a; margin-top: 2px; }
.fp-cari-row { font-size: 11.5px; color: #64748b; display: flex; align-items: center; gap: 5px; }
.fp-cari-row i { font-size: 9px; color: #06b6d4; }
.fp-cari-empty { font-size: 11.5px; color: #94a3b8; font-style: italic; }

.fp-head-r { background: #06b6d4; padding: 20px 16px; display: flex; flex-direction: column; gap: 8px; justify-content: center; }
.fp-cr { display: flex; align-items: flex-start; gap: 8px; }
.fp-cr i { font-size: 11px; color: rgba(255,255,255,.7); margin-top: 2px; flex-shrink: 0; }
.fp-cr span { font-size: 11.5px; color: #fff; line-height: 1.4; }

/* Meta bar */
.fp-meta {
  display: flex; align-items: stretch;
  background: #f8fafc; border-bottom: 3px solid #06b6d4;
  border-top: 1px solid #d1d9e0;
}
.fp-mc { padding: 10px 16px; flex: 1; display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.fp-ms { width: 1px; background: #d1d9e0; margin: 6px 0; flex-shrink: 0; }
.fp-ml { font-size: 9px; font-weight: 700; letter-spacing: 1px; color: #94a3b8; text-transform: uppercase; }
.fp-mv { font-size: 14px; font-weight: 700; color: #0f172a; }
.fp-mv-total { color: #06b6d4; font-size: 15px; }
.fp-mi {
  border-color: transparent !important; background: transparent !important;
  padding: 0 !important; font-size: 13px !important; font-weight: 600 !important;
  box-shadow: none !important; height: auto !important;
}
.fp-mi:focus { border-color: #06b6d4 !important; background: #fff !important; padding: 3px 6px !important; }
.fp-md {
  border: 1px solid transparent; border-radius: 5px; padding: 0;
  font-size: 13px; font-weight: 600; color: #0f172a; background: transparent; cursor: pointer; width: 100%;
}
.fp-md:focus { outline: none; border-color: #06b6d4; background: #fff; padding: 2px 6px; }
.fp-msel { font-size: 12.5px; }
:deep(.fp-msel .p-select) { border-color: transparent !important; background: transparent !important; font-weight: 600; padding: 0 4px !important; }
:deep(.fp-msel .p-select:focus-within) { border-color: #06b6d4 !important; background: #fff !important; }

/* Table Layout */
.fp-tbl-wrap { width: 100%; overflow-x: auto; margin-top: 10px; background: #fff; }
.fp-tbl { width: 100%; border-collapse: collapse; table-layout: fixed; font-size: 13px; }
.fth {
  background: #f8fafc; color: #334155;
  font-size: 11px; font-weight: 700; text-transform: uppercase;
  padding: 10px; border-bottom: 1px solid #e2e8f0; border-top: 1px solid #e2e8f0;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; text-align: left;
}
.fth.c { text-align: center; }
.fth.r { text-align: right; padding-right: 10px; }

.ftr { border-bottom: 1px solid #f1f5f9; transition: background .1s; }
.ftr:hover { background: #f8fafc; }
.ftd {
  padding: 8px 10px; vertical-align: middle; color: #334155;
}
.ftd.c { text-align: center; }
.ftd.r { text-align: right; padding-right: 10px; }
.ftd-no { color: #64748b; font-weight: 700; font-size: 12px; }
.ftd-desc { vertical-align: top; padding: 10px; }
.ftd-linetot { font-weight: 700; color: #0f172a; }

.fi-sel { width: 100%; }
.fi-txt { width: 100%; }
.fi-desc { margin-top: 6px; }
.fi-num { width: 100%; }
.fi-product-label { font-size: 11.5px; font-weight: 700; color: #0f172a; margin-top: 4px; }
.fi-kur-row {
  display: flex; align-items: center; gap: 2px; margin-top: 3px;
  background: #fffbeb; border: 1px solid #fde68a; border-radius: 4px; padding: 1px 2px;
}
.fi-kur-cur { font-size: 10px; font-weight: 700; color: #b45309; padding: 0 1px; flex: 0 0 auto; }
.fi-kur-op {
  flex: 0 0 auto; width: 22px; height: 24px; padding: 0;
  display: flex; align-items: center; justify-content: center;
  border: 1px solid #cbd5e1; border-radius: 4px; background: #fff;
  font-size: 13px; font-weight: 700; color: #475569; cursor: pointer; line-height: 1;
}
.fi-kur-op:hover:not(:disabled) { background: #f1f5f9; border-color: #94a3b8; }
.fi-kur-op:disabled { opacity: .5; cursor: not-allowed; }
.fi-kur-rate { flex: 1 1 auto; width: auto; min-width: 0; }
:deep(.fi-kur-rate.p-inputtext) { padding: 2px 4px !important; font-size: 11px !important; text-align: right; min-width: 0; }

/* Input boxes mock style */
:deep(.ftd .p-select),
:deep(.ftd .p-inputtext),
:deep(.ftd .p-inputnumber),
:deep(.ftd .p-inputnumber-input) {
  border: 1px solid #cbd5e1 !important;
  box-shadow: none !important;
  background: #fff !important;
  border-radius: 6px !important;
  font-size: 12.5px !important;
  width: 100%;
}
:deep(.ftd .p-select:hover),
:deep(.ftd .p-select:focus-within),
:deep(.ftd .p-inputtext:hover),
:deep(.ftd .p-inputtext:focus),
:deep(.ftd .p-inputnumber-input:hover),
:deep(.ftd .p-inputnumber-input:focus) {
  border-color: #06b6d4 !important;
}

:deep(.ftd .p-select-label) { padding: 6px 4px !important; display: flex; align-items: center; justify-content: center; text-overflow: clip; overflow: visible; }
:deep(.ftd .p-select-dropdown) { width: 24px !important; }
:deep(.ftd .p-inputtext) { padding: 6px 8px !important; }
:deep(.ftd .p-inputnumber-input) { padding: 6px 8px !important; }
:deep(.ftd .p-inputnumber) { border: none !important; }

.fi-del {
  width: 24px; height: 24px; border-radius: 4px; border: none;
  background: transparent; cursor: pointer; color: #94a3b8;
  display: flex; align-items: center; justify-content: center; font-size: 10px; transition: all .15s;
}
.fi-del:hover { background: #fee2e2; color: #dc2626; }
.fi-del:disabled { opacity: .3; cursor: not-allowed; }

.fp-addrow {
  display: flex; align-items: center; gap: 6px; width: 100%;
  padding: 9px 14px; background: #f8fafc;
  border: none; border-top: 1px dashed #cbd5e1;
  color: #06b6d4; font-size: 12.5px; font-weight: 600; cursor: pointer; transition: background .15s;
}
.fp-addrow:hover { background: #ecfeff; }

/* Bottom */
.fp-bot { display: grid; grid-template-columns: 1fr 280px; border-top: 1px solid #d1d9e0; }
.fp-bot-l { padding: 20px 24px; border-right: 1px solid #d1d9e0; }
.fp-sec-lbl { font-size: 9.5px; font-weight: 700; letter-spacing: 1.5px; color: #06b6d4; text-transform: uppercase; margin-bottom: 7px; }
.fp-notes { width: 100%; font-size: 12.5px; border-radius: 7px; resize: vertical; color: #475569; }
.fp-sig { margin-top: 16px; }
.fp-sig-name { font-size: 12px; font-weight: 700; color: #0f172a; margin-bottom: 4px; }
.fp-sig-line { width: 160px; border-bottom: 1px solid #94a3b8; padding-bottom: 18px; margin-bottom: 5px; }
.fp-sig-role { font-size: 10.5px; color: #94a3b8; }

.fp-bot-r { padding: 20px 22px; display: flex; flex-direction: column; justify-content: space-between; }
.fp-tots { display: flex; flex-direction: column; gap: 7px; }
.fp-trow { display: flex; justify-content: space-between; align-items: center; font-size: 12.5px; color: #64748b; gap: 16px; }
.fp-trow-disc { color: #dc2626; }
.fp-tsep { border-top: 1px solid #e2e8f0; margin: 3px 0; }
.fp-tgrand {
  display: flex; justify-content: space-between; align-items: center;
  background: #ecfeff; border: 1.5px solid #06b6d4; border-radius: 8px;
  padding: 10px 12px; gap: 16px;
}
.fp-tgrand span:first-child { font-size: 10px; font-weight: 700; letter-spacing: 1px; text-transform: uppercase; color: #0e7490; }
.fp-tgval { font-size: 17px; font-weight: 900; color: #0e7490; font-family: 'SFProNumbers', monospace; }
.fp-deco {
  text-align: right; font-size: 32px; font-weight: 900;
  color: rgba(6,182,212,.1); letter-spacing: 4px; text-transform: uppercase;
  margin-top: 12px; line-height: 1; user-select: none;
}

/* Terms footer */
.fp-terms {
  display: flex; justify-content: space-between; align-items: center;
  background: #f8fafc; border-top: 1px solid #d1d9e0;
  padding: 8px 24px; font-size: 10.5px; color: #94a3b8; gap: 16px; flex-wrap: wrap;
}
.fp-thanks { font-weight: 600; color: #06b6d4; white-space: nowrap; }

/* Dark mode */
:root.p-dark .fp-page { background: #0b0f1a; }
:root.p-dark .fp-bar { background: #111827; border-color: #1f2937; }
:root.p-dark .fp-bar-title { color: #f1f5f9; }
:root.p-dark .fp-doc { background: #111827; border-color: #1f2937; }
:root.p-dark .fp-head { border-color: #1f2937; }
:root.p-dark .fp-head-l { border-color: #1f2937; }
:root.p-dark .fp-co-name { color: #f1f5f9; }
:root.p-dark .fp-divider { border-color: #1f2937; }
:root.p-dark .fp-cari-name { color: #f1f5f9; }
:root.p-dark .fp-meta { background: #0f172a; border-color: #1f2937; }
:root.p-dark .fp-ms { background: #1f2937; }
:root.p-dark .fp-mv { color: #f1f5f9; }
:root.p-dark .fp-md { color: #e2e8f0; }
:root.p-dark .fp-md:focus { background: #1e293b; }
:root.p-dark .fth { background: #0c4a58; }
:root.p-dark .ftr { border-color: #1f2937; }
:root.p-dark .fi-product-label { color: #cbd5e1; }
:root.p-dark .ftr:nth-child(even) { background: #0f172a; }
:root.p-dark .ftr:hover { background: #0c1a2e; }
:root.p-dark .ftd { border-color: #1f2937; color: #e2e8f0; }
:root.p-dark .ftd-linetot { color: #f1f5f9; }
:root.p-dark .fp-addrow { background: #0f172a; border-color: #1f2937; }
:root.p-dark .fp-bot { border-color: #1f2937; }
:root.p-dark .fp-bot-l { border-color: #1f2937; }
:root.p-dark .fp-sig-name { color: #f1f5f9; }
:root.p-dark .fp-tsep { border-color: #1f2937; }
:root.p-dark .fp-tgrand { background: rgba(6,182,212,.1); border-color: #06b6d4; }
:root.p-dark .fp-tgval { color: #06b6d4; }
:root.p-dark .fp-terms { background: #0f172a; border-color: #1f2937; }
</style>
