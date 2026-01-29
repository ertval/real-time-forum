// web/static/js/register.js
import { initPasswordToggles } from "./password-toggle.js";

(() => {
  const form = document.querySelector("form");
  if (!form) return;

  const usernameEl = document.getElementById("username");
  const emailEl = document.getElementById("email");
  const passwordEl = document.getElementById("password");
  const confirmEl = document.getElementById("confirm_password");

  // init password eye
  initPasswordToggles(document);

  const submitBtn = form.querySelector('button[type="submit"]');

  let errorEl = document.querySelector(".auth-error");
  if (!errorEl) {
    errorEl = document.createElement("div");
    errorEl.className = "auth-error";
    errorEl.setAttribute("role", "alert");
    form.prepend(errorEl);
  }

  function showError(msg) {
    errorEl.textContent = msg || "";
    errorEl.style.display = msg ? "block" : "none";
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    showError("");

    const username = usernameEl.value.trim();
    const email = emailEl.value.trim();
    const password = passwordEl.value;
    const confirm = confirmEl.value;

    if (!username || !email || !password || !confirm) {
      showError("All fields are required.");
      return;
    }

    if (password !== confirm) {
      showError("Passwords do not match.");
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
        showError(data?.error?.message || "Registration failed.");
        return;
      }

      window.location.assign("/login");
    } catch {
      showError("Network error. Please try again.");
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
