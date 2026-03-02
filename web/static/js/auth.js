// web/static/js/auth.js
import { API_BASE } from "./utils.js";
import { openAuthModal } from "./auth-modal.js";

export const Auth = {
  checked: false,
  isAuthenticated: false,
  user: null,
  _poller: null,
  _initPromise: null,

  async init() {
    if (this._initPromise) {
      await this._initPromise;
      return this;
    }
    if (this.checked) return this;

    this._initPromise = (async () => {
      this.checked = true;

      try {
        const res = await fetch(`${API_BASE}/users/me`, {
          credentials: "include",
          headers: { Accept: "application/json" },
        });

        if (res.status === 401) {
          this.isAuthenticated = false;
          this.user = null;
          return;
        }

        const payload = await res.json();
        this.user = payload?.data ?? null;
        this.isAuthenticated = !!this.user;
      } catch {
        this.isAuthenticated = false;
        this.user = null;
      }
    })();

    try {
      await this._initPromise;
    } finally {
      this._initPromise = null;
    }

    return this;
  },

  startSessionWatcher(intervalMs = 30000) {
    if (this._poller || !this.isAuthenticated) return;

    this._poller = setInterval(async () => {
      const wasAuthed = this.isAuthenticated;
      this.checked = false;
      await this.init();

      if (wasAuthed && !this.isAuthenticated) {
        openAuthModal();
        this.stopSessionWatcher();
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

Auth.requireOrPrompt = async function () {
  await this.init();

  if (this.isAuthenticated) return true;

  openAuthModal("/login");
  return false;
};
