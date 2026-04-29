// SPA/tests/unit/features/post/post.page.test.js
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';

var createPostComment;
var getPostById;
var getPostComments;
var setupImagePicker;
const OriginalFormData = globalThis.FormData;

vi.mock('../../../../features/post/post.api.js', () => {
	createPostComment = vi.fn();
	getPostById = vi.fn();
	getPostComments = vi.fn();

	return {
		createPostComment,
		getPostById,
		getPostComments,
	};
});

vi.mock('../../../../../web/static/js/image-picker.js', () => {
	setupImagePicker = vi.fn();

	return {
		setupImagePicker,
	};
});

import { initPostDetailPage } from '../../../../features/post/post.page.js';

class FakeCommentsContainer {
	constructor() {
		this.innerHTML = '';
	}
}

class FakeFeedbackNode {
	constructor() {
		this.hidden = true;
		this.textContent = '';
		this.attributes = new Map();
	}

	setAttribute(name, value) {
		this.attributes.set(name, String(value));
	}

	removeAttribute(name) {
		this.attributes.delete(name);
	}
}

class FakeSubmitButton {
	constructor() {
		this.disabled = false;
		this.textContent = 'Post comment';
		this.dataset = {};
	}
}

class FakeImageInput {
	constructor() {
		this.files = [];
	}
}

class FakeImageElement {}

class FakeButton {
	constructor() {
		this.hidden = false;
	}
}

class FakeCommentForm {
	constructor() {
		this.dataset = {};
		this.listeners = new Map();
		this.feedback = new FakeFeedbackNode();
		this.submitButton = new FakeSubmitButton();
		this.imageInput = new FakeImageInput();
		this.imageButton = new FakeButton();
		this.imageClear = new FakeButton();
		this.imageName = {};
		this.imagePreview = { hidden: true };
		this.imagePreviewImage = new FakeImageElement();
		this.fields = { body: '' };
		this.reset = vi.fn(() => {
			this.fields.body = '';
			this.imageInput.files = [];
		});
	}

	addEventListener(type, handler) {
		this.listeners.set(type, handler);
	}

	querySelector(selector) {
		if (selector === '[data-comment-feedback]') {
			return this.feedback;
		}

		if (selector === 'button[type="submit"]') {
			return this.submitButton;
		}

		if (selector === '.comment-image-input') {
			return this.imageInput;
		}

		if (selector === '.comment-image-btn') {
			return this.imageButton;
		}

		if (selector === '.image-clear') {
			return this.imageClear;
		}

		if (selector === '.comment-image-name') {
			return this.imageName;
		}

		if (selector === '.comment-image-preview') {
			return this.imagePreview;
		}

		if (selector === '.comment-image-preview img') {
			return this.imagePreviewImage;
		}

		return null;
	}

	async submit() {
		const handler = this.listeners.get('submit');
		const event = {
			preventDefault: vi.fn(),
		};
		await handler(event);
		return event;
	}
}

class FakeRenderedRoot {
	constructor(markup) {
		this.markup = markup;
		this.attributes = new Map();
		this.innerHTML = markup;
		this.commentsContainer = new FakeCommentsContainer();
		this.commentForm = new FakeCommentForm();
	}

	getAttribute(name) {
		return this.attributes.get(name) ?? null;
	}

	setAttribute(name, value) {
		this.attributes.set(name, String(value));
	}

	querySelector(selector) {
		if (selector === '[data-comments]') {
			return this.commentsContainer;
		}

		if (selector === '#comment-form') {
			return this.commentForm;
		}

		return null;
	}
}

class FakePlaceholderRoot {
	constructor(documentRef) {
		this.documentRef = documentRef;
		this.attributes = new Map();
		this.innerHTML = '';
		this._outerHTML = '';
	}

	getAttribute(name) {
		return this.attributes.get(name) ?? null;
	}

	setAttribute(name, value) {
		this.attributes.set(name, String(value));
	}

	set outerHTML(markup) {
		this._outerHTML = markup;
		this.documentRef.currentRoot = new FakeRenderedRoot(markup);
	}

	get outerHTML() {
		return this._outerHTML;
	}
}

function createPostDetailDocument() {
	const documentRef = {
		currentRoot: null,
		querySelector(selector) {
			if (selector === '[data-screen="post-detail"]') {
				return this.currentRoot;
			}

			return null;
		},
	};

	documentRef.currentRoot = new FakePlaceholderRoot(documentRef);
	return documentRef;
}

beforeEach(() => {
	createPostComment.mockReset();
	getPostById.mockReset();
	getPostComments.mockReset();
	setupImagePicker.mockReset();
	setupImagePicker.mockReturnValue({
		getFile: () => null,
		clearSelectedFile: vi.fn(),
	});
	globalThis.FormData = class FakeFormData {
		constructor(form) {
			this.form = form;
		}

		get(name) {
			return this.form?.fields?.[name] ?? '';
		}
	};
});

afterEach(() => {
	globalThis.FormData = OriginalFormData;
});

