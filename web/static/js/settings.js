// web/static/js/settings.js
import { initSounds } from "./sound-effects.js";

const SFX_KEY = "app:sfx-enabled";

function getSFX() {
  return localStorage.getItem(SFX_KEY) !== "false";
}

function setSFX(val) {
  localStorage.setItem(SFX_KEY, val);
}

let settingsInitialized = false;

function initSettings() {
  if (settingsInitialized) return;
  settingsInitialized = true;

  const modal = document.getElementById("settings-modal");
  if (!modal) return;

  if (getSFX()) {
    initSounds();
  }

  document.addEventListener("click", handleSettingsClick);
  document.addEventListener("change", handleSettingsChange);
}

function handleSettingsClick(e) {
  const modal = document.getElementById("settings-modal");
  const openBtn = e.target.closest("#settings-btn");
  const closeBtn = e.target.closest("#settings-close");
  const backdrop = e.target.closest(".settings-backdrop");
  const logoutBtn = e.target.closest("#settings-logout-btn");

  if (openBtn) {
    modal.classList.remove("hidden");
    modal.setAttribute("aria-hidden", "false");
    document.getElementById("sfxToggle").checked = getSFX();
    document.body.style.overflow = "hidden";
    return;
  }

  if (closeBtn || backdrop) {
    modal.classList.add("hidden");
    modal.setAttribute("aria-hidden", "true");
    document.body.style.overflow = "";
    return;
  }

  if (logoutBtn) {
    handleLogout();
  }
}

function handleSettingsChange(e) {
  if (e.target.id !== "sfxToggle") return;

  const enabled = e.target.checked;
  setSFX(enabled);

  if (enabled) initSounds();
}

async function handleLogout() {
  try {
    await fetch("/api/v1/users/logout", {
      method: "POST",
      credentials: "include",
    });
  } catch (err) {
    console.error("Logout failed:", err);
  }

  window.location.href = "/login";
}

export { initSettings };