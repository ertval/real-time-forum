// web/static/js/forgot-password.js
(() => {
    const form = document.querySelector('form');
    if (!form) return;

    const emailEl = document.getElementById('email');

    let msgEl = document.querySelector('.forgot-password-message');
    if (!msgEl) {
        msgEl = document.createElement('p');
        msgEl.className = 'forgot-password-message muted';
        form.after(msgEl);
    }

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        const email = emailEl.value.trim();
        if (!email) return;

        // Placeholder UX until backend exists
        msgEl.textContent =
            'If an account with this email exists, a reset link will be sent.';
    });
})();
