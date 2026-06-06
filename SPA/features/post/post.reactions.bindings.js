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

// Resolves the reaction target from the container element rather than the
// <input>. The reaction inputs intentionally carry no id attribute (so they
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
	const { container: scope, postId, commentId } = resolveReactionTarget(button);
	if (!scope) {
		return;
	}

	const reaction = getReactionRequest({
		type: button.dataset.reaction,
		postId,
		commentId,
	});

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
