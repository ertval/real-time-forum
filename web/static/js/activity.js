// web/static/js/activity.js
import {
  API_BASE,
  buildImageRequestOptions,
  formatCreatedAt,
  resolveUsername,
  escapeHTML,
  IMAGE_ACCEPT_ATTR,
  toPositiveInt,
} from "./utils.js";
import { renderPostCard, reactionTemplate } from "./posts.js";
import { initReactions } from "./reactions.js";
import { createPagination } from "./pagination.js";
import { uiNotify, uiConfirm } from "./ui-messages.js";
import { playDelete, playUpload } from "./sound-effects.js";
import { Auth } from "./auth.js";
import { setupImagePicker } from "./image-picker.js";
import {
  editCommentButton,
  deleteCommentButton,
} from "./post-actions.js";

const DEFAULT_PAGE = 1;
const DEFAULT_PER_PAGE = 10;
let editBound = false;
let deleteBound = false;
let commentDeleteBound = false;
let commentEditBound = false;
let statusToggleBound = false;
let activeCommentEditor = null;

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

function activityCommentContentTemplate(commentBody, commentImageURL, username) {
  const body =
    typeof commentBody === "string" && commentBody.trim()
      ? `<p class="activity-comment-body">${escapeHTML(commentBody)}</p>`
      : `<p class="activity-comment-body muted">No text body</p>`;

  const imageMarkup =
    typeof commentImageURL === "string" && commentImageURL.trim()
      ? `
          <div class="activity-comment-image-wrap">
            <img
              class="activity-comment-image"
              src="${escapeHTML(commentImageURL)}"
              alt="Comment image by ${escapeHTML(username)}"
              loading="lazy"
            />
          </div>
        `
      : "";

  return `${body}${imageMarkup}`;
}

