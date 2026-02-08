// web/static/js/register.js
import { initPasswordToggles } from "./password-toggle.js";
import { uiNotify } from "./ui-messages.js";

(() => {
  const form = document.querySelector("form");
  if (!form) return;

  const usernameEl = document.getElementById("username");
  const emailEl = document.getElementById("email");
  const passwordEl = document.getElementById("password");
  const confirmEl = document.getElementById("confirm_password");

  if (!usernameEl || !emailEl || !passwordEl || !confirmEl) return;

  // init password eye toggles
  initPasswordToggles(document);

  const submitBtn = form.querySelector('button[type="submit"]');

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

    const username = usernameEl.value.trim();
    const email = emailEl.value.trim();
    const password = passwordEl.value;
    const confirm = confirmEl.value;

    if (!username || !email || !password || !confirm) {
      notify("All fields are required.", "warn");
      return;
    }

    if (password.length < 6) {
      notify("Password must be at least 6 characters.", "warn");
      return;
    }

    if (password !== confirm) {
      notify("Passwords do not match.", "warn");
      return;
    }

    try {
      submitBtn.disabled = true;

      const res = await fetch("/api/v1/users/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          username,
          email,
          password,
        }),
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        notify(
          data?.error?.message ||
            "Registration failed. Please try again.",
          "danger"
        );
        return;
      }

      notify(
        "Account created successfully. You can now sign in.",
        "success"
      );

      // If inside iframe, let parent decide what to do next
      if (window.parent && window.parent !== window) {
        // optional: parent can switch iframe to /login
        window.parent.postMessage(
          { type: "auth:registered" },
          "*"
        );
        return;
      }

      window.location.assign("/login");
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
