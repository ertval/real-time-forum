// web/static/js/reactions.js
import { API_BASE } from "./utils.js";
import { Auth } from "./auth.js";
import { playReaction } from "./sound-effects.js";

let bound = false;

export function initReactions() {
  if (bound) return;
  bound = true;

  document.addEventListener("change", async (e) => {
    const btn = e.target.closest("[data-reaction]");
    if (!btn) return;

    e.stopPropagation();

    // Save previous state for rollback
    const previousState = !btn.checked;

    // Guest → force login modal
    const allowed = await Auth.requireOrPrompt();
    if (!allowed) {
      btn.checked = previousState; // rollback immediately
      return;
    }

    const type = btn.dataset.reaction;
    const postId = btn.dataset.postId;
    const commentId = btn.dataset.commentId;

    const scope = postId
      ? btn.closest("article[data-post-id]")
      : btn.closest(".comment");

    const oppositeType = type === "like" ? "dislike" : "like";
    const opposite = scope?.querySelector(
      `input[data-reaction="${oppositeType}"]`
    );

    const oppositePreviousState = opposite ? opposite.checked : null;

    const url = postId
      ? `${API_BASE}/posts/${postId}/${type}`
      : `${API_BASE}/comments/${commentId}/${type}`;

    try {
      const res = await fetch(url, {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      if (res.status === 401) {
        btn.checked = previousState;
        if (opposite && oppositePreviousState !== null) {
          opposite.checked = oppositePreviousState;
        }
        await Auth.requireOrPrompt();
        return;
      }

      if (!res.ok) {
        btn.checked = previousState;
        if (opposite && oppositePreviousState !== null) {
          opposite.checked = oppositePreviousState;
        }
        return;
      }

      // Only now apply mutual exclusion
      if (btn.checked && opposite) {
        opposite.checked = false;
      }

      playReaction();

      const { data } = await res.json();

      const container = btn.closest(
        postId
          ? `article[data-post-id="${postId}"]`
          : `div[data-comment-id="${commentId}"]`
      );

      if (!container) return;

      container.querySelector("[data-like-count]").textContent =
        data.likes_count;

      container.querySelector("[data-dislike-count]").textContent =
        data.dislikes_count;

    } catch (err) {
      console.error("Reaction failed:", err);

      // Full rollback on network error
      btn.checked = previousState;
      if (opposite && oppositePreviousState !== null) {
        opposite.checked = oppositePreviousState;
      }
    }
  });
}