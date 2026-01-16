'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import {
  ArrowLeft,
  Calendar,
  Trash2,
  CheckCircle,
  Clock,
  AlertCircle,
  Mail,
  Building,
  User,
  Users,
  TrendingUp,
  TrendingDown,
  Minus,
  AlertTriangle,
  Flame,
  Target,
  Phone,
  BarChart3
} from 'lucide-react'

import { API_URL } from '@/lib/config'

interface Employee {
  id: string
  name: string
  position: string
  email: string | null
  department: string | null
  phone: string | null
  mobile: string | null
  manager_id: string | null
  meeting_frequency: string
  meeting_day: string | null
  development_priorities: string | null
  photo_base64: string | null
}

interface Task {
  id: string
  title: string
  status: string
  priority: number
  due_date: string | null
  completed_at: string | null
}

interface Meeting {
  id: string
  date: string
  title: string | null
  summary: string | null
  mood_score: number | null
  analysis: {
    red_flags?: {
      burnout_signs?: string | boolean
      turnover_risk?: string
      team_conflicts?: string | boolean
    }
    mood_trend?: string
    positive_signals?: string[]
    recommendations?: string[]
  } | null
}

interface Dossier {
  employee: Employee
  one_on_one_count: number
  project_meetings_count: number
  tasks: {
    total: number
    done: number
    in_progress: number
  }
  mood_history: { date: string; score: number }[]
  red_flags_history: { date: string; flags: object }[]
  recent_meetings: Meeting[]
}

interface Manager {
  id: string
  name: string
  position: string
}

