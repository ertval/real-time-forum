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

function getReactionScope(button, postId) {
	if (!postId) {
		return button.closest('.comment, .activity-comment');
	}

	// The reaction <input> itself carries data-post-id, so search from its parent
	// to skip the input and find the container element. The feed and activity
	// views use <article data-post-id> while the post-detail route uses
	// <section data-post-id>, so match on the attribute rather than the tag.
	return button.parentElement?.closest('[data-post-id]') ?? null;
}

function restoreReactionState(button, opposite, previousState, oppositePreviousState) {
	button.checked = previousState;
	if (opposite && oppositePreviousState !== null) {
		opposite.checked = oppositePreviousState;
	}
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

async function handleReactionChange(fetchRef, event) {
	const button = getReactionButton(event.target);
	if (!button) {
		return;
	}

	event.stopPropagation();

	const previousState = !button.checked;
	const reaction = getReactionRequest({
		type: button.dataset.reaction,
		postId: button.dataset.postId,
		commentId: button.dataset.commentId,
	});
	const scope = getReactionScope(button, reaction.postId);
	if (!scope) {
		return;
	}

	const oppositeType = getOppositeReactionType(reaction.type);
	const opposite = scope.querySelector(`input[data-reaction="${oppositeType}"]`);
	const oppositePreviousState = opposite ? opposite.checked : null;
	const currentUser = await getCurrentUser(fetchRef);
	if (!currentUser) {
		restoreReactionState(button, opposite, previousState, oppositePreviousState);
		return;
	}

	const nextSelection = applyReactionSelection(previousState ? reaction.type : null, reaction.type);
	const response = await postReaction(fetchRef, reaction);
	if (!response?.ok) {
		restoreReactionState(button, opposite, previousState, oppositePreviousState);
		return;
	}

	button.checked = nextSelection === reaction.type;
	if (opposite) {
		opposite.checked = false;
	}

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

	documentRef.addEventListener('change', (event) => {
		void handleReactionChange(fetchRef, event);
	});

	reactionsInitialized = true;
}
