import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		// The app is a pure SPA (ssr = false in +layout.ts), so it builds to
		// static files. Every unknown path falls back to index.html, where the
		// client-side router takes over.
		adapter: adapter({ pages: 'build', assets: 'build', fallback: 'index.html' }),
		alias: {
			$lib: './src/lib'
		}
	}
};

export default config;
