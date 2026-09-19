import { defineStore } from 'pinia'
import { getMe, login as loginRequest } from '@/api/users'
import type { UserDTO } from '@/types/user'

function loadStoredUser(): UserDTO | null {
  const raw = localStorage.getItem('auth_user')
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserDTO
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('auth_token'),
    user: loadStoredUser(),
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    isAdmin: (state) => state.user?.role === 'admin',
  },
  actions: {
    async login(email: string, password: string) {
      const { data } = await loginRequest({ email, password })
      this.token = data.access_token
      localStorage.setItem('auth_token', this.token)

      const me = await getMe()
      this.user = me.data
      localStorage.setItem('auth_user', JSON.stringify(this.user))
    },
    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_user')
    },
  },
})
