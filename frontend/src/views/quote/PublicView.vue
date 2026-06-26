<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuoteStore } from '@/stores/quote'
import { useToast } from 'primevue/usetoast'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import Textarea from 'primevue/textarea'
import Money from '@/components/Money.vue'

const quoteStore = useQuoteStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()

const loading = ref(true)
const errorMsg = ref('')
const quoteData = ref<any>(null)

const rejectDialogVisible = ref(false)
const rejectNote = ref('')
const processing = ref(false)

const loadQuote = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const token = route.params.token as string
    quoteData.value = await quoteStore.getPublicQuote(token)
  } catch (err: any) {
    errorMsg.value = err.response?.data?.error?.message || 'Teklif yüklenirken bir hata oluştu.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadQuote()
})

const acceptQuote = async () => {
  if (confirm('Teklifi kabul ediyorsunuz. Onaylıyor musunuz?')) {
    processing.value = true
    try {
      const token = route.params.token as string
      await quoteStore.acceptPublicQuote(token)
      toast.add({ severity: 'success', summary: 'Teşekkürler', detail: 'Teklif onaylandı.', life: 10000 })
      await loadQuote()
    } catch (err: any) {
      toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'Onaylama başarısız oldu.', life: 10000 })
    } finally {
      processing.value = false
    }
  }
}

const rejectQuote = async () => {
  if (!rejectNote.value.trim()) {
    toast.add({ severity: 'warn', summary: 'Uyarı', detail: 'Lütfen bir reddetme nedeni girin.', life: 10000 })
    return
  }
  processing.value = true
  try {
    const token = route.params.token as string
    await quoteStore.rejectPublicQuote(token, rejectNote.value)
    toast.add({ severity: 'success', summary: 'İşlem Tamam', detail: 'Teklif reddedildi.', life: 10000 })
    rejectDialogVisible.value = false
    await loadQuote()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: err.response?.data?.error?.message || 'İşlem başarısız oldu.', life: 10000 })
  } finally {
    processing.value = false
  }
}

const fmt = (d?: string) => d ? new Date(d).toLocaleDateString('tr-TR') : '-'
const fmtN = (v?: string | number) => parseFloat(String(v || 0))
const fmtMoney = (v: number) => {
  return new Intl.NumberFormat('tr-TR', { style: 'currency', currency: quoteData.value?.currency || 'TRY', minimumFractionDigits: 2 }).format(v)
}

const lSub  = (i: any) => fmtN(i.quantity) * fmtN(i.unit_price)
const lDisc = (i: any) => lSub(i) * (fmtN(i.discount_rate) / 100)
const lNet  = (i: any) => lSub(i) - lDisc(i)

const statusLabel = (s?: string) => ({ draft: 'Taslak', sent: 'Onayınızı Bekliyor', accepted: 'Kabul Edildi', rejected: 'Reddedildi', expired: 'Süresi Doldu', converted: 'Faturalandı' }[s || ''] || s || '')
const statusSev   = (s?: string) => ({ draft: 'secondary', sent: 'info', accepted: 'success', rejected: 'danger', expired: 'contrast', converted: 'success' }[s || ''] || 'secondary') as any
</script>

