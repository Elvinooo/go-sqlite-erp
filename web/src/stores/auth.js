import { defineStore } from 'pinia'
import { getCurrentUser } from '../api/auth'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: readJSON('currentUser'),
    permissions: readJSON('permissions', []),
    menus: readJSON('menus', [])
  }),
  getters: {
    isLoggedIn: () => Boolean(localStorage.getItem('accessToken')),
    hasPermission: (state) => (code) => {
      if (!code) return true
      if (state.user?.username === 'admin') return true
      const permissions = (state.permissions || []).map((item) => typeof item === 'string' ? item : item?.code).filter(Boolean)
      if (permissions.includes('*')) return true
      return permissions.includes(code)
    }
  },
  actions: {
    async loadCurrentUser() {
      if (!this.isLoggedIn) return
      const data = await getCurrentUser()
      this.user = data
      this.permissions = data.permissions || []
      this.menus = data.menus || []
      localStorage.setItem('currentUser', JSON.stringify(this.user))
      localStorage.setItem('permissions', JSON.stringify(this.permissions))
      localStorage.setItem('menus', JSON.stringify(this.menus))
    },
    clear() {
      this.user = null
      this.permissions = []
      this.menus = []
      localStorage.clear()
      sessionStorage.clear()
    }
  }
})

function readJSON(key, fallback = null) {
  try {
    const value = localStorage.getItem(key)
    return value ? JSON.parse(value) : fallback
  } catch {
    return fallback
  }
}
