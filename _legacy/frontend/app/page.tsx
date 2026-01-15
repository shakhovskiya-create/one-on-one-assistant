'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Users, Calendar, AlertTriangle, Clock, ChevronRight } from 'lucide-react'

import { API_URL } from '@/lib/config'
import { useAuth } from '@/lib/auth'

interface Meeting {
  id: string
  title: string
  date: string
  mood_score: number
  summary: string
  employees: { name: string }
}

interface Task {
  id: string
  title: string
  status: string
  due_date: string
  employee_id: string
}

export default function Dashboard() {
  const { user, subordinates } = useAuth()
  const [recentMeetings, setRecentMeetings] = useState<Meeting[]>([])
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (user) {
      fetchData()
    } else {
      setLoading(false)
    }
  }, [user])

  const fetchData = async () => {
    try {
      // Fetch meetings
      const meetingsRes = await fetch(`${API_URL}/meetings?limit=10`)
      if (meetingsRes.ok) {
        const data = await meetingsRes.json()
        setRecentMeetings(data || [])
      }

      // Fetch tasks
      const tasksRes = await fetch(`${API_URL}/tasks?status=pending&status=in_progress`)
      if (tasksRes.ok) {
        const data = await tasksRes.json()
        setTasks(data || [])
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-ekf-orange"></div>
      </div>
    )
  }

  // Filter subordinates to only show those with departments
  const employeesWithDept = subordinates.filter(emp => emp.department)

  const pendingTasks = tasks.filter(t => t.status === 'pending' || t.status === 'in_progress')
  const overdueTasks = tasks.filter(t => t.due_date && new Date(t.due_date) < new Date() && t.status !== 'done')

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-ekf-dark">Добро пожаловать{user ? `, ${user.name.split(' ')[0]}` : ''}!</h1>
          <p className="text-ekf-gray">Обзор вашей команды и задач</p>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Link href="/employees" className="bg-white p-6 rounded-lg shadow-sm border hover:border-ekf-orange transition-colors">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-primary-50 rounded-lg flex items-center justify-center">
              <Users className="text-ekf-orange" size={24} />
            </div>
            <div>
              <p className="text-2xl font-bold text-ekf-dark">{subordinates.length}</p>
              <p className="text-ekf-gray text-sm">Подчинённых</p>
            </div>
          </div>
        </Link>

        <Link href="/calendar" className="bg-white p-6 rounded-lg shadow-sm border hover:border-ekf-orange transition-colors">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-green-50 rounded-lg flex items-center justify-center">
              <Calendar className="text-green-600" size={24} />
            </div>
            <div>
              <p className="text-2xl font-bold text-ekf-dark">{recentMeetings.length}</p>
              <p className="text-ekf-gray text-sm">Встреч</p>
            </div>
          </div>
        </Link>

        <Link href="/tasks" className="bg-white p-6 rounded-lg shadow-sm border hover:border-ekf-orange transition-colors">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-yellow-50 rounded-lg flex items-center justify-center">
              <Clock className="text-yellow-600" size={24} />
            </div>
            <div>
              <p className="text-2xl font-bold text-ekf-dark">{pendingTasks.length}</p>
              <p className="text-ekf-gray text-sm">Открытых задач</p>
            </div>
          </div>
        </Link>

        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-red-50 rounded-lg flex items-center justify-center">
              <AlertTriangle className="text-red-600" size={24} />
            </div>
            <div>
              <p className="text-2xl font-bold text-ekf-dark">{overdueTasks.length}</p>
              <p className="text-ekf-gray text-sm">Просрочено</p>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Meetings */}
        <div className="bg-white rounded-lg shadow-sm border">
          <div className="p-4 border-b flex items-center justify-between">
            <h2 className="text-lg font-semibold text-ekf-dark">Последние встречи</h2>
            <Link href="/meetings" className="text-ekf-orange text-sm hover:underline flex items-center gap-1">
              Все <ChevronRight size={14} />
            </Link>
          </div>
          <div className="divide-y">
            {recentMeetings.slice(0, 5).map((meeting) => (
              <Link
                key={meeting.id}
                href={`/meetings/${meeting.id}`}
                className="block p-4 hover:bg-gray-50 transition-colors"
              >
                <div className="flex justify-between items-start">
                  <div>
                    <p className="font-medium text-ekf-dark">{meeting.title || meeting.employees?.name}</p>
                    <p className="text-sm text-ekf-gray">{meeting.date}</p>
                  </div>
                  {meeting.mood_score && (
                    <span className={`px-2 py-1 rounded text-sm font-medium ${
                      meeting.mood_score >= 7 ? 'bg-green-100 text-green-700' :
                      meeting.mood_score >= 5 ? 'bg-yellow-100 text-yellow-700' :
                      'bg-red-100 text-red-700'
                    }`}>
                      {meeting.mood_score}/10
                    </span>
                  )}
                </div>
                {meeting.summary && (
                  <p className="text-sm text-ekf-gray mt-2 line-clamp-2">{meeting.summary}</p>
                )}
              </Link>
            ))}
            {recentMeetings.length === 0 && (
              <div className="p-8 text-center">
                <Calendar size={32} className="mx-auto text-ekf-gray-light mb-2" />
                <p className="text-ekf-gray">Нет встреч</p>
                <Link href="/calendar" className="text-ekf-orange text-sm hover:underline">
                  Синхронизировать календарь
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Team */}
        <div className="bg-white rounded-lg shadow-sm border">
          <div className="p-4 border-b flex items-center justify-between">
            <h2 className="text-lg font-semibold text-ekf-dark">Команда</h2>
            <Link href="/employees" className="text-ekf-orange text-sm hover:underline flex items-center gap-1">
              Все <ChevronRight size={14} />
            </Link>
          </div>
          <div className="divide-y">
            {employeesWithDept.slice(0, 6).map((employee) => (
              <Link
                key={employee.id}
                href={`/employees/${employee.id}`}
                className="flex items-center gap-3 p-4 hover:bg-gray-50 transition-colors"
              >
                <div className="w-10 h-10 bg-primary-50 rounded-full flex items-center justify-center">
                  <Users size={18} className="text-ekf-orange" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-ekf-dark truncate">{employee.name}</p>
                  <p className="text-sm text-ekf-gray truncate">{employee.position}</p>
                </div>
                <ChevronRight size={16} className="text-ekf-gray-light" />
              </Link>
            ))}
            {employeesWithDept.length === 0 && subordinates.length === 0 && (
              <div className="p-8 text-center">
                <Users size={32} className="mx-auto text-ekf-gray-light mb-2" />
                <p className="text-ekf-gray">Нет подчинённых</p>
              </div>
            )}
            {employeesWithDept.length === 0 && subordinates.length > 0 && (
              <div className="p-8 text-center">
                <Users size={32} className="mx-auto text-ekf-gray-light mb-2" />
                <p className="text-ekf-gray">{subordinates.length} сотрудников без департамента</p>
                <Link href="/employees?filter=all" className="text-ekf-orange text-sm hover:underline">
                  Показать всех
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-lg shadow-sm border p-6">
        <h2 className="text-lg font-semibold text-ekf-dark mb-4">Быстрые действия</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Link
            href="/calendar"
            className="p-4 border border-gray-200 rounded-lg hover:border-ekf-orange hover:bg-primary-50 transition-colors text-center"
          >
            <Calendar size={24} className="mx-auto text-ekf-orange mb-2" />
            <p className="text-sm font-medium text-ekf-dark">Календарь</p>
          </Link>
          <Link
            href="/script"
            className="p-4 border border-gray-200 rounded-lg hover:border-ekf-orange hover:bg-primary-50 transition-colors text-center"
          >
            <Clock size={24} className="mx-auto text-ekf-orange mb-2" />
            <p className="text-sm font-medium text-ekf-dark">Скрипт встречи</p>
          </Link>
          <Link
            href="/upload"
            className="p-4 border border-gray-200 rounded-lg hover:border-ekf-orange hover:bg-primary-50 transition-colors text-center"
          >
            <AlertTriangle size={24} className="mx-auto text-ekf-orange mb-2" />
            <p className="text-sm font-medium text-ekf-dark">Загрузить запись</p>
          </Link>
          <Link
            href="/analytics"
            className="p-4 border border-gray-200 rounded-lg hover:border-ekf-orange hover:bg-primary-50 transition-colors text-center"
          >
            <Users size={24} className="mx-auto text-ekf-orange mb-2" />
            <p className="text-sm font-medium text-ekf-dark">Аналитика</p>
          </Link>
        </div>
      </div>
    </div>
  )
}
