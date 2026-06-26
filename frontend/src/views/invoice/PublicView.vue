<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getPublicInvoiceApi, disputePublicInvoiceApi, payPublicInvoiceApi } from '@/api/invoice'
import { useToast } from 'primevue/usetoast'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Textarea from 'primevue/textarea'
import Tag from 'primevue/tag'
import { formatDate } from '@/utils/date'
import Money from '@/components/Money.vue'

const route = useRoute()
const toast = useToast()
const token = route.params.token as string

const loading = ref(true)
const invoice = ref<any>(null)
const notFound = ref(false)

const showDisputeModal = ref(false)
const disputeNote = ref('')
const disputeLoading = ref(false)

const loadInvoice = async () => {
  try {
    loading.value = true
    const res = await getPublicInvoiceApi(token)
    invoice.value = res.data.data
  } catch (err: any) {
    if (err.response?.status === 404) {
      notFound.value = true
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Fatura bilgileri yüklenemedi.', life: 10000 })
    }
  } finally {
    loading.value = false
  }
}

const openDispute = () => {
  disputeNote.value = ''
  showDisputeModal.value = true
}

const submitDispute = async () => {
  if (!disputeNote.value) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'İtiraz notu zorunludur.', life: 10000 })
    return
  }
  try {
    disputeLoading.value = true
    await disputePublicInvoiceApi(token, disputeNote.value)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'İtirazınız iletildi.', life: 10000 })
    showDisputeModal.value = false
    await loadInvoice()
  } catch (err: any) {
    if (err.response && err.response.data && err.response.data.error) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'İtiraz işlemi başarısız.', life: 10000 })
    }
  } finally {
    disputeLoading.value = false
  }
}

const payLoading = ref(false)

const payInvoice = async () => {
  try {
    payLoading.value = true
    await payPublicInvoiceApi(token)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Ödemeniz başarıyla alındı.', life: 10000 })
    await loadInvoice()
  } catch (err: any) {
    if (err.response && err.response.data && err.response.data.error) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
    } else {
      toast.add({ severity: 'error', summary: 'Hata', detail: 'Ödeme işlemi başarısız.', life: 10000 })
    }
  } finally {
    payLoading.value = false
  }
}

const getStatusLabel = (status: string) => {
  switch (status) {
    case 'draft': return 'Taslak'
    case 'sent': return 'Ödeme Bekliyor'
    case 'disputed': return 'İtiraz Edildi'
    case 'partial': return 'Kısmi Ödendi'
    case 'paid': return 'Ödendi'
    case 'canceled': return 'İptal Edildi'
    default: return status
  }
}

const getStatusSeverity = (status: string) => {
  switch (status) {
    case 'draft': return 'secondary'
    case 'sent': return 'info'
    case 'disputed': return 'danger'
    case 'partial': return 'warn'
    case 'paid': return 'success'
    case 'canceled': return 'danger'
    default: return 'secondary'
  }
}

onMounted(() => {
  if (token) {
    loadInvoice()
  } else {
    notFound.value = true
    loading.value = false
  }
})
</script>

