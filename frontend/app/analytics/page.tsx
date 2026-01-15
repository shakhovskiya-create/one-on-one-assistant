'use client'

import { Suspense } from 'react'
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
  BarChart,
  Bar,
  Legend,
} from 'recharts'
import {
  TrendingUp,
  TrendingDown,
  Minus,
  AlertTriangle,
  Users,
  Target,
  Calendar,
  CheckCircle,
  Clock,
  MessageSquare,
  Zap,
  RefreshCw
} from 'lucide-react'

import { API_URL } from '@/lib/config'

interface Employee {
  id: string
  name: string
  position: string
}

interface Category {
  id: string
  code: string
  name: string
  description: string | null
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
    team_conflicts?: boolean | string
  }
}

interface Analytics {
  mood_history: MoodData[]
  agreement_stats: AgreementStats
  red_flags_history: RedFlag[]
  total_meetings: number
}

interface CategoryAnalytics {
  category: Category
  meetings_count: number
  avg_duration: number
  avg_mood: number | null
  mood_trend: 'up' | 'down' | 'stable' | null
  agreements_created: number
  agreements_completed: number
  common_topics: string[]
  red_flags_count: number
  // Category-specific
  specific_metrics: Record<string, number | string | null>
}

function AnalyticsContent() {
  const searchParams = useSearchParams()
  const employeeIdParam = searchParams.get('employee')

  const [employees, setEmployees] = useState<Employee[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [selectedEmployee, setSelectedEmployee] = useState(employeeIdParam || '')
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [analytics, setAnalytics] = useState<Analytics | null>(null)
  const [categoryAnalytics, setCategoryAnalytics] = useState<CategoryAnalytics[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchEmployees()
    fetchCategories()
  }, [])

  useEffect(() => {
    if (selectedEmployee) {
      fetchAnalytics(selectedEmployee)
      fetchCategoryAnalytics(selectedEmployee)
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

  const fetchCategories = async () => {
    try {
      const res = await fetch(`${API_URL}/meeting-categories`)
      if (res.ok) {
        const data = await res.json()
        setCategories(data)
      }
    } catch (error) {
      console.error('Failed to fetch categories:', error)
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

  const fetchCategoryAnalytics = async (employeeId: string) => {
    try {
      const res = await fetch(`${API_URL}/analytics/employee/${employeeId}/by-category`)
      if (res.ok) {
        const data = await res.json()
        setCategoryAnalytics(data)
      }
    } catch (error) {
      // If endpoint doesn't exist yet, generate mock data
      console.error('Failed to fetch category analytics:', error)
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
  const CATEGORY_COLORS: Record<string, string> = {
    one_on_one: '#3b82f6',
    team_meeting: '#8b5cf6',
    planning: '#10b981',
    retro: '#f59e0b',
    kickoff: '#ec4899',
    interview: '#6366f1',
    status: '#14b8a6',
    demo: '#f97316',
  }

  const pieData = analytics
    ? [
        { name: 'Выполнено', value: analytics.agreement_stats.completed },
        { name: 'В работе', value: analytics.agreement_stats.pending },
        { name: 'Просрочено', value: analytics.agreement_stats.overdue },
      ].filter((d) => d.value > 0)
    : []

  const selectedEmployeeData = employees.find((e) => e.id === selectedEmployee)
  const moodTrend = getMoodTrend()

  // Category-specific rendering
  const renderCategorySpecificAnalytics = (catCode: string) => {
    const catAnalytics = categoryAnalytics.find(ca => ca.category.code === catCode)

    switch (catCode) {
      case 'one_on_one':
        return (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Mood over time for 1-on-1s */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <TrendingUp size={18} className="text-blue-500" />
                Динамика настроения (1-на-1)
              </h3>
              {analytics && analytics.mood_history.length > 0 ? (
                <ResponsiveContainer width="100%" height={250}>
                  <LineChart data={analytics.mood_history}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="date"
                      tickFormatter={(value) =>
                        new Date(value).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
                      }
                    />
                    <YAxis domain={[0, 10]} />
                    <Tooltip />
                    <Line type="monotone" dataKey="score" stroke="#3b82f6" strokeWidth={2} />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="h-[250px] flex items-center justify-center text-gray-400">
                  Нет данных
                </div>
              )}
            </div>

            {/* Red flags timeline */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <AlertTriangle size={18} className="text-red-500" />
                Красные флаги
              </h3>
              {analytics && analytics.red_flags_history.length > 0 ? (
                <div className="space-y-2 max-h-[250px] overflow-y-auto">
                  {analytics.red_flags_history.slice(0, 5).map((flag, i) => (
                    <div key={i} className="p-3 bg-red-50 rounded-lg text-sm">
                      <p className="font-medium">{new Date(flag.date).toLocaleDateString('ru-RU')}</p>
                      {flag.flags.burnout_signs && <p className="text-red-600">Признаки выгорания</p>}
                      {flag.flags.turnover_risk !== 'low' && (
                        <p className="text-red-600">Риск ухода: {flag.flags.turnover_risk === 'high' ? 'Высокий' : 'Средний'}</p>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="h-[250px] flex items-center justify-center text-green-500">
                  <CheckCircle size={32} className="mr-2" />
                  Красных флагов не обнаружено
                </div>
              )}
            </div>

            {/* Agreements */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <Target size={18} className="text-green-500" />
                Договорённости
              </h3>
              {analytics && analytics.agreement_stats.total > 0 ? (
                <>
                  <ResponsiveContainer width="100%" height={200}>
                    <PieChart>
                      <Pie
                        data={pieData}
                        cx="50%"
                        cy="50%"
                        innerRadius={50}
                        outerRadius={80}
                        paddingAngle={5}
                        dataKey="value"
                      >
                        {pieData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index]} />
                        ))}
                      </Pie>
                      <Tooltip />
                      <Legend />
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="text-center mt-2 text-sm text-gray-500">
                    Всего: {analytics.agreement_stats.total} договорённостей
                  </div>
                </>
              ) : (
                <div className="h-[200px] flex items-center justify-center text-gray-400">
                  Нет договорённостей
                </div>
              )}
            </div>

            {/* Summary stats */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">Сводка по 1-на-1</h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-4 bg-blue-50 rounded-lg">
                  <p className="text-3xl font-bold text-blue-600">{analytics?.total_meetings || 0}</p>
                  <p className="text-sm text-gray-500">Встреч проведено</p>
                </div>
                <div className="text-center p-4 bg-green-50 rounded-lg">
                  <p className="text-3xl font-bold text-green-600">
                    {analytics?.mood_history.length ? (
                      (analytics.mood_history.reduce((s, m) => s + m.score, 0) / analytics.mood_history.length).toFixed(1)
                    ) : '-'}
                  </p>
                  <p className="text-sm text-gray-500">Среднее настроение</p>
                </div>
                <div className="text-center p-4 bg-yellow-50 rounded-lg">
                  <p className="text-3xl font-bold text-yellow-600">{analytics?.agreement_stats.pending || 0}</p>
                  <p className="text-sm text-gray-500">В работе</p>
                </div>
                <div className="text-center p-4 bg-red-50 rounded-lg">
                  <p className="text-3xl font-bold text-red-600">{analytics?.red_flags_history.length || 0}</p>
                  <p className="text-sm text-gray-500">Красных флагов</p>
                </div>
              </div>
            </div>
          </div>
        )

      case 'retro':
        return (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <RefreshCw size={18} className="text-yellow-500" />
                Метрики ретроспектив
              </h3>
              <div className="space-y-4">
                <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                  <span>Что прошло хорошо</span>
                  <span className="text-green-600 font-bold">+24</span>
                </div>
                <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                  <span>Что можно улучшить</span>
                  <span className="text-yellow-600 font-bold">18</span>
                </div>
                <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                  <span>Action items</span>
                  <span className="text-blue-600 font-bold">12</span>
                </div>
                <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                  <span>Выполнено action items</span>
                  <span className="text-green-600 font-bold">8 (67%)</span>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <MessageSquare size={18} className="text-purple-500" />
                Частые темы
              </h3>
              <div className="space-y-2">
                {['Коммуникация в команде', 'Технический долг', 'Процессы код-ревью', 'Планирование спринтов', 'Документация'].map((topic, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <div className="flex-1 bg-gray-100 rounded-full h-4">
                      <div
                        className="bg-purple-500 h-4 rounded-full"
                        style={{ width: `${100 - i * 15}%` }}
                      />
                    </div>
                    <span className="text-sm text-gray-600 w-40 truncate">{topic}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )

      case 'planning':
        return (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <Zap size={18} className="text-green-500" />
                Метрики планирования
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-4 bg-green-50 rounded-lg">
                  <p className="text-3xl font-bold text-green-600">85%</p>
                  <p className="text-sm text-gray-500">Точность оценок</p>
                </div>
                <div className="text-center p-4 bg-blue-50 rounded-lg">
                  <p className="text-3xl font-bold text-blue-600">42</p>
                  <p className="text-sm text-gray-500">Задач запланировано</p>
                </div>
                <div className="text-center p-4 bg-yellow-50 rounded-lg">
                  <p className="text-3xl font-bold text-yellow-600">3</p>
                  <p className="text-sm text-gray-500">Блокеров выявлено</p>
                </div>
                <div className="text-center p-4 bg-purple-50 rounded-lg">
                  <p className="text-3xl font-bold text-purple-600">2.5ч</p>
                  <p className="text-sm text-gray-500">Ср. время планирования</p>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">Выполнение спринтов</h3>
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={[
                  { name: 'Спринт 1', planned: 20, done: 18 },
                  { name: 'Спринт 2', planned: 22, done: 20 },
                  { name: 'Спринт 3', planned: 25, done: 22 },
                  { name: 'Спринт 4', planned: 23, done: 23 },
                ]}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="planned" fill="#94a3b8" name="Запланировано" />
                  <Bar dataKey="done" fill="#22c55e" name="Выполнено" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>
        )

      case 'status':
        return (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4 flex items-center gap-2">
                <Clock size={18} className="text-teal-500" />
                Статус-митинги
              </h3>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span>Проведено статусов</span>
                  <span className="font-bold">24</span>
                </div>
                <div className="flex justify-between items-center">
                  <span>Средняя длительность</span>
                  <span className="font-bold">25 мин</span>
                </div>
                <div className="flex justify-between items-center">
                  <span>Выявлено блокеров</span>
                  <span className="font-bold text-red-600">8</span>
                </div>
                <div className="flex justify-between items-center">
                  <span>Разрешено блокеров</span>
                  <span className="font-bold text-green-600">7</span>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h3 className="font-semibold mb-4">Частота проблем</h3>
              <div className="space-y-3">
                {[
                  { issue: 'Задержки от смежных команд', count: 5 },
                  { issue: 'Нехватка ресурсов', count: 3 },
                  { issue: 'Изменение требований', count: 4 },
                  { issue: 'Технические сложности', count: 2 },
                ].map((item, i) => (
                  <div key={i} className="flex items-center gap-3">
                    <div className="flex-1">
                      <div className="flex justify-between mb-1">
                        <span className="text-sm">{item.issue}</span>
                        <span className="text-sm font-medium">{item.count}</span>
                      </div>
                      <div className="h-2 bg-gray-100 rounded-full">
                        <div
                          className="h-2 bg-teal-500 rounded-full"
                          style={{ width: `${item.count * 20}%` }}
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )

      default:
        return (
          <div className="bg-white rounded-lg shadow-sm border p-12 text-center text-gray-400">
            Выберите категорию для просмотра специализированной аналитики
          </div>
        )
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center flex-wrap gap-4">
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

      {/* Category tabs */}
      {selectedEmployee && (
        <div className="flex gap-2 overflow-x-auto pb-2">
          <button
            onClick={() => setSelectedCategory('all')}
            className={`px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap ${
              selectedCategory === 'all'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Все категории
          </button>
          {categories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.code)}
              className={`px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap ${
                selectedCategory === cat.code
                  ? 'text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              style={{
                backgroundColor: selectedCategory === cat.code ? CATEGORY_COLORS[cat.code] || '#6b7280' : undefined
              }}
            >
              {cat.name}
            </button>
          ))}
        </div>
      )}

      {!selectedEmployee && (
        <div className="bg-white rounded-lg shadow-sm border p-12 text-center">
          <Users size={48} className="mx-auto text-gray-400 mb-4" />
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
          {/* Employee header */}
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <h2 className="text-xl font-semibold">{selectedEmployeeData?.name}</h2>
            <p className="text-gray-500">{selectedEmployeeData?.position}</p>
            <div className="mt-4 flex gap-6 flex-wrap">
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
                  {!moodTrend && <span className="text-gray-400">Недостаточно данных</span>}
                </div>
              </div>
              <div>
                <p className="text-sm text-gray-500">Договорённостей</p>
                <p className="text-2xl font-bold">{analytics.agreement_stats.total}</p>
              </div>
            </div>
          </div>

          {/* Category-specific content */}
          {selectedCategory === 'all' ? (
            // Overview of all categories
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
                          new Date(value).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
                        }
                      />
                      <YAxis domain={[0, 10]} />
                      <Tooltip />
                      <Line type="monotone" dataKey="score" stroke="#3b82f6" strokeWidth={2} />
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
                  <>
                    <ResponsiveContainer width="100%" height={250}>
                      <PieChart>
                        <Pie
                          data={pieData}
                          cx="50%"
                          cy="50%"
                          innerRadius={60}
                          outerRadius={100}
                          paddingAngle={5}
                          dataKey="value"
                        >
                          {pieData.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index]} />
                          ))}
                        </Pie>
                        <Tooltip />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </>
                ) : (
                  <div className="h-[250px] flex items-center justify-center text-gray-400">
                    Нет договоренностей
                  </div>
                )}
              </div>

              {analytics.red_flags_history.length > 0 && (
                <div className="bg-white rounded-lg shadow-sm border p-6 lg:col-span-2">
                  <h3 className="font-semibold mb-4 flex items-center gap-2">
                    <AlertTriangle className="text-red-600" size={20} />
                    История красных флагов
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    {analytics.red_flags_history.map((flag, i) => (
                      <div key={i} className="p-4 bg-red-50 border border-red-200 rounded-lg">
                        <p className="text-sm text-gray-600 mb-2">
                          {new Date(flag.date).toLocaleDateString('ru-RU')}
                        </p>
                        {flag.flags.burnout_signs && (
                          <p className="text-red-700 text-sm">Признаки выгорания</p>
                        )}
                        {flag.flags.turnover_risk !== 'low' && (
                          <p className="text-red-700 text-sm">
                            Риск ухода: {flag.flags.turnover_risk === 'high' ? 'Высокий' : 'Средний'}
                          </p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            renderCategorySpecificAnalytics(selectedCategory)
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
