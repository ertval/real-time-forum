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
    const res = await fetch(`${API_BASE}/posts`, { // `${API_BASE}/posts/public ???
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

  // -----------------------
  // Header
  // -----------------------
  const header = document.createElement("header");
  header.className = "post-header";

  const meta = document.createElement("div");
  meta.className = "post-meta";

  const title = document.createElement("h3");
  title.id = `post-${post.id}-title`;
  title.className = "post-title";
  title.textContent = post.title ?? "";

  const author = document.createElement("p");
  author.className = "post-author muted";
  author.textContent = `Author: ${post.author ?? ""}`;

  meta.appendChild(title);
  meta.appendChild(author);

  // Categories (separate responsibility)
  renderCategories(meta, post.categories);

  const time = document.createElement("time");
  time.className = "post-time muted";
  time.dateTime = post.created_at ?? "";
  time.textContent = formatCreatedAt(post.created_at);

  header.appendChild(meta);
  header.appendChild(time);

  // -----------------------
  // Body
  // -----------------------
  const body = document.createElement("section");
  body.className = "post-body";
  body.setAttribute("aria-label", "Post body");

  const bodyText = document.createElement("p");
  bodyText.textContent = post.body ?? "";

  body.appendChild(bodyText);

  // -----------------------
  // Actions
  // -----------------------
  const actions = document.createElement("section");
  actions.className = "post-actions";
  actions.setAttribute("aria-label", "Post actions");

  actions.appendChild(createReaction("Like", post.likes));
  actions.appendChild(createReaction("Dislike", post.dislikes));

  const right = document.createElement("div");
  right.className = "action-right";

  const commentsBtn = document.createElement("button");
  commentsBtn.className = "btn btn-secondary";
  commentsBtn.type = "button";
  commentsBtn.disabled = true;
  commentsBtn.textContent = "Load comments";

  right.appendChild(commentsBtn);
  actions.appendChild(right);

  // -----------------------
  // Assemble
  // -----------------------
  article.appendChild(header);
  article.appendChild(body);
  article.appendChild(actions);

  return article;
}

// --------------------------------------------------
// CATEGORIES 
// --------------------------------------------------

function renderCategories(container, categories) {
  if (!Array.isArray(categories) || categories.length === 0) return;

  const p = document.createElement("p");
  p.className = "muted post-categories";
  p.textContent = categories.join(", ");

  container.appendChild(p);
}

// --------------------------------------------------
// REACTIONS
// --------------------------------------------------

function createReaction(label, count) {
  const wrap = document.createElement("div");
  wrap.className = "reaction";

  const span = document.createElement("span");
  span.className = "count";
  span.textContent = Number(count ?? 0);

  const btn = document.createElement("button");
  btn.className = "btn btn-ghost";
  btn.type = "button";
  btn.disabled = true;
  btn.textContent = label;

  wrap.appendChild(span);
  wrap.appendChild(btn);

  return wrap;
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

  const d = new Date(iso); // ISO UTC → Date
  return d.toLocaleString(undefined, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatCreatedAt(iso) {
  if (!iso) return "";

  const d = new Date(iso);

  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");

  let hours = d.getHours();
  const minutes = String(d.getMinutes()).padStart(2, "0");

  const ampm = hours >= 12 ? "PM" : "AM";
  hours = hours % 12;
  hours = hours === 0 ? 12 : hours; // 12 AM / 12 PM
  hours = String(hours).padStart(2, "0");

  return `${year}-${month}-${day}, ${hours}:${minutes} ${ampm}`;
}
