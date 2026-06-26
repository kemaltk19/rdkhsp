<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQuoteStore } from '@/stores/quote'
import { useCariStore } from '@/stores/cari'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Money from '@/components/Money.vue'
import QuotePrintTemplate from '@/components/QuotePrintTemplate.vue'
// jsPDF / autotable ana bundle'a girmesin diye exportPDF içinde dinamik yüklenir.
import AuditHistory from '@/components/AuditHistory.vue'

const quoteStore = useQuoteStore()
const cariStore = useCariStore()
const settingsStore = useSettingsStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()

const loading = ref(false)
const errorMsg = ref('')

const quote = computed(() => quoteStore.activeQuote)
const cari = computed(() => {
  if (!quote.value) return null
  return cariStore.caris.find(c => c.id === quote.value!.cari_id) || null
})

const statusOptions = ref([
  { label: 'Taslak', value: 'draft' },
  { label: 'Teklif Gönderildi', value: 'sent' },
  { label: 'Kabul Edildi', value: 'accepted' },
  { label: 'Reddedildi', value: 'rejected' },
  { label: 'Süresi Doldu', value: 'expired' },
])

const loadQuote = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const id = route.params.id as string
    await quoteStore.fetchQuoteByID(id)
    await cariStore.fetchCaris({ page: 1, limit: 1000 })
    await settingsStore.fetchCompanyProfile()
  } catch (err) {
    errorMsg.value = 'Teklif bilgileri yüklenemedi.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadQuote()
})

const goBack = () => router.push('/quotes')
const editQuote = () => { if (quote.value) router.push(`/quotes/${quote.value.id}/edit`) }

const deleteQuote = async () => {
  if (!quote.value) return
  if (confirm('Bu teklifi silmek istediğinize emin misiniz?')) {
    try {
      await quoteStore.deleteQuote(quote.value.id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif silindi', life: 10000 })
      router.push('/quotes')
    } catch { toast.add({ severity: 'error', summary: 'Hata', detail: 'Teklif silinemedi', life: 10000 }) }
  }
}

const handleStatusChange = async (newStatus: string) => {
  if (!quote.value) return
  loading.value = true
  try {
    await quoteStore.updateQuoteStatus(quote.value.id, newStatus)
    toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif durumu güncellendi', life: 10000 })
    await loadQuote()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: 'Hata', detail: 'Durum güncellenirken hata oluştu', life: 10000 })
  } finally {
    loading.value = false
  }
}

const convertToInvoice = async () => {
  if (!quote.value) return
  if (confirm('Bu teklifi taslak bir satış faturasına dönüştürmek istediğinize emin misiniz?')) {
    loading.value = true
    try {
      const draftInvoice = await quoteStore.convertQuote(quote.value.id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif faturaya dönüştürüldü', life: 10000 })
      router.push(`/invoices/${draftInvoice.id}`)
    } catch (err: any) {
      if (err.response?.data?.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Faturaya dönüştürme başarısız', life: 10000 })
      }
    } finally {
      loading.value = false
    }
  }
}

const sendQuote = async () => {
  if (!quote.value) return
  if (confirm('Teklifi müşteriye göndermek istediğinize emin misiniz?')) {
    loading.value = true
    try {
      await quoteStore.sendQuote(quote.value.id)
      toast.add({ severity: 'success', summary: 'Başarılı', detail: 'Teklif e-posta ile gönderildi', life: 10000 })
      await loadQuote()
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.error) {
        toast.add({ severity: 'error', summary: 'Hata', detail: err.response.data.error.message, life: 10000 })
      } else {
        toast.add({ severity: 'error', summary: 'Hata', detail: 'Teklif gönderilemedi', life: 10000 })
      }
    } finally {
      loading.value = false
    }
  }
}

const fmt = (d?: string) => d ? new Date(d).toLocaleDateString('tr-TR') : '-'
const fmtN = (v?: string | number) => parseFloat(String(v || 0))

// Satırın kendi para biriminde KDV hariç/dahil tutarları (gösterim amaçlı, satır kendi dövizinde).
const lSub  = (i: any) => fmtN(i.quantity) * fmtN(i.unit_price)
const lDisc = (i: any) => lSub(i) * (fmtN(i.discount_rate) / 100)
const lNet  = (i: any) => lSub(i) - lDisc(i)
const lTax  = (i: any) => lNet(i) * (fmtN(i.tax_rate) / 100)
const lTot  = (i: any) => lNet(i) + lTax(i)

// Teklif toplamları: backend'in döviz kuru çevirisiyle hesapladığı ve kaydettiği
// değerler kullanılır (quote_service.go convertToDefaultCurrency). Burada tekrar
// toplamak, farklı para biriminde satırlarda yanlış sonuç üretir.
const subtotal      = computed(() => fmtN(quote.value?.subtotal))
const discountTotal = computed(() => fmtN(quote.value?.discount_total))
const netTotal      = computed(() => subtotal.value - discountTotal.value)
const taxTotal      = computed(() => fmtN(quote.value?.tax_total))
const grandTotal    = computed(() => fmtN(quote.value?.total))

