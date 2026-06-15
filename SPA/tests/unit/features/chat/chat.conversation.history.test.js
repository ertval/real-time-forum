import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';
import { initChatConversation } from '../../../../features/chat/chat.conversation.page.js';

async function flushMicrotasks() {
	for (let i = 0; i < 10; i += 1) {
		await Promise.resolve();
	}
}

function createScrollEl({ scrollHeight = 1000, scrollTop = 0 } = {}) {
	const listeners = new Map();
	return {
		scrollHeight,
		scrollTop,
		addEventListener(type, handler) {
			if (!listeners.has(type)) {
				listeners.set(type, []);
			}
			listeners.get(type).push(handler);
		},
		dispatch(type) {
			for (const handler of listeners.get(type) ?? []) {
				handler();
			}
		},
	};
}

// List mock: prepending grows the scroll container, like real prepended nodes.
function createListEl(scrollEl, growthPerPrepend = 300) {
	return {
		chunks: [],
		insertAdjacentHTML(position, html) {
			this.chunks.push({ position, html });
			scrollEl.scrollHeight += growthPerPrepend;
		},
	};
}

function createActiveRoot(scrollEl, listEl) {
	const attrs = new Map();
	const listeners = new Map();
	return {
		innerHTML: '',
		getAttribute(name) {
			return attrs.has(name) ? attrs.get(name) : null;
		},
		setAttribute(name, value) {
			attrs.set(name, String(value));
		},
		addEventListener(type, handler) {
			if (!listeners.has(type)) {
				listeners.set(type, []);
			}
			listeners.get(type).push(handler);
		},
		querySelector(selector) {
			if (selector === '[data-conversation-scroll]') {
				return scrollEl;
			}
			if (selector === '[data-conversation-messages]') {
				return listEl;
			}
			return null;
		},
	};
}

function createDocumentRef(activeRoot) {
	const listeners = new Map();
	return {
		querySelector(selector) {
			return selector === '[data-chat-active]' ? activeRoot : null;
		},
		addEventListener(type, handler) {
			if (!listeners.has(type)) {
				listeners.set(type, []);
			}
			listeners.get(type).push(handler);
		},
		dispatch(type, event) {
			for (const handler of listeners.get(type) ?? []) {
				handler(event);
			}
		},
	};
}

function msg(id) {
	return {
		id,
		sender_id: 2,
		sender_username: 'bob',
		body: `m${id}`,
		created_at: '2026-01-01T00:00:00Z',
	};
}

function makePagingFetch({ meId = 1, initial, older }) {
	return vi.fn(async (url) => {
		const u = String(url);
		if (u.endsWith('/users/me')) {
			return { ok: true, status: 200, json: async () => ({ data: { id: meId } }) };
		}
		if (u.includes('before_id=')) {
			return {
				ok: true,
				status: 200,
				json: async () => ({ data: { messages: older.messages, has_more: older.hasMore } }),
			};
		}
		return {
			ok: true,
			status: 200,
			json: async () => ({ data: { messages: initial.messages, has_more: initial.hasMore } }),
		};
	});
}

async function selectUser(documentRef, { userId = 2, username = 'bob', isOnline = true } = {}) {
	documentRef.dispatch('chat:user-selected', { detail: { userId, username, isOnline } });
	await flushMicrotasks();
}

