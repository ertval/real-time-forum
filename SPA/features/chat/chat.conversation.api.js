import { API_BASE } from '../../core/api/constants.js';

const EMPTY_CONVERSATION = { messages: [], hasMore: false };

// C03 history endpoint: GET /api/v1/chats/:id/messages -> chronological batch of 10.
// Pass { beforeId } to page backwards: returns the 10 messages older than that id.
export async function fetchConversation(fetchRef, userId, { beforeId } = {}) {
	if (typeof fetchRef !== 'function' || !Number.isFinite(userId) || userId <= 0) {
		return { ...EMPTY_CONVERSATION };
	}

	let path = `${API_BASE}/chats/${userId}/messages`;
	if (Number.isFinite(beforeId) && beforeId > 0) {
		path += `?before_id=${beforeId}`;
	}

	try {
		const response = await fetchRef(path, {
			credentials: 'include',
			headers: { Accept: 'application/json' },
		});

		if (!response?.ok) {
			return { ...EMPTY_CONVERSATION };
		}

		const payload = await response.json();
		const data = payload?.data ?? {};
		return {
			messages: Array.isArray(data.messages) ? data.messages : [],
			hasMore: Boolean(data.has_more),
		};
	} catch {
		return { ...EMPTY_CONVERSATION };
	}
}

// Needed to tell own messages apart from the other party's in the rendered history.
export async function fetchCurrentUserId(fetchRef) {
	if (typeof fetchRef !== 'function') {
		return null;
	}

	try {
		const response = await fetchRef(`${API_BASE}/users/me`, {
			credentials: 'include',
			headers: { Accept: 'application/json' },
		});

		if (!response?.ok) {
			return null;
		}

		const payload = await response.json();
		const id = payload?.data?.id;
		return Number.isFinite(id) ? Number(id) : null;
	} catch {
		return null;
	}
}
