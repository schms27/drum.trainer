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
		// Check for a new deployment every 5 minutes. `updated` then shows the
		// reload notice, and the next link click does a full page load.
		version: { pollInterval: 5 * 60 * 1000 },
		alias: {
			$lib: './src/lib'
		}
	}
};

export default config;
