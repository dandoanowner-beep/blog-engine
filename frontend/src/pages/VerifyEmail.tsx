import { useEffect, useState } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { authApi } from '../api/auth'

export default function VerifyEmail() {
  const [params] = useSearchParams()
  const [status, setStatus] = useState<'loading' | 'ok' | 'error'>('loading')

  useEffect(() => {
    const token = params.get('token')
    if (!token) { setStatus('error'); return }
    authApi.verifyEmail(token)
      .then(() => setStatus('ok'))
      .catch(() => setStatus('error'))
  }, [params])

  if (status === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Verifying your email…</p>
      </div>
    )
  }

  if (status === 'ok') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100 max-w-md text-center">
          <div className="text-4xl mb-4">✅</div>
          <h1 className="text-xl font-bold mb-2">Email verified!</h1>
          <p className="text-gray-500 text-sm mb-4">Your account is now active.</p>
          <Link to="/auth/login" className="bg-blue-600 text-white px-6 py-2 rounded font-medium hover:bg-blue-700">
            Sign in
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100 max-w-md text-center">
        <div className="text-4xl mb-4">❌</div>
        <h1 className="text-xl font-bold mb-2">Verification failed</h1>
        <p className="text-gray-500 text-sm mb-4">The link may have expired. Request a new one.</p>
        <Link to="/auth/login" className="text-blue-600 hover:underline text-sm">Back to sign in</Link>
      </div>
    </div>
  )
}
