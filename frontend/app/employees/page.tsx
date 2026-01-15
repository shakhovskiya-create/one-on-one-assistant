'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Plus, Edit, User, Calendar, Users, Filter, Building } from 'lucide-react'

import { API_URL } from '@/lib/config'
import { useAuth } from '@/lib/auth'

interface Employee {
  id: string
  name: string
  position: string
  email: string | null
  department: string | null
  manager_id: string | null
  meeting_frequency: string
  meeting_day: string | null
  development_priorities: string | null
  level?: number
}

export default function EmployeesPage() {
  const { user, subordinates } = useAuth()
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null)
  const [filter, setFilter] = useState<'all' | 'my-team'>('my-team')
  const [formData, setFormData] = useState({
    name: '',
    position: '',
    meeting_frequency: 'weekly',
    meeting_day: '',
    development_priorities: '',
  })

  useEffect(() => {
    fetchEmployees()
  }, [filter, user])

  const fetchEmployees = async () => {
    setLoading(true)
    try {
      let url = `${API_URL}/employees`

      // If filter is 'my-team' and user is logged in, use filtered endpoint
      if (filter === 'my-team' && user) {
        url = `${API_URL}/employees/my-team?manager_id=${user.id}&include_indirect=true`
      }

      const res = await fetch(url)
      if (res.ok) {
        const data = await res.json()
        setEmployees(data)
      }
    } catch (error) {
      console.error('Failed to fetch employees:', error)
    } finally {
      setLoading(false)
    }
  }

  const openModal = (employee?: Employee) => {
    if (employee) {
      setEditingEmployee(employee)
      setFormData({
        name: employee.name,
        position: employee.position,
        meeting_frequency: employee.meeting_frequency,
        meeting_day: employee.meeting_day || '',
        development_priorities: employee.development_priorities || '',
      })
    } else {
      setEditingEmployee(null)
      setFormData({
        name: '',
        position: '',
        meeting_frequency: 'weekly',
        meeting_day: '',
        development_priorities: '',
      })
    }
    setShowModal(true)
  }

  const closeModal = () => {
    setShowModal(false)
    setEditingEmployee(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      const url = editingEmployee
        ? `${API_URL}/employees/${editingEmployee.id}`
        : `${API_URL}/employees`

      const res = await fetch(url, {
        method: editingEmployee ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })

      if (res.ok) {
        await fetchEmployees()
        closeModal()
      }
    } catch (error) {
      console.error('Failed to save employee:', error)
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

  // Group employees by department for better display
  const groupByDepartment = (emps: Employee[]) => {
    const groups: Record<string, Employee[]> = {}
    emps.forEach(emp => {
      const dept = emp.department || 'Без отдела'
      if (!groups[dept]) groups[dept] = []
      groups[dept].push(emp)
    })
    return groups
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  const groupedEmployees = groupByDepartment(employees)

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-bold text-gray-900">Команда</h1>
          {user && (
            <div className="flex border rounded-lg overflow-hidden">
              <button
                onClick={() => setFilter('my-team')}
                className={`flex items-center gap-1 px-3 py-1 text-sm ${
                  filter === 'my-team'
                    ? 'bg-blue-600 text-white'
                    : 'hover:bg-gray-100'
                }`}
              >
                <Users size={14} />
                Мои подчинённые
              </button>
              <button
                onClick={() => setFilter('all')}
                className={`flex items-center gap-1 px-3 py-1 text-sm ${
                  filter === 'all'
                    ? 'bg-blue-600 text-white'
                    : 'hover:bg-gray-100'
                }`}
              >
                <Building size={14} />
                Все
              </button>
            </div>
          )}
        </div>
        <button
          onClick={() => openModal()}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus size={20} />
          Добавить сотрудника
        </button>
      </div>

      {/* Stats */}
      {user && filter === 'my-team' && (
        <div className="bg-blue-50 border border-blue-100 rounded-lg p-4">
          <div className="flex items-center gap-2 text-blue-700">
            <Users size={20} />
            <span className="font-medium">
              {employees.length} {employees.length === 1 ? 'подчинённый' : 'подчинённых'}
            </span>
          </div>
        </div>
      )}

      {employees.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-12 text-center">
          <User size={48} className="mx-auto text-gray-400 mb-4" />
          <p className="text-gray-600 mb-4">
            {filter === 'my-team' ? 'У вас нет подчинённых' : 'Пока нет сотрудников'}
          </p>
          {filter === 'all' && (
            <button
              onClick={() => openModal()}
              className="text-blue-600 hover:underline"
            >
              Добавить первого сотрудника
            </button>
          )}
        </div>
      ) : (
        <div className="space-y-6">
          {Object.entries(groupedEmployees).map(([department, deptEmployees]) => (
            <div key={department}>
              <h2 className="text-sm font-medium text-gray-500 mb-3 flex items-center gap-2">
                <Building size={14} />
                {department}
                <span className="text-gray-400">({deptEmployees.length})</span>
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {deptEmployees.map((employee) => (
                  <div
                    key={employee.id}
                    className="bg-white rounded-lg shadow-sm border p-6 hover:shadow-md transition-shadow"
                  >
                    <div className="flex justify-between items-start mb-4">
                      <div>
                        <h3 className="font-semibold text-lg">{employee.name}</h3>
                        <p className="text-gray-500">{employee.position}</p>
                        {employee.email && (
                          <p className="text-xs text-gray-400 mt-1">{employee.email}</p>
                        )}
                      </div>
                      <button
                        onClick={() => openModal(employee)}
                        className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded"
                      >
                        <Edit size={16} />
                      </button>
                    </div>

                    {/* Level indicator for subordinates */}
                    {filter === 'my-team' && employee.level && (
                      <div className="mb-3">
                        <span className={`text-xs px-2 py-0.5 rounded ${
                          employee.level === 1
                            ? 'bg-green-100 text-green-700'
                            : 'bg-gray-100 text-gray-600'
                        }`}>
                          {employee.level === 1 ? 'Прямой подчинённый' : `Уровень ${employee.level}`}
                        </span>
                      </div>
                    )}

                    <div className="space-y-2 text-sm">
                      <div className="flex items-center gap-2 text-gray-600">
                        <Calendar size={16} />
                        <span>{frequencyLabels[employee.meeting_frequency] || 'Не задано'}</span>
                        {employee.meeting_day && (
                          <span>({dayLabels[employee.meeting_day]})</span>
                        )}
                      </div>
                      {employee.development_priorities && (
                        <p className="text-gray-500 line-clamp-2">
                          {employee.development_priorities}
                        </p>
                      )}
                    </div>

                    <div className="mt-4 pt-4 border-t flex gap-2">
                      <Link
                        href={`/employees/${employee.id}`}
                        className="flex-1 text-center py-2 text-blue-600 hover:bg-blue-50 rounded text-sm"
                      >
                        Досье
                      </Link>
                      <Link
                        href={`/analytics?employee=${employee.id}`}
                        className="flex-1 text-center py-2 text-blue-600 hover:bg-blue-50 rounded text-sm"
                      >
                        Аналитика
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
            <div className="p-6 border-b">
              <h2 className="text-lg font-semibold">
                {editingEmployee ? 'Редактировать сотрудника' : 'Новый сотрудник'}
              </h2>
            </div>
            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ФИО
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full border rounded-lg px-3 py-2"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Должность
                </label>
                <input
                  type="text"
                  value={formData.position}
                  onChange={(e) => setFormData({ ...formData, position: e.target.value })}
                  className="w-full border rounded-lg px-3 py-2"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Частота встреч
                  </label>
                  <select
                    value={formData.meeting_frequency}
                    onChange={(e) =>
                      setFormData({ ...formData, meeting_frequency: e.target.value })
                    }
                    className="w-full border rounded-lg px-3 py-2"
                  >
                    <option value="weekly">Еженедельно</option>
                    <option value="biweekly">Раз в 2 недели</option>
                    <option value="monthly">Ежемесячно</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    День встречи
                  </label>
                  <select
                    value={formData.meeting_day}
                    onChange={(e) =>
                      setFormData({ ...formData, meeting_day: e.target.value })
                    }
                    className="w-full border rounded-lg px-3 py-2"
                  >
                    <option value="">Не задан</option>
                    <option value="monday">Понедельник</option>
                    <option value="tuesday">Вторник</option>
                    <option value="wednesday">Среда</option>
                    <option value="thursday">Четверг</option>
                    <option value="friday">Пятница</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Приоритеты развития
                </label>
                <textarea
                  value={formData.development_priorities}
                  onChange={(e) =>
                    setFormData({ ...formData, development_priorities: e.target.value })
                  }
                  className="w-full border rounded-lg px-3 py-2"
                  rows={3}
                />
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={closeModal}
                  className="flex-1 py-2 border rounded-lg hover:bg-gray-50"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                >
                  {editingEmployee ? 'Сохранить' : 'Добавить'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
