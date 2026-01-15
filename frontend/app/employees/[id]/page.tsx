'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, Calendar, Edit, Trash2, CheckCircle, Clock, AlertCircle } from 'lucide-react'

import { API_URL } from '@/lib/config'

interface Employee {
  id: string
  name: string
  position: string
  meeting_frequency: string
  meeting_day: string | null
  development_priorities: string | null
}

interface Task {
  id: string
  title: string
  status: string
  priority: number
  due_date: string | null
}

interface Meeting {
  id: string
  date: string
  status: string
  summary: string | null
}

export default function EmployeeProfilePage() {
  const params = useParams()
  const router = useRouter()
  const [employee, setEmployee] = useState<Employee | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [meetings, setMeetings] = useState<Meeting[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (params.id) {
      fetchData()
    }
  }, [params.id])

  const fetchData = async () => {
    try {
      const [empRes, tasksRes, meetingsRes] = await Promise.all([
        fetch(`${API_URL}/employees/${params.id}`),
        fetch(`${API_URL}/tasks?assignee_id=${params.id}`),
        fetch(`${API_URL}/meetings?employee_id=${params.id}`)
      ])

      if (empRes.ok) {
        setEmployee(await empRes.json())
      }
      if (tasksRes.ok) {
        setTasks(await tasksRes.json())
      }
      if (meetingsRes.ok) {
        setMeetings(await meetingsRes.json())
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

  const frequencyLabels: Record<string, string> = {
    weekly: 'Еженедельно',
    biweekly: 'Раз в 2 недели',
    monthly: 'Ежемесячно',
  }

  const dayLabels: Record<string, string> = {
    monday: 'Понедельник',
    tuesday: 'Вторник',
    wednesday: 'Среда',
    thursday: 'Четверг',
    friday: 'Пятница',
  }

  const statusConfig: Record<string, { label: string; color: string }> = {
    backlog: { label: 'Бэклог', color: 'bg-gray-100 text-gray-700' },
    todo: { label: 'К выполнению', color: 'bg-blue-100 text-blue-700' },
    in_progress: { label: 'В работе', color: 'bg-yellow-100 text-yellow-700' },
    review: { label: 'На проверке', color: 'bg-purple-100 text-purple-700' },
    done: { label: 'Готово', color: 'bg-green-100 text-green-700' },
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (!employee) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Сотрудник не найден</p>
        <Link href="/employees" className="text-blue-600 hover:underline mt-2 inline-block">
          Вернуться к списку
        </Link>
      </div>
    )
  }

  const activeTasks = tasks.filter(t => t.status !== 'done')
  const completedTasks = tasks.filter(t => t.status === 'done')

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href="/employees"
          className="p-2 hover:bg-gray-100 rounded-lg"
        >
          <ArrowLeft size={20} />
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">{employee.name}</h1>
          <p className="text-gray-500">{employee.position}</p>
        </div>
        <button
          onClick={handleDelete}
          className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
        >
          <Trash2 size={20} />
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Информация */}
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <h2 className="font-semibold mb-4">Информация</h2>
          <div className="space-y-3">
            <div>
              <span className="text-sm text-gray-500">Частота встреч</span>
              <p>{frequencyLabels[employee.meeting_frequency]}</p>
            </div>
            {employee.meeting_day && (
              <div>
                <span className="text-sm text-gray-500">День встречи</span>
                <p>{dayLabels[employee.meeting_day]}</p>
              </div>
            )}
            {employee.development_priorities && (
              <div>
                <span className="text-sm text-gray-500">Приоритеты развития</span>
                <p className="text-sm">{employee.development_priorities}</p>
              </div>
            )}
          </div>
        </div>

        {/* Активные задачи */}
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="font-semibold">Активные задачи</h2>
            <span className="text-sm text-gray-500">{activeTasks.length}</span>
          </div>
          {activeTasks.length === 0 ? (
            <p className="text-gray-400 text-sm">Нет активных задач</p>
          ) : (
            <div className="space-y-2">
              {activeTasks.slice(0, 5).map(task => (
                <div key={task.id} className="flex items-center gap-2 p-2 bg-gray-50 rounded">
                  <Clock size={14} className="text-gray-400" />
                  <span className="flex-1 text-sm truncate">{task.title}</span>
                  <span className={`text-xs px-2 py-0.5 rounded ${statusConfig[task.status]?.color || ''}`}>
                    {statusConfig[task.status]?.label || task.status}
                  </span>
                </div>
              ))}
              {activeTasks.length > 5 && (
                <Link href={`/tasks?assignee=${employee.id}`} className="text-sm text-blue-600 hover:underline">
                  Ещё {activeTasks.length - 5}...
                </Link>
              )}
            </div>
          )}
        </div>

        {/* Последние встречи */}
        <div className="bg-white rounded-lg shadow-sm border p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="font-semibold">Последние встречи</h2>
            <span className="text-sm text-gray-500">{meetings.length}</span>
          </div>
          {meetings.length === 0 ? (
            <p className="text-gray-400 text-sm">Нет встреч</p>
          ) : (
            <div className="space-y-2">
              {meetings.slice(0, 5).map(meeting => (
                <Link
                  key={meeting.id}
                  href={`/meetings/${meeting.id}`}
                  className="flex items-center gap-2 p-2 bg-gray-50 rounded hover:bg-gray-100"
                >
                  <Calendar size={14} className="text-gray-400" />
                  <span className="flex-1 text-sm">
                    {new Date(meeting.date).toLocaleDateString('ru-RU')}
                  </span>
                  {meeting.status === 'completed' ? (
                    <CheckCircle size={14} className="text-green-500" />
                  ) : (
                    <AlertCircle size={14} className="text-yellow-500" />
                  )}
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Статистика */}
      <div className="bg-white rounded-lg shadow-sm border p-6">
        <h2 className="font-semibold mb-4">Статистика</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <p className="text-2xl font-bold text-blue-600">{tasks.length}</p>
            <p className="text-sm text-gray-500">Всего задач</p>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <p className="text-2xl font-bold text-green-600">{completedTasks.length}</p>
            <p className="text-sm text-gray-500">Выполнено</p>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <p className="text-2xl font-bold text-yellow-600">{activeTasks.length}</p>
            <p className="text-sm text-gray-500">В работе</p>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <p className="text-2xl font-bold text-purple-600">{meetings.length}</p>
            <p className="text-sm text-gray-500">Встреч</p>
          </div>
        </div>
      </div>
    </div>
  )
}
