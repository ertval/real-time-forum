// web/static/js/my-posts.js
import { API_BASE, getPaginationFromURL } from "./utils.js";
import { renderPostCard, loadPostCommentsPreview } from "./posts.js";
import { initReactions } from "./reactions.js";
import { uiNotify, uiConfirm } from "./ui-messages.js";
import { playUpload, playDelete } from "./sound-effects.js";

let statusToggleBound = false;
let deleteBound = false;

function start() {
  initStatusFilterUI();
  initStatusToggle();
  initDeletePost();
  boot().catch(() => {
    uiNotify("Failed to load your posts.", { type: "danger" });
  });
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", start);
} else {
  start();
}

function getStatusFilterFromURL() {
  const params = new URLSearchParams(window.location.search);
  const v = (params.get("status") || "all").toLowerCase();
  return v === "draft" || v === "published" ? v : "all";
}

function setStatusFilterInURL(status) {
  const url = new URL(window.location.href);
  const params = url.searchParams;
  params.set("page", "1");

  if (!status || status === "all") {
    params.delete("status");
  } else {
    params.set("status", status);
  }

  history.replaceState({}, "", url.toString());
}

function initStatusFilterUI() {
  const select = document.getElementById("status-filter");
  if (!select) return;

  select.value = getStatusFilterFromURL();

  select.addEventListener("change", () => {
    setStatusFilterInURL(select.value);
    boot().catch(() => {
      uiNotify("Failed to load your posts.", { type: "danger" });
    });
  });
}

async function boot() {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");
  if (!output) return;

  output.innerHTML = "";
  if (empty) empty.hidden = true;

  const { page, perPage } = getPaginationFromURL();
  const status = getStatusFilterFromURL();

  const { posts } = await fetchMyPosts({ page, perPage, status });

  if (!posts.length) {
    if (empty) empty.hidden = false;
    return;
  }

  const fragment = document.createDocumentFragment();
  const articles = [];

  for (const post of posts) {
    const article = renderPostCard(post, {
      clickable: true,
      showStatusToggle: true,
      showDelete: true,
    });
    fragment.appendChild(article);
    articles.push(article);
  }

  output.appendChild(fragment);

  await runWithConcurrencyLimit(
    articles.map((article) => async () => {
      const postId = article.dataset.postId;
      if (postId) await loadPostCommentsPreview(postId, article);
    }),
    4
  );

  initReactions();
  initStatusToggle();
}

function initStatusToggle() {
  if (statusToggleBound) return;
  statusToggleBound = true;

  document.addEventListener(
    "click",
    async (e) => {
      const btn = e.target.closest(".post-status-toggle");
      if (!btn) return;

      e.preventDefault();
      e.stopPropagation();

      const postId = btn.dataset.postId;
      const currentStatus = btn.dataset.currentStatus;
      if (!postId || !currentStatus) return;

      const nextStatus = currentStatus === "draft" ? "published" : "draft";
      btn.disabled = true;

      try {
        const res = await fetch(`${API_BASE}/posts/${postId}`, {
          method: "PATCH",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
          },
          body: JSON.stringify({ status: nextStatus }),
        });

        if (!res.ok) {
          uiNotify("Failed to update post status.", { type: "danger" });
          return;
        }

        // SOUND IMMEDIATELY
        playUpload();

        //  UPDATE BUTTON TEXT + DATA
        btn.dataset.currentStatus = nextStatus;
        btn.textContent = nextStatus === "draft" ? "Publish" : "Draft";

        await boot();

      } catch {
        uiNotify("Failed to update post status.", { type: "danger" });
      } finally {
        btn.disabled = false;
      }
    },
    true
  );
}

function initDeletePost() {
  if (deleteBound) return;
  deleteBound = true;

  document.addEventListener(
    "click",
    async (e) => {
      const btn = e.target.closest(".post-delete");
      if (!btn) return;

      e.preventDefault();
      e.stopPropagation();

      const postId = btn.dataset.postId;
      if (!postId) return;

      const ok = await uiConfirm(
        "Delete this post? This action cannot be undone.",
        {
          type: "danger",
          title: "Delete post",
          okText: "Delete",
          cancelText: "Cancel",
        }
      );

      if (!ok) return;

      btn.disabled = true;

      try {
        const res = await fetch(`${API_BASE}/posts/${postId}`, {
          method: "DELETE",
          credentials: "include",
          headers: { Accept: "application/json" },
        });

        if (res.status === 401) {
          uiNotify("You must be logged in.", { type: "warn" });
          return;
        }

        if (res.status === 404) {
          uiNotify("Post not found.", { type: "warn" });
          await boot();
          return;
        }

        if (!res.ok) {
          uiNotify("Failed to delete post.", { type: "danger" });
          return;
        }

        // Play delete sound exactly once
        playDelete();

        uiNotify("Post deleted.", { type: "success" });
        await boot();

      } catch {
        uiNotify("Failed to delete post.", { type: "danger" });
      } finally {
        if (document.contains(btn)) btn.disabled = false;
      }
    },
    true
  );
}

/*------
  API
------*/

async function fetchMyPosts({ page, perPage, status }) {
  const url = new URL(`${API_BASE}/posts/mine`, window.location.origin);
  url.searchParams.set("page", String(page));
  url.searchParams.set("per_page", String(perPage));

  if (status && status !== "all") {
    url.searchParams.set("status", status);
  }

  const res = await fetch(url.toString(), {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  if (res.status === 401) {
    uiNotify("You must be logged in to view your posts.", { type: "warn" });
    return { posts: [], meta: null };
  }

  if (!res.ok) {
    uiNotify(`Failed to load posts (${res.status}).`, { type: "danger" });
    return { posts: [], meta: null };
  }

  const payload = await res.json();
  return {
    posts: Array.isArray(payload?.data) ? payload.data : [],
    meta: payload?.meta ?? null,
  };
}

/*--------------------
  CONCURRENCY HELPERS
--------------------*/

async function runWithConcurrencyLimit(tasks, limit = 4) {
  const queue = tasks.slice();

  const workers = Array.from({ length: limit }, async () => {
    while (queue.length) {
      const task = queue.shift();
      if (!task) return;
      try {
        await task();
      } catch {}
    }
  });

  await Promise.all(workers);
}
