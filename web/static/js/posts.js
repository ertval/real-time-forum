// web/static/js/posts.js

import {
  API_BASE,
  buildImageRequestOptions,
  formatCreatedAt,
  IMAGE_ACCEPT_ATTR,
  resolveUsername,
  escapeHTML,
} from "./utils.js";

import { Auth } from "./auth.js";
import { playUpload } from "./sound-effects.js";
import { uiNotify } from "./ui-messages.js";
import { setupImagePicker } from "./image-picker.js";

let imageLightbox = null;
let imageLightboxImg = null;
let imageLightboxCloseBtn = null;
let lastFocusedElement = null;

/*-----------
  POST CARD
-----------*/

export function renderPostCard(
  post,
  {
    clickable = true,
    showStatusToggle = false,
    showDelete = false,
    showEdit = false,
  } = {}
) {
  const article = document.createElement("article");
  article.className = "post card card-pad";
  article.dataset.postId = post.id;

  const imageUrl =
    typeof post.image_url === "string" && post.image_url.trim()
      ? post.image_url
      : "";

  const imageMarkup = imageUrl
    ? `
      <div class="post-image">
        <div class="post-image-ambient" aria-hidden="true"></div>
        <img
          src="${escapeHTML(imageUrl)}"
          alt="${escapeHTML(post.title)}"
          loading="lazy"
          class="expandable-image"
          tabindex="0"
          role="button"
          aria-label="Expand post image"
        />
      </div>
    `
    : "";

  const bodyMarkup = post.body
    ? `<p>${escapeHTML(post.body)}</p>`
    : "";

  article.innerHTML = `
    <header class="post-header ${clickable ? "clickable" : ""}">
      <div>
        ${renderCategories(post.categories)}
        <h3 class="post-title">${escapeHTML(post.title)}</h3>
        <p class="muted">Author: ${resolveUsername(post)}</p>
      </div>

      <div class="post-header-right">
        <time class="muted">${formatCreatedAt(post.created_at)}</time>
        ${showStatusToggle ? statusToggleTemplate(post) : ""}
        ${showEdit ? editPostTemplate(post) : ""}
        ${showDelete ? deletePostTemplate(post) : ""}
      </div>
    </header>

    <section class="post-body ${clickable ? "clickable" : ""}">
      ${imageMarkup}
      ${bodyMarkup}
    </section>

    <section class="post-actions">
      ${reactionTemplate(post)}
    </section>

    <section class="post-comments" data-comments></section>
  `;

  if (clickable) {
    article.querySelectorAll(".clickable").forEach(el => {
      el.addEventListener("click", () => {
        window.location.href = `/view-post/${post.id}`;
      });
    });
  }

  const postImage = article.querySelector(".post-image img");
  bindExpandableImage(postImage, "post");

  const postImageFrame = article.querySelector(".post-image");
  const postImageAmbient = article.querySelector(".post-image-ambient");
  syncImageTransparencyPresentation({
    imgEl: postImage,
    frameEl: postImageFrame,
    checkerboardClass: "post-image--checkerboard",
    ambientEl: postImageAmbient,
  });

  article
    .querySelectorAll(".post-comments, textarea, form")
    .forEach(el =>
      el.addEventListener("click", e => e.stopPropagation())
    );

  return article;
}

/*----------
  COMMENTS
----------*/

export async function loadPostCommentsPreview(postId, article) {
  const container = article.querySelector("[data-comments]");
  if (!container) return;

  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (!res.ok) return;

    const payload = await res.json();
    const comments = extractArray(payload);

    const list = document.createElement("div");
    list.className = "comments comments-scroll";

    if (comments.length === 0) {
      list.innerHTML = `<p class="muted">No comments yet.</p>`;
    } else {
      comments.forEach(c => list.appendChild(renderComment(c)));
    }

    container.appendChild(list);
    list.scrollTop = list.scrollHeight;

    maybeRenderCommentForm(container, postId);

  } catch (err) {
    console.error("Failed to load comments:", err);
  }
}