const exchangeRatesInfo = computed(() => {
  const items = quote.value?.items || [];
  const rates = new Map<string, number>();
  for (const item of items) {
    if (item.currency && item.currency.toUpperCase() !== quote.value?.currency?.toUpperCase()) {
      const rate = parseFloat(item.exchange_rate);
      if (!isNaN(rate) && rate > 0) {
        rates.set(item.currency.toUpperCase(), rate);
      }
    }
  }
  return Array.from(rates.entries()).map(([curr, rate]) => ({ curr, rate }));
});

const statusLabel = (s?: string) => ({ draft: 'Taslak', sent: 'Gönderildi', accepted: 'Kabul Edildi', rejected: 'Reddedildi', expired: 'Süresi Doldu', converted: 'Faturalandı' }[s || ''] || s || '')
const statusSev   = (s?: string) => ({ draft: 'secondary', sent: 'info', accepted: 'success', rejected: 'danger', expired: 'contrast', converted: 'success' }[s || ''] || 'secondary') as any

const currency = computed(() => quote.value?.currency || 'TRY')
const fmtMoney = (v: number) => {
  return new Intl.NumberFormat('tr-TR', { style: 'currency', currency: currency.value, minimumFractionDigits: 2 }).format(v)
}

/* ── PDF Export ── */
const exportPDF = async () => {
  if (!quote.value) return
  const { default: jsPDF } = await import('jspdf')
  const { default: autoTable } = await import('jspdf-autotable')
  const doc = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' })
  const qte = quote.value
  const co = settingsStore.company
  const cy = cari.value

  // Header background
  doc.setFillColor(6, 182, 212)
  doc.rect(140, 0, 70, 55, 'F')

  // Company name
  doc.setFont('helvetica', 'bold')
  doc.setFontSize(14)
  doc.setTextColor(15, 23, 42)
  doc.text((co?.name || 'FİRMA').toUpperCase(), 14, 18)

  doc.setFont('helvetica', 'normal')
  doc.setFontSize(8)
  doc.setTextColor(100, 116, 139)
  if (co?.title) doc.text(co.title, 14, 24)

  // Contact in cyan panel
  doc.setTextColor(255, 255, 255)
  doc.setFontSize(8)
  let cy2 = 10
  if (co?.phone)   { doc.text(`Tel: ${co.phone}`, 144, cy2); cy2 += 6 }
  if (co?.email)   { doc.text(`E: ${co.email}`, 144, cy2); cy2 += 6 }
  if (co?.address) { doc.text(co.address.substring(0, 30), 144, cy2); cy2 += 6 }
  if (co?.tax_number) { doc.text(`V.N.: ${co.tax_number}`, 144, cy2) }

  // Quote To
  doc.setTextColor(6, 182, 212)
  doc.setFontSize(7)
  doc.setFont('helvetica', 'bold')
  doc.text('TEKLİF KESİLEN', 14, 34)
  doc.setTextColor(15, 23, 42)
  doc.setFontSize(10)
  doc.text(cy?.name || '-', 14, 40)
  doc.setFont('helvetica', 'normal')
  doc.setFontSize(8)
  doc.setTextColor(100, 116, 139)
  if (cy?.address) doc.text(cy.address, 14, 46)
  if (cy?.phone) doc.text(`Tel: ${cy.phone}`, 14, 51)
  if (cy?.tax_number) doc.text(`V.N.: ${cy.tax_number}`, 14, 56)

  // Meta bar
  doc.setFillColor(248, 250, 252)
  doc.rect(0, 62, 210, 16, 'F')
  doc.setDrawColor(6, 182, 212)
  doc.setLineWidth(0.8)
  doc.line(0, 78, 210, 78)

  const metas = [
    { label: 'TEKLİF NO', val: qte.number || '-' },
    { label: 'TARİH', val: fmt(qte.date) },
    { label: 'GEÇERLİLİK', val: fmt(qte.expiry_date) },
    { label: 'GENEL TOPLAM', val: fmtMoney(grandTotal.value) },
  ]
  metas.forEach((m, i) => {
    const x = 14 + i * 50
    doc.setFont('helvetica', 'bold')
    doc.setFontSize(7)
    doc.setTextColor(148, 163, 184)
    doc.text(m.label, x, 68)
    doc.setFont('helvetica', 'bold')
    doc.setFontSize(9)
    doc.setTextColor(i === 3 ? 6 : 15, i === 3 ? 182 : 23, i === 3 ? 212 : 42)
    doc.text(m.val, x, 75)
  })

  // Table
  const rows = (qte.items || []).map((item, idx) => [
    idx + 1,
    item.description || '-',
    item.unit || 'Adet',
    fmtN(item.quantity),
    fmtMoney(fmtN(item.unit_price)),
    `%${fmtN(item.discount_rate)}`,
    fmtMoney(lDisc(item)),
    `%${fmtN(item.tax_rate)}`,
    fmtMoney(lNet(item)),
  ])

  autoTable(doc, {
    startY: 82,
    head: [['#', 'ÜRÜN / HİZMET', 'BİRİM', 'MİKT.', 'BİRİM FİYAT', 'İND%', 'İNDİRİM', 'KDV%', 'TUTAR (KDV HARİÇ)']],
    body: rows,
    styles: { fontSize: 8, cellPadding: 2.5 },
    headStyles: { fillColor: [14, 116, 144], textColor: [255,255,255], fontStyle: 'bold', fontSize: 7.5 },
    columnStyles: {
      0: { halign: 'center', cellWidth: 8 },
      1: { cellWidth: 'auto' },
      2: { halign: 'center', cellWidth: 16 },
      3: { halign: 'center', cellWidth: 14 },
      4: { halign: 'right', cellWidth: 22 },
      5: { halign: 'center', cellWidth: 12 },
      6: { halign: 'right', cellWidth: 20 },
      7: { halign: 'center', cellWidth: 12 },
      8: { halign: 'right', cellWidth: 26 },
    },
    alternateRowStyles: { fillColor: [250, 252, 255] },
    margin: { left: 14, right: 14 },
  })

  const finalY = (doc as any).lastAutoTable.finalY + 8

  // Totals
  const totLines = [
    { lbl: 'Ara Toplam:', val: fmtMoney(subtotal.value) },
    { lbl: 'Toplam İndirim (−):', val: fmtMoney(discountTotal.value) },
    { lbl: 'Vergi Öncesi:', val: fmtMoney(netTotal.value) },
    { lbl: 'KDV:', val: fmtMoney(taxTotal.value) },
  ]
  let ty = finalY
  totLines.forEach(l => {
    doc.setFont('helvetica', 'normal')
    doc.setFontSize(8.5)
    doc.setTextColor(100, 116, 139)
    doc.text(l.lbl, 130, ty)
    doc.setTextColor(15, 23, 42)
    doc.text(l.val, 196, ty, { align: 'right' })
    ty += 6
  })

  // Grand total box
  doc.setFillColor(236, 254, 255)
  doc.setDrawColor(6, 182, 212)
  doc.setLineWidth(0.6)
  doc.roundedRect(128, ty, 68, 12, 2, 2, 'FD')
  doc.setFont('helvetica', 'bold')
  doc.setFontSize(8)
  doc.setTextColor(14, 116, 144)
  doc.text('GENEL TOPLAM', 132, ty + 7.5)
  doc.setFontSize(12)
  doc.text(fmtMoney(grandTotal.value), 196, ty + 7.5, { align: 'right' })

  // Notes
  if (qte.note) {
    doc.setFont('helvetica', 'bold')
    doc.setFontSize(7)
    doc.setTextColor(6, 182, 212)
    doc.text('NOTLAR / ÖDEME BİLGİLERİ', 14, ty)
    doc.setFont('helvetica', 'normal')
    doc.setTextColor(71, 85, 105)
    doc.setFontSize(8)
    const lines = doc.splitTextToSize(qte.note, 100)
    doc.text(lines, 14, ty + 6)
  }

  // TEKLIF watermark
  doc.setFont('helvetica', 'bold')
  doc.setFontSize(40)
  doc.setTextColor(6, 182, 212)
  doc.setGState(doc.GState({ opacity: 0.08 }))
  doc.text('TEKLİF', 196, 280, { align: 'right' })
  doc.setGState(doc.GState({ opacity: 1 }))

  // Footer
  doc.setFont('helvetica', 'normal')
  doc.setFontSize(7.5)
  doc.setTextColor(148, 163, 184)
  const validD = fmt(qte.expiry_date)
  doc.text(`Bu teklif ${validD} tarihine kadar geçerlidir. Şartlar ve koşullar için iletişime geçin.`, 14, 286)
  doc.setTextColor(6, 182, 212)
  doc.setFont('helvetica', 'bold')
  doc.text('İlginiz için teşekkürler!', 196, 286, { align: 'right' })

  doc.save(`${qte.number || 'teklif'}.pdf`)
}

