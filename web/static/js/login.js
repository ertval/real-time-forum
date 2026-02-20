// web/static/js/login.js
import { initPasswordToggles } from "./password-toggle.js";
import { uiNotify } from "./ui-messages.js";
import { API_BASE } from "./utils.js";

(() => {
  const form = document.querySelector("form");
  const identifierEl = document.getElementById("email");
  const passwordEl = document.getElementById("password");
  const guestBtn = document.getElementById("guest-login-btn");

  if (!form || !identifierEl || !passwordEl) return;

  initPasswordToggles(document);

  const isInIframe = window.parent && window.parent !== window;

  /* =========================
     HIDE GUEST IN IFRAME
  ========================= */

  if (isInIframe && guestBtn) {
    guestBtn.closest(".guest-login")?.remove();
  }

  /* =========================
     OAUTH FIX (CRITICAL)
  ========================= */

  function redirectToOAuth(provider) {
    const url = `${API_BASE}/auth/${provider}`;

    if (isInIframe) {
      // Escape iframe context
      window.top.location.href = url;
    } else {
      window.location.href = url;
    }
  }

  document.addEventListener("click", (e) => {
    const googleBtn = e.target.closest('[href="/api/v1/auth/google"]');
    const githubBtn = e.target.closest('[href="/api/v1/auth/github"]');

    if (googleBtn) {
      e.preventDefault();
      redirectToOAuth("google");
    }

    if (githubBtn) {
      e.preventDefault();
      redirectToOAuth("github");
    }
  });

  /* =========================
     NORMAL LOGIN
  ========================= */

  const submitBtn = form.querySelector('button[type="submit"]');
  const LOGIN_API = form.getAttribute("action");

  function notify(message, type = "danger") {
    if (isInIframe) {
      window.parent.postMessage(
        {
          type: "auth:notify",
          payload: { message, level: type },
        },
        "*"
      );
    } else {
      uiNotify(message, { type });
    }
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const identifier = identifierEl.value.trim();
    const password = passwordEl.value;

    if (!identifier || !password) {
      notify("Email/username and password are required.", "warn");
      return;
    }

    const payload = { password };
    identifier.includes("@")
      ? (payload.email = identifier)
      : (payload.username = identifier);

    try {
      submitBtn.disabled = true;

      const res = await fetch(LOGIN_API, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        credentials: "include",
        body: JSON.stringify(payload),
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        notify(
          data?.error?.message ||
            (res.status === 401
              ? "Invalid credentials."
              : "Login failed. Please try again."),
          "danger"
        );
        return;
      }

      sessionStorage.setItem("auth:login-success", "1");

      if (isInIframe) {
        window.parent.postMessage("auth:success", "*");
        return;
      }

      window.location.href = "/";
    } catch {
      notify("Network error. Please try again.", "danger");
    } finally {
      submitBtn.disabled = false;
    }
  });

  guestBtn?.addEventListener("click", () => {
    window.location.assign("/");
  });
})();