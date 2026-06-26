<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useExpenseStore } from '@/stores/expense'
import { useCariStore } from '@/stores/cari'
import { usePaymentStore } from '@/stores/payment'
import { useSettingsStore } from '@/stores/settings'
import { useCurrencyStore } from '@/stores/currency'
import { getCurrentCompanyDatetimeLocal, toBackendDate } from '@/utils/date'
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
  expenseId: {
    type: String,
    default: null,
  },
})

const emit = defineEmits(['update:visible', 'saved'])

const expenseStore = useExpenseStore()
const cariStore = useCariStore()
const paymentStore = usePaymentStore()
const settingsStore = useSettingsStore()
const currencyStore = useCurrencyStore()
const toast = useToast()

const loading = ref(false)
const errorMsg = ref('')

const isTaxInclusive = ref(true) // Default tax-inclusive (vergiler dahil) flag

const form = ref({
  category_id: '',
  cari_id: null as string | null,
  date: '',
  description: '',
  currency: 'TRY',
  amount: 0, // Input amount entered by the user
  tax_rate: 20, // default 20%
  account_kind: 'cash' as 'cash' | 'bank' | null,
  account_id: null as string | null,
  status: 'paid' as 'paid' | 'unpaid' | 'ignored',
  note: '',
  is_recurring: false, // Her ay otomatik tekrarlansın seçeneği
})

const taxRateOptions = ref([
  { label: '%0', value: 0 },
  { label: '%1', value: 1 },
  { label: '%10', value: 10 },
  { label: '%20', value: 20 },
])

const currencyOptions = ref<{label: string, value: string}[]>([
  { label: 'TRY', value: 'TRY' },
  { label: 'USD', value: 'USD' },
  { label: 'EUR', value: 'EUR' },
  { label: 'GBP', value: 'GBP' }
])

const statusOptions = ref([
  { label: 'Ödendi', value: 'paid' },
  { label: 'Ödenmedi (Borç)', value: 'unpaid' },
  { label: 'Yoksay (İptal/Geçersiz)', value: 'ignored' },
])

// Filter caris: suppliers only
const suppliers = computed(() => {
  return cariStore.caris.filter(c => c.type === 'supplier' || c.type === 'both')
})

// Filter accounts list: combine both cash and bank accounts into a single list
const combinedAccounts = computed(() => {
  const cashList = paymentStore.cashAccounts.map(c => ({
      label: `${c.name} (Kasa - ${c.currency})`,
      id: c.id,
      kind: 'cash',
      balance: c.balance,
      currency: c.currency
  }))
  const bankList = paymentStore.bankAccounts.map(b => ({
      label: `${b.name} (Banka - ${b.currency})`,
      id: b.id,
      kind: 'bank',
      balance: b.balance,
      currency: b.currency
  }))
  return [...cashList, ...bankList]
})

// Active selected account balance and details
const selectedAccountDetail = computed(() => {
  if (!form.value.account_id) return null
  return combinedAccounts.value.find(a => a.id === form.value.account_id)
})

// Computed tax and totals
const calculatedValues = computed(() => {
  const amt = form.value.amount || 0
  const rate = form.value.tax_rate || 0

  if (isTaxInclusive.value) {
    // If tax inclusive: total is the input amount
    // amount_excluding_tax = total / (1 + tax_rate/100)
    const baseAmount = amt / (1 + rate / 100)
    const tax = amt - baseAmount
    return {
      baseAmount: parseFloat(baseAmount.toFixed(4)),
      taxAmount: parseFloat(tax.toFixed(4)),
      totalAmount: parseFloat(amt.toFixed(4))
    }
  } else {
    // If tax exclusive: input amount is base amount
    const tax = amt * (rate / 100)
    const total = amt + tax
    return {
      baseAmount: parseFloat(amt.toFixed(4)),
      taxAmount: parseFloat(tax.toFixed(4)),
      totalAmount: parseFloat(total.toFixed(4))
    }
  }
})

// Dynamic currency symbol for totals based on selected account or cari
const displayCurrency = computed(() => {
  if (form.value.status === 'paid' && selectedAccountDetail.value) {
    return selectedAccountDetail.value.currency
  }
  if (form.value.cari_id) {
    const cari = suppliers.value.find(c => c.id === form.value.cari_id)
    if (cari && cari.currency) {
      return cari.currency
    }
  }
  return 'TRY'
})