function renderComment(comment) {
  const div = document.createElement("div");
  div.className = "comment";
  div.dataset.commentId = comment.id;
  div.id = `comment-${comment.id}`;   

  const commentImageUrl =
    typeof comment.image_url === "string" && comment.image_url.trim()
      ? comment.image_url
      : "";

  const commentBody = typeof comment.body === "string" ? comment.body : "";
  const bodyMarkup = commentBody.trim()
    ? `<p class="comment-text">${escapeHTML(commentBody)}</p>`
    : "";
  const imageMarkup = commentImageUrl
    ? `
      <div class="comment-image">
        <img
          src="${escapeHTML(commentImageUrl)}"
          alt="Comment image by ${escapeHTML(resolveUsername(comment))}"
          loading="lazy"
          tabindex="0"
          role="button"
          aria-label="Expand comment image"
        />
      </div>
    `
    : "";

  div.innerHTML = `
    <div class="comment-meta muted">
      <strong>${resolveUsername(comment)}</strong>
      · ${formatCreatedAt(comment.created_at)}
    </div>

    <div class="comment-body">
      ${bodyMarkup}
      ${imageMarkup}
    </div>

    <div class="comment-actions">
      ${reactionTemplate(comment, true)}
    </div>
  `;

  const commentImage = div.querySelector(".comment-image img");
  bindExpandableImage(commentImage, "comment");
  syncImageTransparencyPresentation({
    imgEl: commentImage,
    frameEl: div.querySelector(".comment-image"),
    checkerboardClass: "comment-image--checkerboard",
  });

  return div;
}

/*--------------
  COMMENT FORM
--------------*/

function maybeRenderCommentForm(container, postId) {
  const form = document.createElement("form");
  form.className = "comment-form";
  form.noValidate = true;

  form.innerHTML = `
    <div class="comment-textarea-wrap">
      <textarea placeholder="Write a comment..." rows="2"></textarea>
      <button type="button" class="comment-image-btn" aria-label="Attach image">
        <img src="/static/img/camera.png" alt="" />
      </button>
      <input type="file" class="comment-image-input" accept="${IMAGE_ACCEPT_ATTR}" hidden />
    </div>
    <div class="comment-image-row">
      <span class="comment-image-name muted" aria-live="polite"></span>
      <button type="button" class="image-clear" aria-label="Remove selected image" hidden>x</button>
    </div>
    <div class="comment-image-preview" hidden>
      <img alt="Selected comment image preview" />
    </div>
    <p class="comment-error" role="alert" hidden></p>
    <button class="btn btn-primary" type="submit">Comment</button>
  `;

  const textarea = form.querySelector("textarea");
  const imageButton = form.querySelector(".comment-image-btn");
  const imageInput = form.querySelector(".comment-image-input");
  const imageName = form.querySelector(".comment-image-name");
  const imageClear = form.querySelector(".image-clear");
  const imagePreview = form.querySelector(".comment-image-preview");
  const imagePreviewImg = imagePreview?.querySelector("img");
  const submitButton = form.querySelector("button[type='submit']");
  let isSubmitting = false;
  let imagePicker = null;

  const ensureImagePicker = () => {
    if (imagePicker) return imagePicker;
    imagePicker = setupImagePicker({
      input: imageInput,
      triggerButton: imageButton,
      clearButton: imageClear,
      nameLabel: imageName,
      previewContainer: imagePreview,
      previewImage: imagePreviewImg,
      onTooLarge: () => {
        uiNotify("Image must be 20MB or smaller.", { type: "danger" });
      },
    });
    return imagePicker;
  };

  imageButton?.addEventListener("click", e => {
    if (imagePicker) return;
    e.preventDefault();
    e.stopPropagation();
    ensureImagePicker();
    imageInput?.click();
  }, { capture: true });

  ["click", "mousedown", "keydown", "submit"].forEach(evt =>
    form.addEventListener(evt, e => {
      e.stopPropagation();
      if (evt === "submit") e.preventDefault();
    })
  );

  form.addEventListener("submit", async () => {
    if (isSubmitting) return;
    isSubmitting = true;
    if (submitButton instanceof HTMLButtonElement) {
      submitButton.disabled = true;
    }

    try {
      const allowed = await Auth.requireOrPrompt();
      if (!allowed) return;

      const errorEl = form.querySelector(".comment-error");
      const body = textarea.value.trim();
      const imageFile =
        imagePicker?.getFile() ||
        imageInput?.files?.[0] ||
        null;
      const hasImage = !!imageFile;

      if (!body && !hasImage) {
        errorEl.textContent = "Cannot submit an empty comment";
        errorEl.hidden = false;
        return;
      }

      errorEl.hidden = true;

      const res = await fetch(
        `${API_BASE}/posts/${postId}/comments`,
        buildImageRequestOptions({
          method: "POST",
          imageFile,
          buildMultipartBody: file => {
            const formData = new FormData();
            formData.append("body", body);
            formData.append("image", file);
            return formData;
          },
          jsonBody: { body },
          multipartHeaders: { Accept: "application/json" },
          jsonHeaders: { Accept: "application/json" },
        })
      );

      if (!res.ok) {
        const payload = await res.json().catch(() => null);
        errorEl.textContent = payload?.error?.message || "Failed to submit comment";
        errorEl.hidden = false;
        return;
      }

      const payload = await res.json();
      const newComment = payload.data ?? payload;

      textarea.value = "";
      if (imagePicker) {
        imagePicker.clearSelectedFile();
      } else if (imageInput) {
        imageInput.value = "";
      }

      playUpload();

      const list = container.querySelector(".comments-scroll");
      list?.appendChild(renderComment(newComment));
      list.scrollTop = list.scrollHeight;

      /* highlight newly added comment */
      const newEl = document.getElementById(`comment-${newComment.id}`);
      if (newEl) {
        newEl.classList.add("highlight-comment");

        // fade-out starts after 1s
        setTimeout(() => {
          newEl.classList.add("fade-out");

          // remove highlight classes once fade-out finishes (~1.2s)
          setTimeout(() => {
            newEl.classList.remove("highlight-comment", "fade-out");
          }, 1200);

        }, 1000);
      }
    } finally {
      isSubmitting = false;
      if (submitButton instanceof HTMLButtonElement) {
        submitButton.disabled = false;
      }
    }
  });

  container.appendChild(form);
}

