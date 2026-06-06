import { describe, expect, test } from 'vitest';
import { renderRosterItem, renderRosterList } from '../../../../features/chat/chat.roster.views.js';

function countOccurrences(haystack, needle) {
	if (!needle) {
		return 0;
	}
	let count = 0;
	let index = haystack.indexOf(needle);
	while (index !== -1) {
		count += 1;
		index = haystack.indexOf(needle, index + needle.length);
	}
	return count;
}

describe('renderRosterList', () => {
	test('renders one <li> per entry, in input order', () => {
		const entries = [
			{ user_id: 1, username: 'alpha', is_online: true, last_message_preview: 'hi' },
			{ user_id: 2, username: 'beta', is_online: false, last_message_preview: null },
			{ user_id: 3, username: 'gamma', is_online: true, last_message_preview: null },
		];

		const html = renderRosterList(entries);
		expect(countOccurrences(html, '<li ')).toBe(3);

		const alphaIdx = html.indexOf('data-roster-user-id="1"');
		const betaIdx = html.indexOf('data-roster-user-id="2"');
		const gammaIdx = html.indexOf('data-roster-user-id="3"');

		expect(alphaIdx).toBeGreaterThan(-1);
		expect(betaIdx).toBeGreaterThan(alphaIdx);
		expect(gammaIdx).toBeGreaterThan(betaIdx);
	});

	test('online entries get the online presence class and data-roster-online="true"', () => {
		const html = renderRosterItem({
			user_id: 5,
			username: 'maria',
			is_online: true,
			last_message_preview: null,
		});

		expect(html).toContain('data-roster-online="true"');
		expect(html).toContain('chat-roster__presence--online');
		expect(html).toContain('chat-roster__item--online');
		expect(html).toContain('aria-label="online"');
	});

	test('offline entries get the offline presence class and data-roster-online="false"', () => {
		const html = renderRosterItem({
			user_id: 9,
			username: 'sam',
			is_online: false,
			last_message_preview: null,
		});

		expect(html).toContain('data-roster-online="false"');
		expect(html).toContain('chat-roster__presence--offline');
		expect(html).toContain('chat-roster__item--offline');
		expect(html).toContain('aria-label="offline"');
	});

	test('renders no preview span when last_message_preview is null', () => {
		const html = renderRosterItem({
			user_id: 1,
			username: 'a',
			is_online: false,
			last_message_preview: null,
		});

		expect(html).not.toContain('chat-roster__preview');
	});

	test('renders the preview span when last_message_preview is a non-empty string', () => {
		const html = renderRosterItem({
			user_id: 1,
			username: 'a',
			is_online: false,
			last_message_preview: 'see you soon',
		});

		expect(html).toContain('chat-roster__preview');
		expect(html).toContain('see you soon');
	});

	test('empty entries array renders the empty-state paragraph', () => {
		const html = renderRosterList([]);
		expect(html).toContain('chat-panel__empty');
		expect(html).toContain('No other users yet');
		expect(html).not.toContain('<li');
	});

	test('non-array input renders the empty-state paragraph', () => {
		expect(renderRosterList(null)).toContain('chat-panel__empty');
		expect(renderRosterList(undefined)).toContain('chat-panel__empty');
	});

	test('username is HTML-escaped', () => {
		const html = renderRosterItem({
			user_id: 1,
			username: '<img onerror=alert(1)>',
			is_online: true,
			last_message_preview: null,
		});

		expect(html).not.toContain('<img onerror=alert(1)>');
		expect(html).toContain('&lt;img onerror=alert(1)&gt;');
	});

	test('preview text is HTML-escaped', () => {
		const html = renderRosterItem({
			user_id: 1,
			username: 'safe',
			is_online: true,
			last_message_preview: '<script>x</script>',
		});

		expect(html).not.toContain('<script>x</script>');
		expect(html).toContain('&lt;script&gt;x&lt;/script&gt;');
	});
});
