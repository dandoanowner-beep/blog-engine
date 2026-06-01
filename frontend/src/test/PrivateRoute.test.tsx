import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import PrivateRoute from '../components/PrivateRoute'
import { useAuthStore } from '../store/auth'

vi.mock('../store/auth', () => ({ useAuthStore: vi.fn() }))

function wrap(user: object | null, roles?: string[], path = '/protected') {
  vi.mocked(useAuthStore).mockReturnValue({ user } as never)
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/protected" element={
          <PrivateRoute roles={roles}>
            <div data-testid="content">Protected</div>
          </PrivateRoute>
        } />
        <Route path="/auth/login" element={<div data-testid="login-page">Login</div>} />
        <Route path="/" element={<div data-testid="home-page">Home</div>} />
      </Routes>
    </MemoryRouter>
  )
}

describe('PrivateRoute', () => {
  it('renders children when user is authenticated', () => {
    wrap({ id: '1', username: 'alice', role: 'user' })
    expect(screen.getByTestId('content')).toBeInTheDocument()
  })

  it('redirects to /auth/login when user is null', () => {
    wrap(null)
    expect(screen.getByTestId('login-page')).toBeInTheDocument()
    expect(screen.queryByTestId('content')).not.toBeInTheDocument()
  })

  it('renders children when user role matches required roles', () => {
    wrap({ id: '1', username: 'admin', role: 'admin' }, ['admin', 'owner'])
    expect(screen.getByTestId('content')).toBeInTheDocument()
  })

  it('redirects to / when user role does not match required roles', () => {
    wrap({ id: '1', username: 'alice', role: 'user' }, ['admin', 'owner'])
    expect(screen.getByTestId('home-page')).toBeInTheDocument()
    expect(screen.queryByTestId('content')).not.toBeInTheDocument()
  })
})
