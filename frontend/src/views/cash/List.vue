<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { usePaymentStore } from '@/stores/payment'

import { useCariStore } from '@/stores/cari'
import { useInvoiceStore } from '@/stores/invoice'
import { getCurrentCompanyDatetimeLocal, toBackendDate } from '@/utils/date'
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
import FormModal from '../payment/FormModal.vue'
import { exportToPDF } from '@/utils/pdfExport'

const paymentStore = usePaymentStore()
const cariStore = useCariStore()
const invoiceStore = useInvoiceStore()
const toast = useToast()

const showPaymentModal = ref(false)
const paymentType = ref<'collection' | 'payment'>('collection')

// Transfer (Virman) state
const showTransferModal = ref(false)
const showTransferConfirmModal = ref(false)
const transferForm = ref({
  from_account_key: '', // e.g. "cash:id"
  to_account_key: '',
  amount: 0, // Kaynak Tutar
  to_amount: 0, // Hedef Tutar
  exchange_rate: 1.0, // Manuel Kur
  exchange_rate_op: '*', // Matematiksel İşlem Operatörü: '*', '/', '+', '-'
  date: getCurrentCompanyDatetimeLocal(),
  description: ''
})

// Logs state
const showLogsModal = ref(false)
const activeLogAccount = ref<{ kind: string; name: string } | null>(null)

// Search & filters
const searchQuery = ref('')
const selectedType = ref('')
const selectedStatus = ref('')
const first = ref(0)
const rows = ref(20)
const page = ref(1)
const dt = ref()
const selectedItems = ref([])

const typeOptions = ref([
  { label: 'Tüm Türler', value: '' },
  { label: 'Tahsilat (Giriş)', value: 'collection' },
  { label: 'Ödeme (Çıkış)', value: 'payment' },
])

const statusOptions = ref([
  { label: 'Tüm Durumlar', value: '' },
  { label: 'Tamamlandı', value: 'completed' },
  { label: 'İptal Edildi', value: 'canceled' },
])

const loadData = async () => {
  const params: any = {
    page: page.value,
    limit: rows.value,
    q: searchQuery.value,
    type: selectedType.value,
    status: selectedStatus.value,
  }
  await paymentStore.fetchPayments(params)
  await paymentStore.fetchAccounts()
}

onMounted(async () => {
  loadData()
  await cariStore.fetchCaris({ page: 1, limit: 1000 })
  await invoiceStore.fetchInvoices({ page: 1, limit: 1000 })
})

watch([searchQuery, selectedType, selectedStatus], () => {
  page.value = 1
  first.value = 0
  loadData()
})

const onPage = (event: any) => {
  page.value = event.page + 1
  rows.value = event.rows
  first.value = event.first
  loadData()
}

const openCollection = () => {
  paymentType.value = 'collection'
  showPaymentModal.value = true
}

const openPayment = () => {
  paymentType.value = 'payment'
  showPaymentModal.value = true
}

const cancelPaymentItem = async (id: string) => {
  if (confirm('Bu ödeme/tahsilat işlemini iptal etmek istediğinize emin misiniz? Cari bakiye, fatura tahsilat miktarı ve kasa durumları ters hareketle düzeltilecektir.')) {
    try {
      await paymentStore.cancelPayment(id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'İşlem iptal edildi', life: 10000 })
      loadData()
    } catch (err: any) {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'İptal işlemi gerçekleştirilemedi', life: 10000 })
    }
  }
}

// Virman (Transfer) Handlers
const openTransfer = () => {
  transferForm.value = {
    from_account_key: '',
    to_account_key: '',
    amount: 0,
    to_amount: 0,
    exchange_rate: 1.0,
    exchange_rate_op: '*',
    date: getCurrentCompanyDatetimeLocal(),
    description: ''
  }
  showTransferModal.value = true
}

const combinedAccounts = computed(() => {
  const cashList = paymentStore.cashAccounts.map(c => ({
    label: `${c.name} (Kasa - ${c.currency})`,
    value: `cash:${c.id}`,
    kind: 'cash',
    id: c.id,
    currency: c.currency,
    name: c.name,
    iban: '-',
    is_default: c.is_default,
    balance: c.balance
  }))
  const bankList = paymentStore.bankAccounts.map(b => ({
    label: `${b.name} (Banka - ${b.currency})`,
    value: `bank:${b.id}`,
    kind: 'bank',
    id: b.id,
    currency: b.currency,
    name: b.name,
    iban: b.iban || '-',
    is_default: false,
    balance: b.balance
  }))
  return [...cashList, ...bankList]
})

