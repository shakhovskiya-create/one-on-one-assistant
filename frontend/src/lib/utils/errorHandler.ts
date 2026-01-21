/**
 * Centralized error handling for the frontend
 */

export interface AppError {
	message: string;
	code?: string;
	status?: number;
	details?: unknown;
	timestamp: string;
}

export type ErrorSeverity = 'info' | 'warning' | 'error' | 'critical';

export interface ErrorHandlerOptions {
	showNotification?: boolean;
	logToConsole?: boolean;
	rethrow?: boolean;
}

const defaultOptions: ErrorHandlerOptions = {
	showNotification: true,
	logToConsole: true,
	rethrow: false
};

// Error notification callbacks
type ErrorCallback = (error: AppError, severity: ErrorSeverity) => void;
const errorCallbacks: ErrorCallback[] = [];

/**
 * Subscribe to error notifications
 */
export function onError(callback: ErrorCallback): () => void {
	errorCallbacks.push(callback);
	return () => {
		const index = errorCallbacks.indexOf(callback);
		if (index > -1) {
			errorCallbacks.splice(index, 1);
		}
	};
}

/**
 * Notify all error subscribers
 */
function notifyError(error: AppError, severity: ErrorSeverity): void {
	errorCallbacks.forEach(callback => {
		try {
			callback(error, severity);
		} catch (e) {
			console.error('Error in error callback:', e);
		}
	});
}

/**
 * Parse error from various sources into AppError
 */
export function parseError(error: unknown): AppError {
	const timestamp = new Date().toISOString();

	if (error instanceof Error) {
		return {
			message: error.message,
			details: (error as Record<string, unknown>).details,
			timestamp
		};
	}

	if (typeof error === 'string') {
		return {
			message: error,
			timestamp
		};
	}

	if (error && typeof error === 'object') {
		const obj = error as Record<string, unknown>;
		return {
			message: (obj.message as string) || (obj.error as string) || 'Unknown error',
			code: obj.code as string,
			status: obj.status as number,
			details: obj.details,
			timestamp
		};
	}

	return {
		message: 'An unexpected error occurred',
		timestamp
	};
}

/**
 * Determine error severity based on status code or error type
 */
export function getErrorSeverity(error: AppError): ErrorSeverity {
	if (error.status) {
		if (error.status >= 500) return 'critical';
		if (error.status >= 400) return 'error';
		if (error.status >= 300) return 'warning';
	}

	if (error.code) {
		if (error.code.startsWith('CRITICAL')) return 'critical';
		if (error.code.startsWith('WARN')) return 'warning';
	}

	return 'error';
}

/**
 * Handle an error with logging and optional notification
 */
export function handleError(
	error: unknown,
	context?: string,
	options: ErrorHandlerOptions = {}
): AppError {
	const opts = { ...defaultOptions, ...options };
	const appError = parseError(error);
	const severity = getErrorSeverity(appError);

	if (context) {
		appError.message = `${context}: ${appError.message}`;
	}

	if (opts.logToConsole) {
		const logMethod = severity === 'critical' || severity === 'error' ? 'error' : 'warn';
		console[logMethod](`[${severity.toUpperCase()}]`, appError);
	}

	if (opts.showNotification) {
		notifyError(appError, severity);
	}

	if (opts.rethrow) {
		throw error;
	}

	return appError;
}

/**
 * Wrap an async function with error handling
 */
export function withErrorHandling<T extends (...args: unknown[]) => Promise<unknown>>(
	fn: T,
	context?: string,
	options?: ErrorHandlerOptions
): T {
	return (async (...args: Parameters<T>) => {
		try {
			return await fn(...args);
		} catch (error) {
			handleError(error, context, options);
			return null;
		}
	}) as T;
}

/**
 * Retry a function with exponential backoff
 */
export async function withRetry<T>(
	fn: () => Promise<T>,
	maxRetries: number = 3,
	baseDelay: number = 1000
): Promise<T> {
	let lastError: unknown;

	for (let attempt = 0; attempt < maxRetries; attempt++) {
		try {
			return await fn();
		} catch (error) {
			lastError = error;

			if (attempt < maxRetries - 1) {
				const delay = baseDelay * Math.pow(2, attempt);
				await new Promise(resolve => setTimeout(resolve, delay));
			}
		}
	}

	throw lastError;
}

/**
 * User-friendly error messages for common HTTP status codes
 */
export function getHumanReadableError(status: number): string {
	const messages: Record<number, string> = {
		400: 'Некорректный запрос. Проверьте введённые данные.',
		401: 'Необходима авторизация. Пожалуйста, войдите в систему.',
		403: 'Доступ запрещён. У вас нет прав для выполнения этого действия.',
		404: 'Запрашиваемый ресурс не найден.',
		408: 'Время ожидания истекло. Попробуйте ещё раз.',
		429: 'Слишком много запросов. Подождите немного.',
		500: 'Внутренняя ошибка сервера. Попробуйте позже.',
		502: 'Сервер временно недоступен.',
		503: 'Сервис временно недоступен. Попробуйте позже.',
		504: 'Сервер не отвечает. Попробуйте позже.'
	};

	return messages[status] || 'Произошла ошибка. Попробуйте ещё раз.';
}
