// SPA/features/post/post.views.js

export function renderPostDetailView(postID) {
	return `
		<section data-screen="post-detail" aria-labelledby="screen-post-title">
			<h1 id="screen-post-title">Post Detail</h1>
			<p data-post-id="${postID}">Viewing post ${postID}</p>
		</section>
	`;
}

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
