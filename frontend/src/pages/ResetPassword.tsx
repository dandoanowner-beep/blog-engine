import { useState, type FormEvent } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { authApi } from '../api/auth'

export default function ResetPassword() {
  const [params] = useSearchParams()
  const token = params.get('token')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleForgot = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await authApi.forgotPassword(email)
      setDone(true)
    } catch {
      setError('Could not send reset email. Check the address and try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleReset = async (e: FormEvent) => {
    e.preventDefault()
    if (!token) return
    setError('')
    setLoading(true)
    try {
      await authApi.resetPassword(token, password)
      setDone(true)
    } catch {
      setError('Reset link is invalid or expired.')
    } finally {
      setLoading(false)
    }
  }

  if (done) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100 max-w-md text-center">
          <div className="text-4xl mb-4">{token ? '🔐' : '📧'}</div>
          <h1 className="text-xl font-bold mb-2">{token ? 'Password updated!' : 'Check your email'}</h1>
          <p className="text-gray-500 text-sm mb-4">
            {token ? 'You can now sign in with your new password.' : `We sent a reset link to ${email}.`}
          </p>
          <Link to="/auth/login" className="text-blue-600 hover:underline text-sm">Back to sign in</Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100 w-full max-w-md">
        <h1 className="text-2xl font-bold mb-6">{token ? 'Set new password' : 'Forgot password'}</h1>

        {error && (
          <div className="bg-red-50 text-red-600 text-sm px-3 py-2 rounded mb-4" role="alert">{error}</div>
        )}

        {token ? (
          <form onSubmit={handleReset} className="space-y-4">
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">New password</label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={8}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <button type="submit" disabled={loading} className="w-full bg-blue-600 text-white py-2 rounded font-medium hover:bg-blue-700 disabled:opacity-60">
              {loading ? 'Updating…' : 'Update password'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleForgot} className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">Email</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <button type="submit" disabled={loading} className="w-full bg-blue-600 text-white py-2 rounded font-medium hover:bg-blue-700 disabled:opacity-60">
              {loading ? 'Sending…' : 'Send reset link'}
            </button>
          </form>
        )}

        <p className="text-center text-sm text-gray-500 mt-4">
          <Link to="/auth/login" className="text-blue-600 hover:underline">Back to sign in</Link>
        </p>
      </div>
    </div>
  )
}
