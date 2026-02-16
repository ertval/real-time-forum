// web/static/js/posts.js

import {
  API_BASE,
  formatCreatedAt,
  resolveUsername,
  escapeHTML,
} from "./utils.js";

import { Auth } from "./auth.js";

/*-----------
  POST CARD
-----------*/

export function renderPostCard(
  post,
  { clickable = true, showStatusToggle = false, showDelete = false } = {}
) {
  const article = document.createElement("article");
  article.className = "post card card-pad";
  article.dataset.postId = post.id;

  const imageUrl =
    typeof post.image_url === "string" && post.image_url.trim()
      ? post.image_url
      : "";

  const imageMarkup = imageUrl
    ? `
      <div class="post-image">
        <img src="${imageUrl}" alt="${escapeHTML(post.title)}" loading="lazy" />
      </div>
    `
    : "";

  const bodyMarkup = post.body
    ? `<p>${escapeHTML(post.body)}</p>`
    : "";

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
      ${imageMarkup}
      ${bodyMarkup}
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

  article
    .querySelectorAll(".post-comments, button, textarea, form")
    .forEach(el =>
      el.addEventListener("click", e => e.stopPropagation())
    );

  return article;
}

/*----------
  COMMENTS
----------*/

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

/*--------------
  COMMENT FORM
--------------*/

function maybeRenderCommentForm(container, postId) {
  const form = document.createElement("form");
  form.className = "comment-form";
  form.noValidate = true;

  form.innerHTML = `
    <textarea placeholder="Write a comment..." rows="2"></textarea>
    <p class="comment-error" role="alert" hidden></p>
    <button class="btn btn-primary" type="submit">Comment</button>
  `;

  ["click", "mousedown", "keydown", "submit"].forEach(evt =>
    form.addEventListener(evt, e => {
      e.stopPropagation();
      if (evt === "submit") e.preventDefault();
    })
  );

  form.addEventListener("submit", async () => {
    const allowed = await Auth.requireOrPrompt();
    if (!allowed) return;

    const textarea = form.querySelector("textarea");
    const errorEl = form.querySelector(".comment-error");
    const body = textarea.value.trim();

    if (!body) {
      errorEl.textContent = "Cannot submit an empty comment";
      errorEl.hidden = false;
      return;
    }

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

    if (!res.ok) return;

    const payload = await res.json();
    const newComment = payload.data ?? payload;

    textarea.value = "";

    const list = container.querySelector(".comments-scroll");
    list?.appendChild(renderComment(newComment));
    list.scrollTop = list.scrollHeight;
  });

  container.appendChild(form);
}

/*--------------------
  REACTIONS TEMPLATE
--------------------*/

export function reactionTemplate(item, isComment = false) {
  const idAttr = isComment
    ? `data-comment-id="${item.id}"`
    : `data-post-id="${item.id}"`;

  return `
    <div class="reaction">
      <label class="reaction-toggle reaction-toggle--like" aria-label="Like">
        <input type="checkbox" data-reaction="like" ${idAttr}>
        <svg viewBox="0 0 32 32" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg">
          <path d="M29.845,17.099l-2.489,8.725C26.989,27.105,25.804,28,24.473,28H11c-0.553,0-1-0.448-1-1V13
          c0-0.215,0.069-0.425,0.198-0.597l5.392-7.24C16.188,4.414,17.05,4,17.974,4C19.643,4,21,5.357,21,7.026V12h5.002
          c1.265,0,2.427,0.579,3.188,1.589C29.954,14.601,30.192,15.88,29.845,17.099z"></path>
          <path d="M7,12H3c-0.553,0-1,0.448-1,1v14c0,0.552,0.447,1,1,1h4c0.553,0,1-0.448,1-1V13C8,12.448,7.553,12,7,12z
          M5,25.5c-0.828,0-1.5-0.672-1.5-1.5c0-0.828,0.672-1.5,1.5-1.5c0.828,0,1.5,0.672,1.5,1.5C6.5,24.828,5.828,25.5,5,25.5z"></path>
        </svg>
      </label>
      <span class="data-like-count" data-like-count>${item.likes ?? 0}</span>
    </div>

    <div class="reaction">
      <label class="reaction-toggle reaction-toggle--dislike" aria-label="Dislike">
        <input type="checkbox" data-reaction="dislike" ${idAttr}>
        <svg viewBox="0 0 32 32" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg">
          <path d="M2.156,14.901l2.489-8.725C5.012,4.895,6.197,4,7.528,4h13.473C21.554,4,22,4.448,22,5v14
          c0,0.215-0.068,0.425-0.197,0.597l-5.392,7.24C15.813,27.586,14.951,28,14.027,28c-1.669,0-3.026-1.357-3.026-3.026V20H5.999
          c-1.265,0-2.427-0.579-3.188-1.589C2.047,17.399,1.809,16.12,2.156,14.901z"></path>
          <path d="M25.001,20h4C29.554,20,30,19.552,30,19V5c0-0.552-0.446-1-0.999-1h-4c-0.553,0-1,0.448-1,1v14
          C24.001,19.552,24.448,20,25.001,20z M27.001,6.5c0.828,0,1.5,0.672,1.5,1.5c0,0.828-0.672,1.5-1.5,1.5c-0.828,0-1.5-0.672-1.5-1.5
          C25.501,7.172,26.173,6.5,27.001,6.5z"></path>
        </svg>
      </label>
      <span class="data-dislike-count" data-dislike-count>${item.dislikes ?? 0}</span>
    </div>
  `;
}

/*---------
  HELPERS
---------*/

function extractArray(payload) {
  if (Array.isArray(payload)) return payload;
  if (payload?.data && Array.isArray(payload.data)) return payload.data;
  if (payload?.data?.data && Array.isArray(payload.data.data)) {
    return payload.data.data;
  }
  return [];
}

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
  if (!Array.isArray(categories) || categories.length === 0) return "";
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
