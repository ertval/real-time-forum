// web/static/js/reactions.js

import { Auth } from "./auth.js";
import { playReaction } from "./sound-effects.js";
import { API_BASE } from "./utils.js";

export function initReactions() {
	document.addEventListener("change", async (e) => {
		const btn = e.target.closest("[data-reaction]");
		if (!btn) return;

		e.stopPropagation();

		const previousState = !btn.checked;

		const allowed = await Auth.requireOrPrompt();
		if (!allowed) {
			btn.checked = previousState;
			return;
		}

		const type = btn.dataset.reaction;
		const postId = btn.dataset.postId;
		const commentId = btn.dataset.commentId;

		const scope = postId
			? btn.closest("article[data-post-id]")
			: btn.closest(".comment, .activity-comment");

		if (!scope) return;

		const oppositeType = type === "like" ? "dislike" : "like";
		const opposite = scope.querySelector(
			`input[data-reaction="${oppositeType}"]`,
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

			if (btn.checked && opposite) {
				opposite.checked = false;
			}

			playReaction();

			const { data } = await res.json();

			scope.querySelector("[data-like-count]").textContent = data.likes_count;

			scope.querySelector("[data-dislike-count]").textContent =
				data.dislikes_count;
		} catch (err) {
			console.error("Reaction failed:", err);

			btn.checked = previousState;
			if (opposite && oppositePreviousState !== null) {
				opposite.checked = oppositePreviousState;
			}
		}
	});
}
