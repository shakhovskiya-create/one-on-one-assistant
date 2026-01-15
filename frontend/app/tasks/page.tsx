'use client'

import { useState, useEffect } from 'react'
import { 
  Plus, 
  Search, 
  Calendar, 
  User, 
  X,
  Check,
  AlertTriangle,
  Flame,
  MoreVertical
} from 'lucide-react'

import { API_URL } from '@/lib/config'

interface Employee {
  id: string
  name: string
}

interface TaskTag {
  id: string
  name: string
  color: string
}

interface Task {
  id: string
  title: string
  description?: string
  status: string
  priority: number
  flag_color?: string
  assignee?: Employee
  co_assignee?: Employee
  due_date?: string
  is_epic: boolean
  parent_id?: string
  tags: TaskTag[]
  blocks: string[]
  blocked_by: string[]
  is_blocking: boolean
  subtasks?: Task[]
  progress?: number
  comments?: Comment[]
  history?: HistoryEntry[]
}

interface Comment {
  id: string
  content: string
  author?: { name: string }
  created_at: string
}

interface HistoryEntry {
  id: string
  field_name: string
  old_value?: string
  new_value?: string
  created_at: string
}

interface KanbanData {
  backlog: Task[]
  todo: Task[]
  in_progress: Task[]
  review: Task[]
  done: Task[]
}

const STATUS_CONFIG: Record<string, { label: string; color: string }> = {
  backlog: { label: 'Бэклог', color: 'bg-gray-100' },
  todo: { label: 'К выполнению', color: 'bg-blue-100' },
  in_progress: { label: 'В работе', color: 'bg-yellow-100' },
  review: { label: 'На проверке', color: 'bg-purple-100' },
  done: { label: 'Готово', color: 'bg-green-100' }
}

const FLAG_COLORS: Record<string, { bg: string; label: string }> = {
  red: { bg: 'bg-red-500', label: 'Критично' },
  orange: { bg: 'bg-orange-500', label: 'Высокий' },
  yellow: { bg: 'bg-yellow-500', label: 'Средний' },
  green: { bg: 'bg-green-500', label: 'Низкий' },
  blue: { bg: 'bg-blue-500', label: 'Информация' },
  purple: { bg: 'bg-purple-500', label: 'Идея' }
}

const PRIORITY_LABELS = ['', 'Критический', 'Высокий', 'Средний', 'Низкий', 'Минимальный']

