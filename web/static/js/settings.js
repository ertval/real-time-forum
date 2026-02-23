// web/static/js/settings.js
import { initSounds } from "./sound-effects.js";

const SFX_KEY = "app:sfx-enabled";

function getSFX() {
  return localStorage.getItem(SFX_KEY) !== "false";
}

function setSFX(val) {
  localStorage.setItem(SFX_KEY, val);
}

function initSettings() {
  const modal = document.getElementById("settings-modal");
  if (!modal) return;

  if (getSFX()) {
    initSounds();
  }

  document.addEventListener("click", (e) => {
    const openBtn = e.target.closest("#settings-btn");
    const closeBtn = e.target.closest("#settings-close");
    const backdrop = e.target.closest(".settings-backdrop");
    const checkbox = document.getElementById("sfxToggle");

    if (openBtn) {
      modal.classList.remove("hidden");

      if (checkbox) {
        checkbox.checked = getSFX();
      }

      document.body.style.overflow = "hidden";
      return;
    }

    if (closeBtn || backdrop) {
      modal.classList.add("hidden");
      document.body.style.overflow = "";
    }
  });

  document.addEventListener("change", (e) => {
    if (e.target.id === "sfxToggle") {
      const enabled = e.target.checked;
      setSFX(enabled);

      if (enabled) {
        initSounds();
      }
    }
  });
}

export { initSettings };