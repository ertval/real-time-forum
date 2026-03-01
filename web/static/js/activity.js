// web/static/js/activity.js
import {
  API_BASE,
  formatCreatedAt,
  resolveUsername,
  escapeHTML,
  toPositiveInt,
} from "./utils.js";
import { renderPostCard, reactionTemplate } from "./posts.js";
import { initReactions } from "./reactions.js";
import { createPagination } from "./pagination.js";
import { uiNotify, uiConfirm } from "./ui-messages.js";
import { playDelete, playUpload } from "./sound-effects.js";
import { Auth } from "./auth.js";

const DEFAULT_PAGE = 1;
const DEFAULT_PER_PAGE = 10;
let editBound = false;
let deleteBound = false;
let commentDeleteBound = false;
let statusToggleBound = false;

function getQueryState() {
  const params = new URLSearchParams(window.location.search);
  const rawStatus = (params.get("status") || "all").toLowerCase();

  return {
    page: toPositiveInt(params.get("page")) || DEFAULT_PAGE,
    perPage: toPositiveInt(params.get("per_page")) || DEFAULT_PER_PAGE,
    status:
      rawStatus === "draft" || rawStatus === "published"
        ? rawStatus
        : "all",
  };
}

function setQueryState({ page, perPage, status }) {
  const params = new URLSearchParams();

  if (page > DEFAULT_PAGE) params.set("page", String(page));
  if (perPage !== DEFAULT_PER_PAGE) params.set("per_page", String(perPage));
  if (status && status !== "all") params.set("status", status);

  const qs = params.toString();
  history.pushState(null, "", qs ? `?${qs}` : window.location.pathname);
}

async function loadActivity({ page, perPage, status }) {
  const url = new URL(`${API_BASE}/users/activity`, window.location.origin);
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
    window.location.href = "/login";
    return null;
  }

  if (!res.ok) {
    throw new Error(`Failed to load activity (${res.status})`);
  }

  return res.json();
}

function asItems(section) {
  return Array.isArray(section?.items) ? section.items : [];
}

function asPagination(section) {
  return section?.pagination ?? {};
}

function asPositiveID(value) {
  const id = Number(value);
  return Number.isInteger(id) && id > 0 ? id : 0;
}

function resolveActivityViewerID(data, fallbackID = 0) {
  const fromAuth = asPositiveID(Auth.user?.id);
  if (fromAuth) return fromAuth;

  const fromFallback = asPositiveID(fallbackID);
  if (fromFallback) return fromFallback;

  const createdItems = asItems(data?.created_posts);
  if (createdItems.length) {
    const fromCreated = asPositiveID(createdItems[0]?.author_id);
    if (fromCreated) return fromCreated;
  }

  const commentItems = asItems(data?.comments);
  if (commentItems.length) {
    const fromComments = asPositiveID(commentItems[0]?.user_id);
    if (fromComments) return fromComments;
  }

  return 0;
}

function renderPostsSection({
  section,
  outputId,
  emptyId,
  countId,
  currentUserID,
}) {
  const output = document.getElementById(outputId);
  const empty = document.getElementById(emptyId);
  const count = document.getElementById(countId);
  if (!output || !empty || !count) return;

  output.innerHTML = "";
  empty.hidden = true;

  const items = asItems(section);
  const pagination = asPagination(section);
  count.textContent = String(pagination.total ?? items.length);

  if (!items.length) {
    empty.hidden = false;
    return;
  }

  const fragment = document.createDocumentFragment();

  for (const post of items) {
    const isOwner = Number(post.author_id) === Number(currentUserID);

    const article = renderPostCard(post, {
      clickable: true,
      showStatusToggle: isOwner,
      showDelete: isOwner,
      showEdit: isOwner,
    });

    // Keep activity page lightweight: no comments preview loading for each card.
    article.querySelector(".post-comments")?.remove();
    fragment.appendChild(article);
  }

  output.appendChild(fragment);
}

