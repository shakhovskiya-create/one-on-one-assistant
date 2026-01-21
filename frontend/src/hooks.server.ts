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
		// If it's /api/admin/... keep as is (admin routes)
		// If it's /api/... without v1 or admin, strip /api (legacy routes)
		if (event.url.pathname.startsWith('/api/v1/') || event.url.pathname.startsWith('/api/admin')) {
			// Protected route - forward as is
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
				// For file uploads, we need to forward the exact body and Content-Type with boundary
				if (contentType?.includes('multipart/form-data')) {
					body = await event.request.arrayBuffer();
					// IMPORTANT: Keep the Content-Type header as-is - it contains the boundary
					// The contentType was already set above, so it includes the boundary
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
