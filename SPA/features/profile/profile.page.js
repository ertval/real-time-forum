// SPA/features/profile/profile.page.js

import { API_BASE } from '../../core/api/constants.js';
import { renderProfileAvatar, renderProfileContent } from './profile.views.js';

const PROFILE_BOUND_ATTR = 'data-profile-bound';

async function loadProfile(fetchRef, userID) {
	const response = await fetchRef(`${API_BASE}/users/${userID}/profile`, {
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});

	if (!response.ok) {
		return null;
	}

	const payload = await response.json();
	return payload.data ?? payload;
}

export function initProfilePage(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const fetchRef = options.fetchRef ?? (typeof fetch === 'function' ? fetch.bind(windowRef) : null);

	if (
		!windowRef ||
		!documentRef ||
		typeof fetchRef !== 'function' ||
		typeof documentRef.querySelector !== 'function'
	) {
		return null;
	}

	const root = documentRef.querySelector('[data-screen="profile"]');
	if (!root) {
		return null;
	}

	if (root.getAttribute(PROFILE_BOUND_ATTR) === 'true') {
		return null;
	}

	root.setAttribute(PROFILE_BOUND_ATTR, 'true');

	const userID = root.getAttribute('data-user-id');
	const contentEl = root.querySelector('#profile-content');
	const avatarEl = root.querySelector('.profile-avatar');

	if (!userID || !contentEl) {
		return null;
	}

	void (async () => {
		const profile = await loadProfile(fetchRef, userID);

		if (!profile) {
			contentEl.innerHTML = `
				<div class="profile-error">
					<h3>User not found</h3>
					<p>The profile you are looking for does not exist or has been removed.</p>
					<button class="auth-button" onclick="window.history.back()">Go Back</button>
				</div>
			`;
			contentEl.classList.remove('skeleton-content');
			if (avatarEl) avatarEl.classList.remove('skeleton');
			return;
		}

		contentEl.innerHTML = renderProfileContent(profile);
		contentEl.classList.remove('skeleton-content');

		if (avatarEl) {
			avatarEl.innerHTML = renderProfileAvatar(profile);
			avatarEl.classList.remove('skeleton');

			// Optional: Generate a color based on username for the avatar
			const colors = [
				'#FF6B6B',
				'#4ECDC4',
				'#45B7D1',
				'#96CEB4',
				'#FFEEAD',
				'#D4A5A5',
				'#9B59B6',
				'#3498DB',
			];
			const colorIndex = profile.username.charCodeAt(0) % colors.length;
			avatarEl.style.background = `linear-gradient(135deg, ${colors[colorIndex]}, ${colors[(colorIndex + 1) % colors.length]})`;
		}
	})();

	return { root };
}
