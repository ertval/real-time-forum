import { escapeHTML } from '../../core/utils/html.js';

export function formatCreatedAt(iso) {
	if (!iso) {
		return '';
	}

	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) {
		return String(iso);
	}

	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const day = String(date.getDate()).padStart(2, '0');
	let hours = date.getHours();
	const minutes = String(date.getMinutes()).padStart(2, '0');
	const suffix = hours >= 12 ? 'PM' : 'AM';
	hours %= 12;
	hours = hours === 0 ? 12 : hours;

	return `${year}-${month}-${day}, ${String(hours).padStart(2, '0')}:${minutes} ${suffix}`;
}

export function resolveUsername(post) {
	return post.username || post.author || (post.user_id ? `User ${post.user_id}` : 'User');
}

export function renderCategories(categories = []) {
	if (!Array.isArray(categories) || categories.length === 0) {
		return '';
	}

	return `
		<ul class="post-categories" aria-label="Post categories">
			${categories
				.map(
					(category) => `<li class="post-category">${escapeHTML(category.name ?? category)}</li>`,
				)
				.join('')}
		</ul>
	`;
}

function reactionTemplate(post) {
	return `
		<div class="reactions" data-reaction-scope="post">
			<label class="reaction-pill">
				<input
					type="checkbox"
					data-reaction="like"
					data-post-id="${escapeHTML(post.id)}"
					${post.current_user_reaction === 'like' ? 'checked' : ''}
				/>
				<span>Like</span>
				<span data-like-count>${Number(post.likes_count ?? 0)}</span>
			</label>

			<label class="reaction-pill">
				<input
					type="checkbox"
					data-reaction="dislike"
					data-post-id="${escapeHTML(post.id)}"
					${post.current_user_reaction === 'dislike' ? 'checked' : ''}
				/>
				<span>Dislike</span>
				<span data-dislike-count>${Number(post.dislikes_count ?? 0)}</span>
			</label>
		</div>
	`;
}

function renderCommentPreview(comment) {
	const body = typeof comment.body === 'string' ? comment.body.trim() : '';
	const previewBody = body ? `<p class="comment-text">${escapeHTML(body)}</p>` : '';

	return `
		<article class="comment" data-comment-id="${escapeHTML(comment.id)}">
			<div class="comment-meta muted">
				<strong><a data-link class="profile-link" href="/profile/${escapeHTML(comment.author_id || comment.user_id || '')}">${escapeHTML(resolveUsername(comment))}</a></strong>
				<span>${formatCreatedAt(comment.created_at)}</span>
			</div>
			<div class="comment-body">
				${previewBody}
			</div>
		</article>
	`;
}

function commentPreviewTemplate(previewComments = [], showCommentPreview = false) {
	if (!showCommentPreview) {
		return '';
	}

	if (!Array.isArray(previewComments) || previewComments.length === 0) {
		return `
			<section class="post-comments" data-comments-preview>
				<p class="muted">No comments yet.</p>
			</section>
		`;
	}

	return `
		<section class="post-comments" data-comments-preview>
			<div class="comments comments-scroll">
				${previewComments.map((comment) => renderCommentPreview(comment)).join('')}
			</div>
		</section>
	`;
}

export function renderPostCard(
	documentRef,
	post,
	{ clickable = true, onNavigate = null, previewComments = [], showCommentPreview = false } = {},
) {
	const article = documentRef.createElement('article');
	article.className = 'post card card-pad';
	article.dataset.postId = String(post.id);

	const imageUrl =
		typeof post.image_url === 'string' && post.image_url.trim() ? post.image_url : '';
	const imageMarkup = imageUrl
		? `
			<div class="post-image">
				<div class="post-image-ambient" aria-hidden="true"></div>
				<img
					src="${escapeHTML(imageUrl)}"
					alt="${escapeHTML(post.title)}"
					loading="lazy"
				/>
			</div>
		`
		: '';
	const bodyMarkup = post.body ? `<p>${escapeHTML(post.body)}</p>` : '';

	article.innerHTML = `
		<header class="post-header ${clickable ? 'clickable' : ''}">
			<div>
				${renderCategories(post.categories)}
				<h3 class="post-title">${escapeHTML(post.title)}</h3>
				<p class="muted">Author: <a data-link class="profile-link" href="/profile/${escapeHTML(post.author_id || post.user_id || '')}">${escapeHTML(resolveUsername(post))}</a></p>
			</div>

			<div class="post-header-right">
				<time class="muted">${formatCreatedAt(post.created_at)}</time>
			</div>
		</header>

		<section class="post-body ${clickable ? 'clickable' : ''}">
			${imageMarkup}
			${bodyMarkup}
		</section>

		<section class="post-actions">
			${reactionTemplate(post)}
		</section>

		${commentPreviewTemplate(previewComments, showCommentPreview)}
	`;

	if (clickable) {
		for (const element of article.querySelectorAll('.clickable')) {
			element.addEventListener('click', () => {
				if (typeof onNavigate === 'function') {
					onNavigate(post);
				}
			});
		}
	}

	return article;
}
