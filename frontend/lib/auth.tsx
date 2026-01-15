'use client'

import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { API_URL } from './config'

interface Employee {
  id: string
  name: string
  email: string
  position: string
  department: string | null
  manager_id: string | null
}

interface AuthContextType {
  user: Employee | null
  token: string | null
  isLoading: boolean
  login: (username: string, password: string) => Promise<{ success: boolean; error?: string }>
  logout: () => void
  subordinates: Employee[]
  canAccessEmployee: (employeeId: string) => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<Employee | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [subordinates, setSubordinates] = useState<Employee[]>([])
  const [isLoading, setIsLoading] = useState(true)

  // Check for existing session on mount
  useEffect(() => {
    const savedToken = localStorage.getItem('auth_token')
    const savedUser = localStorage.getItem('auth_user')

    if (savedToken && savedUser) {
      try {
        const userData = JSON.parse(savedUser)
        setToken(savedToken)
        setUser(userData)
        fetchSubordinates(userData.id)
      } catch {
        localStorage.removeItem('auth_token')
        localStorage.removeItem('auth_user')
      }
    }
    setIsLoading(false)
  }, [])

  const fetchSubordinates = async (userId: string) => {
    try {
      const res = await fetch(`${API_URL}/ad/subordinates/${userId}`)
      if (res.ok) {
        const data = await res.json()
        setSubordinates(data)
      }
    } catch (error) {
      console.error('Failed to fetch subordinates:', error)
    }
  }

  const login = async (username: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const formData = new FormData()
      formData.append('username', username)
      formData.append('password', password)

      const res = await fetch(`${API_URL}/ad/authenticate`, {
        method: 'POST',
        body: formData
      })

      const data = await res.json()

      if (data.authenticated && data.employee) {
        setUser(data.employee)
        setToken(data.token)

        localStorage.setItem('auth_token', data.token)
        localStorage.setItem('auth_user', JSON.stringify(data.employee))

        // Also set currentUserId for backwards compatibility
        localStorage.setItem('currentUserId', data.employee.id)

        fetchSubordinates(data.employee.id)

        return { success: true }
      } else {
        return { success: false, error: data.error || 'Неверные учётные данные' }
      }
    } catch (error) {
      console.error('Login error:', error)
      return { success: false, error: 'Ошибка подключения к серверу' }
    }
  }

  const logout = () => {
    setUser(null)
    setToken(null)
    setSubordinates([])

    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
    localStorage.removeItem('currentUserId')
  }

  const canAccessEmployee = (employeeId: string): boolean => {
    if (!user) return false

    // Can always access self
    if (employeeId === user.id) return true

    // Can access direct and indirect subordinates
    return subordinates.some(sub => sub.id === employeeId)
  }

  return (
    <AuthContext.Provider value={{
      user,
      token,
      isLoading,
      login,
      logout,
      subordinates,
      canAccessEmployee
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

// HOC for protected routes
export function withAuth<P extends object>(
  Component: React.ComponentType<P>
): React.FC<P> {
  return function ProtectedComponent(props: P) {
    const { user, isLoading } = useAuth()

    if (isLoading) {
      return (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      )
    }

    if (!user) {
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
      return null
    }

    return <Component {...props} />
  }
}
