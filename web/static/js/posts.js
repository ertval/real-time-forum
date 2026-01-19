// web/static/js/posts.js

import {
  API_BASE,
  formatCreatedAt,
  resolveUsername,
  escapeHTML,
} from "./utils.js";

import { Auth } from "./auth.js";

/* ==================================================
   POST CARD
================================================== */

export function renderPostCard(
  post,
  { clickable = true, showStatusToggle = false, showDelete = false } = {}
) {
  const article = document.createElement("article");
  article.className = "post card card-pad";
  article.dataset.postId = post.id;

  article.innerHTML = `
    <header class="post-header ${clickable ? "clickable" : ""}">
      <div>
        ${renderCategories(post.categories)}
        <h3 class="post-title">${escapeHTML(post.title)}</h3>
        <p class="muted">Author: ${resolveUsername(post)}</p>
      </div>

      <div class="post-header-right">
        <time class="muted">${formatCreatedAt(post.created_at)}</time>
        ${showStatusToggle ? statusToggleTemplate(post) : ""}
        ${showDelete ? deletePostTemplate(post) : ""}
      </div>
    </header>

    <section class="post-body ${clickable ? "clickable" : ""}">
      <p>${escapeHTML(post.body)}</p>
    </section>

    <section class="post-actions">
      ${reactionTemplate(post)}
    </section>

    <section class="post-comments" data-comments></section>
  `;

  if (clickable) {
    article.querySelectorAll(".clickable").forEach(el => {
      el.addEventListener("click", () => {
        window.location.href = `/view-post/${post.id}`;
      });
    });
  }

  // Prevent bubbling
  article
    .querySelectorAll(".post-actions, .post-comments, button, textarea, form")
    .forEach(el => el.addEventListener("click", e => e.stopPropagation()));

  return article;
}

/* ==================================================
   COMMENTS
================================================== */

function extractArray(payload) {
  if (Array.isArray(payload)) return payload;
  if (payload && Array.isArray(payload.data)) return payload.data;
  return [];
}

export async function loadPostCommentsPreview(postId, article) {
  const container = article.querySelector("[data-comments]");
  if (!container) return;

  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (!res.ok) return;

    const payload = await res.json();
    const comments = extractArray(payload);

    const list = document.createElement("div");
    list.className = "comments comments-scroll";

    if (comments.length === 0) {
      list.innerHTML = `<p class="muted">No comments yet.</p>`;
    } else {
      comments.forEach(c => list.appendChild(renderComment(c)));
    }

    container.appendChild(list);
    list.scrollTop = list.scrollHeight;

    maybeRenderCommentForm(container, postId);
  } catch (err) {
    console.error("Failed to load comments:", err);
  }
}

function renderComment(comment) {
  const div = document.createElement("div");
  div.className = "comment";
  div.dataset.commentId = comment.id;

  div.innerHTML = `
    <div class="comment-meta muted">
      <strong>${resolveUsername(comment)}</strong>
      · ${formatCreatedAt(comment.created_at)}
    </div>

    <div class="comment-body">
      ${escapeHTML(comment.body)}
    </div>

    <div class="comment-actions">
      ${reactionTemplate(comment, true)}
    </div>
  `;

  return div;
}

/* ==================================================
   COMMENT FORM (AUTH-BASED)
================================================== */

function maybeRenderCommentForm(container, postId) {
  // Guest → do nothing (EXPECTED)
  if (!Auth.isAuthenticated) return;

  const form = document.createElement("form");
  form.className = "comment-form";

  form.setAttribute("novalidate", "novalidate");

  form.innerHTML = `
    <textarea placeholder="Write a comment..." rows="2"></textarea>
    <p class="comment-error" role="alert" hidden></p>
    <button class="btn btn-primary" type="submit">Comment</button>
  `;

  ["click", "mousedown", "keydown", "submit"].forEach(evt => {
    form.addEventListener(evt, e => {
      e.stopPropagation();
      if (evt === "submit") e.preventDefault();
    });
  });

  form.addEventListener("submit", async () => {
    const textarea = form.querySelector("textarea");

    /* clears possible previous error as user types */
    textarea.addEventListener("input", () => {
      const errorEl = form.querySelector(".comment-error");
      if (!errorEl) return;

      if (textarea.value.trim()) {
        errorEl.textContent = "";
        errorEl.hidden = true;
      }
    });

    const errorEl = form.querySelector(".comment-error");
    const body = textarea.value.trim();
    if (!body) {
      errorEl.textContent = "Cannot submit an empty comment";
      errorEl.hidden = false;
      textarea.focus();
      return;
    }

    errorEl.textContent = "";
    errorEl.hidden = true;

    const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ body }),
    });

    if (!res.ok) {
      alert("You must be logged in to comment.");
      return;
    }

    const payload = await res.json();
    const newComment = payload.data ?? payload;

    textarea.value = "";

    const commentsList = container.querySelector(".comments-scroll");
    if (!commentsList) return;

    commentsList.appendChild(renderComment(newComment));
    commentsList.scrollTop = commentsList.scrollHeight;
  });

  container.appendChild(form);
}

/* ==================================================
   REACTIONS
================================================== */

export function reactionTemplate(item, isComment = false) {
  const idAttr = isComment
    ? `data-comment-id="${item.id}"`
    : `data-post-id="${item.id}"`;

  return `
    <div class="reaction">
      <span data-like-count>${item.likes ?? 0}</span>
      <button class="btn btn-ghost" data-reaction="like" ${idAttr}>Like</button>
    </div>
    <div class="reaction">
      <span data-dislike-count>${item.dislikes ?? 0}</span>
      <button class="btn btn-ghost" data-reaction="dislike" ${idAttr}>Dislike</button>
    </div>
  `;
}

export function initReactions() {
  document.addEventListener(
    "click",
    async e => {
      const btn = e.target.closest("[data-reaction]");
      if (!btn) return;

      e.preventDefault();
      e.stopPropagation();

      if (!Auth.isAuthenticated) {
        alert("You must be logged in to react.");
        return;
      }

      const type = btn.dataset.reaction;
      const postId = btn.dataset.postId;
      const commentId = btn.dataset.commentId;

      const url = postId
        ? `${API_BASE}/posts/${postId}/${type}`
        : `${API_BASE}/comments/${commentId}/${type}`;

      try {
        const res = await fetch(url, {
          method: "POST",
          credentials: "include",
          headers: { Accept: "application/json" },
        });

        if (!res.ok) return;

        const { data } = await res.json();

        const container = btn.closest(
          postId
            ? `article[data-post-id="${postId}"]`
            : `div[data-comment-id="${commentId}"]`
        );

        if (!container) return;

        container.querySelector("[data-like-count]").textContent =
          data.likes_count;
        container.querySelector("[data-dislike-count]").textContent =
          data.dislikes_count;
      } catch (err) {
        console.error("Reaction failed:", err);
      }
    },
    true
  );
}

/* ==================================================
   HELPERS
================================================== */

function statusToggleTemplate(post) {
  if (!post.status) return "";
  return `
    <button class="btn btn-outline btn-sm post-status-toggle"
      data-post-id="${post.id}"
      data-current-status="${post.status}">
      ${post.status === "draft" ? "Publish" : "Draft"}
    </button>
  `;
}

function renderCategories(categories = []) {
  if (!categories.length) return "";
  return `
    <div class="post-categories">
      ${categories.map(c => `<span class="category-badge">${c.name}</span>`).join("")}
    </div>
  `;
}

function deletePostTemplate(post) {
  return `
    <button class="btn btn-danger btn-sm post-delete"
      data-post-id="${post.id}">
      Delete
    </button>
  `;
}