/*--------------------
  REACTIONS TEMPLATE
--------------------*/

export function reactionTemplate(item, isComment = false) {
  const idAttr = isComment
    ? `data-comment-id="${item.id}"`
    : `data-post-id="${item.id}"`;

  return `
    <div class="reaction">
      <label class="reaction-toggle reaction-toggle--like" aria-label="Like">
        <input type="checkbox" data-reaction="like" ${idAttr}>
        <svg viewBox="0 0 32 32" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg">
          <path d="M29.845,17.099l-2.489,8.725C26.989,27.105,25.804,28,24.473,28H11c-0.553,0-1-0.448-1-1V13
          c0-0.215,0.069-0.425,0.198-0.597l5.392-7.24C16.188,4.414,17.05,4,17.974,4C19.643,4,21,5.357,21,7.026V12h5.002
          c1.265,0,2.427,0.579,3.188,1.589C29.954,14.601,30.192,15.88,29.845,17.099z"></path>
          <path d="M7,12H3c-0.553,0-1,0.448-1,1v14c0,0.552,0.447,1,1,1h4c0.553,0,1-0.448,1-1V13C8,12.448,7.553,12,7,12z
          M5,25.5c-0.828,0-1.5-0.672-1.5-1.5c0-0.828,0.672-1.5,1.5-1.5c0.828,0,1.5,0.672,1.5,1.5C6.5,24.828,5.828,25.5,5,25.5z"></path>
        </svg>
      </label>
      <span class="data-like-count" data-like-count>${item.likes ?? 0}</span>
    </div>

    <div class="reaction">
      <label class="reaction-toggle reaction-toggle--dislike" aria-label="Dislike">
        <input type="checkbox" data-reaction="dislike" ${idAttr}>
        <svg viewBox="0 0 32 32" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg">
          <path d="M2.156,14.901l2.489-8.725C5.012,4.895,6.197,4,7.528,4h13.473C21.554,4,22,4.448,22,5v14
          c0,0.215-0.068,0.425-0.197,0.597l-5.392,7.24C15.813,27.586,14.951,28,14.027,28c-1.669,0-3.026-1.357-3.026-3.026V20H5.999
          c-1.265,0-2.427-0.579-3.188-1.589C2.047,17.399,1.809,16.12,2.156,14.901z"></path>
          <path d="M25.001,20h4C29.554,20,30,19.552,30,19V5c0-0.552-0.446-1-0.999-1h-4c-0.553,0-1,0.448-1,1v14
          C24.001,19.552,24.448,20,25.001,20z M27.001,6.5c0.828,0,1.5,0.672,1.5,1.5c0,0.828-0.672,1.5-1.5,1.5c-0.828,0-1.5-0.672-1.5-1.5
          C25.501,7.172,26.173,6.5,27.001,6.5z"></path>
        </svg>
      </label>
      <span class="data-dislike-count" data-dislike-count>${item.dislikes ?? 0}</span>
    </div>
  `;
}

