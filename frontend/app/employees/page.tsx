'use client'

import { useState, useEffect, useMemo } from 'react'
import Link from 'next/link'
import {
  Plus,
  Edit,
  User,
  Calendar,
  Users,
  Building,
  ChevronDown,
  ChevronRight,
  LayoutGrid,
  GitBranch,
  List,
  Mail,
  Phone
} from 'lucide-react'

import { API_URL } from '@/lib/config'
import { useAuth } from '@/lib/auth'

interface Employee {
  id: string
  name: string
  position: string
  email: string | null
  department: string | null
  manager_id: string | null
  phone: string | null
  mobile: string | null
  photo_base64: string | null
  meeting_frequency: string
  meeting_day: string | null
  development_priorities: string | null
  level?: number
}

type ViewMode = 'tiles' | 'tree' | 'list'

export default function EmployeesPage() {
  const { user } = useAuth()
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null)
  const [filter, setFilter] = useState<'all' | 'my-team'>('my-team')
  const [viewMode, setViewMode] = useState<ViewMode>('tree')
  const [expandedDepts, setExpandedDepts] = useState<Set<string>>(new Set())
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set(['root']))
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

  // Build org tree structure
  const orgTree = useMemo(() => {
    const employeeMap = new Map(employees.map(e => [e.id, e]))
    const children = new Map<string | null, Employee[]>()

    employees.forEach(emp => {
      const managerId = emp.manager_id
      if (!children.has(managerId)) {
        children.set(managerId, [])
      }
      children.get(managerId)!.push(emp)
    })

    // Find root employees (no manager or manager not in list)
    const roots = employees.filter(emp =>
      !emp.manager_id || !employeeMap.has(emp.manager_id)
    )

    return { roots, children }
  }, [employees])

  // Group by department
  const groupedByDepartment = useMemo(() => {
    const groups: Record<string, Employee[]> = {}
    employees.forEach(emp => {
      const dept = emp.department || 'Без отдела'
      if (!groups[dept]) groups[dept] = []
      groups[dept].push(emp)
    })
    return groups
  }, [employees])

  const toggleDept = (dept: string) => {
    const newExpanded = new Set(expandedDepts)
    if (newExpanded.has(dept)) {
      newExpanded.delete(dept)
    } else {
      newExpanded.add(dept)
    }
    setExpandedDepts(newExpanded)
  }

  const toggleNode = (id: string) => {
    const newExpanded = new Set(expandedNodes)
    if (newExpanded.has(id)) {
      newExpanded.delete(id)
    } else {
      newExpanded.add(id)
    }
    setExpandedNodes(newExpanded)
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

  // Render tree node recursively
  const renderTreeNode = (employee: Employee, depth: number = 0) => {
    const children = orgTree.children.get(employee.id) || []
    const hasChildren = children.length > 0
    const isExpanded = expandedNodes.has(employee.id)

    return (
      <div key={employee.id} className="select-none">
        <div
          className={`flex items-center gap-2 py-2 px-3 rounded-lg hover:bg-gray-50 cursor-pointer ${
            depth === 0 ? 'bg-blue-50' : ''
          }`}
          style={{ marginLeft: depth * 24 }}
        >
          {/* Expand/collapse button */}
          <button
            onClick={() => toggleNode(employee.id)}
            className={`p-0.5 rounded ${hasChildren ? 'hover:bg-gray-200' : 'invisible'}`}
          >
            {isExpanded ? (
              <ChevronDown size={16} className="text-gray-500" />
            ) : (
              <ChevronRight size={16} className="text-gray-500" />
            )}
          </button>

          {/* Avatar */}
          <div className="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center flex-shrink-0 overflow-hidden">
            {employee.photo_base64 ? (
              <img
                src={`data:image/jpeg;base64,${employee.photo_base64}`}
                alt=""
                className="w-full h-full object-cover"
              />
            ) : (
              <User size={16} className="text-gray-500" />
            )}
          </div>

          {/* Info */}
          <Link
            href={`/employees/${employee.id}`}
            className="flex-1 min-w-0"
            onClick={(e) => e.stopPropagation()}
          >
            <p className="font-medium text-sm truncate">{employee.name}</p>
            <p className="text-xs text-gray-500 truncate">{employee.position}</p>
          </Link>

          {/* Subordinates count */}
          {hasChildren && (
            <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded">
              {children.length}
            </span>
          )}

          {/* Edit button */}
          <button
            onClick={(e) => {
              e.stopPropagation()
              openModal(employee)
            }}
            className="p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded opacity-0 group-hover:opacity-100"
          >
            <Edit size={14} />
          </button>
        </div>

        {/* Children */}
        {isExpanded && hasChildren && (
          <div className="border-l border-gray-200 ml-6">
            {children.map(child => renderTreeNode(child, depth + 1))}
          </div>
        )}
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-bold text-gray-900">Команда</h1>

          {/* Filter toggle */}
          {user && (
            <div className="flex border rounded-lg overflow-hidden">
              <button
                onClick={() => setFilter('my-team')}
                className={`flex items-center gap-1 px-3 py-1 text-sm ${
                  filter === 'my-team' ? 'bg-blue-600 text-white' : 'hover:bg-gray-100'
                }`}
              >
                <Users size={14} />
                Мои
              </button>
              <button
                onClick={() => setFilter('all')}
                className={`flex items-center gap-1 px-3 py-1 text-sm ${
                  filter === 'all' ? 'bg-blue-600 text-white' : 'hover:bg-gray-100'
                }`}
              >
                <Building size={14} />
                Все
              </button>
            </div>
          )}

          {/* View mode toggle */}
          <div className="flex border rounded-lg overflow-hidden">
            <button
              onClick={() => setViewMode('tree')}
              className={`p-2 ${viewMode === 'tree' ? 'bg-gray-100' : 'hover:bg-gray-50'}`}
              title="Оргструктура"
            >
              <GitBranch size={16} />
            </button>
            <button
              onClick={() => setViewMode('tiles')}
              className={`p-2 ${viewMode === 'tiles' ? 'bg-gray-100' : 'hover:bg-gray-50'}`}
              title="Плитки"
            >
              <LayoutGrid size={16} />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`p-2 ${viewMode === 'list' ? 'bg-gray-100' : 'hover:bg-gray-50'}`}
              title="Список"
            >
              <List size={16} />
            </button>
          </div>
        </div>

        <button
          onClick={() => openModal()}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus size={20} />
          Добавить
        </button>
      </div>

      {/* Stats */}
      <div className="bg-blue-50 border border-blue-100 rounded-lg p-4">
        <div className="flex items-center gap-2 text-blue-700">
          <Users size={20} />
          <span className="font-medium">{employees.length} сотрудников</span>
          {filter === 'my-team' && user && (
            <span className="text-blue-500">в вашем подчинении</span>
          )}
        </div>
      </div>

      {employees.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border p-12 text-center">
          <User size={48} className="mx-auto text-gray-400 mb-4" />
          <p className="text-gray-600 mb-4">
            {filter === 'my-team' ? 'У вас нет подчинённых' : 'Пока нет сотрудников'}
          </p>
        </div>
      ) : (
        <>
          {/* Tree View */}
          {viewMode === 'tree' && (
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="space-y-1">
                {orgTree.roots.map(root => renderTreeNode(root, 0))}
              </div>
            </div>
          )}

          {/* Tiles View */}
          {viewMode === 'tiles' && (
            <div className="space-y-6">
              {Object.entries(groupedByDepartment).map(([dept, deptEmployees]) => (
                <div key={dept}>
                  <button
                    onClick={() => toggleDept(dept)}
                    className="flex items-center gap-2 text-sm font-medium text-gray-500 mb-3 hover:text-gray-700"
                  >
                    {expandedDepts.has(dept) ? (
                      <ChevronDown size={16} />
                    ) : (
                      <ChevronRight size={16} />
                    )}
                    <Building size={14} />
                    {dept}
                    <span className="text-gray-400">({deptEmployees.length})</span>
                  </button>

                  {expandedDepts.has(dept) && (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {deptEmployees.map((employee) => (
                        <div
                          key={employee.id}
                          className="bg-white rounded-lg shadow-sm border p-4 hover:shadow-md transition-shadow"
                        >
                          <div className="flex items-start gap-3">
                            <div className="w-12 h-12 bg-gray-200 rounded-full flex items-center justify-center flex-shrink-0 overflow-hidden">
                              {employee.photo_base64 ? (
                                <img
                                  src={`data:image/jpeg;base64,${employee.photo_base64}`}
                                  alt=""
                                  className="w-full h-full object-cover"
                                />
                              ) : (
                                <User size={24} className="text-gray-500" />
                              )}
                            </div>
                            <div className="flex-1 min-w-0">
                              <h3 className="font-semibold truncate">{employee.name}</h3>
                              <p className="text-sm text-gray-500 truncate">{employee.position}</p>
                              {employee.email && (
                                <p className="text-xs text-gray-400 truncate">{employee.email}</p>
                              )}
                            </div>
                            <button
                              onClick={() => openModal(employee)}
                              className="p-1 text-gray-400 hover:text-gray-600"
                            >
                              <Edit size={14} />
                            </button>
                          </div>
                          <div className="mt-3 pt-3 border-t flex gap-2">
                            <Link
                              href={`/employees/${employee.id}`}
                              className="flex-1 text-center py-1.5 text-blue-600 hover:bg-blue-50 rounded text-sm"
                            >
                              Досье
                            </Link>
                            <Link
                              href={`/analytics?employee=${employee.id}`}
                              className="flex-1 text-center py-1.5 text-blue-600 hover:bg-blue-50 rounded text-sm"
                            >
                              Аналитика
                            </Link>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* List View */}
          {viewMode === 'list' && (
            <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
              <table className="w-full">
                <thead className="bg-gray-50 border-b">
                  <tr>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Сотрудник</th>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Должность</th>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Отдел</th>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Контакты</th>
                    <th className="w-20"></th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {employees.map((employee) => (
                    <tr key={employee.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center overflow-hidden">
                            {employee.photo_base64 ? (
                              <img
                                src={`data:image/jpeg;base64,${employee.photo_base64}`}
                                alt=""
                                className="w-full h-full object-cover"
                              />
                            ) : (
                              <User size={16} className="text-gray-500" />
                            )}
                          </div>
                          <Link
                            href={`/employees/${employee.id}`}
                            className="font-medium text-blue-600 hover:underline"
                          >
                            {employee.name}
                          </Link>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600">{employee.position}</td>
                      <td className="px-4 py-3 text-sm text-gray-600">{employee.department || '-'}</td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3 text-sm text-gray-500">
                          {employee.email && (
                            <span className="flex items-center gap-1">
                              <Mail size={12} />
                              <span className="truncate max-w-32">{employee.email}</span>
                            </span>
                          )}
                          {(employee.phone || employee.mobile) && (
                            <span className="flex items-center gap-1">
                              <Phone size={12} />
                              {employee.mobile || employee.phone}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={() => openModal(employee)}
                          className="p-1 text-gray-400 hover:text-gray-600"
                        >
                          <Edit size={14} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
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
                <label className="block text-sm font-medium text-gray-700 mb-1">ФИО</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full border rounded-lg px-3 py-2"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Должность</label>
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
                  <label className="block text-sm font-medium text-gray-700 mb-1">Частота встреч</label>
                  <select
                    value={formData.meeting_frequency}
                    onChange={(e) => setFormData({ ...formData, meeting_frequency: e.target.value })}
                    className="w-full border rounded-lg px-3 py-2"
                  >
                    <option value="weekly">Еженедельно</option>
                    <option value="biweekly">Раз в 2 недели</option>
                    <option value="monthly">Ежемесячно</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">День встречи</label>
                  <select
                    value={formData.meeting_day}
                    onChange={(e) => setFormData({ ...formData, meeting_day: e.target.value })}
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
                <label className="block text-sm font-medium text-gray-700 mb-1">Приоритеты развития</label>
                <textarea
                  value={formData.development_priorities}
                  onChange={(e) => setFormData({ ...formData, development_priorities: e.target.value })}
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
