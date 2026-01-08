// web/static/js/home.js

import {
  renderPostCard,
  loadPostCommentsPreview,
  initReactions
} from "./posts.js";

import { API_BASE } from "./utils.js";
import { initCategoryFilter } from "./category.js";

/* =========================
   POSTS
========================= */

async function loadPosts(categoryId = "") {
  const url = new URL(`${API_BASE}/posts`, window.location.origin);
  if (categoryId) url.searchParams.set("category_id", categoryId);

  const res = await fetch(url, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  if (!res.ok) return [];
  const payload = await res.json();
  return payload.data ?? [];
}

async function renderPosts(categoryId = "") {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");

  output.innerHTML = "";

  const posts = await loadPosts(categoryId);

  if (posts.length === 0) {
    empty.hidden = false;
    return;
  }

  empty.hidden = true;

  for (const post of posts) {
    const card = renderPostCard(post);
    output.appendChild(card);
    await loadPostCommentsPreview(post.id, card);
  }
}

/* =========================
   INIT
========================= */

document.addEventListener("DOMContentLoaded", async () => {
  const initialCategoryId = await initCategoryFilter(renderPosts);

  await renderPosts(initialCategoryId);

  initReactions();
});