/*---------
  HELPERS
---------*/

function extractArray(payload) {
  if (Array.isArray(payload)) return payload;
  if (payload?.data && Array.isArray(payload.data)) return payload.data;
  if (payload?.data?.data && Array.isArray(payload.data.data)) {
    return payload.data.data;
  }
  return [];
}

function statusToggleTemplate(post) {
  if (!post.status) return "";
  return `
    <button class="btn btn-outline btn-sm post-status-toggle"
      data-post-id="${post.id}"
      data-current-status="${post.status}">
      ${post.status === "draft" ? "Publish" : "Draft"}
    </button>
  `;
}

function renderCategories(categories = []) {
  if (!Array.isArray(categories) || categories.length === 0) return "";
  return `
    <div class="post-categories">
      ${categories.map(c => `<span class="category-badge">${c.name}</span>`).join("")}
    </div>
  `;
}

function deletePostTemplate(post) {
  return `
    <button class="btn btn-danger btn-sm post-delete"
      data-post-id="${post.id}">
      Delete
    </button>
  `;
}

function editPostTemplate(post) {
  return `
    <button class="btn btn-outline btn-sm post-edit"
      data-post-id="${post.id}">
      Edit
    </button>
  `;
}

function bindExpandableImage(imgEl, variant = "post") {
  if (!(imgEl instanceof HTMLImageElement)) return;

  const openImage = e => {
    e.preventDefault();
    e.stopPropagation();
    const useCheckerboard = imgEl.dataset.transparent === "true";
    const imageRect = imgEl.getBoundingClientRect();
    const minDimensions = {
      minWidth: Math.max(0, Math.round(imageRect.width)),
      minHeight: Math.max(0, Math.round(imageRect.height)),
    };

    if (variant === "comment") {
      const postImage = document.querySelector(".post-image img");
      if (postImage instanceof HTMLImageElement) {
        const postRect = postImage.getBoundingClientRect();
        minDimensions.minWidth = Math.max(
          minDimensions.minWidth,
          Math.round(postRect.width)
        );
        minDimensions.minHeight = Math.max(
          minDimensions.minHeight,
          Math.round(postRect.height)
        );
      }
    }

    openImageLightbox(
      imgEl.currentSrc || imgEl.src,
      imgEl.alt,
      variant,
      useCheckerboard,
      minDimensions
    );
  };

  imgEl.addEventListener("click", openImage);
  imgEl.addEventListener("keydown", e => {
    if (e.key === "Enter" || e.key === " ") {
      openImage(e);
    }
  });
}

function syncImageTransparencyPresentation({
  imgEl,
  frameEl = null,
  checkerboardClass = "",
  ambientEl = null,
} = {}) {
  if (!(imgEl instanceof HTMLImageElement)) return;

  const sync = async () => {
    const src = imgEl.currentSrc || imgEl.src;
    if (!src) return;
    const hasTransparency = await isTransparentPng(imgEl, src);
    if ((imgEl.currentSrc || imgEl.src) !== src) return;

    imgEl.dataset.transparent = hasTransparency ? "true" : "false";
    if (frameEl instanceof Element && checkerboardClass) {
      frameEl.classList.toggle(checkerboardClass, hasTransparency);
    }
    if (ambientEl instanceof HTMLElement) {
      ambientEl.style.backgroundImage = hasTransparency ? "" : `url("${src}")`;
    }
  };

  sync();
  imgEl.addEventListener("load", sync);
}

