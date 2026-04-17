// /static/js/activity/render-activity.js

import { deleteCommentButton, editCommentButton } from "../post-actions.js";
import { reactionTemplate, renderPostCard } from "../posts.js";
import { initReactions } from "../reactions.js";
import { escapeHTML, formatCreatedAt, resolveUsername } from "../utils.js";
import { loadActivity } from "./api-activity.js";
import { activityState } from "./state-activity.js";

function asItems(section) {
	return Array.isArray(section?.items) ? section.items : [];
}

function asPagination(section) {
	return section?.pagination ?? {};
}

export async function renderActivity(state, pager) {
	const payload = await loadActivity(state);
	if (!payload) return;

	const data = payload?.data ?? {};

	renderPostsSection(
		data.created_posts,
		"created-posts-output",
		"created-posts-empty",
		"created-count",
		{ showOwnerActions: true },
	);

	renderCommentsSection(data.comments);

	renderPostsSection(
		data.liked_posts,
		"liked-posts-output",
		"liked-posts-empty",
		"liked-count",
	);

	renderPostsSection(
		data.disliked_posts,
		"disliked-posts-output",
		"disliked-posts-empty",
		"disliked-count",
	);

	initReactions();

	const totalPages = Math.max(
		asPagination(data.created_posts).total_pages || 1,
		asPagination(data.comments).total_pages || 1,
		asPagination(data.liked_posts).total_pages || 1,
		asPagination(data.disliked_posts).total_pages || 1,
	);

	pager.set(state.page, totalPages);

	const paginationEl = document.getElementById("activity-pagination");
	if (paginationEl) {
		paginationEl.hidden = totalPages <= 1;
	}
}

function renderPostsSection(
	section,
	outputId,
	emptyId,
	countId,
	{ showOwnerActions = false } = {},
) {
	const output = document.getElementById(outputId);
	const empty = document.getElementById(emptyId);
	const count = document.getElementById(countId);
	if (!output || !empty || !count) return;

	output.innerHTML = "";
	empty.hidden = true;

	const items = asItems(section);
	const pagination = asPagination(section);

	count.textContent = String(pagination.total ?? items.length);

	if (!items.length) {
		empty.hidden = false;
		return;
	}

	items.forEach((post) => {
		const isOwner =
			Number(post.author_id) === Number(activityState.activityUserID);
		const enableActions = showOwnerActions && isOwner;

		const article = renderPostCard(post, {
			clickable: true,
			showStatusToggle: enableActions,
			showDelete: enableActions,
			showEdit: enableActions,
		});

		article.querySelector(".post-comments")?.remove();
		output.appendChild(article);
	});
}

function renderCommentsSection(section) {
	const output = document.getElementById("comments-output");
	const empty = document.getElementById("comments-empty");
	const count = document.getElementById("comments-count");
	if (!output || !empty || !count) return;

	output.innerHTML = "";
	empty.hidden = true;

	const items = asItems(section);
	const pagination = asPagination(section);

	count.textContent = String(pagination.total ?? items.length);

	if (!items.length) {
		empty.hidden = false;
		return;
	}

	items.forEach((comment) => {
		const commentID = Number(comment.id) || 0;
		const username = resolveUsername(comment);
		const commentBody = typeof comment.body === "string" ? comment.body : "";
		const commentImageURL =
			typeof comment.image_url === "string" && comment.image_url.trim()
				? comment.image_url.trim()
				: "";

		const bodyMarkup = commentBody.trim()
			? `<p class="activity-comment-body">${escapeHTML(commentBody)}</p>`
			: "";

		const imageMarkup = commentImageURL
			? `
          <div class="activity-comment-image-wrap">
            <img
              class="activity-comment-image"
              src="${escapeHTML(commentImageURL)}"
              alt="Comment image by ${escapeHTML(username)}"
              loading="lazy"
            />
          </div>
        `
			: "";

		const postForCard = mapActivityCommentPostToCard(comment);
		const postArticle = renderPostCard(postForCard, { clickable: true });
		postArticle.classList.add("activity-comment-related-post");
		postArticle.querySelector(".post-comments")?.remove();

		const article = document.createElement("article");
		article.className = "activity-comment card card-pad";
		article.dataset.commentId = commentID;
		article.dataset.commentBody = commentBody;
		article.dataset.commentImageUrl = commentImageURL;

		article.innerHTML = `
      <header class="activity-comment-head">
        <p class="activity-comment-author muted">
          Author: ${escapeHTML(username)}
        </p>
        <div class="activity-comment-head-right">
          <time class="muted">${formatCreatedAt(comment.created_at)}</time>
          ${editCommentButton(commentID)}
          ${deleteCommentButton(commentID)}
        </div>
      </header>
      <div class="activity-comment-content">
        ${bodyMarkup}
        ${imageMarkup}
      </div>
      <div class="activity-comment-reactions comment">
        ${reactionTemplate(comment, true)}
      </div>
    `;

		const entry = document.createElement("div");
		entry.className = "activity-comment-entry";
		entry.appendChild(postArticle);
		entry.appendChild(article);

		output.appendChild(entry);
	});
}

function mapActivityCommentPostToCard(comment) {
	const post = comment?.post ?? {};
	const categories = normalizePostCategories(post, comment);
	const postReaction = Number(
		post?.my_reaction ??
			post?.myReaction ??
			comment?.post_my_reaction ??
			comment?.postMyReaction ??
			0,
	);

	const postID = Number(comment?.post_id ?? post.id) || 0;
	const createdAt =
		typeof post.created_at === "string" && post.created_at.trim()
			? post.created_at
			: comment?.created_at;

	return {
		id: postID,
		author_id: Number(post.author_id) || 0,
		author: typeof post.author === "string" ? post.author : "",
		title:
			typeof post.title === "string" && post.title.trim()
				? post.title
				: "Untitled",
		body: typeof post.body === "string" ? post.body : "",
		image_url: typeof post.image_url === "string" ? post.image_url : "",
		created_at: createdAt,
		categories,
		likes: Number(post.likes ?? post.likes_count) || 0,
		dislikes: Number(post.dislikes ?? post.dislikes_count) || 0,
		my_reaction: Number.isFinite(postReaction) ? postReaction : 0,
	};
}

function normalizePostCategories(post, comment) {
	const fromNested = Array.isArray(post?.categories) ? post.categories : [];
	const fromComment = Array.isArray(comment?.categories)
		? comment.categories
		: [];
	const source = fromNested.length ? fromNested : fromComment;

	return source
		.map((category) => {
			if (!category || typeof category !== "object") return null;
			const id = Number(category.id ?? category.ID) || 0;
			const name =
				typeof category.name === "string"
					? category.name
					: typeof category.Name === "string"
						? category.Name
						: "";
			if (!id && !name) return null;
			return { id, name };
		})
		.filter(Boolean);
}
