<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAnnouncementStore } from '@/stores/announcement'
import { useSuperadminStore } from '@/stores/superadmin'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'

const announcementStore = useAnnouncementStore()
const superadminStore = useSuperadminStore()
const toast = useToast()
const { t } = useI18n()

const ALL_PLANS = 'all'

const title = ref('')
const body = ref('')
const targetPlanID = ref<string>(ALL_PLANS)
const titleSearch = ref('')
const category = ref('bilgi')
const newCategoryName = ref('')

// Kategoriler artık backend'den dinamik gelir (superadmin ekler/siler).
const categoryOptions = computed(() =>
  announcementStore.categories.map((c) => ({ label: c.name, value: c.slug }))
)

const planOptions = ref<{ label: string; value: string }[]>([
  { label: t('superadmin.announcements.targetAll'), value: ALL_PLANS }
])

onMounted(async () => {
  await announcementStore.fetchAll()
  await announcementStore.fetchCategories()
  if (!categoryOptions.value.some((c) => c.value === category.value)) {
    category.value = categoryOptions.value[0]?.value || 'bilgi'
  }
  await superadminStore.fetchPlans()
  if (superadminStore.plans) {
    planOptions.value = [
      { label: t('superadmin.announcements.targetAll'), value: ALL_PLANS },
      ...superadminStore.plans.map((p: any) => ({
        label: p.name,
        value: p.id
      }))
    ]
  }
})

const handleCreate = async () => {
  if (!title.value.trim() || !body.value.trim()) {
    toast.add({ severity: 'warn', summary: t('superadmin.announcements.title'), detail: t('superadmin.announcements.warnRequired'), life: 10000 })
    return
  }

  try {
    await announcementStore.create({
      title: title.value,
      body: body.value,
      category: category.value,
      target_plan_id: targetPlanID.value === ALL_PLANS ? null : targetPlanID.value
    })
    title.value = ''
    body.value = ''
    category.value = categoryOptions.value[0]?.value || 'bilgi'
    targetPlanID.value = ALL_PLANS
    toast.add({ severity: 'success', summary: t('superadmin.announcements.publish'), detail: t('superadmin.announcements.successPublish'), life: 10000 })
  } catch {
    toast.add({ severity: 'error', summary: t('superadmin.announcements.publish'), detail: t('superadmin.announcements.errorPublish'), life: 10000 })
  }
}

const handleDelete = async (id: string) => {
  if (confirm(t('superadmin.announcements.confirmDelete'))) {
    try {
      await announcementStore.delete(id)
      toast.add({ severity: 'success', summary: t('superadmin.announcements.publish'), detail: t('superadmin.announcements.successDelete'), life: 10000 })
    } catch {
      toast.add({ severity: 'error', summary: t('superadmin.announcements.publish'), detail: t('superadmin.announcements.errorDelete'), life: 10000 })
    }
  }
}

const getPlanName = (planId: string | null) => {
  if (!planId) return t('superadmin.announcements.allPlans')
  const found = planOptions.value.find(p => p.value === planId)
  return found ? found.label : t('superadmin.announcements.customPlan')
}

const getCategoryName = (value: string) => {
  return categoryOptions.value.find((item) => item.value === value)?.label || value || 'Bilgi'
}

const handleAddCategory = async () => {
  const name = newCategoryName.value.trim()
  if (!name) return
  try {
    await announcementStore.createCategory(name)
    newCategoryName.value = ''
    toast.add({ severity: 'success', summary: 'Kategori', detail: 'Kategori eklendi', life: 8000 })
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Kategori eklenemedi'
    toast.add({ severity: 'error', summary: 'Kategori', detail: msg, life: 8000 })
  }
}

const handleDeleteCategory = async (cat: { id: string; name: string; slug: string }) => {
  if (!confirm(`"${cat.name}" kategorisini silmek istediğinize emin misiniz? Bu kategorideki duyurular "Bilgi" olarak işaretlenecek.`)) return
  try {
    await announcementStore.deleteCategory(cat.id)
    if (category.value === cat.slug) {
      category.value = categoryOptions.value[0]?.value || 'bilgi'
    }
    toast.add({ severity: 'success', summary: 'Kategori', detail: 'Kategori silindi', life: 8000 })
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || 'Kategori silinemedi'
    toast.add({ severity: 'error', summary: 'Kategori', detail: msg, life: 8000 })
  }
}

