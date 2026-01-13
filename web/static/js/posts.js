// js/posts.js
import {
  API_BASE,
  formatCreatedAt,
  resolveUsername,
  escapeHTML,
} from "./utils.js";

/* ==================================================
   POST CARD
================================================== */

export function renderPostCard(post, { clickable = true, showStatusToggle = false, showDelete = false } = {}) {
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
    article.querySelectorAll(".clickable").forEach((el) => {
      el.addEventListener("click", () => {
        window.location.href = `/view-post/${post.id}`;
      });
    });
  }

  // stop bubbling so card click doesn't trigger
  article
    .querySelectorAll(".post-actions, .post-comments, button, textarea, form")
    .forEach((el) => {
      el.addEventListener("click", (e) => e.stopPropagation());
    });

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
      comments.forEach((c) => list.appendChild(renderComment(c)));
    }

    container.appendChild(list);
    list.scrollTop = list.scrollHeight;

    await maybeRenderCommentForm(container, postId);
  } catch (err) {
    console.error("Failed to load comments", err);
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
      <div class="reaction">
        <span data-like-count>${comment.likes ?? 0}</span>
        <button
          class="btn btn-ghost"
          type="button"
          data-reaction="like"
          data-comment-id="${comment.id}"
        >
          Like
        </button>
      </div>

      <div class="reaction">
        <span data-dislike-count>${comment.dislikes ?? 0}</span>
        <button
          class="btn btn-ghost"
          type="button"
          data-reaction="dislike"
          data-comment-id="${comment.id}"
        >
          Dislike
        </button>
      </div>
    </div>
  `;

  return div;
}

/* ==================================================
   COMMENT FORM
================================================== */

async function maybeRenderCommentForm(container, postId) {
  try {
    const meRes = await fetch(`${API_BASE}/users/me`, {
      credentials: "include",
    });
    if (!meRes.ok) return;

    const payload = await meRes.json();
    const me = payload?.data;

    const form = document.createElement("form");
    form.className = "comment-form";

    form.innerHTML = `
      <textarea placeholder="Write a comment..." rows="2" required></textarea>
      <button class="btn btn-primary" type="submit">Comment</button>
    `;

    // stop bubbling
    ["click", "mousedown", "keydown", "submit"].forEach((evt) => {
      form.addEventListener(evt, (e) => {
        e.stopPropagation();
        if (evt === "submit") e.preventDefault();
      });
    });

    form.addEventListener("submit", async () => {
      const textarea = form.querySelector("textarea");
      const body = textarea.value.trim();
      if (!body) return;

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

      const createdPayload = await res.json();
      const newComment = createdPayload?.data ?? createdPayload;

      textarea.value = "";

      const commentsList = container.querySelector(".comments-scroll");
      if (!commentsList) return;

      const commentEl = document.createElement("div");
      commentEl.className = "comment";
      commentEl.innerHTML = `
        <div class="comment-meta muted">
          <strong>${escapeHTML(me?.username || "You")}</strong>
          · ${formatCreatedAt(newComment.created_at)}
        </div>
        <div class="comment-body">
          ${escapeHTML(newComment.body)}
        </div>
      `;

      commentsList.appendChild(commentEl);
      commentsList.scrollTop = commentsList.scrollHeight;
    });

    container.appendChild(form);
  } catch (err) {
    console.error("Failed to render comment form", err);
  }
}

/* ==================================================
   REACTIONS
================================================== */

export function reactionTemplate(post) {
  return `
    <div class="reaction">
      <span data-like-count>${post.likes ?? 0}</span>
      <button class="btn btn-ghost" type="button" data-reaction="like" data-post-id="${post.id}">
        Like
      </button>
    </div>
    <div class="reaction">
      <span data-dislike-count>${post.dislikes ?? 0}</span>
      <button class="btn btn-ghost" type="button" data-reaction="dislike" data-post-id="${post.id}">
        Dislike
      </button>
    </div>
  `;
}

let reactionsBound = false;

export function initReactions() {
  document.addEventListener(
    "click",
    async e => {
      const btn = e.target.closest("[data-reaction]");
      if (!btn) return;

      e.preventDefault();
      e.stopPropagation();

      const type = btn.dataset.reaction;
      const postId = btn.dataset.postId;
      const commentId = btn.dataset.commentId;

      let url = null;

      if (postId) {
        url = `${API_BASE}/posts/${postId}/${type}`;
      } else if (commentId) {
        url = `${API_BASE}/comments/${commentId}/${type}`;
      } else {
        return;
      }

      try {
        const res = await fetch(url, {
          method: "POST",
          credentials: "include",
          headers: { Accept: "application/json" },
        });

        if (!res.ok) {
          alert("You must be logged in to react");
          return;
        }

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

function statusToggleTemplate(post) {
  if (!post.status) return "";

  const isDraft = post.status === "draft";

  return `
    <button
      class="btn btn-outline btn-sm post-status-toggle"
      data-post-id="${post.id}"
      data-current-status="${post.status}"
    >
      ${isDraft ? "Publish" : "Draft"}
    </button>
  `;
}

function renderCategories(categories = []) {
  if (!categories.length) return "";

  return `
    <div class="post-categories">
      ${categories.map(c => `
        <span class="category-badge">${c.name}</span>
      `).join("")}
    </div>
  `;
}

function deletePostTemplate(post) {
  return `
    <button
      class="btn btn-danger btn-sm post-delete"
      type="button"
      data-post-id="${post.id}"
      aria-label="Delete post"
    >
      Delete
    </button>
  `;
}
