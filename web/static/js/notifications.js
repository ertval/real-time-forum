// /web/static/js/notifications.js

import { playNotification } from './sound-effects.js';
import { uiNotify } from './ui-messages.js';
import { API_BASE } from './utils.js';

const seenNotificationIds = new Set();
let pollInterval = null;
let initialized = false;

/* -------------------------
   START POLLING
-------------------------- */
export function startNotificationPolling() {
	if (pollInterval) return;

	const bell = document.getElementById('notification-bell');
	if (!bell) return; // guest user

	resetState();
	fetchNotifications();

	pollInterval = setInterval(fetchNotifications, 5000);
}

/* -------------------------
   STOP POLLING
-------------------------- */
export function stopNotificationPolling() {
	if (!pollInterval) return;

	clearInterval(pollInterval);
	pollInterval = null;
	resetState();
}

/* -------------------------
   RESET STATE
-------------------------- */
function resetState() {
	seenNotificationIds.clear();
	initialized = false;
}

/* -------------------------
   FETCH
-------------------------- */
async function fetchNotifications() {
	try {
		const res = await fetch(`${API_BASE}/notifications`, {
			credentials: 'include',
			headers: { Accept: 'application/json' },
		});

		// If session expired → stop polling
		if (res.status === 401) {
			stopNotificationPolling();
			return;
		}

		if (!res.ok) return;

		const json = await res.json();
		const notifications = json.data?.notifications || [];
		const unreadCount = json.data?.unread_count ?? 0;

		updateBadge(unreadCount);
		renderDropdown(notifications);
		processNewNotifications(notifications);

		initialized = true;
	} catch (err) {
		console.error('notifications error:', err);
	}
}

/* -------------------------
   BADGE
-------------------------- */
function updateBadge(count) {
	const badge = document.getElementById('notification-badge');
	if (!badge) return;

	if (count > 0) {
		badge.textContent = count;
		badge.classList.remove('hidden');
	} else {
		badge.textContent = '';
		badge.classList.add('hidden');
	}
}

/* -------------------------
   DROPDOWN RENDER
-------------------------- */
function renderDropdown(notifications) {
	const dropdown = document.getElementById('notification-dropdown');
	if (!dropdown) return;

	dropdown.innerHTML = `
    <div class="notification-header">
      <span>Notifications</span>
      <button id="mark-all-read-btn">Mark all as read</button>
    </div>
  `;

	dropdown.scrollTop = 0;

	const markAllBtn = document.getElementById('mark-all-read-btn');

	if (!notifications.length) {
		if (markAllBtn) markAllBtn.style.display = 'none';

		dropdown.innerHTML += `
      <div class="notification-empty">
        No notifications yet.
      </div>
    `;
		return;
	}

	const hasUnread = notifications.some((n) => !n.is_read);

	if (markAllBtn && !hasUnread) {
		markAllBtn.style.display = 'none';
	}

	notifications.forEach((n) => {
		const div = document.createElement('div');
		div.className = 'notification-item';

		if (!n.is_read) div.classList.add('unread');

		div.innerHTML = buildMessage(n);

		div.addEventListener('click', async () => {
			if (!n.is_read) {
				await markOneAsRead(n.id);

				div.classList.remove('unread');

				const badge = document.getElementById('notification-badge');

				if (badge && badge.textContent !== '') {
					let count = Number(badge.textContent) || 0;
					count = Math.max(count - 1, 0);
					updateBadge(count);
				}
			}

			if (n.post_id) {
				if (n.type === 'comment') {
					window.location.href = `/view-post/${n.post_id}?highlight=last`;
					return;
				}

				if (n.comment_id) {
					window.location.href = `/view-post/${n.post_id}?highlight=${n.comment_id}`;
					return;
				}

				window.location.href = `/view-post/${n.post_id}`;
			}
		});

		dropdown.appendChild(div);
	});

	if (markAllBtn) {
		markAllBtn.addEventListener('click', async (e) => {
			e.stopPropagation();

			await markAllAsRead();

			document.querySelectorAll('.notification-item.unread').forEach((el) => {
				el.classList.remove('unread');
			});

			updateBadge(0);
		});
	}
}

/* -------------------------
   MARK ONE AS READ
-------------------------- */
async function markOneAsRead(id) {
	try {
		await fetch(`${API_BASE}/notifications/${id}/read`, {
			method: 'PATCH',
			credentials: 'include',
		});
	} catch (err) {
		console.error('mark-one error:', err);
	}
}

/* -------------------------
   MARK ALL AS READ
-------------------------- */
async function markAllAsRead() {
	try {
		const res = await fetch(`${API_BASE}/notifications/read-all`, {
			method: 'PATCH',
			credentials: 'include',
		});

		if (res.status === 401) return;
	} catch (err) {
		console.error('mark-all error:', err);
	}
}

/* -------------------------
   TOAST + SOUND
-------------------------- */
function processNewNotifications(notifications) {
	notifications.forEach((n) => {
		if (!initialized) {
			seenNotificationIds.add(n.id);
			return;
		}

		if (!seenNotificationIds.has(n.id)) {
			seenNotificationIds.add(n.id);

			if (!n.is_read) {
				uiNotify(buildMessage(n), { type: 'info', html: true });
				playNotification();
			}
		}
	});
}

/* -------------------------
   MESSAGE BUILDER
-------------------------- */
function buildMessage(n) {
	const truncate = (str, len = 20) => {
		if (!str) return '';
		return str.length > len ? `${str.slice(0, len)}…` : str;
	};

	const title = `<strong>${truncate(n.post_title)}</strong>`;
	const excerpt = `<em>${truncate(n.comment_excerpt)}</em>`;

	switch (n.type) {
		case 'post_like':
			return `${n.actor_username} liked your post: ${title} 👍`;

		case 'post_dislike':
			return `${n.actor_username} disliked your post: ${title} 👎`;

		case 'comment':
			return `${n.actor_username} commented ${excerpt} on ${title} 💬`;

		case 'comment_like':
			return `${n.actor_username} liked your comment: ${excerpt} 👍`;

		case 'comment_dislike':
			return `${n.actor_username} disliked your comment: ${excerpt} 👎`;

		default:
			return 'New notification 🔔';
	}
}

/* -------------------------
   BELL CLICK HANDLER
-------------------------- */
export function initNotificationBell() {
	const bell = document.getElementById('notification-bell');
	const dropdown = document.getElementById('notification-dropdown');

	if (!bell || !dropdown) return;

	bell.addEventListener('click', (e) => {
		e.stopPropagation();
		dropdown.classList.toggle('hidden');
	});

	document.addEventListener('click', (e) => {
		if (!dropdown.contains(e.target) && !bell.contains(e.target)) {
			dropdown.classList.add('hidden');
		}
	});
}
