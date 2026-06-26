<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAnnouncementStore } from '@/stores/announcement'

const announcementStore = useAnnouncementStore()
const errorMessage = ref('')
const expandedAnnouncements = ref<Record<string, boolean>>({})
const titleSearch = ref('')
const selectedCategory = ref('all') // 'all' = tüm kategoriler
const previewLimit = 260

onMounted(async () => {
  try {
    await announcementStore.fetchForTenant()
    // Filtre etiketleri/seçenekleri için kategorileri çek (kritik değil; hata yutulur).
    try { await announcementStore.fetchCategoriesForTenant() } catch {}
  } catch (error) {
    errorMessage.value = 'Duyurular yuklenirken bir sorun olustu.'
  }
})

const announcements = computed(() => announcementStore.announcements)

// Kategori etiketleri backend'den dinamik; bilinmeyen slug kendi değeriyle gösterilir.
const categoryLabels = computed<Record<string, string>>(() => {
  const map: Record<string, string> = { bilgi: 'Bilgi' }
  for (const c of announcementStore.categories) map[c.slug] = c.name
  return map
})

const getCategoryName = (value?: string) => categoryLabels.value[value || 'bilgi'] || value || 'Bilgi'
const getCategoryKey = (value?: string) => (value && categoryLabels.value[value] ? value : 'bilgi')

// Filtre menüsü: "Tümü" + sadece duyurularda gerçekten kullanılan kategoriler.
const categoryFilterOptions = computed(() => {
  const used = new Set(announcements.value.map((a) => a.category || 'bilgi'))
  const opts = [{ label: 'Tüm kategoriler', value: 'all' }]
  for (const c of announcementStore.categories) {
    if (used.has(c.slug)) opts.push({ label: c.name, value: c.slug })
  }
  return opts
})

const filteredAnnouncements = computed(() => {
  const query = titleSearch.value.trim().toLocaleLowerCase('tr-TR')
  return announcements.value.filter((announcement) => {
    const matchesTitle = !query || announcement.title.toLocaleLowerCase('tr-TR').includes(query)
    const matchesCategory =
      selectedCategory.value === 'all' || (announcement.category || 'bilgi') === selectedCategory.value
    return matchesTitle && matchesCategory
  })
})

const latestAnnouncement = computed(() => announcements.value[0] || null)

const formatDate = (value: string) => {
  return new Date(value).toLocaleDateString('tr-TR', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  })
}

