import { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Badge, PriorityBadge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Textarea } from '@/components/ui/Textarea';
import {
  Plus,
  Search,
  Filter,
  Calendar,
  MessageSquare,
  Paperclip,
} from 'lucide-react';

// Types
type TaskStatus = 'backlog' | 'todo' | 'in_progress' | 'review' | 'done';
type TaskPriority = 'critical' | 'high' | 'medium' | 'low';

interface Task {
  id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee?: { id: string; name: string; avatar?: string };
  dueDate?: string;
  tags?: string[];
  commentsCount?: number;
  attachmentsCount?: number;
  storyPoints?: number;
}

// Mock data
const mockTasks: Task[] = [
  {
    id: '1',
    title: 'Разработать API эндпоинты для задач',
    description: 'Создать REST API для CRUD операций с задачами',
    status: 'in_progress',
    priority: 'high',
    assignee: { id: '1', name: 'Иванов Иван' },
    dueDate: '2024-02-15',
    tags: ['backend', 'api'],
    commentsCount: 3,
    attachmentsCount: 1,
    storyPoints: 5,
  },
  {
    id: '2',
    title: 'Исправить критический баг авторизации',
    description: 'Пользователи не могут войти через SSO',
    status: 'in_progress',
    priority: 'critical',
    assignee: { id: '2', name: 'Петров Пётр' },
    dueDate: '2024-02-10',
    tags: ['bug', 'auth'],
    commentsCount: 5,
    storyPoints: 3,
  },
  {
    id: '3',
    title: 'Добавить фильтры в таблицу сотрудников',
    status: 'todo',
    priority: 'medium',
    assignee: { id: '3', name: 'Сидоров Сергей' },
    tags: ['frontend'],
    storyPoints: 2,
  },
  {
    id: '4',
    title: 'Написать тесты для WebSocket',
    status: 'review',
    priority: 'medium',
    assignee: { id: '1', name: 'Иванов Иван' },
    tags: ['testing'],
    storyPoints: 3,
  },
  {
    id: '5',
    title: 'Обновить документацию API',
    status: 'done',
    priority: 'low',
    assignee: { id: '4', name: 'Козлова Мария' },
    tags: ['docs'],
    storyPoints: 1,
  },
  {
    id: '6',
    title: 'Настроить CI/CD пайплайн',
    status: 'backlog',
    priority: 'high',
    tags: ['devops'],
    storyPoints: 8,
  },
  {
    id: '7',
    title: 'Провести код-ревью PR #234',
    status: 'todo',
    priority: 'high',
    assignee: { id: '2', name: 'Петров Пётр' },
    dueDate: '2024-02-11',
    storyPoints: 1,
  },
  {
    id: '8',
    title: 'Оптимизировать запросы к базе данных',
    status: 'backlog',
    priority: 'medium',
    tags: ['performance', 'database'],
    storyPoints: 5,
  },
];

const columns: { id: TaskStatus; title: string; color: string }[] = [
  { id: 'backlog', title: 'Бэклог', color: 'bg-gray-100' },
  { id: 'todo', title: 'К выполнению', color: 'bg-blue-100' },
  { id: 'in_progress', title: 'В работе', color: 'bg-yellow-100' },
  { id: 'review', title: 'На ревью', color: 'bg-purple-100' },
  { id: 'done', title: 'Готово', color: 'bg-green-100' },
];