function renderCommentsSection(section) {
  const output = document.getElementById("comments-output");
  const empty = document.getElementById("comments-empty");
  const count = document.getElementById("comments-count");
  if (!output || !empty || !count) return;

  output.innerHTML = "";
  empty.hidden = true;

  const items = asItems(section);
  const pagination = asPagination(section);
  count.textContent = String(pagination.total ?? items.length);

  if (!items.length) {
    empty.hidden = false;
    return;
  }

  const fragment = document.createDocumentFragment();

  for (const comment of items) {
    const commentID = Number(comment.id) || 0;
    const deleteButtonMarkup =
      commentID > 0
        ? `
          <button
            class="btn btn-danger btn-sm comment-delete"
            data-comment-id="${commentID}"
          >
            Delete
          </button>
        `
        : "";

    const article = document.createElement("article");
    article.className = "activity-comment card card-pad";

    const post = comment.post || {};
    const body =
      typeof comment.body === "string" && comment.body.trim()
        ? `<p class="activity-comment-body">${escapeHTML(comment.body)}</p>`
        : `<p class="activity-comment-body muted">No text body</p>`;

    const imageMarkup =
      typeof comment.image_url === "string" && comment.image_url.trim()
        ? `
          <div class="activity-comment-image-wrap">
            <img
              class="activity-comment-image"
              src="${escapeHTML(comment.image_url)}"
              alt="Comment image by ${escapeHTML(resolveUsername(comment))}"
              loading="lazy"
            />
          </div>
        `
        : "";

    article.innerHTML = `
      <header class="activity-comment-head">
        <div class="activity-comment-post-wrap">
          <p class="muted">On post</p>
          <a
            class="activity-comment-post-link"
            href="/view-post/${Number(post.id) || Number(comment.post_id) || 0}"
          >
            ${escapeHTML(post.title || "Untitled post")}
          </a>
        </div>

        <div class="activity-comment-head-right">
          <time class="muted">${formatCreatedAt(comment.created_at)}</time>
          ${deleteButtonMarkup}
        </div>
      </header>

      <p class="activity-comment-author muted">
        By ${escapeHTML(resolveUsername(comment))}
      </p>

      ${body}
      ${imageMarkup}

      <footer>
        <div
          class="activity-comment-reactions comment"
          data-comment-id="${commentID}"
        >
          ${reactionTemplate(comment, true)}
        </div>
      </footer>
    `;

    fragment.appendChild(article);
  }

  output.appendChild(fragment);
}

function getGlobalTotalPages(data) {
  const totalPages = [
    asPagination(data?.created_posts).total_pages,
    asPagination(data?.liked_posts).total_pages,
    asPagination(data?.disliked_posts).total_pages,
    asPagination(data?.comments).total_pages,
  ]
    .map(v => Number(v) || 0)
    .filter(v => v > 0);

  return totalPages.length ? Math.max(...totalPages) : 1;
}

function initSectionToggles() {
  const toggles = document.querySelectorAll("[data-activity-toggle]");

  for (const toggle of toggles) {
    const controls = toggle.getAttribute("aria-controls");
    if (!controls) continue;

    const content = document.getElementById(controls);
    if (!content) continue;

    const isExpanded = toggle.getAttribute("aria-expanded") === "true";
    content.hidden = !isExpanded;

    toggle.addEventListener("click", () => {
      const expanded = toggle.getAttribute("aria-expanded") === "true";
      const nextExpanded = !expanded;
      toggle.setAttribute("aria-expanded", String(nextExpanded));
      content.hidden = !nextExpanded;
    });
  }
}

