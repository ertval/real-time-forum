// SPA/core/realtime/chat-socket.js
//
// D04 — browser WebSocket transport for the realtime chat.
//
// The socket is a thin, decoupled bridge between the backend WS contract
// (SDS § 5.5) and the rest of the SPA. It does NOT touch the DOM directly:
// every inbound server frame is re-published as a DOM CustomEvent on the
// shared documentRef, so feature slices (roster, conversation) subscribe
// without holding a reference to the socket itself. Outbound frames go the
// other way through `send(type, payload)`.
//
// Reconnect/backoff/disconnected-state recovery are intentionally out of
// scope for D04 (see track-d.md verification gate).

// Inbound server event type -> DOM CustomEvent name. Keeping the mapping in
// one place is the single source of truth shared with the subscribers.
export const WS_EVENTS = Object.freeze({
	PRESENCE_SNAPSHOT: 'chat:presence-snapshot',
	PRESENCE_UPDATE: 'chat:presence-update',
	DM_MESSAGE: 'chat:dm-message',
	ERROR: 'chat:error',
});

// Builds the absolute ws:// or wss:// URL for the same-origin /ws endpoint,
// which the frontend server proxies to the backend (D09). Falls back to ws://
// when location data is unavailable (e.g. during tests).
export function resolveSocketURL(locationRef) {
	const protocol = locationRef?.protocol === 'https:' ? 'wss:' : 'ws:';
	const host = locationRef?.host ?? '';
	return `${protocol}//${host}/ws`;
}

function dispatchEvent(documentRef, name, detail) {
	if (typeof documentRef?.dispatchEvent !== 'function' || typeof CustomEvent !== 'function') {
		return;
	}
	documentRef.dispatchEvent(new CustomEvent(name, { detail }));
}

// Translates a parsed server frame into its corresponding DOM CustomEvent.
// Unknown frame types are ignored — a forward-compatible client must not
// crash on event types it does not understand.
function publishServerEvent(documentRef, frame) {
	const payload = frame?.payload ?? {};

	switch (frame?.type) {
		case 'presence.snapshot':
			dispatchEvent(documentRef, WS_EVENTS.PRESENCE_SNAPSHOT, {
				users: Array.isArray(payload.users) ? payload.users : [],
			});
			return;
		case 'presence.update':
			dispatchEvent(documentRef, WS_EVENTS.PRESENCE_UPDATE, {
				userId: Number(payload.user_id ?? 0),
				isOnline: Boolean(payload.is_online),
			});
			return;
		case 'dm.message':
			dispatchEvent(documentRef, WS_EVENTS.DM_MESSAGE, { message: payload });
			return;
		case 'chat.error':
			dispatchEvent(documentRef, WS_EVENTS.ERROR, {
				code: String(payload.code ?? 'UNKNOWN'),
				message: String(payload.message ?? 'Something went wrong.'),
			});
			return;
		default:
	}
}

export function createChatSocket(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const SocketCtor = options.socketCtor ?? windowRef?.WebSocket ?? null;
	const url = options.url ?? resolveSocketURL(windowRef?.location);

	let socket = null;

	function isOpen() {
		return Boolean(socket) && socket.readyState === 1; // WebSocket.OPEN
	}

	function open() {
		if (typeof SocketCtor !== 'function' || socket) {
			return socket;
		}

		socket = new SocketCtor(url);

		socket.addEventListener?.('message', (event) => {
			let frame = null;
			try {
				frame = JSON.parse(event?.data ?? 'null');
			} catch {
				// Malformed frame: drop it. The connection stays usable.
				return;
			}
			publishServerEvent(documentRef, frame);
		});

		// A transport-level failure is surfaced to the UI as a chat.error so the
		// user is not left silently staring at a dead composer.
		socket.addEventListener?.('error', () => {
			dispatchEvent(documentRef, WS_EVENTS.ERROR, {
				code: 'CONNECTION_ERROR',
				message: 'Connection problem. Some messages may not have been delivered.',
			});
		});

		return socket;
	}

	// Emits a client frame matching the SDS § 5.5 envelope. Returns false when
	// the socket is not open so callers can surface a send failure.
	function send(type, payload) {
		if (!isOpen()) {
			return false;
		}
		socket.send(JSON.stringify({ type, payload }));
		return true;
	}

	function close() {
		if (!socket) {
			return;
		}
		socket.close?.();
		socket = null;
	}

	return {
		open,
		close,
		send,
		isOpen,
		// Resource Management: `using socket = createChatSocket(...)` auto-closes.
		[Symbol.dispose]() {
			close();
		},
	};
}
