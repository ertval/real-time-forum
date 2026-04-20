import path from 'node:path';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	resolve: {
		alias: {
			'/static/js': path.resolve(__dirname, './web/static/js'),
		},
	},
	test: {
		globals: true,
		environment: 'node',
		include: ['**/*.test.{js,mjs,ts}'],
	},
});
