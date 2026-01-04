// --------------------------------------------------
// INIT
// --------------------------------------------------

document.addEventListener("DOMContentLoaded", () => {
  const postId = window.location.pathname.split("/").pop();
  if (!postId) return;

  loadPost(postId);
  loadComments(postId);
  setupCommentForm(postId);
});

// --------------------------------------------------
// LOAD POST
// --------------------------------------------------

async function loadPost(postId) {
  const res = await fetch(`${API_BASE}/posts/${postId}`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) return;

  const { data: post } = await res.json();

  document.querySelector(".viewpost-title").textContent = post.title;
  document.querySelector(".viewpost-body-text").textContent = post.body;

  document.querySelector(".viewpost-author").textContent =
    post.author
      ? `Author: ${post.author}`
      : `Author ID: ${post.author_id}`;

  document.querySelector(".viewpost-time").textContent =
    formatCreatedAt(post.created_at);
}

// --------------------------------------------------
// LOAD COMMENTS (ALL – ASC)
// --------------------------------------------------

async function loadComments(postId) {
  const list = document.getElementById("comments-list");
  if (!list) return;

  const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) return;

  const payload = await res.json();
  const comments = payload.data ?? [];

  list.innerHTML = "";

  comments.forEach(c => {
    list.appendChild(renderComment(c));
  });

  requestAnimationFrame(() => {
    list.scrollTop = list.scrollHeight;
  });
}

// --------------------------------------------------
// COMMENT UI
// --------------------------------------------------

function renderComment(c) {
  const div = document.createElement("div");
  div.className = "comment";

  const username =
    c.username ||
    c.author ||
    (c.user_id ? `User ${c.user_id}` : "User");

  div.innerHTML = `
    <div>
      <strong>${username}:</strong>
      ${c.body}
    </div>
    <div class="muted" style="font-size:12px;">
      ${formatCreatedAt(c.created_at)}
    </div>
  `;

  div.addEventListener("click", e => e.stopPropagation());

  return div;
}

// --------------------------------------------------
// COMMENT FORM
// --------------------------------------------------

async function setupCommentForm(postId) {
  const form = document.getElementById("comment-form");
  if (!form) return;

  const res = await fetch(`${API_BASE}/users/me`, {
    credentials: "include",
  });

  if (!res.ok) {
    form.hidden = true;
    return;
  }

  form.hidden = false;

  form.addEventListener("submit", async e => {
    e.preventDefault();

    const textarea = document.getElementById("comment-body");
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

    if (res.ok) {
      textarea.value = "";
      loadComments(postId); // refresh + auto scroll κάτω
    }
  });
}

// --------------------------------------------------
// REACTIONS (LIKE / DISLIKE)
// --------------------------------------------------

document.addEventListener("click", async (e) => {
  const btn = e.target.closest("[data-reaction]");
  if (!btn) return;

  e.preventDefault();
  e.stopPropagation();

  const type = btn.dataset.reaction;
  const postId = window.location.pathname.split("/").pop();

  if (!type || !postId) return;

  try {
    const res = await fetch(
      `${API_BASE}/posts/${postId}/${type}`,
      {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      }
    );

    if (!res.ok) {
      alert("You must be logged in to react");
      return;
    }

    const { data } = await res.json();

    document.getElementById("like-count").textContent =
      data.likes_count;

    document.getElementById("dislike-count").textContent =
      data.dislikes_count;

  } catch (err) {
    console.error("Reaction failed:", err);
  }
});