async function renderActivity(state, pager) {
  const paginationEl = document.getElementById("activity-pagination");
  if (!paginationEl) return;

  paginationEl.hidden = true;

  const payload = await loadActivity(state);
  if (!payload) return;

  const data = payload?.data ?? {};
  const currentUserID = resolveActivityViewerID(data, activityUserID);
  if (currentUserID > 0) {
    activityUserID = currentUserID;
  }

  renderPostsSection({
    section: data.created_posts,
    outputId: "created-posts-output",
    emptyId: "created-posts-empty",
    countId: "created-count",
    currentUserID,
  });

  renderCommentsSection(data.comments);

  renderPostsSection({
    section: data.liked_posts,
    outputId: "liked-posts-output",
    emptyId: "liked-posts-empty",
    countId: "liked-count",
    currentUserID,
  });

  renderPostsSection({
    section: data.disliked_posts,
    outputId: "disliked-posts-output",
    emptyId: "disliked-posts-empty",
    countId: "disliked-count",
    currentUserID,
  });

  initReactions();

  const totalPages = getGlobalTotalPages(data);
  pager.set(state.page, totalPages);
  paginationEl.hidden = totalPages <= 1;
}

let activityUserID = 0;

function initEditPostNavigation() {
  if (editBound) return;
  editBound = true;

  document.addEventListener(
    "click",
    (e) => {
      const btn = e.target.closest(".post-edit");
      if (!btn) return;

      e.preventDefault();
      e.stopPropagation();

      const postId = btn.dataset.postId;
      if (!postId) return;

      const next = encodeURIComponent(
        `${window.location.pathname}${window.location.search}`
      );
      window.location.href = `/edit-post/${postId}?next=${next}`;
    },
    true
  );
}

function initDeletePost(refresh) {
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

      const ok = await uiConfirm("Delete this post? This action cannot be undone.", {
        type: "danger",
        title: "Delete post",
        okText: "Delete",
        cancelText: "Cancel",
      });
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
        if (res.status === 403) {
          uiNotify("You can only delete your own posts.", { type: "warn" });
          return;
        }
        if (res.status === 404) {
          uiNotify("Post not found.", { type: "warn" });
          await refresh();
          return;
        }
        if (!res.ok) {
          uiNotify("Failed to delete post.", { type: "danger" });
          return;
        }

        playDelete();
        uiNotify("Post deleted.", { type: "success" });
        await refresh();
      } catch (err) {
        console.error("Activity delete failed:", err);
        uiNotify("Failed to delete post.", { type: "danger" });
      } finally {
        if (document.contains(btn)) btn.disabled = false;
      }
    },
    true
  );
}

function initDeleteComment(refresh) {
  if (commentDeleteBound) return;
  commentDeleteBound = true;

  document.addEventListener(
    "click",
    async (e) => {
      const btn = e.target.closest(".comment-delete");
      if (!btn) return;

      e.preventDefault();
      e.stopPropagation();

      const commentId = btn.dataset.commentId;
      if (!commentId) return;

      const ok = await uiConfirm(
        "Delete this comment? This action cannot be undone.",
        {
          type: "danger",
          title: "Delete comment",
          okText: "Delete",
          cancelText: "Cancel",
        }
      );
      if (!ok) return;

      btn.disabled = true;

      try {
        const res = await fetch(`${API_BASE}/comments/${commentId}`, {
          method: "DELETE",
          credentials: "include",
          headers: { Accept: "application/json" },
        });

        if (res.status === 401) {
          uiNotify("You must be logged in.", { type: "warn" });
          return;
        }
        if (res.status === 403) {
          uiNotify("You can only delete your own comments.", { type: "warn" });
          return;
        }
        if (res.status === 404) {
          uiNotify("Comment not found.", { type: "warn" });
          await refresh();
          return;
        }
        if (!res.ok) {
          uiNotify("Failed to delete comment.", { type: "danger" });
          return;
        }

        playDelete();
        uiNotify("Comment deleted.", { type: "success" });
        await refresh();
      } catch (err) {
        console.error("Activity comment delete failed:", err);
        uiNotify("Failed to delete comment.", { type: "danger" });
      } finally {
        if (document.contains(btn)) btn.disabled = false;
      }
    },
    true
  );
}

