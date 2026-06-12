import { describe, expect, test } from 'vitest';
import {
	applySnapshot,
	moveToTopForMessage,
	setPresence,
} from '../../../../features/chat/chat.roster.logic.js';

function roster() {
	return [
		{ user_id: 1, username: 'alpha', is_online: false, last_message_preview: 'hi' },
		{ user_id: 2, username: 'beta', is_online: true, last_message_preview: null },
		{ user_id: 3, username: 'gamma', is_online: false, last_message_preview: 'yo' },
	];
}

describe('applySnapshot', () => {
	test('marks is_online true only for users present and online in the snapshot', () => {
		const result = applySnapshot(roster(), [
			{ user_id: 1, is_online: true },
			{ user_id: 2, is_online: false },
		]);

		expect(result.map((e) => [e.user_id, e.is_online])).toEqual([
			[1, true],
			[2, false],
			[3, false],
		]);
	});

	test('returns a new array and does not mutate the input', () => {
		const entries = roster();
		const result = applySnapshot(entries, [{ user_id: 2, is_online: true }]);

		expect(result).not.toBe(entries);
		expect(result[0]).not.toBe(entries[0]);
		expect(entries[1].is_online).toBe(true); // unchanged
	});

	test('non-array users argument marks everyone offline', () => {
		const result = applySnapshot(roster(), undefined);
		expect(result.every((e) => e.is_online === false)).toBe(true);
	});

	test('non-array entries returns []', () => {
		expect(applySnapshot(null, [])).toEqual([]);
	});
});

describe('setPresence', () => {
	test('flips a single known user without touching others', () => {
		const result = setPresence(roster(), 1, true);
		expect(result[0].is_online).toBe(true);
		expect(result[1].is_online).toBe(true);
		expect(result[2].is_online).toBe(false);
	});

	test('ignores unknown users', () => {
		const entries = roster();
		const result = setPresence(entries, 999, true);
		expect(result.map((e) => e.is_online)).toEqual([false, true, false]);
	});

	test('returns a new array and does not mutate the input', () => {
		const entries = roster();
		const result = setPresence(entries, 2, false);
		expect(result).not.toBe(entries);
		expect(entries[1].is_online).toBe(true);
		expect(result[1].is_online).toBe(false);
	});

	test('non-array entries returns []', () => {
		expect(setPresence(undefined, 1, true)).toEqual([]);
	});
});

describe('moveToTopForMessage', () => {
	test('moves the matching entry to the front and updates the preview', () => {
		const result = moveToTopForMessage(roster(), 3, 'new preview');
		expect(result.map((e) => e.user_id)).toEqual([3, 1, 2]);
		expect(result[0].last_message_preview).toBe('new preview');
	});

	test('keeps the existing preview when the new preview is empty', () => {
		const result = moveToTopForMessage(roster(), 3, '');
		expect(result[0].user_id).toBe(3);
		expect(result[0].last_message_preview).toBe('yo');
	});

	test('unknown user returns a no-op copy in original order', () => {
		const entries = roster();
		const result = moveToTopForMessage(entries, 999, 'x');
		expect(result).not.toBe(entries);
		expect(result.map((e) => e.user_id)).toEqual([1, 2, 3]);
	});

	test('does not mutate the input array', () => {
		const entries = roster();
		moveToTopForMessage(entries, 3, 'changed');
		expect(entries.map((e) => e.user_id)).toEqual([1, 2, 3]);
		expect(entries[2].last_message_preview).toBe('yo');
	});

	test('non-array entries returns []', () => {
		expect(moveToTopForMessage(null, 1, 'x')).toEqual([]);
	});
});
