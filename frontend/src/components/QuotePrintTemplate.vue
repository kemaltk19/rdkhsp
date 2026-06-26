<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { formatDate } from '@/utils/date'
import Money from '@/components/Money.vue'
import { useSettingsStore } from '@/stores/settings'

const props = defineProps<{
  quote: any
  cari: any
}>()

const settingsStore = useSettingsStore()

onMounted(async () => {
  if (!settingsStore.company) {
    await settingsStore.fetchCompanyProfile()
  }
})

// Teklif toplamları: backend'in döviz kuru çevirisiyle hesapladığı ve kaydettiği
// değerler kullanılır (quote_service.go convertToDefaultCurrency). Satırları burada
// tekrar toplamak, farklı para biriminde satırlarda yanlış sonuç üretir.
const subtotal = computed(() => Number(props.quote?.subtotal) || 0)
const taxTotal = computed(() => Number(props.quote?.tax_total) || 0)
const grandTotal = computed(() => Number(props.quote?.total) || 0)

const exchangeRatesInfo = computed(() => {
  const items = props.quote?.items || [];
  const rates = new Map<string, number>();
  for (const item of items) {
    if (item.currency && item.currency.toUpperCase() !== props.quote?.currency?.toUpperCase()) {
      const rate = parseFloat(item.exchange_rate);
      if (!isNaN(rate) && rate > 0) {
        rates.set(item.currency.toUpperCase(), rate);
      }
    }
  }
  return Array.from(rates.entries()).map(([curr, rate]) => ({ curr, rate }));
});

const lSub = (item: any) => item.quantity * item.unit_price
</script>

