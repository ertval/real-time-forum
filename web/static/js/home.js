import {
  renderPostCard,
  loadPostCommentsPreview,
  initReactions
} from "./posts.js";

import { API_BASE } from "./utils.js";

/* =========================
   CATEGORIES
========================= */

async function loadCategories() {
  const res = await fetch(`${API_BASE}/categories`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data) ? data : data.data ?? [];
}

async function populateCategoryFilter() {
  const select = document.getElementById("categoryFilter");
  if (!select) return;

  const categories = await loadCategories();

  // clean (keep "All")
  select.querySelectorAll("option:not(:first-child)").forEach(o => o.remove());

  for (const c of categories) {
    const opt = document.createElement("option");
    opt.value = c.id;
    opt.textContent = c.name;
    select.appendChild(opt);
  }
}

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
  await populateCategoryFilter();
  await renderPosts();

  const select = document.getElementById("categoryFilter");
  if (select) {
    select.addEventListener("change", () => {
      renderPosts(select.value);
    });
  }

  initReactions();
});
