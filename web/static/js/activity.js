// web/static/js/activity.js
import {
  API_BASE,
  formatCreatedAt,
  resolveUsername,
  escapeHTML,
  toPositiveInt,
} from "./utils.js";
import { renderPostCard } from "./posts.js";
import { initReactions } from "./reactions.js";
import { createPagination } from "./pagination.js";
import { uiNotify } from "./ui-messages.js";

const DEFAULT_PAGE = 1;
const DEFAULT_PER_PAGE = 10;

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

function renderPostsSection({
  section,
  outputId,
  emptyId,
  countId,
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
    const article = renderPostCard(post, {
      clickable: true,
      showStatusToggle: false,
      showDelete: false,
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

        <time class="muted">${formatCreatedAt(comment.created_at)}</time>
      </header>

      <p class="activity-comment-author muted">
        By ${escapeHTML(resolveUsername(comment))}
      </p>

      ${body}
      ${imageMarkup}

      <footer class="activity-comment-stats">
        <span>Likes: ${Number(comment.likes) || 0}</span>
        <span>Dislikes: ${Number(comment.dislikes) || 0}</span>
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

async function renderActivity(state, pager) {
  const paginationEl = document.getElementById("activity-pagination");
  if (!paginationEl) return;

  paginationEl.hidden = true;

  const payload = await loadActivity(state);
  if (!payload) return;

  const data = payload?.data ?? {};

  renderPostsSection({
    section: data.created_posts,
    outputId: "created-posts-output",
    emptyId: "created-posts-empty",
    countId: "created-count",
  });

  renderPostsSection({
    section: data.liked_posts,
    outputId: "liked-posts-output",
    emptyId: "liked-posts-empty",
    countId: "liked-count",
  });

  renderPostsSection({
    section: data.disliked_posts,
    outputId: "disliked-posts-output",
    emptyId: "disliked-posts-empty",
    countId: "disliked-count",
  });

  renderCommentsSection(data.comments);

  initReactions();

  const totalPages = getGlobalTotalPages(data);
  pager.set(state.page, totalPages);
  paginationEl.hidden = totalPages <= 1;
}

async function startActivityPage() {
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
