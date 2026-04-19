import { describe, expect, test, vi } from 'vitest';
import { createApp, matchRoute, normalizePathname } from '../../main.js';

function createMockBrowser(initialPath = '/') {
	const documentListeners = new Map();
	const windowListeners = new Map();
	const historyCalls = [];

	const overlay = {
		style: { opacity: '1' },
		removed: false,
		remove() {
			this.removed = true;
		},
	};

	const mainContent = { innerHTML: '' };

	const setPath = (path) => {
		location.pathname = path;
		location.href = `https://example.test${path}`;
	};

	const dispatchWindow = (type, event = {}) => {
		const handlers = windowListeners.get(type) ?? [];
		for (const handler of handlers) {
			handler(event);
		}
	};

	const dispatchDocument = (type, event = {}) => {
		const handlers = documentListeners.get(type) ?? [];
		for (const handler of handlers) {
			handler(event);
		}
	};

	const location = {
		pathname: normalizePathname(initialPath),
		href: `https://example.test${normalizePathname(initialPath)}`,
		assign: vi.fn(),
	};

	const history = {
		stack: [normalizePathname(initialPath)],
		index: 0,
		pushState(_state, _title, path) {
			const normalized = normalizePathname(path);
			this.stack = this.stack.slice(0, this.index + 1);
			this.stack.push(normalized);
			this.index += 1;
			setPath(normalized);
			historyCalls.push({ type: 'push', path: normalized });
		},
		replaceState(_state, _title, path) {
			const normalized = normalizePathname(path);
			this.stack[this.index] = normalized;
			setPath(normalized);
			historyCalls.push({ type: 'replace', path: normalized });
		},
		back() {
			if (this.index === 0) {
				return;
			}
			this.index -= 1;
			setPath(this.stack[this.index]);
			dispatchWindow('popstate', { type: 'popstate' });
		},
		forward() {
			if (this.index >= this.stack.length - 1) {
				return;
			}
			this.index += 1;
			setPath(this.stack[this.index]);
			dispatchWindow('popstate', { type: 'popstate' });
		},
	};

	const windowRef = {
		location,
		history,
		addEventListener(type, handler) {
			if (!windowListeners.has(type)) {
				windowListeners.set(type, []);
			}
			windowListeners.get(type).push(handler);
		},
		removeEventListener(type, handler) {
			const handlers = windowListeners.get(type) ?? [];
			windowListeners.set(
				type,
				handlers.filter((candidate) => candidate !== handler),
			);
		},
		setTimeout(fn) {
			fn();
			return 1;
		},
		clearTimeout() {},
	};

	const documentRef = {
		readyState: 'complete',
		getElementById(id) {
			if (id === 'main-content') {
				return mainContent;
			}
			if (id === 'loading-overlay') {
				return overlay;
			}
			return null;
		},
		addEventListener(type, handler) {
			if (!documentListeners.has(type)) {
				documentListeners.set(type, []);
			}
			documentListeners.get(type).push(handler);
		},
		removeEventListener(type, handler) {
			const handlers = documentListeners.get(type) ?? [];
			documentListeners.set(
				type,
				handlers.filter((candidate) => candidate !== handler),
			);
		},
	};

	function clickLink(href) {
		const anchor = {
			getAttribute(name) {
				if (name === 'href') {
					return href;
				}
				return null;
			},
		};

		const event = {
			type: 'click',
			button: 0,
			metaKey: false,
			ctrlKey: false,
			shiftKey: false,
			altKey: false,
			defaultPrevented: false,
			target: {
				closest(selector) {
					if (selector === 'a[data-link]') {
						return anchor;
					}
					return null;
				},
			},
			preventDefault: vi.fn(() => {
				event.defaultPrevented = true;
			}),
		};

		dispatchDocument('click', event);
		return event;
	}

	return {
		windowRef,
		documentRef,
		mainContent,
		overlay,
		historyCalls,
		clickLink,
		submitForm: (formId) => {
			const form = {
				id: formId,
				getAttribute: (name) => (name === 'id' ? formId : null),
				closest: (selector) => (selector === 'form' ? form : null),
			};
			const event = {
				type: 'submit',
				target: form,
				preventDefault: vi.fn(),
				defaultPrevented: false,
			};
			const handlers = documentListeners.get('submit') ?? [];
			for (const handler of handlers) {
				handler(event);
			}
			return event;
		},
	};
}

