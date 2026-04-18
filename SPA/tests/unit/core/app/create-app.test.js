// SPA/tests/unit/core/app/create-app.test.js

import { describe, expect, test, vi } from 'vitest';
import { createApp } from '../../../../core/app/create-app.js';
import { createMockBrowser } from '../../helpers/browser-mock.js';

describe('SPA Application Engine', () => {
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

	test('boot checks session via GET /api/v1/users/me with cookie credentials', async () => {
		const browser = createMockBrowser('/login');
		const fetchRef = vi.fn(async () => ({ ok: false, status: 401 }));
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef,
		});

		await app.boot();

		expect(fetchRef).toHaveBeenCalledTimes(1);
		expect(fetchRef).toHaveBeenCalledWith('/api/v1/users/me', {
			method: 'GET',
			credentials: 'include',
			headers: {
				Accept: 'application/json',
			},
		});
	});

	test('unauthenticated boot blocks every protected route and redirects to login', async () => {
		const protectedRoutes = ['/', '/post/9', '/create-post', '/edit-post/3', '/activity'];

		for (const path of protectedRoutes) {
			const browser = createMockBrowser(path);
			const app = createApp({
				windowRef: browser.windowRef,
				documentRef: browser.documentRef,
				fetchRef: vi.fn(async () => ({ ok: false, status: 401 })),
			});

			await app.boot();

			expect(browser.windowRef.location.pathname).toBe('/login');
			expect(browser.mainContent.innerHTML).toContain('data-screen="login"');
		}
	});

	test('authenticated boot redirects public-only entry states into forum shell', async () => {
		const browser = createMockBrowser('/register');
		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef: vi.fn(async () => ({ ok: true, status: 200 })),
		});

		await app.boot();

		expect(browser.windowRef.location.pathname).toBe('/');
		expect(browser.mainContent.innerHTML).toContain('data-auth-shell');
		expect(browser.mainContent.innerHTML).toContain('data-screen="feed"');
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

	test('logout button calls API and returns user to login flow', async () => {
		const browser = createMockBrowser('/activity');
		const fetchRef = vi.fn(async (url) => {
			if (url === '/api/v1/users/me') {
				return { ok: true, status: 200 };
			}
			if (url === '/api/v1/users/logout') {
				return { ok: true, status: 200 };
			}
			return { ok: false, status: 404 };
		});

		const app = createApp({
			windowRef: browser.windowRef,
			documentRef: browser.documentRef,
			fetchRef,
		});

		await app.boot();
		expect(browser.mainContent.innerHTML).toContain('data-action="logout"');

		const event = browser.clickLogout();
		await Promise.resolve();

		expect(event.preventDefault).toHaveBeenCalledTimes(1);
		expect(fetchRef).toHaveBeenCalledWith('/api/v1/users/logout', {
			method: 'POST',
			credentials: 'include',
			headers: {
				Accept: 'application/json',
			},
		});
		expect(app.getState().isAuthenticated).toBe(false);
		expect(browser.windowRef.location.pathname).toBe('/login');
		expect(browser.mainContent.innerHTML).toContain('data-screen="login"');
		expect(browser.mainContent.innerHTML).not.toContain('data-auth-shell');
	});

	test('logout is usable from every authenticated route', async () => {
		const protectedRoutes = ['/', '/post/9', '/create-post', '/edit-post/3', '/activity'];

		for (const path of protectedRoutes) {
			const browser = createMockBrowser(path);
			const fetchRef = vi.fn(async (url) => {
				if (url === '/api/v1/users/me') {
					return { ok: true, status: 200 };
				}
				if (url === '/api/v1/users/logout') {
					return { ok: true, status: 200 };
				}
				return { ok: false, status: 404 };
			});

			const app = createApp({
				windowRef: browser.windowRef,
				documentRef: browser.documentRef,
				fetchRef,
			});

			await app.boot();
			expect(browser.mainContent.innerHTML).toContain('data-action="logout"');

			browser.clickLogout();
			await Promise.resolve();

			expect(browser.windowRef.location.pathname).toBe('/login');
			expect(app.getState().isAuthenticated).toBe(false);
		}
	});
});
