// web/static/js/auth-modal.js
import { uiNotify } from './ui-messages.js';

let modalLoaded = false;
let modal = null;
let frame = null;
let isOpen = false;

export async function loadAuthModal() {
	if (modalLoaded) return;

	const res = await fetch('/static/partials/auth-modal.html');
	const html = await res.text();

	document.body.insertAdjacentHTML('beforeend', html);

	modal = document.getElementById('auth-modal');
	frame = document.getElementById('auth-frame');

	modal.querySelector('.auth-close').onclick = modal.querySelector('.auth-backdrop').onclick =
		closeAuthModal;

	modalLoaded = true;
}

export async function openAuthModal(path = '/login') {
	if (isOpen) return;

	if (!modalLoaded) {
		await loadAuthModal();
	}

	frame.style.opacity = '0';
	frame.src = path;

	frame.onload = () => {
		frame.style.opacity = '1';
		modal.classList.add('visible');
	};

	document.body.style.overflow = 'hidden';
	isOpen = true;
}

export function closeAuthModal() {
	if (!modal) return;

	modal.classList.remove('visible');
	document.body.style.overflow = '';
	isOpen = false;
}

/* =========================
   IFRAME → PARENT MESSAGES
========================= */

window.addEventListener('message', (event) => {
	if (!event?.data) return;

	if (event.data.type === 'auth:notify') {
		const { message, level } = event.data.payload || {};
		if (message) {
			uiNotify(message, { type: level || 'info' });
		}
		return;
	}

	if (event.data === 'auth:success') {
		closeAuthModal();
		window.location.href = '/';
	}
});
