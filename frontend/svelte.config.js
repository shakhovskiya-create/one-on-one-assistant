import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			out: 'build'
		}),
		paths: {
			base: '',
			assets: ''
		},
		csrf: {
			checkOrigin: false
		}
	}
};

// Note: If CSRF issues persist, the problem may be in form handling.
// All API requests go through client.ts which uses fetch with JSON body,
// not native form submissions. CSRF typically affects form submissions.

export default config;
