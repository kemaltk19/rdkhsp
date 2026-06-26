<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCariStore } from '@/stores/cari'
import { useInvoiceStore } from '@/stores/invoice'
import { usePaymentStore } from '@/stores/payment'
import { useQuoteStore } from '@/stores/quote'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Money from '@/components/Money.vue'
import FormModal from './FormModal.vue'

const route = useRoute()
const router = useRouter()
const cariStore = useCariStore()
const invoiceStore = useInvoiceStore()
const paymentStore = usePaymentStore()
const quoteStore = useQuoteStore()

const loading = ref(true)
const cariId = ref(route.params.id as string)
const activeTab = ref('profil')
const showEdit = ref(false)

const tabs = [
  { id: 'profil', label: 'Profil', icon: 'pi pi-user' },
  { id: 'kisiler', label: 'Kişiler', icon: 'pi pi-users' },
  { id: 'ekstre', label: 'Hesap ekstresi', icon: 'pi pi-chart-line' },
  { id: 'faturalar', label: 'Faturalar', icon: 'pi pi-file' },
  { id: 'odemeler', label: 'Ödemeler', icon: 'pi pi-money-bill' },
  { id: 'teklifler', label: 'Teklifler', icon: 'pi pi-tags' },
]

const cari = computed(() => cariStore.activeCari)
const initial = computed(() => (cari.value?.name ? cari.value.name.charAt(0).toUpperCase() : '?'))

const activeCurrencies = computed(() => {
  const curs = new Set<string>()
  if (cari.value?.currency) curs.add(cari.value.currency)

  cariStore.activeCariFinancialSummary?.sales?.forEach((s: any) => { if (s.total !== 0) curs.add(s.currency) })
  cariStore.activeCariFinancialSummary?.purchases?.forEach((p: any) => { if (p.total !== 0) curs.add(p.currency) })
  cari.value?.balances?.forEach((b: any) => { if (b.balance !== 0) curs.add(b.currency) })
  
  return Array.from(curs)
})

const standardizedSales = computed(() => activeCurrencies.value.map(curr => {
  const f = cariStore.activeCariFinancialSummary?.sales?.find((s: any) => s.currency === curr)
  return { currency: curr, total: f ? f.total : 0 }
}))
const standardizedPurchases = computed(() => activeCurrencies.value.map(curr => {
  const f = cariStore.activeCariFinancialSummary?.purchases?.find((p: any) => p.currency === curr)
  return { currency: curr, total: f ? f.total : 0 }
}))
const standardizedBalances = computed(() => activeCurrencies.value.map(curr => {
  const f = cari.value?.balances?.find((b: any) => b.currency === curr)
  return { currency: curr, balance: f ? f.balance : 0 }
}))
const standardizedCollections = computed(() => activeCurrencies.value.map(curr => {
  const f = cariStore.activeCariFinancialSummary?.collections?.find((s: any) => s.currency === curr)
  return { currency: curr, total: f ? f.total : 0 }
}))
const standardizedPayments = computed(() => activeCurrencies.value.map(curr => {
  const f = cariStore.activeCariFinancialSummary?.payments?.find((p: any) => p.currency === curr)
  return { currency: curr, total: f ? f.total : 0 }
}))

