import { API_BASE } from '../../core/api/constants.js';

const DEFAULT_PREVIEW_LIMIT = 3;

export async function loadPostCommentsPreview(fetchRef, postId, limit = DEFAULT_PREVIEW_LIMIT) {
	const response = await fetchRef(`${API_BASE}/posts/${postId}/comments`, {
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});

	if (!response.ok) {
		return [];
	}

	const payload = await response.json();
	const comments = Array.isArray(payload?.data) ? payload.data : [];
	return comments.slice(0, limit);
}
