import { WS_EVENTS } from '../../core/realtime/chat-socket.js';
import { throttle } from '../../core/utils/throttle.js';
import { fetchConversation, fetchCurrentUserId } from './chat.conversation.api.js';
import {
	renderConversation,
	renderMessage,
	renderMessageItems,
	renderMessageList,
} from './chat.conversation.views.js';

const ACTIVE_ROOT_SELECTOR = '[data-chat-active]';
const ACTIVE_BOUND_ATTR = 'data-conversation-bound';
const COMPOSER_SELECTOR = '[data-conversation-composer]';
const INPUT_SELECTOR = '[data-conversation-input]';
const SCROLL_CONTAINER_SELECTOR = '[data-conversation-scroll]';
const MESSAGE_LIST_SELECTOR = '[data-conversation-messages]';
const ERROR_SELECTOR = '[data-conversation-error]';

// Outbound composer submissions are published as this DOM event; the app shell
// owns the socket and forwards it as a dm.send frame. Keeps the conversation
// view decoupled from the transport.
export const SEND_MESSAGE_EVENT = 'chat:send-message';

// Trigger an older-history load once the viewport is within this many pixels
// of the top, and rate-limit scroll handling to avoid burst requests.
const SCROLL_LOAD_THRESHOLD_PX = 80;
const SCROLL_THROTTLE_MS = 200;

function renderInto(activeRoot, data) {
	activeRoot.innerHTML = renderConversation(data);
}

function oldestMessageId(messages) {
	if (!Array.isArray(messages) || messages.length === 0) {
		return null;
	}
	const id = Number(messages[0]?.id);
	return Number.isFinite(id) && id > 0 ? id : null;
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

	// Tracks the conversation currently on screen so live dm.message frames can
	// be matched to the active thread.
	const state = { recipientId: 0, username: '', isOnline: false };

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

	// One state object per active conversation. Replacing it on each selection
	// invalidates any older-history fetch still in flight for the prior user.
	let conversationState = null;

	async function loadOlderHistory(scrollEl, listEl) {
		const state = conversationState;
		if (!state || state.isLoading || !state.hasMore || !state.oldestId) {
			return;
		}

		state.isLoading = true;
		// Capture geometry before the prepend so we can restore the viewport.
		const previousHeight = scrollEl.scrollHeight ?? 0;
		const previousTop = scrollEl.scrollTop ?? 0;

		const older = await fetchConversation(fetchRef, state.userId, { beforeId: state.oldestId });

		// Bail if the user switched conversations while the request was pending.
		if (conversationState !== state) {
			return;
		}

		if (older.messages.length > 0) {
			listEl.insertAdjacentHTML?.('afterbegin', renderMessageItems(older.messages, currentUserId));
			state.oldestId = oldestMessageId(older.messages) ?? state.oldestId;

			// Keep the previously-visible message anchored: the list grew above
			// the viewport, so push scrollTop down by exactly that delta.
			const newHeight = scrollEl.scrollHeight ?? previousHeight;
			scrollEl.scrollTop = previousTop + (newHeight - previousHeight);
			state.hasMore = older.hasMore;
		} else {
			state.hasMore = false;
		}

		state.isLoading = false;
	}

	function bindIncrementalLoading() {
		if (typeof activeRoot.querySelector !== 'function') {
			return;
		}

		const scrollEl = activeRoot.querySelector(SCROLL_CONTAINER_SELECTOR);
		const listEl = activeRoot.querySelector(MESSAGE_LIST_SELECTOR);
		if (!scrollEl || !listEl || typeof scrollEl.addEventListener !== 'function') {
			return;
		}

		// innerHTML was just replaced, so this is a fresh node with no prior
		// listeners — a throttled handler can be attached without leaking.
		const onScroll = throttle(() => {
			if ((scrollEl.scrollTop ?? 0) <= SCROLL_LOAD_THRESHOLD_PX) {
				void loadOlderHistory(scrollEl, listEl);
			}
		}, SCROLL_THROTTLE_MS);

		scrollEl.addEventListener('scroll', onScroll);
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

		state.recipientId = detail.userId;
		state.username = detail.username;
		state.isOnline = Boolean(detail.isOnline);

		renderInto(activeRoot, {
			username: detail.username,
			isOnline: state.isOnline,
			messages: conversation.messages,
			currentUserId: resolvedId,
		});

		conversationState = {
			userId: detail.userId,
			oldestId: oldestMessageId(conversation.messages),
			hasMore: Boolean(conversation.hasMore),
			isLoading: false,
		};

		bindIncrementalLoading();
	}

	// Appends a single live message to the open thread, replacing the empty
	// state if necessary, and keeps the viewport pinned to the latest message.
	function appendLiveMessage(message, me) {
		const scroll = activeRoot.querySelector?.(SCROLL_CONTAINER_SELECTOR);
		if (!scroll) {
			return;
		}

		const list = scroll.querySelector?.(MESSAGE_LIST_SELECTOR);
		if (list && typeof list.insertAdjacentHTML === 'function') {
			list.insertAdjacentHTML('beforeend', renderMessage(message, me));
		} else {
			scroll.innerHTML = renderMessageList([message], me);
		}

		if (typeof scroll.scrollHeight === 'number') {
			scroll.scrollTop = scroll.scrollHeight;
		}
	}

	function showError(message) {
		const banner = activeRoot.querySelector?.(ERROR_SELECTOR);
		if (!banner) {
			return;
		}
		banner.textContent = message;
		banner.removeAttribute('hidden');
	}

	function clearError() {
		const banner = activeRoot.querySelector?.(ERROR_SELECTOR);
		banner?.setAttribute('hidden', '');
	}

	documentRef.addEventListener('chat:user-selected', (event) => {
		void onUserSelected(event);
	});

	// Incoming/echoed messages: render live only when they belong to the open
	// thread. The backend echoes the sender's own message back, so sent messages
	// also arrive here — no optimistic rendering needed.
	documentRef.addEventListener(WS_EVENTS.DM_MESSAGE, (event) => {
		const message = event?.detail?.message;
		if (!message || !state.recipientId) {
			return;
		}
		void ensureCurrentUserId().then((me) => {
			const senderId = Number(message.sender_id ?? 0);
			const recipientId = Number(message.recipient_id ?? 0);
			const otherUserId = senderId === me ? recipientId : senderId;
			if (otherUserId !== state.recipientId) {
				return;
			}
			clearError();
			appendLiveMessage(message, me);
		});
	});

	documentRef.addEventListener(WS_EVENTS.ERROR, (event) => {
		showError(event?.detail?.message ?? 'Something went wrong.');
	});

	// Composer submit emits the outbound message through the shell-owned socket.
	// Offline recipients are already blocked via the disabled input/button.
	activeRoot.addEventListener('submit', (event) => {
		const composer = event.target?.closest?.(COMPOSER_SELECTOR);
		if (!composer) {
			return;
		}
		event.preventDefault?.();

		const input = composer.querySelector?.(INPUT_SELECTOR);
		const body = String(input?.value ?? '').trim();
		if (!body || !state.recipientId) {
			return;
		}

		documentRef.dispatchEvent?.(
			new CustomEvent(SEND_MESSAGE_EVENT, {
				detail: { recipientId: state.recipientId, body },
				bubbles: true,
			}),
		);

		if (input) {
			input.value = '';
		}
		clearError();
	});

	return { activeRoot };
}
