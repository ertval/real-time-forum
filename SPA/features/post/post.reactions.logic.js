export function getReactionRequest({ type, postId, commentId }) {
	return {
		type,
		postId: postId || null,
		commentId: commentId || null,
	};
}

export function getOppositeReactionType(type) {
	return type === 'like' ? 'dislike' : 'like';
}

export function applyReactionSelection(currentReaction, nextReaction) {
	return currentReaction === nextReaction ? null : nextReaction;
}

export function getReactionCounts(payload) {
	return {
		likesCount: payload?.data?.likes_count ?? 0,
		dislikesCount: payload?.data?.dislikes_count ?? 0,
	};
}
