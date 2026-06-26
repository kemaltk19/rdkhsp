<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useCariStore } from '@/stores/cari'
import { useSettingsStore } from '@/stores/settings'
import { useCurrencyStore } from '@/stores/currency'
import { useToast } from 'primevue/usetoast'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { countries } from '@/constants/countries'

const props = defineProps({
  visible: { type: Boolean, required: true },
  cariId: { type: String, default: null },
})
const emit = defineEmits(['update:visible', 'saved'])

const cariStore     = useCariStore()
const settingsStore = useSettingsStore()
const currencyStore = useCurrencyStore()
const toast         = useToast()

const loading       = ref(false)
const errorMsg      = ref('')
const invalidFields = ref<string[]>([])
const sameAsBilling = ref(false)

const form = ref({
  type: 'customer', group: '', code: '',
  name: '', contact_name: '',
  tax_office: '', tax_number: '',
  email: '', phone: '', landline: '', fax: '',
  address: '', city: '', district: '', postal_code: '', country: 'Türkiye',
  shipping_address: '', shipping_city: '', shipping_district: '',
  shipping_postal_code: '', shipping_country: 'Türkiye',
  currency: 'TRY', note: '',
})

const isEdit = computed(() => !!props.cariId)

const typeOptions = ref([
  { label: 'Müşteri',                         value: 'customer'  },
  { label: 'Tedarikçi',                       value: 'supplier'  },
  { label: 'Her İkisi (Müşteri + Tedarikçi)', value: 'both'      },
])
const groupOptions    = ref<{label:string,value:string}[]>([])
const currencyOptions = ref([
  { label: 'Türk Lirası (TRY)', value: 'TRY' },
  { label: 'Amerikan Doları (USD)', value: 'USD' },
  { label: 'Avro (EUR)', value: 'EUR' },
])

const copyBillingToShipping = () => {
  form.value.shipping_country    = form.value.country
  form.value.shipping_city       = form.value.city
  form.value.shipping_district   = form.value.district
  form.value.shipping_postal_code= form.value.postal_code
  form.value.shipping_address    = form.value.address
}

watch(sameAsBilling, v => { if (v) copyBillingToShipping() })
watch(
  () => [form.value.country, form.value.city, form.value.district, form.value.postal_code, form.value.address],
  () => { if (sameAsBilling.value) copyBillingToShipping() }
)

const close = () => emit('update:visible', false)

const loadCari = async () => {
  if (!props.cariId) return
  loading.value = true
  try {
    await cariStore.fetchCariByID(props.cariId)
    const c = cariStore.activeCari
    if (c) {
      form.value = {
        type: c.type || 'customer', group: c.group || '', code: c.code,
        name: c.name || '', contact_name: c.contact_name || '',
        tax_office: c.tax_office || '', tax_number: c.tax_number || '',
        email: c.email || '', phone: c.phone || '',
        landline: c.landline || '', fax: c.fax || '',
        address: c.address || '', city: c.city || '',
        district: c.district || '', postal_code: c.postal_code || '',
        country: c.country || 'Türkiye',
        shipping_address: c.shipping_address || '',
        shipping_city: c.shipping_city || '',
        shipping_district: c.shipping_district || '',
        shipping_postal_code: c.shipping_postal_code || '',
        shipping_country: c.shipping_country || 'Türkiye',
        currency: c.currency || 'TRY', note: c.note || '',
      }
    }
  } catch { errorMsg.value = 'Cari bilgileri yüklenemedi.' }
  finally  { loading.value = false }
}

onMounted(async () => {
  try {
    const gVal = await settingsStore.fetchSetting('cari_groups')
    if (gVal) {
      groupOptions.value = JSON.parse(gVal).map((g: string) => ({ label: g, value: g }))
    }
    await currencyStore.fetchCurrencies()
    if (currencyStore.currencies.length > 0) {
      currencyOptions.value = currencyStore.currencies.map(c => ({
        label: `${c.name} (${c.code})`, value: c.code,
      }))
    }
  } catch {}

  if (isEdit.value) {
    await loadCari()
  }
  // Yeni cari kartında kod alanı boş bırakılır: gerçek kod kaydetme anında
  // backend'de üretilir (önceden çekilen bir "önizleme" kodu kaydetme anına
  // kadar başka bir cari tarafından alınabildiğinden çakışmaya yol açıyordu).
})

