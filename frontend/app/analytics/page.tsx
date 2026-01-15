'use client'

import { Suspense } from 'react';
import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import { TrendingUp, TrendingDown, Minus, AlertTriangle } from 'lucide-react'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000'

interface Employee {
  id: string
  name: string
  position: string
}

interface MoodData {
  date: string
  score: number
}

interface AgreementStats {
  total: number
  completed: number
  pending: number
  overdue: number
}

interface RedFlag {
  date: string
  flags: {
    burnout_signs: boolean | string
    turnover_risk: string
  }
}

interface Analytics {
  mood_history: MoodData[]
  agreement_stats: AgreementStats
  red_flags_history: RedFlag[]
  total_meetings: number
}

function AnalyticsContent() {
  const searchParams = useSearchParams()
  const employeeIdParam = searchParams.get('employee')
  
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selectedEmployee, setSelectedEmployee] = useState(employeeIdParam || '')
  const [analytics, setAnalytics] = useState<Analytics | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchEmployees()
  }, [])

  useEffect(() => {
    if (selectedEmployee) {
      fetchAnalytics(selectedEmployee)
    }
  }, [selectedEmployee])

  const fetchEmployees = async () => {
    try {
      const res = await fetch(`${API_URL}/employees`)
      if (res.ok) {
        const data = await res.json()
        setEmployees(data)
        if (!selectedEmployee && data.length > 0) {
          setSelectedEmployee(employeeIdParam || data[0].id)
        }
      }
    } catch (error) {
      console.error('Failed to fetch employees:', error)
    }
  }

  const fetchAnalytics = async (employeeId: string) => {
    setLoading(true)
    try {
      const res = await fetch(`${API_URL}/analytics/employee/${employeeId}`)
      if (res.ok) {
        const data = await res.json()
        setAnalytics(data)
      }
    } catch (error) {
      console.error('Failed to fetch analytics:', error)
    } finally {
      setLoading(false)
    }
  }

  const getMoodTrend = () => {
    if (!analytics || analytics.mood_history.length < 2) return null
    const recent = analytics.mood_history.slice(-3)
    if (recent.length < 2) return null
    
    const avg = recent.reduce((sum, m) => sum + m.score, 0) / recent.length
    const prevAvg =
      analytics.mood_history
        .slice(-6, -3)
        .reduce((sum, m) => sum + m.score, 0) /
      Math.max(analytics.mood_history.slice(-6, -3).length, 1)
    
    const diff = avg - prevAvg
    if (diff > 0.5) return 'up'
    if (diff < -0.5) return 'down'
    return 'stable'
  }

  const COLORS = ['#22c55e', '#f59e0b', '#ef4444', '#94a3b8']

  const pieData = analytics
    ? [
        { name: 'Выполнено', value: analytics.agreement_stats.completed },
        { name: 'В работе', value: analytics.agreement_stats.pending },
        { name: 'Просрочено', value: analytics.agreement_stats.overdue },
      ].filter((d) => d.value > 0)
    : []

  const selectedEmployeeData = employees.find((e) => e.id === selectedEmployee)
  const moodTrend = getMoodTrend()

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Аналитика</h1>
        <select
          value={selectedEmployee}
          onChange={(e) => setSelectedEmployee(e.target.value)}
          className="border rounded-lg px-4 py-2"
        >
          <option value="">Выберите сотрудника</option>
          {employees.map((emp) => (
            <option key={emp.id} value={emp.id}>
              {emp.name}
            </option>
          ))}
        </select>
      </div>

      {!selectedEmployee && (
        <div className="bg-white rounded-lg shadow-sm border p-12 text-center">
          <p className="text-gray-500">Выберите сотрудника для просмотра аналитики</p>
        </div>
      )}

      {loading && (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      )}

      {selectedEmployee && analytics && !loading && (
        <>
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <h2 className="text-xl font-semibold">{selectedEmployeeData?.name}</h2>
            <p className="text-gray-500">{selectedEmployeeData?.position}</p>
            <div className="mt-4 flex gap-6">
              <div>
                <p className="text-sm text-gray-500">Всего встреч</p>
                <p className="text-2xl font-bold">{analytics.total_meetings}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Тренд настроения</p>
                <div className="flex items-center gap-2">
                  {moodTrend === 'up' && (
                    <>
                      <TrendingUp className="text-green-600" size={24} />
                      <span className="text-green-600 font-medium">Растет</span>
                    </>
                  )}
                  {moodTrend === 'down' && (
                    <>
                      <TrendingDown className="text-red-600" size={24} />
                      <span className="text-red-600 font-medium">Падает</span>
                    </>
                  )}
                  {moodTrend === 'stable' && (
                    <>
                      <Minus className="text-gray-600" size={24} />
                      <span className="text-gray-600 font-medium">Стабильно</span>
                    </>
                  )}
                  {!moodTrend && (
                    <span className="text-gray-400">Недостаточно данных</span>
                  )}
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">Динамика настроения</h3>
              {analytics.mood_history.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={analytics.mood_history}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="date"
                      tickFormatter={(value) =>
                        new Date(value).toLocaleDateString('ru-RU', {
                          day: 'numeric',
                          month: 'short',
                        })
                      }
                    />
                    <YAxis domain={[0, 10]} />
                    <Tooltip
                      labelFormatter={(value) =>
                        new Date(value).toLocaleDateString('ru-RU')
                      }
                      formatter={(value: number) => [`${value}/10`, 'Настроение']}
                    />
                    <Line
                      type="monotone"
                      dataKey="score"
                      stroke="#3b82f6"
                      strokeWidth={2}
                      dot={{ fill: '#3b82f6' }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="h-[300px] flex items-center justify-center text-gray-400">
                  Нет данных о настроении
                </div>
              )}
            </div>

            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">Статус договоренностей</h3>
              {analytics.agreement_stats.total > 0 ? (
                <div className="flex items-center justify-center">
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={pieData}
                        cx="50%"
                        cy="50%"
                        innerRadius={60}
                        outerRadius={100}
                        paddingAngle={5}
                        dataKey="value"
                        label={({ name, percent }) =>
                          `${name} ${(percent * 100).toFixed(0)}%`
                        }
                      >
                        {pieData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              ) : (
                <div className="h-[300px] flex items-center justify-center text-gray-400">
                  Нет договоренностей
                </div>
              )}
              <div className="grid grid-cols-3 gap-4 mt-4">
                <div className="text-center">
                  <p className="text-2xl font-bold text-green-600">
                    {analytics.agreement_stats.completed}
                  </p>
                  <p className="text-sm text-gray-500">Выполнено</p>
                </div>
                <div className="text-center">
                  <p className="text-2xl font-bold text-yellow-600">
                    {analytics.agreement_stats.pending}
                  </p>
                  <p className="text-sm text-gray-500">В работе</p>
                </div>
                <div className="text-center">
                  <p className="text-2xl font-bold text-red-600">
                    {analytics.agreement_stats.overdue}
                  </p>
                  <p className="text-sm text-gray-500">Просрочено</p>
                </div>
              </div>
            </div>
          </div>

          {analytics.red_flags_history.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <AlertTriangle className="text-red-600" size={20} />
                История красных флагов
              </h3>
              <div className="space-y-3">
                {analytics.red_flags_history.map((flag, i) => (
                  <div
                    key={i}
                    className="p-4 bg-red-50 border border-red-200 rounded-lg"
                  >
                    <p className="text-sm text-gray-600 mb-2">
                      {new Date(flag.date).toLocaleDateString('ru-RU')}
                    </p>
                    {flag.flags.burnout_signs && (
                      <p className="text-red-700">
                        Признаки выгорания:{' '}
                        {typeof flag.flags.burnout_signs === 'string'
                          ? flag.flags.burnout_signs
                          : 'Да'}
                      </p>
                    )}
                    {flag.flags.turnover_risk !== 'low' && (
                      <p className="text-red-700">
                        Риск ухода:{' '}
                        {flag.flags.turnover_risk === 'high' ? 'Высокий' : 'Средний'}
                      </p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}

export default function AnalyticsPage() {
  return (
    <Suspense fallback={<div className="p-8">Загрузка...</div>}>
      <AnalyticsContent />
    </Suspense>
  )
}
