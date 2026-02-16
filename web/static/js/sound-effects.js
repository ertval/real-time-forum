// web/static/js/sound-effects.js

// --- Preload sounds ---
const reactionSound = new Audio("/static/sounds/reaction.mp3");
const uploadSound = new Audio("/static/sounds/upload.mp3");

reactionSound.volume = 0.35;
uploadSound.volume = 0.35;

export function playReaction() {
    try {
        reactionSound.currentTime = 0;
        reactionSound.play();
    } catch (_) {}
}

export function playUpload() {
    try {
        uploadSound.currentTime = 0;
        uploadSound.play();
    } catch (_) {}
}

// Optional: preload in background for smoother first play
export function initSounds() {
    reactionSound.load();
    uploadSound.load();
}