// Derived category
const selectedCategory = computed(() => {
  return expenseStore.categories.find(c => c.id === form.value.category_id)
})

const isRecurringDisabled = computed(() => {
  if (!selectedCategory.value) return false
  if (props.expenseId && form.value.is_recurring) return false // We are editing the recurring one itself
  return selectedCategory.value.has_active_recurring
})

watch(() => form.value.category_id, () => {
  if (isRecurringDisabled.value && form.value.is_recurring) {
    form.value.is_recurring = false
  }
})

onMounted(async () => {
  form.value.date = getCurrentCompanyDatetimeLocal()
  
  loading.value = true
  try {
    const kdvRatesStr = await settingsStore.fetchSetting('kdv_rates')
    if (kdvRatesStr) {
      try {
        const parsed = JSON.parse(kdvRatesStr)
        if (Array.isArray(parsed)) {
          taxRateOptions.value = parsed.map((r: number) => ({ label: `%${r}`, value: r }))
        }
      } catch (e) {}
    }

    await expenseStore.fetchCategories()
    await cariStore.fetchCaris({ page: 1, limit: 1000 })
    await paymentStore.fetchAccounts()
    await currencyStore.fetchCurrencies()

    if (currencyStore.currencies.length > 0) {
      currencyOptions.value = currencyStore.currencies.map(c => ({ label: c.code, value: c.code }))
    }

    if (props.expenseId) {
      const data = await expenseStore.getExpenseByID(props.expenseId)
      // Check if amount is actually tax-exclusive vs total on database load
      const dbAmount = parseFloat(data.amount) || 0
      const dbTotal = parseFloat(data.total) || 0
      const diff = Math.abs(dbTotal - dbAmount)

      // Set defaults on load based on saved DB record properties
      form.value = {
        category_id: data.category_id,
        cari_id: data.cari_id,
        date: data.date.substring(0, 16),
        description: data.description || '',
        currency: data.currency || 'TRY',
        amount: diff > 0.01 ? dbTotal : dbAmount, 
        tax_rate: parseFloat(data.tax_rate) || 0,
        account_kind: (data.account_kind || 'cash') as 'cash' | 'bank',
        account_id: data.account_id,
        status: data.status as 'paid' | 'unpaid' | 'ignored',
        note: data.note || '',
        is_recurring: data.is_recurring || false,
      }
      isTaxInclusive.value = diff > 0.01
    } else {
      setDefaultAccount()
    }
  } catch (err) {
    errorMsg.value = 'Form verileri yüklenemedi.'
  } finally {
    loading.value = false
  }
})

const setDefaultAccount = () => {
  if (combinedAccounts.value.length > 0) {
    const activeCurrency = form.value.currency || 'TRY'
    let def = combinedAccounts.value.find(a => {
      if (a.kind === 'cash') {
        const acc = paymentStore.cashAccounts.find(c => c.id === a.id)
        return acc && acc.is_default && acc.currency === activeCurrency
      }
      return false
    })
    
    if (!def) {
      def = combinedAccounts.value.find(a => a.currency === activeCurrency)
    }

    if (def) {
      form.value.account_id = def.id
      form.value.account_kind = def.kind as 'cash' | 'bank'
    } else {
      // Fallback if no account matches the currency
      form.value.account_id = combinedAccounts.value[0].id
      form.value.account_kind = combinedAccounts.value[0].kind as 'cash' | 'bank'
    }
  } else {
    form.value.account_id = null
    form.value.account_kind = null
  }
}

// When status changes to unpaid or ignored, clear account fields
watch(() => form.value.status, (newStatus) => {
  if (newStatus === 'unpaid' || newStatus === 'ignored') {
    form.value.account_kind = null
    form.value.account_id = null
  } else {
    setDefaultAccount()
  }
})

// Auto-select currency when cari is selected
watch(() => form.value.cari_id, (newCariId) => {
  if (newCariId) {
    const cari = suppliers.value.find(c => c.id === newCariId)
    if (cari && cari.currency) {
      form.value.currency = cari.currency
    }
  }
})

// Automatically re-select account when currency changes
watch(() => form.value.currency, () => {
  if (form.value.status === 'paid') {
    setDefaultAccount()
  }
})

// Update form.currency before submitting
watch(displayCurrency, (newVal) => {
  form.value.currency = newVal
})

const close = () => {
  emit('update:visible', false)
}