const handleSubmit = async () => {
  if (!form.value.name) {
    errorMsg.value = 'Firma adı zorunludur.'
    return
  }
  if (sameAsBilling.value) copyBillingToShipping()
  loading.value  = true
  errorMsg.value = ''
  invalidFields.value = []
  try {
    const payload = { ...form.value }
    // Yeni cari kartında code alanı sadece önizleme amaçlı dolduruldu (next-code API).
    // Gerçek kod kaydetme anında backend'de üretilir; aksi halde önizleme alındıktan
    // sonra başka bir kayıt aynı kodu almışsa "kod zaten kullanımda" çakışması oluşur.
    if (!isEdit.value) payload.code = ''
    if (payload.name)              payload.name              = payload.name.toUpperCase()
    if (payload.contact_name)      payload.contact_name      = payload.contact_name.toUpperCase()
    if (payload.tax_office)        payload.tax_office        = payload.tax_office.toUpperCase()
    if (payload.city)              payload.city              = payload.city.toUpperCase()
    if (payload.district)          payload.district          = payload.district.toUpperCase()
    if (payload.shipping_city)     payload.shipping_city     = payload.shipping_city.toUpperCase()
    if (payload.shipping_district) payload.shipping_district = payload.shipping_district.toUpperCase()

    if (isEdit.value) {
      await cariStore.updateCari(props.cariId!, payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Cari kart güncellendi', life: 10000 })
    } else {
      await cariStore.createCari(payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Cari kart oluşturuldu', life: 10000 })
    }
    emit('saved')
    close()
  } catch (err: any) {
    if (err.response?.data?.error) {
      const code = err.response.data.error.code
      if (code === 'DUPLICATE_TAX_NUMBER') {
        errorMsg.value = 'Bu VN/TCKN başka bir cariye aittir.'
        invalidFields.value.push('tax_number')
      } else if (code === 'DUPLICATE_PHONE') {
        errorMsg.value = 'Bu telefon numarası başka bir cariye aittir.'
        invalidFields.value.push('phone')
      } else {
        errorMsg.value = err.response.data.error.message
      }
    } else {
      errorMsg.value = 'İşlem sırasında bir hata oluştu.'
    }
  } finally { loading.value = false }
}
</script>

<template>
  <Dialog
    :visible="props.visible"
    @update:visible="close"
    :modal="true"
    :style="{ width: '96%', maxWidth: '1080px' }"
    :draggable="false"
  >
    <!-- HEADER -->
    <template #header>
      <div class="cf-header">
        <span class="cf-title">{{ isEdit ? 'Cari Kartı Düzenle' : 'Yeni Cari Kart' }}</span>
        <div class="cf-code-group">
          <span class="cf-code-lbl">Cari Kodu</span>
          <InputText v-model="form.code" disabled placeholder="Otomatik Üretilecek" class="cf-code-box" />
        </div>
      </div>
    </template>

    <!-- BODY -->
    <div class="cf-body">
      <Message v-if="errorMsg" severity="error" class="mb-4">{{ errorMsg }}</Message>

      <!-- ─── BÖLÜM 1 : FİRMA KİMLİĞİ ─── -->
      <div class="cf-section">
        <div class="cf-section-title"><i class="pi pi-building"></i> Firma Kimliği</div>
        <div class="cf-row">
          <div class="cf-field" style="min-width:180px; flex:1.2">
            <label class="cf-lbl">Cari Tipi <em>*</em></label>
            <Select v-model="form.type" :options="typeOptions" optionLabel="label" optionValue="value"
              :disabled="loading" size="small" class="w-full" />
          </div>
          <div class="cf-field" style="min-width:180px; flex:1.2">
            <label class="cf-lbl">Cari Grubu</label>
            <Select v-model="form.group" :options="groupOptions" optionLabel="label" optionValue="value"
              showClear :disabled="loading" placeholder="Grup seçin" size="small" class="w-full" />
          </div>
          <div class="cf-field" style="min-width:160px; flex:1">
            <label class="cf-lbl">Para Birimi</label>
            <Select v-model="form.currency" :options="currencyOptions" optionLabel="label" optionValue="value"
              :disabled="isEdit || loading" size="small" class="w-full" />
          </div>
        </div>

        <div class="cf-row">
          <div class="cf-field" style="flex:2">
            <label class="cf-lbl">Firma Adı <em>*</em></label>
            <InputText v-model="form.name" placeholder="Ör: GLOBAL COMPANY A.Ş." maxlength="255"
              :disabled="loading" class="w-full uppercase-input"
              :class="{ 'p-invalid': invalidFields.includes('name') }" />
          </div>
          <div class="cf-field" style="flex:1.5">
            <label class="cf-lbl">Yetkili Ad Soyad</label>
            <InputText v-model="form.contact_name" placeholder="Ör: AHMET YILMAZ" maxlength="255"
              :disabled="loading" class="w-full uppercase-input" />
          </div>
        </div>

        <div class="cf-row">
          <div class="cf-field" style="flex:1.5">
            <label class="cf-lbl">Vergi Dairesi</label>
            <InputText v-model="form.tax_office" placeholder="Ör: BAŞKENT VERGİ DAİRESİ" maxlength="255"
              :disabled="loading" class="w-full uppercase-input" />
          </div>
          <div class="cf-field" style="min-width:160px; flex:0.8">
            <label class="cf-lbl">VN / TCKN</label>
            <InputText v-model="form.tax_number" placeholder="10 veya 11 haneli" maxlength="50"
              :disabled="loading" class="w-full"
              :class="{ 'p-invalid': invalidFields.includes('tax_number') }" />
          </div>
        </div>
      </div>

      <!-- ─── BÖLÜM 2 : İLETİŞİM ─── -->
      <div class="cf-section">
        <div class="cf-section-title"><i class="pi pi-phone"></i> İletişim</div>
        <div class="cf-row">
          <div class="cf-field" style="flex:1.2; min-width:200px">
            <label class="cf-lbl">Cep Telefonu</label>
            <InputText v-model="form.phone" placeholder="+90 5xx xxx xx xx" maxlength="50"
              :disabled="loading" class="w-full"
              :class="{ 'p-invalid': invalidFields.includes('phone') }" />
          </div>
          <div class="cf-field" style="flex:1.2; min-width:200px">
            <label class="cf-lbl">Sabit Telefon</label>
            <InputText v-model="form.landline" placeholder="+90 2xx xxx xx xx" maxlength="50"
              :disabled="loading" class="w-full" />
          </div>
          <div class="cf-field" style="flex:1; min-width:180px">
            <label class="cf-lbl">Faks</label>
            <InputText v-model="form.fax" placeholder="+90 2xx xxx xx xx" maxlength="50"
              :disabled="loading" class="w-full" />
          </div>
          <div class="cf-field" style="flex:1.8; min-width:240px">
            <label class="cf-lbl">E-posta</label>
            <InputText v-model="form.email" type="email" placeholder="eposta@firma.com" maxlength="255"
              :disabled="loading" class="w-full" />
          </div>
        </div>
      </div>

      <!-- ─── BÖLÜM 3 : ADRESLER ─── -->
      <div class="cf-section">
        <div class="cf-addr-grid">

          <!-- Fatura Adresi -->
          <div>
            <div class="cf-section-title"><i class="pi pi-map-marker"></i> Fatura Adresi</div>
            <div class="cf-stack">
              <div class="cf-field">
                <label class="cf-lbl">Ülke</label>
                <Select v-model="form.country" :options="countries" filter placeholder="Ülke seçin"
                  :disabled="loading" size="small" class="w-full" />
              </div>
              <div class="cf-row">
                <div class="cf-field" style="flex:1">
                  <label class="cf-lbl">İl</label>
                  <InputText v-model="form.city" placeholder="Ör: İSTANBUL" maxlength="255"
                    :disabled="loading" class="w-full uppercase-input" />
                </div>
                <div class="cf-field" style="flex:1">
                  <label class="cf-lbl">İlçe</label>
                  <InputText v-model="form.district" placeholder="Ör: KADIKÖY" maxlength="255"
                    :disabled="loading" class="w-full uppercase-input" />
                </div>
                <div class="cf-field" style="flex:0.6; min-width:110px">
                  <label class="cf-lbl">Posta Kodu</label>
                  <InputText v-model="form.postal_code" placeholder="34000" maxlength="20"
                    :disabled="loading" class="w-full" />
                </div>
              </div>
              <div class="cf-field">
                <label class="cf-lbl">Adres</label>
                <Textarea v-model="form.address" :rows="4"
                  placeholder="Mahalle, sokak, bina no, daire no..."
                  maxlength="2000" :disabled="loading" class="w-full" />
              </div>
            </div>
          </div>

          <!-- Sevk Adresi -->
          <div>
            <div class="cf-addr-shipping-header">
              <div class="cf-section-title"><i class="pi pi-truck"></i> Sevk Adresi</div>
              <div class="cf-copy-bar">
                <label class="cf-chk">
                  <input type="checkbox" v-model="sameAsBilling" />
                  <span>Fatura adresiyle aynı</span>
                </label>
                <Button type="button" class="cf-copy-btn"
                  @click="copyBillingToShipping"
                  :disabled="sameAsBilling || loading">
                  <i class="pi pi-copy"></i> Kopyala
                </button>
              </div>
            </div>
            <div class="cf-stack">
              <div class="cf-field">
                <label class="cf-lbl">Ülke</label>
                <Select v-model="form.shipping_country" :options="countries" filter placeholder="Ülke seçin"
                  :disabled="sameAsBilling || loading" size="small" class="w-full" />
              </div>
              <div class="cf-row">
                <div class="cf-field" style="flex:1">
                  <label class="cf-lbl">İl</label>
                  <InputText v-model="form.shipping_city" placeholder="Ör: İSTANBUL" maxlength="255"
                    :disabled="sameAsBilling || loading" class="w-full uppercase-input" />
                </div>
                <div class="cf-field" style="flex:1">
                  <label class="cf-lbl">İlçe</label>
                  <InputText v-model="form.shipping_district" placeholder="Ör: KADIKÖY" maxlength="255"
                    :disabled="sameAsBilling || loading" class="w-full uppercase-input" />
                </div>
                <div class="cf-field" style="flex:0.6; min-width:110px">
                  <label class="cf-lbl">Posta Kodu</label>
                  <InputText v-model="form.shipping_postal_code" placeholder="34000" maxlength="20"
                    :disabled="sameAsBilling || loading" class="w-full" />
                </div>
              </div>
              <div class="cf-field">
                <label class="cf-lbl">Adres</label>
                <Textarea v-model="form.shipping_address" :rows="4"
                  placeholder="Mahalle, sokak, bina no, daire no..."
                  maxlength="2000" :disabled="sameAsBilling || loading" class="w-full" />
              </div>
            </div>
          </div>

        </div><!-- /cf-addr-grid -->
      </div>

      <!-- ─── BÖLÜM 4 : DAHİLİ NOTLAR ─── -->
      <div class="cf-section">
        <div class="cf-section-title"><i class="pi pi-comment"></i> Dahili Notlar</div>
        <div class="cf-field">
          <Textarea v-model="form.note" :rows="3"
            placeholder="Bu cari hakkında dahili notlar, özel koşullar, hatırlatmalar..."
            maxlength="2000" :disabled="loading" class="w-full" />
        </div>
      </div>

    </div><!-- /cf-body -->

    <!-- FOOTER -->
    <template #footer>
      <div class="cf-footer">
        <Button label="Vazgeç" text @click="close" :disabled="loading" />
        <Button label="Kaydet" icon="pi pi-check" outlined @click="handleSubmit" :loading="loading" severity="primary" />
      </div>
    </template>
  </Dialog>
</template>

<style scoped>
/* ============================
   CARI FORM — yatay tasarım
   ============================ */

/* HEADER */
.cf-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 2.5rem;
  gap: 1rem;
}
.cf-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: #1a202c;
  white-space: nowrap;
}
:root.p-dark .cf-title { color: #f1f5f9; }

.cf-code-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}
.cf-code-lbl {
  font-size: 0.75rem;
  font-weight: 600;
  color: #718096;
  white-space: nowrap;
}
.cf-code-box {
  width: 9.5rem;
  font-size: 0.85rem !important;
  font-weight: 700 !important;
  padding: 0.35rem 0.6rem !important;
  background: #f7f8fb !important;
}

