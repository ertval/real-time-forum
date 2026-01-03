// web/static/js/view-post.js

document.addEventListener("DOMContentLoaded", () => {
  const parts = window.location.pathname.split("/");
  const postId = parts[parts.length - 1];

  if (!postId || isNaN(postId)) {
    console.error("Invalid post id");
    return;
  }

  loadPost(postId);
});

async function loadPost(postId) {
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}`, {
      headers: { Accept: "application/json" },
      credentials: "include",
    });

    if (!res.ok) {
      throw new Error("Post not found");
    }

    const payload = await res.json();
    const post = payload.data;

    renderPost(post);
  } catch (err) {
    console.error(err);
    alert("Failed to load post");
  }
}

function renderPost(post) {
  // title
  document.querySelector(".viewpost-title").textContent =
    post.title ?? "";

  // body
  document.querySelector(".viewpost-body-text").textContent =
    post.body ?? "";

  // author (fallback)
  document.querySelector(".viewpost-author").textContent =
    post.author
      ? `Author: ${post.author}`
      : `Author ID: ${post.author_id}`;

  // time
  const timeEl = document.querySelector(".viewpost-time");
  timeEl.dateTime = post.created_at;
  timeEl.textContent = formatCreatedAt(post.created_at);

  // reactions
  document.getElementById("like-count").textContent =
    post.likes ?? 0;

  document.getElementById("dislike-count").textContent =
    post.dislikes ?? 0;
}

function formatCreatedAt(iso) {
  if (!iso) return "";
  return iso.slice(0, 10); // YYYY-MM-DD
}
