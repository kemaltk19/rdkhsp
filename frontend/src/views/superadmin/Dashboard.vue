<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-semibold text-surface-900 dark:text-surface-0 tracking-tight">{{ $t('superadmin.dashboard.title') }}</h1>
      <Button icon="pi pi-refresh" :label="$t('superadmin.plans.refresh')" @click="fetchStats" :loading="loading" class="p-button-outlined" />
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <!-- Total Companies -->
      <div class="bg-white dark:bg-slate-800 shadow rounded-2xl p-6! border border-slate-100 dark:border-slate-700 relative overflow-hidden group">
        <div class="absolute -right-6 -top-6 bg-primary-100 dark:bg-primary-900/40 rounded-full w-32 h-32 blur-2xl opacity-50 transition-transform group-hover:scale-150 duration-500"></div>
        <div class="flex items-center justify-between mb-4 relative z-10">
          <span class="text-slate-500 dark:text-slate-400 font-medium">{{ $t('superadmin.dashboard.totalCompanies') }}</span>
          <div class="w-10 h-10 flex items-center justify-center bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 rounded-xl">
            <i class="pi pi-building text-xl"></i>
          </div>
        </div>
        <div class="text-4xl font-bold text-slate-900 dark:text-white relative z-10">
          <Skeleton v-if="loading" width="4rem" height="2.5rem" />
          <span v-else>{{ stats?.total_companies || 0 }}</span>
        </div>
      </div>

      <!-- Active Companies -->
      <div class="bg-white dark:bg-slate-800 shadow rounded-2xl p-6! border border-slate-100 dark:border-slate-700 relative overflow-hidden group">
        <div class="absolute -right-6 -top-6 bg-green-100 dark:bg-green-900/40 rounded-full w-32 h-32 blur-2xl opacity-50 transition-transform group-hover:scale-150 duration-500"></div>
        <div class="flex items-center justify-between mb-4 relative z-10">
          <span class="text-slate-500 dark:text-slate-400 font-medium">{{ $t('superadmin.dashboard.activeCompanies') }}</span>
          <div class="w-10 h-10 flex items-center justify-center bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400 rounded-xl">
            <i class="pi pi-check-circle text-xl"></i>
          </div>
        </div>
        <div class="text-4xl font-bold text-slate-900 dark:text-white relative z-10">
          <Skeleton v-if="loading" width="4rem" height="2.5rem" />
          <span v-else>{{ stats?.active_companies || 0 }}</span>
        </div>
      </div>

      <!-- Trial Companies -->
      <div class="bg-white dark:bg-slate-800 shadow rounded-2xl p-6! border border-slate-100 dark:border-slate-700 relative overflow-hidden group">
        <div class="absolute -right-6 -top-6 bg-amber-100 dark:bg-amber-900/40 rounded-full w-32 h-32 blur-2xl opacity-50 transition-transform group-hover:scale-150 duration-500"></div>
        <div class="flex items-center justify-between mb-4 relative z-10">
          <span class="text-slate-500 dark:text-slate-400 font-medium">{{ $t('superadmin.dashboard.trialCompanies') }}</span>
          <div class="w-10 h-10 flex items-center justify-center bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 rounded-xl">
            <i class="pi pi-clock text-xl"></i>
          </div>
        </div>
        <div class="text-4xl font-bold text-slate-900 dark:text-white relative z-10">
          <Skeleton v-if="loading" width="4rem" height="2.5rem" />
          <span v-else>{{ stats?.trial_companies || 0 }}</span>
        </div>
      </div>

      <!-- Total Users -->
      <div class="bg-white dark:bg-slate-800 shadow rounded-2xl p-6! border border-slate-100 dark:border-slate-700 relative overflow-hidden group">
        <div class="absolute -right-6 -top-6 bg-indigo-100 dark:bg-indigo-900/40 rounded-full w-32 h-32 blur-2xl opacity-50 transition-transform group-hover:scale-150 duration-500"></div>
        <div class="flex items-center justify-between mb-4 relative z-10">
          <span class="text-slate-500 dark:text-slate-400 font-medium">{{ $t('superadmin.dashboard.totalUsers') }}</span>
          <div class="w-10 h-10 flex items-center justify-center bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-xl">
            <i class="pi pi-users text-xl"></i>
          </div>
        </div>
        <div class="text-4xl font-bold text-slate-900 dark:text-white relative z-10">
          <Skeleton v-if="loading" width="4rem" height="2.5rem" />
          <span v-else>{{ stats?.total_users || 0 }}</span>
        </div>
      </div>

      <!-- Monthly Recurring Revenue (MRR) -->
      <div class="bg-white dark:bg-slate-800 shadow rounded-2xl p-6! border border-slate-100 dark:border-slate-700 relative overflow-hidden group">
        <div class="absolute -right-6 -top-6 bg-cyan-100 dark:bg-cyan-900/40 rounded-full w-32 h-32 blur-2xl opacity-50 transition-transform group-hover:scale-150 duration-500"></div>
        <div class="flex items-center justify-between mb-4 relative z-10">
          <span class="text-slate-500 dark:text-slate-400 font-medium">{{ $t('superadmin.dashboard.mrr') }}</span>
          <div class="w-10 h-10 flex items-center justify-center bg-cyan-100 dark:bg-cyan-900/30 text-cyan-600 dark:text-cyan-400 rounded-xl">
            <i class="pi pi-money-bill text-xl"></i>
          </div>
        </div>
        <div class="text-3xl font-bold text-slate-900 dark:text-white relative z-10">
          <Skeleton v-if="loading" width="6rem" height="2.5rem" />
          <span v-else>{{ stats?.mrr || 0 }} TRY</span>
        </div>
      </div>

      <!-- Churn (Past 30 days) -->
      <div class="bg-white dark:bg-slate-800 shadow rounded-2xl p-6! border border-slate-100 dark:border-slate-700 relative overflow-hidden group">
        <div class="absolute -right-6 -top-6 bg-rose-100 dark:bg-rose-900/40 rounded-full w-32 h-32 blur-2xl opacity-50 transition-transform group-hover:scale-150 duration-500"></div>
        <div class="flex items-center justify-between mb-4 relative z-10">
          <span class="text-slate-500 dark:text-slate-400 font-medium">{{ $t('superadmin.dashboard.churn') }}</span>
          <div class="w-10 h-10 flex items-center justify-center bg-rose-100 dark:bg-rose-900/30 text-rose-600 dark:text-rose-400 rounded-xl">
            <i class="pi pi-user-minus text-xl"></i>
          </div>
        </div>
        <div class="text-4xl font-bold text-slate-900 dark:text-white relative z-10">
          <Skeleton v-if="loading" width="4rem" height="2.5rem" />
          <span v-else>{{ stats?.churn || 0 }}</span>
        </div>
      </div>
    </div>

    <!-- Plan Distribution and Recent Companies -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="bg-white dark:bg-slate-800 shadow rounded-2xl border border-slate-100 dark:border-slate-700 p-6 flex flex-col gap-4">
        <h2 class="text-xl font-semibold text-slate-900 dark:text-white">{{ $t('superadmin.dashboard.planDistribution') }}</h2>
        <div v-if="loading">
          <Skeleton width="100%" height="150px" />
        </div>
        <div v-else class="flex flex-col gap-3">
          <div v-for="plan in stats?.plan_distribution" :key="plan.plan_name" class="flex justify-between items-center p-3 bg-slate-50 dark:bg-slate-900/40 rounded-lg">
            <span class="font-medium text-slate-700 dark:text-slate-300">{{ plan.plan_name }}</span>
            <span class="px-2.5 py-1 bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 font-bold text-xs rounded-full">
              {{ plan.count }} Şirket
            </span>
          </div>
        </div>
      </div>

      <div class="lg:col-span-2 bg-white dark:bg-slate-800 shadow rounded-2xl border border-slate-100 dark:border-slate-700 overflow-hidden">
      <div class="p-6! border-b border-slate-100 dark:border-slate-700 flex justify-between items-center">
        <h2 class="text-xl font-semibold text-slate-900 dark:text-white">{{ $t('superadmin.dashboard.recentCompanies') }}</h2>
        <Button :label="$t('superadmin.dashboard.viewAll')" icon="pi pi-arrow-right" iconPos="right" text @click="router.push('/superadmin/companies')" />
      </div>
      
      <div v-if="loading" class="p-6">
        <Skeleton width="100%" height="150px" />
      </div>
      
      <DataTable v-else :value="stats?.recent_companies" responsiveLayout="scroll" :pt="{ root: { class: 'border-none' } }">
        <Column field="name" :header="$t('superadmin.companies.name')">
          <template #body="{ data }">
            <div class="flex items-center gap-3">
              <Avatar :label="data.name.substring(0, 2).toUpperCase()" shape="circle" class="bg-primary-100 text-primary-700 font-bold" />
              <div class="flex flex-col">
                <span class="font-medium text-surface-900 dark:text-surface-0">{{ data.name }}</span>
                <span class="text-sm text-surface-500 dark:text-surface-400">{{ data.admin_email }}</span>
              </div>
            </div>
          </template>
        </Column>
        <Column field="subscription_status" :header="$t('superadmin.companies.status')">
          <template #body="{ data }">
            <Tag :severity="getSubscriptionSeverity(data.subscription_status)" :value="data.subscription_status.toUpperCase()" class="text-xs font-semibold px-2 py-1" />
          </template>
        </Column>
        <Column field="stats.total_users" :header="$t('superadmin.companies.users')" />
        <Column field="stats.total_invoices" :header="$t('superadmin.companies.invoices')" />
        <Column field="created_at" header="Kayıt Tarihi">
          <template #body="{ data }">
            {{ formatDate(data.created_at) }}
          </template>
        </Column>
      </DataTable>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboardStatsApi } from '@/api/superadmin'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const toast = useToast()
const loading = ref(true)
const stats = ref<any>(null)

const fetchStats = async () => {
  loading.value = true
  try {
    const res = await getDashboardStatsApi()
    stats.value = res.data.data
  } catch (error: any) {
    console.error('Superadmin Dashboard error:', error)
    toast.add({
      severity: 'error',
      summary: 'Hata',
      detail: 'İstatistikler alınırken bir hata oluştu.',
      life: 10000
    })
  } finally {
    loading.value = false
  }
}

const getSubscriptionSeverity = (status: string) => {
  switch (status) {
    case 'active':
      return 'success'
    case 'trial':
      return 'warning'
    case 'canceled':
      return 'danger'
    default:
      return 'info'
  }
}

const formatDate = (val: string) => {
  if (!val) return ''
  return new Date(val).toLocaleDateString('tr-TR', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

onMounted(() => {
  fetchStats()
})
</script>
