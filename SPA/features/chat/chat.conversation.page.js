import { fetchConversation, fetchCurrentUserId } from './chat.conversation.api.js';
import { renderConversation } from './chat.conversation.views.js';

const ACTIVE_ROOT_SELECTOR = '[data-chat-active]';
const ACTIVE_BOUND_ATTR = 'data-conversation-bound';
const COMPOSER_SELECTOR = '[data-conversation-composer]';

function renderInto(activeRoot, data) {
	activeRoot.innerHTML = renderConversation(data);
}

export function initChatConversation(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const fetchRef = options.fetchRef ?? (typeof fetch === 'function' ? fetch : null);

	if (
		!windowRef ||
		!documentRef ||
		typeof fetchRef !== 'function' ||
		typeof documentRef.querySelector !== 'function' ||
		typeof documentRef.addEventListener !== 'function'
	) {
		return null;
	}

	const activeRoot = documentRef.querySelector(ACTIVE_ROOT_SELECTOR);
	if (!activeRoot) {
		return null;
	}

	if (activeRoot.getAttribute(ACTIVE_BOUND_ATTR) === 'true') {
		return null;
	}

	activeRoot.setAttribute(ACTIVE_BOUND_ATTR, 'true');

	// Resolve the signed-in user once so own/incoming messages can be styled apart.
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

	async function onUserSelected(event) {
		const detail = event?.detail;
		if (!detail || !Number.isFinite(detail.userId) || detail.userId <= 0) {
			return;
		}

		const [resolvedId, conversation] = await Promise.all([
			ensureCurrentUserId(),
			fetchConversation(fetchRef, detail.userId),
		]);

		renderInto(activeRoot, {
			username: detail.username,
			isOnline: Boolean(detail.isOnline),
			messages: conversation.messages,
			currentUserId: resolvedId,
		});
	}

	documentRef.addEventListener('chat:user-selected', (event) => {
		void onUserSelected(event);
	});

	// Composer submit is wired to the WebSocket in D04; here we only stop the
	// default form navigation. Offline users are already blocked via disabled inputs.
	activeRoot.addEventListener('submit', (event) => {
		const composer = event.target?.closest?.(COMPOSER_SELECTOR);
		if (!composer) {
			return;
		}
		event.preventDefault?.();
	});

	return { activeRoot };
}
