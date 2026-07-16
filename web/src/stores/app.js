import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
  state: () => ({
    theme: localStorage.getItem('theme') || 'light',
    sidebarCollapsed: false
  }),
  actions: {
    toggleTheme() {
      this.theme = this.theme === 'dark' ? 'light' : 'dark'
      localStorage.setItem('theme', this.theme)
      document.documentElement.dataset.theme = this.theme
    },
    initTheme() {
      document.documentElement.dataset.theme = this.theme
    }
  }
})
