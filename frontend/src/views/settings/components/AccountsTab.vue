<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { usePaymentStore } from '@/stores/payment'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import Money from '@/components/Money.vue'
import Tag from 'primevue/tag'

import { ALL_CURRENCIES } from '@/utils/currencies'
import { useCurrencyStore } from '@/stores/currency'

const paymentStore = usePaymentStore()
const currencyStore = useCurrencyStore()
const toast = useToast()

const accountTypes = ref([
  { label: 'Kasa', value: 'cash' },
  { label: 'Banka', value: 'bank' }
])

const showAddDialog = ref(false)

const form = ref({
  type: 'cash',
  code: '',
  name: '',
  account_no: '',
  description: '',
  iban: '',
  currency: 'TRY',
  is_default: false
})

const editingAccountId = ref<string | null>(null)
const editingAccountType = ref<'cash' | 'bank' | null>(null)

const editingAccount = ref({
  code: '',
  name: '',
  account_no: '',
  description: '',
  iban: '',
  currency: 'TRY',
  is_default: false
})

const currencies = computed(() => {
  if (currencyStore.currencies.length === 0) return ALL_CURRENCIES
  return currencyStore.currencies.map((c: any) => ({
    label: `${c.name} (${c.code})`,
    value: c.code
  }))
})

onMounted(async () => {
  await Promise.all([
    paymentStore.fetchAccounts(),
    currencyStore.fetchCurrencies()
  ])
})

const combinedAccounts = computed(() => {
  const cash = paymentStore.cashAccounts.map((a: any) => ({ ...a, type: 'cash' }))
  const bank = paymentStore.bankAccounts.map((a: any) => ({ ...a, type: 'bank' }))
  return [...cash, ...bank]
})

const handleCreateAccount = async () => {
  if (!form.value.name.trim()) return
  try {
    if (form.value.type === 'cash') {
      await paymentStore.createCashAccount({
        code: form.value.code,
        name: form.value.name,
        account_no: form.value.account_no,
        description: form.value.description,
        currency: form.value.currency,
        is_default: form.value.is_default
      })
    } else {
      await paymentStore.createBankAccount({
        code: form.value.code,
        name: form.value.name,
        account_no: form.value.account_no,
        description: form.value.description,
        currency: form.value.currency,
        iban: form.value.iban
      })
    }
    form.value = { type: form.value.type, code: '', name: '', account_no: '', description: '', iban: '', currency: 'TRY', is_default: false }
    showAddDialog.value = false
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Hesap oluşturuldu', life: 10000 })
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'İşlem gerçekleştirilemedi', life: 10000 })
  }
}

const startEditAccount = (acc: any) => {
  editingAccountId.value = acc.id
  editingAccountType.value = acc.type
  editingAccount.value = {
    code: acc.code || '',
    name: acc.name,
    account_no: acc.account_no || '',
    description: acc.description || '',
    iban: acc.iban || '',
    currency: acc.currency,
    is_default: acc.is_default || false
  }
}

const handleUpdateAccount = async () => {
  if (!editingAccountId.value || !editingAccountType.value || !editingAccount.value.name.trim()) return
  try {
    if (editingAccountType.value === 'cash') {
      await paymentStore.updateCashAccount(editingAccountId.value, {
        code: editingAccount.value.code,
        name: editingAccount.value.name,
        account_no: editingAccount.value.account_no,
        description: editingAccount.value.description,
        currency: editingAccount.value.currency,
        is_default: editingAccount.value.is_default
      })
    } else {
      await paymentStore.updateBankAccount(editingAccountId.value, {
        code: editingAccount.value.code,
        name: editingAccount.value.name,
        account_no: editingAccount.value.account_no,
        description: editingAccount.value.description,
        currency: editingAccount.value.currency,
        iban: editingAccount.value.iban
      })
    }
    editingAccountId.value = null
    editingAccountType.value = null
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Hesap güncellendi', life: 10000 })
  } catch (err: any) {
    if (err.response?.data?.error?.message) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'İşlem gerçekleştirilemedi', life: 10000 })
    }
  }
}

