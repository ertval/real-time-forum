// web/static/js/post-actions.js

export function statusToggleButton(postId, currentStatus) {
	const normalized = currentStatus === "draft" ? "draft" : "published";

	const iconSrc =
		normalized === "draft"
			? "/static/img/publish.png"
			: "/static/img/draft.png";

	const label = normalized === "draft" ? "Publish post" : "Move to draft";

	return `
    <button 
      class="action-icon post-status-toggle"
      data-post-id="${postId}"
      data-current-status="${normalized}"
      aria-label="${label}"
      title="${label}"
    >
      <img src="${iconSrc}" alt="" />
    </button>
  `;
}

export function editPostButton(postId) {
	return `
    <button
      class="action-icon post-edit"
      data-post-id="${postId}"
      aria-label="Edit post"
      title="Edit post"
    >
      <img src="/static/img/edit.png" alt="" />
    </button>
  `;
}

export function deletePostButton(postId) {
	return `
    <button
      class="action-icon delete-icon post-delete"
      data-post-id="${postId}"
      aria-label="Delete post"
      title="Delete post"
    >
      <img src="/static/img/delete.png" alt="" />
    </button>
  `;
}

export function editCommentButton(commentId) {
	return `
    <button
      class="action-icon comment-edit"
      data-comment-id="${commentId}"
      aria-label="Edit comment"
    >
      <img src="/static/img/edit.png" alt="Edit" />
    </button>
  `;
}

export function deleteCommentButton(commentId) {
	return `
    <button
      class="action-icon delete-icon comment-delete"
      data-comment-id="${commentId}"
      aria-label="Delete comment"
    >
      <img src="/static/img/delete.png" alt="Delete" />
    </button>
  `;
}
