// /static/js/activity/posts-activity.js

import { API_BASE } from "../utils.js";
import { uiNotify, uiConfirm } from "../ui-messages.js";
import { playDelete, playUpload } from "../sound-effects.js";

let editBound = false;
let deleteBound = false;
let statusToggleBound = false;

export function initEditPostNavigation() {
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

export function initDeletePost(refresh) {
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

export function initStatusToggle(refresh) {
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

      const nextStatus =
        currentStatus === "draft" ? "published" : "draft";

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