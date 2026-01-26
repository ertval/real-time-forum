// web/static/js/auth-modal.js
import { API_BASE } from "./utils.js";
import { Auth } from "./auth.js";

let modalLoaded = false;
let modal = null;

/* ==================================================
   LOAD MODAL (SAFE, ONCE)
================================================== */
export async function loadAuthModal() {
  if (modalLoaded) return;

  const res = await fetch("/static/partials/auth-modal.html");
  const html = await res.text();

  document.body.insertAdjacentHTML("beforeend", html);
  modal = document.getElementById("auth-modal");

  bindUI();
  modalLoaded = true;
}

/* ==================================================
   OPEN MODAL (ALWAYS SAFE)
================================================== */
export async function openAuthModal() {
  // ensure modal is loaded
  if (!modalLoaded) {
    await loadAuthModal();
  }

  modal.classList.remove("hidden");
}

/* ==================================================
   CLOSE
================================================== */
function closeAuthModal() {
  modal.classList.add("hidden");
}

/* ==================================================
   UI BINDINGS
================================================== */
function bindUI() {
  const closeBtn = modal.querySelector(".auth-close");
  const backdrop = modal.querySelector(".auth-backdrop");

  const tabs = modal.querySelectorAll(".auth-tab");
  const loginForm = modal.querySelector("#auth-login");
  const registerForm = modal.querySelector("#auth-register");
  const errorEl = modal.querySelector(".auth-error");

  closeBtn.onclick = backdrop.onclick = closeAuthModal;

  // Tabs
  tabs.forEach(tab => {
    tab.onclick = () => {
      tabs.forEach(t => t.classList.remove("active"));
      tab.classList.add("active");

      errorEl.classList.add("hidden");

      if (tab.dataset.tab === "login") {
        loginForm.classList.remove("hidden");
        registerForm.classList.add("hidden");
      } else {
        registerForm.classList.remove("hidden");
        loginForm.classList.add("hidden");
      }
    };
  });

  loginForm.onsubmit = e => submitAuth(e, "/users/login");
  registerForm.onsubmit = e => submitAuth(e, "/users/register");
}

/* ==================================================
   SUBMIT
================================================== */
async function submitAuth(e, endpoint) {
  e.preventDefault();

  const form = e.target;
  const errorEl = modal.querySelector(".auth-error");
  const payload = Object.fromEntries(new FormData(form));

  if (endpoint.includes("register")) {
    if (payload.password !== payload.confirm_password) {
      errorEl.textContent = "Passwords do not match";
      errorEl.classList.remove("hidden");
      return;
    }
    delete payload.confirm_password;
  }

  const res = await fetch(`${API_BASE}${endpoint}`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    errorEl.textContent =
      endpoint.includes("login")
        ? "Invalid credentials"
        : "Registration failed";
    errorEl.classList.remove("hidden");
    return;
  }

  // refresh auth state
  Auth.checked = false;
  await Auth.init();

  closeAuthModal();

  // notify app (header, buttons, etc)
  window.dispatchEvent(new Event("auth:changed"));
}
