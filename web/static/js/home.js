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
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    const payload = await res.json();
    if (!res.ok || !payload?.data?.length) {
      showEmpty(output, emptyEl);
      return;
    }

    emptyEl.hidden = true;
    output.innerHTML = "";

    for (const post of payload.data) {
      const card = renderPostCard(post);
      output.appendChild(card);

      await loadPostCommentsPreview(post.id, card);
    }
  } catch (err) {
    console.error(err);
    showEmpty(output, emptyEl);
  }
}

function showEmpty(output, emptyEl) {
  output.innerHTML = "";
  emptyEl.hidden = false;
}

// --------------------------------------------------
// POST CARD
// --------------------------------------------------

function renderPostCard(post) {
  const article = document.createElement("article");
  article.className = "post card card-pad";
  article.dataset.postId = post.id;

  article.innerHTML = `
    <header class="post-header clickable">
      <div>
        <h3 class="post-title">${post.title}</h3>
        <p class="muted">Author ID: ${post.author_id}</p>
      </div>
      <time class="muted">${formatCreatedAt(post.created_at)}</time>
    </header>

    <section class="post-body clickable">
      <p>${post.body}</p>
    </section>

    <section class="post-actions">
      ${reactionTemplate(post)}
    </section>

    <section class="post-comments" data-comments></section>
  `;

  // navigation ONLY from header & body
  article.querySelectorAll(".clickable").forEach(el => {
    el.addEventListener("click", () => {
      window.location.href = `/view-post/${post.id}`;
    });
  });

  // ABSOLUTE STOP on interactive areas
  article
    .querySelectorAll(
      ".post-actions, .post-comments, button, textarea, form"
    )
    .forEach(el => {
      el.addEventListener("click", e => e.stopPropagation());
    });

  return article;
}

// --------------------------------------------------
// COMMENTS PREVIEW (LAST 3)
// --------------------------------------------------

async function loadPostCommentsPreview(postId, article) {
  const container = article.querySelector("[data-comments]");
  if (!container) return;

  try {
    const res = await fetch(
      `${API_BASE}/posts/${postId}/comments`,
      { headers: { Accept: "application/json" } }
    );

    if (!res.ok) return;

    const payload = await res.json();
    const comments = payload?.data ?? [];

    if (!comments.length) {
      await maybeRenderCommentForm(container, postId);
      return;
    }

    const list = document.createElement("div");
    list.className = "comments comments-scroll"; // ⬅ scroll box

    comments.forEach(c => {
      const item = document.createElement("div");
      item.className = "comment";

      const username =
        c.username ||
        c.author ||
        (c.user_id ? `User ${c.user_id}` : "User");

      item.innerHTML = `
        <div class="comment-meta muted">
          <strong>${username}</strong> • ${formatCreatedAt(c.created_at)}
        </div>
        <div class="comment-body">
          ${c.body}
        </div>
      `;

      list.appendChild(item); 
    });

    container.appendChild(list);

    list.scrollTop = list.scrollHeight;

    await maybeRenderCommentForm(container, postId);

  } catch (err) {
    console.error(err);
  }
}



// --------------------------------------------------
// COMMENT FORM
// --------------------------------------------------

async function maybeRenderCommentForm(container, postId) {
  try {
    const res = await fetch(`${API_BASE}/users/me`, {
      credentials: "include",
    });

    if (!res.ok) return;

    const form = document.createElement("form");
    form.className = "comment-form";

    form.innerHTML = `
      <textarea
        placeholder="Write a comment..."
        required
      ></textarea>
      <button class="btn btn-primary" type="submit">
        Comment
      </button>
    `;

    // 🛑 STOP EVERYTHING
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

      const res = await fetch(
        `${API_BASE}/posts/${postId}/comments`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
          },
          body: JSON.stringify({ body }),
        }
      );

      if (res.ok) {
        location.reload();
      }
    });

    container.appendChild(form);
  } catch {}
}

// --------------------------------------------------
// REACTIONS (display only for now)
// --------------------------------------------------

function reactionTemplate(post) {
  return `
    <div class="reaction">
      <span data-like-count>${post.likes ?? 0}</span>
      <button
        class="btn btn-ghost"
        type="button"
        data-reaction="like"
        data-post-id="${post.id}"
      >
        Like
      </button>
    </div>

    <div class="reaction">
      <span data-dislike-count>${post.dislikes ?? 0}</span>
      <button
        class="btn btn-ghost"
        type="button"
        data-reaction="dislike"
        data-post-id="${post.id}"
      >
        Dislike
      </button>
    </div>
  `;
}

// --------------------------------------------------
// REACTIONS (LOGIC) – FIXED
// --------------------------------------------------

document.addEventListener(
  "click",
  async (e) => {
    const btn = e.target.closest("[data-reaction]");
    if (!btn) return;

    e.preventDefault();

    const postId = btn.dataset.postId;
    const type = btn.dataset.reaction;

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

      const article = document.querySelector(
        `article[data-post-id="${postId}"]`
      );

      article.querySelector("[data-like-count]").textContent =
        data.likes_count;
      article.querySelector("[data-dislike-count]").textContent =
        data.dislikes_count;

    } catch (err) {
      console.error(err);
    }
  },
  true 
);