const filteredAnnouncements = computed(() => {
  const query = titleSearch.value.trim().toLocaleLowerCase('tr-TR')
  if (!query) return announcementStore.announcements

  return announcementStore.announcements.filter((announcement) =>
    announcement.title.toLocaleLowerCase('tr-TR').includes(query)
  )
})
</script>

<template>
  <div class="announcements-management p-6 flex flex-col gap-6">
    <div class="announcement-search-bar">
      <div>
        <label for="announcement-title-search">Duyuru başlığı ara</label>
        <span>{{ filteredAnnouncements.length }} / {{ announcementStore.announcements.length }} duyuru</span>
      </div>

      <div class="title-search">
        <i class="pi pi-search"></i>
        <InputText
          id="announcement-title-search"
          v-model="titleSearch"
          placeholder="Başlığa göre ara"
          class="w-full"
        />
      </div>
    </div>

    <div class="flex flex-col md:flex-row gap-6">
      <!-- Create Announcement Form -->
      <Card class="w-full md:w-1/3 shrink-0">
        <template #title>
          <div class="text-lg font-bold text-slate-800 dark:text-white flex items-center gap-2">
            <i class="pi pi-megaphone text-indigo-500"></i>
            {{ $t('superadmin.announcements.title') }}
          </div>
        </template>
        <template #content>
          <div class="flex flex-col gap-4">
            <div class="flex flex-col gap-1">
              <label class="text-sm font-bold text-slate-700 dark:text-slate-200">{{ $t('superadmin.announcements.historyTitle') }}</label>
              <InputText v-model="title" :placeholder="$t('superadmin.announcements.titlePlaceholder')" class="w-full" maxlength="255" />
            </div>

            <div class="flex flex-col gap-1">
              <label class="text-sm font-bold text-slate-700 dark:text-slate-200">{{ $t('superadmin.announcements.targetPlan') }}</label>
              <Select v-model="targetPlanID" :options="planOptions" optionLabel="label" optionValue="value" class="w-full" />
            </div>

            <div class="flex flex-col gap-1">
              <label class="text-sm font-bold text-slate-700 dark:text-slate-200">Kategori</label>
              <Select v-model="category" :options="categoryOptions" optionLabel="label" optionValue="value" class="w-full" placeholder="Kategori seçin" />
            </div>

            <!-- Kategori yönetimi: ekle / sil -->
            <div class="flex flex-col gap-2 rounded-lg border border-slate-200 dark:border-slate-700 p-3">
              <label class="text-xs font-bold uppercase tracking-wide text-slate-500 dark:text-slate-400">Kategori Yönetimi</label>
              <div class="flex gap-2">
                <InputText v-model="newCategoryName" placeholder="Yeni kategori adı" class="flex-1" maxlength="64" @keyup.enter="handleAddCategory" />
                <Button icon="pi pi-plus" severity="success" outlined @click="handleAddCategory" :disabled="!newCategoryName.trim()" />
              </div>
              <div class="flex flex-wrap gap-2 mt-1">
                <span
                  v-for="cat in announcementStore.categories"
                  :key="cat.id"
                  class="category-chip"
                  :class="`category-${cat.slug}`"
                >
                  {{ cat.name }}
                  <button
                    v-if="cat.slug !== 'bilgi'"
                    type="button"
                    class="chip-remove"
                    title="Sil"
                    @click="handleDeleteCategory(cat)"
                  >
                    <i class="pi pi-times"></i>
                  </button>
                </span>
              </div>
            </div>

            <div class="flex flex-col gap-1">
              <label class="text-sm font-bold text-slate-700 dark:text-slate-200">{{ $t('superadmin.announcements.body') }}</label>
              <Textarea v-model="body" rows="6" :placeholder="$t('superadmin.announcements.bodyPlaceholder')" class="w-full" />
            </div>

            <Button :label="$t('superadmin.announcements.publish')" icon="pi pi-send" @click="handleCreate" :loading="announcementStore.loading" class="mt-2 w-full" outlined severity="success" />
          </div>
        </template>
      </Card>

      <!-- History / List -->
      <Card class="flex-1">
        <template #title>
          <div class="text-lg font-bold text-slate-800 dark:text-white flex items-center gap-2">
            <i class="pi pi-history text-indigo-500"></i>
            {{ $t('superadmin.announcements.historyTitle') }}
          </div>
        </template>
        <template #content>
          <DataTable
            :value="filteredAnnouncements"
            responsiveLayout="scroll"
            class="p-datatable-sm w-full"
            :emptyMessage="titleSearch ? 'Aramaya uygun duyuru bulunamadı.' : 'Henüz duyuru yok.'"
          >
            <Column field="title" :header="$t('superadmin.announcements.headerTitle')" sortable></Column>
            <Column field="category" header="Kategori" sortable>
              <template #body="{ data }">
                <span class="category-badge" :class="`category-${data.category || 'bilgi'}`">
                  {{ getCategoryName(data.category || 'bilgi') }}
                </span>
              </template>
            </Column>
            <Column field="target_plan_id" :header="$t('superadmin.announcements.headerTarget')" sortable>
              <template #body="{ data }">
                <span class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-xs font-semibold">
                  {{ getPlanName(data.target_plan_id) }}
                </span>
              </template>
            </Column>
            <Column field="created_at" :header="$t('superadmin.announcements.headerDate')" sortable>
              <template #body="{ data }">
                <span>{{ new Date(data.created_at).toLocaleString('tr-TR') }}</span>
              </template>
            </Column>
            <Column :header="$t('superadmin.announcements.headerAction')" style="width: 100px" bodyClass="text-center">
              <template #body="{ data }">
                <Button icon="pi pi-trash" class="p-button-text p-button-sm rounded-md" @click="handleDelete(data.id)" severity="danger" />
              </template>
            </Column>
          </DataTable>
        </template>
      </Card>
    </div>
  </div>
