import { defineStore } from 'pinia'
import * as api from '@/api/settings'

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    values: {} as Record<string, string>,
    loaded: false,
  }),
  getters: {
    get: (s) => (key: string, def = '') => s.values[key] ?? def,
    theme: (s) => (s.values.theme === 'dark' ? 'dark' : 'light'),
    autoCompleteParent: (s) => s.values.auto_complete_parent === 'true',
    heatmapTaskWeight: (s) => parseFloat(s.values.heatmap_task_weight ?? '0.5') || 0.5,
    weekStartDay: (s) => {
      const n = parseInt(s.values.week_start_day ?? '1', 10)
      return n >= 0 && n <= 6 ? n : 1
    },
    notifyEnabled: (s) => s.values.notify_overdue === 'true',
    heatmapGoalWeight: (s) => parseFloat(s.values.heatmap_goal_weight ?? '0.5') || 0.5,
  },
  actions: {
    async load() {
      try {
        const res = await api.getSettings()
        this.values = res.data
      } finally {
        this.loaded = true
        this.applyTheme()
      }
    },
    async set(key: string, value: string) {
      this.values[key] = value
      await api.putSettings({ [key]: value })
      if (key === 'theme') this.applyTheme()
    },
    applyTheme() {
      document.documentElement.classList.toggle('dark', this.theme === 'dark')
    },
  },
})