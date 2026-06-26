<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { usePaymentStore } from '@/stores/payment'
import { useCariStore } from '@/stores/cari'
import { useInvoiceStore } from '@/stores/invoice'
import { useToast } from 'primevue/usetoast'
import { getCurrentCompanyDatetimeLocal, toBackendDate, toCompanyDatetimeLocal } from '@/utils/date'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Money from '@/components/Money.vue'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
  type: {
    type: String, // 'collection' or 'payment'
    required: true,
  },
})

const emit = defineEmits(['update:visible', 'saved'])

const paymentStore = usePaymentStore()
const cariStore = useCariStore()
const invoiceStore = useInvoiceStore()
const toast = useToast()

const loading = ref(false)
const errorMsg = ref('')

const form = ref({
  cari_id: '',
  type: props.type,
  date: '',
  method: 'cash',
  account_kind: 'cash',
  account_id: '',
  amount: 0,
  currency: 'TRY',
  invoice_id: null as string | null,
  reference: '',
  note: '',
})

const methodOptions = ref([
  { label: 'Nakit', value: 'cash' },
  { label: 'Havale/EFT', value: 'bank' },
  { label: 'Kredi Kartı', value: 'card' },
  { label: 'Çek', value: 'check' },
])

// Filter caris based on transaction type
const filteredCaris = computed(() => {
  return cariStore.caris.filter(c => {
    if (props.type === 'collection') {
      return c.type === 'customer' || c.type === 'both'
    } else {
      return c.type === 'supplier' || c.type === 'both'
    }
  })
})

// Group accounts by kind for a single dropdown
const groupedAccounts = computed(() => {
  const groups = []
  
  if (paymentStore.cashAccounts.length > 0) {
    groups.push({
      label: 'Kasa Hesapları (Nakit)',
      items: paymentStore.cashAccounts.map(a => ({
        ...a,
        displayLabel: `${a.name} (${a.currency})`,
        _kind: 'cash',
        _method: 'cash'
      }))
    })
  }
  
  if (paymentStore.bankAccounts.length > 0) {
    groups.push({
      label: 'Banka Hesapları (Havale/EFT)',
      items: paymentStore.bankAccounts.map(a => ({
        ...a,
        displayLabel: `${a.name} (${a.currency})`,
        _kind: 'bank',
        _method: 'bank'
      }))
    })
  }
  
  return groups
})

// Filter accounts strictly by invoice currency if an invoice is selected
const filteredGroupedAccounts = computed(() => {
  const groups = []
  let activeCurrency = ''
  
  if (form.value.invoice_id) {
    const inv = invoiceStore.invoices.find(i => i.id === form.value.invoice_id)
    if (inv) {
      activeCurrency = inv.currency
    }
  }

  const filteredCash = paymentStore.cashAccounts
    .filter(a => !activeCurrency || a.currency === activeCurrency)
    .map(a => ({
      ...a,
      displayLabel: `${a.name} (${a.currency})`,
      _kind: 'cash',
      _method: 'cash'
    }))
    
  if (filteredCash.length > 0) {
    groups.push({
      label: 'Kasa Hesapları (Nakit)',
      items: filteredCash
    })
  }
  
  const filteredBank = paymentStore.bankAccounts
    .filter(a => !activeCurrency || a.currency === activeCurrency)
    .map(a => ({
      ...a,
      displayLabel: `${a.name} (${a.currency})`,
      _kind: 'bank',
      _method: 'bank'
    }))
    
  if (filteredBank.length > 0) {
    groups.push({
      label: 'Banka Hesapları (Havale/EFT)',
      items: filteredBank
    })
  }
  
  return groups
})

// Filter unpaid invoices of the selected Cari for allocation
const unpaidInvoices = computed(() => {
  if (!form.value.cari_id) return []
  // Match invoices by selected Cari ID, and unpaid status
  return invoiceStore.invoices.filter(i => {
    const isCari = i.cari_id === form.value.cari_id
    const isUnpaid = i.status !== 'paid' && i.status !== 'canceled'
    const isMatchingType = props.type === 'collection' ? i.type === 'sales' : i.type === 'purchase'
    return isCari && isUnpaid && isMatchingType
  })
})

onMounted(async () => {
  form.value.date = getCurrentCompanyDatetimeLocal()
  
  loading.value = true
  try {
    await paymentStore.fetchAccounts()
    await cariStore.fetchCaris({ page: 1, limit: 1000 })
    await invoiceStore.fetchInvoices({ page: 1, limit: 1000 })
    
    // Set default cash account if exists
    setDefaultAccount()
  } catch (err) {
    errorMsg.value = 'Hesap bilgileri yüklenemedi.'
  } finally {
    loading.value = false
  }
})

