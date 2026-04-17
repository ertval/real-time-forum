// /static/js/activity/posts-activity.js

import { playDelete, playUpload } from '../sound-effects.js';
import { uiConfirm, uiNotify } from '../ui-messages.js';
import { API_BASE } from '../utils.js';

let editBound = false;
let deleteBound = false;
let statusToggleBound = false;

export function initEditPostNavigation() {
	if (editBound) return;
	editBound = true;

	document.addEventListener('click', handleEditPostNavigation, true);
}

export function initDeletePost(refresh) {
	if (deleteBound) return;
	deleteBound = true;

	document.addEventListener(
		'click',
		(e) => {
			void handleDeletePostClick(e, refresh);
		},
		true,
	);
}

export function initStatusToggle(refresh) {
	if (statusToggleBound) return;
	statusToggleBound = true;

	document.addEventListener(
		'click',
		(e) => {
			void handleStatusToggleClick(e, refresh);
		},
		true,
	);
}

function getActionButton(target, selector) {
	if (!(target instanceof Element)) return null;
	return target.closest(selector);
}

function handleEditPostNavigation(e) {
	const btn = getActionButton(e.target, '.post-edit');
	if (!btn) return;

	e.preventDefault();
	e.stopPropagation();

	const postId = btn.dataset.postId;
	if (!postId) return;

	const next = encodeURIComponent(`${window.location.pathname}${window.location.search}`);
	window.location.href = `/edit-post/${postId}?next=${next}`;
}

async function handleDeletePostClick(e, refresh) {
	const btn = getActionButton(e.target, '.post-delete');
	if (!btn) return;

	e.preventDefault();
	e.stopPropagation();

	const postId = btn.dataset.postId;
	if (!postId) return;

	const ok = await uiConfirm('Delete this post? This action cannot be undone.', {
		type: 'danger',
		title: 'Delete post',
		okText: 'Delete',
		cancelText: 'Cancel',
	});

	if (!ok) return;

	btn.disabled = true;

	try {
		const res = await deletePost(postId);
		if (!(await handleDeletePostResponse(res, refresh))) return;

		playDelete();
		uiNotify('Post deleted.', { type: 'success' });
		await refresh();
	} catch (err) {
		console.error('Activity delete failed:', err);
		uiNotify('Failed to delete post.', { type: 'danger' });
	} finally {
		if (document.contains(btn)) btn.disabled = false;
	}
}

async function deletePost(postId) {
	return fetch(`${API_BASE}/posts/${postId}`, {
		method: 'DELETE',
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});
}

async function handleDeletePostResponse(res, refresh) {
	if (res.status === 401) {
		uiNotify('You must be logged in.', { type: 'warn' });
		return false;
	}

	if (res.status === 403) {
		uiNotify('You can only delete your own posts.', { type: 'warn' });
		return false;
	}

	if (res.status === 404) {
		uiNotify('Post not found.', { type: 'warn' });
		await refresh();
		return false;
	}

	if (!res.ok) {
		uiNotify('Failed to delete post.', { type: 'danger' });
		return false;
	}

	return true;
}

async function handleStatusToggleClick(e, refresh) {
	const btn = getActionButton(e.target, '.post-status-toggle');
	if (!btn) return;

	e.preventDefault();
	e.stopPropagation();

	const postId = btn.dataset.postId;
	const currentStatus = btn.dataset.currentStatus;
	if (!postId || !currentStatus) return;

	const nextStatus = currentStatus === 'draft' ? 'published' : 'draft';

	btn.disabled = true;

	try {
		const res = await updatePostStatus(postId, nextStatus);
		if (!res.ok) {
			uiNotify('Failed to update post status.', { type: 'danger' });
			return;
		}

		playUpload();
		updateStatusToggleButton(btn, nextStatus);
		await refresh();
	} catch (err) {
		console.error('Activity status toggle failed:', err);
		uiNotify('Failed to update post status.', { type: 'danger' });
	} finally {
		if (document.contains(btn)) btn.disabled = false;
	}
}

async function updatePostStatus(postId, status) {
	return fetch(`${API_BASE}/posts/${postId}`, {
		method: 'PATCH',
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			Accept: 'application/json',
		},
		body: JSON.stringify({ status }),
	});
}

function updateStatusToggleButton(btn, nextStatus) {
	btn.dataset.currentStatus = nextStatus;

	const img = btn.querySelector('img');
	if (img) {
		img.src = nextStatus === 'draft' ? '/static/img/publish.png' : '/static/img/draft.png';
	}

	btn.title = nextStatus === 'draft' ? 'Publish post' : 'Move to draft';
}
