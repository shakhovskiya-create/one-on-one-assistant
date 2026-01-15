'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  Users,
  Calendar,
  CalendarDays,
  FileText,
  Upload,
  BarChart3,
  Home,
  CheckSquare,
  LogOut,
  User
} from 'lucide-react'
import { useAuth } from '@/lib/auth'

export function SidebarWrapper() {
  const pathname = usePathname()

  // Don't show sidebar on login page
  if (pathname === '/login') {
    return null
  }

  return <Sidebar />
}

function Sidebar() {
  const { user, logout, subordinates } = useAuth()
  const pathname = usePathname()

  const navItems = [
    { href: '/', icon: Home, label: 'Дашборд' },
    { href: '/employees', icon: Users, label: 'Команда' },
    { href: '/calendar', icon: CalendarDays, label: 'Календарь' },
    { href: '/meetings', icon: Calendar, label: 'Встречи' },
    { href: '/tasks', icon: CheckSquare, label: 'Задачи' },
    { href: '/script', icon: FileText, label: 'Скрипт встречи' },
    { href: '/upload', icon: Upload, label: 'Загрузить запись' },
    { href: '/analytics', icon: BarChart3, label: 'Аналитика' },
  ]

  const isActive = (href: string) => {
    if (href === '/') return pathname === '/'
    return pathname.startsWith(href)
  }

  return (
    <aside className="w-64 min-w-64 flex-shrink-0 bg-white border-r border-gray-200 min-h-screen flex flex-col">
      {/* Header */}
      <div className="p-4 border-b">
        <h1 className="text-xl font-bold text-blue-600">1-on-1 Assistant</h1>
      </div>

      {/* User info */}
      {user && (
        <div className="p-4 border-b bg-gray-50">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
              <User size={20} className="text-blue-600" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="font-medium text-sm truncate">{user.name}</p>
              <p className="text-xs text-gray-500 truncate">{user.position}</p>
            </div>
          </div>
          {subordinates.length > 0 && (
            <div className="mt-2 text-xs text-gray-500">
              <Users size={12} className="inline mr-1" />
              {subordinates.length} подчинённых
            </div>
          )}
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={`flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
              isActive(item.href)
                ? 'bg-blue-50 text-blue-600 font-medium'
                : 'text-gray-700 hover:bg-gray-100'
            }`}
          >
            <item.icon size={20} />
            <span>{item.label}</span>
          </Link>
        ))}
      </nav>

      {/* Logout */}
      {user && (
        <div className="p-4 border-t">
          <button
            onClick={logout}
            className="flex items-center gap-3 px-3 py-2 w-full text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <LogOut size={20} />
            <span>Выйти</span>
          </button>
        </div>
      )}
    </aside>
  )
}
