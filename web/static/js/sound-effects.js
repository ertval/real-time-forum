// web/static/js/sound-effects.js

const SFX_KEY = 'app:sfx-enabled';

function isSFXEnabled() {
	if (typeof localStorage === 'undefined') {
		return false;
	}

	return localStorage.getItem(SFX_KEY) !== 'false';
}

function createSound(src) {
	if (typeof Audio === 'undefined') {
		return {
			currentTime: 0,
			volume: 1,
			load() {},
			play() {
				return Promise.resolve();
			},
		};
	}

	return new Audio(src);
}

// --- Preload sounds ---
const reactionSound = createSound('/static/sounds/reaction.mp3');
const uploadSound = createSound('/static/sounds/upload.mp3');
const deleteSound = createSound('/static/sounds/delete.mp3');
const notificationSound = createSound('/static/sounds/notification.mp3');

// Default volumes
reactionSound.volume = 0.35;
uploadSound.volume = 0.35;
deleteSound.volume = 0.4;
notificationSound.volume = 0.3;

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
