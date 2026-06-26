<template>
  <div class="p-inputgroup flex-1 flex flex-row items-stretch w-full">
    <Select
      v-model="selectedCode"
      :options="countries"
      optionLabel="label"
      optionValue="code"
      class="w-24 phone-code-select"
      filter
      :disabled="disabled"
      placeholder="Ülke"
    >
      <template #value="slotProps">
        <div v-if="slotProps.value" class="flex items-center">
          <span class="mr-1 text-sm">{{ getFlag(slotProps.value) }}</span>
          <span class="text-sm font-medium">{{ slotProps.value }}</span>
        </div>
        <span v-else class="text-sm">
          {{ slotProps.placeholder }}
        </span>
      </template>
      <template #option="slotProps">
        <div class="flex items-center text-sm">
          <span class="mr-2">{{ getFlag(slotProps.option.code) }}</span>
          <span>{{ slotProps.option.name }} ({{ slotProps.option.code }})</span>
        </div>
      </template>
    </Select>
    <InputText
      v-model="phoneNumber"
      :placeholder="placeholder"
      :id="id"
      :disabled="disabled"
      :maxlength="maxlength"
      class="flex-1 min-w-0"
      @input="emitUpdate"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  id?: string
  disabled?: boolean
  maxlength?: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

import { phoneCodes } from '@/constants/phoneCodes'

const settingsStore = useSettingsStore()

const countries = phoneCodes

// Add label for dropdown filtering
countries.forEach(c => (c as any).label = `${c.name} (${c.code})`)

const selectedCode = ref('+90')
const phoneNumber = ref('')

const getFlag = (code: string) => {
  const country = countries.find(c => c.code === code)
  return country ? country.flag : '🌐'
}

const parseModelValue = (val: string) => {
  if (!val) {
    phoneNumber.value = ''
    return
  }
  
  // Sort by code length descending to match +358 before +35
  const sortedCodes = [...countries].sort((a, b) => b.code.length - a.code.length)
  
  let found = false
  for (const c of sortedCodes) {
    if (val.startsWith(c.code)) {
      selectedCode.value = c.code
      phoneNumber.value = val.substring(c.code.length).trim()
      found = true
      break
    }
  }

  // If no prefix matched, maybe they just typed the number
  if (!found) {
    if (val.startsWith('+')) {
      // It's an unknown international format
      selectedCode.value = '+90' // default fallback
      phoneNumber.value = val
    } else {
      // It's a local number
      phoneNumber.value = val
    }
  }
}

const emitUpdate = () => {
  if (!phoneNumber.value) {
    emit('update:modelValue', '')
  } else {
    // Trim any spaces and ensure it has the country code
    emit('update:modelValue', `${selectedCode.value}${phoneNumber.value.trim()}`)
  }
}

watch(selectedCode, () => {
  emitUpdate()
})

watch(() => props.modelValue, (newVal) => {
  // Prevent re-parsing if the value was just emitted by us
  const combined = phoneNumber.value ? `${selectedCode.value}${phoneNumber.value.trim()}` : ''
  if (newVal !== combined) {
    parseModelValue(newVal)
  }
})

onMounted(async () => {
  if (props.modelValue) {
    parseModelValue(props.modelValue)
  } else {
    // Try to get default from company
    if (!settingsStore.company) {
      await settingsStore.fetchCompanyProfile()
    }
    if (settingsStore.company?.country) {
      const companyCountry = countries.find(c => c.name.toLowerCase() === settingsStore.company.country.toLowerCase())
      if (companyCountry) {
        selectedCode.value = companyCountry.code
      }
    }
  }
})
</script>

<style scoped>
:deep(.phone-code-select) {
  min-width: 85px;
}
:deep(.phone-code-select .p-select-label) {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}
</style>
