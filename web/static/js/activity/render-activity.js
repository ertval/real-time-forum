// /static/js/activity/render-activity.js

import { loadActivity } from "./api-activity.js";
import { renderPostCard } from "../posts.js";
import { initReactions } from "../reactions.js";
import { formatCreatedAt, resolveUsername, escapeHTML } from "../utils.js";
import { editCommentButton, deleteCommentButton } from "../post-actions.js";
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
    "created-count"
  );

  renderCommentsSection(data.comments);

  renderPostsSection(
    data.liked_posts,
    "liked-posts-output",
    "liked-posts-empty",
    "liked-count"
  );

  renderPostsSection(
    data.disliked_posts,
    "disliked-posts-output",
    "disliked-posts-empty",
    "disliked-count"
  );

  initReactions();

  const totalPages = Math.max(
    asPagination(data.created_posts).total_pages || 1,
    asPagination(data.comments).total_pages || 1,
    asPagination(data.liked_posts).total_pages || 1,
    asPagination(data.disliked_posts).total_pages || 1
  );

  pager.set(state.page, totalPages);

  const paginationEl = document.getElementById("activity-pagination");
  if (paginationEl) {
    paginationEl.hidden = totalPages <= 1;
  }
}

function renderPostsSection(section, outputId, emptyId, countId) {
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

  items.forEach(post => {
    const isOwner =
      Number(post.author_id) === Number(activityState.activityUserID);

    const article = renderPostCard(post, {
      clickable: true,
      showStatusToggle: isOwner,
      showDelete: isOwner,
      showEdit: isOwner,
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

  items.forEach(comment => {
    const commentID = Number(comment.id) || 0;
    const username = resolveUsername(comment);

    const article = document.createElement("article");
    article.className = "activity-comment card card-pad";
    article.dataset.commentId = commentID;

    article.innerHTML = `
      <header class="activity-comment-head">
        <div>
          <p class="muted">On post</p>
          <a href="/view-post/${comment.post_id}" class="activity-comment-post-link">
            ${escapeHTML(comment.post?.title || "Untitled")}
          </a>
        </div>
        <div class="activity-comment-head-right">
          <time class="muted">${formatCreatedAt(comment.created_at)}</time>
          ${editCommentButton(commentID)}
          ${deleteCommentButton(commentID)}
        </div>
      </header>
      <p class="activity-comment-author muted">
        By ${escapeHTML(username)}
      </p>
      <p class="activity-comment-body">
        ${escapeHTML(comment.body || "")}
      </p>
    `;

    output.appendChild(article);
  });
}