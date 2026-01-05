// web/static/js/reactions.js


function bindReactions(container = document) {
  container.addEventListener("click", async (e) => {
    const likeBtn = e.target.closest("[data-like]");
    const dislikeBtn = e.target.closest("[data-dislike]");

    if (!likeBtn && !dislikeBtn) return;

    e.stopPropagation(); 

    const postId = likeBtn
      ? likeBtn.dataset.like
      : dislikeBtn.dataset.dislike;

    const type = likeBtn ? "like" : "dislike";

    try {
      const res = await fetch(`/api/v1/posts/${postId}/${type}`, {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      if (!res.ok) {
        alert("You must be logged in to react");
        return;
      }

      const payload = await res.json();
      const data = payload.data;

      // ενημέρωση counters
      const postEl = container.querySelector(
        `[data-post-id="${postId}"]`
      );

      if (!postEl) return;

      postEl.querySelector(".like-count").textContent =
        data.likes_count;

      postEl.querySelector(".dislike-count").textContent =
        data.dislikes_count;

    } catch (err) {
      console.error("Reaction failed:", err);
    }
  });
}


async function handleReaction(postId, type, btn) {
  if (!postId) return;

  btn.disabled = true;

  try {
    const res = await fetch(`${API_BASE}/posts/${postId}/${type}`, {
      method: "POST",
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (!res.ok) {
      if (res.status === 401) {
        window.location.href = "/login";
        return;
      }
      throw new Error("Reaction failed");
    }

    const payload = await res.json();
    updateCounters(postId, payload.data);
  } catch (err) {
    console.error(err);
  } finally {
    btn.disabled = false;
  }
}

function updateCounters(postId, data) {
  const card = document.querySelector(`[data-post-id="${postId}"]`);
  if (!card) return;

  const likeCount = card.querySelector(".like-count");
  const dislikeCount = card.querySelector(".dislike-count");

  if (likeCount) likeCount.textContent = data.likes;
  if (dislikeCount) dislikeCount.textContent = data.dislikes;
}
