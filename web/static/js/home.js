import {
  renderPostCard,
  loadPostCommentsPreview,
  initReactions
} from "./posts.js";

import { API_BASE } from "./utils.js";

document.addEventListener("DOMContentLoaded", () => {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");

  if (output) {
    loadPublicPosts(output, empty);
  }

  initReactions();
});

async function loadPublicPosts(output, emptyEl) {
  try {
    const res = await fetch(`${API_BASE}/posts`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    const payload = await res.json();

    if (!res.ok || !Array.isArray(payload.data) || payload.data.length === 0) {
      emptyEl.hidden = false;
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
    emptyEl.hidden = false;
  }
}
