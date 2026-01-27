import { NavLink } from 'react-router-dom';
import {
  Home,
  Users,
  CheckSquare,
  Calendar,
  Mail,
  MessageCircle,
  Headphones,
  BarChart2,
  LogOut,
  User,
} from 'lucide-react';
import { cn } from '@/lib/utils/cn';
import { useAuthStore } from '@/stores/auth';

const navItems = [
  { to: '/', icon: Home, label: 'Главная' },
  { to: '/employees', icon: Users, label: 'Сотрудники' },
  { to: '/tasks', icon: CheckSquare, label: 'Задачи' },
  { to: '/meetings', icon: Calendar, label: 'Встречи' },
  { to: '/mail', icon: Mail, label: 'Почта' },
  { to: '/messenger', icon: MessageCircle, label: 'Сообщения' },
  { to: '/service-desk', icon: Headphones, label: 'SD' },
  { to: '/analytics', icon: BarChart2, label: 'Аналитика' },
];

export function GlobalNav() {
  const { user, logout } = useAuthStore();

  return (
    <header className="h-12 bg-ekf-dark text-white flex items-center px-4 fixed top-0 left-0 right-0 z-50">
      {/* Logo */}
      <div className="flex items-center gap-2 mr-8">
        <div className="w-8 h-8 rounded bg-ekf-red flex items-center justify-center font-bold text-sm">
          EKF
        </div>
        <span className="font-semibold text-lg hidden md:block">Hub</span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 flex items-center gap-1">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
                isActive
                  ? 'bg-ekf-red text-white'
                  : 'text-gray-300 hover:bg-white/10 hover:text-white'
              )
            }
          >
            <item.icon size={18} />
            <span className="hidden lg:block">{item.label}</span>
          </NavLink>
        ))}
      </nav>

      {/* User Menu */}
      <div className="flex items-center gap-4">
        {user && (
          <NavLink
            to="/profile"
            className="flex items-center gap-2 text-sm text-gray-300 hover:text-white"
          >
            {user.photo_base64 ? (
              <img
                src={`data:image/jpeg;base64,${user.photo_base64}`}
                alt={user.name}
                className="w-8 h-8 rounded-full object-cover"
              />
            ) : (
              <div className="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center">
                <User size={16} />
              </div>
            )}
            <span className="hidden md:block">{user.name}</span>
          </NavLink>
        )}
        <button
          onClick={() => logout()}
          className="p-2 text-gray-400 hover:text-white rounded-lg hover:bg-white/10"
          title="Выйти"
        >
          <LogOut size={18} />
        </button>
      </div>
    </header>
  );
}
