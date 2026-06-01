import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'

interface Props {
  children: ReactNode
  roles?: string[]
}

export default function PrivateRoute({ children, roles }: Props) {
  const { user } = useAuthStore()

  if (!user) return <Navigate to="/auth/login" replace />
  if (roles && !roles.includes(user.role)) return <Navigate to="/" replace />

  return <>{children}</>
}