function closeActiveCommentEditor({ playCancelSound = false } = {}) {
  if (!activeCommentEditor || typeof activeCommentEditor.close !== "function") {
    return;
  }
  activeCommentEditor.close({ playCancelSound });
  activeCommentEditor = null;
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
    const commentBody =
      typeof comment.body === "string" ? comment.body : "";
    const commentImageURL =
      typeof comment.image_url === "string" && comment.image_url.trim()
        ? comment.image_url.trim()
        : "";
    const username = resolveUsername(comment);

    const editButtonMarkup =
      commentID > 0 ? editCommentButton(commentID) : "";

    const deleteButtonMarkup =
      commentID > 0 ? deleteCommentButton(commentID) : "";

    const article = document.createElement("article");
    article.className = "activity-comment card card-pad";
    article.dataset.commentId = String(commentID);
    article.dataset.commentBody = commentBody;
    if (commentImageURL) {
      article.dataset.commentImageUrl = commentImageURL;
    }

    const post = comment.post || {};

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
          ${editButtonMarkup}
          ${deleteButtonMarkup}
        </div>
      </header>

      <p class="activity-comment-author muted">
        By ${escapeHTML(username)}
      </p>

      <section class="activity-comment-content">
        ${activityCommentContentTemplate(commentBody, commentImageURL, username)}
      </section>

      <footer class="activity-comment-footer">
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

let sectionToggleBound = false;

function initSectionToggles() {

   if (sectionToggleBound) return;
    sectionToggleBound = true;

  const sections = document.querySelectorAll(".activity-section");

  for (const section of sections) {
    const head = section.querySelector(".activity-section-head");
    const toggle = section.querySelector("[data-activity-toggle]");

    if (!head || !toggle) continue;

    const controls = toggle.getAttribute("aria-controls");
    const content = controls ? document.getElementById(controls) : null;
    if (!content) continue;

    // Initial state
    const isExpanded = toggle.getAttribute("aria-expanded") === "true";
    content.hidden = !isExpanded;

    const toggleSection = () => {
      const expanded = toggle.getAttribute("aria-expanded") === "true";
      const nextExpanded = !expanded;
      toggle.setAttribute("aria-expanded", String(nextExpanded));
      content.hidden = !nextExpanded;
    };

    // Click anywhere on header
    head.addEventListener("click", (e) => {
      // If user clicked directly on the button, ignore (already handled)
      if (e.target.closest("[data-activity-toggle]")) return;
      toggleSection();
    });

    // Keep button working normally
    toggle.addEventListener("click", (e) => {
      e.stopPropagation();
      toggleSection();
    });
  }
}

function openSectionByHash() {
  const hash = window.location.hash.replace("#", "");
  if (!hash) return;

    // close all sections first
  document.querySelectorAll("[data-activity-toggle]").forEach(toggle => {
    const controls = toggle.getAttribute("aria-controls");
    const content = controls ? document.getElementById(controls) : null;
    if (!content) return;
    toggle.setAttribute("aria-expanded", "false");
    content.hidden = true;
  });

  const map = {
    created: "created-section-content",
    comments: "comments-section-content",
    liked: "liked-section-content",
    disliked: "disliked-section-content",
  };

  const targetId = map[hash];
  if (!targetId) return;

  const content = document.getElementById(targetId);
  if (!content) return;

  const toggle = document.querySelector(
    `[aria-controls="${targetId}"]`
  );

  if (toggle) {
    toggle.setAttribute("aria-expanded", "true");
  }

  content.hidden = false;

  content.scrollIntoView({
    behavior: "smooth",
    block: "start",
  });
}

async function renderActivity(state, pager) {
  const paginationEl = document.getElementById("activity-pagination");
  if (!paginationEl) return;

  closeActiveCommentEditor();
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

function openCommentInlineEditor(article, refresh) {
  if (!(article instanceof HTMLElement)) return;

  const commentId = article.dataset.commentId;
  if (!commentId) return;

  if (activeCommentEditor?.commentId === commentId) return;

  closeActiveCommentEditor();

  const content = article.querySelector(".activity-comment-content");
  const footer = article.querySelector(".activity-comment-footer");
  if (!(content instanceof HTMLElement)) return;

  const originalBody = article.dataset.commentBody ?? "";
  const originalImageURL = article.dataset.commentImageUrl ?? "";
  const originalMarkup = content.innerHTML;

  article.classList.add("is-editing");
  if (footer instanceof HTMLElement) {
    footer.hidden = true;
  }

  content.innerHTML = `
    <form class="comment-form activity-comment-edit-form" data-comment-edit-form novalidate>
      <div class="comment-textarea-wrap">
        <textarea rows="3" placeholder="Edit your comment..."></textarea>
        <button type="button" class="comment-image-btn" aria-label="Attach image">
          <img src="/static/img/camera.png" alt="" />
        </button>
        <input type="file" class="comment-image-input" accept="${IMAGE_ACCEPT_ATTR}" hidden />
      </div>

      <div class="comment-image-row">
        <span class="comment-image-name muted" aria-live="polite"></span>
        <button type="button" class="image-clear" aria-label="Remove selected image" hidden>x</button>
      </div>

      <div class="comment-image-preview activity-comment-edit-preview" hidden>
        <img alt="Selected comment image preview" />
      </div>

      <p class="comment-error" role="alert" hidden></p>

      <div class="activity-comment-edit-actions">
        <button type="submit" class="btn btn-primary btn-sm comment-update">Update</button>
        <button type="button" class="btn btn-outline btn-sm comment-cancel-edit">Cancel</button>
      </div>
    </form>
  `;

  const form = content.querySelector("[data-comment-edit-form]");
  if (!(form instanceof HTMLFormElement)) {
    content.innerHTML = originalMarkup;
    article.classList.remove("is-editing");
    if (footer instanceof HTMLElement) {
      footer.hidden = false;
    }
    return;
  }

  const textarea = form.querySelector("textarea");
  const imageButton = form.querySelector(".comment-image-btn");
  const imageInput = form.querySelector(".comment-image-input");
  const imageName = form.querySelector(".comment-image-name");
  const imageClear = form.querySelector(".image-clear");
  const imagePreview = form.querySelector(".comment-image-preview");
  const imagePreviewImg = imagePreview?.querySelector("img");
  const errorEl = form.querySelector(".comment-error");
  const updateBtn = form.querySelector(".comment-update");
  const cancelBtn = form.querySelector(".comment-cancel-edit");

  if (textarea instanceof HTMLTextAreaElement) {
    textarea.value = originalBody;
  }

  let clearedPersistedImage = false;
  const picker = setupImagePicker({
    input: imageInput,
    triggerButton: imageButton,
    clearButton: imageClear,
    nameLabel: imageName,
    previewContainer: imagePreview,
    previewImage: imagePreviewImg,
    persistedUrl: originalImageURL || null,
    persistedLabel: originalImageURL ? "Current image" : "",
    onTooLarge: () => {
      uiNotify("Image must be 20MB or smaller.", { type: "danger" });
    },
    onClearPersisted: () => {
      clearedPersistedImage = true;
    },
  });

  const closeEditor = ({ playCancelSound = false, restore = true } = {}) => {
    picker?.destroy?.();
    if (restore && document.contains(article)) {
      content.innerHTML = originalMarkup;
      article.classList.remove("is-editing");
      if (footer instanceof HTMLElement) {
        footer.hidden = false;
      }
    }
    if (playCancelSound) {
      playDelete();
    }
  };

  activeCommentEditor = { commentId, close: closeEditor };

  let isSubmitting = false;
  const setBusy = (busy) => {
    if (updateBtn instanceof HTMLButtonElement) {
      updateBtn.disabled = busy;
    }
    if (cancelBtn instanceof HTMLButtonElement) {
      cancelBtn.disabled = busy;
    }
    if (imageButton instanceof HTMLButtonElement) {
      imageButton.disabled = busy;
    }
  };

  cancelBtn?.addEventListener("click", (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (isSubmitting) return;

    if (activeCommentEditor?.commentId === commentId) {
      activeCommentEditor = null;
    }
    closeEditor({ playCancelSound: true, restore: true });
  });

  ["click", "mousedown", "keydown"].forEach((evt) =>
    form.addEventListener(evt, (e) => e.stopPropagation())
  );

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (isSubmitting) return;

    const body = textarea instanceof HTMLTextAreaElement ? textarea.value.trim() : "";
    const imageFile = picker?.getFile?.() || null;
    const keepsExistingImage =
      !!originalImageURL && !clearedPersistedImage && !imageFile;
    const hasAnyImage = !!imageFile || keepsExistingImage;

    if (!body && !hasAnyImage) {
      if (errorEl instanceof HTMLElement) {
        errorEl.textContent = "Cannot save an empty comment.";
        errorEl.hidden = false;
      }
      return;
    }
    if (errorEl instanceof HTMLElement) {
      errorEl.hidden = true;
    }

    isSubmitting = true;
    setBusy(true);

    try {
      const requestOptions = buildImageRequestOptions({
        method: "PATCH",
        imageFile,
        buildMultipartBody: (file) => {
          const formData = new FormData();
          formData.append("body", body);
          formData.append("image", file);
          return formData;
        },
        jsonBody: {
          body,
          ...(clearedPersistedImage && !imageFile ? { remove_image: true } : {}),
        },
        multipartHeaders: { Accept: "application/json" },
        jsonHeaders: { Accept: "application/json" },
      });

      const res = await fetch(`${API_BASE}/comments/${commentId}`, requestOptions);
      const payload = await res.json().catch(() => null);
      const message = payload?.error?.message || "Failed to update comment.";

      if (res.status === 401) {
        uiNotify("You must be logged in.", { type: "warn" });
        return;
      }
      if (res.status === 403) {
        uiNotify("You can only update your own comments.", { type: "warn" });
        return;
      }
      if (res.status === 404) {
        uiNotify("Comment not found.", { type: "warn" });
        if (activeCommentEditor?.commentId === commentId) {
          activeCommentEditor = null;
        }
        closeEditor({ restore: false });
        await refresh();
        return;
      }
      if (!res.ok) {
        if (errorEl instanceof HTMLElement) {
          errorEl.textContent = message;
          errorEl.hidden = false;
        } else {
          uiNotify(message, { type: "danger" });
        }
        return;
      }

      playUpload();
      uiNotify("Comment updated.", { type: "success" });
      if (activeCommentEditor?.commentId === commentId) {
        activeCommentEditor = null;
      }
      closeEditor({ restore: false });
      await refresh();
    } catch (err) {
      console.error("Activity comment update failed:", err);
      if (errorEl instanceof HTMLElement) {
        errorEl.textContent = "Failed to update comment.";
        errorEl.hidden = false;
      } else {
        uiNotify("Failed to update comment.", { type: "danger" });
      }
    } finally {
      isSubmitting = false;
      setBusy(false);
    }
  });
}

