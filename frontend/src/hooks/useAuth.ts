import { useState, useEffect, useCallback } from 'react'
import { userApi } from '@/api/user'

interface User {
  id: number
  username: string
  is_admin: boolean
}

export function useAuth() {
  const [user, setUser] = useState<User | null>(() => {
    const stored = localStorage.getItem('user')
    return stored ? JSON.parse(stored) : null
  })
  const [loading, setLoading] = useState(false)

  const login = useCallback((token: string, userData: User) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(userData))
    setUser(userData)
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setUser(null)
    window.location.href = '/login'
  }, [])

  const fetchProfile = useCallback(async () => {
    setLoading(true)
    try {
      const res = await userApi.profile()
      const u = { id: res.data.id, username: res.data.username, is_admin: res.data.is_admin }
      localStorage.setItem('user', JSON.stringify(u))
      setUser(u)
    } catch {
      logout()
    } finally {
      setLoading(false)
    }
  }, [logout])

  return { user, login, logout, fetchProfile, loading }
}
