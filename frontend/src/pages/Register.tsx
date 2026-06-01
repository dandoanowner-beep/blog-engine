import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import { authApi } from '../api/auth'

export default function Register() {
  const [email, setEmail] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')
  const { register, loading } = useAuthStore()

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      await register(email, username, password)
      setDone(true)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setError(msg ?? 'Registration failed — please try again')
    }
  }

  if (done) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100 max-w-md text-center">
          <div className="text-4xl mb-4">📧</div>
          <h1 className="text-xl font-bold mb-2">Check your email</h1>
          <p className="text-gray-500 text-sm">
            We sent a verification link to <strong>{email}</strong>.<br />
            Click it to activate your account.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100 w-full max-w-md">
        <h1 className="text-2xl font-bold mb-6">Create your account</h1>

        {error && (
          <div className="bg-red-50 text-red-600 text-sm px-3 py-2 rounded mb-4" role="alert">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
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

          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">Username</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              minLength={3}
              maxLength={30}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">Password</label>
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

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded font-medium hover:bg-blue-700 disabled:opacity-60"
          >
            {loading ? 'Creating account…' : 'Create account'}
          </button>
        </form>

        <hr className="my-4" />

        <button
          onClick={authApi.googleLogin}
          className="w-full border border-gray-300 py-2 rounded font-medium flex items-center justify-center gap-2 hover:bg-gray-50"
        >
          <img src="/google.svg" alt="" className="w-4 h-4" />
          Continue with Google
        </button>

        <p className="text-center text-sm text-gray-500 mt-4">
          Already have an account?{' '}
          <Link to="/auth/login" className="text-blue-600 hover:underline">Sign in</Link>
        </p>
      </div>
    </div>
  )
}