export default function TasksPage() {
  const [kanban, setKanban] = useState<KanbanData | null>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [tags, setTags] = useState<TaskTag[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showTaskModal, setShowTaskModal] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [filterAssignee, setFilterAssignee] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')
  const [draggedTask, setDraggedTask] = useState<Task | null>(null)

  useEffect(() => {
    fetchData()
  }, [filterAssignee])

  const fetchData = async () => {
    setLoading(true)
    try {
      const kanbanUrl = filterAssignee 
        ? `${API_URL}/kanban?assignee_id=${filterAssignee}` 
        : `${API_URL}/kanban`
      
      const [kanbanRes, employeesRes, tagsRes] = await Promise.all([
        fetch(kanbanUrl),
        fetch(`${API_URL}/employees`),
        fetch(`${API_URL}/tags`)
      ])
      
      if (kanbanRes.ok) setKanban(await kanbanRes.json())
      if (employeesRes.ok) setEmployees(await employeesRes.json())
      if (tagsRes.ok) setTags(await tagsRes.json())
    } catch (error) {
      console.error('Failed to fetch data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDragStart = (e: React.DragEvent, task: Task) => {
    setDraggedTask(task)
    e.dataTransfer.effectAllowed = 'move'
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }

  const handleDrop = async (e: React.DragEvent, newStatus: string) => {
    e.preventDefault()
    if (!draggedTask || draggedTask.status === newStatus) {
      setDraggedTask(null)
      return
    }

    if (draggedTask.blocked_by.length > 0 && newStatus === 'done') {
      alert('Задача заблокирована другими задачами!')
      setDraggedTask(null)
      return
    }

    try {
      await fetch(`${API_URL}/kanban/move?task_id=${draggedTask.id}&new_status=${newStatus}`, {
        method: 'PUT'
      })
      fetchData()
    } catch (error) {
      console.error('Failed to move task:', error)
    }
    
    setDraggedTask(null)
  }

  const openTaskDetails = async (taskId: string) => {
    try {
      const res = await fetch(`${API_URL}/tasks/${taskId}`)
      if (res.ok) {
        const task = await res.json()
        setSelectedTask(task)
        setShowTaskModal(true)
      }
    } catch (error) {
      console.error('Failed to fetch task:', error)
    }
  }

  const filterTasks = (tasks: Task[]) => {
    if (!searchQuery) return tasks
    return tasks.filter(t => 
      t.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      t.description?.toLowerCase().includes(searchQuery.toLowerCase())
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
    <div className="h-full flex flex-col">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Задачи</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus size={20} />
          Новая задача
        </button>
      </div>

      <div className="flex gap-4 mb-6">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
          <input
            type="text"
            placeholder="Поиск задач..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border rounded-lg"
          />
        </div>
        <select
          value={filterAssignee}
          onChange={(e) => setFilterAssignee(e.target.value)}
          className="px-4 py-2 border rounded-lg"
        >
          <option value="">Все исполнители</option>
          {employees.map(emp => (
            <option key={emp.id} value={emp.id}>{emp.name}</option>
          ))}
        </select>
      </div>

      <div className="flex-1 overflow-x-auto">
        <div className="flex gap-4 h-full min-w-max pb-4">
          {Object.entries(STATUS_CONFIG).map(([status, config]) => (
            <div
              key={status}
              className={`w-80 flex-shrink-0 ${config.color} rounded-lg p-4`}
              onDragOver={handleDragOver}
              onDrop={(e) => handleDrop(e, status)}
            >
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-semibold text-gray-700">{config.label}</h3>
                <span className="text-sm text-gray-500 bg-white px-2 py-1 rounded-full">
                  {kanban ? filterTasks(kanban[status as keyof KanbanData]).length : 0}
                </span>
              </div>
              
              <div className="space-y-3">
                {kanban && filterTasks(kanban[status as keyof KanbanData]).map(task => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    onDragStart={handleDragStart}
                    onClick={() => openTaskDetails(task.id)}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {showCreateModal && (
        <CreateTaskModal
          employees={employees}
          tags={tags}
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false)
            fetchData()
          }}
        />
      )}

      {showTaskModal && selectedTask && (
        <TaskDetailsModal
          task={selectedTask}
          employees={employees}
          tags={tags}
          onClose={() => {
            setShowTaskModal(false)
            setSelectedTask(null)
          }}
          onUpdated={() => {
            fetchData()
            if (selectedTask) openTaskDetails(selectedTask.id)
          }}
        />
      )}
    </div>
  )
}

function TaskCard({ 
  task, 
  onDragStart, 
  onClick 
}: { 
  task: Task
  onDragStart: (e: React.DragEvent, task: Task) => void
  onClick: () => void 
}) {
  const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== 'done'
  const flagConfig = task.flag_color ? FLAG_COLORS[task.flag_color] : null

  return (
    <div
      draggable
      onDragStart={(e) => onDragStart(e, task)}
      onClick={onClick}
      className={`bg-white rounded-lg p-4 shadow-sm cursor-pointer hover:shadow-md transition-shadow ${
        task.is_blocking ? 'ring-2 ring-red-400' : ''
      } ${(task.blocked_by?.length ?? 0) > 0 ? 'opacity-60' : ''}`}
    >
      <div className="flex items-center gap-2 mb-2">
        {flagConfig && (
          <div className={`w-3 h-3 rounded-full ${flagConfig.bg}`} />
        )}
        {task.is_epic && (
          <span className="text-xs bg-purple-100 text-purple-700 px-2 py-0.5 rounded">Эпик</span>
        )}
        {task.is_blocking && (
          <span title="Блокирует другие задачи">
            <Flame size={14} className="text-red-500" />
          </span>
        )}
        {task.blocked_by?.length > 0 && (
          <span title="Заблокирована">
            <AlertTriangle size={14} className="text-yellow-500" />
          </span>
        )}
      </div>

      <h4 className="font-medium text-gray-900 mb-2 line-clamp-2">{task.title}</h4>

      {task.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {task.tags.slice(0, 3).map(tag => (
            <span
              key={tag.id}
              className="text-xs px-2 py-0.5 rounded"
              style={{ backgroundColor: `${tag.color}20`, color: tag.color }}
            >
              {tag.name}
            </span>
          ))}
          {task.tags.length > 3 && (
            <span className="text-xs text-gray-500">+{task.tags.length - 3}</span>
          )}
        </div>
      )}

      {task.is_epic && task.progress !== undefined && (
        <div className="mb-2">
          <div className="h-1.5 bg-gray-200 rounded-full overflow-hidden">
            <div 
              className="h-full bg-green-500 transition-all"
              style={{ width: `${task.progress}%` }}
            />
          </div>
          <span className="text-xs text-gray-500">{task.progress}%</span>
        </div>
      )}

      <div className="flex items-center justify-between text-sm text-gray-500">
        <div className="flex items-center gap-2">
          {task.assignee && (
            <div className="flex items-center gap-1">
              <User size={14} />
              <span className="truncate max-w-[80px]">{task.assignee.name.split(' ')[0]}</span>
            </div>
          )}
        </div>
        
        {task.due_date && (
          <div className={`flex items-center gap-1 ${isOverdue ? 'text-red-500 font-medium' : ''}`}>
            <Calendar size={14} />
            <span>{new Date(task.due_date).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })}</span>
          </div>
        )}
      </div>
    </div>
  )
}

function CreateTaskModal({ 
  employees, 
  tags, 
  onClose, 
  onCreated 
}: { 
  employees: Employee[]
  tags: TaskTag[]
  onClose: () => void
  onCreated: () => void 
}) {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    status: 'todo',
    priority: 3,
    flag_color: '',
    assignee_id: '',
    co_assignee_id: '',
    due_date: '',
    is_epic: false,
    tags: [] as string[]
  })
  const [saving, setSaving] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.title.trim()) return

    setSaving(true)
    try {
      const res = await fetch(`${API_URL}/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          flag_color: formData.flag_color || null,
          assignee_id: formData.assignee_id || null,
          co_assignee_id: formData.co_assignee_id || null,
          due_date: formData.due_date || null
        })
      })
      
      if (res.ok) {
        onCreated()
      }
    } catch (error) {
      console.error('Failed to create task:', error)
    } finally {
      setSaving(false)
    }
  }

  const toggleTag = (tagName: string) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.includes(tagName)
        ? prev.tags.filter(t => t !== tagName)
        : [...prev.tags, tagName]
    }))
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl w-full max-w-lg max-h-[90vh] overflow-y-auto m-4">
        <div className="flex items-center justify-between p-4 border-b">
          <h2 className="text-lg font-semibold">Новая задача</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X size={24} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Название *</label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg"
              placeholder="Что нужно сделать?"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg"
              rows={3}
              placeholder="Подробности задачи..."
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg"
              >
                {Object.entries(STATUS_CONFIG).map(([value, { label }]) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Приоритет</label>
              <select
                value={formData.priority}
                onChange={(e) => setFormData({ ...formData, priority: Number(e.target.value) })}
                className="w-full px-3 py-2 border rounded-lg"
              >
                {[1, 2, 3, 4, 5].map(p => (
                  <option key={p} value={p}>{PRIORITY_LABELS[p]}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Исполнитель</label>
              <select
                value={formData.assignee_id}
                onChange={(e) => setFormData({ ...formData, assignee_id: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg"
              >
                <option value="">Не назначен</option>
                {employees.map(emp => (
                  <option key={emp.id} value={emp.id}>{emp.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Соисполнитель</label>
              <select
                value={formData.co_assignee_id}
                onChange={(e) => setFormData({ ...formData, co_assignee_id: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg"
              >
                <option value="">Не назначен</option>
                {employees.map(emp => (
                  <option key={emp.id} value={emp.id}>{emp.name}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Срок</label>
              <input
                type="date"
                value={formData.due_date}
                onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Флаг</label>
              <div className="flex gap-2">
                {Object.entries(FLAG_COLORS).map(([color, { bg, label }]) => (
                  <button
                    key={color}
                    type="button"
                    onClick={() => setFormData({ ...formData, flag_color: formData.flag_color === color ? '' : color })}
                    className={`w-8 h-8 rounded-full ${bg} ${formData.flag_color === color ? 'ring-2 ring-offset-2 ring-gray-400' : ''}`}
                    title={label}
                  />
                ))}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="is_epic"
              checked={formData.is_epic}
              onChange={(e) => setFormData({ ...formData, is_epic: e.target.checked })}
              className="w-4 h-4"
            />
            <label htmlFor="is_epic" className="text-sm text-gray-700">Это эпик (содержит подзадачи)</label>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Теги</label>
            <div className="flex flex-wrap gap-2">
              {tags.map(tag => (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => toggleTag(tag.name)}
                  className={`px-3 py-1 rounded-full text-sm transition-colors ${
                    formData.tags.includes(tag.name)
                      ? 'bg-blue-100 text-blue-700 ring-2 ring-blue-300'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {tag.name}
                </button>
              ))}
            </div>
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={saving || !formData.title.trim()}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              {saving ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function TaskDetailsModal({
  task,
  employees,
  tags,
  onClose,
  onUpdated
}: {
  task: Task
  employees: Employee[]
  tags: TaskTag[]
  onClose: () => void
  onUpdated: () => void
}) {
  const [activeTab, setActiveTab] = useState<'details' | 'comments' | 'history'>('details')
  const [newComment, setNewComment] = useState('')
  const [editing, setEditing] = useState(false)
  const [formData, setFormData] = useState({
    title: task.title,
    description: task.description || '',
    status: task.status,
    priority: task.priority,
    flag_color: task.flag_color || '',
    assignee_id: task.assignee?.id || '',
    co_assignee_id: task.co_assignee?.id || '',
    due_date: task.due_date || ''
  })

  const handleSave = async () => {
    try {
      await fetch(`${API_URL}/tasks/${task.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          flag_color: formData.flag_color || null,
          assignee_id: formData.assignee_id || null,
          co_assignee_id: formData.co_assignee_id || null,
          due_date: formData.due_date || null
        })
      })
      setEditing(false)
      onUpdated()
    } catch (error) {
      console.error('Failed to update task:', error)
    }
  }

  const handleAddComment = async () => {
    if (!newComment.trim()) return
    
    try {
      await fetch(`${API_URL}/task-comments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          task_id: task.id,
          author_id: employees[0]?.id,
          content: newComment
        })
      })
      setNewComment('')
      onUpdated()
    } catch (error) {
      console.error('Failed to add comment:', error)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Удалить задачу?')) return
    
    try {
      await fetch(`${API_URL}/tasks/${task.id}`, { method: 'DELETE' })
      onClose()
      onUpdated()
    } catch (error) {
      console.error('Failed to delete task:', error)
    }
  }

  const statusConfig = STATUS_CONFIG[task.status]

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl w-full max-w-2xl max-h-[90vh] overflow-hidden m-4 flex flex-col">
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-2">
            {task.is_epic && (
              <span className="text-xs bg-purple-100 text-purple-700 px-2 py-1 rounded">Эпик</span>
            )}
            {statusConfig && (
              <span className={`text-xs px-2 py-1 rounded ${statusConfig.color}`}>
                {statusConfig.label}
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setEditing(!editing)}
              className="p-2 text-gray-400 hover:text-gray-600"
            >
              {editing ? <X size={20} /> : <MoreVertical size={20} />}
            </button>
            <button onClick={onClose} className="p-2 text-gray-400 hover:text-gray-600">
              <X size={20} />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          {editing ? (
            <div className="space-y-4">
              <input
                type="text"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                className="w-full text-xl font-semibold px-3 py-2 border rounded-lg"
              />
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg"
                rows={4}
                placeholder="Описание..."
              />
              <div className="grid grid-cols-2 gap-4">
                <select
                  value={formData.status}
                  onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                  className="px-3 py-2 border rounded-lg"
                >
                  {Object.entries(STATUS_CONFIG).map(([value, { label }]) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </select>
                <select
                  value={formData.assignee_id}
                  onChange={(e) => setFormData({ ...formData, assignee_id: e.target.value })}
                  className="px-3 py-2 border rounded-lg"
                >
                  <option value="">Не назначен</option>
                  {employees.map(emp => (
                    <option key={emp.id} value={emp.id}>{emp.name}</option>
                  ))}
                </select>
              </div>
              <input
                type="date"
                value={formData.due_date}
                onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                className="px-3 py-2 border rounded-lg"
              />
              <div className="flex gap-3">
                <button
                  onClick={() => setEditing(false)}
                  className="px-4 py-2 border rounded-lg"
                >
                  Отмена
                </button>
                <button
                  onClick={handleSave}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg"
                >
                  Сохранить
                </button>
                <button
                  onClick={handleDelete}
                  className="px-4 py-2 bg-red-600 text-white rounded-lg ml-auto"
                >
                  Удалить
                </button>
              </div>
            </div>
          ) : (
            <div>
              <h2 className="text-xl font-semibold mb-2">{task.title}</h2>
              {task.description && (
                <p className="text-gray-600 mb-4">{task.description}</p>
              )}

              <div className="grid grid-cols-2 gap-4 mb-6">
                <div>
                  <span className="text-sm text-gray-500">Исполнитель</span>
                  <p className="font-medium">{task.assignee?.name || 'Не назначен'}</p>
                </div>
                {task.co_assignee && (
                  <div>
                    <span className="text-sm text-gray-500">Соисполнитель</span>
                    <p className="font-medium">{task.co_assignee.name}</p>
                  </div>
                )}
                {task.due_date && (
                  <div>
                    <span className="text-sm text-gray-500">Срок</span>
                    <p className="font-medium">{new Date(task.due_date).toLocaleDateString('ru-RU')}</p>
                  </div>
                )}
                <div>
                  <span className="text-sm text-gray-500">Приоритет</span>
                  <p className="font-medium">{PRIORITY_LABELS[task.priority]}</p>
                </div>
              </div>

              {task.tags.length > 0 && (
                <div className="mb-6">
                  <span className="text-sm text-gray-500 block mb-2">Теги</span>
                  <div className="flex flex-wrap gap-2">
                    {task.tags.map(tag => (
                      <span
                        key={tag.id}
                        className="px-3 py-1 rounded-full text-sm"
                        style={{ backgroundColor: `${tag.color}20`, color: tag.color }}
                      >
                        {tag.name}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {task.is_epic && task.subtasks && task.subtasks.length > 0 && (
                <div className="mb-6">
                  <span className="text-sm text-gray-500 block mb-2">
                    Подзадачи ({task.subtasks.filter(s => s.status === 'done').length}/{task.subtasks.length})
                  </span>
                  <div className="space-y-2">
                    {task.subtasks.map(subtask => (
                      <div key={subtask.id} className="flex items-center gap-2 p-2 bg-gray-50 rounded">
                        {subtask.status === 'done' ? (
                          <Check size={16} className="text-green-500" />
                        ) : (
                          <div className="w-4 h-4 border-2 rounded" />
                        )}
                        <span className={subtask.status === 'done' ? 'line-through text-gray-400' : ''}>
                          {subtask.title}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="border-t pt-4">
                <div className="flex gap-4 mb-4">
                  <button
                    onClick={() => setActiveTab('comments')}
                    className={`text-sm font-medium ${activeTab === 'comments' ? 'text-blue-600' : 'text-gray-500'}`}
                  >
                    Комментарии ({task.comments?.length || 0})
                  </button>
                  <button
                    onClick={() => setActiveTab('history')}
                    className={`text-sm font-medium ${activeTab === 'history' ? 'text-blue-600' : 'text-gray-500'}`}
                  >
                    История
                  </button>
                </div>

                {activeTab === 'comments' && (
                  <div className="space-y-3">
                    {task.comments?.map((comment) => (
                      <div key={comment.id} className="p-3 bg-gray-50 rounded-lg">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-medium text-sm">{comment.author?.name || 'Аноним'}</span>
                          <span className="text-xs text-gray-400">
                            {new Date(comment.created_at).toLocaleString('ru-RU')}
                          </span>
                        </div>
                        <p className="text-gray-700">{comment.content}</p>
                      </div>
                    ))}
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={newComment}
                        onChange={(e) => setNewComment(e.target.value)}
                        placeholder="Добавить комментарий..."
                        className="flex-1 px-3 py-2 border rounded-lg"
                        onKeyDown={(e) => e.key === 'Enter' && handleAddComment()}
                      />
                      <button
                        onClick={handleAddComment}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg"
                      >
                        Отправить
                      </button>
                    </div>
                  </div>
                )}

                {activeTab === 'history' && (
                  <div className="space-y-2">
                    {task.history?.map((entry) => (
                      <div key={entry.id} className="text-sm">
                        <span className="text-gray-400">
                          {new Date(entry.created_at).toLocaleString('ru-RU')}
                        </span>
                        {' — '}
                        <span className="font-medium">{entry.field_name}</span>
                        {' изменено с '}
                        <span className="text-red-500">{entry.old_value || '(пусто)'}</span>
                        {' на '}
                        <span className="text-green-500">{entry.new_value || '(пусто)'}</span>
                      </div>
                    ))}
                    {!task.history?.length && (
                      <p className="text-gray-400 text-sm">История изменений пуста</p>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
