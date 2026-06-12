// SPA/features/chat/chat.roster.logic.js
//
// Pure state transitions for the live roster. Each helper takes the current
// entries array and returns a NEW array (immutable updates via array-by-copy),
// so the page layer can diff/re-render without worrying about shared mutation.

function entryId(entry) {
	return Number(entry?.user_id ?? 0);
}

// presence.snapshot lists every currently-online user. Anyone absent from the
// set is offline. Returns a copy with is_online reconciled against the snapshot.
export function applySnapshot(entries, users) {
	if (!Array.isArray(entries)) {
		return [];
	}

	const onlineIds = new Set(
		(Array.isArray(users) ? users : [])
			.filter((user) => Boolean(user?.is_online))
			.map((user) => Number(user?.user_id ?? 0)),
	);

	return entries.map((entry) => ({ ...entry, is_online: onlineIds.has(entryId(entry)) }));
}

// presence.update flips a single user's online state. Unknown users (not in the
// roster) are left untouched — the roster is sourced from the C05 API.
export function setPresence(entries, userId, isOnline) {
	if (!Array.isArray(entries)) {
		return [];
	}

	const target = Number(userId ?? 0);
	return entries.map((entry) =>
		entryId(entry) === target ? { ...entry, is_online: Boolean(isOnline) } : entry,
	);
}

// Message activity (sent or received) moves the conversation partner to the top
// of the roster and refreshes their last-message preview — mirroring the
// backend C05 ordering (most-recent-message first) without a refetch.
export function moveToTopForMessage(entries, otherUserId, preview) {
	if (!Array.isArray(entries)) {
		return [];
	}

	const target = Number(otherUserId ?? 0);
	const index = entries.findIndex((entry) => entryId(entry) === target);
	if (index === -1) {
		return entries.slice();
	}

	const updated = {
		...entries[index],
		last_message_preview:
			typeof preview === 'string' && preview.length > 0
				? preview
				: entries[index].last_message_preview,
	};

	return [updated, ...entries.slice(0, index), ...entries.slice(index + 1)];
}
