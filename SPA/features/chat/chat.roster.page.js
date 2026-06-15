import { WS_EVENTS } from '../../core/realtime/chat-socket.js';
import { fetchCurrentUserId } from './chat.conversation.api.js';
import { fetchRoster } from './chat.roster.api.js';
import { applySnapshot, moveToTopForMessage, setPresence } from './chat.roster.logic.js';
import { renderRosterList } from './chat.roster.views.js';

const ROSTER_BOUND_ATTR = 'data-roster-bound';
const ROSTER_ROOT_SELECTOR = '[data-chat-roster]';
const ROSTER_LIST_ID = 'roster-list';
const ROW_SELECTOR = '[data-roster-user-id]';

function clearSelection(rosterRoot) {
	if (typeof rosterRoot.querySelectorAll !== 'function') {
		return;
	}
	for (const node of rosterRoot.querySelectorAll('[aria-pressed="true"]')) {
		node.setAttribute('aria-pressed', 'false');
		node.classList?.remove('is-selected');
	}
}

function markSelected(row) {
	row.setAttribute('aria-pressed', 'true');
	row.classList?.add('is-selected');
}

function emitSelection(row) {
	const userId = Number(row.getAttribute('data-roster-user-id') ?? 0);
	const username = row.getAttribute('data-roster-username') ?? '';
	const isOnline = row.getAttribute('data-roster-online') === 'true';

	row.dispatchEvent(
		new CustomEvent('chat:user-selected', {
			detail: { userId, username, isOnline },
			bubbles: true,
		}),
	);

	return userId;
}

export function initChatRoster(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const fetchRef = options.fetchRef ?? (typeof fetch === 'function' ? fetch : null);

	if (
		!windowRef ||
		!documentRef ||
		typeof fetchRef !== 'function' ||
		typeof documentRef.querySelector !== 'function'
	) {
		return null;
	}

	const rosterRoot = documentRef.querySelector(ROSTER_ROOT_SELECTOR);
	if (!rosterRoot) {
		return null;
	}

	if (rosterRoot.getAttribute(ROSTER_BOUND_ATTR) === 'true') {
		return null;
	}

	rosterRoot.setAttribute(ROSTER_BOUND_ATTR, 'true');

	// Single source of truth for the painted roster. Presence and message events
	// mutate this through the pure helpers, then trigger a re-render.
	const state = { entries: [], selectedUserId: 0 };

	// The signed-in user id is needed to work out the "other party" of a live
	// dm.message so the right roster row floats to the top. Resolved lazily.
	let currentUserId = null;
	let currentUserPromise = null;
	function ensureCurrentUserId() {
		if (currentUserId !== null) {
			return Promise.resolve(currentUserId);
		}
		if (!currentUserPromise) {
			currentUserPromise = fetchCurrentUserId(fetchRef).then((id) => {
				currentUserId = id;
				return id;
			});
		}
		return currentUserPromise;
	}

	// Re-applies the active selection after innerHTML rebuilds the rows, looking
	// the row up by id since the previously selected element no longer exists.
	function reapplySelection() {
		clearSelection(rosterRoot);
		if (!state.selectedUserId) {
			return;
		}
		const selected = rosterRoot.querySelector?.(`[data-roster-user-id="${state.selectedUserId}"]`);
		if (selected) {
			markSelected(selected);
		}
	}

	function renderRoster() {
		const listMount = rosterRoot.querySelector?.(`#${ROSTER_LIST_ID}`);
		if (!listMount) {
			return;
		}
		listMount.innerHTML = renderRosterList(state.entries);
		reapplySelection();
	}

	function selectFromRow(row) {
		if (!row) {
			return;
		}
		clearSelection(rosterRoot);
		markSelected(row);
		state.selectedUserId = emitSelection(row);
	}

	rosterRoot.addEventListener('click', (event) => {
		selectFromRow(event.target?.closest?.(ROW_SELECTOR));
	});

	rosterRoot.addEventListener('keydown', (event) => {
		if (event.key !== 'Enter' && event.key !== ' ' && event.key !== 'Spacebar') {
			return;
		}
		const row = event.target?.closest?.(ROW_SELECTOR);
		if (!row) {
			return;
		}
		event.preventDefault?.();
		selectFromRow(row);
	});

	// --- Live WebSocket-driven updates (D04) ---------------------------------
	if (typeof documentRef.addEventListener === 'function') {
		documentRef.addEventListener(WS_EVENTS.PRESENCE_SNAPSHOT, (event) => {
			state.entries = applySnapshot(state.entries, event?.detail?.users);
			renderRoster();
		});

		documentRef.addEventListener(WS_EVENTS.PRESENCE_UPDATE, (event) => {
			const detail = event?.detail;
			if (!detail) {
				return;
			}
			state.entries = setPresence(state.entries, detail.userId, detail.isOnline);
			renderRoster();
		});

		documentRef.addEventListener(WS_EVENTS.DM_MESSAGE, (event) => {
			const message = event?.detail?.message;
			if (!message) {
				return;
			}
			void ensureCurrentUserId().then((me) => {
				const senderId = Number(message.sender_id ?? 0);
				const recipientId = Number(message.recipient_id ?? 0);
				const otherUserId = senderId === me ? recipientId : senderId;
				state.entries = moveToTopForMessage(state.entries, otherUserId, String(message.body ?? ''));
				renderRoster();
			});
		});
	}

	void (async () => {
		state.entries = await fetchRoster(fetchRef);
		renderRoster();
	})();

	return { rosterRoot };
}
