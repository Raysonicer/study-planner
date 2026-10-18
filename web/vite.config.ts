import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { VitePWA } from 'vite-plugin-pwa'

// 纯前端版：产物可部署到任意子路径（如 GitHub Pages 的 /study-planner/），
// 路由已切 hash 模式，base 用相对路径即可。
export default defineConfig({
  base: './',
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      manifest: {
        name: 'StudyPilot · 学习任务计划',
        short_name: 'StudyPilot',
        description: '目标 · 任务 · 日程 · 统计 · 复盘',
        theme_color: '#4c6ef5',
        background_color: '#e9effd',
        display: 'standalone',
        start_url: './',
        scope: './',
        icons: [{ src: './icon.svg', sizes: 'any', type: 'image/svg+xml', purpose: 'any' }],
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
