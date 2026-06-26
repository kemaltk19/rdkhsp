// vite.config.ts
import { defineConfig } from "file:///C:/ai-folder/lite-Radikal-hesap/frontend/node_modules/vite/dist/node/index.js";
import vue from "file:///C:/ai-folder/lite-Radikal-hesap/frontend/node_modules/@vitejs/plugin-vue/dist/index.mjs";
import tailwindcss from "file:///C:/ai-folder/lite-Radikal-hesap/frontend/node_modules/@tailwindcss/vite/dist/index.mjs";
import path from "path";
var __vite_injected_original_dirname = "C:\\ai-folder\\lite-Radikal-hesap\\frontend";
var vite_config_default = defineConfig({
  plugins: [
    vue(),
    tailwindcss()
  ],
  resolve: {
    alias: {
      "@": path.resolve(__vite_injected_original_dirname, "./src")
    }
  },
  server: {
    port: 9e3,
    strictPort: true
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes("node_modules")) return;
          if (/[\\/]node_modules[\\/](vue|@vue|vue-router|pinia|@intlify|vue-i18n)[\\/]/.test(id)) {
            return "vendor-vue";
          }
        }
      }
    },
    // Ağır export kütüphaneleri (jspdf, xlsx) dinamik import edildiğinden ana
    // bundle'a girmez; geri kalan parçalar 600 KB altında kalır.
    chunkSizeWarningLimit: 600
  }
});
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZS5jb25maWcudHMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCJDOlxcXFxhaS1mb2xkZXJcXFxcbGl0ZS1SYWRpa2FsLWhlc2FwXFxcXGZyb250ZW5kXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ZpbGVuYW1lID0gXCJDOlxcXFxhaS1mb2xkZXJcXFxcbGl0ZS1SYWRpa2FsLWhlc2FwXFxcXGZyb250ZW5kXFxcXHZpdGUuY29uZmlnLnRzXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ltcG9ydF9tZXRhX3VybCA9IFwiZmlsZTovLy9DOi9haS1mb2xkZXIvbGl0ZS1SYWRpa2FsLWhlc2FwL2Zyb250ZW5kL3ZpdGUuY29uZmlnLnRzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSAndml0ZSdcbmltcG9ydCB2dWUgZnJvbSAnQHZpdGVqcy9wbHVnaW4tdnVlJ1xuaW1wb3J0IHRhaWx3aW5kY3NzIGZyb20gJ0B0YWlsd2luZGNzcy92aXRlJ1xuaW1wb3J0IHBhdGggZnJvbSAncGF0aCdcblxuLy8gaHR0cHM6Ly92aXRlLmRldi9jb25maWcvXG5leHBvcnQgZGVmYXVsdCBkZWZpbmVDb25maWcoe1xuICBwbHVnaW5zOiBbXG4gICAgdnVlKCksXG4gICAgdGFpbHdpbmRjc3MoKSxcbiAgXSxcbiAgcmVzb2x2ZToge1xuICAgIGFsaWFzOiB7XG4gICAgICAnQCc6IHBhdGgucmVzb2x2ZShfX2Rpcm5hbWUsICcuL3NyYycpLFxuICAgIH0sXG4gIH0sXG4gIHNlcnZlcjoge1xuICAgIHBvcnQ6IDkwMDAsXG4gICAgc3RyaWN0UG9ydDogdHJ1ZSxcbiAgfSxcbiAgYnVpbGQ6IHtcbiAgICByb2xsdXBPcHRpb25zOiB7XG4gICAgICBvdXRwdXQ6IHtcbiAgICAgICAgbWFudWFsQ2h1bmtzKGlkKSB7XG4gICAgICAgICAgaWYgKCFpZC5pbmNsdWRlcygnbm9kZV9tb2R1bGVzJykpIHJldHVyblxuICAgICAgICAgIC8vIFx1MDBDN2VraXJkZWsgVnVlIGVrb3Npc3RlbWkgXHUyMDE0IGhlciBzYXlmYWRhIGdlcmVrbGksIHRlayBwYXJcdTAwRTdhZGEga2Fsc1x1MDEzMW4uXG4gICAgICAgICAgaWYgKC9bXFxcXC9dbm9kZV9tb2R1bGVzW1xcXFwvXSh2dWV8QHZ1ZXx2dWUtcm91dGVyfHBpbmlhfEBpbnRsaWZ5fHZ1ZS1pMThuKVtcXFxcL10vLnRlc3QoaWQpKSB7XG4gICAgICAgICAgICByZXR1cm4gJ3ZlbmRvci12dWUnXG4gICAgICAgICAgfVxuICAgICAgICAgIC8vIFByaW1lVnVlIHZlIGRpXHUwMTFGZXIgXHUwMEZDXHUwMEU3XHUwMEZDbmNcdTAwRkMgcGFydGkga29kdSBWaXRlJ1x1MDEzMW4gb3RvbWF0aWsgYlx1MDBGNmxtZXNpbmVcbiAgICAgICAgICAvLyBiXHUwMTMxcmFrXHUwMTMxbFx1MDEzMXI7IGJcdTAwRjZ5bGVjZSBiaWxlXHUwMTVGZW5sZXIgaWxnaWxpIHNheWZhIHBhclx1MDBFN2FsYXJcdTAxMzFuYSBkYVx1MDExRlx1MDEzMWxcdTAxMzFyIHZlXG4gICAgICAgICAgLy8gdGVrIGRldiBcInByaW1ldnVlXCIgcGFyXHUwMEU3YXNcdTAxMzEgb2x1XHUwMTVGbWF6LlxuICAgICAgICB9LFxuICAgICAgfSxcbiAgICB9LFxuICAgIC8vIEFcdTAxMUZcdTAxMzFyIGV4cG9ydCBrXHUwMEZDdFx1MDBGQ3BoYW5lbGVyaSAoanNwZGYsIHhsc3gpIGRpbmFtaWsgaW1wb3J0IGVkaWxkaVx1MDExRmluZGVuIGFuYVxuICAgIC8vIGJ1bmRsZSdhIGdpcm1lejsgZ2VyaSBrYWxhbiBwYXJcdTAwRTdhbGFyIDYwMCBLQiBhbHRcdTAxMzFuZGEga2FsXHUwMTMxci5cbiAgICBjaHVua1NpemVXYXJuaW5nTGltaXQ6IDYwMCxcbiAgfSxcbn0pXG4iXSwKICAibWFwcGluZ3MiOiAiO0FBQWtULFNBQVMsb0JBQW9CO0FBQy9VLE9BQU8sU0FBUztBQUNoQixPQUFPLGlCQUFpQjtBQUN4QixPQUFPLFVBQVU7QUFIakIsSUFBTSxtQ0FBbUM7QUFNekMsSUFBTyxzQkFBUSxhQUFhO0FBQUEsRUFDMUIsU0FBUztBQUFBLElBQ1AsSUFBSTtBQUFBLElBQ0osWUFBWTtBQUFBLEVBQ2Q7QUFBQSxFQUNBLFNBQVM7QUFBQSxJQUNQLE9BQU87QUFBQSxNQUNMLEtBQUssS0FBSyxRQUFRLGtDQUFXLE9BQU87QUFBQSxJQUN0QztBQUFBLEVBQ0Y7QUFBQSxFQUNBLFFBQVE7QUFBQSxJQUNOLE1BQU07QUFBQSxJQUNOLFlBQVk7QUFBQSxFQUNkO0FBQUEsRUFDQSxPQUFPO0FBQUEsSUFDTCxlQUFlO0FBQUEsTUFDYixRQUFRO0FBQUEsUUFDTixhQUFhLElBQUk7QUFDZixjQUFJLENBQUMsR0FBRyxTQUFTLGNBQWMsRUFBRztBQUVsQyxjQUFJLDJFQUEyRSxLQUFLLEVBQUUsR0FBRztBQUN2RixtQkFBTztBQUFBLFVBQ1Q7QUFBQSxRQUlGO0FBQUEsTUFDRjtBQUFBLElBQ0Y7QUFBQTtBQUFBO0FBQUEsSUFHQSx1QkFBdUI7QUFBQSxFQUN6QjtBQUNGLENBQUM7IiwKICAibmFtZXMiOiBbXQp9Cg==
