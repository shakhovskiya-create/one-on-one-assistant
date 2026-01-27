import { useAuthStore } from '@/stores/auth';

export default function Dashboard() {
  const user = useAuthStore((state) => state.user);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">
          Добро пожаловать, {user?.name || 'Пользователь'}!
        </h1>
        <p className="text-gray-500 mt-1">
          EKF Hub — корпоративный портал
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="card p-6">
          <h3 className="text-sm font-medium text-gray-500">Задачи в работе</h3>
          <p className="mt-2 text-3xl font-bold text-gray-900">12</p>
        </div>
        <div className="card p-6">
          <h3 className="text-sm font-medium text-gray-500">Встречи сегодня</h3>
          <p className="mt-2 text-3xl font-bold text-gray-900">3</p>
        </div>
        <div className="card p-6">
          <h3 className="text-sm font-medium text-gray-500">Непрочитанных писем</h3>
          <p className="mt-2 text-3xl font-bold text-gray-900">8</p>
        </div>
        <div className="card p-6">
          <h3 className="text-sm font-medium text-gray-500">Активные заявки</h3>
          <p className="mt-2 text-3xl font-bold text-gray-900">2</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Последние задачи</h2>
          <p className="text-gray-500">Здесь будут отображаться последние задачи...</p>
        </div>
        <div className="card p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Предстоящие встречи</h2>
          <p className="text-gray-500">Здесь будут отображаться предстоящие встречи...</p>
        </div>
      </div>
    </div>
  );
}
