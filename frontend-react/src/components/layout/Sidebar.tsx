import { NavLink, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  ListTodo,
  Layers,
  GitBranch,
  Rocket,
  TestTube,
  FileText,
  Github,
  Book,
  Calendar,
  Users,
  Building2,
  Clock,
  History,
  FileAudio,
  Brain,
  Plus,
  BarChart2,
} from 'lucide-react';
import { cn } from '@/lib/utils/cn';

interface SidebarSection {
  title: string;
  items: {
    to: string;
    icon: React.ElementType;
    label: string;
    badge?: string | number;
  }[];
}

const tasksSidebar: SidebarSection[] = [
  {
    title: 'Планирование',
    items: [
      { to: '/tasks', icon: LayoutDashboard, label: 'Доска задач' },
      { to: '/tasks/backlog', icon: ListTodo, label: 'Бэклог' },
      { to: '/tasks/roadmap', icon: Layers, label: 'Roadmap' },
    ],
  },
  {
    title: 'Спринты',
    items: [
      { to: '/sprints', icon: GitBranch, label: 'Все спринты' },
    ],
  },
  {
    title: 'Релизы',
    items: [
      { to: '/releases', icon: Rocket, label: 'Все релизы' },
    ],
  },
  {
    title: 'Тестирование',
    items: [
      { to: '/tests/plans', icon: TestTube, label: 'Тест-планы' },
      { to: '/tests/cases', icon: FileText, label: 'Тест-кейсы' },
    ],
  },
  {
    title: 'Интеграции',
    items: [
      { to: '/github', icon: Github, label: 'GitHub' },
      { to: '/confluence', icon: Book, label: 'Confluence' },
    ],
  },
];

const meetingsSidebar: SidebarSection[] = [
  {
    title: 'Календарь',
    items: [
      { to: '/calendar', icon: Calendar, label: 'Мой календарь' },
      { to: '/calendar/team', icon: Users, label: 'Команда' },
      { to: '/calendar/rooms', icon: Building2, label: 'Переговорные' },
    ],
  },
  {
    title: 'Встречи',
    items: [
      { to: '/meetings', icon: Clock, label: 'Предстоящие' },
      { to: '/meetings/past', icon: History, label: 'Прошедшие' },
    ],
  },
  {
    title: 'Анализ',
    items: [
      { to: '/meetings/transcriptions', icon: FileAudio, label: 'Транскрипции' },
      { to: '/meetings/analysis', icon: Brain, label: 'AI Анализ' },
    ],
  },
];

const serviceDeskSidebar: SidebarSection[] = [
  {
    title: 'Заявки',
    items: [
      { to: '/service-desk', icon: ListTodo, label: 'Мои заявки' },
      { to: '/service-desk/create', icon: Plus, label: 'Создать заявку' },
    ],
  },
  {
    title: 'Агент',
    items: [
      { to: '/service-desk/agent', icon: Users, label: 'Консоль агента' },
      { to: '/service-desk/stats', icon: BarChart2, label: 'Статистика' },
    ],
  },
];

function getSidebarForPath(pathname: string): SidebarSection[] | null {
  if (pathname.startsWith('/tasks') || pathname.startsWith('/sprints') || pathname.startsWith('/releases')) {
    return tasksSidebar;
  }
  if (pathname.startsWith('/meetings') || pathname.startsWith('/calendar')) {
    return meetingsSidebar;
  }
  if (pathname.startsWith('/service-desk')) {
    return serviceDeskSidebar;
  }
  return null;
}

export function Sidebar() {
  const location = useLocation();
  const sections = getSidebarForPath(location.pathname);

  if (!sections) {
    return null;
  }

  return (
    <aside className="w-60 bg-ekf-dark-lighter text-white fixed top-12 left-0 bottom-0 overflow-y-auto">
      <div className="p-4">
        {sections.map((section, idx) => (
          <div key={idx} className={cn('mb-6', idx > 0 && 'border-t border-white/10 pt-4')}>
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2 px-2">
              {section.title}
            </h3>
            <nav className="space-y-1">
              {section.items.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.to === '/tasks' || item.to === '/meetings' || item.to === '/service-desk'}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center gap-3 px-2 py-2 rounded-lg text-sm transition-colors',
                      isActive
                        ? 'bg-ekf-red text-white'
                        : 'text-gray-300 hover:bg-white/10 hover:text-white'
                    )
                  }
                >
                  <item.icon size={18} />
                  <span>{item.label}</span>
                  {item.badge && (
                    <span className="ml-auto bg-ekf-red text-white text-xs px-2 py-0.5 rounded-full">
                      {item.badge}
                    </span>
                  )}
                </NavLink>
              ))}
            </nav>
          </div>
        ))}
      </div>
    </aside>
  );
}
