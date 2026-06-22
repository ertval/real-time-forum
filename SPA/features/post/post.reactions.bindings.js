import { getCurrentUser, postReaction } from './post.reactions.api.js';
import {
	applyReactionSelection,
	getOppositeReactionType,
	getReactionCounts,
	getReactionRequest,
} from './post.reactions.logic.js';

let reactionsInitialized = false;

function getReactionButton(target) {
	if (!(target instanceof Element)) {
		return null;
	}

	return target.closest('[data-reaction]');
}

function isPressed(button) {
	return button?.getAttribute('aria-pressed') === 'true';
}

function setPressed(button, pressed) {
	if (button) {
		button.setAttribute('aria-pressed', pressed ? 'true' : 'false');
	}
}

// Resolves the reaction target from the container element rather than the
// <button>. The reaction buttons intentionally carry no id attribute (so they
// never collide with the unique [data-post-id] container selector), so the id
// is read from the nearest [data-post-id]/[data-comment-id] ancestor. The
// .reactions wrapper's data-reaction-scope hook picks which container to find:
// posts live under [data-post-id] (<article> in the feed/activity, <section> on
// the detail route), comments under .comment / .activity-comment.
function resolveReactionTarget(button) {
	const wrapper = button.closest('[data-reaction-scope]');
	const scope = wrapper?.dataset.reactionScope;

	if (scope === 'comment') {
		const container = button.closest('[data-comment-id]');
		return { container, postId: null, commentId: container?.dataset.commentId ?? null };
	}

	const container = button.closest('[data-post-id]');
	return { container, postId: container?.dataset.postId ?? null, commentId: null };
}

function applyReactionCountsToDom(scope, payload) {
	const { likesCount, dislikesCount } = getReactionCounts(payload);
	const likeCount = scope?.querySelector('[data-like-count]');
	const dislikeCount = scope?.querySelector('[data-dislike-count]');

	if (likeCount) {
		likeCount.textContent = String(likesCount);
	}

	if (dislikeCount) {
		dislikeCount.textContent = String(dislikesCount);
	}
}

async function handleReactionClick(fetchRef, event) {
	const button = getReactionButton(event.target);
	if (!button) {
		return;
	}

	event.stopPropagation();

	const { container: scope, postId, commentId } = resolveReactionTarget(button);
	if (!scope) {
		return;
	}

	const reaction = getReactionRequest({
		type: button.dataset.reaction,
		postId,
		commentId,
	});

	// The button's aria-pressed reflects the persisted reaction: clicking the
	// active button clears it, clicking the other switches. Nothing is mutated
	// until the server confirms, so a failed or anonymous request simply leaves
	// the existing state untouched — there is nothing to roll back.
	const wasActive = isPressed(button);

	const currentUser = await getCurrentUser(fetchRef);
	if (!currentUser) {
		return;
	}

	const nextSelection = applyReactionSelection(wasActive ? reaction.type : null, reaction.type);
	const response = await postReaction(fetchRef, reaction);
	if (!response?.ok) {
		return;
	}

	const oppositeType = getOppositeReactionType(reaction.type);
	const opposite = scope.querySelector(`[data-reaction="${oppositeType}"]`);
	setPressed(button, nextSelection === reaction.type);
	setPressed(opposite, false);

	applyReactionCountsToDom(scope, await response.json());
}

export function initReactionBindings({
	documentRef = typeof document !== 'undefined' ? document : null,
	fetchRef = typeof fetch === 'function' ? fetch.bind(globalThis) : null,
} = {}) {
	if (
		reactionsInitialized ||
		!documentRef ||
		typeof documentRef.addEventListener !== 'function' ||
		typeof fetchRef !== 'function'
	) {
		return;
	}

	documentRef.addEventListener('click', (event) => {
		void handleReactionClick(fetchRef, event);
	});

	reactionsInitialized = true;
}
