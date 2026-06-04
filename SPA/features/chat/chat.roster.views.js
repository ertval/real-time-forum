import { escapeHTML } from '../../core/utils/html.js';

export function renderRosterItem(entry) {
	const userId = Number(entry?.user_id ?? 0);
	const username = String(entry?.username ?? '');
	const isOnline = Boolean(entry?.is_online);
	const presenceLabel = isOnline ? 'online' : 'offline';
	const presenceClass = isOnline
		? 'chat-roster__presence chat-roster__presence--online'
		: 'chat-roster__presence chat-roster__presence--offline';
	const itemClass = isOnline
		? 'chat-roster__item chat-roster__item--online'
		: 'chat-roster__item chat-roster__item--offline';

	const rawPreview = entry?.last_message_preview;
	const previewMarkup =
		typeof rawPreview === 'string' && rawPreview.length > 0
			? `<span class="chat-roster__preview">${escapeHTML(rawPreview)}</span>`
			: '';

	return `
		<li class="${itemClass}"
			data-roster-user-id="${userId}"
			data-roster-username="${escapeHTML(username)}"
			data-roster-online="${isOnline ? 'true' : 'false'}"
			tabindex="0"
			role="button"
			aria-pressed="false">
			<span class="${presenceClass}" aria-label="${presenceLabel}"></span>
			<span class="chat-roster__name">${escapeHTML(username)}</span>
			${previewMarkup}
		</li>
	`;
}

export function renderRosterList(entries) {
	if (!Array.isArray(entries) || entries.length === 0) {
		return '<p class="chat-panel__empty">No other users yet</p>';
	}

	const items = entries.map((entry) => renderRosterItem(entry)).join('');
	return `<ul class="chat-roster" data-roster-list>${items}</ul>`;
}
