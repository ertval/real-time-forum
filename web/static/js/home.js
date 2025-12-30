// web/static/js/home.js

const API_BASE = "http://localhost:8080/api/v1";

document.addEventListener("DOMContentLoaded", () => {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");

  if (!output) return;

  loadPublicPosts(output, empty);
});

async function loadPublicPosts(output, emptyEl) {
  try {
    const res = await fetch(`${API_BASE}/posts/public`, {
      method: "GET",
      credentials: "include", // safe even if not needed
      headers: { "Accept": "application/json" },
    });

    const payload = await res.json().catch(() => null);

    if (!res.ok) {
      console.error("Failed to load posts:", res.status, payload);
      showEmpty(output, emptyEl);
      return;
    }

    // Your API shape: { data: [...], meta: {...} } OR { data: null, meta: {...} }
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
  // PublicPost fields from backend: id, title, body, author, categories[], likes, dislikes, created_at
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
        <h3 id="post-${post.id}-title" class="post-title">${escapeHtml(post.title ?? "")}</h3>
        <p class="post-author muted">Author: ${escapeHtml(post.author ?? "")}</p>
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
        <button class="btn btn-secondary" type="button" disabled>Load comments</button>
      </div>
    </section>
  `;

  return article;
}

function formatCreatedAt(iso) {
  if (!iso) return "";
  // handles: 2025-12-30T10:20:30Z or "2025-12-30 10:20"
  return String(iso).replace("T", " ").replace("Z", "").slice(0, 16);
}

// Minimal escaping to prevent HTML injection
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