function isPngSource(src) {
  if (!src) return false;
  try {
    const parsed = new URL(src, window.location.href);
    return parsed.pathname.toLowerCase().endsWith(".png");
  } catch {
    return src.split("?")[0].toLowerCase().endsWith(".png");
  }
}

async function isTransparentPng(imgEl, srcHint = "") {
  const src = srcHint || imgEl.currentSrc || imgEl.src;
  if (!isPngSource(src)) return false;

  if (!imgEl.naturalWidth || !imgEl.naturalHeight) return false;

  try {
    const sampleWidth = Math.min(80, imgEl.naturalWidth);
    const sampleHeight = Math.min(80, imgEl.naturalHeight);

    const canvas = document.createElement("canvas");
    canvas.width = sampleWidth;
    canvas.height = sampleHeight;

    const ctx = canvas.getContext("2d", { willReadFrequently: true });
    if (!ctx) return false;

    ctx.drawImage(imgEl, 0, 0, sampleWidth, sampleHeight);
    const { data } = ctx.getImageData(0, 0, sampleWidth, sampleHeight);

    for (let i = 3; i < data.length; i += 4) {
      if (data[i] < 250) return true;
    }
  } catch {
    return false;
  }

  return false;
}

function ensureImageLightbox() {
  if (imageLightbox) return;

  const wrapper = document.createElement("div");
  wrapper.className = "image-lightbox";
  wrapper.hidden = true;
  wrapper.innerHTML = `
    <div class="image-lightbox-backdrop" data-close-lightbox></div>
    <figure class="image-lightbox-content" role="dialog" aria-modal="true" aria-label="Expanded post image">
      <button type="button" class="image-lightbox-close" aria-label="Close expanded image">
        <img src="/static/img/close.png" alt="" aria-hidden="true" class="image-lightbox-close-icon" />
      </button>
      <img class="image-lightbox-image" alt="Expanded post image" />
    </figure>
  `;

  document.body.appendChild(wrapper);
  imageLightbox = wrapper;
  imageLightboxImg = wrapper.querySelector(".image-lightbox-image");
  imageLightboxCloseBtn = wrapper.querySelector(".image-lightbox-close");

  wrapper.addEventListener("click", e => {
    const target = e.target;
    if (!(target instanceof Element)) return;
    if (
      target.matches("[data-close-lightbox]") ||
      target.closest(".image-lightbox-close")
    ) {
      closeImageLightbox();
    }
  });

  document.addEventListener("keydown", e => {
    if (e.key === "Escape" && imageLightbox && !imageLightbox.hidden) {
      closeImageLightbox();
    }
  });
}

function openImageLightbox(
  src,
  alt = "",
  variant = "post",
  useCheckerboard = false,
  minDimensions = {}
) {
  if (!src) return;
  ensureImageLightbox();
  if (!imageLightbox || !imageLightboxImg) return;

  lastFocusedElement = document.activeElement;
  const { minWidth = 0, minHeight = 0 } = minDimensions;
  imageLightboxImg.style.setProperty(
    "--lightbox-min-width",
    `${Math.max(0, minWidth)}px`
  );
  imageLightboxImg.style.setProperty(
    "--lightbox-min-height",
    `${Math.max(0, minHeight)}px`
  );
  imageLightboxImg.src = src;
  imageLightboxImg.alt = alt || "Expanded post image";
  if (useCheckerboard) {
    imageLightbox.dataset.checkerboard = "true";
  } else {
    delete imageLightbox.dataset.checkerboard;
  }
  imageLightbox.dataset.variant = variant;
  imageLightbox.hidden = false;
  document.body.classList.add("image-lightbox-open");
  imageLightboxCloseBtn?.focus();
}

function closeImageLightbox() {
  if (!imageLightbox || imageLightbox.hidden) return;

  imageLightbox.hidden = true;
  delete imageLightbox.dataset.variant;
  delete imageLightbox.dataset.checkerboard;
  document.body.classList.remove("image-lightbox-open");
  imageLightboxImg?.style.removeProperty("--lightbox-min-width");
  imageLightboxImg?.style.removeProperty("--lightbox-min-height");
  imageLightboxImg?.removeAttribute("src");

  if (lastFocusedElement instanceof HTMLElement) {
    lastFocusedElement.focus();
  }
}
