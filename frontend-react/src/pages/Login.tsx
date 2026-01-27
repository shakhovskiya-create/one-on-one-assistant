import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import { User, Lock, AlertCircle } from 'lucide-react';

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const login = useAuthStore((state) => state.login);
  const navigate = useNavigate();
  const location = useLocation();

  // Redirect to previous page or dashboard after login
  const from = (location.state as { from?: Location })?.from?.pathname || '/';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!username.trim()) {
      setError('Введите логин');
      return;
    }

    if (!password) {
      setError('Введите пароль');
      return;
    }

    setIsLoading(true);

    try {
      const result = await login(username.trim(), password);

      if (result.success) {
        navigate(from, { replace: true });
      } else {
        setError(result.error || 'Неверный логин или пароль');
      }
    } catch {
      setError('Ошибка соединения с сервером');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-100 to-gray-200 p-4">
      <div className="w-full max-w-md">
        {/* Logo and Title */}
        <div className="text-center mb-8">
          <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-ekf-red to-ekf-red-dark flex items-center justify-center mx-auto mb-4 shadow-lg">
            <span className="text-white font-bold text-3xl">EKF</span>
          </div>
          <h1 className="text-3xl font-bold text-gray-900">EKF Hub</h1>
          <p className="text-gray-500 mt-2">Корпоративный портал</p>
        </div>

        {/* Login Card */}
        <Card variant="elevated" padding="lg">
          <form onSubmit={handleSubmit} className="space-y-5">
            <Input
              label="Логин"
              name="username"
              type="text"
              value={username}
              onChange={(e) => {
                setUsername(e.target.value);
                setError('');
              }}
              placeholder="Введите ваш логин AD"
              leftIcon={<User className="h-5 w-5" />}
              disabled={isLoading}
              autoComplete="username"
              autoFocus
            />

            <Input
              label="Пароль"
              name="password"
              type="password"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                setError('');
              }}
              placeholder="Введите пароль"
              leftIcon={<Lock className="h-5 w-5" />}
              disabled={isLoading}
              autoComplete="current-password"
            />

            {/* Error Message */}
            {error && (
              <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                <AlertCircle className="h-5 w-5 flex-shrink-0" />
                <span>{error}</span>
              </div>
            )}

            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isLoading}
              className="w-full"
            >
              {isLoading ? 'Вход...' : 'Войти'}
            </Button>
          </form>

          {/* Help Text */}
          <div className="mt-6 pt-6 border-t border-gray-200 text-center">
            <p className="text-sm text-gray-500">
              Используйте учётные данные Active Directory
            </p>
          </div>
        </Card>

        {/* Footer */}
        <div className="mt-8 text-center text-sm text-gray-400">
          <p>© {new Date().getFullYear()} EKF Group. Все права защищены.</p>
        </div>
      </div>
    </div>
  );
}
