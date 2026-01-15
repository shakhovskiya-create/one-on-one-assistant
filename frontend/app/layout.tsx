import './globals.css'
import type { Metadata } from 'next'
import Link from 'next/link'
import { Users, Calendar, FileText, Upload, BarChart3, Home } from 'lucide-react'

export const metadata: Metadata = {
  title: '1-on-1 Assistant',
  description: 'Инструмент для проведения и анализа встреч 1-на-1',
}

function Sidebar() {
  const navItems = [
    { href: '/', icon: Home, label: 'Дашборд' },
    { href: '/employees', icon: Users, label: 'Команда' },
    { href: '/meetings', icon: Calendar, label: 'Встречи' },
    { href: '/script', icon: FileText, label: 'Скрипт встречи' },
    { href: '/upload', icon: Upload, label: 'Загрузить запись' },
    { href: '/analytics', icon: BarChart3, label: 'Аналитика' },
  ]

  return (
    <aside className="w-64 bg-white border-r border-gray-200 min-h-screen p-4">
      <div className="mb-8">
        <h1 className="text-xl font-bold text-blue-600">1-on-1 Assistant</h1>
      </div>
      <nav className="space-y-2">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="flex items-center gap-3 px-3 py-2 text-gray-700 rounded-lg hover:bg-gray-100 transition-colors"
          >
            <item.icon size={20} />
            <span>{item.label}</span>
          </Link>
        ))}
      </nav>
    </aside>
  )
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ru">
      <body className="flex">
        <Sidebar />
        <main className="flex-1 p-8">{children}</main>
      </body>
    </html>
  )
}
