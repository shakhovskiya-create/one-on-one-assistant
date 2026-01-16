import { writable } from 'svelte/store';
import type { Employee } from '$lib/api/client';

// Current user
export const currentUser = writable<Employee | null>(null);

// Auth state
export const isAuthenticated = writable(false);
export const authToken = writable<string | null>(null);

// Loading states
export const isLoading = writable(false);

// Notifications
export interface Notification {
	id: string;
	type: 'success' | 'error' | 'info' | 'warning';
	message: string;
	duration?: number;
}

function createNotificationStore() {
	const { subscribe, update } = writable<Notification[]>([]);

	return {
		subscribe,
		add: (notification: Omit<Notification, 'id'>) => {
			const id = Math.random().toString(36).substr(2, 9);
			const newNotification = { ...notification, id };

			update(n => [...n, newNotification]);

			if (notification.duration !== 0) {
				setTimeout(() => {
					update(n => n.filter(item => item.id !== id));
				}, notification.duration || 5000);
			}
		},
		remove: (id: string) => {
			update(n => n.filter(item => item.id !== id));
		},
		clear: () => {
			update(() => []);
		}
	};
}

export const notifications = createNotificationStore();

// Sidebar state
export const sidebarOpen = writable(true);
