import type { Handle } from '@sveltejs/kit';

const BACKEND_URL = process.env.BACKEND_URL || 'http://backend:8080';

export const handle: Handle = async ({ event, resolve }) => {
	// Proxy /api requests to backend
	if (event.url.pathname.startsWith('/api/')) {
		// Remove /api prefix and forward to backend
		// /api/v1/employees -> /api/v1/employees (keep /api/v1 for protected routes)
		// /api/employees -> /employees (legacy routes)

		let backendPath = event.url.pathname;

		// If it's /api/v1/... keep as is (protected routes)
		// If it's /api/... without v1, strip /api (legacy routes)
		if (event.url.pathname.startsWith('/api/v1/')) {
			// Protected route - forward as /api/v1/...
			backendPath = event.url.pathname;
		} else {
			// Legacy route - strip /api
			backendPath = event.url.pathname.replace(/^\/api/, '');
		}

		const backendUrl = `${BACKEND_URL}${backendPath}${event.url.search}`;

		console.log(`[Proxy] ${event.url.pathname} -> ${backendUrl}`);

		try {
			const headers = new Headers();
			// Forward relevant headers
			const authHeader = event.request.headers.get('Authorization');
			if (authHeader) {
				headers.set('Authorization', authHeader);
			}
			const contentType = event.request.headers.get('Content-Type');
			if (contentType) {
				headers.set('Content-Type', contentType);
			}

			let body: BodyInit | null = null;
			if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
				// For file uploads, pass body as-is
				if (contentType?.includes('multipart/form-data')) {
					body = await event.request.arrayBuffer();
					// Don't set Content-Type for multipart - let fetch set it with boundary
					headers.delete('Content-Type');
				} else {
					body = await event.request.text();
				}
			}

			const response = await fetch(backendUrl, {
				method: event.request.method,
				headers,
				body
			});

			// Return the response as-is
			return new Response(response.body, {
				status: response.status,
				statusText: response.statusText,
				headers: {
					'Content-Type': response.headers.get('Content-Type') || 'application/json'
				}
			});
		} catch (error) {
			console.error('Proxy error:', error);
			return new Response(JSON.stringify({ error: 'Backend unavailable', details: String(error) }), {
				status: 502,
				headers: { 'Content-Type': 'application/json' }
			});
		}
	}

	return resolve(event);
};