// Automatically set default account on load (first available)
const setDefaultAccount = () => {
  if (groupedAccounts.value.length > 0 && groupedAccounts.value[0].items.length > 0) {
    form.value.account_id = groupedAccounts.value[0].items[0].id
  }
}

// When account changes, automatically set method, kind, and currency!
watch(() => form.value.account_id, (newId) => {
  if (!newId) return
  let selected = null
  for (const group of groupedAccounts.value) {
    const found = group.items.find(i => i.id === newId)
    if (found) {
      selected = found
      break
    }
  }
  if (selected) {
    form.value.account_kind = selected._kind as any
    form.value.method = selected._method as any
    form.value.currency = selected.currency
    
    // Clear invoice only if selected invoice currency does not match the new account's currency
    if (form.value.invoice_id) {
      const inv = invoiceStore.invoices.find(i => i.id === form.value.invoice_id)
      if (inv && inv.currency !== selected.currency) {
        form.value.invoice_id = null
      }
    }
  }
})

// When Cari changes, automatically select the account matching Cari's default currency
watch(() => form.value.cari_id, (newCariId) => {
  form.value.invoice_id = null
  if (!newCariId) {
    form.value.account_id = ''
    return
  }
  
  const cari = cariStore.caris.find(c => c.id === newCariId)
  if (cari) {
    const targetCurrency = cari.currency || 'TRY'
    let foundAcc = null
    for (const group of groupedAccounts.value) {
      const match = group.items.find(a => a.currency === targetCurrency)
      if (match) {
        foundAcc = match
        break
      }
    }
    if (foundAcc) {
      form.value.account_id = foundAcc.id
      form.value.currency = foundAcc.currency
      form.value.account_kind = foundAcc._kind as any
      form.value.method = foundAcc._method as any
    } else {
      form.value.account_id = ''
    }
  } else {
    form.value.account_id = ''
  }
})

// Auto-populate invoice details, match currency, and auto-select matching account on selection
watch(() => form.value.invoice_id, (newInvoiceId) => {
  if (newInvoiceId) {
    const inv = invoiceStore.invoices.find(i => i.id === newInvoiceId)
    if (inv) {
      form.value.currency = inv.currency
      
      // Auto-select account matching invoice currency
      let foundAcc = null
      for (const group of groupedAccounts.value) {
        const match = group.items.find(a => a.currency === inv.currency)
        if (match) {
          foundAcc = match
          break
        }
      }
      if (foundAcc) {
        form.value.account_id = foundAcc.id
        form.value.account_kind = foundAcc._kind as any
        form.value.method = foundAcc._method as any
      } else {
        form.value.account_id = ''
        toast.add({
          severity: 'warn',
          summary: 'Hesap Bulunamadı',
          detail: `Fatura para birimi olan (${inv.currency}) ile eşleşen bir Kasa veya Banka hesabı bulunamadı!`,
          life: 10000
        })
      }

      const remaining = parseFloat(inv.total || '0') - (parseFloat(inv.paid_total || '0') || 0)
      form.value.amount = remaining > 0 ? remaining : 0
      form.value.note = `${inv.number} nolu faturanın ${props.type === 'collection' ? 'tahsilatı' : 'ödemesi'}`
    }
  } else {
    // Fallback to Cari currency default account when invoice is cleared
    if (form.value.cari_id) {
      const cari = cariStore.caris.find(c => c.id === form.value.cari_id)
      if (cari) {
        const targetCurrency = cari.currency || 'TRY'
        let foundAcc = null
        for (const group of groupedAccounts.value) {
          const match = group.items.find(a => a.currency === targetCurrency)
          if (match) {
            foundAcc = match
            break
          }
        }
        if (foundAcc) {
          form.value.account_id = foundAcc.id
          form.value.currency = foundAcc.currency
          form.value.account_kind = foundAcc._kind as any
          form.value.method = foundAcc._method as any
        } else {
          form.value.account_id = ''
        }
      }
    }
  }
})

const close = () => {
  emit('update:visible', false)
}

