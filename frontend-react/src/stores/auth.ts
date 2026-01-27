import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { auth as authApi } from '@/lib/api/client';
import type { User, AuthState } from '@/types';

interface AuthActions {
  login: (username: string, password: string) => Promise<{
    success: boolean;
    error?: string;
    forcePasswordChange?: boolean;
    userId?: string;
  }>;
  logout: () => Promise<void>;
  setUser: (user: User | null) => void;
  setSubordinates: (subordinates: User[]) => void;
  fetchSubordinates: (userId: string) => Promise<void>;
  changePassword: (userId: string, oldPassword: string, newPassword: string) => Promise<{
    success: boolean;
    error?: string;
  }>;
  refreshToken: () => Promise<boolean>;
  canAccessEmployee: (employeeId: string) => boolean;
}

type AuthStore = AuthState & AuthActions;

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // State
      user: null,
      token: null,
      isLoading: false,
      subordinates: [],

      // Actions
      login: async (username: string, password: string) => {
        set({ isLoading: true });
        try {
          const data = await authApi.login(username, password);

          if (data.authenticated && data.employee) {
            const user = data.employee;
            const token = data.token || 'authenticated';

            // Check if password change is required
            if (data.force_password_change) {
              set({ isLoading: false });
              return { success: true, forcePasswordChange: true, userId: user.id };
            }

            set({ user, token, isLoading: false });

            // Fetch subordinates in background
            get().fetchSubordinates(user.id);

            return { success: true };
          } else {
            set({ isLoading: false });
            return { success: false, error: data.error || 'Неверные учётные данные' };
          }
        } catch (error) {
          set({ isLoading: false });
          console.error('Login error:', error);
          return { success: false, error: 'Ошибка подключения к серверу' };
        }
      },

      logout: async () => {
        try {
          await authApi.logout();
        } catch {
          // Ignore errors, still clear local state
        }
        set({ user: null, token: null, subordinates: [] });
      },

      setUser: (user) => set({ user }),

      setSubordinates: (subordinates) => set({ subordinates }),

      fetchSubordinates: async (userId: string) => {
        try {
          const subordinates = await authApi.getSubordinates(userId);
          set({ subordinates: subordinates || [] });
        } catch (error) {
          console.error('Failed to fetch subordinates:', error);
        }
      },

      changePassword: async (userId: string, oldPassword: string, newPassword: string) => {
        try {
          const result = await authApi.changePassword(userId, oldPassword, newPassword);
          return { success: result.success };
        } catch (error) {
          console.error('Change password error:', error);
          return { success: false, error: 'Ошибка подключения к серверу' };
        }
      },

      refreshToken: async () => {
        try {
          const data = await authApi.refresh();
          if (data.token) {
            set({ token: data.token });
            return true;
          }
          return false;
        } catch (error) {
          console.error('Token refresh error:', error);
          return false;
        }
      },

      canAccessEmployee: (employeeId: string) => {
        const { user, subordinates } = get();
        if (!user) return false;
        if (employeeId === user.id) return true;
        return subordinates.some((sub) => sub.id === employeeId);
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
      }),
    }
  )
);

// Selectors
export const selectUser = (state: AuthStore) => state.user;
export const selectIsAuthenticated = (state: AuthStore) => !!state.user;
export const selectIsLoading = (state: AuthStore) => state.isLoading;
export const selectSubordinates = (state: AuthStore) => state.subordinates;
