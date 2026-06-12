import { escapeHTML } from '../../core/utils/html.js';

export function formatTimestamp(value) {
	if (!value) {
		return '';
	}

	const date = new Date(value);
	if (Number.isNaN(date.getTime())) {
		return String(value);
	}

	return date.toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' });
}

export function renderMessage(message, currentUserId) {
	const senderId = Number(message?.sender_id ?? 0);
	const isOwn = Number.isFinite(currentUserId) && senderId === currentUserId;
	const username = String(message?.sender_username ?? '');
	const body = String(message?.body ?? '');
	const rawCreatedAt = message?.created_at ?? '';
	const messageId = Number(message?.id ?? 0);

	const itemClass = isOwn
		? 'chat-conversation__message chat-conversation__message--own'
		: 'chat-conversation__message chat-conversation__message--incoming';

	return `
		<li class="${itemClass}" data-message-id="${messageId}">
			<div class="chat-conversation__meta">
				<span class="chat-conversation__sender">${escapeHTML(username)}</span>
				<time class="chat-conversation__time" datetime="${escapeHTML(rawCreatedAt)}">${escapeHTML(formatTimestamp(rawCreatedAt))}</time>
			</div>
			<p class="chat-conversation__body">${escapeHTML(body)}</p>
		</li>
	`;
}

export function renderMessageList(messages, currentUserId) {
	if (!Array.isArray(messages) || messages.length === 0) {
		return '<p class="chat-conversation__empty" data-conversation-empty>No messages yet. Say hello!</p>';
	}

	const items = messages.map((message) => renderMessage(message, currentUserId)).join('');
	return `<ul class="chat-conversation__messages" data-conversation-messages>${items}</ul>`;
}

export function renderComposer(isOnline) {
	const disabledAttr = isOnline ? '' : 'disabled';
	const placeholder = isOnline ? 'Type a message…' : 'User is offline';
	const offlineNote = isOnline
		? ''
		: '<p class="chat-conversation__offline-note" data-conversation-offline>You cannot message users while they are offline.</p>';

	return `
		<form class="chat-conversation__composer" data-conversation-composer>
			<input
				class="chat-conversation__input"
				data-conversation-input
				type="text"
				name="message"
				autocomplete="off"
				placeholder="${placeholder}"
				${disabledAttr} />
			<button class="chat-conversation__send" data-conversation-send type="submit" ${disabledAttr}>Send</button>
		</form>
		${offlineNote}
	`;
}

export function renderConversation({ username, isOnline, messages, currentUserId } = {}) {
	const safeUsername = String(username ?? '');
	const online = Boolean(isOnline);
	const presenceClass = online
		? 'chat-conversation__presence chat-conversation__presence--online'
		: 'chat-conversation__presence chat-conversation__presence--offline';

	return `
		<div class="chat-conversation" data-conversation-view>
			<header class="chat-conversation__header">
				<span class="chat-conversation__title">${escapeHTML(safeUsername)}</span>
				<span class="${presenceClass}">${online ? 'online' : 'offline'}</span>
			</header>
			<p class="chat-conversation__error" data-conversation-error role="alert" hidden></p>
			<div class="chat-conversation__scroll" data-conversation-scroll>
				${renderMessageList(messages, currentUserId)}
			</div>
			${renderComposer(online)}
		</div>
	`;
}