const handleDeleteAccount = async (id: string, type: 'cash'|'bank') => {
  if (confirm('Bu hesabı silmek istediğinize emin misiniz?')) {
    try {
      if (type === 'cash') {
        await paymentStore.deleteCashAccount(id)
      } else {
        await paymentStore.deleteBankAccount(id)
      }
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Hesap silindi', life: 10000 })
    } catch (err: any) {
      if (err.response?.data?.error?.message) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Bu hesap silinemez. Hareket görmüş olabilir.', life: 10000 })
      }
    }
  }
}
</script>

<template>
  <div class="accounts-tab">
    <div class="mb-4 flex justify-between items-center">
      <div>
        <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-2">Hesap Tanımları</h3>
        <p class="text-sm text-slate-500">Nakit ve Banka işlemlerinizi takip edeceğiniz hesapları tanımlayın.</p>
      </div>
      <Button label="Yeni Hesap Ekle" icon="pi pi-plus" @click="showAddDialog = true" severity="success" />
    </div>

    <!-- ADD DIALOG -->
    <Dialog v-model:visible="showAddDialog" modal header="Yeni Hesap Ekle" :style="{ width: '95vw', maxWidth: '800px' }">
      <div class="flex flex-col gap-4 mt-2">
        <div class="flex flex-col gap-2">
          <label class="font-medium text-sm text-slate-700">Hesap Türü</label>
          <Select v-model="form.type" :options="accountTypes" optionLabel="label" optionValue="value" class="w-full" />
        </div>
        <div class="flex flex-col gap-2">
          <label class="font-medium text-sm text-slate-700">Kodu</label>
          <InputText v-model="form.code" placeholder="Otomatik atanması için boş bırakın" class="w-full" />
        </div>
        <div class="flex flex-col gap-2">
          <label class="font-medium text-sm text-slate-700">{{ form.type === 'cash' ? 'Kasa Adı' : 'Banka ve Şube Adı' }}</label>
          <InputText v-model="form.name" :placeholder="form.type === 'cash' ? 'Örn: Merkez Kasa' : 'Örn: Garanti BBVA Maslak'" class="w-full" />
        </div>
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex flex-col gap-2 flex-1">
            <label class="font-medium text-sm text-slate-700">Hesap No</label>
            <InputText v-model="form.account_no" class="w-full" />
          </div>
          <div class="flex flex-col gap-2 flex-1">
            <label class="font-medium text-sm text-slate-700">Para Birimi</label>
            <Select v-model="form.currency" :options="currencies" optionLabel="label" optionValue="value" class="w-full" filter />
          </div>
        </div>
        <div class="flex flex-col gap-2" v-if="form.type === 'bank'">
          <label class="font-medium text-sm text-slate-700">IBAN</label>
          <InputText v-model="form.iban" class="w-full" />
        </div>
        <div class="flex flex-col gap-2">
          <label class="font-medium text-sm text-slate-700">Açıklama</label>
          <InputText v-model="form.description" class="w-full" />
        </div>
        <div v-if="form.type === 'cash'" class="flex items-center gap-2 mt-2">
          <Checkbox v-model="form.is_default" :binary="true" inputId="cash_default_new_modal" />
          <label for="cash_default_new_modal" class="text-sm text-slate-700 dark:text-slate-300">Varsayılan Kasa</label>
        </div>
      </div>
      <template #footer>
        <Button label="İptal" icon="pi pi-times" @click="showAddDialog = false" class="p-button-text" />
        <Button label="Ekle" icon="pi pi-check" @click="handleCreateAccount" :disabled="!form.name.trim()" autofocus severity="success" />
      </template>
    </Dialog>

    <DataTable :value="combinedAccounts" class="p-datatable-sm" responsiveLayout="scroll">
      <Column header="Tür" style="width: 100px">
        <template #body="{ data }">
          <Tag :value="data.type === 'cash' ? 'KASA' : 'BANKA'" :severity="data.type === 'cash' ? 'success' : 'info'" />
        </template>
      </Column>
      
      <Column field="code" header="Kodu" style="width: 100px">
        <template #body="{ data }">
          <div v-if="editingAccountId === data.id">
            <InputText v-model="editingAccount.code" class="w-full p-inputtext-sm" />
          </div>
          <span v-else>{{ data.code || '-' }}</span>
        </template>
      </Column>

      <Column field="name" header="Hesap Adı">
        <template #body="{ data }">
          <div v-if="editingAccountId === data.id">
            <InputText v-model="editingAccount.name" class="w-full p-inputtext-sm" />
          </div>
          <span v-else>{{ data.name }}</span>
        </template>
      </Column>

      <Column field="account_no" header="Hesap No" style="width: 150px">
        <template #body="{ data }">
          <div v-if="editingAccountId === data.id">
            <InputText v-model="editingAccount.account_no" class="w-full p-inputtext-sm" />
          </div>
          <span v-else>{{ data.account_no || '-' }}</span>
        </template>
      </Column>

      <Column field="iban" header="IBAN" style="width: 200px">
        <template #body="{ data }">
          <div v-if="editingAccountId === data.id && data.type === 'bank'">
            <InputText v-model="editingAccount.iban" class="w-full p-inputtext-sm" />
          </div>
          <span v-else-if="data.type === 'bank'">{{ data.iban || '-' }}</span>
          <span v-else class="text-slate-400">-</span>
        </template>
      </Column>

      <Column field="description" header="Açıklama" style="width: 200px">
        <template #body="{ data }">
          <div v-if="editingAccountId === data.id">
            <InputText v-model="editingAccount.description" class="w-full p-inputtext-sm" />
          </div>
          <span v-else>{{ data.description || '-' }}</span>
        </template>
      </Column>
      
      <Column field="currency" header="Para Birimi" style="width: 150px">
        <template #body="{ data }">
          <div v-if="editingAccountId === data.id">
            <Select v-model="editingAccount.currency" :options="currencies" optionLabel="label" optionValue="value" class="w-full p-inputtext-sm" filter />
          </div>
          <span v-else>{{ data.currency }}</span>
        </template>
      </Column>
      
      <Column field="balance" header="Bakiye" style="width: 150px; text-align: right;">
        <template #body="{ data }">
          <Money :value="data.balance" :currency="data.currency" />
        </template>
      </Column>

      <Column field="is_default" header="Varsayılan" style="width: 100px; text-align: center;">
        <template #body="{ data }">
          <div v-if="data.type === 'cash'">
            <div v-if="editingAccountId === data.id">
              <Checkbox v-model="editingAccount.is_default" :binary="true" />
            </div>
            <i v-else-if="data.is_default" class="pi pi-check text-green-500"></i>
            <span v-else>-</span>
          </div>
          <div v-else class="text-slate-400">-</div>
        </template>
      </Column>
      
      <Column header="İşlemler" style="width: 100px; text-align: center">
        <template #body="{ data }">
          <div class="flex gap-2 justify-center" v-if="editingAccountId === data.id">
            <Button icon="pi pi-check" class="p-button-rounded p-button-text p-button-sm" @click="handleUpdateAccount" severity="primary" />
            <Button icon="pi pi-times" class="p-button-rounded p-button-text p-button-sm" @click="editingAccountId = null; editingAccountType = null" severity="warn" />
          </div>
          <div class="flex gap-2 justify-center" v-else>
            <Button icon="pi pi-pencil" class="p-button-rounded p-button-text p-button-sm" @click="startEditAccount(data)" severity="warn" />
            <Button icon="pi pi-trash" class="p-button-rounded p-button-text p-button-sm" @click="handleDeleteAccount(data.id, data.type)" severity="danger" />
          </div>
        </template>
      </Column>
      
      <template #empty>
        <div class="text-center p-4 text-slate-500">
          Kayıtlı hesap bulunamadı.
        </div>
      </template>
    </DataTable>
  </div>
</template>
