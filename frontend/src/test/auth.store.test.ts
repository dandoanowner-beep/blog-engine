import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from '../store/auth'
import { authApi } from '../api/auth'

vi.mock('../api/auth', () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    googleLogin: vi.fn(),
    verifyEmail: vi.fn(),
    forgotPassword: vi.fn(),
    resetPassword: vi.fn(),
  },
}))

vi.mock('../api/client', () => ({
  tokenStore: { get: vi.fn(), set: vi.fn(), clear: vi.fn() },
  default: {},
}))

describe('useAuthStore', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: null, loading: false })
    vi.clearAllMocks()
  })

  it('starts with null user and loading=false', () => {
    const { user, loading } = useAuthStore.getState()
    expect(user).toBeNull()
    expect(loading).toBe(false)
  })

  it('login: sets user on success', async () => {
    const mockUser = { id: '1', username: 'alice', role: 'user' as const, verified: true }
    vi.mocked(authApi.login).mockResolvedValue({ access_token: 'tok', user: mockUser })

    await useAuthStore.getState().login('a@b.com', 'pass')

    expect(useAuthStore.getState().user).toEqual(mockUser)
    expect(useAuthStore.getState().loading).toBe(false)
  })

  it('login: loading=false and throws on error', async () => {
    vi.mocked(authApi.login).mockRejectedValue(new Error('invalid'))

    await expect(useAuthStore.getState().login('x@y.com', 'bad')).rejects.toThrow('invalid')
    expect(useAuthStore.getState().loading).toBe(false)
  })

  it('register: calls authApi.register', async () => {
    vi.mocked(authApi.register).mockResolvedValue({ data: { user_id: '1', message: 'ok' }, status: 201, statusText: 'Created', headers: {}, config: {} as never })

    await useAuthStore.getState().register('a@b.com', 'alice', 'password123')

    expect(authApi.register).toHaveBeenCalledWith('a@b.com', 'alice', 'password123')
  })

  it('logout: clears user', async () => {
    useAuthStore.setState({ user: { id: '1', username: 'alice', role: 'user', verified: true } })
    vi.mocked(authApi.logout).mockResolvedValue(undefined)

    await useAuthStore.getState().logout()

    expect(useAuthStore.getState().user).toBeNull()
  })

  it('setUser: updates user directly', () => {
    const u = { id: '2', username: 'bob', role: 'user' as const, verified: false }
    useAuthStore.getState().setUser(u)
    expect(useAuthStore.getState().user).toEqual(u)
  })

  it('register: sets loading=false and rethrows on error', async () => {
    vi.mocked(authApi.register).mockRejectedValue(new Error('email already taken'))
    await expect(useAuthStore.getState().register('a@b.com', 'alice', 'pass')).rejects.toThrow('email already taken')
    expect(useAuthStore.getState().loading).toBe(false)
  })
})
