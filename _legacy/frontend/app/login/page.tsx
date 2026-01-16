'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, Lock, User } from 'lucide-react'
import { useAuth } from '@/lib/auth'

export default function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const { user, login } = useAuth()
  const router = useRouter()

  // Redirect if already logged in
  useEffect(() => {
    if (user) {
      router.push('/')
    }
  }, [user, router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    const result = await login(username, password)

    if (result.success) {
      router.push('/')
    } else {
      setError(result.error || 'Ошибка входа')
    }

    setLoading(false)
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-ekf-light">
      <div className="max-w-md w-full">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-3 mb-4">
            <div className="w-12 h-12 bg-ekf-orange rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-lg">EKF</span>
            </div>
            <div className="text-left">
              <h1 className="text-2xl font-bold text-ekf-dark">Team Hub</h1>
              <p className="text-xs text-ekf-gray">Управление командой</p>
            </div>
          </div>
          <p className="text-ekf-gray mt-4">Войдите с учётными данными AD</p>
        </div>

        {/* Login Form */}
        <div className="bg-white rounded-lg shadow-lg p-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Error message */}
            {error && (
              <div className="bg-red-50 text-red-700 px-4 py-3 rounded-lg text-sm border border-red-200">
                {error}
              </div>
            )}

            {/* Username */}
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-ekf-dark mb-1">
                Логин
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <User size={18} className="text-ekf-gray-light" />
                </div>
                <input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="block w-full pl-10 pr-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
                  placeholder="DOMAIN\\username или email"
                  required
                  autoFocus
                />
              </div>
            </div>

            {/* Password */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-ekf-dark mb-1">
                Пароль
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Lock size={18} className="text-ekf-gray-light" />
                </div>
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="block w-full pl-10 pr-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
                  placeholder="Пароль AD"
                  required
                />
              </div>
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={loading}
              className="w-full flex items-center justify-center gap-2 py-2.5 px-4 bg-ekf-orange text-white rounded-lg hover:bg-ekf-orange-dark focus:ring-2 focus:ring-offset-2 focus:ring-ekf-orange disabled:opacity-50 transition-colors"
            >
              {loading ? (
                <>
                  <Loader2 size={18} className="animate-spin" />
                  Вход...
                </>
              ) : (
                'Войти'
              )}
            </button>
          </form>

          {/* Help text */}
          <p className="mt-6 text-center text-sm text-ekf-gray">
            Используйте ваши корпоративные учётные данные Active Directory
          </p>
        </div>

        {/* Footer */}
        <p className="mt-8 text-center text-xs text-ekf-gray-light">
          Для работы требуется подключение on-prem коннектора
        </p>
      </div>
    </div>
  )
}
