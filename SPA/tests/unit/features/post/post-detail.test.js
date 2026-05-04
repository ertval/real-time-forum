// SPA/tests/unit/features/post/post-detail.test.js

import { describe, expect, test, vi } from 'vitest';
import { initPostDetailPage } from '../../../../features/post/post-detail.page.js';

describe('initPostDetailPage', () => {
	test('returns null if parameters are invalid', () => {
		expect(initPostDetailPage()).toBeNull();
	});

	test('initializes and calls fetch for post content and comments', async () => {
		const fetchMock = vi.fn().mockImplementation(() =>
			Promise.resolve({
				ok: true,
				json: () =>
					Promise.resolve({
						data: {
							id: 1,
							title: 'Sample Post',
							body: 'This is the post body',
						},
					}),
			}),
		);

		const rootMock = {
			getAttribute: vi.fn().mockReturnValue('1'),
			setAttribute: vi.fn(),
			querySelector: vi.fn().mockReturnValue({
				innerHTML: '',
			}),
		};

		const docMock = {
			querySelector: vi.fn().mockReturnValue(rootMock),
		};

		const options = {
			windowRef: {},
			documentRef: docMock,
			fetchRef: fetchMock,
		};

		const result = initPostDetailPage(options);
		expect(result).not.toBeNull();
		expect(result.root).toBe(rootMock);
		expect(rootMock.setAttribute).toHaveBeenCalledWith('data-post-detail-bound', 'true');
	});
});
