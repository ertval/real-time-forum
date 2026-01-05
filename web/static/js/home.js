// web/static/js/home.js

document.addEventListener("DOMContentLoaded", () => {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");

  if (output) {
    loadPublicPosts(output, empty);
  }
});

// --------------------------------------------------
// POSTS
// --------------------------------------------------

async function loadPublicPosts(output, emptyEl) {
  try {
    const res = await fetch(`${API_BASE}/posts`, {
      method: "GET",
      credentials: "include",
      headers: { Accept: "application/json" },
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

    // 🔥 enable reactions AFTER render
    bindReactions(output);

  } catch (err) {
    console.error("Error loading posts:", err);
    showEmpty(output, emptyEl);
  }
}

function showEmpty(output, emptyEl) {
  output.innerHTML = "";
  if (emptyEl) emptyEl.hidden = false;
}

// --------------------------------------------------
// POST CARD
// --------------------------------------------------

function renderPostCard(post) {
  const article = document.createElement("article");
  article.className = "post card card-pad";
  article.dataset.postId = post.id;
  article.setAttribute("aria-labelledby", `post-${post.id}-title`);

  // Click → /view-post/{id}
  article.style.cursor = "pointer";
 
  article.addEventListener("click", (e) => {
    if (e.target.closest(".post-actions")) {
      return;
    }

    window.location.href = `/view-post/${post.id}`;
  });


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
  author.textContent = post.author
    ? `Author: ${post.author}`
    : `Author ID: ${post.author_id}`;

  meta.appendChild(title);
  meta.appendChild(author);

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

  actions.appendChild(createReaction("Like", post.likes, post.id));
  actions.appendChild(createReaction("Dislike", post.dislikes, post.id));

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
// REACTIONS (UI)
// --------------------------------------------------

function createReaction(label, count, postId) {
  const wrap = document.createElement("div");
  wrap.className = "reaction";

  const span = document.createElement("span");
  span.className = `count ${label === "Like" ? "like-count" : "dislike-count"}`;
  span.textContent = Number(count ?? 0);

  const btn = document.createElement("button");
  btn.className = "btn btn-ghost";
  btn.type = "button";
  btn.textContent = label;

  if (label === "Like") {
    btn.dataset.like = postId;
  } else {
    btn.dataset.dislike = postId;
  }

  wrap.appendChild(span);
  wrap.appendChild(btn);

  return wrap;
}

// --------------------------------------------------
// REACTIONS (LOGIC)
// --------------------------------------------------

function bindReactions(container) {
  container.addEventListener("click", async (e) => {
    const likeBtn = e.target.closest("[data-like]");
    const dislikeBtn = e.target.closest("[data-dislike]");

    if (!likeBtn && !dislikeBtn) return;

    e.stopPropagation();

    const postId = likeBtn
      ? likeBtn.dataset.like
      : dislikeBtn.dataset.dislike;

    const type = likeBtn ? "like" : "dislike";

    try {
      const res = await fetch(`/api/v1/posts/${postId}/${type}`, {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      if (!res.ok) {
        alert("You must be logged in to react");
        return;
      }

      const payload = await res.json();
      const data = payload.data;

      const postEl = container.querySelector(
        `[data-post-id="${postId}"]`
      );

      if (!postEl) return;

      postEl.querySelector(".like-count").textContent =
        data.likes_count;

      postEl.querySelector(".dislike-count").textContent =
        data.dislikes_count;

    } catch (err) {
      console.error("Reaction failed:", err);
    }
  });
}

