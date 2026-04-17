// SPA/main.js

import { createApp } from './core/app/create-app.js';
import { matchRoute, normalizePathname } from './core/router/routes.js';

export { createApp, matchRoute, normalizePathname };

const app = createApp();

if (app) {
	if (document.readyState === 'loading') {
		document.addEventListener(
			'DOMContentLoaded',
			() => {
				void app.boot();
			},
			{ once: true },
		);
	} else {
		void app.boot();
	}
}
