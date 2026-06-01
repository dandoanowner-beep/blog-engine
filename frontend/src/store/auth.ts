import { create } from 'zustand'
import type { User } from '../types'
import { authApi } from '../api/auth'
import { tokenStore } from '../api/client'

interface AuthState {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setUser: (user: User | null) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  loading: false,

  login: async (email, password) => {
    set({ loading: true })
    try {
      const data = await authApi.login(email, password)
      set({ user: data.user, loading: false })
    } catch (err) {
      set({ loading: false })
      throw err
    }
  },

  register: async (email, username, password) => {
    set({ loading: true })
    try {
      await authApi.register(email, username, password)
      set({ loading: false })
    } catch (err) {
      set({ loading: false })
      throw err
    }
  },

  logout: async () => {
    await authApi.logout()
    tokenStore.clear()
    set({ user: null })
  },

  setUser: (user) => set({ user }),
}))
