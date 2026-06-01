import type { ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import NotificationBell from './NotificationBell'

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/auth/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-6xl mx-auto px-4 h-14 flex items-center justify-between">
          <Link to="/" className="font-bold text-lg text-blue-600">BlogEngine</Link>

          <nav className="flex items-center gap-4">
            <Link to="/" className="text-sm text-gray-600 hover:text-blue-600">Explore</Link>
            <Link to="/search" className="text-sm text-gray-600 hover:text-blue-600">Search</Link>

            {user ? (
              <>
                <Link to="/editor" className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700">
                  Write
                </Link>
                <NotificationBell />
                <Link to={`/profile/${user.username}`} className="flex items-center gap-1">
                  {user.avatar_url
                    ? <img src={user.avatar_url} alt={user.username} className="w-7 h-7 rounded-full" />
                    : <span className="w-7 h-7 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-xs font-medium">
                        {user.username[0].toUpperCase()}
                      </span>
                  }
                </Link>
                {(user.role === 'admin' || user.role === 'owner') && (
                  <Link to="/admin" className="text-xs text-gray-500 hover:text-blue-600">Admin</Link>
                )}
                <button onClick={handleLogout} className="text-sm text-gray-500 hover:text-red-500">
                  Logout
                </button>
              </>
            ) : (
              <>
                <Link to="/auth/login" className="text-sm text-gray-600 hover:text-blue-600">Sign in</Link>
                <Link to="/auth/register" className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700">
                  Get started
                </Link>
              </>
            )}
          </nav>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-6">{children}</main>
    </div>
  )
}