describe('SPA routing for A03/A04', () => {
	test('normalizes and matches required routes', () => {
		expect(normalizePathname('')).toBe('/');
		expect(normalizePathname('/view-post/42')).toBe('/post/42');
		expect(normalizePathname('/activity/')).toBe('/activity');

		expect(matchRoute('/')?.route.id).toBe('feed');
		expect(matchRoute('/post/7')?.params.id).toBe('7');
		expect(matchRoute('/edit-post/15')?.params.id).toBe('15');
		expect(matchRoute('/does-not-exist')).toBeNull();

		// Regression: A03 Malformed URI should not crash
		expect(matchRoute('/post/%E0%A4%A')).toBeNull();
	});

	test('boot applies auth guard and redirects protected deep links', async () => {
		const browser = createMockBrowser('/activity');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: false, status: 401 })),
			setTimeoutRef: (fn) => {
				fn();
				return 1;
			},
		});

		await app.boot();

		expect(browser.windowRef.location.pathname).toBe('/login');
		expect(browser.mainContent.innerHTML).toContain('data-screen="login"');
		expect(browser.overlay.style.opacity).toBe('0');
		expect(browser.overlay.removed).toBe(true);
	});

	test('internal link navigation updates route without reload', async () => {
		const browser = createMockBrowser('/login');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: false, status: 401 })),
		});

		await app.boot();
		const clickEvent = browser.clickLink('/register');

		expect(clickEvent.preventDefault).toHaveBeenCalledTimes(1);
		expect(browser.windowRef.location.pathname).toBe('/register');
		expect(browser.mainContent.innerHTML).toContain('data-screen="register"');
		expect(browser.windowRef.location.assign).not.toHaveBeenCalled();
		expect(
			browser.historyCalls.some((call) => call.type === 'push' && call.path === '/register'),
		).toBe(true);
	});

	test('browser back and forward render supported routes', async () => {
		const browser = createMockBrowser('/');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: true, status: 200 })),
		});

		await app.boot();
		app.navigate('/create-post');
		app.navigate('/post/42');

		expect(browser.mainContent.innerHTML).toContain('data-screen="post-detail"');

		browser.windowRef.history.back();
		expect(browser.mainContent.innerHTML).toContain('data-screen="create-post"');

		browser.windowRef.history.back();
		expect(browser.mainContent.innerHTML).toContain('data-screen="feed"');

		browser.windowRef.history.forward();
		expect(browser.mainContent.innerHTML).toContain('data-screen="create-post"');
	});

	test('deep-link boot renders expected authenticated routes', async () => {
		const cases = [
			{ path: '/', expected: 'data-screen="feed"' },
			{ path: '/post/9', expected: 'data-screen="post-detail"' },
			{ path: '/create-post', expected: 'data-screen="create-post"' },
			{ path: '/edit-post/3', expected: 'data-screen="edit-post"' },
			{ path: '/activity', expected: 'data-screen="activity"' },
		];

		for (const testCase of cases) {
			const browser = createMockBrowser(testCase.path);
			const app = createApp({
				windowRef: browser.windowRef,
				documentRef: browser.documentRef,
				fetchRef: vi.fn(async () => ({ ok: true, status: 200 })),
			});

			await app.boot();
			expect(browser.mainContent.innerHTML).toContain(testCase.expected);
		}
	});

	test('authenticated routes render inside a shared shell with reserved chat regions', async () => {
		const browser = createMockBrowser('/');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: true, status: 200 })),
		});

		await app.boot();

		expect(browser.mainContent.innerHTML).toContain('data-auth-shell');
		expect(browser.mainContent.innerHTML).toContain('aria-label="Forum navigation"');
		expect(browser.mainContent.innerHTML).toContain('data-action="logout"');
		expect(browser.mainContent.innerHTML).toContain('data-chat-roster');
		expect(browser.mainContent.innerHTML).toContain('data-chat-active');
		expect(browser.mainContent.innerHTML).toContain('data-screen="feed"');
	});

	test('navigation and logout stay visible across protected route changes', async () => {
		const browser = createMockBrowser('/');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: true, status: 200 })),
		});

		await app.boot();
		app.navigate('/create-post');
		app.navigate('/activity');

		expect(browser.mainContent.innerHTML).toContain('data-screen="activity"');
		expect(browser.mainContent.innerHTML).toContain('aria-label="Forum navigation"');
		expect(browser.mainContent.innerHTML).toContain('data-action="logout"');
		expect(browser.mainContent.innerHTML).toContain('data-chat-roster');
		expect(browser.mainContent.innerHTML).toContain('data-chat-active');
	});

	test('public routes render outside authenticated shell', async () => {
		const browser = createMockBrowser('/login');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: false, status: 401 })),
		});

		await app.boot();
		expect(browser.mainContent.innerHTML).toContain('data-screen="login"');
		expect(browser.mainContent.innerHTML).not.toContain('data-auth-shell');

		browser.clickLink('/register');
		expect(browser.mainContent.innerHTML).toContain('data-screen="register"');
		expect(browser.mainContent.innerHTML).not.toContain('data-auth-shell');
	});

	test('form submission should be intercepted and prevented', async () => {
		const browser = createMockBrowser('/login');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: false, status: 401 })),
		});

		await app.boot();

		const event = browser.submitForm('login-form');
		expect(event.preventDefault).toHaveBeenCalled();
	});

	test('non-critical forms should not be intercepted by default', async () => {
		const browser = createMockBrowser('/');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: true, status: 200 })),
		});

		await app.boot();

		const event = browser.submitForm('search-form');
		expect(event.preventDefault).not.toHaveBeenCalled();
	});
});
