import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true, // sends httpOnly refresh cookie
  headers: { 'Content-Type': 'application/json' },
})

// Attach access token from memory on every request
api.interceptors.request.use((config) => {
  const token = tokenStore.get()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Auto-refresh on 401
api.interceptors.response.use(
  (res) => res,
  async (err) => {
    if (err.response?.status === 401 && !err.config._retry) {
      err.config._retry = true
      try {
        const { data } = await axios.post<{ access_token: string }>(
          '/api/v1/auth/refresh',
          {},
          { withCredentials: true }
        )
        tokenStore.set(data.access_token)
        err.config.headers.Authorization = `Bearer ${data.access_token}`
        return api(err.config)
      } catch {
        tokenStore.clear()
        localStorage.removeItem('blog_engine_user') // BUG-009: stale persisted session
        window.location.href = '/auth/login'
      }
    }
    return Promise.reject(err)
  }
)

// In-memory token store (XSS-safe — no localStorage)
export const tokenStore = {
  _token: null as string | null,
  get: () => tokenStore._token,
  set: (t: string) => { tokenStore._token = t },
  clear: () => { tokenStore._token = null },
}

export default api
