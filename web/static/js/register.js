// web/static/js/register.js
import { initPasswordToggles } from './password-toggle.js';
import { uiNotify } from './ui-messages.js';

(() => {
	const form = document.querySelector('form');
	if (!form) return;

	const usernameEl = document.getElementById('username');
	const emailEl = document.getElementById('email');
	const passwordEl = document.getElementById('password');
	const confirmEl = document.getElementById('confirm_password');

	if (!usernameEl || !emailEl || !passwordEl || !confirmEl) return;

	initPasswordToggles(document);

	const submitBtn = form.querySelector('button[type="submit"]');

	function notify(message, type = 'danger') {
		if (window.parent && window.parent !== window) {
			window.parent.postMessage(
				{
					type: 'auth:notify',
					payload: { message, level: type },
				},
				'*',
			);
		} else {
			uiNotify(message, { type });
		}
	}

	function getRegistrationError(username, email, password, confirm) {
		if (!username || !email || !password || !confirm) {
			return 'All fields are required.';
		}

		if (password.length < 8) {
			return 'Password must be at least 8 characters long.';
		}

		if (password !== confirm) {
			return 'Passwords do not match.';
		}

		return '';
	}

	form.addEventListener('submit', async (e) => {
		e.preventDefault();

		const username = usernameEl.value.trim();
		const email = emailEl.value.trim();
		const password = passwordEl.value;
		const confirm = confirmEl.value;

		const error = getRegistrationError(username, email, password, confirm);
		if (error) {
			notify(error, 'warn');
			return;
		}

		try {
			submitBtn.disabled = true;

			const res = await fetch('/api/v1/users/register', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Accept: 'application/json',
				},
				credentials: 'include',
				body: JSON.stringify({ username, email, password }),
			});

			const data = await res.json().catch(() => null);

			if (!res.ok) {
				notify(data?.error?.message || 'Registration failed. Please try again.', 'danger');
				return;
			}

			sessionStorage.setItem('auth:login-success', '1');

			if (window.parent && window.parent !== window) {
				window.parent.postMessage('auth:success', '*');
				return;
			}

			window.location.assign('/');
		} catch {
			notify('Network error. Please try again.', 'danger');
		} finally {
			submitBtn.disabled = false;
		}
	});
})();
