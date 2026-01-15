'use client'

import { useState, useEffect, useMemo, useRef, useCallback } from 'react'
import {
  Plus,
  Search,
  Calendar,
  User,
  X,
  Check,
  AlertTriangle,
  Flame,
  MoreVertical,
  GripVertical,
  LayoutGrid,
  List,
  Table2,
  ChevronDown,
  Filter,
  SortAsc,
  SortDesc,
  MessageSquare,
  AtSign,
  Send,
  Settings2,
  Eye,
  EyeOff,
  ArrowUpDown
} from 'lucide-react'
import {
  DndContext,
  DragOverlay,
  closestCorners,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragStartEvent,
  DragEndEvent,
  DragOverEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

import { API_URL } from '@/lib/config'

// Types
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
  tags?: TaskTag[]
  blocks?: string[]
  blocked_by?: string[]
  is_blocking?: boolean
  subtasks?: Task[]
  progress?: number
  comments?: Comment[]
  history?: HistoryEntry[]
  created_at?: string
  updated_at?: string
}

interface Comment {
  id: string
  content: string
  author?: { id: string; name: string }
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

type ViewMode = 'board' | 'list' | 'table'
type SortField = 'title' | 'status' | 'priority' | 'due_date' | 'assignee' | 'created_at'
type SortDirection = 'asc' | 'desc'

// Config
const STATUS_CONFIG: Record<string, { label: string; color: string; bgDrop: string; textColor: string }> = {
  backlog: { label: 'Бэклог', color: 'bg-gray-100', bgDrop: 'bg-gray-200', textColor: 'text-gray-700' },
  todo: { label: 'К выполнению', color: 'bg-blue-100', bgDrop: 'bg-blue-200', textColor: 'text-blue-700' },
  in_progress: { label: 'В работе', color: 'bg-yellow-100', bgDrop: 'bg-yellow-200', textColor: 'text-yellow-700' },
  review: { label: 'На проверке', color: 'bg-purple-100', bgDrop: 'bg-purple-200', textColor: 'text-purple-700' },
  done: { label: 'Готово', color: 'bg-green-100', bgDrop: 'bg-green-200', textColor: 'text-green-700' }
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
const PRIORITY_COLORS = ['', 'text-red-600', 'text-orange-600', 'text-yellow-600', 'text-blue-600', 'text-gray-600']

// Column visibility config
interface ColumnConfig {
  id: string
  label: string
  visible: boolean
  width?: string
}

const DEFAULT_COLUMNS: ColumnConfig[] = [
  { id: 'title', label: 'Название', visible: true, width: 'w-64' },
  { id: 'status', label: 'Статус', visible: true, width: 'w-32' },
  { id: 'priority', label: 'Приоритет', visible: true, width: 'w-28' },
  { id: 'assignee', label: 'Исполнитель', visible: true, width: 'w-36' },
  { id: 'due_date', label: 'Срок', visible: true, width: 'w-28' },
  { id: 'tags', label: 'Теги', visible: true, width: 'w-40' },
  { id: 'comments', label: 'Комментарии', visible: false, width: 'w-20' },
  { id: 'created_at', label: 'Создана', visible: false, width: 'w-28' },
]

// Main Component
export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [kanban, setKanban] = useState<KanbanData | null>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [tags, setTags] = useState<TaskTag[]>([])
  const [loading, setLoading] = useState(true)

  // View state
  const [viewMode, setViewMode] = useState<ViewMode>('board')
  const [columns, setColumns] = useState<ColumnConfig[]>(DEFAULT_COLUMNS)
  const [showColumnSettings, setShowColumnSettings] = useState(false)

  // Filter & Sort
  const [searchQuery, setSearchQuery] = useState('')
  const [filterAssignee, setFilterAssignee] = useState('')
  const [filterStatus, setFilterStatus] = useState('')
  const [filterPriority, setFilterPriority] = useState('')
  const [sortField, setSortField] = useState<SortField>('created_at')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')
  const [showFilters, setShowFilters] = useState(false)

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showTaskModal, setShowTaskModal] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)

  // DnD state
  const [activeTask, setActiveTask] = useState<Task | null>(null)
  const [overColumn, setOverColumn] = useState<string | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      const [tasksRes, kanbanRes, employeesRes, tagsRes] = await Promise.all([
        fetch(`${API_URL}/tasks`),
        fetch(`${API_URL}/kanban`),
        fetch(`${API_URL}/employees`),
        fetch(`${API_URL}/tags`)
      ])

      if (tasksRes.ok) setTasks(await tasksRes.json())
      if (kanbanRes.ok) setKanban(await kanbanRes.json())
      if (employeesRes.ok) setEmployees(await employeesRes.json())
      if (tagsRes.ok) setTags(await tagsRes.json())
    } catch (error) {
      console.error('Failed to fetch data:', error)
    } finally {
      setLoading(false)
    }
  }

  // Filter and sort tasks
  const filteredTasks = useMemo(() => {
    let result = [...tasks]

    // Search
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      result = result.filter(t =>
        t.title.toLowerCase().includes(query) ||
        t.description?.toLowerCase().includes(query)
      )
    }

    // Filters
    if (filterAssignee) {
      result = result.filter(t => t.assignee?.id === filterAssignee)
    }
    if (filterStatus) {
      result = result.filter(t => t.status === filterStatus)
    }
    if (filterPriority) {
      result = result.filter(t => t.priority === parseInt(filterPriority))
    }

    // Sort
    result.sort((a, b) => {
      let comparison = 0
      switch (sortField) {
        case 'title':
          comparison = a.title.localeCompare(b.title)
          break
        case 'status':
          const statusOrder = ['backlog', 'todo', 'in_progress', 'review', 'done']
          comparison = statusOrder.indexOf(a.status) - statusOrder.indexOf(b.status)
          break
        case 'priority':
          comparison = a.priority - b.priority
          break
        case 'due_date':
          if (!a.due_date && !b.due_date) comparison = 0
          else if (!a.due_date) comparison = 1
          else if (!b.due_date) comparison = -1
          else comparison = new Date(a.due_date).getTime() - new Date(b.due_date).getTime()
          break
        case 'assignee':
          comparison = (a.assignee?.name || 'zzz').localeCompare(b.assignee?.name || 'zzz')
          break
        case 'created_at':
          comparison = new Date(a.created_at || 0).getTime() - new Date(b.created_at || 0).getTime()
          break
      }
      return sortDirection === 'asc' ? comparison : -comparison
    })

    return result
  }, [tasks, searchQuery, filterAssignee, filterStatus, filterPriority, sortField, sortDirection])

  // Filtered kanban for board view
  const filteredKanban = useMemo(() => {
    if (!kanban) return null

    const filterFn = (task: Task) => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        if (!task.title.toLowerCase().includes(query) &&
            !task.description?.toLowerCase().includes(query)) {
          return false
        }
      }
      if (filterAssignee && task.assignee?.id !== filterAssignee) return false
      if (filterPriority && task.priority !== parseInt(filterPriority)) return false
      return true
    }

    return {
      backlog: kanban.backlog.filter(filterFn),
      todo: kanban.todo.filter(filterFn),
      in_progress: kanban.in_progress.filter(filterFn),
      review: kanban.review.filter(filterFn),
      done: kanban.done.filter(filterFn),
    }
  }, [kanban, searchQuery, filterAssignee, filterPriority])

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  const toggleColumn = (columnId: string) => {
    setColumns(prev => prev.map(col =>
      col.id === columnId ? { ...col, visible: !col.visible } : col
    ))
  }

  // DnD handlers
  const handleDragStart = (event: DragStartEvent) => {
    const taskId = event.active.id as string
    if (kanban) {
      for (const status of Object.keys(kanban) as Array<keyof KanbanData>) {
        const task = kanban[status].find(t => t.id === taskId)
        if (task) {
          setActiveTask(task)
          break
        }
      }
    }
  }

  const handleDragOver = (event: DragOverEvent) => {
    const { over } = event
    if (over) {
      if (Object.keys(STATUS_CONFIG).includes(over.id as string)) {
        setOverColumn(over.id as string)
      } else if (kanban) {
        for (const status of Object.keys(kanban) as Array<keyof KanbanData>) {
          if (kanban[status].find(t => t.id === over.id)) {
            setOverColumn(status)
            break
          }
        }
      }
    } else {
      setOverColumn(null)
    }
  }

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event
    setActiveTask(null)
    setOverColumn(null)

    if (!over || !kanban) return

    const taskId = active.id as string
    let newStatus: keyof KanbanData | null = null

    if (Object.keys(STATUS_CONFIG).includes(over.id as string)) {
      newStatus = over.id as keyof KanbanData
    } else {
      for (const status of Object.keys(kanban) as Array<keyof KanbanData>) {
        if (kanban[status].find(t => t.id === over.id)) {
          newStatus = status
          break
        }
      }
    }

    if (!newStatus) return

    let currentTask: Task | null = null
    let currentStatus: keyof KanbanData | null = null

    for (const status of Object.keys(kanban) as Array<keyof KanbanData>) {
      const task = kanban[status].find(t => t.id === taskId)
      if (task) {
        currentTask = task
        currentStatus = status
        break
      }
    }

    if (!currentTask || !currentStatus || currentStatus === newStatus) return

    // Update local state optimistically
    const newKanban: KanbanData = { ...kanban }
    newKanban[currentStatus] = kanban[currentStatus].filter(t => t.id !== taskId)
    newKanban[newStatus] = [...kanban[newStatus], { ...currentTask, status: newStatus }]
    setKanban(newKanban)

    // Also update tasks array
    setTasks(prev => prev.map(t => t.id === taskId ? { ...t, status: newStatus! } : t))

    try {
      await fetch(`${API_URL}/kanban/move?task_id=${taskId}&new_status=${newStatus}`, { method: 'PUT' })
    } catch (error) {
      console.error('Failed to move task:', error)
      fetchData()
    }
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

  const activeFiltersCount = [filterAssignee, filterStatus, filterPriority].filter(Boolean).length

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold text-gray-900">Задачи</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus size={20} />
          Новая задача
        </button>
      </div>

      {/* Toolbar */}
      <div className="flex flex-wrap gap-3 mb-4 items-center">
        {/* View Toggle - Slack-style */}
        <div className="flex bg-gray-100 rounded-lg p-1">
          <button
            onClick={() => setViewMode('board')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              viewMode === 'board' ? 'bg-white shadow text-gray-900' : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <LayoutGrid size={16} />
            Доска
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              viewMode === 'list' ? 'bg-white shadow text-gray-900' : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <List size={16} />
            Список
          </button>
          <button
            onClick={() => setViewMode('table')}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              viewMode === 'table' ? 'bg-white shadow text-gray-900' : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <Table2 size={16} />
            Таблица
          </button>
        </div>

        {/* Search */}
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
          <input
            type="text"
            placeholder="Поиск..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-4 py-2 border rounded-lg text-sm"
          />
        </div>

        {/* Filter Button */}
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`flex items-center gap-1.5 px-3 py-2 border rounded-lg text-sm ${
            activeFiltersCount > 0 ? 'border-blue-500 text-blue-600 bg-blue-50' : 'text-gray-600'
          }`}
        >
          <Filter size={16} />
          Фильтры
          {activeFiltersCount > 0 && (
            <span className="ml-1 px-1.5 py-0.5 bg-blue-600 text-white text-xs rounded-full">
              {activeFiltersCount}
            </span>
          )}
        </button>

        {/* Column Settings (for list/table views) */}
        {(viewMode === 'list' || viewMode === 'table') && (
          <div className="relative">
            <button
              onClick={() => setShowColumnSettings(!showColumnSettings)}
              className="flex items-center gap-1.5 px-3 py-2 border rounded-lg text-sm text-gray-600"
            >
              <Settings2 size={16} />
              Колонки
            </button>

            {showColumnSettings && (
              <div className="absolute right-0 top-full mt-1 w-56 bg-white border rounded-lg shadow-lg z-20 py-2">
                <div className="px-3 py-1.5 text-xs font-medium text-gray-500 uppercase">Показать колонки</div>
                {columns.map(col => (
                  <button
                    key={col.id}
                    onClick={() => toggleColumn(col.id)}
                    className="w-full flex items-center gap-2 px-3 py-2 hover:bg-gray-50 text-sm"
                  >
                    {col.visible ? <Eye size={16} className="text-blue-600" /> : <EyeOff size={16} className="text-gray-400" />}
                    {col.label}
                  </button>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <div className="flex flex-wrap gap-3 mb-4 p-3 bg-gray-50 rounded-lg">
          <select
            value={filterAssignee}
            onChange={(e) => setFilterAssignee(e.target.value)}
            className="px-3 py-1.5 border rounded-lg text-sm"
          >
            <option value="">Все исполнители</option>
            {employees.map(emp => (
              <option key={emp.id} value={emp.id}>{emp.name}</option>
            ))}
          </select>

          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-3 py-1.5 border rounded-lg text-sm"
          >
            <option value="">Все статусы</option>
            {Object.entries(STATUS_CONFIG).map(([value, { label }]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>

          <select
            value={filterPriority}
            onChange={(e) => setFilterPriority(e.target.value)}
            className="px-3 py-1.5 border rounded-lg text-sm"
          >
            <option value="">Все приоритеты</option>
            {[1, 2, 3, 4, 5].map(p => (
              <option key={p} value={p}>{PRIORITY_LABELS[p]}</option>
            ))}
          </select>

          {activeFiltersCount > 0 && (
            <button
              onClick={() => {
                setFilterAssignee('')
                setFilterStatus('')
                setFilterPriority('')
              }}
              className="px-3 py-1.5 text-sm text-red-600 hover:text-red-700"
            >
              Сбросить фильтры
            </button>
          )}
        </div>
      )}

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {viewMode === 'board' && filteredKanban && (
          <BoardView
            kanban={filteredKanban}
            sensors={sensors}
            activeTask={activeTask}
            overColumn={overColumn}
            onDragStart={handleDragStart}
            onDragOver={handleDragOver}
            onDragEnd={handleDragEnd}
            onTaskClick={openTaskDetails}
          />
        )}

        {viewMode === 'list' && (
          <ListView
            tasks={filteredTasks}
            columns={columns.filter(c => c.visible)}
            sortField={sortField}
            sortDirection={sortDirection}
            onSort={handleSort}
            onTaskClick={openTaskDetails}
          />
        )}

        {viewMode === 'table' && (
          <TableView
            tasks={filteredTasks}
            columns={columns.filter(c => c.visible)}
            employees={employees}
            sortField={sortField}
            sortDirection={sortDirection}
            onSort={handleSort}
            onTaskClick={openTaskDetails}
            onTaskUpdate={fetchData}
          />
        )}
      </div>

      {/* Modals */}
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

// ============ Board View (Kanban) ============
function BoardView({
  kanban,
  sensors,
  activeTask,
  overColumn,
  onDragStart,
  onDragOver,
  onDragEnd,
  onTaskClick
}: {
  kanban: KanbanData
  sensors: ReturnType<typeof useSensors>
  activeTask: Task | null
  overColumn: string | null
  onDragStart: (event: DragStartEvent) => void
  onDragOver: (event: DragOverEvent) => void
  onDragEnd: (event: DragEndEvent) => void
  onTaskClick: (id: string) => void
}) {
  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDragEnd={onDragEnd}
    >
      <div className="flex gap-4 h-full overflow-x-auto pb-4">
        {Object.entries(STATUS_CONFIG).map(([status, config]) => {
          const tasks = kanban[status as keyof KanbanData]
          const isOver = overColumn === status

          return (
            <KanbanColumn
              key={status}
              id={status}
              title={config.label}
              color={isOver ? config.bgDrop : config.color}
              count={tasks.length}
              tasks={tasks}
              onTaskClick={onTaskClick}
            />
          )
        })}
      </div>

      <DragOverlay>
        {activeTask && (
          <div className="rotate-3 opacity-90">
            <TaskCard task={activeTask} isDragging />
          </div>
        )}
      </DragOverlay>
    </DndContext>
  )
}

function KanbanColumn({ id, title, color, count, tasks, onTaskClick }: {
  id: string
  title: string
  color: string
  count: number
  tasks: Task[]
  onTaskClick: (id: string) => void
}) {
  const { setNodeRef } = useSortable({ id, data: { type: 'column' } })

  return (
    <div ref={setNodeRef} className={`w-72 flex-shrink-0 ${color} rounded-lg p-3 transition-colors duration-200`}>
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-semibold text-gray-700 text-sm">{title}</h3>
        <span className="text-xs text-gray-500 bg-white px-2 py-0.5 rounded-full">{count}</span>
      </div>

      <SortableContext items={tasks.map(t => t.id)} strategy={verticalListSortingStrategy}>
        <div className="space-y-2 min-h-[100px]">
          {tasks.map(task => (
            <SortableTaskCard key={task.id} task={task} onClick={() => onTaskClick(task.id)} />
          ))}
        </div>
      </SortableContext>
    </div>
  )
}

function SortableTaskCard({ task, onClick }: { task: Task; onClick: () => void }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: task.id })
  const style = { transform: CSS.Transform.toString(transform), transition }

  if (isDragging) {
    return <div ref={setNodeRef} style={style} className="bg-blue-50 border-2 border-dashed border-blue-300 rounded-lg p-4 h-[100px]" />
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners} onClick={onClick} className="touch-none">
      <TaskCard task={task} />
    </div>
  )
}

function TaskCard({ task, isDragging = false }: { task: Task; isDragging?: boolean }) {
  const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== 'done'
  const flagConfig = task.flag_color ? FLAG_COLORS[task.flag_color] : null

  return (
    <div className={`bg-white rounded-lg p-3 shadow-sm cursor-grab active:cursor-grabbing hover:shadow-md transition-all ${isDragging ? 'shadow-xl' : ''}`}>
      <div className="flex items-center gap-1.5 mb-2">
        <GripVertical size={12} className="text-gray-300" />
        {flagConfig && <div className={`w-2 h-2 rounded-full ${flagConfig.bg}`} />}
        {task.is_epic && <span className="text-xs bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded">Эпик</span>}
      </div>

      <h4 className="font-medium text-gray-900 text-sm mb-2 line-clamp-2">{task.title}</h4>

      {(task.tags?.length ?? 0) > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {task.tags!.slice(0, 2).map(tag => (
            <span key={tag.id} className="text-xs px-1.5 py-0.5 rounded" style={{ backgroundColor: `${tag.color}20`, color: tag.color }}>
              {tag.name}
            </span>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between text-xs text-gray-500">
        {task.assignee && (
          <div className="flex items-center gap-1">
            <User size={12} />
            <span className="truncate max-w-[70px]">{task.assignee.name.split(' ')[0]}</span>
          </div>
        )}
        {task.due_date && (
          <div className={`flex items-center gap-1 ${isOverdue ? 'text-red-500 font-medium' : ''}`}>
            <Calendar size={12} />
            <span>{new Date(task.due_date).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })}</span>
          </div>
        )}
      </div>
    </div>
  )
}

// ============ List View ============
function ListView({
  tasks,
  columns,
  sortField,
  sortDirection,
  onSort,
  onTaskClick
}: {
  tasks: Task[]
  columns: ColumnConfig[]
  sortField: SortField
  sortDirection: SortDirection
  onSort: (field: SortField) => void
  onTaskClick: (id: string) => void
}) {
  return (
    <div className="bg-white rounded-lg border overflow-hidden h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center border-b bg-gray-50 text-xs font-medium text-gray-500 uppercase">
        {columns.map(col => (
          <button
            key={col.id}
            onClick={() => ['title', 'status', 'priority', 'due_date', 'assignee', 'created_at'].includes(col.id) && onSort(col.id as SortField)}
            className={`flex items-center gap-1 px-4 py-3 ${col.width} hover:bg-gray-100 text-left`}
          >
            {col.label}
            {sortField === col.id && (
              sortDirection === 'asc' ? <SortAsc size={14} /> : <SortDesc size={14} />
            )}
          </button>
        ))}
      </div>

      {/* Rows */}
      <div className="flex-1 overflow-y-auto">
        {tasks.map(task => (
          <div
            key={task.id}
            onClick={() => onTaskClick(task.id)}
            className="flex items-center border-b hover:bg-gray-50 cursor-pointer transition-colors"
          >
            {columns.map(col => (
              <div key={col.id} className={`px-4 py-3 ${col.width} truncate`}>
                {col.id === 'title' && (
                  <div className="flex items-center gap-2">
                    {task.flag_color && <div className={`w-2 h-2 rounded-full ${FLAG_COLORS[task.flag_color]?.bg}`} />}
                    <span className="font-medium text-gray-900">{task.title}</span>
                  </div>
                )}
                {col.id === 'status' && (
                  <span className={`inline-flex px-2 py-1 rounded text-xs ${STATUS_CONFIG[task.status]?.color} ${STATUS_CONFIG[task.status]?.textColor}`}>
                    {STATUS_CONFIG[task.status]?.label}
                  </span>
                )}
                {col.id === 'priority' && (
                  <span className={`text-sm ${PRIORITY_COLORS[task.priority]}`}>
                    {PRIORITY_LABELS[task.priority]}
                  </span>
                )}
                {col.id === 'assignee' && (
                  <div className="flex items-center gap-1.5 text-sm text-gray-600">
                    {task.assignee ? (
                      <>
                        <User size={14} />
                        {task.assignee.name}
                      </>
                    ) : (
                      <span className="text-gray-400">—</span>
                    )}
                  </div>
                )}
                {col.id === 'due_date' && (
                  <span className={`text-sm ${task.due_date && new Date(task.due_date) < new Date() && task.status !== 'done' ? 'text-red-500' : 'text-gray-600'}`}>
                    {task.due_date ? new Date(task.due_date).toLocaleDateString('ru-RU') : '—'}
                  </span>
                )}
                {col.id === 'tags' && (
                  <div className="flex gap-1">
                    {task.tags?.slice(0, 2).map(tag => (
                      <span key={tag.id} className="text-xs px-1.5 py-0.5 rounded" style={{ backgroundColor: `${tag.color}20`, color: tag.color }}>
                        {tag.name}
                      </span>
                    ))}
                  </div>
                )}
                {col.id === 'comments' && (
                  <div className="flex items-center gap-1 text-gray-500">
                    <MessageSquare size={14} />
                    <span className="text-sm">{task.comments?.length || 0}</span>
                  </div>
                )}
                {col.id === 'created_at' && (
                  <span className="text-sm text-gray-500">
                    {task.created_at ? new Date(task.created_at).toLocaleDateString('ru-RU') : '—'}
                  </span>
                )}
              </div>
            ))}
          </div>
        ))}

        {tasks.length === 0 && (
          <div className="flex items-center justify-center h-32 text-gray-400">
            Нет задач
          </div>
        )}
      </div>
    </div>
  )
}

// ============ Table View (Editable) ============
function TableView({
  tasks,
  columns,
  employees,
  sortField,
  sortDirection,
  onSort,
  onTaskClick,
  onTaskUpdate
}: {
  tasks: Task[]
  columns: ColumnConfig[]
  employees: Employee[]
  sortField: SortField
  sortDirection: SortDirection
  onSort: (field: SortField) => void
  onTaskClick: (id: string) => void
  onTaskUpdate: () => void
}) {
  const [editingCell, setEditingCell] = useState<{ taskId: string; field: string } | null>(null)
  const [editValue, setEditValue] = useState('')

  const handleCellClick = (taskId: string, field: string, currentValue: string) => {
    if (['status', 'priority', 'assignee'].includes(field)) {
      setEditingCell({ taskId, field })
      setEditValue(currentValue)
    }
  }

  const handleCellChange = async (taskId: string, field: string, value: string) => {
    setEditingCell(null)

    try {
      const updateData: Record<string, string | number | null> = {}
      if (field === 'status') updateData.status = value
      if (field === 'priority') updateData.priority = parseInt(value)
      if (field === 'assignee') updateData.assignee_id = value || null

      await fetch(`${API_URL}/tasks/${taskId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updateData)
      })
      onTaskUpdate()
    } catch (error) {
      console.error('Failed to update task:', error)
    }
  }

  return (
    <div className="bg-white rounded-lg border overflow-hidden h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center border-b bg-gray-50 text-xs font-medium text-gray-500 uppercase sticky top-0">
        {columns.map(col => (
          <button
            key={col.id}
            onClick={() => ['title', 'status', 'priority', 'due_date', 'assignee', 'created_at'].includes(col.id) && onSort(col.id as SortField)}
            className={`flex items-center gap-1 px-3 py-2 ${col.width} hover:bg-gray-100 text-left border-r last:border-r-0`}
          >
            {col.label}
            {sortField === col.id && (
              sortDirection === 'asc' ? <SortAsc size={12} /> : <SortDesc size={12} />
            )}
          </button>
        ))}
      </div>

      {/* Rows */}
      <div className="flex-1 overflow-y-auto">
        {tasks.map(task => (
          <div key={task.id} className="flex items-center border-b hover:bg-blue-50/30">
            {columns.map(col => (
              <div
                key={col.id}
                className={`px-3 py-2 ${col.width} border-r last:border-r-0 text-sm`}
                onClick={() => {
                  if (col.id === 'title') {
                    onTaskClick(task.id)
                  } else if (['status', 'priority', 'assignee'].includes(col.id)) {
                    const val = col.id === 'status' ? task.status :
                               col.id === 'priority' ? task.priority.toString() :
                               task.assignee?.id || ''
                    handleCellClick(task.id, col.id, val)
                  }
                }}
              >
                {col.id === 'title' && (
                  <div className="flex items-center gap-2 cursor-pointer hover:text-blue-600">
                    {task.flag_color && <div className={`w-2 h-2 rounded-full ${FLAG_COLORS[task.flag_color]?.bg}`} />}
                    <span className="truncate">{task.title}</span>
                  </div>
                )}

                {col.id === 'status' && (
                  editingCell?.taskId === task.id && editingCell?.field === 'status' ? (
                    <select
                      autoFocus
                      value={editValue}
                      onChange={(e) => handleCellChange(task.id, 'status', e.target.value)}
                      onBlur={() => setEditingCell(null)}
                      className="w-full px-1 py-0.5 border rounded text-xs"
                    >
                      {Object.entries(STATUS_CONFIG).map(([value, { label }]) => (
                        <option key={value} value={value}>{label}</option>
                      ))}
                    </select>
                  ) : (
                    <span className={`inline-flex px-2 py-0.5 rounded text-xs cursor-pointer ${STATUS_CONFIG[task.status]?.color} ${STATUS_CONFIG[task.status]?.textColor}`}>
                      {STATUS_CONFIG[task.status]?.label}
                    </span>
                  )
                )}

                {col.id === 'priority' && (
                  editingCell?.taskId === task.id && editingCell?.field === 'priority' ? (
                    <select
                      autoFocus
                      value={editValue}
                      onChange={(e) => handleCellChange(task.id, 'priority', e.target.value)}
                      onBlur={() => setEditingCell(null)}
                      className="w-full px-1 py-0.5 border rounded text-xs"
                    >
                      {[1, 2, 3, 4, 5].map(p => (
                        <option key={p} value={p}>{PRIORITY_LABELS[p]}</option>
                      ))}
                    </select>
                  ) : (
                    <span className={`cursor-pointer ${PRIORITY_COLORS[task.priority]}`}>
                      {PRIORITY_LABELS[task.priority]}
                    </span>
                  )
                )}

                {col.id === 'assignee' && (
                  editingCell?.taskId === task.id && editingCell?.field === 'assignee' ? (
                    <select
                      autoFocus
                      value={editValue}
                      onChange={(e) => handleCellChange(task.id, 'assignee', e.target.value)}
                      onBlur={() => setEditingCell(null)}
                      className="w-full px-1 py-0.5 border rounded text-xs"
                    >
                      <option value="">Не назначен</option>
                      {employees.map(emp => (
                        <option key={emp.id} value={emp.id}>{emp.name}</option>
                      ))}
                    </select>
                  ) : (
                    <div className="flex items-center gap-1 cursor-pointer text-gray-600">
                      {task.assignee ? (
                        <>
                          <User size={12} />
                          <span className="truncate">{task.assignee.name}</span>
                        </>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </div>
                  )
                )}

                {col.id === 'due_date' && (
                  <span className={`${task.due_date && new Date(task.due_date) < new Date() && task.status !== 'done' ? 'text-red-500' : 'text-gray-600'}`}>
                    {task.due_date ? new Date(task.due_date).toLocaleDateString('ru-RU') : '—'}
                  </span>
                )}

                {col.id === 'tags' && (
                  <div className="flex gap-1 overflow-hidden">
                    {task.tags?.slice(0, 2).map(tag => (
                      <span key={tag.id} className="text-xs px-1 py-0.5 rounded whitespace-nowrap" style={{ backgroundColor: `${tag.color}20`, color: tag.color }}>
                        {tag.name}
                      </span>
                    ))}
                  </div>
                )}

                {col.id === 'comments' && (
                  <div className="flex items-center gap-1 text-gray-500">
                    <MessageSquare size={12} />
                    {task.comments?.length || 0}
                  </div>
                )}

                {col.id === 'created_at' && (
                  <span className="text-gray-500">
                    {task.created_at ? new Date(task.created_at).toLocaleDateString('ru-RU') : '—'}
                  </span>
                )}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}

// ============ Create Task Modal ============
function CreateTaskModal({ employees, tags, onClose, onCreated }: {
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
      if (res.ok) onCreated()
    } catch (error) {
      console.error('Failed to create task:', error)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl w-full max-w-lg max-h-[90vh] overflow-y-auto m-4">
        <div className="flex items-center justify-between p-4 border-b">
          <h2 className="text-lg font-semibold">Новая задача</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={24} /></button>
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
              <label className="block text-sm font-medium text-gray-700 mb-1">Срок</label>
              <input
                type="date"
                value={formData.due_date}
                onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Флаг приоритета</label>
            <div className="flex gap-2">
              {Object.entries(FLAG_COLORS).map(([color, { bg, label }]) => (
                <button
                  key={color}
                  type="button"
                  onClick={() => setFormData({ ...formData, flag_color: formData.flag_color === color ? '' : color })}
                  className={`w-7 h-7 rounded-full ${bg} ${formData.flag_color === color ? 'ring-2 ring-offset-2 ring-gray-400' : ''}`}
                  title={label}
                />
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Теги</label>
            <div className="flex flex-wrap gap-2">
              {tags.map(tag => (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => setFormData(prev => ({
                    ...prev,
                    tags: prev.tags.includes(tag.name) ? prev.tags.filter(t => t !== tag.name) : [...prev.tags, tag.name]
                  }))}
                  className={`px-3 py-1 rounded-full text-sm ${
                    formData.tags.includes(tag.name) ? 'bg-blue-100 text-blue-700 ring-2 ring-blue-300' : 'bg-gray-100 text-gray-700'
                  }`}
                >
                  {tag.name}
                </button>
              ))}
            </div>
          </div>

          <div className="flex gap-3 pt-4">
            <button type="button" onClick={onClose} className="flex-1 px-4 py-2 border rounded-lg">Отмена</button>
            <button type="submit" disabled={saving || !formData.title.trim()} className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg disabled:opacity-50">
              {saving ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// ============ Task Details Modal with @mentions ============
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
  const [showMentions, setShowMentions] = useState(false)
  const [mentionFilter, setMentionFilter] = useState('')
  const [editing, setEditing] = useState(false)
  const [formData, setFormData] = useState({
    title: task.title,
    description: task.description || '',
    status: task.status,
    priority: task.priority,
    flag_color: task.flag_color || '',
    assignee_id: task.assignee?.id || '',
    due_date: task.due_date || ''
  })
  const commentInputRef = useRef<HTMLTextAreaElement>(null)

  const taskComments = task.comments ?? []
  const taskHistory = task.history ?? []

  // Handle @mention
  const handleCommentChange = (value: string) => {
    setNewComment(value)

    const lastAt = value.lastIndexOf('@')
    if (lastAt !== -1 && lastAt === value.length - 1) {
      setShowMentions(true)
      setMentionFilter('')
    } else if (lastAt !== -1) {
      const textAfterAt = value.substring(lastAt + 1)
      if (!textAfterAt.includes(' ')) {
        setShowMentions(true)
        setMentionFilter(textAfterAt.toLowerCase())
      } else {
        setShowMentions(false)
      }
    } else {
      setShowMentions(false)
    }
  }

  const insertMention = (employee: Employee) => {
    const lastAt = newComment.lastIndexOf('@')
    const newValue = newComment.substring(0, lastAt) + `@${employee.name} `
    setNewComment(newValue)
    setShowMentions(false)
    commentInputRef.current?.focus()
  }

  const filteredEmployees = employees.filter(emp =>
    emp.name.toLowerCase().includes(mentionFilter)
  )

  const handleSave = async () => {
    try {
      await fetch(`${API_URL}/tasks/${task.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          flag_color: formData.flag_color || null,
          assignee_id: formData.assignee_id || null,
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

  // Render comment with @mentions highlighted
  const renderCommentContent = (content: string) => {
    const parts = content.split(/(@\w+(?:\s\w+)?)/g)
    return parts.map((part, i) => {
      if (part.startsWith('@')) {
        return <span key={i} className="text-blue-600 font-medium bg-blue-50 px-0.5 rounded">{part}</span>
      }
      return part
    })
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl w-full max-w-2xl max-h-[90vh] overflow-hidden m-4 flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-2">
            {task.is_epic && <span className="text-xs bg-purple-100 text-purple-700 px-2 py-1 rounded">Эпик</span>}
            <span className={`text-xs px-2 py-1 rounded ${STATUS_CONFIG[task.status]?.color}`}>
              {STATUS_CONFIG[task.status]?.label}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <button onClick={() => setEditing(!editing)} className="p-2 text-gray-400 hover:text-gray-600">
              {editing ? <X size={20} /> : <MoreVertical size={20} />}
            </button>
            <button onClick={onClose} className="p-2 text-gray-400 hover:text-gray-600"><X size={20} /></button>
          </div>
        </div>

        {/* Content */}
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
              <div className="flex gap-3">
                <button onClick={() => setEditing(false)} className="px-4 py-2 border rounded-lg">Отмена</button>
                <button onClick={handleSave} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Сохранить</button>
                <button onClick={handleDelete} className="px-4 py-2 bg-red-600 text-white rounded-lg ml-auto">Удалить</button>
              </div>
            </div>
          ) : (
            <div>
              <h2 className="text-xl font-semibold mb-2">{task.title}</h2>
              {task.description && <p className="text-gray-600 mb-4">{task.description}</p>}

              <div className="grid grid-cols-2 gap-4 mb-6">
                <div>
                  <span className="text-sm text-gray-500">Исполнитель</span>
                  <p className="font-medium">{task.assignee?.name || 'Не назначен'}</p>
                </div>
                {task.due_date && (
                  <div>
                    <span className="text-sm text-gray-500">Срок</span>
                    <p className="font-medium">{new Date(task.due_date).toLocaleDateString('ru-RU')}</p>
                  </div>
                )}
                <div>
                  <span className="text-sm text-gray-500">Приоритет</span>
                  <p className={`font-medium ${PRIORITY_COLORS[task.priority]}`}>{PRIORITY_LABELS[task.priority]}</p>
                </div>
              </div>

              {/* Tabs */}
              <div className="border-t pt-4">
                <div className="flex gap-4 mb-4">
                  <button
                    onClick={() => setActiveTab('comments')}
                    className={`flex items-center gap-1.5 text-sm font-medium ${activeTab === 'comments' ? 'text-blue-600' : 'text-gray-500'}`}
                  >
                    <MessageSquare size={16} />
                    Комментарии ({taskComments.length})
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
                    {/* Comments list */}
                    {taskComments.map(comment => (
                      <div key={comment.id} className="p-3 bg-gray-50 rounded-lg">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-medium text-sm">{comment.author?.name || 'Аноним'}</span>
                          <span className="text-xs text-gray-400">
                            {new Date(comment.created_at).toLocaleString('ru-RU')}
                          </span>
                        </div>
                        <p className="text-gray-700">{renderCommentContent(comment.content)}</p>
                      </div>
                    ))}

                    {/* Comment input with @mentions */}
                    <div className="relative">
                      <div className="flex items-center gap-2 mb-2">
                        <AtSign size={16} className="text-gray-400" />
                        <span className="text-xs text-gray-500">Используйте @ для упоминания коллег</span>
                      </div>
                      <textarea
                        ref={commentInputRef}
                        value={newComment}
                        onChange={(e) => handleCommentChange(e.target.value)}
                        placeholder="Написать комментарий..."
                        className="w-full px-3 py-2 border rounded-lg resize-none"
                        rows={2}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' && !e.shiftKey) {
                            e.preventDefault()
                            handleAddComment()
                          }
                        }}
                      />

                      {/* Mentions dropdown */}
                      {showMentions && filteredEmployees.length > 0 && (
                        <div className="absolute bottom-full left-0 w-64 bg-white border rounded-lg shadow-lg mb-1 py-1 max-h-48 overflow-y-auto z-10">
                          {filteredEmployees.map(emp => (
                            <button
                              key={emp.id}
                              onClick={() => insertMention(emp)}
                              className="w-full flex items-center gap-2 px-3 py-2 hover:bg-gray-50 text-left"
                            >
                              <User size={16} className="text-gray-400" />
                              <span>{emp.name}</span>
                            </button>
                          ))}
                        </div>
                      )}

                      <div className="flex justify-end mt-2">
                        <button
                          onClick={handleAddComment}
                          disabled={!newComment.trim()}
                          className="flex items-center gap-1.5 px-4 py-2 bg-blue-600 text-white rounded-lg disabled:opacity-50"
                        >
                          <Send size={16} />
                          Отправить
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'history' && (
                  <div className="space-y-2">
                    {taskHistory.map(entry => (
                      <div key={entry.id} className="text-sm">
                        <span className="text-gray-400">{new Date(entry.created_at).toLocaleString('ru-RU')}</span>
                        {' — '}
                        <span className="font-medium">{entry.field_name}</span>
                        {' изменено с '}
                        <span className="text-red-500">{entry.old_value || '(пусто)'}</span>
                        {' на '}
                        <span className="text-green-500">{entry.new_value || '(пусто)'}</span>
                      </div>
                    ))}
                    {taskHistory.length === 0 && <p className="text-gray-400 text-sm">История пуста</p>}
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
