import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  build: {
    outDir: '../backend/web/dist',
    emptyOutDir: true,
  },
  server: {
    port: 27138,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:27138',
        changeOrigin: true,
      },
    },
  },
})
