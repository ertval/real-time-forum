(() => {
    const form = document.querySelector('form');
    const identifierEl = document.getElementById('email');
    const passwordEl = document.getElementById('password');
    const errorEl = document.getElementById('login-error');
    const toggleBtn = document.querySelector('.password-toggle');

    if (!form || !identifierEl || !passwordEl || !errorEl) return;

    const submitBtn = form.querySelector('button[type="submit"]');
    const LOGIN_API = form.getAttribute('action');

    function showError(msg) {
        errorEl.textContent = msg || '';
        errorEl.style.display = msg ? 'block' : 'none';
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        showError('');

        const identifier = identifierEl.value.trim();
        const password = passwordEl.value;

        if (!identifier || !password) {
            showError('Email/username and password are required.');
            return;
        }

        const payload = { password };
        identifier.includes('@')
            ? payload.email = identifier
            : payload.username = identifier;

        try {
            submitBtn.disabled = true;

            const res = await fetch(LOGIN_API, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(payload),
            });

            const data = await res.json().catch(() => null);

            if (!res.ok) {
                showError(
                    data?.error?.message ||
                    (res.status === 401 ? 'Invalid credentials.' : 'Login failed.')
                );
                return;
            }

            window.location.assign('/');
        } catch {
            showError('Network error. Please try again.');
        } finally {
            submitBtn.disabled = false;
        }
    });

    toggleBtn?.addEventListener('click', () => {
        const show = passwordEl.type === 'password';
        passwordEl.type = show ? 'text' : 'password';
        toggleBtn.setAttribute('aria-label', show ? 'Hide password' : 'Show password');
    });

    document.getElementById('guest-login-btn')?.addEventListener('click', () => {
        window.location.assign('/');
    });
})();
