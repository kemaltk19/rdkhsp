<template>
  <div class="card max-w-4xl mx-auto mt-4 p-6 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-xl shadow-sm">
    <div class="mb-6 pb-4 border-b border-slate-200 dark:border-slate-700">
      <h2 class="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <i class="pi pi-envelope text-primary-500 text-2xl"></i>
        {{ $t('superadmin.emailSettings.title') }}
      </h2>
      <p class="text-sm text-slate-500 dark:text-slate-400 mt-2">
        {{ $t('superadmin.emailSettings.desc') }}
      </p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
      <div class="field">
        <label class="block mb-2 font-semibold text-sm text-slate-700 dark:text-slate-300">{{ $t('superadmin.emailSettings.host') }}</label>
        <InputText v-model.trim="form.host" placeholder="smtp.gmail.com" class="w-full" maxlength="255" />
      </div>

      <div class="field">
        <label class="block mb-2 font-semibold text-sm text-slate-700 dark:text-slate-300">{{ $t('superadmin.emailSettings.port') }}</label>
        <InputText v-model.trim="form.port" placeholder="587" class="w-full" maxlength="10" />
      </div>

      <div class="field">
        <label class="block mb-2 font-semibold text-sm text-slate-700 dark:text-slate-300">{{ $t('superadmin.emailSettings.username') }}</label>
        <InputText v-model.trim="form.username" placeholder="kullanici@firma.com" class="w-full" maxlength="255" />
      </div>

      <div class="field">
        <label class="block mb-2 font-semibold text-sm text-slate-700 dark:text-slate-300">{{ $t('superadmin.emailSettings.password') }}</label>
        <InputText
          v-model="form.password"
          type="password"
          :placeholder="hasPassword ? $t('superadmin.emailSettings.passwordPlaceholderEdit') : $t('superadmin.emailSettings.passwordPlaceholder')"
          class="w-full"
          maxlength="255"
        />
        <small class="text-slate-500 dark:text-slate-400 mt-1 block" v-if="hasPassword">
          <i class="pi pi-info-circle mr-1 text-xs"></i>{{ $t('superadmin.emailSettings.passwordInfo') }}
        </small>
      </div>

      <div class="field">
        <label class="block mb-2 font-semibold text-sm text-slate-700 dark:text-slate-300">{{ $t('superadmin.emailSettings.fromEmail') }}</label>
        <InputText v-model.trim="form.from_email" placeholder="no-reply@firma.com" class="w-full" maxlength="255" />
      </div>

      <div class="field">
        <label class="block mb-2 font-semibold text-sm text-slate-700 dark:text-slate-300">{{ $t('superadmin.emailSettings.fromName') }}</label>
        <InputText v-model.trim="form.from_name" placeholder="Radikal Hesap" class="w-full" maxlength="255" />
      </div>

      <div class="field col-span-1 md:col-span-2 mt-2 bg-slate-50 dark:bg-slate-800 p-4 rounded-lg border border-slate-200 dark:border-slate-700">
        <div class="flex items-center gap-3">
          <ToggleSwitch inputId="enabled" v-model="form.enabled" />
          <label for="enabled" class="font-semibold text-slate-700 dark:text-slate-300 cursor-pointer select-none">
            {{ $t('superadmin.emailSettings.active') }}
          </label>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-2 ml-14">
          {{ $t('superadmin.emailSettings.activeDesc') }}
        </p>
      </div>
    </div>

    <div class="flex justify-end gap-3 mb-8">
      <Button :label="$t('superadmin.emailSettings.save')" icon="pi pi-check" :loading="saving" @click="save" class="w-full md:w-auto" severity="primary" />
    </div>

    <div class="bg-primary-50 dark:bg-primary-900/20 p-5 rounded-lg border border-primary-200 dark:border-primary-800/30">
      <div class="mb-3">
        <label class="flex font-bold text-primary-900 dark:text-primary-100 items-center gap-2">
          <i class="pi pi-send text-primary-600 dark:text-primary-400"></i> {{ $t('superadmin.emailSettings.testTitle') }}
        </label>
        <p class="text-sm text-primary-700 dark:text-primary-300 mt-1">{{ $t('superadmin.emailSettings.testDesc') }}</p>
      </div>
      <div class="flex flex-col md:flex-row gap-3 items-start md:items-center">
        <InputText v-model.trim="testTo" :placeholder="$t('superadmin.emailSettings.testToPlaceholder')" class="w-full md:flex-1" maxlength="255" />
        <Button :label="$t('superadmin.emailSettings.testBtn')" icon="pi pi-paperclip" severity="secondary" :loading="testing" @click="sendTest" class="w-full md:w-auto" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'
import { getEmailSettingsApi, updateEmailSettingsApi, testEmailApi } from '@/api/superadmin'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import ToggleSwitch from 'primevue/toggleswitch'

const toast = useToast()
const { t } = useI18n()

const form = ref<any>({
  host: '',
  port: '587',
  username: '',
  password: '',
  from_email: '',
  from_name: '',
  enabled: false,
})
const hasPassword = ref(false)
const testTo = ref('')
const saving = ref(false)
const testing = ref(false)

const load = async () => {
  try {
    const res = await getEmailSettingsApi()
    const s = res.data.data
    form.value = {
      host: s.host || '',
      port: s.port || '587',
      username: s.username || '',
      password: '',
      from_email: s.from_email || '',
      from_name: s.from_name || '',
      enabled: !!s.enabled,
    }
    hasPassword.value = !!s.has_password
  } catch (e) {
    toast.add({ severity: 'error', summary: t('superadmin.emailSettings.errorLoad'), detail: t('superadmin.emailSettings.errorLoad'), life: 10000 })
  }
}

const save = async () => {
  saving.value = true
  try {
    await updateEmailSettingsApi(form.value)
    toast.add({ severity: 'success', summary: t('superadmin.emailSettings.successSave'), detail: t('superadmin.emailSettings.successSave'), life: 10000 })
    form.value.password = ''
    await load()
  } catch (e) {
    toast.add({ severity: 'error', summary: t('superadmin.emailSettings.errorSave'), detail: t('superadmin.emailSettings.errorSave'), life: 10000 })
  } finally {
    saving.value = false
  }
}

const sendTest = async () => {
  if (!testTo.value) {
    toast.add({ severity: 'warn', summary: t('superadmin.emailSettings.warnEmail'), detail: t('superadmin.emailSettings.warnEmail'), life: 10000 })
    return
  }
  testing.value = true
  try {
    await testEmailApi(testTo.value)
    toast.add({ severity: 'success', summary: t('superadmin.emailSettings.successTest'), detail: t('superadmin.emailSettings.successTest'), life: 10000 })
  } catch (e: any) {
    const msg = e?.response?.data?.error?.message || t('superadmin.emailSettings.errorTest')
    toast.add({ severity: 'error', summary: t('superadmin.emailSettings.errorTest'), detail: msg, life: 10000 })
  } finally {
    testing.value = false
  }
}

onMounted(load)
</script>
