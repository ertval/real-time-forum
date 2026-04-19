import { renderActivityView } from '../../features/activity/activity.views.js';
import { renderLoginView, renderRegisterView } from '../../features/auth/auth.views.js';
import { renderFeedView } from '../../features/feed/feed.views.js';
import {
	renderCreatePostView,
	renderEditPostView,
	renderPostDetailView,
} from '../../features/post/post.views.js';
import { escapeHTML } from '../utils/html.js';

export function renderTemplate(match) {
	const { id, title } = match.route;

	if (id === 'login') {
		return renderLoginView();
	}

	if (id === 'register') {
		return renderRegisterView();
	}

	if (id === 'post-detail') {
		const postID = escapeHTML(match.params.id || '');
		return renderPostDetailView(postID);
	}

	if (id === 'edit-post') {
		const postID = escapeHTML(match.params.id || '');
		return renderEditPostView(postID);
	}

	if (id === 'create-post') {
		return renderCreatePostView();
	}

	if (id === 'activity') {
		return renderActivityView();
	}

	return renderFeedView(escapeHTML(title));
}