const fromAccount = computed(() => combinedAccounts.value.find(o => o.value === transferForm.value.from_account_key))
const toAccount = computed(() => combinedAccounts.value.find(o => o.value === transferForm.value.to_account_key))

const isCrossCurrency = computed(() => {
  return fromAccount.value && toAccount.value && fromAccount.value.currency !== toAccount.value.currency
})

const rateLabelText = computed(() => {
  if (!isCrossCurrency.value || !fromAccount.value || !toAccount.value) return 'Kur'
  if (fromAccount.value.currency === 'TRY') {
    return `1 ${toAccount.value.currency} = ? TRY (Kur)`
  }
  return `1 ${fromAccount.value.currency} = ? ${toAccount.value.currency} (Kur)`
})

const rateFormulaHint = computed(() => {
  if (!isCrossCurrency.value || !fromAccount.value || !toAccount.value) return ''
  const op = transferForm.value.exchange_rate_op || '*'
  return `${fromAccount.value.currency} ${op} Kur = ${toAccount.value.currency}`
})

const calcToAmount = () => {
  if (!isCrossCurrency.value) {
    transferForm.value.to_amount = transferForm.value.amount
    return
  }
  const rate = transferForm.value.exchange_rate
  if (rate === null || rate === undefined || rate <= 0) {
    transferForm.value.to_amount = 0
    return
  }
  const op = transferForm.value.exchange_rate_op || '*'
  
  let targetVal = 0
  if (op === '/') {
    if (rate === 0) return
    targetVal = transferForm.value.amount / rate
  } else if (op === '+') {
    targetVal = transferForm.value.amount + rate
  } else if (op === '-') {
    targetVal = transferForm.value.amount - rate
  } else {
    // default multiplication '*'
    targetVal = transferForm.value.amount * rate
  }

  if (targetVal < 0) {
    transferForm.value.to_amount = 0
  } else {
    transferForm.value.to_amount = parseFloat(targetVal.toFixed(4))
  }
}

// Auto-populate bakiye when source account is selected
watch(() => transferForm.value.from_account_key, (newKey) => {
  if (newKey) {
    const acc = combinedAccounts.value.find(o => o.value === newKey)
    if (acc) {
      transferForm.value.amount = parseFloat(acc.balance) || 0
      
      // Auto-set suggested operator based on TRY direction as helper
      if (acc.currency === 'TRY') {
        transferForm.value.exchange_rate_op = '/'
      } else {
        transferForm.value.exchange_rate_op = '*'
      }
      calcToAmount()
    }
  }
})

// Recalculate target to_amount when amount, exchange_rate or exchange_rate_op changes
watch([() => transferForm.value.amount, () => transferForm.value.exchange_rate, () => transferForm.value.exchange_rate_op], calcToAmount)

// Recalculate target to_amount when target account changes (cross-currency state changes)
watch(() => transferForm.value.to_account_key, calcToAmount)

const requestTransfer = () => {
  if (!transferForm.value.from_account_key || !transferForm.value.to_account_key) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Lütfen kaynak ve hedef hesapları seçin', life: 10000 })
    return
  }

  if (!transferForm.value.amount || transferForm.value.amount <= 0) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Transfer tutarı sıfırdan büyük olmalıdır', life: 10000 })
    return
  }

  const fromOption = combinedAccounts.value.find(o => o.value === transferForm.value.from_account_key)
  const toOption = combinedAccounts.value.find(o => o.value === transferForm.value.to_account_key)

  if (!fromOption || !toOption) return

  if (fromOption.value === toOption.value) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Aynı hesaba transfer yapılamaz.', life: 10000 })
    return
  }

  // Open confirmation modal
  showTransferConfirmModal.value = true
}