function initStatusToggle(refresh) {
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

        if (res.status === 401) {
          uiNotify("You must be logged in.", { type: "warn" });
          return;
        }

        if (res.status === 403) {
          uiNotify("You can only update your own posts.", { type: "warn" });
          return;
        }

        if (res.status === 404) {
          uiNotify("Post not found.", { type: "warn" });
          await refresh();
          return;
        }

        if (!res.ok) {
          uiNotify("Failed to update post status.", { type: "danger" });
          return;
        }

        playUpload();
        btn.dataset.currentStatus = nextStatus;
        btn.textContent = nextStatus === "draft" ? "Publish" : "Draft";
        await refresh();
      } catch (err) {
        console.error("Activity status toggle failed:", err);
        uiNotify("Failed to update post status.", { type: "danger" });
      } finally {
        if (document.contains(btn)) btn.disabled = false;
      }
    },
    true
  );
}

async function startActivityPage() {
  initSectionToggles();

  await Auth.init();
  activityUserID = asPositiveID(Auth.user?.id);

  const prevBtn = document.getElementById("activity-prev-page");
  const nextBtn = document.getElementById("activity-next-page");
  const numbersEl = document.getElementById("activity-page-numbers");
  const statusSelect = document.getElementById("activity-status-filter");
  const perPageSelect = document.getElementById("activity-per-page");

  if (
    !(prevBtn instanceof HTMLButtonElement) ||
    !(nextBtn instanceof HTMLButtonElement) ||
    !(numbersEl instanceof HTMLElement) ||
    !(statusSelect instanceof HTMLSelectElement) ||
    !(perPageSelect instanceof HTMLSelectElement)
  ) {
    return;
  }

  let state = getQueryState();

  const refresh = async () => {
    await renderActivity(state, pager);
  };

  statusSelect.value = state.status;
  perPageSelect.value = String(state.perPage);

  const pager = createPagination({
    prevBtn,
    nextBtn,
    numbersEl,
    onPageChange: async (page) => {
      state.page = page;
      setQueryState(state);
      try {
        await renderActivity(state, pager);
      } catch (err) {
        console.error("Activity pagination failed:", err);
        uiNotify("Failed to load activity.", { type: "danger" });
      }
    },
  });

  prevBtn.addEventListener("click", () => pager.prev());
  nextBtn.addEventListener("click", () => pager.next());

  statusSelect.addEventListener("change", async () => {
    state.status = statusSelect.value;
    state.page = DEFAULT_PAGE;
    setQueryState(state);
    try {
      await renderActivity(state, pager);
    } catch (err) {
      console.error("Activity status filter failed:", err);
      uiNotify("Failed to load activity.", { type: "danger" });
    }
  });

  perPageSelect.addEventListener("change", async () => {
    state.perPage = Number(perPageSelect.value) || DEFAULT_PER_PAGE;
    state.page = DEFAULT_PAGE;
    setQueryState(state);
    try {
      await renderActivity(state, pager);
    } catch (err) {
      console.error("Activity per-page update failed:", err);
      uiNotify("Failed to load activity.", { type: "danger" });
    }
  });

  try {
    await renderActivity(state, pager);
  } catch (err) {
    console.error("Activity render failed:", err);
    uiNotify("Failed to load activity.", { type: "danger" });
  }

  window.addEventListener("popstate", async () => {
    state = getQueryState();
    statusSelect.value = state.status;
    perPageSelect.value = String(state.perPage);

    try {
      await renderActivity(state, pager);
    } catch (err) {
      console.error("Activity history navigation failed:", err);
      uiNotify("Failed to load activity.", { type: "danger" });
    }
  });

  initEditPostNavigation();
  initStatusToggle(refresh);
  initDeletePost(refresh);
  initDeleteComment(refresh);
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => {
    startActivityPage().catch((err) => {
      console.error("Activity startup failed:", err);
      uiNotify("Failed to load activity.", { type: "danger" });
    });
  });
} else {
  startActivityPage().catch((err) => {
    console.error("Activity startup failed:", err);
    uiNotify("Failed to load activity.", { type: "danger" });
  });
}
