// SPA/features/post/post.api.js

import { buildImageRequestOptions } from '../../core/shared/utils.js';
import { API_BASE } from '../../core/api/constants.js';

async function getNormalizedPayload(response) {
	if (!response.ok) {
		return null;
	}

	const payload = await response.json();

	if (Array.isArray(payload)) {
		return payload;
	}

	return payload?.data ?? payload ?? null;
}

export async function getPostById(fetchRef, postId) {
	try {
		const response = await fetchRef(`${API_BASE}/posts/${postId}`, {
			credentials: 'include',
			headers: { Accept: 'application/json' },
		});

		return await getNormalizedPayload(response);
	} catch {
		return null;
	}
}

export async function getPostComments(fetchRef, postId) {
	try {
		const response = await fetchRef(`${API_BASE}/posts/${postId}/comments`, {
			credentials: 'include',
			headers: { Accept: 'application/json' },
		});

		return await getNormalizedPayload(response);
	} catch {
		return null;
	}
}

export async function createPostComment(fetchRef, postId, { body = '', imageFile = null } = {}) {
	try {
		const response = await fetchRef(
			`${API_BASE}/posts/${postId}/comments`,
			buildImageRequestOptions({
				method: 'POST',
				imageFile,
				buildMultipartBody: (file) => {
					const formData = new FormData();
					formData.append('body', body);
					formData.append('image', file);
					return formData;
				},
				jsonBody: { body },
				multipartHeaders: { Accept: 'application/json' },
				jsonHeaders: { Accept: 'application/json' },
			}),
		);

		return {
			ok: response.ok,
			status: response.status,
			data: await getNormalizedPayload(response),
		};
	} catch {
		return { ok: false, status: 0, data: null };
	}
}
