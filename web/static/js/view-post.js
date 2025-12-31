// web/static/js/view-post.js
const API_BASE = "/api/v1";

document.addEventListener("DOMContentLoaded", async () => {
  const pathParts = window.location.pathname.split("/");
  const postId = pathParts[pathParts.length - 1];

  if (!postId || isNaN(postId)) {
    console.error("Invalid post id");
    return;
  }

  try {
    const res = await fetch(`/api/v1/posts/${postId}`);
    if (!res.ok) throw new Error("Post not found");

    const post = await res.json();

    document.getElementById("post-title").textContent = post.title;
    document.getElementById("post-author").textContent =
      "Author: " + post.author;
    document.getElementById("post-body").textContent = post.body;

    const timeEl = document.getElementById("post-time");
    timeEl.textContent = new Date(post.created_at).toLocaleString();
    timeEl.dateTime = post.created_at;

    document.getElementById("like-count").textContent = post.likes;
    document.getElementById("dislike-count").textContent = post.dislikes;

  } catch (err) {
    console.error(err);
  }
});

async function loadPost(postId) {
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}`, {
      headers: { "Accept": "application/json" },
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
  document.querySelector(".viewpost-title").textContent = post.title;
  document.querySelector(".viewpost-body-text").textContent = post.body;

  document.querySelector(".viewpost-author").textContent =
    `Author: ${post.author ?? "unknown"}`;

  const timeEl = document.querySelector(".viewpost-time");
  timeEl.dateTime = post.created_at;
  timeEl.textContent = formatCreatedAt(post.created_at);

  document.getElementById("like-count").textContent = post.likes ?? 0;
  document.getElementById("dislike-count").textContent = post.dislikes ?? 0;
}

function formatCreatedAt(iso) {
  if (!iso) return "";
  return iso.slice(0, 10); 
}
