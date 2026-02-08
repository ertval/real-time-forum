// web/static/js/login.js
import { initPasswordToggles } from "./password-toggle.js";
import { uiNotify } from "./ui-messages.js";

(() => {
  const form = document.querySelector("form");
  const identifierEl = document.getElementById("email");
  const passwordEl = document.getElementById("password");

  if (!form || !identifierEl || !passwordEl) return;

  initPasswordToggles(document);

  const submitBtn = form.querySelector('button[type="submit"]');
  const LOGIN_API = form.getAttribute("action");

  function notify(message, type = "danger") {
    // If inside iframe → delegate to parent
    if (window.parent && window.parent !== window) {
      window.parent.postMessage(
        {
          type: "auth:notify",
          payload: { message, level: type },
        },
        "*"
      );
    } else {
      // Normal page
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

      notify("Signed in successfully.", "success");

      // iframe → parent handles success
      if (window.parent && window.parent !== window) {
        window.parent.postMessage("auth:success", "*");
        return;
      }

      window.location.assign("/");
    } catch {
      notify("Network error. Please try again.", "danger");
    } finally {
      submitBtn.disabled = false;
    }
  });

  document
    .getElementById("guest-login-btn")
    ?.addEventListener("click", () => {
      window.location.assign("/");
    });
})();
