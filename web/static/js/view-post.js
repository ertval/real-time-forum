import {
  renderPostCard,
  loadPostCommentsPreview,
  initReactions,
} from "./posts.js";

import { API_BASE } from "./utils.js";

document.addEventListener("DOMContentLoaded", async () => {
  const postId = window.location.pathname.split("/").pop();
  if (!postId) return;

  const container = document.getElementById("post-output");
  if (!container) return;

  await loadAndRenderPost(postId, container);

  initReactions();
});

// ==================================================
// LOAD + RENDER SINGLE POST
// ==================================================

async function loadAndRenderPost(postId, container) {
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (!res.ok) {
      container.innerHTML = `<p class="muted">Post not found.</p>`;
      return;
    }

    const { data: post } = await res.json();

    const article = renderPostCard(post, { clickable: false });

    container.innerHTML = "";
    container.appendChild(article);

    // comments preview (same logic as home)
    await loadPostCommentsPreview(postId, article);

  } catch (err) {
    console.error("Failed to load post:", err);
    container.innerHTML = `<p class="muted">Failed to load post.</p>`;
  }
}
