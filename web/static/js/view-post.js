import {
  loadPostCommentsPreview,
  initReactions,
} from "./posts.js";

import {
  API_BASE,
  formatCreatedAt,
  resolveUsername,
  escapeHTML,
} from "./utils.js";

document.addEventListener("DOMContentLoaded", async () => {
  const postId = window.location.pathname.split("/").pop();
  if (!postId) return;

  const article = document.querySelector("article[data-post-id]");
  if (!article) return;

  article.dataset.postId = postId;

  await loadPost(postId, article);

  // comments (reuse posts.js)
  const commentsContainer = article.querySelector("[data-comments]");
  if (commentsContainer) {
    await loadPostCommentsPreview(postId, article);
  }

  // reactions (shared with home)
  initReactions();
});

// --------------------------------------------------
// LOAD POST
// --------------------------------------------------

async function loadPost(postId, article) {
  const res = await fetch(`${API_BASE}/posts/${postId}`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) return;

  const { data: post } = await res.json();

  // title / body
  article.querySelector(".viewpost-title").textContent = post.title;
  article.querySelector(".viewpost-body-text").innerHTML =
    escapeHTML(post.body);

  // author / time
  article.querySelector(".viewpost-author").textContent =
    `Author: ${resolveUsername(post)}`;

  article.querySelector(".viewpost-time").textContent =
    formatCreatedAt(post.created_at);

  // reactions counts (SAME selectors as home)
  article.querySelector("[data-like-count]").textContent =
    post.likes ?? 0;

  article.querySelector("[data-dislike-count]").textContent =
    post.dislikes ?? 0;

  // buttons dataset (CRITICAL for initReactions)
  article
    .querySelectorAll("[data-reaction]")
    .forEach(btn => {
      btn.dataset.postId = postId;
    });
}
