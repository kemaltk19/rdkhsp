<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import { countries } from '@/constants/countries'
import { sectors } from '@/constants/sectors'
import { phoneCodesOptions } from '@/constants/phoneCodes'
import { metaForCountry, browserTimezone } from '@/utils/localeMap'

const authStore = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const companyName = ref('')
const fullName = ref('')
const phoneCode = ref('+90')
const phoneNumber = ref('')
const email = ref('')
const industry = ref<string | null>(null)
const country = ref<string | null>('Türkiye')
const city = ref('')
const district = ref('')
const address = ref('')
const password = ref('')
const passwordConfirm = ref('')

// Otomatik ayarlanan (ülkeden) — kullanıcı isterse düzeltebilir
const currency = ref('TRY')
const locale = ref('tr')
const timezone = ref(browserTimezone())

const loading = ref(false)
const errorMsg = ref('')

const countryOptions = countries.map((c) => ({ label: c, value: c }))
const sectorOptions = sectors.map((s) => ({ label: s, value: s }))

const currencyOptions = [
  { label: '₺ Türk Lirası (TRY)', value: 'TRY' },
  { label: '$ Amerikan Doları (USD)', value: 'USD' },
  { label: '€ Euro (EUR)', value: 'EUR' },
  { label: '£ İngiliz Sterlini (GBP)', value: 'GBP' },
  { label: '₽ Rus Rublesi (RUB)', value: 'RUB' },
]
const localeOptions = [
  { label: 'Türkçe', value: 'tr' },
  { label: 'English', value: 'en' },
]

// Ülke değişince para birimi + dil + saat dilimi otomatik gelsin
watch(country, (val) => {
  const meta = metaForCountry(val)
  currency.value = meta.currency
  locale.value = meta.locale
  timezone.value = meta.timezone
  // Telefon kodunu da ülkeye göre dene
  const match = phoneCodesOptions.find((p) => p.name === val)
  if (match) phoneCode.value = match.code
})

onMounted(() => {
  // İlk açılışta TR varsayılanı + tarayıcı tz
  const meta = metaForCountry(country.value)
  currency.value = meta.currency
  locale.value = meta.locale
  timezone.value = browserTimezone() || meta.timezone
})

const fullPhone = computed(() => {
  const num = phoneNumber.value.trim()
  if (!num) return ''
  return `${phoneCode.value}${num.replace(/^0+/, '')}`
})

