import './globals.css'
import type { Metadata } from 'next'
import { Providers } from './providers'
import { SidebarWrapper } from './sidebar'

export const metadata: Metadata = {
  title: '1-on-1 Assistant',
  description: 'Инструмент для проведения и анализа встреч 1-на-1',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ru">
      <body className="flex">
        <Providers>
          <SidebarWrapper />
          <main className="flex-1 p-8 overflow-x-auto">{children}</main>
        </Providers>
      </body>
    </html>
  )
}
