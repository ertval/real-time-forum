// web/static/js/header-loader.js
import { initHeader } from "./header.js";
import { Auth } from "./auth.js";
import { loadAuthModal } from "./auth-modal.js";
import { closeAuthModal } from "./auth-modal.js";
import { initSettings } from "./settings.js";

(async function bootstrap() {
  await loadAuthModal();
  await Auth.init();

  if (Auth.isAuthenticated) {
    Auth.startSessionWatcher();
  }

  initSettings();
  initHeader();
})();

window.addEventListener("message", async (e) => {
  if (e.data !== "auth:success") return;

  closeAuthModal();

  Auth.checked = false;
  await Auth.init();

  window.dispatchEvent(new Event("auth:changed"));
});