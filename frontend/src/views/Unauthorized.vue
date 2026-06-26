<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'

const router = useRouter()

const goHome = () => {
  router.push('/dashboard')
}
</script>

<template>
  <div class="unauthorized-container flex flex-col items-center justify-center min-h-[60vh] px-4 text-center">
    <div class="icon-box w-24 h-24 bg-red-50 dark:bg-red-950/20 rounded-full flex items-center justify-center text-red-600 dark:text-red-500 mb-6 border border-red-200 dark:border-red-900/50 shadow-sm">
      <i class="pi pi-shield text-5xl"></i>
    </div>
    
    <h1 class="text-3xl font-extrabold text-red-600 dark:text-red-500 mb-3 tracking-tight">
      Bu İşlem İçin Yetkiniz Bulunmamaktadır!
    </h1>
    
    <p class="text-lg text-slate-700 dark:text-slate-300 font-medium max-w-lg mb-2">
      Girmek istediğiniz sayfayı/modülü görüntüleme iznine sahip değilsiniz.
    </p>
    
    <p class="text-sm text-slate-500 dark:text-slate-400 max-w-md mb-8">
      Yetki değişiklikleri veya daha fazla bilgi için lütfen <strong>şirket yöneticinize başvurun</strong>.
    </p>
    
    <Button label="Kontrol Paneline Dön" icon="pi pi-home" @click="goHome" class="p-button-danger px-6 py-3 text-base font-semibold" outlined />
  </div>
</template>

<style scoped>
.unauthorized-container {
  animation: fadeIn 0.4s ease-out;
}
.icon-box i {
  animation: pulse 2s infinite ease-in-out;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.08); }
}
</style>