/* BODY */
.cf-body {
  padding: 0 1.5rem 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* SECTION */
.cf-section {
  padding: 1rem 0 0.75rem;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.cf-section:last-child { border-bottom: none; }
:root.p-dark .cf-section { border-bottom-color: rgba(255,255,255,0.05); }

.cf-section-title {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: #06b6d4;
}
.cf-section-title .pi { font-size: 0.72rem; }

/* ROW: alanlar yan yana, doğal genişlikte */
.cf-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  align-items: flex-start;
}

/* STACK: dikey sıralama (adres içi) */
.cf-stack {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

/* FIELD */
.cf-field {
  display: flex;
  flex-direction: column;
  gap: 0.28rem;
}

/* LABEL */
.cf-lbl {
  font-size: 0.72rem;
  font-weight: 700;
  color: #4a5568;
  letter-spacing: 0.01em;
  white-space: nowrap;
}
:root.p-dark .cf-lbl { color: #94a3b8; }
.cf-lbl em { color: #e53e3e; font-style: normal; margin-left: 2px; }

/* Address 2-column grid */
.cf-addr-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 2rem;
  align-items: start;
}
@media (max-width: 700px) {
  .cf-addr-grid { grid-template-columns: 1fr; }
}

.cf-addr-shipping-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

/* Copy bar */
.cf-copy-bar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.cf-chk {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #4a5568;
  cursor: pointer;
  user-select: none;
}
:root.p-dark .cf-chk { color: #94a3b8; }
.cf-copy-btn {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  font-weight: 600;
  color: #06b6d4;
  background: none;
  border: 1px solid rgba(6,182,212,0.3);
  border-radius: 0.3rem;
  padding: 0.2rem 0.55rem;
  cursor: pointer;
  transition: background 0.15s;
}
.cf-copy-btn:hover { background: rgba(6,182,212,0.07); }
.cf-copy-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.cf-copy-btn .pi { font-size: 0.68rem; }

/* FOOTER */
.cf-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 0.9rem 1.5rem;
  border-top: 1px solid #f0f2f5;
  background: #fafbfc;
  border-radius: 0 0 0.75rem 0.75rem;
}
:root.p-dark .cf-footer {
  border-top-color: rgba(255,255,255,0.06);
  background: #1e293b;
}
</style>