/* ── Excel Export ── */
const exportExcel = () => {
  if (!quote.value) return
  const qte = quote.value
  const headers = ['#', 'Ürün/Hizmet', 'Birim', 'Miktar', 'Birim Fiyat', 'İnd%', 'İndirim', 'KDV%', 'Tutar (KDV Hariç)', 'KDV Tutar', 'Genel Toplam']
  const rows = (qte.items || []).map((item, idx) => [
    idx + 1, item.description || '',
    item.unit || 'Adet', fmtN(item.quantity),
    fmtN(item.unit_price), fmtN(item.discount_rate),
    lDisc(item).toFixed(2), fmtN(item.tax_rate),
    lNet(item).toFixed(2), lTax(item).toFixed(2), lTot(item).toFixed(2),
  ])
  const summary = [
    [], ['', '', '', '', '', '', '', '', 'Ara Toplam', '', subtotal.value.toFixed(2)],
    ['', '', '', '', '', '', '', '', 'İndirim', '', discountTotal.value.toFixed(2)],
    ['', '', '', '', '', '', '', '', 'Vergi Öncesi', '', netTotal.value.toFixed(2)],
    ['', '', '', '', '', '', '', '', 'KDV', '', taxTotal.value.toFixed(2)],
    ['', '', '', '', '', '', '', '', 'GENEL TOPLAM', '', grandTotal.value.toFixed(2)],
  ]
  const csvContent = [headers, ...rows, ...summary]
    .map(r => r.map(v => `"${String(v).replace(/"/g, '""')}"`).join(','))
    .join('\n')
  const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url; a.download = `${qte.number || 'teklif'}.csv`
  a.click(); URL.revokeObjectURL(url)
}