const handleRegister = async () => {
  errorMsg.value = ''
  if (!companyName.value || !fullName.value || !email.value || !password.value) {
    errorMsg.value = 'Lütfen zorunlu alanları doldurun.'
    return
  }
  if (!phoneNumber.value.trim()) {
    errorMsg.value = 'Telefon numarası zorunludur.'
    return
  }
  if (password.value.length < 8) {
    errorMsg.value = 'Şifre en az 8 karakter olmalıdır.'
    return
  }
  if (password.value !== passwordConfirm.value) {
    errorMsg.value = 'Şifreler eşleşmiyor.'
    return
  }

  loading.value = true
  try {
    await authStore.register({
      company_name: companyName.value,
      name: fullName.value,
      email: email.value,
      password: password.value,
      phone: fullPhone.value,
      industry: industry.value || '',
      country: country.value || '',
      city: city.value,
      district: district.value,
      address: address.value,
      currency: currency.value,
      locale: locale.value,
      timezone: timezone.value,
    })
    // Yeni kullanıcı için karşılama popup'ı dashboard'da gösterilsin
    sessionStorage.setItem('show_welcome', '1')
    router.push('/dashboard')
  } catch (err: any) {
    if (err.response && err.response.data && err.response.data.error) {
      errorMsg.value = err.response.data.error.message
    } else {
      errorMsg.value = 'Sunucuyla bağlantı kurulamadı.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="register-view">
    <h2 class="title">{{ t('auth.register') }}</h2>
    <p class="subtitle">{{ t('auth.registerText') }}</p>

    <Message v-if="errorMsg" severity="error" class="mb-4">{{ errorMsg }}</Message>

    <form @submit.prevent="handleRegister" class="form">
      <div class="grid">
        <div class="field">
          <label for="companyName">{{ t('auth.companyName') }} *</label>
          <InputText id="companyName" v-model="companyName" class="w-full" maxlength="255"
            placeholder="Ör: Radikal Teknoloji" :disabled="loading" />
        </div>

        <div class="field">
          <label for="fullName">{{ t('auth.fullName') }} *</label>
          <InputText id="fullName" v-model="fullName" class="w-full" maxlength="255"
            placeholder="Ör: Ahmet Yılmaz" :disabled="loading" />
        </div>

        <div class="field">
          <label>Telefon Numarası *</label>
          <div class="phone-row">
            <Select v-model="phoneCode" :options="phoneCodesOptions" optionLabel="label"
              optionValue="code" filter class="phone-code" :disabled="loading" />
            <InputText v-model="phoneNumber" class="w-full" placeholder="5XX XXX XX XX"
              maxlength="20" :disabled="loading" />
          </div>
        </div>

        <div class="field">
          <label for="email">Mail Adresiniz *</label>
          <InputText id="email" v-model="email" type="email" class="w-full" maxlength="255"
            placeholder="isim@firma.com" :disabled="loading" />
        </div>

        <div class="field">
          <label>Sektör</label>
          <Select v-model="industry" :options="sectorOptions" optionLabel="label"
            optionValue="value" filter placeholder="Sektör seçin" class="w-full"
            :disabled="loading" showClear />
        </div>

        <div class="field">
          <label>Ülke</label>
          <Select v-model="country" :options="countryOptions" optionLabel="label"
            optionValue="value" filter placeholder="Ülke seçin" class="w-full"
            :disabled="loading" />
        </div>

        <div class="field">
          <label for="city">İl</label>
          <InputText id="city" v-model="city" class="w-full" maxlength="255" placeholder="Ör: Ankara"
            :disabled="loading" />
        </div>

        <div class="field">
          <label for="district">İlçe</label>
          <InputText id="district" v-model="district" class="w-full" maxlength="255" placeholder="Ör: Çankaya"
            :disabled="loading" />
        </div>

        <div class="field field-full">
          <label for="address">Adres</label>
          <Textarea id="address" v-model="address" class="w-full" rows="2"
            placeholder="Açık adres" :disabled="loading" autoResize />
        </div>

        <div class="field">
          <label for="password">Şifre *</label>
          <Password id="password" v-model="password" class="w-full" toggle-mask
            promptLabel="Şifre belirleyin" weakLabel="Zayıf" mediumLabel="Orta"
            strongLabel="Güçlü" :disabled="loading" />
        </div>

        <div class="field">
          <label for="passwordConfirm">Şifre (Tekrar) *</label>
          <Password id="passwordConfirm" v-model="passwordConfirm" class="w-full"
            :feedback="false" toggle-mask :disabled="loading" />
        </div>
      </div>

      <!-- Otomatik bölge ayarları (ülkeden gelir, düzeltilebilir) -->
      <div class="auto-box">
        <span class="auto-label">Bölge Ayarları (otomatik)</span>
        <div class="auto-row">
          <div class="field">
            <label>Para Birimi</label>
            <Select v-model="currency" :options="currencyOptions" optionLabel="label"
              optionValue="value" class="w-full" :disabled="loading" />
          </div>
          <div class="field">
            <label>Dil</label>
            <Select v-model="locale" :options="localeOptions" optionLabel="label"
              optionValue="value" class="w-full" :disabled="loading" />
          </div>
        </div>
        <small class="tz-note">Saat dilimi: {{ timezone }}</small>
      </div>

      <Button type="submit" :label="loading ? t('auth.submitting') : t('auth.register')"
        icon="pi pi-user-plus" class="w-full submit-btn" :loading="loading" />
    </form>

    <div class="footer">
      <span>{{ t('auth.haveAccount') }}</span>
      <router-link to="/login" class="link-btn">{{ t('auth.login') }}</router-link>
    </div>
  </div>
</template>

<style scoped>
.register-view {
  display: flex;
  flex-direction: column;
}

.title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
  text-align: center;
}

.subtitle {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.9rem;
  margin-bottom: 1.25rem;
  text-align: center;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.85rem 1rem;
}

/* Dar sağ panelde 2 sütun sıkışık duruyordu; ~520px altında tek sütun */
@media (max-width: 520px) {
  .grid {
    grid-template-columns: 1fr;
  }
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.field-full {
  grid-column: 1 / -1;
}

.field label {
  font-size: 0.82rem;
  font-weight: 600;
}

.phone-row {
  display: flex;
  gap: 0.5rem;
}

.phone-code {
  width: 130px;
  flex-shrink: 0;
}

.auto-box {
  background: var(--p-surface-50, #f8fafc);
  border: 1px solid var(--p-surface-200, #e2e8f0);
  border-radius: 10px;
  padding: 0.85rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.auto-label {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--p-primary-color, #06b6d4);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.auto-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.tz-note {
  color: var(--p-text-muted-color, #64748b);
  font-size: 0.78rem;
}

.submit-btn {
  margin-top: 0.25rem;
}

.footer {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.6rem;
  margin-top: 1.25rem;
  font-size: 0.9rem;
}

/* Giriş yap butonu biraz küçük (kayıt sayfasında ikincil aksiyon) */
.link-btn {
  display: inline-block;
  padding: 0.35rem 0.9rem;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--p-primary-color, #06b6d4);
  border: 1px solid var(--p-primary-color, #06b6d4);
  border-radius: 8px;
  text-decoration: none;
  transition: all 0.15s;
}

.link-btn:hover {
  background: var(--p-primary-color, #06b6d4);
  color: #fff;
}

:deep(.p-password),
:deep(.p-password-input) {
  width: 100%;
}

@media (max-width: 520px) {
  .auto-row {
    grid-template-columns: 1fr;
  }
}
</style>