describe('incremental history loading', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	test('scrolling to the top loads the next batch of 10 with before_id', async () => {
		const scrollEl = createScrollEl({ scrollTop: 0 });
		const listEl = createListEl(scrollEl);
		const activeRoot = createActiveRoot(scrollEl, listEl);
		const documentRef = createDocumentRef(activeRoot);

		const initial = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 11)), hasMore: true };
		const older = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 1)), hasMore: false };
		const fetchRef = makePagingFetch({ initial, older });

		initChatConversation({ windowRef: {}, documentRef, fetchRef });
		await selectUser(documentRef);

		scrollEl.dispatch('scroll');
		await flushMicrotasks();

		// Oldest rendered id was 11, so the older page must be requested with before_id=11.
		const olderCall = fetchRef.mock.calls.find(([url]) => String(url).includes('before_id='));
		expect(olderCall?.[0]).toBe('/api/v1/chats/2/messages?before_id=11');
		expect(listEl.chunks).toHaveLength(1);
		expect(listEl.chunks[0].position).toBe('afterbegin');
		// 10 prepended items.
		expect(listEl.chunks[0].html.match(/data-message-id/g)).toHaveLength(10);
	});

	test('repeated scroll events do not cause burst requests', async () => {
		const scrollEl = createScrollEl({ scrollTop: 0 });
		const listEl = createListEl(scrollEl);
		const activeRoot = createActiveRoot(scrollEl, listEl);
		const documentRef = createDocumentRef(activeRoot);

		const initial = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 11)), hasMore: true };
		const older = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 1)), hasMore: true };
		const fetchRef = makePagingFetch({ initial, older });

		initChatConversation({ windowRef: {}, documentRef, fetchRef });
		await selectUser(documentRef);

		for (let i = 0; i < 25; i += 1) {
			scrollEl.dispatch('scroll');
		}
		await flushMicrotasks();

		const olderCalls = fetchRef.mock.calls.filter(([url]) => String(url).includes('before_id='));
		expect(olderCalls).toHaveLength(1);
	});

	test('preserves scroll position when older history is prepended', async () => {
		const scrollEl = createScrollEl({ scrollHeight: 1000, scrollTop: 5 });
		const listEl = createListEl(scrollEl, 300);
		const activeRoot = createActiveRoot(scrollEl, listEl);
		const documentRef = createDocumentRef(activeRoot);

		const initial = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 11)), hasMore: true };
		const older = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 1)), hasMore: false };
		const fetchRef = makePagingFetch({ initial, older });

		initChatConversation({ windowRef: {}, documentRef, fetchRef });
		await selectUser(documentRef);

		scrollEl.dispatch('scroll');
		await flushMicrotasks();

		// previousTop(5) + (newHeight 1300 - previousHeight 1000) = 305
		expect(scrollEl.scrollTop).toBe(305);
	});

	test('does not request older history when the backend reports none', async () => {
		const scrollEl = createScrollEl({ scrollTop: 0 });
		const listEl = createListEl(scrollEl);
		const activeRoot = createActiveRoot(scrollEl, listEl);
		const documentRef = createDocumentRef(activeRoot);

		const initial = { messages: [msg(1), msg(2)], hasMore: false };
		const older = { messages: [], hasMore: false };
		const fetchRef = makePagingFetch({ initial, older });

		initChatConversation({ windowRef: {}, documentRef, fetchRef });
		await selectUser(documentRef);

		scrollEl.dispatch('scroll');
		await flushMicrotasks();

		const olderCalls = fetchRef.mock.calls.filter(([url]) => String(url).includes('before_id='));
		expect(olderCalls).toHaveLength(0);
	});

	test('does not load when the viewport is not near the top', async () => {
		const scrollEl = createScrollEl({ scrollTop: 500 });
		const listEl = createListEl(scrollEl);
		const activeRoot = createActiveRoot(scrollEl, listEl);
		const documentRef = createDocumentRef(activeRoot);

		const initial = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 11)), hasMore: true };
		const older = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 1)), hasMore: true };
		const fetchRef = makePagingFetch({ initial, older });

		initChatConversation({ windowRef: {}, documentRef, fetchRef });
		await selectUser(documentRef);

		scrollEl.dispatch('scroll');
		await flushMicrotasks();

		const olderCalls = fetchRef.mock.calls.filter(([url]) => String(url).includes('before_id='));
		expect(olderCalls).toHaveLength(0);
	});

	test('stops paging once a batch reports has_more=false', async () => {
		const scrollEl = createScrollEl({ scrollTop: 0 });
		const listEl = createListEl(scrollEl);
		const activeRoot = createActiveRoot(scrollEl, listEl);
		const documentRef = createDocumentRef(activeRoot);

		const initial = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 11)), hasMore: true };
		const older = { messages: Array.from({ length: 10 }, (_, i) => msg(i + 1)), hasMore: false };
		const fetchRef = makePagingFetch({ initial, older });

		initChatConversation({ windowRef: {}, documentRef, fetchRef });
		await selectUser(documentRef);

		// First scroll consumes the only older batch (has_more=false afterwards).
		scrollEl.dispatch('scroll');
		await flushMicrotasks();
		// Let the throttle cooldown lapse so a later scroll is a fresh leading call.
		vi.advanceTimersByTime(500);
		await flushMicrotasks();

		scrollEl.dispatch('scroll');
		await flushMicrotasks();

		const olderCalls = fetchRef.mock.calls.filter(([url]) => String(url).includes('before_id='));
		expect(olderCalls).toHaveLength(1);
	});
});
