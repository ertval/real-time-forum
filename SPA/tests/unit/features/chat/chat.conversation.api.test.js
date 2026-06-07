import { describe, expect, test, vi } from 'vitest';
import {
	fetchConversation,
	fetchCurrentUserId,
} from '../../../../features/chat/chat.conversation.api.js';

describe('fetchConversation', () => {
	test('returns empty conversation when fetchRef is not a function', async () => {
		const result = await fetchConversation(null, 1);
		expect(result).toEqual({ messages: [], hasMore: false });
	});

	test('returns empty conversation for an invalid user id', async () => {
		const fetchRef = vi.fn();
		expect(await fetchConversation(fetchRef, 0)).toEqual({ messages: [], hasMore: false });
		expect(await fetchConversation(fetchRef, -3)).toEqual({ messages: [], hasMore: false });
		expect(await fetchConversation(fetchRef, Number.NaN)).toEqual({ messages: [], hasMore: false });
		expect(fetchRef).not.toHaveBeenCalled();
	});

	test('requests the C03 history endpoint and unwraps data', async () => {
		const fetchRef = vi.fn(async () => ({
			ok: true,
			status: 200,
			json: async () => ({
				data: {
					messages: [{ id: 1, sender_id: 9, body: 'hi', created_at: '2026-01-01T00:00:00Z' }],
					has_more: true,
				},
			}),
		}));

		const result = await fetchConversation(fetchRef, 42);

		expect(fetchRef).toHaveBeenCalledWith(
			'/api/v1/chats/42/messages',
			expect.objectContaining({ credentials: 'include' }),
		);
		expect(result.messages).toHaveLength(1);
		expect(result.hasMore).toBe(true);
	});

	test('returns empty conversation on non-ok responses', async () => {
		const fetchRef = vi.fn(async () => ({ ok: false, status: 500, json: async () => ({}) }));
		expect(await fetchConversation(fetchRef, 5)).toEqual({ messages: [], hasMore: false });
	});

	test('returns empty conversation when fetch throws', async () => {
		const fetchRef = vi.fn(async () => {
			throw new Error('network');
		});
		expect(await fetchConversation(fetchRef, 5)).toEqual({ messages: [], hasMore: false });
	});

	test('defaults messages to an array when payload is malformed', async () => {
		const fetchRef = vi.fn(async () => ({
			ok: true,
			status: 200,
			json: async () => ({ data: { messages: 'nope' } }),
		}));
		const result = await fetchConversation(fetchRef, 5);
		expect(result.messages).toEqual([]);
		expect(result.hasMore).toBe(false);
	});
});

describe('fetchCurrentUserId', () => {
	test('returns null when fetchRef is not a function', async () => {
		expect(await fetchCurrentUserId(undefined)).toBeNull();
	});

	test('returns the id from the data wrapper', async () => {
		const fetchRef = vi.fn(async () => ({
			ok: true,
			status: 200,
			json: async () => ({ data: { id: 17, username: 'alice' } }),
		}));
		expect(await fetchCurrentUserId(fetchRef)).toBe(17);
		expect(fetchRef).toHaveBeenCalledWith(
			'/api/v1/users/me',
			expect.objectContaining({ credentials: 'include' }),
		);
	});

	test('returns null on non-ok response', async () => {
		const fetchRef = vi.fn(async () => ({ ok: false, status: 401, json: async () => ({}) }));
		expect(await fetchCurrentUserId(fetchRef)).toBeNull();
	});

	test('returns null when id is missing', async () => {
		const fetchRef = vi.fn(async () => ({
			ok: true,
			status: 200,
			json: async () => ({ data: {} }),
		}));
		expect(await fetchCurrentUserId(fetchRef)).toBeNull();
	});

	test('returns null when fetch throws', async () => {
		const fetchRef = vi.fn(async () => {
			throw new Error('network');
		});
		expect(await fetchCurrentUserId(fetchRef)).toBeNull();
	});
});
