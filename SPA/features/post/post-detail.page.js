// SPA/features/post/post-detail.page.js

import { API_BASE } from '../../core/api/constants.js';
import { escapeHTML } from '../../core/utils/html.js';

const POST_DETAIL_BOUND_ATTR = 'data-post-detail-bound';

async function fetchPost(fetchRef, postID) {
	const response = await fetchRef(`${API_BASE}/posts/${postID}`, {
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});
	if (!response.ok) return null;
	const payload = await response.json();
	return payload.data ?? payload;
}

async function fetchComments(fetchRef, postID) {
	const response = await fetchRef(`${API_BASE}/posts/${postID}/comments`, {
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});
	if (!response.ok) return [];
	const payload = await response.json();
	return payload.data ?? payload;
}

function renderPostContent(contentEl, post) {
	if (!contentEl || !post) return;
	contentEl.innerHTML = `
		<h1 class="post-title">${escapeHTML(post.title)}</h1>
		<p class="post-meta">Author: <a data-link class="profile-link" href="/profile/${escapeHTML(post.author_id || post.user_id || '')}">${escapeHTML(post.username || post.author || 'User')}</a></p>
		<p class="post-body">${escapeHTML(post.body || '')}</p>
	`;
}

function renderCommentsList(commentsListEl, comments) {
	if (!commentsListEl) return;
	if (!Array.isArray(comments) || comments.length === 0) {
		commentsListEl.innerHTML = '<p>No comments yet.</p>';
		return;
	}

	commentsListEl.innerHTML = comments
		.map(
			(comment) => `
		<div class="comment" data-comment-id="${escapeHTML(comment.id)}">
			<p class="comment-meta">
				<strong>
					<a data-link class="profile-link" href="/profile/${escapeHTML(comment.author_id || comment.user_id || '')}">
						${escapeHTML(comment.username || comment.author || (comment.user_id ? `User ${comment.user_id}` : 'User'))}
					</a>
				</strong>
			</p>
			<p class="comment-body">${escapeHTML(comment.body || '')}</p>
		</div>
	`,
		)
		.join('');
}

export function initPostDetailPage(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const fetchRef = options.fetchRef ?? (typeof fetch === 'function' ? fetch.bind(windowRef) : null);

	if (
		!windowRef ||
		!documentRef ||
		typeof fetchRef !== 'function' ||
		typeof documentRef.querySelector !== 'function'
	) {
		return null;
	}

	const root = documentRef.querySelector('[data-screen="post-detail"]');
	if (!root || root.getAttribute(POST_DETAIL_BOUND_ATTR) === 'true') return null;

	root.setAttribute(POST_DETAIL_BOUND_ATTR, 'true');

	const postID = root.getAttribute('data-post-id');
	const contentEl = root.querySelector('#post-content');
	const commentsListEl = root.querySelector('#comments-list');

	if (!postID) return null;

	void (async () => {
		const post = await fetchPost(fetchRef, postID);
		renderPostContent(contentEl, post);

		const comments = await fetchComments(fetchRef, postID);
		renderCommentsList(commentsListEl, comments);
	})();

	return { root };
}
