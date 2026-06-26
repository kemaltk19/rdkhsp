<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useAuthStore } from '@/stores/auth'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'
import ToggleSwitch from 'primevue/toggleswitch'
import Button from 'primevue/button'

const settingsStore = useSettingsStore()
const authStore = useAuthStore()
const toast = useToast()
const { t } = useI18n()

const modules = ref([
  { id: 'caris', name: 'Cari Kartlar', desc: 'Müşteri ve tedarikçilerin borç/alacak takibi.' },
  { id: 'invoices', name: 'Faturalar', desc: 'Satış ve alış faturalarının düzenlenmesi.' },
  { id: 'quotes', name: 'Teklifler', desc: 'Müşterilere teklif/proforma belgesi oluşturma.' },
  { id: 'payments', name: 'Tahsilat ve Ödemeler', desc: 'Nakit, banka ve pos hareketleri.' },
  { id: 'cash', name: 'Kasa / Banka', desc: 'Kasa ve banka hesap tanımları, virman işlemleri.' },
  { id: 'expenses', name: 'Giderler', desc: 'Şirket içi masraf ve harcama belgeleri.' },
  { id: 'products', name: 'Ürün ve Stok', desc: 'Stok takibi, ürün tanımları ve depolar.' },
  { id: 'reports', name: 'Raporlar', desc: 'Kâr-zarar, KDV ve özet finansal analizler.' },
  { id: 'employees', name: 'Personel Yönetimi', desc: 'Personel kayıtları, maaş ve avans takibi.' },
  { id: 'projects', name: 'Proje Yönetimi', desc: 'Proje takibi, personel ataması ve durum yönetimi.' }
])
// Not: Dashboard ve Ayarlar kasıtlı olarak listede yok — her zaman açık kalmalı,
// aksi halde admin modülleri kapatıp kendi hesabını kilitleyebilir.

const enabledList = ref<string[]>([])

onMounted(async () => {
  const comp = await settingsStore.fetchCompanyProfile()
  if (comp) {
    if (comp.enabled_modules) {
      try {
        enabledList.value = JSON.parse(comp.enabled_modules)
      } catch {
        enabledList.value = modules.value.map(m => m.id)
      }
    } else {
      // Default all enabled if null
      enabledList.value = modules.value.map(m => m.id)
    }
  }
})

const isModuleChecked = (id: string) => {
  return enabledList.value.includes(id)
}

const toggleModule = (id: string) => {
  if (enabledList.value.includes(id)) {
    enabledList.value = enabledList.value.filter(x => x !== id)
  } else {
    enabledList.value.push(id)
  }
}

const handleSave = async () => {
  try {
    await settingsStore.updateEnabledModules(enabledList.value)
    toast.add({ severity: 'success', summary: t('superadmin.modules.title'), detail: t('superadmin.modules.successSave'), life: 10000 })
    // Reload page to reflect menu changes or update layout state
    setTimeout(() => {
      window.location.reload()
    }, 1000)
  } catch {
    toast.add({ severity: 'error', summary: t('superadmin.modules.title'), detail: t('superadmin.modules.errorSave'), life: 10000 })
  }
}
</script>

<template>
  <div class="modules-tab">
    <div class="modules-grid">
      <div v-for="m in modules" :key="m.id" class="module-row">
        <div>
          <div class="module-name">{{ m.name }}</div>
          <div class="module-desc">{{ m.desc }}</div>
        </div>
        <ToggleSwitch :modelValue="isModuleChecked(m.id)" @update:modelValue="toggleModule(m.id)" />
      </div>
    </div>

    <div class="flex justify-end mt-3">
      <Button :label="$t('superadmin.modules.save')" icon="pi pi-check" size="small" class="w-full sm:w-auto" @click="handleSave" :loading="settingsStore.loading" outlined severity="primary" />
    </div>
  </div>
</template>

<style scoped>
.modules-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0 1.25rem;
}
@media (max-width: 900px) {
  .modules-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 560px) {
  .modules-grid { grid-template-columns: 1fr; }
}
.module-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.45rem 0;
  border-bottom: 1px solid #f0f2f5;
}
.module-row:last-child { border-bottom: none; }
:root.p-dark .module-row { border-bottom-color: rgba(255,255,255,0.05); }
.module-name { font-size: 0.8rem; font-weight: 600; color: #1a202c; }
:root.p-dark .module-name { color: #f1f5f9; }
.module-desc { font-size: 0.68rem; color: #718096; margin-top: 0.05rem; }
</style>