const handleTransfer = async () => {
  showTransferConfirmModal.value = false

  const fromOption = combinedAccounts.value.find(o => o.value === transferForm.value.from_account_key)
  const toOption = combinedAccounts.value.find(o => o.value === transferForm.value.to_account_key)

  if (!fromOption || !toOption) return

  try {
    const payload = {
      from_kind: fromOption.kind,
      from_id: fromOption.id,
      to_kind: toOption.kind,
      to_id: toOption.id,
      amount: transferForm.value.amount.toString(),
      to_amount: (fromOption.currency !== toOption.currency) ? transferForm.value.to_amount.toString() : transferForm.value.amount.toString(),
      exchange_rate: (fromOption.currency !== toOption.currency) ? transferForm.value.exchange_rate.toString() : '1.0',
      exchange_rate_op: (fromOption.currency !== toOption.currency) ? transferForm.value.exchange_rate_op : '*',
      date: toBackendDate(transferForm.value.date),
      description: transferForm.value.description
    }

    await paymentStore.transferCash(payload)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Transfer başarıyla gerçekleştirildi', life: 10000 })
    showTransferModal.value = false
    loadData()
  } catch (err: any) {
    if (err.response && err.response.data && err.response.data.error) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Transfer gerçekleştirilemedi', life: 10000 })
    }
  }
}

// Log history Dialog Handlers
const viewAccountLogs = async (kind: 'cash' | 'bank', acc: any) => {
  activeLogAccount.value = { kind, name: acc.name }
  try {
    await paymentStore.fetchCashTransactions({
      account_kind: kind,
      account_id: acc.id
    })
    showLogsModal.value = true
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Hesap hareketleri yüklenemedi', life: 10000 })
  }
}

// Format helpers
const getCariName = (cariId: string) => {
  const cari = cariStore.caris.find(c => c.id === cariId)
  return cari ? cari.name : 'Yükleniyor...'
}

const getAccountName = (kind: string, id: string) => {
  if (kind === 'cash') {
    const acc = paymentStore.cashAccounts.find(a => a.id === id)
    return acc ? `${acc.name} (Kasa)` : 'Kasa Hesabı'
  } else {
    const acc = paymentStore.bankAccounts.find(a => a.id === id)
    return acc ? `${acc.name} (Banka)` : 'Banka Hesabı'
  }
}

const getPaymentTypeLabel = (type: string) => {
  return type === 'collection' ? 'Tahsilat' : 'Ödeme'
}

const getPaymentTypeSeverity = (type: string) => {
  return type === 'collection' ? 'success' : 'warn'
}

const getStatusLabel = (status: string) => {
  return status === 'completed' ? 'Tamamlandı' : 'İptal Edildi'
}

const getStatusSeverity = (status: string) => {
  return status === 'completed' ? 'success' : 'danger'
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('tr-TR')
}

const getSourceTypeLabel = (src: string) => {
  switch (src) {
    case 'payment': return 'Cari Ödeme'
    case 'expense': return 'Gider Ödemesi'
    case 'transfer': return 'Virman Transfer'
    case 'manual': return 'Manuel Harek'
    default: return src
  }
}

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
    const dataToExport = selectedItems.value.length > 0 ? selectedItems.value : combinedAccounts.value
    const columns = [
      { header: 'Kasa/Banka Adı', dataKey: 'name' },
      { header: 'Tür', dataKey: 'kind' },
      { header: 'Hesap No', dataKey: 'account_no' },
      { header: 'IBAN', dataKey: 'iban' },
      { header: 'Bakiye', dataKey: 'balance' },
      { header: 'Döviz', dataKey: 'currency' }
    ]
    exportToPDF('Kasa_Banka_Hesaplari_Listesi', columns, dataToExport.map(item => ({
      ...item,
      kind: item.kind === 'cash' ? 'Nakit Kasa' : 'Banka'
    })))
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'PDF olarak dışa aktarıldı', life: 10000 })
  }
}
const exportCSV = () => {
  exportData('excel')
}
</script>

