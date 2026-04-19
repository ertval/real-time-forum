// web/static/js/header-loader.js

import { Auth } from './auth.js';
import { closeAuthModal, loadAuthModal } from './auth-modal.js';
import { initHeader } from './header.js';
import { initSettings } from './settings.js';

function applyAuthUI() {
	if (Auth.isAuthenticated) {
		document.body.classList.add('is-authenticated');
		Auth.startSessionWatcher();
	} else {
		document.body.classList.remove('is-authenticated');
	}
}

async function initialize() {
	await loadAuthModal();
	await Auth.init();
	applyAuthUI();
	initSettings();
	initHeader();
}

initialize();

/* Login via modal */
window.addEventListener('message', async (e) => {
	if (e.data !== 'auth:success') return;

	closeAuthModal();

	Auth.checked = false;
	await Auth.init();

	applyAuthUI();
	initHeader();
});
