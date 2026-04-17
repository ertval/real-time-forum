// web/static/js/reactions.js

import { Auth } from './auth.js';
import { playReaction } from './sound-effects.js';
import { API_BASE } from './utils.js';

export function initReactions() {
	document.addEventListener('change', (e) => {
		void handleReactionChange(e);
	});
}

async function handleReactionChange(e) {
	const btn = getReactionButton(e.target);
	if (!btn) return;

	e.stopPropagation();

	const previousState = !btn.checked;
	const allowed = await Auth.requireOrPrompt();
	if (!allowed) {
		btn.checked = previousState;
		return;
	}

	const reaction = readReactionTarget(btn);
	const scope = getReactionScope(btn, reaction.postId);
	if (!scope) return;

	const opposite = getOppositeReaction(scope, reaction.type);
	const oppositePreviousState = opposite ? opposite.checked : null;
	const res = await submitReaction(reaction);

	if (!res) {
		restoreReactionState(btn, opposite, previousState, oppositePreviousState);
		return;
	}

	if (res.status === 401) {
		restoreReactionState(btn, opposite, previousState, oppositePreviousState);
		await Auth.requireOrPrompt();
		return;
	}

	if (!res.ok) {
		restoreReactionState(btn, opposite, previousState, oppositePreviousState);
		return;
	}

	if (btn.checked && opposite) {
		opposite.checked = false;
	}

	playReaction();
	applyReactionCounts(scope, await res.json());
}

function getReactionButton(target) {
	if (!(target instanceof Element)) return null;
	return target.closest('[data-reaction]');
}

function readReactionTarget(btn) {
	return {
		type: btn.dataset.reaction,
		postId: btn.dataset.postId,
		commentId: btn.dataset.commentId,
	};
}

function getReactionScope(btn, postId) {
	return postId ? btn.closest('article[data-post-id]') : btn.closest('.comment, .activity-comment');
}

function getOppositeReaction(scope, type) {
	const oppositeType = type === 'like' ? 'dislike' : 'like';
	return scope.querySelector(`input[data-reaction="${oppositeType}"]`);
}

async function submitReaction({ type, postId, commentId }) {
	const url = postId
		? `${API_BASE}/posts/${postId}/${type}`
		: `${API_BASE}/comments/${commentId}/${type}`;

	try {
		return await fetch(url, {
			method: 'POST',
			credentials: 'include',
			headers: { Accept: 'application/json' },
		});
	} catch (err) {
		console.error('Reaction failed:', err);
		return null;
	}
}

function restoreReactionState(btn, opposite, previousState, oppositePreviousState) {
	btn.checked = previousState;
	if (opposite && oppositePreviousState !== null) {
		opposite.checked = oppositePreviousState;
	}
}

function applyReactionCounts(scope, payload) {
	const { data } = payload;
	scope.querySelector('[data-like-count]').textContent = data.likes_count;
	scope.querySelector('[data-dislike-count]').textContent = data.dislikes_count;
}