const formatTime = (value: string) => {
  return new Date(value).toLocaleTimeString('tr-TR', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

const isLongBody = (body: string) => body.length > previewLimit

const previewBody = (id: string, body: string) => {
  if (!isLongBody(body) || expandedAnnouncements.value[id]) {
    return body
  }

  return `${body.slice(0, previewLimit).trim()}...`
}

const toggleAnnouncement = (id: string) => {
  expandedAnnouncements.value = {
    ...expandedAnnouncements.value,
    [id]: !expandedAnnouncements.value[id],
  }
}
</script>

<template>
  <div class="announcements-board">
    <section class="board-header">
      <div>
        <span class="eyebrow">Merkez bildirimleri</span>
        <h1>Duyuru Panosu</h1>
        <p>Sistem yoneticilerinden gelen guncellemeler, bakim notlari ve operasyon duyurulari.</p>
      </div>

      <div class="summary-panel">
        <span class="summary-label">Toplam duyuru</span>
        <strong>{{ announcements.length }}</strong>
        <span v-if="latestAnnouncement" class="summary-date">
          Son: {{ formatDate(latestAnnouncement.created_at) }}
        </span>
        <span v-else class="summary-date">Yeni duyuru yok</span>
      </div>
    </section>

    <div v-if="errorMessage" class="state-message error-state">
      <i class="pi pi-exclamation-triangle"></i>
      <span>{{ errorMessage }}</span>
    </div>

    <section v-if="!errorMessage" class="search-panel">
      <div>
        <label for="tenant-announcement-search">Duyuru başlığı ara</label>
        <span>{{ filteredAnnouncements.length }} / {{ announcements.length }} duyuru</span>
      </div>

      <div class="search-controls">
        <div class="category-filter">
          <i class="pi pi-filter"></i>
          <select
            id="tenant-announcement-category"
            v-model="selectedCategory"
            aria-label="Kategoriye göre filtrele"
          >
            <option v-for="opt in categoryFilterOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>

        <div class="title-search">
          <i class="pi pi-search"></i>
          <input
            id="tenant-announcement-search"
            v-model="titleSearch"
            type="search"
            placeholder="Başlığa göre ara"
          />
        </div>
      </div>
    </section>

    <div v-if="!errorMessage && announcementStore.loading" class="announcement-list">
      <article v-for="item in 3" :key="item" class="announcement-card skeleton-card">
        <div class="date-badge skeleton-block"></div>
        <div class="announcement-content">
          <div class="skeleton-line title-line"></div>
          <div class="skeleton-line"></div>
          <div class="skeleton-line short-line"></div>
        </div>
      </article>
    </div>

    <div v-else-if="!errorMessage && announcements.length === 0" class="state-message empty-state">
      <div class="state-icon">
        <i class="pi pi-inbox"></i>
      </div>
      <div>
        <h2>Henuz yayinlanmis bir duyuru yok</h2>
        <p>Yeni bir sistem duyurusu geldiginde burada listelenecek.</p>
      </div>
    </div>

    <div v-else-if="!errorMessage && filteredAnnouncements.length === 0" class="state-message empty-state">
      <div class="state-icon">
        <i class="pi pi-search"></i>
      </div>
      <div>
        <h2>Aramaya uygun duyuru bulunamadı</h2>
        <p>Başlık aramasını değiştirerek tekrar deneyin.</p>
      </div>
    </div>

    <div v-else-if="!errorMessage" class="announcement-list">
      <article
        v-for="ann in filteredAnnouncements"
        :key="ann.id"
        class="announcement-card"
      >
        <div class="date-badge">
          <span>{{ formatDate(ann.created_at).split(' ')[0] }}</span>
          <small>{{ formatDate(ann.created_at).split(' ').slice(1).join(' ') }}</small>
        </div>

        <div class="announcement-content">
          <div class="announcement-meta">
            <span><i class="pi pi-megaphone"></i> Sistem duyurusu</span>
            <span class="category-badge" :class="`category-${getCategoryKey(ann.category)}`">
              {{ getCategoryName(ann.category) }}
            </span>
            <time>{{ formatTime(ann.created_at) }}</time>
          </div>
          <h2>{{ ann.title }}</h2>
          <p>{{ previewBody(ann.id, ann.body) }}</p>
          <button
            v-if="isLongBody(ann.body)"
            type="button"
            class="read-more-btn"
            @click="toggleAnnouncement(ann.id)"
          >
            <span>{{ expandedAnnouncements[ann.id] ? 'Daha az' : 'Daha fazla' }}</span>
            <i :class="expandedAnnouncements[ann.id] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"></i>
          </button>
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.announcements-board {
  max-width: 980px;
  margin: 0 auto;
  padding: 0.25rem 0 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.board-header {
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.5rem;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  border-radius: 8px;
  background:
    linear-gradient(135deg, rgba(6, 182, 212, 0.1), rgba(15, 23, 42, 0.02)),
    var(--p-content-background, #ffffff);
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  margin-bottom: 0.5rem;
  color: #0891b2;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.board-header h1 {
  margin: 0;
  color: var(--p-text-color, #0f172a);
  font-size: 1.75rem;
  line-height: 1.2;
  font-weight: 800;
}

.board-header p {
  max-width: 620px;
  margin: 0.5rem 0 0;
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.95rem;
  line-height: 1.6;
}

.summary-panel {
  min-width: 170px;
  padding: 1rem;
  border: 1px solid rgba(6, 182, 212, 0.22);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.72);
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.summary-label,
.summary-date {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.78rem;
  font-weight: 700;
}

.summary-panel strong {
  color: var(--p-text-color, #0f172a);
  font-size: 2.25rem;
  line-height: 1.1;
}

.search-panel {
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

.search-panel label {
  display: block;
  color: var(--p-text-color, #0f172a);
  font-size: 0.95rem;
  font-weight: 800;
}

.search-panel span {
  display: block;
  margin-top: 0.2rem;
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.82rem;
  font-weight: 700;
}

.search-controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.category-filter {
  position: relative;
}

.category-filter i {
  position: absolute;
  left: 0.7rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--p-text-muted-color, #64748b);
  pointer-events: none;
}

.category-filter select {
  height: 42px;
  padding: 0 2rem 0 2.1rem;
  border: 1px solid var(--p-content-border-color, #cbd5e1);
  border-radius: 8px;
  outline: none;
  background: var(--p-content-background, #ffffff);
  color: var(--p-text-color, #0f172a);
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
  appearance: none;
}

.category-filter select:focus {
  border-color: #0891b2;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.14);
}

.title-search {
  width: min(100%, 340px);
  position: relative;
}

.title-search i {
  position: absolute;
  left: 0.85rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--p-text-muted-color, #64748b);
}

.title-search input {
  width: 100%;
  height: 42px;
  padding: 0 0.85rem 0 2.45rem;
  border: 1px solid var(--p-content-border-color, #cbd5e1);
  border-radius: 8px;
  outline: none;
  background: var(--p-content-background, #ffffff);
  color: var(--p-text-color, #0f172a);
  font-size: 0.92rem;
  font-weight: 600;
}

.title-search input:focus {
  border-color: #0891b2;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.14);
}

.announcement-list {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.announcement-card {
  display: grid;
  grid-template-columns: 112px 1fr;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  border-radius: 8px;
  background: var(--p-content-background, #ffffff);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
}

.date-badge {
  min-height: 96px;
  border-radius: 8px;
  background: #0f172a;
  color: #f8fafc;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.date-badge span {
  font-size: 1.7rem;
  line-height: 1;
  font-weight: 800;
}

.date-badge small {
  margin-top: 0.35rem;
  max-width: 86px;
  color: #cbd5e1;
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.25;
}

.announcement-content {
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.announcement-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.55rem;
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.78rem;
  font-weight: 700;
}

.announcement-meta span {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  color: #0891b2;
}

/* Kategori rozeti — "Sistem duyurusu" etiketinin yanında, tarih en sağda kalsın. */
.announcement-meta .category-badge {
  margin-right: auto;
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

.announcement-content h2 {
  margin: 0;
  color: var(--p-text-color, #0f172a);
  font-size: 1.12rem;
  line-height: 1.35;
  font-weight: 800;
}

.announcement-content p {
  margin: 0.55rem 0 0;
  color: var(--p-text-muted-color, #475569);
  font-size: 0.95rem;
  line-height: 1.65;
  white-space: pre-line;
}

.read-more-btn {
  width: fit-content;
  margin-top: 0.8rem;
  padding: 0.45rem 0.7rem;
  border: 1px solid rgba(6, 182, 212, 0.26);
  border-radius: 8px;
  background: rgba(6, 182, 212, 0.08);
  color: #0891b2;
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.82rem;
  font-weight: 800;
  cursor: pointer;
  transition: background-color 0.18s, border-color 0.18s;
}

.read-more-btn:hover {
  background: rgba(6, 182, 212, 0.14);
  border-color: rgba(6, 182, 212, 0.42);
}

.state-message {
  min-height: 220px;
  border: 1px solid var(--p-content-border-color, #e2e8f0);
  border-radius: 8px;
  background: var(--p-content-background, #ffffff);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  text-align: left;
}

.state-icon {
  width: 52px;
  height: 52px;
  border-radius: 8px;
  background: rgba(6, 182, 212, 0.1);
  color: #0891b2;
  display: grid;
  place-items: center;
  font-size: 1.4rem;
}

.state-message h2 {
  margin: 0;
  color: var(--p-text-color, #0f172a);
  font-size: 1rem;
}

.state-message p,
.state-message span {
  margin: 0.25rem 0 0;
  color: var(--p-text-muted-color, #64748b);
  font-weight: 600;
}

.error-state {
  min-height: auto;
  padding: 1rem;
  justify-content: flex-start;
  color: #b91c1c;
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.22);
}

.skeleton-card {
  box-shadow: none;
}

.skeleton-block,
.skeleton-line {
  position: relative;
  overflow: hidden;
  background: #e2e8f0;
}

.skeleton-block::after,
.skeleton-line::after {
  content: "";
  position: absolute;
  inset: 0;
  transform: translateX(-100%);
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.65), transparent);
  animation: shimmer 1.3s infinite;
}

.skeleton-line {
  height: 12px;
  width: 100%;
  border-radius: 999px;
  margin-top: 0.8rem;
}

.title-line {
  width: 58%;
  height: 18px;
  margin-top: 0;
}

.short-line {
  width: 72%;
}

:root.p-dark .board-header,
:root.p-dark .announcement-card,
:root.p-dark .state-message,
:root.p-dark .search-panel,
:root.p-dark .title-search input {
  background: #1e293b;
  border-color: #334155;
}

:root.p-dark .summary-panel {
  background: rgba(15, 23, 42, 0.72);
}

:root.p-dark .date-badge {
  background: #020617;
}

:root.p-dark .skeleton-block,
:root.p-dark .skeleton-line {
  background: #334155;
}

@keyframes shimmer {
  100% {
    transform: translateX(100%);
  }
}

@media (max-width: 720px) {
  .board-header {
    flex-direction: column;
    padding: 1.1rem;
  }

  .summary-panel {
    min-width: 0;
  }

  .announcement-card {
    grid-template-columns: 1fr;
  }

  .search-panel {
    align-items: stretch;
    flex-direction: column;
  }

  .search-controls {
    flex-direction: column;
    align-items: stretch;
  }

  .category-filter select {
    width: 100%;
  }

  .title-search {
    width: 100%;
  }

  .date-badge {
    min-height: auto;
    padding: 0.8rem;
    align-items: flex-start;
    text-align: left;
  }

  .date-badge small {
    max-width: none;
  }

  .announcement-meta {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.35rem;
  }

  .state-message {
    padding: 1.25rem;
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
