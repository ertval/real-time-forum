// web/static/js/home.js

const API_BASE = "/api/v1";

document.addEventListener("DOMContentLoaded", () => {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");

  if (output) {
    loadPublicPosts(output, empty);
  }

  loadGreeting();
});

// --------------------------------------------------
// POSTS
// --------------------------------------------------

async function loadPublicPosts(output, emptyEl) {
  try {
    const res = await fetch(`${API_BASE}/posts/public`, {
      method: "GET",
      credentials: "include",
      headers: { "Accept": "application/json" },
    });

    const payload = await res.json().catch(() => null);

    if (!res.ok) {
      console.error("Failed to load posts:", res.status, payload);
      showEmpty(output, emptyEl);
      return;
    }

    const posts = Array.isArray(payload?.data) ? payload.data : [];

    if (posts.length === 0) {
      showEmpty(output, emptyEl);
      return;
    }

    if (emptyEl) emptyEl.hidden = true;
    output.innerHTML = "";

    for (const post of posts) {
      output.appendChild(renderPostCard(post));
    }
  } catch (err) {
    console.error("Error loading posts:", err);
    showEmpty(output, emptyEl);
  }
}

function showEmpty(output, emptyEl) {
  output.innerHTML = "";
  if (emptyEl) emptyEl.hidden = false;
}

function renderPostCard(post) {
  const article = document.createElement("article");
  article.className = "post card card-pad";
  article.setAttribute("aria-labelledby", `post-${post.id}-title`);

  const created = formatCreatedAt(post.created_at);
  const categories = Array.isArray(post.categories) ? post.categories : [];

  const categoriesHtml =
    categories.length > 0
      ? `<p class="muted post-categories">${escapeHtml(categories.join(", "))}</p>`
      : "";

  article.innerHTML = `
    <header class="post-header">
      <div class="post-meta">
        <h3 id="post-${post.id}-title" class="post-title">
          ${escapeHtml(post.title ?? "")}
        </h3>
        <p class="post-author muted">
          Author: ${escapeHtml(post.author ?? "")}
        </p>
        ${categoriesHtml}
      </div>

      <time class="post-time muted" datetime="${escapeAttr(post.created_at ?? "")}">
        ${escapeHtml(created)}
      </time>
    </header>

    <section class="post-body" aria-label="Post body">
      <p>${escapeHtml(post.body ?? "")}</p>
    </section>

    <section class="post-actions" aria-label="Post actions">
      <div class="reaction">
        <span class="count">${Number(post.likes ?? 0)}</span>
        <button class="btn btn-ghost" type="button" disabled>Like</button>
      </div>

      <div class="reaction">
        <span class="count">${Number(post.dislikes ?? 0)}</span>
        <button class="btn btn-ghost" type="button" disabled>Dislike</button>
      </div>

      <div class="action-right">
        <button class="btn btn-secondary" type="button" disabled>
          Load comments
        </button>
      </div>
    </section>
  `;

  return article;
}

// --------------------------------------------------
// GREETING
// --------------------------------------------------

async function loadGreeting() {
  const greetingEl = document.getElementById("greeting");
  if (!greetingEl) return;

  try {
    const res = await fetch(`${API_BASE}/users/me`, {
      credentials: "include",
    });

    if (!res.ok) {
      greetingEl.textContent = "Hello, Guest";
      return;
    }

    const payload = await res.json();
    const username = payload?.data?.username;

    greetingEl.textContent = username
      ? `Hello, ${username}`
      : "Hello, Guest";
  } catch {
    greetingEl.textContent = "Hello, Guest";
  }
}

// --------------------------------------------------
// HELPERS
// --------------------------------------------------

function formatCreatedAt(iso) {
  if (!iso) return "";
  return String(iso).replace("T", " ").replace("Z", "").slice(0, 16);
}

function escapeHtml(s) {
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function escapeAttr(s) {
  return escapeHtml(s);
}
