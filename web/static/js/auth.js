// web/static/js/auth.js
import { API_BASE } from "./utils.js";

export const Auth = {
  checked: false,
  isAuthenticated: false,
  user: null,
  _poller: null,

  async init() {
    if (this.checked) return this;

    this.checked = true;

    try {
      const res = await fetch(`${API_BASE}/users/me`, {
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      if (res.status === 401) {
        // Guest or session invalidated
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

  // --------------------------------------------------
  // Periodic session check (single-session enforcement UX)
  // --------------------------------------------------
  startSessionWatcher(intervalMs = 30_000) {
    if (this._poller) return; // avoid duplicates

    this._poller = setInterval(async () => {
      const wasAuthenticated = this.isAuthenticated;

      // force re-check
      this.checked = false;
      await this.init();

      // session was invalidated elsewhere
      if (wasAuthenticated && !this.isAuthenticated) {
        console.info("Auth: session invalidated remotely");
        window.location.href = "/login";
      }
    }, intervalMs);
  },

  stopSessionWatcher() {
    if (this._poller) {
      clearInterval(this._poller);
      this._poller = null;
    }
  }
};
