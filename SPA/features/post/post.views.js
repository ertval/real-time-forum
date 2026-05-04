// SPA/features/post/post.views.js

export function renderPostDetailView(postID) {
	return `
		<section data-screen="post-detail" data-post-id="${postID}">
			<div class="post-detail-container">
				<div id="post-content" class="post-content">
					<p>Loading post details...</p>
				</div>
				<div id="comments-section" class="comments-section">
					<h3>Comments</h3>
					<div id="comments-list">
						<p>Loading comments...</p>
					</div>
				</div>
			</div>
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