const printQuote = () => {
  const originalTitle = document.title
  document.title = quote.value?.number || 'Teklif'
  const style = document.createElement('style')
  style.innerHTML = `@media print { 
    @page { size: A4 portrait; margin: 0; }
    body { margin: 0; padding: 0; background: #fff; }
    
    /* Hide layout elements globally */
    .sidebar, .header, .no-print { display: none !important; }
    
    /* Remove padding/margin from layout containers */
    .app-container, .main-layout, .content, .fade-in-container, .id-page {
      padding: 0 !important;
      margin: 0 !important;
      background: #fff !important;
      min-height: 0 !important;
      display: block !important;
    }
    
    /* Document adjustments */
    .id-doc { display: none !important; }
    #quote-print-container { display: block !important; }
    .fp-doc { 
      width: 210mm !important;
      min-height: 297mm !important;
      margin: 0 auto !important; 
      padding: 0 !important; 
      border: none !important; 
      box-shadow: none !important; 
      -webkit-print-color-adjust: exact !important; 
      print-color-adjust: exact !important; 
    }
  }`
  document.head.appendChild(style)
  setTimeout(() => {
    window.print()
    document.head.removeChild(style)
    document.title = originalTitle
  }, 100)
}
</script>

<template>
  <div class="id-page">

    <!-- Action Bar -->
    <div class="id-bar no-print">
      <div class="id-bar-l">
        <Button class="id-back" @click="goBack"><i class="pi pi-arrow-left"></i> Geri</button>
        <span class="id-bar-title">Teklif Detayı</span>
        <Tag v-if="quote" :value="statusLabel(quote.status)" :severity="statusSev(quote.status)" />
      </div>
      <div class="id-bar-r">
        <!-- Durum Değiştirici -->
        <Select
          v-if="quote && quote.status !== 'converted'"
          v-model="quote.status"
          :options="statusOptions"
          optionLabel="label"
          optionValue="value"
          class="id-status-sel"
          @change="handleStatusChange($event.value)"
          :disabled="loading"
        />

        <Button v-if="quote?.status === 'draft' || quote?.status === 'rejected'" class="id-btn id-btn-ok" @click="sendQuote" :disabled="loading" title="Gönder">
          <i class="pi pi-send"></i> Gönder
        </button>
        <Button v-if="quote?.status !== 'converted' && quote?.status !== 'accepted'" class="id-btn id-btn-warn" @click="editQuote" :disabled="loading" title="Düzenle" severity="warn">
          <i class="pi pi-pencil"></i> Düzenle
        </button>
        <Button v-if="quote?.status !== 'converted'" class="id-btn id-btn-ok" @click="convertToInvoice" :disabled="loading" title="Faturalandır">
          <i class="pi pi-file-export"></i> Faturalandır
        </button>
        <Button v-if="quote?.status !== 'converted' && quote?.status !== 'accepted'" class="id-btn id-btn-danger" @click="deleteQuote" :disabled="loading" severity="danger">
          <i class="pi pi-trash"></i> Sil
        </button>
        <div class="id-sep"></div>
        <Button class="id-btn id-btn-export" @click="printQuote" title="Yazdır / PDF" severity="contrast"><i class="pi pi-print"></i> Yazdır / PDF İndir</button>
        <Button class="id-btn id-btn-export" @click="exportExcel" title="Excel" severity="contrast"><i class="pi pi-file-excel"></i> Excel</button>
      </div>
    </div>

    <Message v-if="errorMsg" severity="error" class="id-err no-print">{{ errorMsg }}</Message>
    <Message v-if="quote && quote.status === 'rejected'" severity="warn" class="id-err no-print" :closable="false">
      <strong>Müşteri Teklifi Reddetti:</strong> {{ quote.reject_note || 'Reddetme nedeni belirtilmemiş.' }}
    </Message>

    <!-- Converted Banner -->
    <div v-if="quote?.status === 'converted' && quote?.converted_invoice_id" class="id-conv-banner no-print">
      <Message severity="success" :closable="false">
        Bu teklif faturaya dönüştürülmüştür. 
        <router-link :to="`/invoices/${quote.converted_invoice_id}`" class="id-conv-link">
          Dönüştürülen Faturayı Görüntüle &rarr;
        </router-link>
      </Message>
    </div>

    <div v-if="loading && !quote" class="id-loading no-print">Yükleniyor...</div>

    <!-- ═══ QUOTE DOCUMENT (design mode) ═══ -->
    <div v-if="quote" id="quote-print" class="id-doc">

      <!-- Header -->
      <div class="id-head">
        <div class="id-head-l">
          <div class="id-logo-row">
            <div class="id-logo"><i class="pi pi-file-edit"></i></div>
            <div>
              <div class="id-co-name">{{ settingsStore.company?.name || 'FİRMA ADINIZ' }}</div>
              <div class="id-co-sub">{{ settingsStore.company?.title || '' }}</div>
            </div>
          </div>
          <div class="id-divider"></div>
          <div class="id-to-lbl">TEKLİF KESİLEN</div>
          <div v-if="cari" class="id-cari-info">
            <div class="id-cari-name">{{ cari.name }}</div>
            <div v-if="cari.address" class="id-cari-row"><i class="pi pi-map-marker"></i>{{ cari.address }}</div>
            <div v-if="cari.phone" class="id-cari-row"><i class="pi pi-phone"></i>{{ cari.phone }}</div>
            <div v-if="cari.email" class="id-cari-row"><i class="pi pi-envelope"></i>{{ cari.email }}</div>
            <div v-if="cari.tax_number" class="id-cari-row"><i class="pi pi-id-card"></i>V.D.: {{ cari.tax_office }} | V.N.: {{ cari.tax_number }}</div>
          </div>
          <div v-else class="id-cari-row">Cari bilgisi bulunamadı</div>
        </div>
        <div class="id-head-r">
          <div v-if="settingsStore.company?.phone" class="id-cr"><i class="pi pi-phone"></i><span>{{ settingsStore.company.phone }}</span></div>
          <div v-if="settingsStore.company?.website" class="id-cr"><i class="pi pi-globe"></i><span>{{ settingsStore.company.website }}</span></div>
          <div v-if="settingsStore.company?.email" class="id-cr"><i class="pi pi-envelope"></i><span>{{ settingsStore.company.email }}</span></div>
          <div v-if="settingsStore.company?.address" class="id-cr"><i class="pi pi-map-marker"></i><span>{{ settingsStore.company.address }}</span></div>
          <div v-if="settingsStore.company?.tax_number" class="id-cr"><i class="pi pi-id-card"></i><span>V.N.: {{ settingsStore.company.tax_number }}</span></div>
        </div>
      </div>

      <!-- Meta bar -->
      <div class="id-meta">
        <div class="id-mc">
          <div class="id-ml">GENEL TOPLAM</div>
          <div class="id-mv id-mv-total"><Money :value="grandTotal.toString()" :currency="quote.currency" /></div>
        </div>
        <div class="id-ms"></div>
        <div class="id-mc">
          <div class="id-ml">TEKLİF NO</div>
          <div class="id-mv">{{ quote.number || '-' }}</div>
        </div>
        <div class="id-ms"></div>
        <div class="id-mc">
          <div class="id-ml">TEKLİF TARİHİ</div>
          <div class="id-mv">{{ fmt(quote.date) }}</div>
        </div>
        <div class="id-ms"></div>
        <div class="id-mc">
          <div class="id-ml">GEÇERLİLİK TARİHİ</div>
          <div class="id-mv">{{ fmt(quote.expiry_date) }}</div>
        </div>
        <div class="id-ms" v-if="quote.created_by_user"></div>
        <div class="id-mc" v-if="quote.created_by_user">
          <div class="id-ml">OLUŞTURAN</div>
          <div class="id-mv">{{ quote.created_by_user.name }}</div>
        </div>
        <div class="id-ms" v-if="quote.updated_by_user"></div>
        <div class="id-mc" v-if="quote.updated_by_user">
          <div class="id-ml">SON DÜZENLEYEN</div>
          <div class="id-mv">{{ quote.updated_by_user.name }}</div>
        </div>
      </div>

      <!-- Table -->
      <div class="id-tbl-wrap">
        <table class="id-tbl">
          <colgroup>
            <col style="width:36px">
            <col style="width:auto">
            <col style="width:70px">
            <col style="width:60px">
            <col style="width:100px">
            <col style="width:48px">
            <col style="width:90px">
            <col style="width:58px">
            <col style="width:110px">
          </colgroup>
          <thead>
            <tr>
              <th class="ith c">#</th>
              <th class="ith">ÜRÜN / HİZMET AÇIKLAMASI</th>
              <th class="ith c">BİRİM</th>
              <th class="ith c">MİKT.</th>
              <th class="ith r">BİRİM FİYAT</th>
              <th class="ith c">İND%</th>
              <th class="ith r">İNDİRİM</th>
              <th class="ith c">KDV%</th>
              <th class="ith r">TUTAR (KDV HARİÇ)</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, idx) in quote.items" :key="item.id" class="itr" :class="{ 'itr-even': idx % 2 === 1 }">
              <td class="itd c itd-no">{{ idx + 1 }}</td>
              <td class="itd">
                <div class="itd-desc">{{ item.description || '-' }}</div>
              </td>
              <td class="itd c">{{ item.unit || 'Adet' }}</td>
              <td class="itd c">{{ fmtN(item.quantity) }}</td>
              <td class="itd r"><Money :value="item.unit_price" :currency="item.currency || quote.currency" /></td>
              <td class="itd c">%{{ fmtN(item.discount_rate) }}</td>
              <td class="itd r itd-disc"><Money :value="lDisc(item).toString()" :currency="item.currency || quote.currency" /></td>
              <td class="itd c">%{{ fmtN(item.tax_rate) }}</td>
              <td class="itd r itd-tot"><Money :value="lNet(item).toString()" :currency="item.currency || quote.currency" /></td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Bottom -->
      <div class="id-bot">
        <div class="id-bot-l">
          <div v-if="quote.note">
            <div class="id-sec-lbl">NOTLAR / ÖDEME BİLGİLERİ</div>
            <div class="id-note">{{ quote.note }}</div>
          </div>
          <div class="id-sig">
            <div class="id-sig-name">{{ settingsStore.company?.name || '' }}</div>
            <div class="id-sig-line"></div>
            <div class="id-sig-role">Yetkili / Authorized</div>
          </div>
        </div>
        <div class="id-bot-r">
          <div class="id-tots">
            <div class="id-trow"><span>Ara Toplam</span><Money :value="subtotal.toString()" :currency="quote.currency" /></div>
            <div v-if="discountTotal > 0" class="id-trow id-trow-disc"><span>İndirim (−)</span><Money :value="discountTotal.toString()" :currency="quote.currency" /></div>
            <div class="id-trow"><span>Vergi Öncesi</span><Money :value="netTotal.toString()" :currency="quote.currency" /></div>
            <div class="id-trow"><span>KDV</span><Money :value="taxTotal.toString()" :currency="quote.currency" /></div>
            <div class="id-tsep"></div>
            <div class="id-tgrand">
              <span>GENEL TOPLAM</span>
              <span class="id-tgval"><Money :value="grandTotal.toString()" :currency="quote.currency" /></span>
            </div>

            <div v-if="exchangeRatesInfo.length > 0" class="mt-3 text-[11px] text-slate-500 bg-slate-50 p-2 rounded border border-slate-200">
              <div class="font-bold mb-1 text-slate-600">Kur Bilgisi</div>
              <div v-for="r in exchangeRatesInfo" :key="r.curr" class="flex justify-between gap-4">
                <span>{{ r.curr }}</span>
                <span>{{ new Intl.NumberFormat('tr-TR', { minimumFractionDigits: 4, maximumFractionDigits: 4 }).format(r.rate) }}</span>
              </div>
            </div>
          </div>
          <div class="id-deco">TEKLİF</div>
        </div>
      </div>

      <div class="id-terms">
        <span>Teklif tutarları aksi belirtilmedikçe teklif tarihinden itibaren geçerlidir.</span>
        <span class="id-thanks">İş birliğiniz için teşekkürler!</span>
      </div>
    </div>

    <!-- Audit History -->
    <div v-if="quote" class="max-w-[1040px] mx-auto mt-4 no-print">
      <AuditHistory module="quote" :recordId="quote.id" />
    </div>

    <!-- ═══ PRINT TEMPLATE (Hidden in Web View) ═══ -->
    <div id="quote-print-container" class="print-only">
      <QuotePrintTemplate v-if="quote" :quote="quote" :cari="cari" />
    </div>

  </div>
