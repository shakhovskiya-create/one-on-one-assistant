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
			checkOrigin: false // Required for local development
		}
	}
};

export default config;
