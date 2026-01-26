// web/static/js/auth.js
import { API_BASE } from "./utils.js";
import { openAuthModal } from "./auth-modal.js";

export const Auth = {
  checked: false,
  isAuthenticated: false,
  user: null,
  _poller: null,

  /* ==================================================
     INITIAL AUTH CHECK
  ================================================== */
  async init() {
    if (this.checked) return this;

    this.checked = true;

    try {
      const res = await fetch(`${API_BASE}/users/me`, {
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      if (res.status === 401) {
        // Guest OR session invalidated
        this.isAuthenticated = false;
        this.user = null;
        return this;
      }

      if (!res.ok) {
        console.error("Auth: unexpected error", res.status);
        this.isAuthenticated = false;
        this.user = null;
        return this;
      }

      const payload = await res.json();
      this.user = payload?.data ?? null;
      this.isAuthenticated = !!this.user;

      return this;

    } catch (err) {
      console.error("Auth: network failure", err);
      this.isAuthenticated = false;
      this.user = null;
      return this;
    }
  },

  /* ==================================================
     SINGLE-SESSION WATCHER (UX ENFORCEMENT)
  ================================================== */
  startSessionWatcher(intervalMs = 30_000) {
    if (this._poller || !this.isAuthenticated) return;

    this._poller = setInterval(async () => {
      const wasAuthenticated = this.isAuthenticated;

      // force re-check
      this.checked = false;
      await this.init();

      // session invalidated elsewhere
      if (wasAuthenticated && !this.isAuthenticated) {
        console.info("Auth: session invalidated remotely");
        this.stopSessionWatcher();
        openAuthModal();
      }
    }, intervalMs);
  },

  stopSessionWatcher() {
    if (this._poller) {
      clearInterval(this._poller);
      this._poller = null;
    }
  },
};

/* ==================================================
   AUTH GUARD (USED BY UI ACTIONS)
================================================== */
Auth.requireOrPrompt = async function () {
  await this.init();

  if (this.isAuthenticated) return true;

  openAuthModal();
  return false;
};
