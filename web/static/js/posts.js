import {
  API_BASE,
  formatCreatedAt,
  resolveUsername,
  escapeHTML
} from "./utils.js";

/* ==================================================
   CATEGORY FILTER
================================================== */

async function loadCategories() {
  const res = await fetch(`${API_BASE}/categories`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) throw new Error("Failed to load categories");
  return await res.json();
}

function extractArray(payload) {
  if (Array.isArray(payload)) return payload;
  if (payload && Array.isArray(payload.data)) return payload.data;
  return [];
}

async function populateCategorySelect() {
  const select = document.getElementById("categoryFilter");
  if (!select) return;

  const payload = await loadCategories();
  const categories = extractArray(payload);

  // remove previous options except "All"
  select.querySelectorAll("option:not(:first-child)").forEach(o => o.remove());

  categories.forEach(c => {
    const opt = document.createElement("option");
    opt.value = c.id;
    opt.textContent = c.name;
    select.appendChild(opt);
  });
}

function getCategoryIdFromURL() {
  const params = new URLSearchParams(window.location.search);
  return params.get("category_id") || "";
}

function setCategoryIdToURL(categoryId) {
  const url = new URL(window.location.href);
  if (categoryId) {
    url.searchParams.set("category_id", categoryId);
  } else {
    url.searchParams.delete("category_id");
  }
  history.pushState({}, "", url);
}

/* ==================================================
   POSTS
================================================== */

async function fetchPosts(categoryId = "") {
  const url = new URL(`${API_BASE}/posts`, window.location.origin);
  if (categoryId) url.searchParams.set("category_id", categoryId);

  const res = await fetch(url, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) throw new Error("Failed to load posts");
  return await res.json();
}

async function refreshPosts(categoryId = "") {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");
  if (!output) return;

  output.innerHTML = "";

  const payload = await fetchPosts(categoryId);
  const posts = extractArray(payload);

  if (posts.length === 0) {
    empty.hidden = false;
    return;
  }

  empty.hidden = true;

  posts.forEach(post => {
    const card = renderPostCard(post);
    output.appendChild(card);
    loadPostCommentsPreview(post.id, card);
  });
}

/* ==================================================
   POST CARD
================================================== */

export function renderPostCard(post, { clickable = true } = {}) {
  const article = document.createElement("article");
  article.className = "post card card-pad";
  article.dataset.postId = post.id;

  article.innerHTML = `
    <header class="post-header ${clickable ? "clickable" : ""}">
      <div>
        <h3 class="post-title">${escapeHTML(post.title)}</h3>
        <p class="muted">Author: ${resolveUsername(post)}</p>
      </div>
      <time class="muted">${formatCreatedAt(post.created_at)}</time>
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

  article
    .querySelectorAll(".post-actions, .post-comments, button, textarea, form")
    .forEach(el => {
      el.addEventListener("click", e => e.stopPropagation());
    });

  return article;
}

/* ==================================================
   COMMENTS
================================================== */

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

    await maybeRenderCommentForm(container, postId);
  } catch (err) {
    console.error("Failed to load comments", err);
  }
}

function renderComment(comment) {
  const div = document.createElement("div");
  div.className = "comment";

  div.innerHTML = `
    <div class="comment-meta muted">
      <strong>${resolveUsername(comment)}</strong>
      · ${formatCreatedAt(comment.created_at)}
    </div>
    <div class="comment-body">
      ${escapeHTML(comment.body)}
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

    const { data: me } = await meRes.json();

    const form = document.createElement("form");
    form.className = "comment-form";

    form.innerHTML = `
      <textarea placeholder="Write a comment..." rows="2" required></textarea>
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

      const { data: newComment } = await res.json();
      textarea.value = "";

      const commentsList = container.querySelector(".comments-scroll");
      if (!commentsList) return;

      const commentEl = document.createElement("div");
      commentEl.className = "comment";
      commentEl.innerHTML = `
        <div class="comment-meta muted">
          <strong>${escapeHTML(me.username)}</strong>
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
      if (!type || !postId) return;

      try {
        const res = await fetch(`${API_BASE}/posts/${postId}/${type}`, {
          method: "POST",
          credentials: "include",
          headers: { Accept: "application/json" },
        });

        if (!res.ok) {
          alert("You must be logged in to react");
          return;
        }

        const { data } = await res.json();

        const article = document.querySelector(
          `article[data-post-id="${postId}"]`
        );
        if (article) {
          article.querySelector("[data-like-count]").textContent = data.likes_count;
          article.querySelector("[data-dislike-count]").textContent = data.dislikes_count;
        }
      } catch (err) {
        console.error("Reaction failed:", err);
      }
    },
    true
  );
}

/* ==================================================
   INIT
================================================== */

document.addEventListener("DOMContentLoaded", async () => {
  const select = document.getElementById("categoryFilter");
  if (!select) return;

  await populateCategorySelect();

  const categoryId = getCategoryIdFromURL();
  select.value = categoryId;

  await refreshPosts(categoryId);

  select.addEventListener("change", async () => {
    setCategoryIdToURL(select.value);
    await refreshPosts(select.value);
  });

  window.addEventListener("popstate", async () => {
    const id = getCategoryIdFromURL();
    select.value = id;
    await refreshPosts(id);
  });

  initReactions();
});