<template>
  <div class="payments-list-container">
    <!-- Header -->
    <div class="header-section flex justify-between items-center">
      <div class="flex gap-2">
        <Button label="Transfer (Virman)" icon="pi pi-arrows-h" class="p-button-secondary" @click="openTransfer" />
      </div>
      <div class="flex gap-2">
        <Button label="Aktar" icon="pi pi-upload" class="p-button-outlined" @click="toggleExportMenu" aria-haspopup="true" aria-controls="export_menu" severity="contrast" />
        <Menu ref="exportMenu" id="export_menu" :model="exportOptions" :popup="true" />
      </div>
    </div>

    <!-- Accounts DataTable -->
    <Card class="table-card">
      <template #title>
        <div class="px-4 pt-4 pb-2 border-b border-slate-100 dark:border-slate-800">
          <span class="text-md font-medium text-slate-700 dark:text-slate-200">Kasa ve Banka Hesapları</span>
        </div>
      </template>
      <template #content>
        <DataTable
          ref="dt"
          :value="combinedAccounts"
          v-model:selection="selectedItems"
          class="p-datatable-sm"
          responsiveLayout="scroll"
          dataKey="value"
        >
          <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
            <template #body="slotProps">
              {{ slotProps.index + 1 }}
            </template>
          </Column>
          <Column selectionMode="multiple" headerStyle="width: 3rem"></Column>
          <Column field="code" header="Kodu" sortable style="min-width: 120px; width: 10%">
            <template #body="{ data }">
              <span class="text-sm text-slate-600 dark:text-slate-400">{{ data.code || '-' }}</span>
            </template>
          </Column>
          <Column field="kind" header="Tür" sortable style="min-width: 120px; width: 10%">
            <template #body="{ data }">
              <Tag :value="data.kind === 'cash' ? 'Nakit Kasa' : 'Banka'" :severity="data.kind === 'cash' ? 'success' : 'info'" />
            </template>
          </Column>
          <Column field="name" header="Kasa Adı" sortable style="min-width: 200px; width: 25%">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200 truncate max-w-[180px] inline-block align-middle" :title="data.name">
                {{ data.name }}
              </span>
              <Tag v-if="data.is_default" value="Varsayılan" severity="secondary" class="text-[10px] ml-2 px-1.5 py-0.5 align-middle" />
            </template>
          </Column>
          <Column field="account_no" header="Hesap No" sortable style="min-width: 150px; width: 15%">
            <template #body="{ data }">
              <span class="text-sm text-slate-600 dark:text-slate-400">{{ data.account_no || '-' }}</span>
            </template>
          </Column>
          <Column field="iban" header="IBAN" sortable style="min-width: 200px; width: 15%">
            <template #body="{ data }">
              <span class="tabular-nums text-sm text-slate-600 dark:text-slate-400">{{ data.iban || '-' }}</span>
            </template>
          </Column>
          <Column field="description" header="Açıklama" sortable style="min-width: 150px; width: 10%">
            <template #body="{ data }">
              <span class="text-sm text-slate-600 dark:text-slate-400">{{ data.description || '-' }}</span>
            </template>
          </Column>
          <Column field="balance" header="Bakiye" sortable style="min-width: 150px; width: 10%" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200">
                <Money :value="data.balance" :currency="data.currency" />
              </span>
            </template>
          </Column>
          <Column header="İşlemler" style="min-width: 100px; width: 5%" headerClass="text-center" bodyClass="text-center">
            <template #body="{ data }">
              <Button icon="pi pi-history" class="p-button-text p-button-secondary p-button-sm" @click="viewAccountLogs(data.kind, data)" v-tooltip.top="'Hesap Hareketleri'" />
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Cash Transactions Logs Dialog stays the same -->

    <!-- Transfer (Virman) Dialog -->
    <Dialog v-model:visible="showTransferModal" header="Hesaplar Arası Transfer (Virman)" :modal="true" :style="{ width: '90%', maxWidth: '500px' }">
      <div class="flex flex-col gap-4 py-2">
        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Kaynak Hesap *</label>
          <Select
            v-model="transferForm.from_account_key"
            :options="combinedAccounts"
            optionLabel="label"
            optionValue="value"
            placeholder="Gönderen hesabı seçin..."
            class="w-full"
            filter
          />
        </div>

        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Hedef Hesap *</label>
          <Select
            v-model="transferForm.to_account_key"
            :options="combinedAccounts"
            optionLabel="label"
            optionValue="value"
            placeholder="Alıcı hesabı seçin..."
            class="w-full"
            filter
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-semibold">Gönderilen (Miktar) *</label>
            <InputNumber v-model="transferForm.amount" mode="decimal" :minFractionDigits="2" class="w-full" />
          </div>

          <div class="flex flex-col gap-1">
            <label class="text-xs font-semibold">Tarih *</label>
            <InputText type="datetime-local" v-model="transferForm.date" class="w-full" />
          </div>
        </div>

        <!-- Cross Currency Fields (Kur, İşlem Operatörü ve Alınan) -->
        <div v-if="isCrossCurrency" class="flex flex-col gap-3 border p-3 rounded bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700">
          <div class="grid grid-cols-2 gap-3">
            <div class="flex flex-col gap-1">
              <label class="text-xs font-semibold text-cyan-600 dark:text-cyan-400">{{ rateLabelText }} *</label>
              <InputNumber v-model="transferForm.exchange_rate" mode="decimal" :minFractionDigits="4" :maxFractionDigits="6" class="w-full" />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-xs font-semibold text-cyan-600 dark:text-cyan-400">İşlem İşareti *</label>
              <Select
                v-model="transferForm.exchange_rate_op"
                :options="[
                  { label: 'Çarp (×)', value: '*' },
                  { label: 'Böl (÷)', value: '/' },
                  { label: 'Topla (+)', value: '+' },
                  { label: 'Çıkar (-)', value: '-' }
                ]"
                optionLabel="label"
                optionValue="value"
                placeholder="İşlem Seçin"
                class="w-full"
              />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3 mt-1">
            <div class="flex flex-col gap-1">
              <label class="text-xs font-semibold text-cyan-600 dark:text-cyan-400">Formül Önizleme</label>
              <span class="text-xs text-slate-500 font-medium mt-1.5">{{ rateFormulaHint }}</span>
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-xs font-semibold text-cyan-600 dark:text-cyan-400">Alınan (Hedef Tutar) *</label>
              <InputNumber v-model="transferForm.to_amount" mode="decimal" :minFractionDigits="2" class="w-full" />
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-1">
          <label class="text-xs font-semibold">Açıklama</label>
          <InputText v-model="transferForm.description" placeholder="Ör: Günlük kasa devri, virman..." class="w-full" />
        </div>

        <div class="flex justify-end gap-2 mt-4">
          <Button label="İptal" class="p-button-text p-button-secondary" @click="showTransferModal = false" outlined />
          <Button label="Transferi Gerçekleştir" icon="pi pi-check" @click="requestTransfer" outlined severity="primary" />
        </div>
      </div>
    </Dialog>

    <!-- Confirmation & Summary Info Dialog -->
    <Dialog v-model:visible="showTransferConfirmModal" header="Transfer İşlemi Onayı" :modal="true" :style="{ width: '90%', maxWidth: '450px' }">
      <div class="flex flex-col gap-4 py-2">
        <div class="flex items-start gap-3 bg-amber-50 dark:bg-amber-950/30 p-3 rounded border border-amber-200 dark:border-amber-900/50">
          <i class="pi pi-exclamation-triangle text-amber-500 text-xl mt-0.5"></i>
          <div>
            <h4 class="font-bold text-amber-800 dark:text-amber-300 text-sm mb-1">Virman Özeti ve Onay</h4>
            <p class="text-xs text-amber-700 dark:text-amber-400 leading-normal">
              <strong>{{ fromAccount?.name }}</strong> hesabından <strong>{{ transferForm.amount }} {{ fromAccount?.currency }}</strong> çıkış yapılacaktır.
            </p>
            <p v-if="isCrossCurrency" class="text-xs text-amber-700 dark:text-amber-400 leading-normal mt-1">
              <strong>{{ transferForm.exchange_rate }}</strong> kuru ve <strong>"{{ transferForm.exchange_rate_op === '*' ? 'Çarp (×)' : transferForm.exchange_rate_op === '/' ? 'Böl (÷)' : transferForm.exchange_rate_op === '+' ? 'Topla (+)' : 'Çıkar (-)' }}"</strong> işlemiyle hesaplanarak, 
              <strong>{{ toAccount?.name }}</strong> hesabına <strong>{{ transferForm.to_amount }} {{ toAccount?.currency }}</strong> giriş yapılacaktır.
            </p>
            <p v-else class="text-xs text-amber-700 dark:text-amber-400 leading-normal mt-1">
              Bire bir kur ile <strong>{{ toAccount?.name }}</strong> hesabına <strong>{{ transferForm.amount }} {{ toAccount?.currency }}</strong> giriş yapılacaktır.
            </p>
          </div>
        </div>
        <p class="text-sm text-slate-600 dark:text-slate-400">Bu işlemi onaylıyor musunuz?</p>
        <div class="flex justify-end gap-2 mt-2">
          <Button label="İptal" class="p-button-text p-button-secondary" @click="showTransferConfirmModal = false" outlined />
          <Button label="Onayla ve Transfer Et" icon="pi pi-check" @click="handleTransfer" outlined severity="primary" />
        </div>
      </div>
    </Dialog>

    <!-- Cash Transaction Logs history Dialog -->
    <Dialog
      v-model:visible="showLogsModal"
      :header="activeLogAccount ? `${activeLogAccount.name} - Hesap Hareketleri Logu` : 'Kasa/Banka Hesap Hareketleri'"
      :modal="true"
      :style="{ width: '90%', maxWidth: '850px' }"
    >
      <div v-if="paymentStore.loading" class="text-center py-8">
        <i class="pi pi-spin pi-spinner text-2xl mb-2"></i>
        <div>Yükleniyor...</div>
      </div>
      <div v-else>
        <DataTable
          :value="paymentStore.transactions"
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
              <span class="text-xs">{{ formatDate(data.date) }}</span>
            </template>
          </Column>
          <Column field="type" header="Tür" style="min-width: 120px">
            <template #body="{ data }">
              <Tag :value="data.type === 'in' ? 'Giriş (+)' : 'Çıkış (-)'" :severity="data.type === 'in' ? 'success' : 'danger'" class="text-xs" />
            </template>
          </Column>
          <Column field="amount" header="Tutar" sortable style="min-width: 120px" headerClass="col-right" bodyClass="col-right">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-200 text-xs">
                <Money :value="data.amount" />
              </span>
            </template>
          </Column>
          <Column field="balance_after" header="Bakiye Sonrası" style="min-width: 150px">
            <template #body="{ data }">
              <span class="font-medium text-slate-700 dark:text-slate-300 text-xs">
                <Money :value="data.balance_after" />
              </span>
            </template>
          </Column>
          <Column field="source_type" header="Kaynak" style="min-width: 150px">
            <template #body="{ data }">
              <Tag :value="getSourceTypeLabel(data.source_type)" severity="info" class="text-xs" />
            </template>
          </Column>
          <Column field="description" header="Açıklama" style="min-width: 280px">
            <template #body="{ data }">
              <span class="text-xs text-slate-500">{{ data.description }}</span>
            </template>
          </Column>
        </DataTable>
      </div>
    </Dialog>

    <!-- Collection/Payment Creation Form Modal -->
    <FormModal
      v-if="showPaymentModal"
      v-model:visible="showPaymentModal"
      :type="paymentType"
      @saved="loadData"
    />
  </div>
</template>

<style scoped>
.payments-list-container {
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

.account-summary-card {
  border-radius: 12px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
}

:root.p-dark .account-summary-card {
  border-color: #334155;
  background-color: #1e293b;
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
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.search-input {
  position: relative;
  flex: 1;
  min-width: 280px;
  max-width: 400px;
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

.select-filters {
  display: flex;
  gap: 0.5rem;
}

.type-select, .status-select {
  width: 160px;
}

.actions-cell {
  display: flex;
  justify-content: center;
  gap: 0.25rem;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.field-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--p-text-color, #475569);
}

:root.p-dark .field-group label {
  color: #cbd5e1;
}

.grid {
  display: grid;
}

.grid-cols-1 {
  grid-template-columns: repeat(1, 1fr);
}

.gap-6 {
  gap: 1.5rem;
}

.space-y-3 > * + * {
  margin-top: 0.75rem;
}


/* Helper to override PrimeVue Button paddings when nested */
.p-button-sm {
  font-size: 0.8rem;
}
</style>

