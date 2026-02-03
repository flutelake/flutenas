import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	base: '/',
	server: {
		host: '0.0.0.0',
		port: 5173,
	},
	// build: {
	// 	sourcemap: true, // 或者使用 'hidden' 以避免在生产环境暴露源码
	// }
});
