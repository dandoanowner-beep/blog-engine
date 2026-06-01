import api, { tokenStore } from './client'
import type { User } from '../types'

export interface LoginResponse { access_token: string; user: User }
export interface RegisterResponse { user_id: string; message: string }

export const authApi = {
  register: (email: string, username: string, password: string) =>
    api.post<RegisterResponse>('/auth/register', { email, username, password }),

  login: async (email: string, password: string) => {
    const res = await api.post<LoginResponse>('/auth/login', { email, password })
    tokenStore.set(res.data.access_token)
    return res.data
  },

  logout: async () => {
    await api.post('/auth/logout')
    tokenStore.clear()
  },

  verifyEmail: (token: string) =>
    api.get(`/auth/verify?token=${token}`),

  forgotPassword: (email: string) =>
    api.post('/auth/forgot-password', { email }),

  resetPassword: (token: string, password: string) =>
    api.post('/auth/reset-password', { token, password }),

  googleLogin: () => {
    window.location.href = '/api/v1/auth/google'
  },
}
