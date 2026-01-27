import { useAuthStore } from '@/stores/auth';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge, StatusBadge, PriorityBadge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import {
  CheckSquare,
  Calendar,
  Mail,
  Ticket,
  TrendingUp,
  Clock,
  Users,
  ArrowRight,
} from 'lucide-react';
import { Link } from 'react-router-dom';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  trend?: { value: number; isPositive: boolean };
  href?: string;
}

function StatCard({ title, value, icon, trend, href }: StatCardProps) {
  const content = (
    <Card className="hover:shadow-md transition-shadow">
      <div className="p-6">
        <div className="flex items-center justify-between">
          <div className="p-2 rounded-lg bg-ekf-red/10 text-ekf-red">
            {icon}
          </div>
          {trend && (
            <div
              className={`flex items-center text-sm ${
                trend.isPositive ? 'text-green-600' : 'text-red-600'
              }`}
            >
              <TrendingUp
                className={`h-4 w-4 mr-1 ${!trend.isPositive && 'rotate-180'}`}
              />
              {trend.value}%
            </div>
          )}
        </div>
        <div className="mt-4">
          <p className="text-3xl font-bold text-gray-900">{value}</p>
          <p className="text-sm text-gray-500 mt-1">{title}</p>
        </div>
      </div>
    </Card>
  );

  if (href) {
    return <Link to={href}>{content}</Link>;
  }

  return content;
}

// Mock data for tasks
const recentTasks = [
  { id: 1, title: 'Обновить документацию API', status: 'in_progress' as const, priority: 'high' as const, assignee: 'Иванов И.' },
  { id: 2, title: 'Исправить баг авторизации', status: 'review' as const, priority: 'critical' as const, assignee: 'Петров П.' },
  { id: 3, title: 'Добавить фильтры в отчёты', status: 'open' as const, priority: 'medium' as const, assignee: 'Сидоров С.' },
  { id: 4, title: 'Оптимизировать запросы к БД', status: 'done' as const, priority: 'low' as const, assignee: 'Козлов К.' },
];

// Mock data for meetings
const upcomingMeetings = [
  { id: 1, title: 'Стендап команды', time: '10:00', participants: ['Иванов И.', 'Петров П.', 'Сидоров С.'] },
  { id: 2, title: 'Ревью спринта', time: '14:00', participants: ['Вся команда'] },
  { id: 3, title: '1:1 с руководителем', time: '16:30', participants: ['Козлов К.'] },
];

export default function Dashboard() {
  const user = useAuthStore((state) => state.user);

  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Добро пожаловать, {user?.name || 'Пользователь'}!
          </h1>
          <p className="text-gray-500 mt-1">
            Вот что происходит в вашем рабочем пространстве
          </p>
        </div>
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <Clock className="h-4 w-4" />
          {new Date().toLocaleDateString('ru-RU', {
            weekday: 'long',
            day: 'numeric',
            month: 'long',
          })}
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Задачи в работе"
          value={12}
          icon={<CheckSquare className="h-6 w-6" />}
          href="/tasks"
        />
        <StatCard
          title="Встречи сегодня"
          value={3}
          icon={<Calendar className="h-6 w-6" />}
          trend={{ value: 15, isPositive: true }}
          href="/meetings"
        />
        <StatCard
          title="Непрочитанных писем"
          value={8}
          icon={<Mail className="h-6 w-6" />}
          href="/mail"
        />
        <StatCard
          title="Активные заявки"
          value={2}
          icon={<Ticket className="h-6 w-6" />}
          href="/service-desk"
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Tasks */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle>Последние задачи</CardTitle>
            <Link
              to="/tasks"
              className="text-sm text-ekf-red hover:text-ekf-red-dark flex items-center gap-1"
            >
              Все задачи
              <ArrowRight className="h-4 w-4" />
            </Link>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {recentTasks.map((task) => (
                <div
                  key={task.id}
                  className="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {task.title}
                    </p>
                    <p className="text-xs text-gray-500 mt-0.5">{task.assignee}</p>
                  </div>
                  <div className="flex items-center gap-2 ml-4">
                    <PriorityBadge priority={task.priority} size="sm" />
                    <StatusBadge status={task.status} size="sm" />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Upcoming Meetings */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle>Предстоящие встречи</CardTitle>
            <Link
              to="/meetings"
              className="text-sm text-ekf-red hover:text-ekf-red-dark flex items-center gap-1"
            >
              Все встречи
              <ArrowRight className="h-4 w-4" />
            </Link>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {upcomingMeetings.map((meeting) => (
                <div
                  key={meeting.id}
                  className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                >
                  <div className="flex items-center justify-center w-12 h-12 bg-ekf-red/10 text-ekf-red rounded-lg font-semibold text-sm">
                    {meeting.time}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900">
                      {meeting.title}
                    </p>
                    <div className="flex items-center gap-1 mt-1">
                      <Users className="h-3 w-3 text-gray-400" />
                      <p className="text-xs text-gray-500">
                        {meeting.participants.join(', ')}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Team Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Активность команды</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="flex -space-x-3">
              <Avatar name="Иванов Иван" size="md" showStatus status="online" className="ring-2 ring-white" />
              <Avatar name="Петров Пётр" size="md" showStatus status="online" className="ring-2 ring-white" />
              <Avatar name="Сидоров Сергей" size="md" showStatus status="away" className="ring-2 ring-white" />
              <Avatar name="Козлов Константин" size="md" showStatus status="offline" className="ring-2 ring-white" />
              <div className="flex items-center justify-center w-10 h-10 rounded-full bg-gray-100 text-gray-600 text-sm font-medium ring-2 ring-white">
                +5
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="success">3 онлайн</Badge>
              <Badge variant="default">2 заняты</Badge>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