export default function EmployeeDossierPage() {
  const params = useParams()
  const router = useRouter()
  const [dossier, setDossier] = useState<Dossier | null>(null)
  const [manager, setManager] = useState<Manager | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (params.id) {
      fetchData()
    }
  }, [params.id])

  const fetchData = async () => {
    try {
      const [dossierRes, tasksRes] = await Promise.all([
        fetch(`${API_URL}/employees/${params.id}/dossier`),
        fetch(`${API_URL}/tasks?assignee_id=${params.id}`)
      ])

      if (dossierRes.ok) {
        const dossierData = await dossierRes.json()
        setDossier(dossierData)

        // Fetch manager if exists
        if (dossierData.employee.manager_id) {
          const managerRes = await fetch(`${API_URL}/employees/${dossierData.employee.manager_id}`)
          if (managerRes.ok) {
            setManager(await managerRes.json())
          }
        }
      }

      if (tasksRes.ok) {
        setTasks(await tasksRes.json())
      }
    } catch (error) {
      console.error('Failed to fetch data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Удалить сотрудника?')) return

    try {
      const res = await fetch(`${API_URL}/employees/${params.id}`, {
        method: 'DELETE'
      })
      if (res.ok) {
        router.push('/employees')
      }
    } catch (error) {
      console.error('Failed to delete:', error)
    }
  }

  // Calculate average mood
  const getAverageMood = (history: { date: string; score: number }[]) => {
    if (history.length === 0) return null
    return (history.reduce((sum, h) => sum + h.score, 0) / history.length).toFixed(1)
  }

  // Get mood trend
  const getMoodTrend = (history: { date: string; score: number }[]) => {
    if (history.length < 2) return 'stable'
    const recent = history.slice(-3)
    const older = history.slice(-6, -3)

    if (older.length === 0) return 'stable'

    const recentAvg = recent.reduce((sum, h) => sum + h.score, 0) / recent.length
    const olderAvg = older.reduce((sum, h) => sum + h.score, 0) / older.length

    if (recentAvg - olderAvg > 0.5) return 'improving'
    if (olderAvg - recentAvg > 0.5) return 'declining'
    return 'stable'
  }

  const getMoodColor = (score: number) => {
    if (score >= 7) return 'text-green-600 bg-green-100'
    if (score >= 5) return 'text-yellow-600 bg-yellow-100'
    return 'text-red-600 bg-red-100'
  }

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'improving':
        return <TrendingUp className="text-green-500" size={16} />
      case 'declining':
        return <TrendingDown className="text-red-500" size={16} />
      default:
        return <Minus className="text-gray-400" size={16} />
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (!dossier) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Сотрудник не найден</p>
        <Link href="/employees" className="text-blue-600 hover:underline mt-2 inline-block">
          Вернуться к списку
        </Link>
      </div>
    )
  }

  const { employee, mood_history, red_flags_history, recent_meetings } = dossier
  const avgMood = getAverageMood(mood_history)
  const moodTrend = getMoodTrend(mood_history)

  const activeTasks = tasks.filter(t => t.status !== 'done')
  const completedTasks = tasks.filter(t => t.status === 'done')
  const overdueTasks = tasks.filter(t =>
    t.due_date && new Date(t.due_date) < new Date() && t.status !== 'done'
  )

  // Count recent red flags
  const recentRedFlags = red_flags_history.filter(rf => {
    const date = new Date(rf.date)
    const monthAgo = new Date()
    monthAgo.setMonth(monthAgo.getMonth() - 1)
    return date > monthAgo
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/employees" className="p-2 hover:bg-gray-100 rounded-lg">
          <ArrowLeft size={20} />
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">Досье сотрудника</h1>
        </div>
        <Link
          href={`/analytics?employee=${employee.id}`}
          className="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-gray-50"
        >
          <BarChart3 size={18} />
          Аналитика
        </Link>
        <button
          onClick={handleDelete}
          className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
        >
          <Trash2 size={20} />
        </button>
      </div>

      {/* Profile Card */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-700 rounded-lg p-6 text-white">
        <div className="flex items-start gap-6">
          {/* Avatar */}
          <div className="w-24 h-24 bg-white/20 rounded-full flex items-center justify-center flex-shrink-0">
            {employee.photo_base64 ? (
              <img
                src={`data:image/jpeg;base64,${employee.photo_base64}`}
                alt={employee.name}
                className="w-full h-full rounded-full object-cover"
              />
            ) : (
              <User size={48} className="text-white/80" />
            )}
          </div>

          {/* Info */}
          <div className="flex-1">
            <h2 className="text-2xl font-bold">{employee.name}</h2>
            <p className="text-blue-100 text-lg">{employee.position}</p>

            <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              {employee.email && (
                <div className="flex items-center gap-2">
                  <Mail size={14} className="text-blue-200" />
                  <span className="truncate">{employee.email}</span>
                </div>
              )}
              {employee.department && (
                <div className="flex items-center gap-2">
                  <Building size={14} className="text-blue-200" />
                  <span>{employee.department}</span>
                </div>
              )}
              {(employee.phone || employee.mobile) && (
                <div className="flex items-center gap-2">
                  <Phone size={14} className="text-blue-200" />
                  <span>{employee.mobile || employee.phone}</span>
                </div>
              )}
              {manager && (
                <div className="flex items-center gap-2">
                  <Users size={14} className="text-blue-200" />
                  <Link
                    href={`/employees/${manager.id}`}
                    className="hover:underline"
                  >
                    {manager.name}
                  </Link>
                </div>
              )}
            </div>
          </div>

          {/* Quick Stats */}
          <div className="flex gap-4">
            {avgMood && (
              <div className="text-center bg-white/10 rounded-lg p-4">
                <div className="flex items-center gap-1 justify-center">
                  <span className="text-2xl font-bold">{avgMood}</span>
                  {getTrendIcon(moodTrend)}
                </div>
                <p className="text-xs text-blue-200">Настроение</p>
              </div>
            )}
            <div className="text-center bg-white/10 rounded-lg p-4">
              <p className="text-2xl font-bold">{dossier.one_on_one_count}</p>
              <p className="text-xs text-blue-200">1-на-1</p>
            </div>
            <div className="text-center bg-white/10 rounded-lg p-4">
              <p className="text-2xl font-bold">{tasks.length}</p>
              <p className="text-xs text-blue-200">Задач</p>
            </div>
          </div>
        </div>
      </div>

      {/* Red Flags Alert */}
      {recentRedFlags.length > 0 && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="text-red-500 flex-shrink-0" size={20} />
            <div>
              <h3 className="font-semibold text-red-800">
                Обнаружено {recentRedFlags.length} красных флагов за последний месяц
              </h3>
              <div className="mt-2 space-y-1">
                {recentRedFlags.slice(0, 3).map((rf, i) => (
                  <p key={i} className="text-sm text-red-700">
                    {new Date(rf.date).toLocaleDateString('ru-RU')}: {' '}
                    {(rf.flags as { burnout_signs?: string | boolean })?.burnout_signs && 'Признаки выгорания'}
                    {(rf.flags as { turnover_risk?: string })?.turnover_risk === 'high' && ' Высокий риск ухода'}
                  </p>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Mood History */}
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <h3 className="font-semibold mb-4 flex items-center gap-2">
            <TrendingUp size={18} className="text-gray-400" />
            История настроения
          </h3>

          {mood_history.length === 0 ? (
            <p className="text-gray-400 text-sm">Нет данных</p>
          ) : (
            <div className="space-y-2">
              {/* Simple bar chart */}
              <div className="flex items-end gap-1 h-20 mb-4">
                {mood_history.slice(-10).map((m, i) => (
                  <div
                    key={i}
                    className={`flex-1 rounded-t ${
                      m.score >= 7 ? 'bg-green-400' : m.score >= 5 ? 'bg-yellow-400' : 'bg-red-400'
                    }`}
                    style={{ height: `${m.score * 10}%` }}
                    title={`${m.date}: ${m.score}/10`}
                  />
                ))}
              </div>

              {/* Recent values */}
              {mood_history.slice(-5).reverse().map((m, i) => (
                <div key={i} className="flex items-center justify-between text-sm">
                  <span className="text-gray-500">
                    {new Date(m.date).toLocaleDateString('ru-RU')}
                  </span>
                  <span className={`px-2 py-0.5 rounded ${getMoodColor(m.score)}`}>
                    {m.score}/10
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Tasks Overview */}
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <h3 className="font-semibold mb-4 flex items-center gap-2">
            <Target size={18} className="text-gray-400" />
            Задачи
          </h3>

          <div className="grid grid-cols-2 gap-3 mb-4">
            <div className="text-center p-3 bg-blue-50 rounded-lg">
              <p className="text-2xl font-bold text-blue-600">{activeTasks.length}</p>
              <p className="text-xs text-gray-500">Активные</p>
            </div>
            <div className="text-center p-3 bg-green-50 rounded-lg">
              <p className="text-2xl font-bold text-green-600">{completedTasks.length}</p>
              <p className="text-xs text-gray-500">Выполнено</p>
            </div>
          </div>

          {overdueTasks.length > 0 && (
            <div className="p-3 bg-red-50 rounded-lg mb-4">
              <div className="flex items-center gap-2 text-red-700">
                <AlertCircle size={16} />
                <span className="font-medium">{overdueTasks.length} просроченных</span>
              </div>
            </div>
          )}

          {/* Recent tasks */}
          <div className="space-y-2">
            {activeTasks.slice(0, 4).map(task => (
              <div key={task.id} className="flex items-center gap-2 p-2 bg-gray-50 rounded text-sm">
                <Clock size={14} className="text-gray-400 flex-shrink-0" />
                <span className="truncate flex-1">{task.title}</span>
              </div>
            ))}
          </div>

          <Link
            href={`/tasks?assignee=${employee.id}`}
            className="block text-center text-sm text-blue-600 hover:underline mt-4"
          >
            Все задачи
          </Link>
        </div>

        {/* Recent Meetings */}
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <h3 className="font-semibold mb-4 flex items-center gap-2">
            <Calendar size={18} className="text-gray-400" />
            Последние встречи
          </h3>

          {recent_meetings.length === 0 ? (
            <p className="text-gray-400 text-sm">Нет встреч</p>
          ) : (
            <div className="space-y-3">
              {recent_meetings.slice(0, 5).map(meeting => (
                <Link
                  key={meeting.id}
                  href={`/meetings/${meeting.id}`}
                  className="block p-3 bg-gray-50 rounded-lg hover:bg-gray-100"
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm font-medium">
                      {new Date(meeting.date).toLocaleDateString('ru-RU')}
                    </span>
                    {meeting.mood_score && (
                      <span className={`text-xs px-2 py-0.5 rounded ${getMoodColor(meeting.mood_score)}`}>
                        {meeting.mood_score}/10
                      </span>
                    )}
                  </div>
                  {meeting.summary && (
                    <p className="text-xs text-gray-500 line-clamp-2">{meeting.summary}</p>
                  )}
                  {meeting.analysis?.red_flags?.burnout_signs && (
                    <div className="flex items-center gap-1 mt-1 text-xs text-red-600">
                      <Flame size={12} />
                      Признаки выгорания
                    </div>
                  )}
                </Link>
              ))}
            </div>
          )}

          <Link
            href={`/meetings?employee=${employee.id}`}
            className="block text-center text-sm text-blue-600 hover:underline mt-4"
          >
            Все встречи
          </Link>
        </div>
      </div>

      {/* Development Priorities */}
      {employee.development_priorities && (
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <h3 className="font-semibold mb-3 flex items-center gap-2">
            <Target size={18} className="text-gray-400" />
            Приоритеты развития
          </h3>
          <p className="text-gray-700 whitespace-pre-wrap">{employee.development_priorities}</p>
        </div>
      )}

      {/* Recommendations from last meeting */}
      {recent_meetings[0]?.analysis?.recommendations && recent_meetings[0].analysis.recommendations.length > 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
          <h3 className="font-semibold mb-3 text-yellow-800">
            Рекомендации с последней встречи
          </h3>
          <ul className="space-y-2">
            {recent_meetings[0].analysis.recommendations.map((rec: string, i: number) => (
              <li key={i} className="flex items-start gap-2 text-yellow-700">
                <CheckCircle size={16} className="flex-shrink-0 mt-0.5" />
                <span>{rec}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