describe('initPostDetailPage', () => {
	test('fetches post and comments using the URL id and renders comments', async () => {
		const documentRef = createPostDetailDocument();
		const windowRef = { location: { pathname: '/posts/42' } };
		const fetchRef = vi.fn();

		getPostById.mockResolvedValue({
			id: 42,
			title: 'Route hydration',
			body: 'Loaded body',
			username: 'alice',
			created_at: '2026-04-29T09:30:00Z',
			categories: ['SPA'],
		});
		getPostComments.mockResolvedValue([
			{
				id: 9,
				body: 'First reply',
				username: 'bob',
				created_at: '2026-04-29T10:00:00Z',
			},
		]);

		const result = await initPostDetailPage({ windowRef, documentRef, fetchRef });

		expect(getPostById).toHaveBeenCalledWith(fetchRef, '42');
		expect(getPostComments).toHaveBeenCalledWith(fetchRef, '42');
		expect(documentRef.currentRoot.markup).toContain('Route hydration');
		expect(documentRef.currentRoot.commentsContainer.innerHTML).toContain('First reply');
		expect(result).toEqual({
			post: expect.objectContaining({ id: 42, title: 'Route hydration' }),
			comments: [expect.objectContaining({ id: 9, body: 'First reply' })],
		});
	});

	test('renders an error state when the post request fails', async () => {
		const documentRef = createPostDetailDocument();
		const windowRef = { location: { pathname: '/posts/77' } };

		getPostById.mockResolvedValue(null);

		const result = await initPostDetailPage({
			windowRef,
			documentRef,
			fetchRef: vi.fn(),
		});

		expect(documentRef.currentRoot.innerHTML).toContain('Post unavailable');
		expect(documentRef.currentRoot.innerHTML).toContain('Unable to load this post right now.');
		expect(getPostComments).not.toHaveBeenCalled();
		expect(result).toBeNull();
	});

	test('submits a comment, clears the form, and reloads comments', async () => {
		const documentRef = createPostDetailDocument();
		const windowRef = { location: { pathname: '/posts/42' } };
		const fetchRef = vi.fn();

		getPostById.mockResolvedValue({
			id: 42,
			title: 'Route hydration',
			body: 'Loaded body',
			username: 'alice',
			created_at: '2026-04-29T09:30:00Z',
			categories: ['SPA'],
		});
		getPostComments
			.mockResolvedValueOnce([
				{ id: 9, body: 'First reply', username: 'bob', created_at: '2026-04-29T10:00:00Z' },
			])
			.mockResolvedValueOnce([
				{ id: 9, body: 'First reply', username: 'bob', created_at: '2026-04-29T10:00:00Z' },
				{ id: 10, body: 'New reply', username: 'alice', created_at: '2026-04-29T10:05:00Z' },
			]);
		createPostComment.mockResolvedValue({ ok: true, status: 201, data: { id: 10 } });

		await initPostDetailPage({ windowRef, documentRef, fetchRef });

		documentRef.currentRoot.commentForm.fields.body = 'New reply';
		await documentRef.currentRoot.commentForm.submit();

		expect(createPostComment).toHaveBeenCalledWith(fetchRef, '42', {
			body: 'New reply',
			imageFile: null,
		});
		expect(documentRef.currentRoot.commentForm.reset).toHaveBeenCalledTimes(1);
		expect(getPostComments).toHaveBeenCalledTimes(2);
		expect(documentRef.currentRoot.commentsContainer.innerHTML).toContain('New reply');
	});

	test('allows image-only comment submission via the shared picker flow', async () => {
		const documentRef = createPostDetailDocument();
		const windowRef = { location: { pathname: '/posts/42' } };
		const fetchRef = vi.fn();
		const imageFile = { name: 'reply.png', size: 128, type: 'image/png' };
		const clearSelectedFile = vi.fn();

		setupImagePicker.mockReturnValue({
			getFile: () => imageFile,
			clearSelectedFile,
		});

		getPostById.mockResolvedValue({
			id: 42,
			title: 'Route hydration',
			body: 'Loaded body',
			username: 'alice',
			created_at: '2026-04-29T09:30:00Z',
			categories: ['SPA'],
		});
		getPostComments.mockResolvedValueOnce([]).mockResolvedValueOnce([
			{
				id: 11,
				body: '',
				username: 'alice',
				created_at: '2026-04-29T10:05:00Z',
				image_url: '/static/uploads/comment/reply.png',
			},
		]);
		createPostComment.mockResolvedValue({ ok: true, status: 201, data: { id: 11 } });

		await initPostDetailPage({ windowRef, documentRef, fetchRef });
		await documentRef.currentRoot.commentForm.submit();

		expect(createPostComment).toHaveBeenCalledWith(fetchRef, '42', {
			body: '',
			imageFile,
		});
		expect(clearSelectedFile).toHaveBeenCalledTimes(1);
		expect(documentRef.currentRoot.commentsContainer.innerHTML).toContain(
			'/static/uploads/comment/reply.png',
		);
	});
});
