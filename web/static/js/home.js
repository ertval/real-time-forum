// /web/static/js/home.js

import {
  renderPostCard,
  loadPostCommentsPreview,
} from "./posts.js";

import { initReactions } from "./reactions.js";

import { API_BASE } from "./utils.js";
import { initCategoryFilter } from "./category.js";
import { createPagination } from "./pagination.js";
import { uiNotify } from "./ui-messages.js";

/*-----------
  URL STATE
-----------*/

function getQueryState() {
  const params = new URLSearchParams(window.location.search);

  return {
    page: Number(params.get("page")) || 1,
    perPage: Number(params.get("per_page")) || 10,
    categoryId: params.get("category_id") || "",
  };
}

function setQueryState({ page, perPage, categoryId }) {
  const params = new URLSearchParams();

  if (page > 1) params.set("page", page);
  if (perPage !== 10) params.set("per_page", perPage);
  if (categoryId) params.set("category_id", categoryId);

  history.pushState(null, "", `?${params.toString()}`);
}

/*--------
  POSTS
--------*/

async function loadPosts({ page, perPage, categoryId }) {
  const url = new URL(`${API_BASE}/posts`, window.location.origin);

  url.searchParams.set("page", page);
  if (perPage > 0) {
    url.searchParams.set("per_page", perPage);
  }
  if (categoryId) url.searchParams.set("category_id", categoryId);

  const res = await fetch(url, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  if (!res.ok) return null;
  return res.json();
}

async function renderPosts(state, pager, paginationEl) {
  const output = document.getElementById("posts-output");
  const empty = document.getElementById("posts-empty");

  output.innerHTML = "";
  empty.hidden = true;

  paginationEl.hidden = true;
  pager.set(1, 1);

  const payload = await loadPosts(state);

  if (!payload || payload.data.length === 0) {
    empty.hidden = false;
    return;
  }

  for (const post of payload.data) {
    const card = renderPostCard(post);
    output.appendChild(card);
    await loadPostCommentsPreview(post.id, card);
  }

  if (state.perPage === 0) return;

  if (payload.meta && payload.meta.pagination) {
    const { page, total_pages } = payload.meta.pagination;

    if (total_pages > 1) {
      paginationEl.hidden = false;
      pager.set(page, total_pages);
    }
  }
}

/*------
  INIT
------*/

document.addEventListener("DOMContentLoaded", async () => {
  const loginSuccess = sessionStorage.getItem("auth:login-success");
  if (loginSuccess) {
    uiNotify("Welcome! You are now signed in.", { type: "success" });
    sessionStorage.removeItem("auth:login-success");
  }

  const paginationEl = document.getElementById("pagination");
  const prevBtn = document.getElementById("prevPage");
  const nextBtn = document.getElementById("nextPage");
  const numbersEl = document.getElementById("pageNumbers");
  const perPageSelect = document.getElementById("perPageSelect");

  let state = getQueryState();
  perPageSelect.value = state.perPage;

  const pager = createPagination({
    prevBtn,
    nextBtn,
    numbersEl,
    onPageChange: page => {
      state.page = page;
      setQueryState(state);
      renderPosts(state, pager, paginationEl);
    },
  });

  prevBtn.addEventListener("click", () => pager.prev());
  nextBtn.addEventListener("click", () => pager.next());

  perPageSelect.addEventListener("change", () => {
    state.perPage = Number(perPageSelect.value);
    state.page = 1;
    setQueryState(state);
    renderPosts(state, pager, paginationEl);
  });

  state.categoryId = await initCategoryFilter(categoryId => {
    state.categoryId = categoryId;
    state.page = 1;
    setQueryState(state);
    renderPosts(state, pager, paginationEl);
  });

  await renderPosts(state, pager, paginationEl);
  initReactions();

  window.addEventListener("popstate", () => {
    state = getQueryState();
    perPageSelect.value = state.perPage;
    renderPosts(state, pager, paginationEl);
  });
});