<template>
  <div class="public-page p-4 md:p-8 min-h-screen bg-slate-50">
    <div v-if="loading" class="text-center py-20 text-slate-500">
      <i class="pi pi-spin pi-spinner text-3xl mb-4"></i>
      <p>Teklif yükleniyor...</p>
    </div>

    <div v-else-if="errorMsg" class="max-w-2xl mx-auto py-20">
      <Message severity="error" :closable="false" class="mb-4">{{ errorMsg }}</Message>
      <div class="text-center mt-6">
        <a href="https://radikalhesap.com" class="text-slate-500 hover:text-slate-800 underline">Radikal Hesap'a Dön</a>
      </div>
    </div>

    <div v-else-if="quoteData" class="max-w-4xl mx-auto">
      
      <!-- Top banner / actions -->
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 p-6 mb-6 flex flex-col md:flex-row justify-between items-center gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-800">Teklif Özeti</h1>
          <p class="text-slate-500 text-sm mt-1">Lütfen aşağıdaki teklif detaylarını inceleyip kararınızı iletin.</p>
        </div>
        <div class="flex items-center gap-3">
          <Tag :value="statusLabel(quoteData.status)" :severity="statusSev(quoteData.status)" class="px-3 py-1.5 text-sm" />
          <template v-if="quoteData.status === 'sent'">
            <button @click="rejectDialogVisible = true" :disabled="processing" class="px-4 py-2 bg-rose-50 hover:bg-rose-100 text-rose-600 font-semibold rounded-lg transition-colors border border-rose-200">
              Reddet
            </button>
            <button @click="acceptQuote" :disabled="processing" class="px-5 py-2 bg-emerald-500 hover:bg-emerald-600 text-white font-semibold rounded-lg shadow-sm shadow-emerald-200 transition-colors">
              <i class="pi pi-check mr-1"></i> Kabul Ediyorum
            </button>
          </template>
        </div>
      </div>

      <!-- Quote Document Details -->
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        
        <div class="p-6 md:p-8 border-b border-slate-100 bg-slate-50/50 flex flex-col md:flex-row justify-between gap-6">
          <div>
            <div class="text-xs font-bold tracking-wider text-cyan-600 mb-1">MÜŞTERİ</div>
            <div class="text-lg font-bold text-slate-800">{{ quoteData.cari?.name }}</div>
            <div class="text-sm text-slate-500 mt-2">{{ quoteData.cari?.email }}</div>
            <div class="text-sm text-slate-500">{{ quoteData.cari?.phone }}</div>
          </div>
          
          <div class="flex flex-wrap gap-8 md:justify-end text-right">
            <div>
              <div class="text-[10px] font-bold tracking-widest text-slate-400 mb-1 uppercase">Teklif No</div>
              <div class="text-sm font-semibold text-slate-700">{{ quoteData.number }}</div>
            </div>
            <div>
              <div class="text-[10px] font-bold tracking-widest text-slate-400 mb-1 uppercase">Tarih</div>
              <div class="text-sm font-semibold text-slate-700">{{ fmt(quoteData.date) }}</div>
            </div>
            <div>
              <div class="text-[10px] font-bold tracking-widest text-slate-400 mb-1 uppercase">Geçerlilik</div>
              <div class="text-sm font-semibold text-slate-700">{{ fmt(quoteData.expiry_date) }}</div>
            </div>
          </div>
        </div>

        <div class="p-6 md:p-8 overflow-x-auto">
          <table class="w-full text-sm text-left">
            <thead>
              <tr class="border-b-2 border-slate-200 text-slate-500">
                <th class="py-3 font-semibold w-12 text-center">#</th>
                <th class="py-3 font-semibold">Ürün / Hizmet</th>
                <th class="py-3 font-semibold text-center">Miktar</th>
                <th class="py-3 font-semibold text-right">Birim Fiyat</th>
                <th class="py-3 font-semibold text-right">Tutar</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, idx) in quoteData.items" :key="item.id" class="border-b border-slate-100 last:border-0 hover:bg-slate-50/50">
                <td class="py-4 text-center text-slate-400">{{ Number(idx) + 1 }}</td>
                <td class="py-4 font-medium text-slate-700">{{ item.description }}</td>
                <td class="py-4 text-center text-slate-600">{{ fmtN(item.quantity) }} {{ item.unit }}</td>
                <td class="py-4 text-right text-slate-600"><Money :value="item.unit_price" :currency="item.currency || quoteData.currency" /></td>
                <td class="py-4 text-right font-semibold text-slate-800"><Money :value="lNet(item).toString()" :currency="item.currency || quoteData.currency" /></td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="p-6 md:p-8 bg-slate-50/50 border-t border-slate-100 flex flex-col md:flex-row justify-between gap-8">
          <div class="flex-1">
            <div v-if="quoteData.note">
              <div class="text-xs font-bold tracking-wider text-slate-400 mb-2 uppercase">Notlar</div>
              <p class="text-sm text-slate-600 whitespace-pre-line leading-relaxed">{{ quoteData.note }}</p>
            </div>
            <div v-if="quoteData.reject_note" class="mt-4 p-4 bg-rose-50 rounded-lg border border-rose-100">
              <div class="text-xs font-bold tracking-wider text-rose-500 mb-2 uppercase">Reddetme Nedeni</div>
              <p class="text-sm text-rose-700 whitespace-pre-line">{{ quoteData.reject_note }}</p>
            </div>
          </div>
          
          <div class="w-full md:w-64 shrink-0 flex flex-col gap-2 text-sm">
            <div class="flex justify-between text-slate-500">
              <span>Ara Toplam</span>
              <span class="font-medium text-slate-700">{{ fmtMoney(fmtN(quoteData.subtotal)) }}</span>
            </div>
            <div v-if="fmtN(quoteData.discount_total) > 0" class="flex justify-between text-rose-500">
              <span>İndirim</span>
              <span class="font-medium">-{{ fmtMoney(fmtN(quoteData.discount_total)) }}</span>
            </div>
            <div class="flex justify-between text-slate-500">
              <span>KDV</span>
              <span class="font-medium text-slate-700">{{ fmtMoney(fmtN(quoteData.tax_total)) }}</span>
            </div>
            <div class="flex justify-between items-center mt-2 pt-3 border-t border-slate-200">
              <span class="font-bold text-slate-800">GENEL TOPLAM</span>
              <span class="text-xl font-bold text-cyan-600">{{ fmtMoney(fmtN(quoteData.total)) }}</span>
            </div>
          </div>
        </div>

      </div>

      <!-- Footer branding -->
      <div class="mt-12 text-center text-slate-400 text-xs">
        Bu belge <a href="https://radikalhesap.com" class="text-cyan-600 font-semibold hover:underline">Radikal Hesap</a> altyapısı ile güvenle oluşturulmuştur.
      </div>
    </div>

    <!-- Reject Dialog -->
    <Dialog v-model:visible="rejectDialogVisible" modal header="Teklifi Reddet" :style="{ width: '450px' }">
      <div class="flex flex-col gap-3 mt-2">
        <label for="rejectNote" class="text-sm font-medium text-slate-700">Lütfen reddetme nedeninizi kısaca belirtin:</label>
        <Textarea id="rejectNote" v-model="rejectNote" rows="4" class="w-full" placeholder="Örn: Fiyat yüksek geldi..." />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2 mt-4">
          <button @click="rejectDialogVisible = false" class="px-4 py-2 text-slate-600 font-medium hover:bg-slate-100 rounded-md transition-colors">İptal</button>
          <button @click="rejectQuote" :disabled="processing" class="px-4 py-2 bg-rose-500 hover:bg-rose-600 text-white font-medium rounded-md transition-colors shadow-sm">Teklifi Reddet</button>
        </div>
      </template>
    </Dialog>

  </div>
</template>

<style scoped>
.public-page {
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
}
</style>