const loadData = async () => {
  loading.value = true
  try {
    await Promise.all([
      cariStore.fetchCariByID(cariId.value),
      cariStore.fetchTransactions(cariId.value, { page: 1, limit: 100 }),
      invoiceStore.fetchInvoices({ cari_id: cariId.value, limit: 100 }),
      paymentStore.fetchPayments({ cari_id: cariId.value, limit: 100 }),
      quoteStore.fetchQuotes({ cari_id: cariId.value, limit: 100 }),
      cariStore.fetchFinancialSummary(cariId.value),
    ])
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

const goBack = () => router.push('/caris')

const typeLabel = (t: string) => ({ customer: 'Müşteri', supplier: 'Tedarikçi', both: 'Her İkisi' } as any)[t] || t
const typeSeverity = (t: string) => ({ customer: 'info', supplier: 'warn', both: 'success' } as any)[t] || 'secondary'
const txTypeLabel = (t: string) => (t === 'debit' ? 'Borç' : 'Alacak')
const txTypeSeverity = (t: string) => (t === 'debit' ? 'success' : 'danger')
const txSourceLabel = (s: string) => ({ invoice: 'Fatura', payment: 'Ödeme / Tahsilat', expense: 'Gider', manual: 'Manuel' } as any)[s] || s
const invTypeLabel = (t: string) => (t === 'sales' ? 'Satış' : 'Alış')
const invStatusLabel = (s: string, type?: string) => {
  if (s === 'sent') return type === 'purchase' ? 'Alındı' : 'Gönderildi'
  return ({ draft: 'Taslak', partial: 'Kısmi', paid: 'Ödendi', canceled: 'İptal' } as any)[s] || s
}
const invStatusSeverity = (s: string) => ({ draft: 'secondary', sent: 'info', partial: 'warn', paid: 'success', canceled: 'danger' } as any)[s] || 'secondary'
const payTypeLabel = (t: string) => (t === 'collection' ? 'Tahsilat' : 'Ödeme')
const payTypeSeverity = (t: string) => (t === 'collection' ? 'success' : 'danger')
const payMethodLabel = (m: string) => ({ cash: 'Nakit', bank: 'Havale/EFT', card: 'Kredi Kartı', check: 'Çek' } as any)[m] || m
const quoteStatusLabel = (s: string) => ({ draft: 'Taslak', sent: 'Gönderildi', accepted: 'Kabul', rejected: 'Red', expired: 'Süresi doldu', converted: 'Faturalaştı' } as any)[s] || s
const quoteStatusSeverity = (s: string) => ({ draft: 'secondary', sent: 'info', accepted: 'success', rejected: 'danger', expired: 'warn', converted: 'success' } as any)[s] || 'secondary'

const fmtDate = (d: string) => (d ? new Date(d).toLocaleDateString('tr-TR', { day: 'numeric', month: 'short', year: 'numeric' }) : '')
const balColor = (b: string | number) => {
  const v = typeof b === 'string' ? parseFloat(b) : b
  if (v > 0) return 'pos'
  if (v < 0) return 'neg'
  return 'zero'
}

const hasBilling = computed(() => !!(cari.value?.address || cari.value?.city || cari.value?.district || cari.value?.postal_code))
const hasShipping = computed(() => !!(cari.value?.shipping_address || cari.value?.shipping_city || cari.value?.shipping_district || cari.value?.shipping_postal_code))


</script>

<template>
  <div v-if="!loading && cari" class="cari-detail">

    <!-- HEADER -->
    <div class="dh card">
      <div class="dh-left">
        <div class="dh-avatar">{{ initial }}</div>
        <div class="dh-info">
          <div class="dh-title">
            <h1>{{ cari.name }}</h1>
            <Tag :value="typeLabel(cari.type)" :severity="typeSeverity(cari.type)" />
          </div>
          <div class="dh-meta">
            <span class="code">#{{ cari.code }}</span>
            <span v-if="cari.contact_name"><i class="pi pi-user"></i>{{ cari.contact_name }}</span>
            <span v-if="cari.phone" :title="'Cep: ' + cari.phone"><i class="pi pi-mobile"></i>{{ cari.phone }}</span>
            <span v-if="cari.landline" :title="'Sabit: ' + cari.landline"><i class="pi pi-phone"></i>{{ cari.landline }}</span>
            <span v-if="cari.email"><i class="pi pi-envelope"></i>{{ cari.email }}</span>
          </div>
        </div>
      </div>
      <div class="dh-actions">
        <Button label="Geri" icon="pi pi-arrow-left" text @click="goBack" />
        <Button label="Düzenle" icon="pi pi-pencil" outlined @click="showEdit = true" severity="warn" />
      </div>
    </div>

    <!-- STAT CARDS -->
    <div class="stat-row">
      <div class="stat card">
        <span class="stat-label">Güncel bakiye</span>
        <div class="stat-values">
          <span v-for="b in standardizedBalances" :key="b.currency" class="stat-val" :class="balColor(b.balance)">
            <Money :value="b.balance" :currency="b.currency" />
          </span>
        </div>
      </div>
      <div class="stat card">
        <span class="stat-label">Toplam satış</span>
        <div class="stat-values">
          <span v-for="s in standardizedSales" :key="s.currency" class="stat-val">
            <Money :value="s.total" :currency="s.currency" />
          </span>
        </div>
      </div>
      <div class="stat card">
        <span class="stat-label">Tahsil edilen</span>
        <div class="stat-values">
          <span v-for="c in standardizedCollections" :key="c.currency" class="stat-val pos">
            <Money :value="c.total" :currency="c.currency" />
          </span>
        </div>
      </div>
      <div class="stat card" v-if="cari.type === 'supplier' || cari.type === 'both'">
        <span class="stat-label">Toplam alış</span>
        <div class="stat-values">
          <span v-for="p in standardizedPurchases" :key="p.currency" class="stat-val">
            <Money :value="p.total" :currency="p.currency" />
          </span>
        </div>
      </div>
      <div class="stat card" v-if="cari.type === 'supplier' || cari.type === 'both'">
        <span class="stat-label">Ödenen</span>
        <div class="stat-values">
          <span v-for="p in standardizedPayments" :key="p.currency" class="stat-val neg">
            <Money :value="p.total" :currency="p.currency" />
          </span>
        </div>
      </div>
    </div>

    <!-- TABS -->
    <div class="d-tabs">
      <Button v-for="t in tabs" :key="t.id" type="button" :class="['d-tab', activeTab === t.id && 'active']" @click="activeTab = t.id">
        <i :class="t.icon"></i>{{ t.label }}
      </button>
    </div>

    <!-- PROFIL -->
    <div v-show="activeTab === 'profil'" class="profil-grid">
      <div class="card">
        <h3 class="card-h">Cari bilgileri</h3>
        <div class="info">
          <div class="irow"><span>Cari kodu</span><b>{{ cari.code }}</b></div>
          <div class="irow"><span>Firma adı</span><b>{{ cari.name }}</b></div>
          <div class="irow" v-if="cari.contact_name"><span>Ad Soyad</span><b>{{ cari.contact_name }}</b></div>
          <div class="irow" v-if="cari.title"><span>Resmi ünvan</span><b>{{ cari.title }}</b></div>
          <div class="irow" v-if="cari.tax_number"><span>Vergi</span><b>{{ cari.tax_office }} V.D. / {{ cari.tax_number }}</b></div>
          <div class="irow" v-if="cari.phone"><span>Cep Telefonu</span><b>{{ cari.phone }}</b></div>
          <div class="irow" v-if="cari.landline"><span>Sabit Telefon</span><b>{{ cari.landline }}</b></div>
          <div class="irow" v-if="cari.fax"><span>Faks</span><b>{{ cari.fax }}</b></div>
          <div class="irow" v-if="cari.email"><span>E-posta</span><b>{{ cari.email }}</b></div>
          <div class="irow"><span>Para birimi</span><b>{{ cari.currency }}</b></div>
          <div class="irow" v-if="cari.note"><span>Notlar</span><b class="note">{{ cari.note }}</b></div>
        </div>
      </div>

      <div class="card">
        <h3 class="card-h">Adresler</h3>
        <div class="addr-block">
          <div class="addr-title">Fatura adresi</div>
          <p v-if="hasBilling" class="addr-text">
            {{ cari.address }}<br v-if="cari.address" />
            {{ [cari.district, cari.city, cari.postal_code].filter(Boolean).join(' / ') }}<br v-if="cari.city || cari.district" />
            {{ cari.country }}
          </p>
          <p v-else class="addr-empty">Fatura adresi girilmemiş.</p>
        </div>
        <div class="addr-block">
          <div class="addr-title">Sevk adresi</div>
          <p v-if="hasShipping" class="addr-text">
            {{ cari.shipping_address }}<br v-if="cari.shipping_address" />
            {{ [cari.shipping_district, cari.shipping_city, cari.shipping_postal_code].filter(Boolean).join(' / ') }}<br v-if="cari.shipping_city || cari.shipping_district" />
            {{ cari.shipping_country }}
          </p>
          <p v-else class="addr-empty">Sevk adresi girilmemiş.</p>
        </div>
      </div>
    </div>

    <!-- KİŞİLER -->
    <div v-show="activeTab === 'kisiler'" class="card">
      <h3 class="card-h">İletişim Kişisi (Yetkili)</h3>
      <div class="info" v-if="cari.contact_name">
        <div class="irow"><span>Ad Soyad</span><b>{{ cari.contact_name }}</b></div>
        <div class="irow" v-if="cari.phone"><span>Cep Telefonu</span><b>{{ cari.phone }}</b></div>
        <div class="irow" v-if="cari.landline"><span>Sabit Telefon</span><b>{{ cari.landline }}</b></div>
        <div class="irow" v-if="cari.email"><span>E-posta</span><b>{{ cari.email }}</b></div>
      </div>
      <div class="info" v-else>
        <p class="empty-msg">Bu cari için yetkili kişi / ad soyad bilgisi girilmemiştir.</p>
      </div>
    </div>

    <!-- EKSTRE -->
    <div v-show="activeTab === 'ekstre'" class="card">
      <h3 class="card-h">Hesap ekstresi / hareketler</h3>
      <DataTable :value="cariStore.activeCariTxs" class="p-datatable-sm" emptyMessage="Henüz hareket yok." paginator :rows="5" :rowsPerPageOptions="[5,10,20]" sortField="date" :sortOrder="-1">
        <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
          <template #body="slotProps">
            {{ slotProps.index + 1 }}
          </template>
        </Column>
        <Column field="date" header="Tarih" sortable style="min-width:120px"><template #body="{ data }">{{ fmtDate(data.date) }}</template></Column>
        <Column field="description" header="Açıklama" style="min-width:260px"></Column>
        <Column field="source_type" header="Kaynak" style="min-width:130px"><template #body="{ data }"><Tag :value="txSourceLabel(data.source_type)" severity="secondary" /></template></Column>
        <Column field="type" header="Tip" style="min-width:100px"><template #body="{ data }"><Tag :value="txTypeLabel(data.type)" :severity="txTypeSeverity(data.type)" /></template></Column>
        <Column header="Tutar" style="min-width:120px"><template #body="{ data }"><span class="num"><Money :value="data.amount" :currency="data.currency" /></span></template></Column>
        <Column header="Bakiye" style="min-width:120px"><template #body="{ data }"><span class="num" :class="balColor(data.balance_after)"><Money :value="data.balance_after" :currency="data.currency" /></span></template></Column>
      </DataTable>
    </div>

    <!-- FATURALAR -->
    <div v-show="activeTab === 'faturalar'" class="card">
      <h3 class="card-h">Faturalar</h3>
      <DataTable :value="invoiceStore.invoices" class="p-datatable-sm" emptyMessage="Fatura yok." paginator :rows="5" :rowsPerPageOptions="[5,10,20]" sortField="date" :sortOrder="-1">
        <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
          <template #body="slotProps">
            {{ slotProps.index + 1 }}
          </template>
        </Column>
        <Column field="number" header="Fatura no" style="min-width:140px"><template #body="{ data }"><b>{{ data.number }}</b></template></Column>
        <Column field="type" header="Tip" style="min-width:100px"><template #body="{ data }"><Tag :value="invTypeLabel(data.type)" :severity="data.type === 'sales' ? 'info' : 'warn'" /></template></Column>
        <Column field="date" header="Tarih" style="min-width:120px"><template #body="{ data }">{{ fmtDate(data.date) }}</template></Column>
        <Column field="status" header="Durum" style="min-width:110px"><template #body="{ data }"><Tag :value="invStatusLabel(data.status, data.type)" :severity="invStatusSeverity(data.status)" /></template></Column>
        <Column header="Toplam / Kalan" style="min-width:145px">
          <template #body="{ data }">
            <span class="num font-semibold"><Money :value="data.total" :currency="data.currency" /></span>
            <div v-if="data.status === 'partial'" class="partial-remaining text-right">
              <i class="pi pi-hourglass text-[10px] mr-1"></i>
              <span class="tabular-nums"><Money :value="(parseFloat(data.total || '0') - parseFloat(data.paid_total || '0')).toString()" :currency="data.currency" /></span>
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- ODEMELER -->
    <div v-show="activeTab === 'odemeler'" class="card">
      <h3 class="card-h">Ödemeler / tahsilatlar</h3>
      <DataTable :value="paymentStore.payments" class="p-datatable-sm" emptyMessage="Ödeme/tahsilat yok." paginator :rows="5" :rowsPerPageOptions="[5,10,20]" sortField="date" :sortOrder="-1">
        <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
          <template #body="slotProps">
            {{ slotProps.index + 1 }}
          </template>
        </Column>
        <Column field="reference" header="Makbuz / Ref No" style="min-width:140px"><template #body="{ data }"><b>{{ data.reference || '-' }}</b></template></Column>
        <Column field="type" header="İşlem" style="min-width:100px"><template #body="{ data }"><Tag :value="payTypeLabel(data.type)" :severity="payTypeSeverity(data.type)" /></template></Column>
        <Column field="date" header="Tarih" style="min-width:120px"><template #body="{ data }">{{ fmtDate(data.date) }}</template></Column>
        <Column field="method" header="Yöntem" style="min-width:130px"><template #body="{ data }">{{ payMethodLabel(data.method) }}</template></Column>
        <Column header="Tutar" style="min-width:120px"><template #body="{ data }"><span class="num"><Money :value="data.amount" :currency="data.currency" /></span></template></Column>
      </DataTable>
    </div>

    <!-- TEKLIFLER -->
    <div v-show="activeTab === 'teklifler'" class="card">
      <h3 class="card-h">Teklifler</h3>
      <DataTable :value="quoteStore.quotes" class="p-datatable-sm" emptyMessage="Teklif yok." paginator :rows="5" :rowsPerPageOptions="[5,10,20]" sortField="date" :sortOrder="-1">
        <Column header="#" headerStyle="width: 2.5rem" bodyClass="text-xs font-mono text-slate-400 dark:text-slate-500">
          <template #body="slotProps">
            {{ slotProps.index + 1 }}
          </template>
        </Column>
        <Column field="number" header="Teklif no" style="min-width:140px"><template #body="{ data }"><b>{{ data.number }}</b></template></Column>
        <Column field="date" header="Tarih" style="min-width:120px"><template #body="{ data }">{{ fmtDate(data.date) }}</template></Column>
        <Column field="status" header="Durum" style="min-width:110px"><template #body="{ data }"><Tag :value="quoteStatusLabel(data.status)" :severity="quoteStatusSeverity(data.status)" /></template></Column>
        <Column header="Toplam" style="min-width:120px"><template #body="{ data }"><span class="num"><Money :value="data.total" :currency="data.currency" /></span></template></Column>
      </DataTable>
    </div>

    <FormModal v-if="showEdit" v-model:visible="showEdit" :cari-id="cariId" @saved="loadData" />
  </div>

  <div v-else class="loading-state">
    <i class="pi pi-spin pi-spinner"></i>
    <span>Cari detayları yükleniyor...</span>
  </div>
</template>

<style scoped>
.cari-detail { display: flex; flex-direction: column; gap: 1.25rem; }

/* Header */
.dh { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
.dh-left { display: flex; align-items: center; gap: 16px; min-width: 0; }
.dh-avatar { width: 56px; height: 56px; border-radius: 14px; background: #06b6d4; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 24px; font-weight: 700; flex-shrink: 0; }
.dh-title { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.dh-title h1 { font-size: 1.4rem; font-weight: 700; margin: 0; }
.dh-meta { display: flex; flex-wrap: wrap; gap: 14px; margin-top: 6px; color: #64748b; font-size: 0.88rem; }
.dh-meta span { display: inline-flex; align-items: center; gap: 5px; }
.dh-meta .code { font-weight: 600; color: #0891b2; }
.dh-actions { display: flex; gap: 8px; flex-shrink: 0; }

/* Stats */
.stat-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 1rem; margin: 0; }
.stat { margin-bottom: 0; padding: 1.1rem 1.25rem; }
.stat-label { font-size: 0.82rem; font-weight: 600; color: #64748b; }
.stat-values { display: flex; flex-direction: column; gap: 2px; margin-top: 8px; }
.stat-val { font-size: 1.35rem; font-weight: 700; letter-spacing: -0.02em; }
.stat-val.pos { color: #16a34a; }
.stat-val.neg { color: #dc2626; }
.stat-val.zero { color: #475569; }

/* Tabs */
.d-tabs { display: flex; gap: 4px; flex-wrap: wrap; border-bottom: 1px solid rgba(38, 43, 67, 0.1); }
.d-tab { background: transparent; border: none; padding: 10px 16px; font-size: 0.92rem; color: #64748b; cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -1px; display: inline-flex; align-items: center; gap: 7px; }
.d-tab:hover { color: #0891b2; }
.d-tab.active { color: #0891b2; font-weight: 600; border-bottom-color: #06b6d4; }

/* Profil */
.profil-grid { display: grid; grid-template-columns: 1fr; gap: 1.25rem; }
@media (min-width: 1024px) { .profil-grid { grid-template-columns: 3fr 2fr; } }
.card-h { font-size: 1rem; font-weight: 600; margin: 0 0 1rem; }
.info { display: flex; flex-direction: column; }
.irow { display: flex; gap: 16px; padding: 10px 0; border-bottom: 1px solid rgba(38, 43, 67, 0.06); }
.irow:last-child { border-bottom: none; }
.irow > span { width: 130px; flex-shrink: 0; color: #64748b; font-size: 0.85rem; }
.irow > b { flex: 1; font-weight: 500; font-size: 0.92rem; }
.irow > b.note { white-space: pre-wrap; font-weight: 400; color: #475569; font-style: italic; }

.addr-block { margin-bottom: 1.25rem; }
.addr-block:last-child { margin-bottom: 0; }
.addr-title { font-size: 0.82rem; font-weight: 600; color: #0891b2; margin-bottom: 6px; }
.addr-text { margin: 0; font-size: 0.92rem; line-height: 1.55; color: #334155; }
.addr-empty { margin: 0; font-size: 0.88rem; color: #94a3b8; font-style: italic; }

.num { display: block; text-align: right; font-variant-numeric: tabular-nums; }
.num.pos { color: #16a34a; font-weight: 600; }
.num.neg { color: #dc2626; font-weight: 600; }

.partial-remaining {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  font-size: 0.78rem;
  color: #d97706;
  margin-top: 2px;
}
:root.p-dark .partial-remaining {
  color: #fbbf24;
}

.loading-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 5rem 0; gap: 1rem; color: #64748b; }
.loading-state i { font-size: 2.5rem; }

/* Dark mode */
:root.p-dark .stat-label,
:root.p-dark .dh-meta,
:root.p-dark .irow > span { color: #94a3b8; }
:root.p-dark .d-tabs,
:root.p-dark .irow { border-color: rgba(255, 255, 255, 0.08); }
:root.p-dark .addr-text { color: #cbd5e1; }
:root.p-dark .stat-val.zero { color: #cbd5e1; }
</style>
