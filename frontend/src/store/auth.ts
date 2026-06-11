import { create } from 'zustand'
import type { User } from '../types'
import { authApi } from '../api/auth'
import { tokenStore } from '../api/client'

// BUG-009: the user object (not the token) is persisted so the UI survives a
// page reload. The access token stays in memory; the axios interceptor
// re-acquires it via POST /auth/refresh (httpOnly cookie) on the first 401.
const USER_KEY = 'blog_engine_user'

export function loadPersistedUser(): User | null {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) ?? 'null')
  } catch {
    return null
  }
}

function persistUser(user: User | null) {
  if (user) localStorage.setItem(USER_KEY, JSON.stringify(user))
  else localStorage.removeItem(USER_KEY)
}

interface AuthState {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setUser: (user: User | null) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: loadPersistedUser(),
  loading: false,

  login: async (email, password) => {
    set({ loading: true })
    try {
      const data = await authApi.login(email, password)
      persistUser(data.user)
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
    persistUser(null)
    set({ user: null })
  },

  setUser: (user) => {
    persistUser(user)
    set({ user })
  },
}))
