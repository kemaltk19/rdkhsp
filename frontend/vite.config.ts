import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 9000,
    strictPort: true,
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return
          // Çekirdek Vue ekosistemi — her sayfada gerekli, tek parçada kalsın.
          if (/[\\/]node_modules[\\/](vue|@vue|vue-router|pinia|@intlify|vue-i18n)[\\/]/.test(id)) {
            return 'vendor-vue'
          }
          // PrimeVue ve diğer üçüncü parti kodu Vite'ın otomatik bölmesine
          // bırakılır; böylece bileşenler ilgili sayfa parçalarına dağılır ve
          // tek dev "primevue" parçası oluşmaz.
        },
      },
    },
    // Ağır export kütüphaneleri (jspdf, xlsx) dinamik import edildiğinden ana
    // bundle'a girmez; geri kalan parçalar 600 KB altında kalır.
    chunkSizeWarningLimit: 600,
  },
})
