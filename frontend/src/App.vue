<script setup lang="ts">
import { onMounted } from 'vue'
import Toast from 'primevue/toast'
import ConfirmDialog from 'primevue/confirmdialog'
import { useTheme } from '@/composables/useTheme'

const { applyTheme } = useTheme()

onMounted(() => {
  applyTheme()
})
</script>

<template>
  <router-view />
  <Toast />
  <ConfirmDialog />
</template>

<style>
/* Reset and Base Styles */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

/* Scrollbar styling */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 9999px;
}

:root.p-dark ::-webkit-scrollbar-thumb {
  background: #475569;
}
</style>