function initEditComment(refresh) {
  if (commentEditBound) return;
  commentEditBound = true;

  document.addEventListener(
    "click",
    (e) => {
      const btn = e.target.closest(".comment-edit");
      if (!btn) return;

      e.preventDefault();
      e.stopPropagation();

      const article = btn.closest(".activity-comment");
      if (!(article instanceof HTMLElement)) return;
      openCommentInlineEditor(article, refresh);
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

        const img = btn.querySelector("img");
        if (img) {
          img.src =
            nextStatus === "draft"
              ? "/static/img/publish.png"
              : "/static/img/draft.png";
        }

        btn.title =
          nextStatus === "draft"
            ? "Publish post"
            : "Move to draft";

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
        openSectionByHash(); 
      } catch (err) {
        console.error("Activity pagination failed:", err);
        uiNotify("Failed to load activity.", { type: "danger" });
      }
    },
  });

  const refresh = async () => {
    await renderActivity(state, pager);
    openSectionByHash();
  };

  prevBtn.addEventListener("click", () => pager.prev());
  nextBtn.addEventListener("click", () => pager.next());

  statusSelect.addEventListener("change", async () => {
    state.status = statusSelect.value;
    state.page = DEFAULT_PAGE;
    setQueryState(state);
    try {
      await renderActivity(state, pager);
      openSectionByHash(); // <-- add here
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
      openSectionByHash(); // <-- add here
    } catch (err) {
      console.error("Activity per-page update failed:", err);
      uiNotify("Failed to load activity.", { type: "danger" });
    }
  });

  /* ---------------- INITIAL RENDER ---------------- */

  try {
    await renderActivity(state, pager);
    openSectionByHash(); // <-- THIS WAS MISSING
  } catch (err) {
    console.error("Activity render failed:", err);
    uiNotify("Failed to load activity.", { type: "danger" });
  }

  /* ---------------- HISTORY NAVIGATION ---------------- */

  window.addEventListener("popstate", async () => {
    state = getQueryState();
    statusSelect.value = state.status;
    perPageSelect.value = String(state.perPage);

    try {
      await renderActivity(state, pager);
      openSectionByHash(); // <-- also here
    } catch (err) {
      console.error("Activity history navigation failed:", err);
      uiNotify("Failed to load activity.", { type: "danger" });
    }
  });

  window.addEventListener("hashchange", openSectionByHash);

  initEditPostNavigation();
  initStatusToggle(refresh);
  initDeletePost(refresh);
  initEditComment(refresh);
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