</template>

<style scoped>
/* Kategori rozeti — duyuru listesindeki Kategori sütununda renkli etiket. */
.category-badge {
  display: inline-block;
  padding: 0.12rem 0.55rem;
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 800;
  line-height: 1.5;
  border: 1px solid transparent;
}

.category-bilgi {
  color: #0369a1;
  background: rgba(2, 132, 199, 0.12);
  border-color: rgba(2, 132, 199, 0.25);
}

.category-egitim {
  color: #047857;
  background: rgba(5, 150, 105, 0.12);
  border-color: rgba(5, 150, 105, 0.25);
}

.category-hata {
  color: #b91c1c;
  background: rgba(220, 38, 38, 0.12);
  border-color: rgba(220, 38, 38, 0.25);
}

.category-ozellik {
  color: #7c3aed;
  background: rgba(124, 58, 237, 0.12);
  border-color: rgba(124, 58, 237, 0.25);
}

/* Kategori yönetimi chip'i (ekle/sil paneli) */
.category-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.18rem 0.6rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 800;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  background: var(--p-content-background, #f8fafc);
  color: var(--p-text-color, #334155);
}

.chip-remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border: none;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.08);
  color: inherit;
  cursor: pointer;
  font-size: 0.6rem;
  line-height: 1;
}

.chip-remove:hover {
  background: rgba(220, 38, 38, 0.18);
  color: #b91c1c;
}

.announcement-search-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  border-radius: 8px;
  background: var(--p-content-background, #ffffff);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
}

.announcement-search-bar label {
  display: block;
  color: var(--p-text-color, #0f172a);
  font-size: 0.95rem;
  font-weight: 800;
}

.announcement-search-bar span {
  display: block;
  margin-top: 0.2rem;
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.82rem;
  font-weight: 700;
}

.title-search {
  width: min(100%, 320px);
  position: relative;
}

.title-search i {
  position: absolute;
  left: 0.8rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--p-text-muted-color, #64748b);
  z-index: 1;
}

.title-search :deep(.p-inputtext) {
  padding-left: 2.35rem;
  border-radius: 8px;
}

@media (max-width: 720px) {
  .announcement-search-bar {
    align-items: stretch;
    flex-direction: column;
  }

  .title-search {
    width: 100%;
  }
}
</style>