const handleSubmit = async () => {
  if (!form.value.category_id) {
    errorMsg.value = 'Lütfen bir gider kategorisi seçin.'
    return
  }
  if (form.value.amount <= 0) {
    errorMsg.value = 'Gider tutarı sıfırdan büyük olmalıdır.'
    return
  }
  if (form.value.status === 'paid' && !form.value.account_id) {
    errorMsg.value = 'Ödenen giderler için kasa/banka hesabı zorunludur.'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const calculated = calculatedValues.value
    const payload = {
      ...form.value,
      date: toBackendDate(form.value.date),
      // Base amount (without tax) is saved as amount, total is saved as total
      amount: calculated.baseAmount.toString(),
      tax_rate: form.value.tax_rate.toString(),
    }

    if (props.expenseId) {
      await expenseStore.updateExpense(props.expenseId, payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Gider kaydı güncellendi', life: 10000 })
    } else {
      await expenseStore.createExpense(payload)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Gider kaydı oluşturuldu', life: 10000 })
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
    :header="props.expenseId ? 'Gider Kaydını Düzenle' : 'Yeni Gider Fişi / Kaydı'"
    :modal="true"
    :style="{ width: '95%', maxWidth: '920px' }"
  >
    <Message v-if="errorMsg" severity="error" class="mb-4">{{ errorMsg }}</Message>

    <form @submit.prevent="handleSubmit">
      <div class="cari-form grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Left Column: Temel Gider Bilgileri -->
        <div class="form-col flex flex-col gap-3">
          <div class="section-title">Temel Gider Bilgileri</div>
          
          <div class="frow-vert">
            <label for="category">Gider Kategorisi *</label>
            <Select
              id="category"
              v-model="form.category_id"
              :options="expenseStore.categories"
              optionLabel="name"
              optionValue="id"
              placeholder="Kategori seçin..."
              class="w-full"
              filter
              :disabled="loading"
              size="small"
            />
          </div>

          <div class="frow-vert">
            <label for="cari">Tedarikçi (Cari Hesap)</label>
            <Select
              id="cari"
              v-model="form.cari_id"
              :options="suppliers"
              optionLabel="name"
              optionValue="id"
              placeholder="İsteğe bağlı tedarikçi seçin..."
              class="w-full"
              filter
              showClear
              :disabled="loading"
              size="small"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="frow-vert">
              <label for="date">Belge Tarihi *</label>
              <input
                id="date"
                type="datetime-local"
                v-model="form.date"
                class="p-inputtext w-full p-inputtext-sm"
                required
                :disabled="loading"
              />
            </div>
            <div class="frow-vert">
              <label for="description">Açıklama / Detay</label>
              <InputText
                id="description"
                v-model="form.description"
                placeholder="Ör: Yemek faturası"
                class="w-full p-inputtext-sm"
                maxlength="2000"
                :disabled="loading"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="frow-vert">
              <div class="flex justify-between items-center">
                <label for="amount">Tutar *</label>
                <div class="flex items-center gap-1 cursor-pointer" @click="isTaxInclusive = !isTaxInclusive">
                  <span class="text-[10px] font-bold text-slate-500" :class="isTaxInclusive ? 'text-green-600 dark:text-green-400' : 'text-slate-400'">KDV Dahil</span>
                  <i class="pi" :class="isTaxInclusive ? 'pi-check-circle text-green-600' : 'pi-circle text-slate-400'" style="font-size: 0.8rem"></i>
                </div>
              </div>
              <InputNumber
                id="amount"
                v-model="form.amount"
                class="w-full"
                mode="decimal"
                :minFractionDigits="2"
                :maxFractionDigits="2"
                :disabled="loading"
                size="small"
              />
            </div>
            <div class="frow-vert">
              <label for="tax_rate">KDV Oranı</label>
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
        </div>

        <!-- Right Column: Ödeme & Hesap Bilgileri -->
        <div class="form-col flex flex-col gap-3">
          <div class="section-title">Hesap & Ödeme Bilgileri</div>

          <!-- Calculated Totals Info -->
          <div class="totals-summary p-3 rounded-lg border border-dashed border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/30 text-xs">
            <div class="flex justify-between text-slate-500 dark:text-slate-400">
              <span>KDV Hariç Tutar:</span>
              <span>{{ calculatedValues.baseAmount.toFixed(2) }} {{ displayCurrency }}</span>
            </div>
            <div class="flex justify-between text-slate-500 dark:text-slate-400 mt-1">
              <span>KDV Tutarı (%{{ form.tax_rate }}):</span>
              <span>{{ calculatedValues.taxAmount.toFixed(2) }} {{ displayCurrency }}</span>
            </div>
            <div class="flex justify-between text-sm font-bold border-t border-slate-200 dark:border-slate-800 pt-2 mt-2">
              <span>Genel Toplam:</span>
              <span class="text-slate-800 dark:text-slate-100">{{ calculatedValues.totalAmount.toFixed(2) }} {{ displayCurrency }}</span>
            </div>
          </div>

          <!-- Recurring switch -->
          <div class="bg-sky-50/40 dark:bg-sky-950/10 p-2.5 rounded border border-sky-100/50 dark:border-sky-900/30 flex justify-between items-center text-xs">
            <div>
              <span class="font-bold text-slate-800 dark:text-slate-200 block">Tekrarlayan Fatura</span>
              <span class="text-[10px] text-slate-500 block mt-0.5">Her ay bu tarihte otomatik tekrarlanır.</span>
            </div>
            <div class="cursor-pointer" @click="!isRecurringDisabled && (form.is_recurring = !form.is_recurring)">
              <i class="pi" :class="[form.is_recurring ? 'pi-check-circle text-sky-600 text-lg' : 'pi-circle text-slate-400 text-lg', isRecurringDisabled ? 'opacity-50 cursor-not-allowed' : '']"></i>
            </div>
          </div>
          <p v-if="isRecurringDisabled" class="text-red-500 text-[10px] font-semibold">
            * Bu kategoride aktif bir tekrarlayan fiş bulunuyor.
          </p>

          <div class="grid grid-cols-2 gap-3">
            <div class="frow-vert">
              <label for="status">Ödeme Durumu *</label>
              <Select
                id="status"
                v-model="form.status"
                :options="statusOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                :disabled="loading"
                size="small"
              />
            </div>

            <div class="frow-vert">
              <label for="currency">Para Birimi</label>
              <Select
                id="currency"
                v-model="form.currency"
                :options="currencyOptions"
                optionLabel="label"
                optionValue="value"
                class="w-full"
                :disabled="loading"
                size="small"
              />
            </div>
          </div>

          <!-- Combined Account Selection -->
          <div class="frow-vert" v-if="form.status === 'paid'">
            <label for="account">Ödeme Yapılan Hesap *</label>
            <Select
              id="account"
              v-model="form.account_id"
              :options="combinedAccounts"
              optionLabel="label"
              optionValue="id"
              placeholder="Kasa veya Banka Seçin..."
              class="w-full"
              filter
              @change="(e) => {
                const acc = combinedAccounts.find(a => a.id === e.value);
                if (acc) form.account_kind = acc.kind as 'cash' | 'bank';
              }"
              :disabled="loading"
              size="small"
            />
            <!-- Balance Warning Info -->
            <div v-if="selectedAccountDetail" class="mt-1 flex justify-between items-center text-[10px]">
              <span class="text-slate-500">Mevcut Bakiye:</span>
              <span class="font-bold" :class="(selectedAccountDetail.currency === 'TRY' && parseFloat(selectedAccountDetail.balance) < calculatedValues.totalAmount) ? 'text-red-500' : 'text-slate-700 dark:text-slate-300'">
                {{ parseFloat(selectedAccountDetail.balance).toFixed(2) }} {{ selectedAccountDetail.currency }}
                <span v-if="selectedAccountDetail.currency === 'TRY' && parseFloat(selectedAccountDetail.balance) < calculatedValues.totalAmount" class="text-red-500 font-normal">(Yetersiz)</span>
              </span>
            </div>
          </div>

          <!-- Note -->
          <div class="frow-vert">
            <label for="note">Notlar</label>
            <Textarea
              id="note"
              v-model="form.note"
              rows="1"
              class="w-full p-textarea-sm"
              placeholder="İsteğe bağlı not..."
              maxlength="2000"
              :disabled="loading"
            />
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

.totals-summary {
  background-color: var(--p-content-background, #f8fafc);
  padding: 0.75rem 1rem;
  border-radius: 8px;
  border: 1px dashed var(--p-content-border-color, #cbd5e1);
}

:root.p-dark .totals-summary {
  background-color: #0f172a;
  border-color: #334155;
}

.footer-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
