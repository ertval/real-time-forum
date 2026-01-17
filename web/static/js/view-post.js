// /web/static/js/view-post.js

import {
  renderPostCard,
  loadPostCommentsPreview,
  initReactions,
} from "./posts.js";

import { API_BASE } from "./utils.js";

/* ==================================================
   HELPERS
================================================== */

function getPostIdFromURL() {
  const parts = window.location.pathname.split("/");
  const id = Number(parts[parts.length - 1]);
  return Number.isFinite(id) && id > 0 ? id : null;
}

/* ==================================================
   INIT
================================================== */

document.addEventListener("DOMContentLoaded", async () => {
  const postId = getPostIdFromURL();
  if (!postId) return;

  const container = document.getElementById("post-output");
  if (!container) return;

  const result = await loadAndRenderPost(postId, container);
  if (!result) return;

  const { article, categoryId } = result;

  // reactions AFTER DOM is ready
  initReactions();

  // navigation only if category exists
  if (categoryId) {
    initPostNavigation(postId, categoryId);
  }
});

/* ==================================================
   LOAD + RENDER SINGLE POST
   Returns { article, categoryId } | null
================================================== */

async function loadAndRenderPost(postId, container) {
  try {
    const res = await fetch(`${API_BASE}/posts/${postId}`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (!res.ok) {
      container.innerHTML = `<p class="muted">Post not found.</p>`;
      return null;
    }

    const { data: post } = await res.json();

    const article = renderPostCard(post, { clickable: false });

    container.innerHTML = "";
    container.appendChild(article);

    // comments preview AFTER article exists
    await loadPostCommentsPreview(postId, article);

    const categoryId =
      Array.isArray(post.categories) && post.categories.length > 0
        ? post.categories[0].id
        : null;

    return { article, categoryId };
  } catch (err) {
    console.error("Failed to load post:", err);
    container.innerHTML = `<p class="muted">Failed to load post.</p>`;
    return null;
  }
}

/* ==================================================
   POST NAVIGATION (CATEGORY-AWARE)
================================================== */

async function initPostNavigation(postId, categoryId) {
  const prevBtn = document.getElementById("post-prev");
  const nextBtn = document.getElementById("post-next");

  if (!prevBtn || !nextBtn) return;

  // reset
  prevBtn.hidden = true;
  nextBtn.hidden = true;
  prevBtn.onclick = null;
  nextBtn.onclick = null;

  try {
    const res = await fetch(
      `${API_BASE}/posts/${postId}/nav?category_id=${categoryId}`,
      { credentials: "include" }
    );

    if (!res.ok) return;

    const { prev_id, next_id } = (await res.json()).data || {};

    if (typeof prev_id === "number" && prev_id > 0) {
      prevBtn.hidden = false;
      prevBtn.onclick = () => {
        window.location.href = `/view-post/${prev_id}`;
      };
    }

    if (typeof next_id === "number" && next_id > 0) {
      nextBtn.hidden = false;
      nextBtn.onclick = () => {
        window.location.href = `/view-post/${next_id}`;
      };
    }

    /* =========================
       KEYBOARD NAVIGATION
    ========================= */

    document.addEventListener("keydown", (e) => {
      const tag = e.target.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA") return;

      if (e.key === "ArrowLeft" && !prevBtn.hidden) {
        prevBtn.click();
      }

      if (e.key === "ArrowRight" && !nextBtn.hidden) {
        nextBtn.click();
      }
    });

  } catch (err) {
    console.error("Failed to load post navigation:", err);
  }
}