const handleSubmit = async () => {
  if (!form.value.cari_id) {
    errorMsg.value = 'Lütfen bir Cari Hesap seçin.'
    return
  }
  if (!form.value.account_id) {
    errorMsg.value = 'Lütfen işlem yapılacak Kasa veya Banka hesabını seçin.'
    return
  }
  if (form.value.amount <= 0) {
    errorMsg.value = 'İşlem tutarı sıfırdan büyük olmalıdır.'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const payload = {
      ...form.value,
      date: toBackendDate(form.value.date),
      amount: form.value.amount.toString(), // string format for GORM money type
    }
    
    await paymentStore.createPayment(payload)
    toast.add({
      severity: 'success',
      summary: 'Başarılı',
      detail: props.type === 'collection' ? 'Tahsilat başarıyla kaydedildi' : 'Ödeme başarıyla kaydedildi',
      life: 10000
    })
    
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
    :header="props.type === 'collection' ? 'Yeni Tahsilat (Giriş) Kaydet' : 'Yeni Ödeme (Çıkış) Kaydet'"
    :modal="true"
    :style="{ width: '95vw', maxWidth: '800px' }"
  >
    <Message v-if="errorMsg" severity="error" class="mb-4">{{ errorMsg }}</Message>

    <form @submit.prevent="handleSubmit" class="form-grid">
      <!-- Row 1 -->
      <div class="field-col col-12">
        <label for="cari">Cari Hesap *</label>
        <Select
          id="cari"
          v-model="form.cari_id"
          :options="filteredCaris"
          optionLabel="name"
          optionValue="id"
          placeholder="Cari seçin..."
          class="w-full"
          filter
          :disabled="loading"
        />
      </div>

      <!-- Row 2 (İlişkili Fatura - Kasa/Banka üstüne taşındı) -->
      <div class="field-col col-12" v-if="unpaidInvoices.length > 0">
        <label for="invoice">İlişkili Fatura (Kapatılacak Fatura)</label>
        <Select
          id="invoice"
          v-model="form.invoice_id"
          :options="unpaidInvoices"
          optionLabel="number"
          optionValue="id"
          placeholder="İsteğe bağlı fatura kapatma..."
          class="w-full"
          showClear
          :disabled="loading || !form.cari_id"
        >
          <template #option="{ option }">
            <div class="flex justify-between w-full text-xs">
              <span class="font-bold">{{ option.number }}</span>
              <span>
                Kalan: <Money :value="(parseFloat(option.total || '0') - (parseFloat(option.paid_total || '0') || 0)).toString()" :currency="option.currency" />
              </span>
            </div>
          </template>
        </Select>
      </div>

      <!-- Row 3 (Tarih ve Kasa/Banka) -->
      <div class="field-col md:col-6">
        <label for="date">İşlem Tarihi *</label>
        <input
          id="date"
          type="datetime-local"
          v-model="form.date"
          class="p-inputtext w-full"
          required
          :disabled="loading"
        />
      </div>

      <div class="field-col md:col-6">
        <label for="account">{{ props.type === 'collection' ? 'Tahsilatın' : 'Ödemenin' }} Yapılacağı Kasa / Banka *</label>
        <Select
          id="account"
          v-model="form.account_id"
          :options="filteredGroupedAccounts"
          optionLabel="displayLabel"
          optionValue="id"
          optionGroupLabel="label"
          optionGroupChildren="items"
          placeholder="Kasa veya Banka seçin..."
          class="w-full"
          :disabled="loading || !form.cari_id"
        />
        <p v-if="filteredGroupedAccounts.length === 0" class="text-xs text-red-500 mt-1">
          * Lütfen önce Ayarlar altından Kasa veya Banka hesabı tanımlayın.
        </p>
      </div>

      <!-- Row 5 -->
      <div class="field-col col-12">
        <label for="amount">İşlem Tutarı ({{ form.currency }}) *</label>
        <InputNumber
          id="amount"
          v-model="form.amount"
          class="w-full"
          mode="decimal"
          :minFractionDigits="2"
          :maxFractionDigits="2"
          :disabled="loading"
        />
      </div>

      <!-- Row 6 -->
      <div class="field-col col-12">
        <label for="reference">Referans No / Dekont / Belge No</label>
        <InputText
          id="reference"
          v-model="form.reference"
          placeholder="Ör: DEK-98321"
          class="w-full"
          :disabled="loading"
          maxlength="100"
        />
      </div>

      <!-- Row 7 -->
      <div class="field-col col-12">
        <label for="note">Açıklama / Dahili Notlar</label>
        <Textarea
          id="note"
          v-model="form.note"
          rows="2"
          class="w-full"
          placeholder="Bu ödeme/tahsilat işlemi ile ilgili notlar..."
          :disabled="loading"
          maxlength="2000"
        />
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
.form-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 1rem;
  padding: 0.5rem 0;
}

.field-col {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.col-12 {
  grid-column: span 12;
}

.md\:col-6 {
  grid-column: span 12;
}

@media (min-width: 768px) {
  .md\:col-6 {
    grid-column: span 6;
  }
}

.field-col label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--p-text-color, #334155);
}

.footer-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
