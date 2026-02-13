// web/static/js/sound-effects.js

const likeSound = new Audio("/static/sounds/like.mp3");
const dislikeSound = new Audio("/static/sounds/dislike.mp3");

likeSound.volume = 0.3;
dislikeSound.volume = 0.3;

export function playLike() {
    likeSound.currentTime = 0;
    likeSound.play().catch(() => {});
}

export function playDislike() {
    dislikeSound.currentTime = 0;
    dislikeSound.play().catch(() => {});
}
