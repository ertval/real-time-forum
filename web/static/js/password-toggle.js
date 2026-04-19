// web/static/js/password-toggle.js

export function initPasswordToggles(root = document) {
	const toggles = root.querySelectorAll('.password-toggle');
	if (!toggles.length) return;

	toggles.forEach((btn) => {
		if (btn.dataset.bound) return;
		btn.dataset.bound = '1';

		btn.addEventListener('click', () => {
			const form = btn.closest('form');
			if (!form) return;

			const inputs = form.querySelectorAll(
				'input[type="password"], input[type="text"][data-password]',
			);

			if (!inputs.length) return;

			const shouldShow = inputs[0].type === 'password';

			inputs.forEach((input) => {
				input.type = shouldShow ? 'text' : 'password';
				input.dataset.password = '1'; // marker
			});

			form.querySelectorAll('.password-toggle').forEach((eye) => {
				eye.setAttribute('aria-label', shouldShow ? 'Hide password' : 'Show password');
				eye.title = shouldShow ? 'Hide password' : 'Show password';
			});
		});
	});
}