// Task Card Component
function TaskCard({ task, onClick }: { task: Task; onClick: () => void }) {
  return (
    <div
      onClick={onClick}
      className="bg-white rounded-lg border border-gray-200 p-3 shadow-sm hover:shadow-md transition-shadow cursor-pointer"
    >
      {/* Priority & Tags */}
      <div className="flex items-center gap-2 mb-2 flex-wrap">
        <PriorityBadge priority={task.priority} size="sm" />
        {task.tags?.slice(0, 2).map((tag) => (
          <Badge key={tag} variant="outline" size="sm">
            {tag}
          </Badge>
        ))}
      </div>

      {/* Title */}
      <h4 className="text-sm font-medium text-gray-900 mb-2 line-clamp-2">
        {task.title}
      </h4>

      {/* Meta */}
      <div className="flex items-center justify-between text-xs text-gray-500">
        <div className="flex items-center gap-3">
          {task.commentsCount !== undefined && task.commentsCount > 0 && (
            <span className="flex items-center gap-1">
              <MessageSquare className="h-3 w-3" />
              {task.commentsCount}
            </span>
          )}
          {task.attachmentsCount !== undefined && task.attachmentsCount > 0 && (
            <span className="flex items-center gap-1">
              <Paperclip className="h-3 w-3" />
              {task.attachmentsCount}
            </span>
          )}
          {task.storyPoints !== undefined && (
            <span className="flex items-center justify-center w-5 h-5 rounded bg-gray-100 text-gray-600 font-medium">
              {task.storyPoints}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {task.dueDate && (
            <span className="flex items-center gap-1 text-gray-400">
              <Calendar className="h-3 w-3" />
              {new Date(task.dueDate).toLocaleDateString('ru-RU', {
                day: 'numeric',
                month: 'short',
              })}
            </span>
          )}
          {task.assignee && (
            <Avatar name={task.assignee.name} size="xs" />
          )}
        </div>
      </div>
    </div>
  );
}

// Kanban Column Component
function KanbanColumn({
  column,
  tasks,
  onTaskClick,
}: {
  column: { id: TaskStatus; title: string; color: string };
  tasks: Task[];
  onTaskClick: (task: Task) => void;
}) {
  const totalPoints = tasks.reduce((sum, t) => sum + (t.storyPoints || 0), 0);

  return (
    <div className="flex flex-col min-w-[300px] w-[300px]">
      {/* Column Header */}
      <div className={`rounded-t-lg px-3 py-2 ${column.color}`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h3 className="font-medium text-gray-700">{column.title}</h3>
            <span className="flex items-center justify-center w-5 h-5 rounded-full bg-white text-xs font-medium text-gray-600">
              {tasks.length}
            </span>
          </div>
          <span className="text-xs text-gray-500">{totalPoints} SP</span>
        </div>
      </div>

      {/* Tasks */}
      <div className="flex-1 bg-gray-50 rounded-b-lg p-2 space-y-2 min-h-[200px]">
        {tasks.map((task) => (
          <TaskCard key={task.id} task={task} onClick={() => onTaskClick(task)} />
        ))}
        {tasks.length === 0 && (
          <div className="flex items-center justify-center h-20 text-sm text-gray-400">
            Нет задач
          </div>
        )}
      </div>
    </div>
  );
}

// Task Detail Modal
function TaskDetailModal({
  task,
  isOpen,
  onClose,
}: {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
}) {
  if (!task) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={task.title} size="lg">
      <div className="space-y-6">
        {/* Status & Priority */}
        <div className="flex items-center gap-4">
          <div>
            <label className="text-xs text-gray-500 block mb-1">Статус</label>
            <Badge variant="info">
              {columns.find((c) => c.id === task.status)?.title}
            </Badge>
          </div>
          <div>
            <label className="text-xs text-gray-500 block mb-1">Приоритет</label>
            <PriorityBadge priority={task.priority} />
          </div>
          {task.storyPoints !== undefined && (
            <div>
              <label className="text-xs text-gray-500 block mb-1">Story Points</label>
              <Badge variant="default">{task.storyPoints}</Badge>
            </div>
          )}
        </div>

        {/* Description */}
        {task.description && (
          <div>
            <label className="text-xs text-gray-500 block mb-1">Описание</label>
            <p className="text-gray-700">{task.description}</p>
          </div>
        )}

        {/* Assignee & Due Date */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-xs text-gray-500 block mb-1">Исполнитель</label>
            {task.assignee ? (
              <div className="flex items-center gap-2">
                <Avatar name={task.assignee.name} size="sm" />
                <span className="text-sm text-gray-700">{task.assignee.name}</span>
              </div>
            ) : (
              <span className="text-sm text-gray-400">Не назначен</span>
            )}
          </div>
          <div>
            <label className="text-xs text-gray-500 block mb-1">Срок</label>
            {task.dueDate ? (
              <span className="text-sm text-gray-700">
                {new Date(task.dueDate).toLocaleDateString('ru-RU')}
              </span>
            ) : (
              <span className="text-sm text-gray-400">Не указан</span>
            )}
          </div>
        </div>

        {/* Tags */}
        {task.tags && task.tags.length > 0 && (
          <div>
            <label className="text-xs text-gray-500 block mb-1">Теги</label>
            <div className="flex flex-wrap gap-2">
              {task.tags.map((tag) => (
                <Badge key={tag} variant="outline">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
}

// Create Task Modal
function CreateTaskModal({
  isOpen,
  onClose,
}: {
  isOpen: boolean;
  onClose: () => void;
}) {
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Создать задачу"
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Отмена
          </Button>
          <Button variant="primary" onClick={onClose}>
            Создать
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <Input label="Название" placeholder="Введите название задачи" />
        <Textarea label="Описание" placeholder="Описание задачи..." rows={3} />
        <div className="grid grid-cols-2 gap-4">
          <Select
            label="Приоритет"
            options={[
              { value: 'critical', label: 'Критический' },
              { value: 'high', label: 'Высокий' },
              { value: 'medium', label: 'Средний' },
              { value: 'low', label: 'Низкий' },
            ]}
          />
          <Select
            label="Статус"
            options={[
              { value: 'backlog', label: 'Бэклог' },
              { value: 'todo', label: 'К выполнению' },
              { value: 'in_progress', label: 'В работе' },
              { value: 'review', label: 'На ревью' },
              { value: 'done', label: 'Готово' },
            ]}
          />
        </div>
        <Input label="Story Points" type="number" placeholder="0" />
      </div>
    </Modal>
  );
}

export default function Tasks() {
  const [tasks] = useState<Task[]>(mockTasks);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const filteredTasks = tasks.filter((task) =>
    task.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getTasksByStatus = (status: TaskStatus) =>
    filteredTasks.filter((task) => task.status === status);

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Задачи</h1>
          <p className="text-gray-500 mt-1">Sprint 10 · 2 недели осталось</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input
              type="text"
              placeholder="Поиск задач..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ekf-red focus:border-transparent w-64"
            />
          </div>
          <Button variant="outline" leftIcon={<Filter className="h-4 w-4" />}>
            Фильтры
          </Button>
          <Button
            variant="primary"
            leftIcon={<Plus className="h-4 w-4" />}
            onClick={() => setIsCreateModalOpen(true)}
          >
            Создать задачу
          </Button>
        </div>
      </div>

      {/* Kanban Board */}
      <div className="flex-1 overflow-x-auto">
        <div className="flex gap-4 pb-4 min-h-full">
          {columns.map((column) => (
            <KanbanColumn
              key={column.id}
              column={column}
              tasks={getTasksByStatus(column.id)}
              onTaskClick={setSelectedTask}
            />
          ))}
        </div>
      </div>

      {/* Modals */}
      <TaskDetailModal
        task={selectedTask}
        isOpen={!!selectedTask}
        onClose={() => setSelectedTask(null)}
      />
      <CreateTaskModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />
    </div>
  );
}
