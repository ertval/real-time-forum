// /static/js/activity/comments-activity.js

import {
  API_BASE,
  buildImageRequestOptions,
  IMAGE_ACCEPT_ATTR,
} from "../utils.js";
import { uiNotify, uiConfirm } from "../ui-messages.js";
import { playDelete, playUpload } from "../sound-effects.js";
import { setupImagePicker } from "../image-picker.js";

let commentDeleteBound = false;
let commentEditBound = false;
let activeCommentEditor = null;

function closeActiveCommentEditor() {
  if (!activeCommentEditor?.close) return;
  activeCommentEditor.close();
  activeCommentEditor = null;
}

export function initEditComment(refresh) {
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

export function initDeleteComment(refresh) {
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

/* ===============================
   INLINE EDITOR (unchanged logic)
=================================*/

function openCommentInlineEditor(article, refresh) {
  if (!(article instanceof HTMLElement)) return;

  const commentId = article.dataset.commentId;
  if (!commentId) return;

  if (activeCommentEditor?.commentId === commentId) return;

  closeActiveCommentEditor();

  const content = article.querySelector(".activity-comment-content");
  if (!(content instanceof HTMLElement)) return;

  const originalMarkup = content.innerHTML;
  const originalBody = article.dataset.commentBody ?? "";
  const originalImageURL = article.dataset.commentImageUrl ?? "";

  article.classList.add("is-editing");

  content.innerHTML = `
    <form class="comment-form activity-comment-edit-form" data-comment-edit-form novalidate>
      <div class="comment-textarea-wrap">
        <textarea rows="3"></textarea>
        <button type="button" class="comment-image-btn">
          <img src="/static/img/camera.png" alt="" />
        </button>
        <input type="file" class="comment-image-input" accept="${IMAGE_ACCEPT_ATTR}" hidden />
      </div>

      <div class="comment-image-preview activity-comment-edit-preview" hidden>
        <img />
      </div>

      <div class="activity-comment-edit-actions">
        <button type="submit" class="btn btn-primary btn-sm">Update</button>
        <button type="button" class="btn btn-outline btn-sm cancel-edit">Cancel</button>
      </div>
    </form>
  `;

  const form = content.querySelector("[data-comment-edit-form]");
  const textarea = form.querySelector("textarea");
  const cancelBtn = form.querySelector(".cancel-edit");

  textarea.value = originalBody;

  const picker = setupImagePicker({
    input: form.querySelector(".comment-image-input"),
    triggerButton: form.querySelector(".comment-image-btn"),
    previewContainer: form.querySelector(".comment-image-preview"),
    previewImage: form.querySelector(".comment-image-preview img"),
    persistedUrl: originalImageURL || null,
  });

  activeCommentEditor = {
    commentId,
    close() {
      picker?.destroy?.();
      content.innerHTML = originalMarkup;
      article.classList.remove("is-editing");
    },
  };

  cancelBtn.addEventListener("click", (e) => {
    e.preventDefault();
    closeActiveCommentEditor();
  });

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const body = textarea.value.trim();
    const imageFile = picker?.getFile?.() || null;

    if (!body && !imageFile && !originalImageURL) {
      uiNotify("Cannot save empty comment.", { type: "warn" });
      return;
    }

    try {
      const requestOptions = buildImageRequestOptions({
        method: "PATCH",
        imageFile,
        jsonBody: { body },
      });

      const res = await fetch(
        `${API_BASE}/comments/${commentId}`,
        requestOptions
      );

      if (!res.ok) {
        uiNotify("Failed to update comment.", { type: "danger" });
        return;
      }

      playUpload();
      uiNotify("Comment updated.", { type: "success" });

      closeActiveCommentEditor();
      await refresh();

    } catch (err) {
      console.error(err);
      uiNotify("Failed to update comment.", { type: "danger" });
    }
  });
}