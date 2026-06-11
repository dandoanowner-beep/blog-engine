import type { ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../store/auth'
import NotificationBell from './NotificationBell'
import LanguageToggle from './LanguageToggle'

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuthStore()
  const { t } = useTranslation()
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
            {/* page links: 40px between each link, and 40px before the language toggle
                (nav gap-4 = 16px, so mr-[24px] tops it up to 40px) */}
            <div className="flex items-center gap-[40px] mr-[24px]">
              <Link to="/" className="text-sm font-bold text-gray-600 hover:text-blue-600">{t('nav.articles')}</Link>
              <Link to="/portfolio" className="text-sm font-bold text-gray-600 hover:text-blue-600">{t('nav.portfolio')}</Link>
              <Link to="/author" className="text-sm font-bold text-gray-600 hover:text-blue-600">{t('nav.author')}</Link>
              <Link to="/categories" className="text-sm font-bold text-gray-600 hover:text-blue-600">{t('nav.categories')}</Link>
              <Link to="/forum" className="text-sm font-bold text-gray-600 hover:text-blue-600">{t('nav.forums')}</Link>
            </div>

            <LanguageToggle />

            {user ? (
              <>
                {user.role === 'owner' && (
                  <Link to="/editor" className="text-sm text-gray-600 hover:text-blue-600">
                    {t('nav.write')}
                  </Link>
                )}
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
                  <Link to="/admin" className="text-xs text-gray-500 hover:text-blue-600">{t('nav.admin')}</Link>
                )}
                <button onClick={handleLogout} className="text-sm text-gray-500 hover:text-red-500">
                  {t('nav.logout')}
                </button>
              </>
            ) : (
              <>
                <Link to="/auth/login" className="text-sm text-gray-600 hover:text-blue-600">{t('nav.signIn')}</Link>
                <Link to="/auth/register" className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700">
                  {t('nav.getStarted')}
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
