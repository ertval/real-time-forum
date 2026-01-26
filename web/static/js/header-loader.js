// web/static/js/header-loader.js

import { initHeader } from "./header.js";
import { Auth } from "./auth.js";
import { loadAuthModal } from "./auth-modal.js";

(async function bootstrap() {
  // preload auth modal ONCE
  await loadAuthModal();

  // auth state
  await Auth.init();

  if (Auth.isAuthenticated) {
    Auth.startSessionWatcher();
  }

  initHeader();
})();
