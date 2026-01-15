'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Users, Calendar, AlertTriangle, CheckCircle, Clock } from 'lucide-react'

import { API_URL } from '@/lib/config'

interface Employee {
  id: string
  name: string
  position: string
}

interface Meeting {
  id: string
  date: string
  mood_score: number
  summary: string
  employees: { name: string }
}

interface Agreement {
  id: string
  task: string
  deadline: string
  meetings: {
    employees: { name: string }
  }
}

interface RedFlag {
  employee: string
  date: string
  flags: {
    burnout_signs: boolean | string
    turnover_risk: string
  }
}

export default function Dashboard() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [recentMeetings, setRecentMeetings] = useState<Meeting[]>([])
  const [pendingAgreements, setPendingAgreements] = useState<Agreement[]>([])
  const [redFlags, setRedFlags] = useState<RedFlag[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchDashboard()
  }, [])

  const fetchDashboard = async () => {
    try {
      const res = await fetch(`${API_URL}/analytics/dashboard`)
      if (res.ok) {
        const data = await res.json()
        setEmployees(data.employees || [])
        setRecentMeetings(data.recent_meetings || [])
        setPendingAgreements(data.pending_agreements || [])
        setRedFlags(data.red_flags || [])
      }
    } catch (error) {
      console.error('Failed to fetch dashboard:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold text-gray-900">Дашборд</h1>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center gap-3">
            <Users className="text-blue-600" size={24} />
            <div>
              <p className="text-2xl font-bold">{employees.length}</p>
              <p className="text-gray-500 text-sm">Сотрудников</p>
            </div>
          </div>
        </div>
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center gap-3">
            <Calendar className="text-green-600" size={24} />
            <div>
              <p className="text-2xl font-bold">{recentMeetings.length}</p>
              <p className="text-gray-500 text-sm">Встреч за месяц</p>
            </div>
          </div>
        </div>
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center gap-3">
            <Clock className="text-yellow-600" size={24} />
            <div>
              <p className="text-2xl font-bold">{pendingAgreements.length}</p>
              <p className="text-gray-500 text-sm">Открытых задач</p>
            </div>
          </div>
        </div>
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center gap-3">
            <AlertTriangle className="text-red-600" size={24} />
            <div>
              <p className="text-2xl font-bold">{redFlags.length}</p>
              <p className="text-gray-500 text-sm">Красных флагов</p>
            </div>
          </div>
        </div>
      </div>

      {/* Red Flags */}
      {redFlags.length > 0 && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <h2 className="text-lg font-semibold text-red-800 mb-4 flex items-center gap-2">
            <AlertTriangle size={20} />
            Требуют внимания
          </h2>
          <div className="space-y-3">
            {redFlags.map((flag, i) => (
              <div key={i} className="bg-white p-4 rounded border border-red-200">
                <p className="font-medium">{flag.employee}</p>
                <p className="text-sm text-gray-600">Встреча: {flag.date}</p>
                {flag.flags.burnout_signs && (
                  <p className="text-sm text-red-600 mt-1">
                    Признаки выгорания: {typeof flag.flags.burnout_signs === 'string' 
                      ? flag.flags.burnout_signs 
                      : 'Да'}
                  </p>
                )}
                {flag.flags.turnover_risk !== 'low' && (
                  <p className="text-sm text-red-600">
                    Риск ухода: {flag.flags.turnover_risk === 'high' ? 'Высокий' : 'Средний'}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Recent Meetings */}
        <div className="bg-white rounded-lg shadow-sm border">
          <div className="p-4 border-b">
            <h2 className="text-lg font-semibold">Последние встречи</h2>
          </div>
          <div className="divide-y">
            {recentMeetings.slice(0, 5).map((meeting) => (
              <Link
                key={meeting.id}
                href={`/meetings/${meeting.id}`}
                className="block p-4 hover:bg-gray-50"
              >
                <div className="flex justify-between items-start">
                  <div>
                    <p className="font-medium">{meeting.employees?.name}</p>
                    <p className="text-sm text-gray-500">{meeting.date}</p>
                  </div>
                  {meeting.mood_score && (
                    <span className={`px-2 py-1 rounded text-sm ${
                      meeting.mood_score >= 7 ? 'bg-green-100 text-green-800' :
                      meeting.mood_score >= 5 ? 'bg-yellow-100 text-yellow-800' :
                      'bg-red-100 text-red-800'
                    }`}>
                      {meeting.mood_score}/10
                    </span>
                  )}
                </div>
                {meeting.summary && (
                  <p className="text-sm text-gray-600 mt-2 line-clamp-2">{meeting.summary}</p>
                )}
              </Link>
            ))}
            {recentMeetings.length === 0 && (
              <p className="p-4 text-gray-500 text-center">Нет встреч</p>
            )}
          </div>
        </div>

        {/* Pending Agreements */}
        <div className="bg-white rounded-lg shadow-sm border">
          <div className="p-4 border-b">
            <h2 className="text-lg font-semibold">Открытые договоренности</h2>
          </div>
          <div className="divide-y">
            {pendingAgreements.slice(0, 5).map((agreement) => (
              <div key={agreement.id} className="p-4">
                <div className="flex justify-between items-start">
                  <div>
                    <p className="font-medium">{agreement.task}</p>
                    <p className="text-sm text-gray-500">
                      {agreement.meetings?.employees?.name}
                    </p>
                  </div>
                  {agreement.deadline && (
                    <span className="text-sm text-gray-600">
                      до {agreement.deadline}
                    </span>
                  )}
                </div>
              </div>
            ))}
            {pendingAgreements.length === 0 && (
              <p className="p-4 text-gray-500 text-center">Нет открытых задач</p>
            )}
          </div>
        </div>
      </div>

      {/* Team */}
      <div className="bg-white rounded-lg shadow-sm border">
        <div className="p-4 border-b flex justify-between items-center">
          <h2 className="text-lg font-semibold">Команда</h2>
          <Link href="/employees" className="text-blue-600 text-sm hover:underline">
            Все сотрудники
          </Link>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-4">
          {employees.map((employee) => (
            <Link
              key={employee.id}
              href={`/employees/${employee.id}`}
              className="p-4 border rounded-lg hover:border-blue-300 transition-colors"
            >
              <p className="font-medium">{employee.name}</p>
              <p className="text-sm text-gray-500">{employee.position}</p>
            </Link>
          ))}
          {employees.length === 0 && (
            <p className="text-gray-500 col-span-3 text-center py-4">
              <Link href="/employees" className="text-blue-600 hover:underline">
                Добавьте первого сотрудника
              </Link>
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
