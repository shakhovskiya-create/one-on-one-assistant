'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Calendar, User, ChevronRight } from 'lucide-react'

const API_URL = process.env.API_URL || 'http://localhost:8000'

interface Meeting {
  id: string
  date: string
  mood_score: number | null
  summary: string | null
  employees: {
    name: string
    position: string
  }
}

export default function MeetingsPage() {
  const [meetings, setMeetings] = useState<Meeting[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchMeetings()
  }, [])

  const fetchMeetings = async () => {
    try {
      const res = await fetch(`${API_URL}/meetings`)
      if (res.ok) {
        const data = await res.json()
        setMeetings(data)
      }
    } catch (error) {
      console.error('Failed to fetch meetings:', error)
    } finally {
      setLoading(false)
    }
  }

  const groupByMonth = (meetings: Meeting[]) => {
    const groups: Record<string, Meeting[]> = {}
    
    meetings.forEach((meeting) => {
      const date = new Date(meeting.date)
      const key = date.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })
      if (!groups[key]) {
        groups[key] = []
      }
      groups[key].push(meeting)
    })
    
    return groups
  }

  const groupedMeetings = groupByMonth(meetings)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Встречи</h1>
        <Link
          href="/upload"
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Загрузить запись
        </Link>
      </div>

      {meetings.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-12 text-center">
          <Calendar size={48} className="mx-auto text-gray-400 mb-4" />
          <p className="text-gray-600 mb-4">Пока нет записанных встреч</p>
          <Link href="/upload" className="text-blue-600 hover:underline">
            Загрузить первую запись
          </Link>
        </div>
      ) : (
        <div className="space-y-8">
          {Object.entries(groupedMeetings).map(([month, monthMeetings]) => (
            <div key={month}>
              <h2 className="text-lg font-semibold text-gray-700 mb-4 capitalize">
                {month}
              </h2>
              <div className="bg-white rounded-lg shadow-sm border divide-y">
                {monthMeetings.map((meeting) => (
                  <Link
                    key={meeting.id}
                    href={`/meetings/${meeting.id}`}
                    className="flex items-center justify-between p-4 hover:bg-gray-50"
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
                        <User className="text-blue-600" size={24} />
                      </div>
                      <div>
                        <p className="font-medium">{meeting.employees?.name}</p>
                        <p className="text-sm text-gray-500">
                          {new Date(meeting.date).toLocaleDateString('ru-RU', {
                            weekday: 'long',
                            day: 'numeric',
                            month: 'long',
                          })}
                        </p>
                        {meeting.summary && (
                          <p className="text-sm text-gray-600 mt-1 line-clamp-1 max-w-md">
                            {meeting.summary}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      {meeting.mood_score && (
                        <span
                          className={`px-3 py-1 rounded-full text-sm font-medium ${
                            meeting.mood_score >= 7
                              ? 'bg-green-100 text-green-800'
                              : meeting.mood_score >= 5
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {meeting.mood_score}/10
                        </span>
                      )}
                      <ChevronRight className="text-gray-400" size={20} />
                    </div>
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
