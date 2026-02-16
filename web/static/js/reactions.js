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

    // Guest → force login modal
    const allowed = await Auth.requireOrPrompt();
    if (!allowed) return;

    const type = btn.dataset.reaction;
    const postId = btn.dataset.postId;
    const commentId = btn.dataset.commentId;

    // Determine if scope is a post or a comment
    const scope = postId
      ? btn.closest("article[data-post-id]")
      : btn.closest(".comment");

    // Mutually exclusive toggles
    if (btn.checked && scope) {
      const oppositeType = type === "like" ? "dislike" : "like";
      const opposite = scope.querySelector(
        `input[data-reaction="${oppositeType}"]`
      );
      if (opposite) opposite.checked = false;
    }

    const url = postId
      ? `${API_BASE}/posts/${postId}/${type}`
      : `${API_BASE}/comments/${commentId}/${type}`;

    try {
      const res = await fetch(url, {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      if (!res.ok) return;

      // Play unified reaction sound
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
    }
  });
}
