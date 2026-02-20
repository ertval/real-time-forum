// web/static/js/sound-effects.js

const SFX_KEY = "app:sfx-enabled";

function isSFXEnabled() {
  return localStorage.getItem(SFX_KEY) !== "false";
}

// --- Preload sounds ---
const reactionSound = new Audio("/static/sounds/reaction.mp3");
const uploadSound = new Audio("/static/sounds/upload.mp3");
const deleteSound = new Audio("/static/sounds/delete.mp3");
const notificationSound = new Audio("/static/sounds/notification.mp3");

// Default volumes
reactionSound.volume = 0.35;
uploadSound.volume = 0.35;
deleteSound.volume = 0.40;
notificationSound.volume = 0.30;

// --- Playback helpers ---
export function playReaction() {
  if (!isSFXEnabled()) return;
  try {
    reactionSound.currentTime = 0;
    reactionSound.play().catch(() => {});
  } catch (_) {}
}

export function playUpload() {
  if (!isSFXEnabled()) return;
  try {
    uploadSound.currentTime = 0;
    uploadSound.play().catch(() => {});
  } catch (_) {}
}

export function playDelete() {
  if (!isSFXEnabled()) return;
  try {
    deleteSound.currentTime = 0;
    deleteSound.play().catch(() => {});
  } catch (_) {}
}

export function playNotification() {
  if (!isSFXEnabled()) return;
  try {
    notificationSound.currentTime = 0;
    notificationSound.play().catch(() => {});
  } catch (_) {}
}

// --- Optional preload ---
export function initSounds() {
  reactionSound.load();
  uploadSound.load();
  deleteSound.load();
  notificationSound.load();
}