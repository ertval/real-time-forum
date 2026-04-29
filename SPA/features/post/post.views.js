// SPA/features/post/post.views.js

export { renderPostDetailView } from './post-detail.views.js';

export function renderEditPostView(postID) {
	return `
		<section data-screen="edit-post" aria-labelledby="screen-edit-post-title">
			<h1 id="screen-edit-post-title">Edit Post</h1>
			<p data-post-id="${postID}">Editing post ${postID}</p>
		</section>
	`;
}

export function renderCreatePostView() {
	return `
		<section data-screen="create-post" aria-labelledby="screen-create-post-title">
			<h1 id="screen-create-post-title">Create Post</h1>
			<p>Compose and publish a new post.</p>
		</section>
	`;
}
