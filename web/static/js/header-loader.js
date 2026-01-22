// web/static/js/header-loader.js

import { initHeader } from "./header.js";
import { Auth } from "./auth.js";

// --------------------------------------------------
// Initialize auth state (who am I?)
// --------------------------------------------------
await Auth.init();

// --------------------------------------------------
// Start session watcher ONLY for authenticated users
// (detect remote logout / superseded session)
// --------------------------------------------------
if (Auth.isAuthenticated) {
  Auth.startSessionWatcher();
}

// --------------------------------------------------
// Initialize header UI (login/logout buttons, user info)
// --------------------------------------------------
initHeader();