</template>

<style scoped>
.id-page { min-height: 100vh; background: #f1f5f9; padding-bottom: 60px; }

/* Action bar */
.id-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 20px; background: #fff; border-bottom: 1px solid #e2e8f0;
  gap: 10px; flex-wrap: wrap; position: sticky; top: 0; z-index: 20;
  box-shadow: 0 1px 3px rgba(15,23,42,.06);
}
.id-bar-l { display: flex; align-items: center; gap: 10px; }
.id-bar-r { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.id-back {
  display: flex; align-items: center; gap: 5px; padding: 6px 12px;
  border: 1px solid #cbd5e1; border-radius: 7px; background: transparent;
  cursor: pointer; font-size: 12.5px; color: #64748b; transition: all .15s;
}
.id-back:hover { border-color: #06b6d4; color: #06b6d4; }
.id-bar-title { font-size: 14px; font-weight: 700; color: #0f172a; }
.id-sep { width: 1px; height: 24px; background: #e2e8f0; }

.id-status-sel { font-size: 12.5px; height: 32px; min-width: 140px; }
:deep(.id-status-sel .p-select-label) { padding: 4px 8px; }

.id-btn {
  display: flex; align-items: center; gap: 5px; padding: 6px 13px;
  border-radius: 7px; border: 1px solid; background: transparent;
  font-size: 12.5px; font-weight: 500; cursor: pointer; transition: all .15s;
}
.id-btn:disabled { opacity: .4; cursor: not-allowed; }
.id-btn-warn { border-color: #f59e0b; color: #f59e0b; }
.id-btn-warn:hover { background: rgba(245,158,11,.08); }
.id-btn-ok { border-color: #10b981; color: #10b981; }
.id-btn-ok:hover { background: rgba(16,185,129,.08); }
.id-btn-danger { border-color: #ef4444; color: #ef4444; }
.id-btn-danger:hover { background: rgba(239,68,68,.08); }
.id-btn-export { border-color: #64748b; color: #64748b; }
.id-btn-export:hover { background: rgba(100,116,139,.08); }
.id-btn-pdf { border-color: #06b6d4; color: #06b6d4; }
.id-btn-pdf:hover { background: #ecfeff; }

.id-err { margin: 8px 20px 0; }
.id-conv-banner { margin: 8px 20px 0; }
.id-conv-link { font-weight: bold; text-decoration: underline; margin-left: 6px; }
.id-conv-link:hover { opacity: 0.8; }
.id-loading { text-align: center; padding: 40px; color: #94a3b8; }

/* Document */
.id-doc {
  max-width: 1040px; margin: 16px auto;
  background: #fff; border-radius: 10px;
  border: 1px solid #d1d9e0;
  box-shadow: 0 2px 12px rgba(15,23,42,.07);
  overflow: hidden;
}

.id-head { display: grid; grid-template-columns: 1fr 220px; border-bottom: 1px solid #d1d9e0; }
.id-head-l { padding: 22px 24px; display: flex; flex-direction: column; gap: 5px; border-right: 1px solid #d1d9e0; }
.id-logo-row { display: flex; align-items: center; gap: 12px; margin-bottom: 4px; }
.id-logo { width: 42px; height: 42px; border-radius: 9px; background: #06b6d4; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 18px; flex-shrink: 0; }
.id-co-name { font-size: 15px; font-weight: 800; color: #0f172a; text-transform: uppercase; letter-spacing: .3px; }
.id-co-sub { font-size: 11px; color: #94a3b8; }
.id-divider { border-top: 1px solid #e8edf5; margin: 4px 0; }
.id-to-lbl { font-size: 9.5px; font-weight: 700; letter-spacing: 1.5px; color: #06b6d4; text-transform: uppercase; }
.id-cari-name { font-size: 13px; font-weight: 700; color: #0f172a; margin-top: 2px; }
.id-cari-row { font-size: 11.5px; color: #64748b; display: flex; align-items: center; gap: 5px; }
.id-cari-row i { font-size: 9px; color: #06b6d4; }
.id-head-r { background: #06b6d4; padding: 20px 16px; display: flex; flex-direction: column; gap: 8px; justify-content: center; }
.id-cr { display: flex; align-items: flex-start; gap: 8px; }
.id-cr i { font-size: 11px; color: rgba(255,255,255,.7); margin-top: 2px; flex-shrink: 0; }
.id-cr span { font-size: 11.5px; color: #fff; line-height: 1.4; }

.id-meta { display: flex; align-items: stretch; background: #f8fafc; border-bottom: 3px solid #06b6d4; border-top: 1px solid #d1d9e0; }
.id-mc { padding: 10px 16px; flex: 1; display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.id-ms { width: 1px; background: #d1d9e0; margin: 6px 0; flex-shrink: 0; }
.id-ml { font-size: 9px; font-weight: 700; letter-spacing: 1px; color: #94a3b8; text-transform: uppercase; }
.id-mv { font-size: 13px; font-weight: 700; color: #0f172a; }
.id-mv-total { color: #06b6d4; font-size: 15px; }

.id-tbl-wrap { overflow-x: auto; border-top: 1px solid #d1d9e0; }
.id-tbl { width: 100%; border-collapse: collapse; table-layout: fixed; font-size: 12px; }
.ith {
  background: #0e7490; color: #fff;
  font-size: 10px; font-weight: 700; letter-spacing: .6px; text-transform: uppercase;
  padding: 9px 8px; border-right: 1px solid rgba(255,255,255,.15);
  white-space: nowrap; overflow: hidden;
}
.ith:last-child { border-right: none; }
.ith.c { text-align: center; }
.ith.r { text-align: right; padding-right: 10px; }
.itr { border-bottom: 1px solid #e2e8f0; }
.itr-even { background: #fafcff; }
.itr:hover { background: #f0fbfd; }
.itd { padding: 8px 8px; vertical-align: middle; color: #334155; border-right: 1px solid #e2e8f0; overflow: hidden; }
.itd:last-child { border-right: none; }
.itd.c { text-align: center; }
.itd.r { text-align: right; padding-right: 10px; }
.itd-no { color: #94a3b8; font-weight: 700; font-size: 11px; padding-left: 10px; }
.itd-desc { font-size: 12.5px; color: #0f172a; font-weight: 500; line-height: 1.4; }
.itd-disc { color: #d97706; font-weight: 600; }
.itd-tot { font-weight: 700; color: #0f172a; }

.id-bot { display: grid; grid-template-columns: 1fr 280px; border-top: 1px solid #d1d9e0; }
.id-bot-l { padding: 20px 24px; border-right: 1px solid #d1d9e0; }
.id-sec-lbl { font-size: 9.5px; font-weight: 700; letter-spacing: 1.5px; color: #06b6d4; text-transform: uppercase; margin-bottom: 6px; }
.id-note { font-size: 12.5px; color: #475569; line-height: 1.6; white-space: pre-line; }
.id-sig { margin-top: 20px; }
.id-sig-name { font-size: 12px; font-weight: 700; color: #0f172a; margin-bottom: 4px; }
.id-sig-line { width: 160px; border-bottom: 1px solid #94a3b8; padding-bottom: 18px; margin-bottom: 5px; }
.id-sig-role { font-size: 10.5px; color: #94a3b8; }

.id-bot-r { padding: 20px 22px; display: flex; flex-direction: column; justify-content: space-between; }
.id-tots { display: flex; flex-direction: column; gap: 7px; }
.id-trow { display: flex; justify-content: space-between; align-items: center; font-size: 12.5px; color: #64748b; gap: 16px; }
.id-trow-disc { color: #dc2626; }
.id-tsep { border-top: 1px solid #e2e8f0; margin: 3px 0; }
.id-tgrand {
  display: flex; justify-content: space-between; align-items: center;
  background: #ecfeff; border: 1.5px solid #06b6d4; border-radius: 8px; padding: 10px 12px; gap: 16px;
}
.id-tgrand span:first-child { font-size: 10px; font-weight: 700; letter-spacing: 1px; text-transform: uppercase; color: #0e7490; }
.id-tgval { font-size: 17px; font-weight: 900; color: #0e7490; font-family: 'SFProNumbers', monospace; }
.id-deco {
  text-align: right; font-size: 32px; font-weight: 900;
  color: rgba(6,182,212,.1); letter-spacing: 4px; text-transform: uppercase;
  margin-top: 12px; line-height: 1; user-select: none;
}
.id-terms {
  display: flex; justify-content: space-between; align-items: center;
  background: #f8fafc; border-top: 1px solid #d1d9e0;
  padding: 8px 24px; font-size: 10.5px; color: #94a3b8; gap: 16px; flex-wrap: wrap;
}
.id-thanks { font-weight: 600; color: #06b6d4; white-space: nowrap; }

/* Print */
.print-only { display: none; }
@media print {
  .no-print { display: none !important; }
  .id-page { background: #fff !important; padding: 0 !important; }
  .print-only { display: block !important; }
}

/* Dark mode */
:root.p-dark .id-page { background: #0b0f1a; }
:root.p-dark .id-bar { background: #111827; border-color: #1f2937; }
:root.p-dark .id-bar-title { color: #f1f5f9; }
:root.p-dark .id-doc { background: #111827; border-color: #1f2937; }
:root.p-dark .id-head { border-color: #1f2937; }
:root.p-dark .id-head-l { border-color: #1f2937; }
:root.p-dark .id-co-name { color: #f1f5f9; }
:root.p-dark .id-cari-name { color: #f1f5f9; }
:root.p-dark .id-divider { border-color: #1f2937; }
:root.p-dark .id-meta { background: #0f172a; border-color: #1f2937; }
:root.p-dark .id-ms { background: #1f2937; }
:root.p-dark .id-mv { color: #f1f5f9; }
:root.p-dark .ith { background: #0c4a58; }
:root.p-dark .itr { border-color: #1f2937; }
:root.p-dark .itr-even { background: #0f172a; }
:root.p-dark .itr:hover { background: #0c1a2e; }
:root.p-dark .itd { border-color: #1f2937; color: #e2e8f0; }
:root.p-dark .itd-desc { color: #f1f5f9; }
:root.p-dark .itd-disc { color: #fbbf24; }
:root.p-dark .itd-tot { color: #f1f5f9; }
:root.p-dark .id-bot { border-color: #1f2937; }
:root.p-dark .id-bot-l { border-color: #1f2937; }
:root.p-dark .id-sig-name { color: #f1f5f9; }
:root.p-dark .id-tsep { border-color: #1f2937; }
:root.p-dark .id-tgrand { background: rgba(6,182,212,.1); border-color: #06b6d4; }
:root.p-dark .id-tgval { color: #06b6d4; }
:root.p-dark .id-terms { background: #0f172a; border-color: #1f2937; }
</style>
