// web/static/js/comments.js

/**
 * Bind comments logic inside a container
 * Used by: home.js, view-post.js
 */
function bindComments(container) {
  container.addEventListener("click", async (e) => {
    const btn = e.target.closest("[data-comments]");
    if (!btn) return;

    e.stopPropagation();

    const postId = btn.dataset.comments;
    const postEl = btn.closest("[data-post-id]");
    if (!postEl) return;

    let commentsEl = postEl.querySelector(".comments");

    // Toggle
    if (commentsEl) {
      commentsEl.hidden = !commentsEl.hidden;
      return;
    }

    commentsEl = createCommentsBlock(postId);
    postEl.appendChild(commentsEl);

    await loadComments(postId, commentsEl);
  });
}

/* -------------------------------------------------- */
/* LOAD COMMENTS                                     */
/* -------------------------------------------------- */

async function loadComments(postId, wrapper) {
  const list = wrapper.querySelector(".comments-list");
  list.innerHTML = "Loading…";

  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    const payload = await res.json().catch(() => null);
    const comments = Array.isArray(payload?.data) ? payload.data : [];

    list.innerHTML = "";

    if (comments.length === 0) {
      list.innerHTML = `<p class="muted">No comments yet.</p>`;
      return;
    }

    for (const c of comments) {
      list.appendChild(renderComment(c));
    }
  } catch (err) {
    console.error("Failed to load comments", err);
    list.innerHTML = `<p class="muted">Failed to load comments.</p>`;
  }
}

/* -------------------------------------------------- */
/* CREATE COMMENT                                    */
/* -------------------------------------------------- */

async function createComment(postId, textarea, list) {
  const body = textarea.value.trim();
  if (!body) return;

  textarea.disabled = true;

  try {
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
    const comment = payload?.data;

    if (comment) {
      list.prepend(renderComment(comment));
      textarea.value = "";
    }
  } catch (err) {
    console.error("Create comment failed", err);
  } finally {
    textarea.disabled = false;
  }
}

/* -------------------------------------------------- */
/* UI HELPERS                                        */
/* -------------------------------------------------- */

function createCommentsBlock(postId) {
  const wrapper = document.createElement("div");
  wrapper.className = "comments";

  wrapper.innerHTML = `
    <h4 class="comments-title">Comments</h4>
    <div class="comments-list"></div>
    <form class="comment-form">
      <textarea class="comment-input" rows="2" required></textarea>
      <button class="btn btn-primary" type="submit">Post comment</button>
    </form>
  `;

  const form = wrapper.querySelector(".comment-form");
  const textarea = wrapper.querySelector(".comment-input");
  const list = wrapper.querySelector(".comments-list");

  form.addEventListener("submit", (e) => {
    e.preventDefault();
    createComment(postId, textarea, list);
  });

  return wrapper;
}

function renderComment(comment) {
  const el = document.createElement("div");
  el.className = "comment";

  el.innerHTML = `
    <div class="comment-header">
      <span class="muted">
        ${comment.author ?? "User"} · ${formatCreatedAt(comment.created_at)}
      </span>
    </div>
    <p class="comment-text">${comment.body}</p>
  `;

  return el;
}
