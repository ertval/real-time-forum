// SPA/features/post/post.page.js

import { matchRoute } from '../../core/router/routes.js';
import { setupImagePicker } from '../../core/shared/image-picker.js';
import { escapeHTML } from '../../core/utils/html.js';
import { createPostComment, getPostById, getPostComments } from './post.api.js';
import { formatCreatedAt, resolveUsername } from './post-card.views.js';
import { renderPostDetailView } from './post-detail.views.js';

const POST_DETAIL_BOUND_ATTR = 'data-post-detail-bound';

function renderCommentsMarkup(comments) {
	if (!Array.isArray(comments) || comments.length === 0) {
		return '<p class="muted">No comments yet.</p>';
	}

	return comments
		.map((comment) => {
			const body = comment.body ? escapeHTML(comment.body) : '';
			const author = escapeHTML(resolveUsername(comment));
			const createdAt = formatCreatedAt(comment.created_at);
			const createdAtMarkup = createdAt
				? `<time class="comment-date" datetime="${escapeHTML(comment.created_at)}">${escapeHTML(createdAt)}</time>`
				: '';
			const imageMarkup =
				typeof comment.image_url === 'string' && comment.image_url.trim()
					? `
						<div class="comment-image">
							<img src="${escapeHTML(comment.image_url)}" alt="Comment attachment" loading="lazy" />
						</div>
					`
					: '';

			return `
				<article class="comment card card-pad" data-comment-id="${escapeHTML(String(comment.id ?? ''))}">
					<header class="comment-header">
						<p class="muted">Author: ${author}</p>
						${createdAtMarkup}
					</header>
					<div class="comment-body">
						<p>${body}</p>
						${imageMarkup}
					</div>
				</article>
			`;
		})
		.join('');
}

function renderPostError(root, message) {
	root.innerHTML = `
		<section class="post-detail-error" role="alert">
			<h1 id="screen-post-title">Post unavailable</h1>
			<p>${escapeHTML(message)}</p>
		</section>
	`;
}

async function reloadComments(fetchRef, postId, commentsContainer) {
	commentsContainer.innerHTML = `
		<div class="post-comments-loading" aria-busy="true">
			<p>Loading comments...</p>
		</div>
	`;

	const comments = await getPostComments(fetchRef, postId);
	if (!comments) {
		commentsContainer.innerHTML = `
			<p class="muted" role="alert">Unable to load comments right now.</p>
		`;
		return null;
	}

	commentsContainer.innerHTML = renderCommentsMarkup(comments);
	return comments;
}

function setCommentFormFeedback(form, message, isError = false) {
	const feedback = form?.querySelector?.('[data-comment-feedback]');
	if (!feedback) {
		return;
	}

	if (!message) {
		feedback.hidden = true;
		feedback.textContent = '';
		feedback.removeAttribute('role');
		return;
	}

	feedback.hidden = false;
	feedback.textContent = message;
	feedback.setAttribute('role', isError ? 'alert' : 'status');
}

function setCommentFormBusy(form, isBusy) {
	const submitButton = form?.querySelector?.('button[type="submit"]');
	if (!submitButton) {
		return;
	}

	if (isBusy) {
		if (!submitButton.dataset.originalLabel) {
			submitButton.dataset.originalLabel = submitButton.textContent ?? '';
		}
		submitButton.disabled = true;
		submitButton.textContent = 'Posting...';
		return;
	}

	submitButton.disabled = false;
	if (submitButton.dataset.originalLabel) {
		submitButton.textContent = submitButton.dataset.originalLabel;
		delete submitButton.dataset.originalLabel;
	}
}

function bindCommentForm({ form, fetchRef, postId, commentsContainer }) {
	if (!form || form.dataset.bound === 'true') {
		return;
	}

	form.dataset.bound = 'true';
	const imageInput = form.querySelector('.comment-image-input');
	const imagePicker = setupImagePicker({
		input: imageInput,
		triggerButton: form.querySelector('.comment-image-btn'),
		clearButton: form.querySelector('.image-clear'),
		nameLabel: form.querySelector('.comment-image-name'),
		previewContainer: form.querySelector('.comment-image-preview'),
		previewImage: form.querySelector('.comment-image-preview img'),
		onTooLarge: () => {
			setCommentFormFeedback(form, 'Image must be 20MB or smaller.', true);
		},
	});

	form.addEventListener('submit', async (event) => {
		event.preventDefault();

		const formData = new FormData(form);
		const body = typeof formData.get('body') === 'string' ? formData.get('body').trim() : '';
		const imageFile = imagePicker?.getFile?.() ?? imageInput?.files?.[0] ?? null;

		if (!body && !imageFile) {
			setCommentFormFeedback(form, 'Cannot submit an empty comment.', true);
			return;
		}

		setCommentFormFeedback(form, '');
		setCommentFormBusy(form, true);

		try {
			const result = await createPostComment(fetchRef, postId, { body, imageFile });
			if (!result?.ok) {
				setCommentFormFeedback(form, 'Unable to post comment right now.', true);
				return;
			}

			form.reset();
			imagePicker?.clearSelectedFile?.();
			await reloadComments(fetchRef, postId, commentsContainer);
			setCommentFormFeedback(form, 'Comment posted.');
		} finally {
			setCommentFormBusy(form, false);
		}
	});
}

function resolvePostDetailContext(documentRef, match) {
	const root = documentRef.querySelector('[data-screen="post-detail"]');
	if (!root || root.getAttribute(POST_DETAIL_BOUND_ATTR) === 'true') {
		return null;
	}

	const postId = match.params.id;
	if (!postId) {
		renderPostError(root, 'The requested post is missing an identifier.');
		return null;
	}

	root.setAttribute(POST_DETAIL_BOUND_ATTR, 'true');
	root.innerHTML = `
		<div class="post-detail-loading" aria-busy="true">
			<p>Loading post...</p>
		</div>
	`;

	return { root, postId };
}

export async function initPostDetailPage(options = {}) {
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

	const match = matchRoute(windowRef.location.pathname || '/');
	if (match?.route?.id !== 'post-detail') {
		return null;
	}

	const context = resolvePostDetailContext(documentRef, match);
	if (!context) {
		return null;
	}

	const { root, postId } = context;

	const post = await getPostById(fetchRef, postId);
	if (!post) {
		renderPostError(root, 'Unable to load this post right now.');
		return null;
	}

	root.outerHTML = renderPostDetailView(post);

	const nextRoot = documentRef.querySelector('[data-screen="post-detail"]');
	if (!nextRoot) {
		return null;
	}

	nextRoot.setAttribute(POST_DETAIL_BOUND_ATTR, 'true');

	const commentsContainer = nextRoot.querySelector('[data-comments]');
	if (!commentsContainer) {
		return { post, comments: [] };
	}

	const commentForm = nextRoot.querySelector('#comment-form');
	bindCommentForm({ form: commentForm, fetchRef, postId, commentsContainer });

	const comments = await reloadComments(fetchRef, postId, commentsContainer);
	return { post, comments };
}