<template>
  <div class="fp-doc">
    <!-- Header -->
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
        <div class="fp-cari-name">{{ cari?.name || 'Müşteri / Tedarikçi' }}</div>
        <div v-if="cari?.address" class="fp-cari-row"><i class="pi pi-map-marker"></i>{{ cari.address }}</div>
        <div v-if="cari?.phone" class="fp-cari-row"><i class="pi pi-phone"></i>{{ cari.phone }}</div>
        <div v-if="cari?.email" class="fp-cari-row"><i class="pi pi-envelope"></i>{{ cari.email }}</div>
        <div v-if="cari?.tax_number" class="fp-cari-row"><i class="pi pi-id-card"></i>V.D.: {{ cari.tax_office }} | V.N.: {{ cari.tax_number }}</div>
      </div>
      <div class="fp-head-r">
        <div v-if="settingsStore.company?.phone" class="fp-cr"><i class="pi pi-phone"></i><span>{{ settingsStore.company.phone }}</span></div>
        <div v-if="settingsStore.company?.website" class="fp-cr"><i class="pi pi-globe"></i><span>{{ settingsStore.company.website }}</span></div>
        <div v-if="settingsStore.company?.email" class="fp-cr"><i class="pi pi-envelope"></i><span>{{ settingsStore.company.email }}</span></div>
        <div v-if="settingsStore.company?.address" class="fp-cr"><i class="pi pi-map-marker"></i><span>{{ settingsStore.company.address }}</span></div>
        <div v-if="settingsStore.company?.tax_number" class="fp-cr"><i class="pi pi-id-card"></i><span>V.N.: {{ settingsStore.company.tax_number }}</span></div>
      </div>
    </div>

    <!-- Meta bar -->
    <div class="fp-meta">
      <div class="fp-mc">
        <div class="fp-ml">GENEL TOPLAM</div>
        <div class="fp-mv fp-mv-total"><Money :value="grandTotal.toString()" :currency="quote?.currency" /></div>
      </div>
      <div class="fp-ms"></div>
      <div class="fp-mc">
        <div class="fp-ml">TEKLİF NO</div>
        <div class="fp-mv uppercase-input">{{ quote?.number || 'TASLAK' }}</div>
      </div>
      <div class="fp-ms"></div>
      <div class="fp-mc">
        <div class="fp-ml">TEKLİF TARİHİ</div>
        <div class="fp-mv">{{ formatDate(quote?.date) }}</div>
      </div>
      <div class="fp-ms"></div>
      <div class="fp-mc">
        <div class="fp-ml">GEÇERLİLİK TARİHİ</div>
        <div class="fp-mv">{{ formatDate(quote?.expiry_date) }}</div>
      </div>
      <div class="fp-ms"></div>
      <div class="fp-mc">
        <div class="fp-ml">DÖVİZ</div>
        <div class="fp-mv font-bold">{{ quote?.currency }}</div>
      </div>
    </div>

    <!-- Items Table -->
    <div class="fp-tbl-wrap">
      <table class="fp-tbl">
        <colgroup>
          <col style="width:30px">
          <col style="width:auto">
          <col style="width:80px">
          <col style="width:70px">
          <col style="width:105px">
          <col style="width:90px">
          <col style="width:130px">
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
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, idx) in quote?.items" :key="idx" class="ftr">
            <td class="ftd c ftd-no">{{ idx + 1 }}</td>
            <td class="ftd ftd-desc">
              <div class="font-semibold">{{ item.description }}</div>
            </td>
            <td class="ftd c">{{ item.unit }}</td>
            <td class="ftd c">{{ Number(item.quantity) }}</td>
            <td class="ftd r"><Money :value="item.unit_price" :currency="item.currency || quote?.currency" /></td>
            <td class="ftd c">%{{ item.tax_rate }}</td>
            <td class="ftd r ftd-linetot"><Money :value="lSub(item).toString()" :currency="item.currency || quote?.currency" /></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Bottom: Notes + Totals -->
    <div class="fp-bot">
      <div class="fp-bot-l">
        <div class="fp-sec-lbl">NOTLAR / ÖDEME BİLGİLERİ</div>
        <div class="fp-notes whitespace-pre-wrap">{{ quote?.note || '-' }}</div>
        <div class="fp-sig">
          <div class="fp-sig-name">{{ settingsStore.company?.name || '' }}</div>
          <div class="fp-sig-line"></div>
          <div class="fp-sig-role">Yetkili / Authorized</div>
        </div>
      </div>

      <div class="fp-bot-r">
        <div class="fp-tots">
          <div class="fp-trow"><span>Ara Toplam</span><Money :value="subtotal.toString()" :currency="quote?.currency" /></div>
          <div class="fp-trow">
            <span class="whitespace-nowrap">İndirim / Ek (-/+)</span>
            <span class="font-semibold"><Money :value="(quote?.discount_total || 0).toString()" :currency="quote?.currency" /></span>
          </div>
          <div class="fp-trow"><span>KDV</span><Money :value="taxTotal.toString()" :currency="quote?.currency" /></div>
          <div class="fp-tsep"></div>
          <div class="fp-tgrand">
            <span>GENEL TOPLAM</span>
            <span class="fp-tgval"><Money :value="grandTotal.toString()" :currency="quote?.currency" /></span>
          </div>

          <div v-if="exchangeRatesInfo.length > 0" class="mt-3 text-[11px] text-slate-500 bg-slate-50 p-2 rounded border border-slate-200" style="margin-top: 10px; padding: 8px; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 4px; font-size: 10px; color: #64748b;">
            <div style="font-weight: 700; margin-bottom: 4px; color: #475569;">Kur Bilgisi</div>
            <div v-for="r in exchangeRatesInfo" :key="r.curr" style="display: flex; justify-content: space-between; gap: 16px;">
              <span>{{ r.curr }}</span>
              <span>{{ new Intl.NumberFormat('tr-TR', { minimumFractionDigits: 4, maximumFractionDigits: 4 }).format(r.rate) }}</span>
            </div>
          </div>
        </div>
        <div class="fp-deco">TEKLİF</div>
      </div>
    </div>

    <div class="fp-terms">
      <span>Teklif tutarları aksi belirtilmedikçe teklif tarihinden itibaren geçerlidir.</span>
      <span class="fp-thanks">İş birliğiniz için teşekkürler!</span>
    </div>
  </div>
