// static/js/settings-loader.js

import { initSettings } from "./settings.js";

export async function loadSettingsModal() {
  const res = await fetch("/static/partials/settings-modal.html");
  const html = await res.text();

  document.body.insertAdjacentHTML("beforeend", html);

  // IMPORTANT: initialize AFTER injection
  initSettings();
}