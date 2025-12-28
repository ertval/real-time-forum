(() => {
    const form = document.querySelector('form');
    const emailEl = document.getElementById('email');
    const passwordEl = document.getElementById('password');
    const errorEl = document.getElementById('login-error');



    if (!form || !emailEl || !passwordEl || !errorEl) {
        // Either the script is included on the wrong page,
        // or the HTML changed.
        return;
    }

    const submitBtn = form.querySelector('button[type="submit"]');
    const toggleBtn = document.querySelector('.password-toggle');


    function showError(message) {
        if (!message) {
            errorEl.textContent = '';
            errorEl.style.display = 'none';
            return;
        }
        errorEl.textContent = message;
        errorEl.style.display = 'block';
    }

    const LOGIN_API = form.getAttribute('action'); // "/api/v1/users/login"

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        showError('');

        const identifier = emailEl.value.trim();
        const password = passwordEl.value;

        if (!identifier || !password) {
            showError('Email or username and password are required.');
            return;
        }

        const body = {password};

        // Decide which field to send
        if (identifier.includes('@')) {
            body.email = identifier;
        } else {
            body.username = identifier;
        }

        try {
            submitBtn && (submitBtn.disabled = true);
            const res = await fetch(LOGIN_API, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(body),
            });

            const payload = await res.json().catch(() => null);

            if (!res.ok) {
                // APIResponse{ Error: err }
                showError(
                    payload?.error?.message ||
                    (res.status === 401 ? 'Invalid credentials.' : 'Login failed.')
                );
                return;
            }

            // Success → session cookie is set
            window.location.assign('/home');
        } catch {
            showError('Network error. Please try again.');
        }  finally {
            submitBtn && (submitBtn.disabled = false);
        }
    });

    toggleBtn?.addEventListener('click', () => {
        const show = passwordEl.type === 'password';
        passwordEl.type = show ? 'text' : 'password';
        toggleBtn.setAttribute('aria-label', show ? 'Hide password' : 'Show password');
        toggleBtn.title = show ? 'Hide password' : 'Show password';
    });
})();