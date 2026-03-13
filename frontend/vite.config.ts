import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/api': 'http://localhost:9705',
			'/ws': {
				target: 'ws://localhost:9705',
				ws: true
			}
		}
	},
	resolve: {
		conditions: ['browser']
	},
	test: {
		environment: 'jsdom',
		include: ['src/**/*.test.ts'],
		setupFiles: ['src/test-setup.ts']
	}
});
