<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'

const authStore = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const identifier = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

const handleLogin = async () => {
  if (!identifier.value || !password.value) {
    errorMsg.value = 'Lütfen tüm alanları doldurun.'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    // Backend "email" alanını hem e-posta hem telefon olarak kabul eder
    await authStore.login({
      email: identifier.value.trim(),
      password: password.value,
    })
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
  <div class="login-view">
    <h2 class="title">{{ t('auth.login') }}</h2>
    <p class="subtitle">{{ t('auth.loginText') }}</p>

    <Message v-if="errorMsg" severity="error" class="mb-4">{{ errorMsg }}</Message>

    <form @submit.prevent="handleLogin" class="form">
      <div class="field">
        <label for="identifier">Mail Adresi veya Telefon</label>
        <InputText
          id="identifier"
          v-model="identifier"
          class="w-full"
          placeholder="isim@firma.com veya 5XX XXX XX XX"
          :disabled="loading"
        />
      </div>

      <div class="field">
        <label for="password">{{ t('auth.password') }}</label>
        <Password
          id="password"
          v-model="password"
          class="w-full"
          :feedback="false"
          toggle-mask
          :disabled="loading"
        />
      </div>

      <Button
        type="submit"
        :label="loading ? t('auth.submitting') : t('auth.login')"
        icon="pi pi-sign-in"
        class="w-full submit-btn"
        :loading="loading"
      />
    </form>

    <div class="footer">
      <span>{{ t('auth.noAccount') }}</span>
      <router-link to="/register" class="link-btn">{{ t('auth.register') }}</router-link>
    </div>
  </div>
</template>

<style scoped>
.login-view {
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
  margin-bottom: 1.5rem;
  text-align: center;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field label {
  font-size: 0.85rem;
  font-weight: 600;
}

.submit-btn {
  margin-top: 0.25rem;
}

.footer {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.6rem;
  margin-top: 1.5rem;
  font-size: 0.9rem;
}

/* Kayıt ol butonu biraz küçük (giriş sayfasında ikincil aksiyon) */
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
</style>
