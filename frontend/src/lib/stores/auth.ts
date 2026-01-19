import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';

// Use /api prefix which will be proxied to backend
const API_URL = '/api';

export interface User {
	id: string;
	name: string;
	email: string;
	position: string;
	department: string | null;
	manager_id: string | null;
	photo_base64?: string;
}

interface AuthState {
	user: User | null;
	token: string | null;
	isLoading: boolean;
	subordinates: User[];
}

function createAuthStore() {
	const initial: AuthState = {
		user: null,
		token: null,
		isLoading: true,
		subordinates: []
	};

	const { subscribe, set, update } = writable<AuthState>(initial);

	// Initialize from localStorage
	if (browser) {
		const savedToken = localStorage.getItem('auth_token');
		const savedUser = localStorage.getItem('auth_user');

		if (savedToken && savedUser) {
			try {
				const userData = JSON.parse(savedUser);
				update(state => ({ ...state, user: userData, token: savedToken, isLoading: false }));
				// Fetch subordinates
				fetchSubordinates(userData.id);
			} catch {
				localStorage.removeItem('auth_token');
				localStorage.removeItem('auth_user');
				update(state => ({ ...state, isLoading: false }));
			}
		} else {
			update(state => ({ ...state, isLoading: false }));
		}
	}

	async function fetchSubordinates(userId: string) {
		try {
			const authToken = browser ? localStorage.getItem('auth_token') : null;
			const res = await fetch(`${API_URL}/v1/ad/subordinates/${userId}`, {
				headers: {
					...(authToken && authToken !== 'authenticated' ? { 'Authorization': `Bearer ${authToken}` } : {})
				}
			});
			if (res.ok) {
				const data = await res.json();
				update(state => ({ ...state, subordinates: data || [] }));
			}
		} catch (error) {
			console.error('Failed to fetch subordinates:', error);
		}
	}

	async function login(username: string, password: string): Promise<{ success: boolean; error?: string; forcePasswordChange?: boolean; userId?: string }> {
		try {
			const formData = new FormData();
			formData.append('username', username);
			formData.append('password', password);

			const res = await fetch(`${API_URL}/v1/ad/authenticate`, {
				method: 'POST',
				body: formData
			});

			const data = await res.json();

			if (data.authenticated && data.employee) {
				const user = data.employee;
				const token = data.token || 'authenticated';

				// Check if password change is required
				if (data.force_password_change) {
					return { success: true, forcePasswordChange: true, userId: user.id };
				}

				update(state => ({ ...state, user, token, isLoading: false }));

				if (browser) {
					localStorage.setItem('auth_token', token);
					localStorage.setItem('auth_user', JSON.stringify(user));
					localStorage.setItem('currentUserId', user.id);
					// Save credentials for EWS services (mail, calendar)
					sessionStorage.setItem('ews_credentials', JSON.stringify({ username, password }));
				}

				fetchSubordinates(user.id);

				return { success: true };
			} else {
				return { success: false, error: data.error || 'Неверные учётные данные' };
			}
		} catch (error) {
			console.error('Login error:', error);
			return { success: false, error: 'Ошибка подключения к серверу' };
		}
	}

	async function changePassword(userId: string, oldPassword: string, newPassword: string): Promise<{ success: boolean; error?: string }> {
		try {
			const authToken = browser ? localStorage.getItem('auth_token') : null;
			const res = await fetch(`${API_URL}/v1/users/change-password`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					...(authToken && authToken !== 'authenticated' ? { 'Authorization': `Bearer ${authToken}` } : {})
				},
				body: JSON.stringify({
					user_id: userId,
					old_password: oldPassword,
					new_password: newPassword
				})
			});

			const data = await res.json();

			if (data.success) {
				return { success: true };
			} else {
				return { success: false, error: data.error || 'Не удалось изменить пароль' };
			}
		} catch (error) {
			console.error('Change password error:', error);
			return { success: false, error: 'Ошибка подключения к серверу' };
		}
	}

	async function refreshToken(): Promise<boolean> {
		try {
			const currentState = get({ subscribe });
			if (!currentState.token || currentState.token === 'authenticated') {
				return false;
			}

			const res = await fetch(`${API_URL}/v1/auth/refresh`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${currentState.token}`
				}
			});

			if (res.ok) {
				const data = await res.json();
				if (data.token) {
					update(state => ({ ...state, token: data.token }));
					if (browser) {
						localStorage.setItem('auth_token', data.token);
					}
					return true;
				}
			}
			return false;
		} catch (error) {
			console.error('Token refresh error:', error);
			return false;
		}
	}

	function logout() {
		update(state => ({ ...state, user: null, token: null, subordinates: [] }));

		if (browser) {
			localStorage.removeItem('auth_token');
			localStorage.removeItem('auth_user');
			localStorage.removeItem('currentUserId');
		}
	}

	function canAccessEmployee(employeeId: string): boolean {
		const state = get({ subscribe });
		if (!state.user) return false;
		if (employeeId === state.user.id) return true;
		return state.subordinates.some(sub => sub.id === employeeId);
	}

	return {
		subscribe,
		login,
		logout,
		canAccessEmployee,
		fetchSubordinates,
		changePassword,
		refreshToken
	};
}

export const auth = createAuthStore();

// Derived stores for convenience
export const user = derived(auth, $auth => $auth.user);
export const isAuthenticated = derived(auth, $auth => !!$auth.user);
export const isLoading = derived(auth, $auth => $auth.isLoading);
export const subordinates = derived(auth, $auth => $auth.subordinates);
