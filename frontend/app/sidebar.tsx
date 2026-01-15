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
  User,
  Settings
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
    { href: '/settings', icon: Settings, label: 'Настройки' },
  ]

  const isActive = (href: string) => {
    if (href === '/') return pathname === '/'
    return pathname.startsWith(href)
  }

  return (
    <aside className="w-64 min-w-64 flex-shrink-0 bg-white border-r border-gray-200 min-h-screen flex flex-col">
      {/* Header with EKF brand */}
      <div className="p-4 border-b bg-ekf-dark">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-ekf-orange rounded flex items-center justify-center">
            <span className="text-white font-bold text-sm">EKF</span>
          </div>
          <div>
            <h1 className="text-lg font-bold text-white">Team Hub</h1>
            <p className="text-xs text-gray-400">Управление командой</p>
          </div>
        </div>
      </div>

      {/* User info */}
      {user && (
        <div className="p-4 border-b bg-gray-50">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-ekf-orange/10 rounded-full flex items-center justify-center border-2 border-ekf-orange">
              <User size={20} className="text-ekf-orange" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="font-medium text-sm text-ekf-dark truncate">{user.name}</p>
              <p className="text-xs text-ekf-gray truncate">{user.position}</p>
            </div>
          </div>
          {subordinates.length > 0 && (
            <div className="mt-2 text-xs text-ekf-gray flex items-center gap-1">
              <Users size={12} />
              <span>{subordinates.length} подчинённых</span>
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
                ? 'bg-ekf-orange/10 text-ekf-orange font-medium'
                : 'text-ekf-dark hover:bg-gray-100'
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
            className="flex items-center gap-3 px-3 py-2 w-full text-ekf-gray hover:bg-gray-100 hover:text-ekf-dark rounded-lg transition-colors"
          >
            <LogOut size={20} />
            <span>Выйти</span>
          </button>
        </div>
      )}
    </aside>
  )
}
