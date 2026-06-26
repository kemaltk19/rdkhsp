<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const isRegister = computed(() => route.path.startsWith('/register'))

const features = [
  { icon: 'pi pi-users', title: 'Cari Yönetimi', desc: 'Müşteri ve tedarikçi hesaplarınızı tek yerden yönetin.' },
  { icon: 'pi pi-file', title: 'Fatura & Teklif', desc: 'Satış/alış faturaları ve teklifleri saniyeler içinde oluşturun.' },
  { icon: 'pi pi-wallet', title: 'Kasa & Banka', desc: 'Nakit akışınızı ve banka hesaplarınızı anlık takip edin.' },
  { icon: 'pi pi-chart-line', title: 'Raporlar', desc: 'Gelir, gider ve kâr/zarar tablolarını canlı görün.' },
]
</script>

<template>
  <div class="auth-layout" :class="{ 'reg-mode': isRegister }">
    <!-- SOL: Tanıtım (%70) -->
    <section class="auth-hero">
      <div class="hero-inner">
        <div class="brand">
          <span class="logo-icon">▲</span>
          <span class="brand-text">Radikal Hesap</span>
        </div>

        <h1 class="hero-title">
          İşletmenizin finansal kontrolü<br />
          <span class="accent">tek panelde.</span>
        </h1>
        <p class="hero-sub">
          Ön muhasebe, fatura, stok, kasa ve raporlama. Bulut tabanlı,
          her cihazdan erişilebilir, modern ve hızlı.
        </p>

        <ul class="feature-list">
          <li v-for="f in features" :key="f.title" class="feature-item">
            <span class="feature-icon"><i :class="f.icon"></i></span>
            <div class="feature-text">
              <strong>{{ f.title }}</strong>
              <span>{{ f.desc }}</span>
            </div>
          </li>
        </ul>

        <div class="hero-footer">
          <span class="badge"><i class="pi pi-check-circle"></i> 14 gün ücretsiz deneme</span>
          <span class="badge"><i class="pi pi-lock"></i> KVKK uyumlu &amp; güvenli</span>
        </div>
      </div>
    </section>

    <!-- SAĞ: Form (%30) -->
    <section class="auth-form-pane">
      <div class="form-scroll">
        <div class="brand brand-mobile">
          <span class="logo-icon">▲</span>
          <span class="brand-text">Radikal Hesap</span>
        </div>
        <router-view />
      </div>
    </section>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: flex;
  background-color: var(--p-content-background, #f8fafc);
}

/* SOL HERO %70 */
.auth-hero {
  flex: 0 0 70%;
  max-width: 70%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: #f1f5f9;
  background: linear-gradient(135deg, #0f172a 0%, #134e57 55%, #0e7490 100%);
  overflow: hidden;
}

.auth-hero::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(circle at 80% 15%, rgba(6, 182, 212, 0.25) 0%, transparent 45%),
    radial-gradient(circle at 15% 85%, rgba(34, 211, 238, 0.18) 0%, transparent 40%);
  pointer-events: none;
}

.hero-inner {
  position: relative;
  max-width: 560px;
  width: 100%;
}

.hero-title {
  font-size: 2.6rem;
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: -0.03em;
  margin: 1.5rem 0 1rem;
}

.accent {
  color: #22d3ee;
}

.hero-sub {
  font-size: 1.05rem;
  color: #cbd5e1;
  line-height: 1.6;
  margin-bottom: 2.25rem;
  max-width: 480px;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0 0 2.25rem;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.1rem 1.5rem;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
}

.feature-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(34, 211, 238, 0.12);
  border: 1px solid rgba(34, 211, 238, 0.25);
  color: #22d3ee;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.05rem;
}

.feature-text {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.feature-text strong {
  font-size: 0.95rem;
  font-weight: 700;
  color: #f8fafc;
}

.feature-text span {
  font-size: 0.82rem;
  color: #94a3b8;
  line-height: 1.4;
}

.hero-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: #e2e8f0;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0.4rem 0.8rem;
  border-radius: 999px;
}

.badge .pi {
  color: #22d3ee;
}

/* SAĞ FORM %30 */
.auth-form-pane {
  flex: 0 0 30%;
  max-width: 30%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem 2.25rem;
  background: var(--p-content-background, #ffffff);
  overflow-y: auto;
}

/* Kayıt modunda sağ panel genişler, hero daralır (reg-mode layout class'ıyla) */
.reg-mode .auth-form-pane {
  flex-basis: 38%;
  max-width: 38%;
}
.reg-mode .auth-hero {
  flex-basis: 62%;
  max-width: 62%;
}

.form-scroll {
  width: 100%;
  max-width: 420px;
}
.reg-mode .form-scroll {
  max-width: 480px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.brand-mobile {
  display: none;
  justify-content: center;
  margin-bottom: 1.75rem;
}

.logo-icon {
  font-size: 1.75rem;
  color: #06b6d4;
}

.brand-text {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.025em;
}

:root.p-dark .auth-form-pane {
  background: #1e293b;
}

:root.p-dark .brand-text {
  color: #f8fafc;
}

/* RESPONSIVE: tablet ve altı → tek kolon, hero gizlenir */
@media (max-width: 1024px) {
  .auth-layout {
    flex-direction: column;
  }
  .auth-hero {
    display: none;
  }
  .auth-form-pane {
    flex: 1 1 auto;
    max-width: 100%;
    padding: 2rem 1.25rem;
  }
  .reg-mode .auth-form-pane {
    flex: 1 1 auto;
    max-width: 100%;
  }
  .brand-mobile {
    display: flex;
  }
  .form-scroll {
    max-width: 520px;
    margin: 0 auto;
  }
}
</style>
