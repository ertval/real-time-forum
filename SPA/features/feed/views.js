export function renderFeedView(title) {
	return `
		<section data-screen="feed" aria-labelledby="screen-feed-title">
			<h1 id="screen-feed-title">${title}</h1>
			<p>Browse the latest discussions.</p>
		</section>
	`;
}
