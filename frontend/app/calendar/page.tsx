'use client'

import { useState, useEffect, useMemo } from 'react'
import {
  ChevronLeft,
  ChevronRight,
  Plus,
  RefreshCw,
  Clock,
  Users,
  MapPin,
  X,
  Check,
  Loader2,
  AlertCircle
} from 'lucide-react'

import { API_URL } from '@/lib/config'
import { useAuth } from '@/lib/auth'

interface CalendarEvent {
  id: string
  subject: string
  body: string | null
  start: string
  end: string
  location: string | null
  organizer: string | null
  attendees: {
    email: string
    name: string | null
    response: string | null
    optional?: boolean
  }[]
  is_recurring: boolean
  is_cancelled: boolean
}

interface Employee {
  id: string
  name: string
  email: string
  position: string
}

export default function CalendarPage() {
  const { user } = useAuth()
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [currentDate, setCurrentDate] = useState(new Date())
  const [view, setView] = useState<'week' | 'day'>('week')
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)
  const [showNewMeeting, setShowNewMeeting] = useState(false)
  const [selectedSlot, setSelectedSlot] = useState<{ date: Date; hour: number } | null>(null)
  const [connectorStatus, setConnectorStatus] = useState<boolean>(false)

  useEffect(() => {
    checkConnectorStatus()
    fetchEmployees()
  }, [])

  useEffect(() => {
    if (user?.id) {
      fetchCalendar()
    }
  }, [user?.id, currentDate])

  const checkConnectorStatus = async () => {
    try {
      const res = await fetch(`${API_URL}/connector/status`)
      if (res.ok) {
        const data = await res.json()
        setConnectorStatus(data.connected)
      }
    } catch (error) {
      console.error('Failed to check connector status:', error)
    }
  }

  const fetchEmployees = async () => {
    try {
      const res = await fetch(`${API_URL}/employees`)
      if (res.ok) {
        const data = await res.json()
        setEmployees(data)
      }
    } catch (error) {
      console.error('Failed to fetch employees:', error)
    }
  }

  const fetchCalendar = async () => {
    if (!user?.id) return

    setLoading(true)
    try {
      const res = await fetch(`${API_URL}/calendar/${user.id}?days_back=7&days_forward=30`)
      if (res.ok) {
        const data = await res.json()
        setEvents(Array.isArray(data) ? data : [])
      }
    } catch (error) {
      console.error('Failed to fetch calendar:', error)
    } finally {
      setLoading(false)
    }
  }

  const syncCalendar = async () => {
    if (!user?.id) return

    setSyncing(true)
    try {
      const formData = new FormData()
      formData.append('employee_id', user.id)

      const res = await fetch(`${API_URL}/calendar/sync`, {
        method: 'POST',
        body: formData
      })

      if (res.ok) {
        const data = await res.json()
        alert(`Синхронизировано ${data.synced} встреч из ${data.total_events}`)
        fetchCalendar()
      } else {
        const error = await res.json()
        alert(`Ошибка: ${error.detail || 'Не удалось синхронизировать'}`)
      }
    } catch (error) {
      console.error('Failed to sync calendar:', error)
      alert('Ошибка при синхронизации календаря')
    } finally {
      setSyncing(false)
    }
  }

  // Get week days
  const weekDays = useMemo(() => {
    const start = new Date(currentDate)
    start.setDate(start.getDate() - start.getDay() + 1) // Start from Monday

    return Array.from({ length: 7 }, (_, i) => {
      const day = new Date(start)
      day.setDate(start.getDate() + i)
      return day
    })
  }, [currentDate])

  // Hours for the time grid
  const hours = Array.from({ length: 24 }, (_, i) => i)
  const workingHours = hours.filter(h => h >= 8 && h <= 20)

  // Navigate
  const goToday = () => setCurrentDate(new Date())
  const goPrev = () => {
    const newDate = new Date(currentDate)
    newDate.setDate(newDate.getDate() - (view === 'week' ? 7 : 1))
    setCurrentDate(newDate)
  }
  const goNext = () => {
    const newDate = new Date(currentDate)
    newDate.setDate(newDate.getDate() + (view === 'week' ? 7 : 1))
    setCurrentDate(newDate)
  }

  // Get events for a specific day and hour
  const getEventsForSlot = (date: Date, hour: number) => {
    return events.filter(event => {
      if (!event.start) return false
      const eventStart = new Date(event.start)
      const eventDate = eventStart.toDateString()
      const eventHour = eventStart.getHours()
      return eventDate === date.toDateString() && eventHour === hour
    })
  }

  // Get event duration in hours
  const getEventDuration = (event: CalendarEvent) => {
    if (!event.start || !event.end) return 1
    const start = new Date(event.start)
    const end = new Date(event.end)
    return Math.max(1, (end.getTime() - start.getTime()) / (1000 * 60 * 60))
  }

  // Format date for header
  const formatWeekRange = () => {
    const start = weekDays[0]
    const end = weekDays[6]
    const options: Intl.DateTimeFormatOptions = { day: 'numeric', month: 'long' }

    if (start.getMonth() === end.getMonth()) {
      return `${start.getDate()} - ${end.toLocaleDateString('ru-RU', options)} ${end.getFullYear()}`
    }
    return `${start.toLocaleDateString('ru-RU', options)} - ${end.toLocaleDateString('ru-RU', options)} ${end.getFullYear()}`
  }

  const isToday = (date: Date) => {
    const today = new Date()
    return date.toDateString() === today.toDateString()
  }

  // Handle slot click for creating new meeting
  const handleSlotClick = (date: Date, hour: number) => {
    setSelectedSlot({ date, hour })
    setShowNewMeeting(true)
  }

  if (!user) {
    return (
      <div className="flex flex-col items-center justify-center h-96">
        <Loader2 size={32} className="animate-spin text-ekf-orange mb-4" />
        <p className="text-ekf-gray">Загрузка...</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-bold text-ekf-dark">Календарь</h1>
          <div className="flex items-center gap-1">
            <button
              onClick={goPrev}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <ChevronLeft size={20} className="text-ekf-dark" />
            </button>
            <button
              onClick={goToday}
              className="px-3 py-1 text-sm hover:bg-gray-100 rounded-lg text-ekf-dark transition-colors"
            >
              Сегодня
            </button>
            <button
              onClick={goNext}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <ChevronRight size={20} className="text-ekf-dark" />
            </button>
          </div>
          <span className="text-lg font-medium text-ekf-gray">
            {formatWeekRange()}
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* View toggle */}
          <div className="flex border border-gray-200 rounded-lg overflow-hidden">
            <button
              onClick={() => setView('day')}
              className={`px-3 py-1.5 text-sm font-medium transition-colors ${
                view === 'day'
                  ? 'bg-ekf-orange text-white'
                  : 'bg-white text-ekf-dark hover:bg-gray-50'
              }`}
            >
              День
            </button>
            <button
              onClick={() => setView('week')}
              className={`px-3 py-1.5 text-sm font-medium transition-colors ${
                view === 'week'
                  ? 'bg-ekf-orange text-white'
                  : 'bg-white text-ekf-dark hover:bg-gray-50'
              }`}
            >
              Неделя
            </button>
          </div>

          {/* Connector status */}
          <div className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded text-xs font-medium ${
            connectorStatus ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
          }`}>
            <div className={`w-2 h-2 rounded-full ${connectorStatus ? 'bg-green-500' : 'bg-red-500'}`} />
            {connectorStatus ? 'Exchange' : 'Нет связи'}
          </div>

          {/* Sync button */}
          <button
            onClick={syncCalendar}
            disabled={syncing || !connectorStatus}
            className="flex items-center gap-2 px-3 py-2 text-sm border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 transition-colors"
          >
            <RefreshCw size={16} className={`text-ekf-dark ${syncing ? 'animate-spin' : ''}`} />
            <span className="text-ekf-dark">Синхронизировать</span>
          </button>

          {/* New meeting button */}
          <button
            onClick={() => setShowNewMeeting(true)}
            className="flex items-center gap-2 px-4 py-2 bg-ekf-orange text-white rounded-lg hover:bg-ekf-orange-dark transition-colors"
          >
            <Plus size={16} />
            Новая встреча
          </button>
        </div>
      </div>

      {/* Info banner if no events */}
      {!loading && events.length === 0 && (
        <div className="bg-primary-50 border border-primary-100 rounded-lg p-4 mb-4 flex items-center gap-3">
          <AlertCircle size={20} className="text-ekf-orange" />
          <div>
            <p className="font-medium text-ekf-dark">Календарь пуст</p>
            <p className="text-sm text-ekf-gray">
              {connectorStatus
                ? 'Нажмите "Синхронизировать" чтобы загрузить встречи из Exchange'
                : 'Подключите on-prem коннектор для синхронизации с Exchange'
              }
            </p>
          </div>
        </div>
      )}

      {/* Calendar Grid */}
      {loading ? (
        <div className="flex items-center justify-center h-96">
          <Loader2 className="animate-spin text-ekf-orange" size={32} />
        </div>
      ) : (
        <div className="flex-1 bg-white rounded-lg shadow-sm border overflow-hidden">
          <div className="flex flex-col h-full">
            {/* Day headers */}
            <div className="flex border-b bg-gray-50">
              <div className="w-16 flex-shrink-0 border-r" /> {/* Time column */}
              {(view === 'week' ? weekDays : [currentDate]).map((day, i) => (
                <div
                  key={i}
                  className={`flex-1 p-2 text-center border-r last:border-r-0 ${
                    isToday(day) ? 'bg-primary-50' : ''
                  }`}
                >
                  <div className="text-xs text-ekf-gray">
                    {day.toLocaleDateString('ru-RU', { weekday: 'short' })}
                  </div>
                  <div className={`text-lg font-semibold ${
                    isToday(day) ? 'text-ekf-orange' : 'text-ekf-dark'
                  }`}>
                    {day.getDate()}
                  </div>
                </div>
              ))}
            </div>

            {/* Time grid */}
            <div className="flex-1 overflow-y-auto">
              {workingHours.map(hour => (
                <div key={hour} className="flex border-b min-h-[60px]">
                  {/* Time label */}
                  <div className="w-16 flex-shrink-0 border-r p-1 text-xs text-ekf-gray text-right pr-2">
                    {hour.toString().padStart(2, '0')}:00
                  </div>

                  {/* Day cells */}
                  {(view === 'week' ? weekDays : [currentDate]).map((day, dayIndex) => {
                    const slotEvents = getEventsForSlot(day, hour)

                    return (
                      <div
                        key={dayIndex}
                        className={`flex-1 border-r last:border-r-0 p-1 cursor-pointer hover:bg-gray-50 transition-colors ${
                          isToday(day) ? 'bg-primary-50/30' : ''
                        }`}
                        onClick={() => handleSlotClick(day, hour)}
                      >
                        {slotEvents.map((event, eventIndex) => (
                          <div
                            key={eventIndex}
                            className="bg-ekf-orange text-white text-xs p-1.5 rounded mb-1 cursor-pointer hover:bg-ekf-orange-dark transition-colors"
                            style={{
                              minHeight: `${Math.min(getEventDuration(event), 3) * 50}px`
                            }}
                            onClick={(e) => {
                              e.stopPropagation()
                              // TODO: Open event details
                            }}
                          >
                            <div className="font-medium truncate">{event.subject}</div>
                            <div className="opacity-75 truncate">
                              {new Date(event.start).toLocaleTimeString('ru-RU', {
                                hour: '2-digit',
                                minute: '2-digit'
                              })}
                              {event.location && ` - ${event.location}`}
                            </div>
                            {event.attendees && event.attendees.length > 0 && (
                              <div className="opacity-75 truncate text-[10px] mt-0.5">
                                <Users size={10} className="inline mr-1" />
                                {event.attendees.length} участников
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    )
                  })}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* New Meeting Modal */}
      {showNewMeeting && user && (
        <NewMeetingModal
          employees={employees}
          currentUserId={user.id}
          initialDate={selectedSlot?.date || new Date()}
          initialHour={selectedSlot?.hour || 10}
          connectorStatus={connectorStatus}
          onClose={() => {
            setShowNewMeeting(false)
            setSelectedSlot(null)
          }}
          onCreated={() => {
            setShowNewMeeting(false)
            setSelectedSlot(null)
            fetchCalendar()
          }}
        />
      )}
    </div>
  )
}

// New Meeting Modal Component
function NewMeetingModal({
  employees,
  currentUserId,
  initialDate,
  initialHour,
  connectorStatus,
  onClose,
  onCreated
}: {
  employees: Employee[]
  currentUserId: string
  initialDate: Date
  initialHour: number
  connectorStatus: boolean
  onClose: () => void
  onCreated: () => void
}) {
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')
  const [location, setLocation] = useState('')
  const [startDate, setStartDate] = useState(initialDate.toISOString().split('T')[0])
  const [startTime, setStartTime] = useState(`${initialHour.toString().padStart(2, '0')}:00`)
  const [endTime, setEndTime] = useState(`${(initialHour + 1).toString().padStart(2, '0')}:00`)
  const [selectedAttendees, setSelectedAttendees] = useState<string[]>([])
  const [creating, setCreating] = useState(false)
  const [findingSlots, setFindingSlots] = useState(false)
  const [freeSlots, setFreeSlots] = useState<{ start: string; end: string }[]>([])

  const handleCreate = async () => {
    if (!subject.trim()) {
      alert('Введите тему встречи')
      return
    }

    setCreating(true)
    try {
      const formData = new FormData()
      formData.append('organizer_id', currentUserId)
      formData.append('subject', subject)
      formData.append('start', `${startDate}T${startTime}:00`)
      formData.append('end', `${startDate}T${endTime}:00`)
      formData.append('attendee_ids', JSON.stringify(selectedAttendees))
      formData.append('body', body)
      formData.append('location', location)

      const res = await fetch(`${API_URL}/calendar/meeting`, {
        method: 'POST',
        body: formData
      })

      if (res.ok) {
        onCreated()
      } else {
        const error = await res.json()
        alert(`Ошибка: ${error.detail || 'Не удалось создать встречу'}`)
      }
    } catch (error) {
      console.error('Failed to create meeting:', error)
      alert('Ошибка при создании встречи')
    } finally {
      setCreating(false)
    }
  }

  const findFreeSlots = async () => {
    if (selectedAttendees.length === 0) {
      alert('Выберите участников')
      return
    }

    setFindingSlots(true)
    try {
      const startDateTime = new Date(startDate)
      startDateTime.setHours(8, 0, 0, 0)

      const endDateTime = new Date(startDate)
      endDateTime.setHours(20, 0, 0, 0)

      const params = new URLSearchParams({
        attendee_ids: [...selectedAttendees, currentUserId].join(','),
        duration_minutes: '60',
        start: startDateTime.toISOString(),
        end: endDateTime.toISOString()
      })

      const res = await fetch(`${API_URL}/calendar/free-slots?${params}`)
      if (res.ok) {
        const data = await res.json()
        setFreeSlots(data)
      }
    } catch (error) {
      console.error('Failed to find free slots:', error)
    } finally {
      setFindingSlots(false)
    }
  }

  const toggleAttendee = (id: string) => {
    setSelectedAttendees(prev =>
      prev.includes(id)
        ? prev.filter(a => a !== id)
        : [...prev, id]
    )
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <h2 className="text-lg font-semibold text-ekf-dark">Новая встреча</h2>
          <button onClick={onClose} className="p-1 hover:bg-gray-100 rounded transition-colors">
            <X size={20} className="text-ekf-gray" />
          </button>
        </div>

        {/* Form */}
        <div className="p-4 space-y-4">
          {/* Subject */}
          <div>
            <label className="block text-sm font-medium text-ekf-dark mb-1">
              Тема встречи *
            </label>
            <input
              type="text"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
              placeholder="Например: 1-на-1 с Иваном"
            />
          </div>

          {/* Date and Time */}
          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-ekf-dark mb-1">
                Дата
              </label>
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-ekf-dark mb-1">
                Начало
              </label>
              <input
                type="time"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-ekf-dark mb-1">
                Окончание
              </label>
              <input
                type="time"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
                className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
              />
            </div>
          </div>

          {/* Location */}
          <div>
            <label className="block text-sm font-medium text-ekf-dark mb-1">
              <MapPin size={14} className="inline mr-1" />
              Место
            </label>
            <input
              type="text"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
              placeholder="Переговорная или ссылка на Zoom"
            />
          </div>

          {/* Attendees */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="block text-sm font-medium text-ekf-dark">
                <Users size={14} className="inline mr-1" />
                Участники
              </label>
              {connectorStatus && selectedAttendees.length > 0 && (
                <button
                  onClick={findFreeSlots}
                  disabled={findingSlots}
                  className="text-sm text-ekf-orange hover:underline flex items-center gap-1"
                >
                  <Clock size={14} />
                  {findingSlots ? 'Поиск...' : 'Найти свободное время'}
                </button>
              )}
            </div>

            <div className="border border-gray-200 rounded-lg max-h-40 overflow-y-auto">
              {employees
                .filter(emp => emp.id !== currentUserId)
                .map(emp => (
                  <label
                    key={emp.id}
                    className="flex items-center gap-3 p-2 hover:bg-gray-50 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={selectedAttendees.includes(emp.id)}
                      onChange={() => toggleAttendee(emp.id)}
                      className="rounded text-ekf-orange focus:ring-ekf-orange"
                    />
                    <div>
                      <div className="font-medium text-sm text-ekf-dark">{emp.name}</div>
                      <div className="text-xs text-ekf-gray">{emp.position}</div>
                    </div>
                  </label>
                ))}
            </div>
          </div>

          {/* Free Slots */}
          {freeSlots.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-ekf-dark mb-2">
                Свободные слоты
              </label>
              <div className="flex flex-wrap gap-2">
                {freeSlots.slice(0, 5).map((slot, i) => (
                  <button
                    key={i}
                    onClick={() => {
                      const start = new Date(slot.start)
                      setStartTime(`${start.getHours().toString().padStart(2, '0')}:${start.getMinutes().toString().padStart(2, '0')}`)
                      const end = new Date(slot.start)
                      end.setHours(end.getHours() + 1)
                      setEndTime(`${end.getHours().toString().padStart(2, '0')}:${end.getMinutes().toString().padStart(2, '0')}`)
                    }}
                    className="px-3 py-1 text-sm border border-gray-200 rounded-lg hover:bg-primary-50 hover:border-ekf-orange transition-colors"
                  >
                    {new Date(slot.start).toLocaleTimeString('ru-RU', {
                      hour: '2-digit',
                      minute: '2-digit'
                    })}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-ekf-dark mb-1">
              Описание
            </label>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              rows={3}
              className="w-full px-3 py-2 border border-gray-200 rounded-lg resize-none focus:ring-2 focus:ring-ekf-orange focus:border-ekf-orange outline-none"
              placeholder="Повестка встречи..."
            />
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-4 border-t bg-gray-50">
          <div className="text-sm text-ekf-gray">
            {connectorStatus
              ? 'Встреча будет создана в Exchange и отправлены приглашения'
              : 'Exchange недоступен - встреча будет создана только в приложении'
            }
          </div>
          <div className="flex gap-2">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-gray-200 text-ekf-dark rounded-lg hover:bg-gray-100 transition-colors"
            >
              Отмена
            </button>
            <button
              onClick={handleCreate}
              disabled={creating || !subject.trim()}
              className="px-4 py-2 bg-ekf-orange text-white rounded-lg hover:bg-ekf-orange-dark disabled:opacity-50 flex items-center gap-2 transition-colors"
            >
              {creating ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  Создание...
                </>
              ) : (
                <>
                  <Check size={16} />
                  Создать встречу
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
