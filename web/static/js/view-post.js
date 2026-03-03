// /web/static/js/view-post.js

import {
  renderPostCard,
  loadPostCommentsPreview,
} from "./posts.js";

import { initReactions } from "./reactions.js";
import { API_BASE } from "./utils.js";
import { uiNotify } from "./ui-messages.js";

/*---------
  HELPERS
---------*/

function getPostIdFromURL() {
  const parts = window.location.pathname.split("/");
  const id = Number(parts[parts.length - 1]);
  return Number.isFinite(id) && id > 0 ? id : null;
}

/*----------------------------------
  RELIABLE HIGHLIGHT WITH RETRIES
----------------------------------*/

async function highlightComment(commentId) {
  const numericId = Number(commentId);
  if (!numericId) return;

  for (let i = 0; i < 20; i++) {
    const el = document.getElementById(`comment-${numericId}`);
    if (el) {
      el.scrollIntoView({ behavior: "smooth", block: "center" });
      el.classList.add("highlight-comment");

      // Auto-hide highlight after 2.5 seconds
      setTimeout(() => {
        el.classList.add("fade-out");
      }, 2500);

      return;
    }
    await new Promise(res => setTimeout(res, 50));
  }
}

/*------
  INIT
------*/

document.addEventListener("DOMContentLoaded", async () => {
  const container = document.getElementById("post-output");
  if (!container) return;

  const postId = getPostIdFromURL();
  if (!postId) {
    container.innerHTML = `<p class="muted">Invalid post ID.</p>`;
    uiNotify("Invalid post ID.", { type: "danger" });
    return;
  }

  const result = await loadAndRenderPost(postId, container);
  if (!result) return;

  const { categoryId } = result;

  initReactions();

  if (categoryId) {
    initPostNavigation(postId, categoryId);
  }
});

/*-------------------------------
  LOAD + RENDER SINGLE POST
-------------------------------*/

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

    // Load comments into DOM
    await loadPostCommentsPreview(postId, article);

    /* ----------------------------------------------------
      HIGHLIGHT COMMENT (WITH RETRIES + "last" support)
    ---------------------------------------------------- */
    const params = new URLSearchParams(window.location.search);
    const highlight = params.get("highlight");

        if (highlight) {
          if (highlight === "last") {
            // Highlight newest comment
            for (let i = 0; i < 20; i++) {
              const all = document.querySelectorAll(".comment");
              if (all.length > 0) {
                const last = all[all.length - 1];
                last.scrollIntoView({ behavior: "smooth", block: "center" });
                last.classList.add("highlight-comment");

                setTimeout(() => last.classList.add("fade-out"), 1000);
                setTimeout(() => last.classList.remove("highlight-comment", "fade-out"), 2200);
                break;
              }
              await new Promise(r => setTimeout(r, 50));
            }
          } else {
            // Normal highlight by ID
            highlightComment(highlight);
          }
        }

        const categoryId =
          Array.isArray(post.categories) && post.categories.length > 0
            ? post.categories[0].id
            : null;

        return { categoryId };

      } catch (err) {
        console.error("Failed to load post:", err);
        container.innerHTML = `<p class="muted">Failed to load post.</p>`;
        return null;
      }
    }

/*-------------------------------------
  POST NAVIGATION
-------------------------------------*/

async function initPostNavigation(postId, categoryId) {
  const prevBtn = document.getElementById("post-prev");
  const nextBtn = document.getElementById("post-next");

  if (!prevBtn || !nextBtn) return;

  prevBtn.hidden = true;
  nextBtn.hidden = true;

  try {
    const res = await fetch(
      `${API_BASE}/posts/${postId}/nav?category_id=${categoryId}`,
      { credentials: "include" }
    );

    if (!res.ok) return;

    const { prev_id, next_id } = (await res.json()).data || {};

    if (prev_id > 0) {
      prevBtn.hidden = false;
      prevBtn.onclick = () => (window.location.href = `/view-post/${prev_id}`);
    }

    if (next_id > 0) {
      nextBtn.hidden = false;
      nextBtn.onclick = () => (window.location.href = `/view-post/${next_id}`);
    }

    document.addEventListener("keydown", (e) => {
      if (["INPUT", "TEXTAREA"].includes(e.target.tagName)) return;

      if (e.key === "ArrowLeft" && !prevBtn.hidden) prevBtn.click();
      if (e.key === "ArrowRight" && !nextBtn.hidden) nextBtn.click();
    });

  } catch (err) {
    console.error("Failed to load post navigation:", err);
  }
}