<template>
  <div class="public-view-container min-h-screen bg-slate-50 dark:bg-slate-900 flex items-center justify-center p-4">
    <div v-if="loading" class="text-center p-8">
      <i class="pi pi-spin pi-spinner text-4xl text-primary-500 mb-4"></i>
      <div class="text-slate-600 dark:text-slate-400">Yükleniyor...</div>
    </div>

    <div v-else-if="notFound" class="text-center p-8 bg-white dark:bg-slate-800 rounded-xl shadow-lg max-w-md w-full">
      <i class="pi pi-exclamation-circle text-5xl text-red-500 mb-4"></i>
      <h2 class="text-xl font-bold mb-2">Fatura Bulunamadı</h2>
      <p class="text-slate-500">Bu bağlantı geçersiz veya fatura silinmiş olabilir.</p>
    </div>

    <div v-else-if="invoice" class="w-full max-w-3xl">
      <Card class="invoice-card shadow-xl border-0 overflow-hidden">
        <template #content>
          <!-- Header -->
          <div class="flex flex-col md:flex-row justify-between items-start md:items-center border-b border-slate-200 dark:border-slate-700 pb-6 mb-6">
            <div>
              <h1 class="text-2xl font-bold text-slate-800 dark:text-slate-100 mb-1">
                Fatura <span class="text-primary-600">#{{ invoice.number }}</span>
              </h1>
              <div class="text-slate-500 text-sm">Düzenlenme: {{ formatDate(invoice.date) }}</div>
              <div class="text-slate-500 text-sm font-semibold mt-1">Son Ödeme: {{ formatDate(invoice.due_date) }}</div>
            </div>
            <div class="mt-4 md:mt-0 text-right">
              <Tag :value="getStatusLabel(invoice.status)" :severity="getStatusSeverity(invoice.status)" class="text-sm px-3 py-1" />
            </div>
          </div>

          <!-- Customer Info -->
          <div class="mb-8 p-4 bg-slate-50 dark:bg-slate-800/50 rounded-lg">
            <h3 class="text-sm font-bold text-slate-400 uppercase tracking-wider mb-2">Sayın</h3>
            <div class="font-bold text-lg text-slate-800 dark:text-slate-100">{{ invoice.cari?.name || 'Müşteri' }}</div>
          </div>

          <!-- Items Table -->
          <div class="overflow-x-auto mb-8">
            <table class="w-full text-left text-sm text-slate-600 dark:text-slate-300">
              <thead class="bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200">
                <tr>
                  <th class="p-3 rounded-tl-lg">Açıklama</th>
                  <th class="p-3">Miktar</th>
                  <th class="p-3">Birim Fiyat</th>
                  <th class="p-3 text-right rounded-tr-lg">Toplam</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in invoice.items" :key="item.id" class="border-b border-slate-100 dark:border-slate-800 last:border-0">
                  <td class="p-3">{{ item.description }}</td>
                  <td class="p-3">{{ item.quantity }} {{ item.unit }}</td>
                  <td class="p-3"><Money :value="item.unit_price" :currency="invoice.currency" /></td>
                  <td class="p-3 text-right font-medium"><Money :value="item.line_total" :currency="invoice.currency" /></td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Totals -->
          <div class="flex justify-end mb-8">
            <div class="w-full max-w-sm space-y-3">
              <div class="flex justify-between text-slate-600 dark:text-slate-400">
                <span>Ara Toplam:</span>
                <Money :value="invoice.subtotal" :currency="invoice.currency" />
              </div>
              <div v-if="parseFloat(invoice.discount_total) > 0" class="flex justify-between text-slate-600 dark:text-slate-400">
                <span>İndirim:</span>
                <Money :value="invoice.discount_total" :currency="invoice.currency" />
              </div>
              <div class="flex justify-between text-slate-600 dark:text-slate-400">
                <span>Vergi Toplamı:</span>
                <Money :value="invoice.tax_total" :currency="invoice.currency" />
              </div>
              <div class="flex justify-between text-xl font-bold text-slate-800 dark:text-slate-100 pt-3 border-t border-slate-200 dark:border-slate-700">
                <span>Genel Toplam:</span>
                <Money :value="invoice.total" :currency="invoice.currency" />
              </div>
            </div>
          </div>

          <!-- Disputed Note View -->
          <div v-if="invoice.status === 'disputed'" class="mb-8 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <div class="flex items-center gap-2 text-red-600 dark:text-red-400 font-bold mb-2">
              <i class="pi pi-info-circle"></i> Bu faturaya itiraz ettiniz
            </div>
            <p class="text-sm text-slate-700 dark:text-slate-300">{{ invoice.dispute_note || 'İtiraz notu faturaya eklendi.' }}</p>
          </div>

          <!-- Actions -->
          <div class="flex flex-col sm:flex-row gap-3 pt-6 border-t border-slate-200 dark:border-slate-700" v-if="invoice.status === 'sent' || invoice.status === 'partial'">
            <Button label="Hemen Öde" icon="pi pi-credit-card" class="w-full sm:flex-1 p-button-lg" @click="payInvoice" :loading="payLoading" severity="warn" />
            <Button label="Faturaya İtiraz Et" icon="pi pi-exclamation-triangle" class="w-full sm:flex-1 p-button-outlined p-button-danger p-button-lg" @click="openDispute" />
          </div>
        </template>
      </Card>

      <!-- Dispute Modal -->
      <Dialog v-model:visible="showDisputeModal" header="Faturaya İtiraz Et" :modal="true" class="w-full max-w-lg">
        <div class="mb-4">
          <p class="text-slate-600 dark:text-slate-400 mb-4">Lütfen faturadaki yanlışlığı veya itiraz nedeninizi açıklayın. Faturayı gönderen firma konuyla ilgili size geri dönüş yapacaktır.</p>
          <Textarea v-model="disputeNote" rows="5" class="w-full" placeholder="Örn: Hizmet bedeli eksik hesaplanmış..." />
        </div>
        <template #footer>
          <Button label="Vazgeç" icon="pi pi-times" class="p-button-text p-button-secondary" @click="showDisputeModal = false" :disabled="disputeLoading" />
          <Button label="İtirazı Gönder" icon="pi pi-send" class="p-button-danger" @click="submitDispute" :loading="disputeLoading" />
        </template>
      </Dialog>
    </div>
  </div>
</template>

<style scoped>
.invoice-card {
  border-radius: 1rem;
}
</style>
