// web/static/js/auth.js
import { API_BASE } from "./utils.js";

export const Auth = {
  checked: false,
  isAuthenticated: false,
  user: null,

  async init() {
    if (this.checked) return this;

    this.checked = true;

    try {
      const res = await fetch(`${API_BASE}/users/me`, {
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      if (res.status === 401) {
        // Guest session – EXPECTED
        console.info("Auth: guest session");
        this.isAuthenticated = false;
        this.user = null;
        return this;
      }

      if (!res.ok) {
        console.error("Auth: unexpected error", res.status);
        this.isAuthenticated = false;
        return this;
      }

      const payload = await res.json();
      this.user = payload?.data ?? null;
      this.isAuthenticated = !!this.user;

      return this;

    } catch (err) {
      console.error("Auth: network failure", err);
      this.isAuthenticated = false;
      return this;
    }
  }
};