</template>

<style scoped>
/* Document Layout Match - A4 proportions */
.fp-doc {
  background: #fff;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* Header: Split Panel */
.fp-head { display: flex; align-items: stretch; background: #fff; }
.fp-head-l { flex: 1; padding: 24px; display: flex; flex-direction: column; justify-content: center; }
.fp-logo-row { display: flex; align-items: center; gap: 14px; }
.fp-logo { width: 44px; height: 44px; background: #06b6d4; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 20px; }
.fp-co-name { font-size: 18px; font-weight: 800; color: #0f172a; letter-spacing: -0.5px; }
.fp-co-sub { font-size: 12px; color: #64748b; margin-top: 2px; }
.fp-divider { height: 1px; background: #e2e8f0; margin: 20px 0; width: 100%; }
.fp-to-label { font-size: 10px; font-weight: 700; color: #06b6d4; letter-spacing: 1px; margin-bottom: 6px; }
.fp-cari-name { font-size: 14px; font-weight: 700; color: #0f172a; margin-bottom: 6px; }
.fp-cari-row { font-size: 12px; color: #475569; display: flex; align-items: center; gap: 6px; margin-bottom: 4px; }
.fp-cari-row i { color: #94a3b8; font-size: 11px; }

.fp-head-r { width: 280px; background: #06b6d4; padding: 24px; color: #fff; display: flex; flex-direction: column; justify-content: center; gap: 8px; }
.fp-cr { display: flex; align-items: center; gap: 8px; font-size: 12px; font-weight: 500; }
.fp-cr i { opacity: 0.8; font-size: 12px; }

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
.uppercase-input { text-transform: uppercase; }

/* Table Layout */
.fp-tbl-wrap { width: 100%; margin-top: 10px; flex: 1; }
.fp-tbl { width: 100%; border-collapse: collapse; table-layout: fixed; font-size: 13px; }
.fth {
  background: #f8fafc; color: #334155;
  font-size: 11px; font-weight: 700; text-transform: uppercase;
  padding: 10px; border-bottom: 1px solid #e2e8f0; border-top: 1px solid #e2e8f0;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; text-align: left;
}
.fth.c { text-align: center; }
.fth.r { text-align: right; padding-right: 10px; }

.ftr { border-bottom: 1px solid #f1f5f9; }
.ftd { padding: 12px 10px; vertical-align: middle; color: #334155; }
.ftd.c { text-align: center; }
.ftd.r { text-align: right; padding-right: 10px; }
.ftd-no { color: #64748b; font-weight: 700; font-size: 12px; }
.ftd-desc { vertical-align: top; }
.ftd-linetot { font-weight: 700; color: #0f172a; }

/* Bottom */
.fp-bot { display: grid; grid-template-columns: 1fr 280px; border-top: 1px solid #d1d9e0; margin-top: auto; }
.fp-bot-l { padding: 20px 24px; border-right: 1px solid #d1d9e0; }
.fp-sec-lbl { font-size: 9.5px; font-weight: 700; letter-spacing: 1.5px; color: #06b6d4; text-transform: uppercase; margin-bottom: 7px; }
.fp-notes { width: 100%; font-size: 12.5px; color: #475569; line-height: 1.5; }
.fp-sig { margin-top: 40px; }
.fp-sig-name { font-size: 12px; font-weight: 700; color: #0f172a; margin-bottom: 4px; }
.fp-sig-line { width: 160px; border-bottom: 1px solid #94a3b8; padding-bottom: 18px; margin-bottom: 5px; }
.fp-sig-role { font-size: 10.5px; color: #94a3b8; }

.fp-bot-r { padding: 20px 22px; display: flex; flex-direction: column; justify-content: space-between; }
.fp-tots { display: flex; flex-direction: column; gap: 7px; }
.fp-trow { display: flex; justify-content: space-between; align-items: center; font-size: 12.5px; color: #64748b; gap: 16px; }
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
</style>
