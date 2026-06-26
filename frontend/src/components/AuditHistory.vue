<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuditStore } from '@/stores/audit'

const props = defineProps<{
  module: string
  recordId: string
}>()

const auditStore = useAuditStore()
const loading = ref(false)

const loadLogs = async () => {
  loading.value = true
  try {
    await auditStore.fetchLogs({ module: props.module, record_id: props.recordId })
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadLogs()
})

const getActionLabel = (act: string) => {
  switch (act) {
    case 'create': return 'Oluşturuldu'
    case 'update': return 'Güncellendi'
    case 'delete': return 'Silindi'
    case 'cancel': return 'İptal Edildi'
    default: return act
  }
}

const getActionSeverityClass = (act: string) => {
  switch (act) {
    case 'create': return 'badge-create'
    case 'update': return 'badge-update'
    case 'delete': return 'badge-delete'
    case 'cancel': return 'badge-cancel'
    default: return 'badge-other'
  }
}

const fmtDate = (d: string) => {
  return new Date(d).toLocaleString('tr-TR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <div class="audit-history">
    <div class="ah-title flex justify-between items-center mb-4">
      <h3 class="text-sm font-bold tracking-wide uppercase text-slate-500 dark:text-slate-400">İşlem Geçmişi</h3>
      <button @click="loadLogs" class="refresh-btn" :disabled="loading">
        <i class="pi pi-refresh" :class="{ 'pi-spin': loading }"></i>
      </button>
    </div>

    <div v-if="loading && auditStore.logs.length === 0" class="ah-loading text-center py-4 text-slate-400">
      <i class="pi pi-spin pi-spinner mr-2"></i> Yükleniyor...
    </div>

    <div v-else-if="auditStore.logs.length === 0" class="ah-empty text-center py-4 text-slate-400 text-xs">
      Henüz işlem geçmişi kaydı bulunmuyor.
    </div>

    <div v-else class="ah-timeline">
      <div v-for="log in auditStore.logs" :key="log.id" class="ah-item">
        <div class="ah-badge" :class="getActionSeverityClass(log.action)">
          {{ getActionLabel(log.action) }}
        </div>
        <div class="ah-details">
          <div class="ah-meta">
            <span class="ah-user font-semibold text-slate-700 dark:text-slate-200">{{ log.user_name }}</span>
            <span v-if="log.user_role" class="ah-role text-xs text-slate-400"> ({{ log.user_role }})</span>
            <span class="ah-date text-xs text-slate-400 ml-auto">{{ fmtDate(log.created_at) }}</span>
          </div>
          <div class="ah-summary text-xs text-slate-500 mt-1" v-if="log.summary">
            Detay: {{ log.summary }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.audit-history {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
}
:root.p-dark .audit-history {
  background: #1e293b;
  border-color: #334155;
}
.refresh-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 4px;
}
.refresh-btn:hover {
  color: #06b6d4;
}
.ah-timeline {
  display: flex;
  flex-direction: column;
  gap: 12px;
  position: relative;
}
.ah-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 10px;
}
:root.p-dark .ah-item {
  border-color: #334155;
}
.ah-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}
.ah-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 9999px;
  text-transform: uppercase;
  min-width: 80px;
  text-align: center;
}
.badge-create { background: #ecfdf5; color: #059669; }
.badge-update { background: #eff6ff; color: #2563eb; }
.badge-cancel { background: #fffbeb; color: #d97706; }
.badge-delete { background: #fef2f2; color: #dc2626; }
.badge-other { background: #f8fafc; color: #64748b; }

:root.p-dark .badge-create { background: rgba(5,150,105,0.15); color: #34d399; }
:root.p-dark .badge-update { background: rgba(37,99,235,0.15); color: #60a5fa; }
:root.p-dark .badge-cancel { background: rgba(217,119,6,0.15); color: #fbbf24; }
:root.p-dark .badge-delete { background: rgba(220,38,38,0.15); color: #f87171; }

.ah-details {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.ah-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  font-size: 12px;
}
</style>
