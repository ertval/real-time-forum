// SPA/core/app/create-app.js

import { renderAuthenticatedShell } from '../../features/shell/shell.views.js';
import { renderTemplate } from '../router/render-template.js';
import { matchRoute, normalizePathname } from '../router/routes.js';

function hideLoadingOverlay(documentRef, setTimeoutRef) {
	const overlay = documentRef.getElementById('loading-overlay');
	if (!overlay) {
		return;
	}

	overlay.style.opacity = '0';
	setTimeoutRef(() => overlay.remove(), 200);
}

async function resolveSession(fetchRef) {
	if (typeof fetchRef !== 'function') {
		return { isAuthenticated: false, status: 0 };
	}

	try {
		const response = await fetchRef('/api/v1/users/me', {
			method: 'GET',
			credentials: 'include',
			headers: {
				Accept: 'application/json',
			},
		});

		return { isAuthenticated: response.ok, status: response.status };
	} catch {
		return { isAuthenticated: false, status: 0 };
	}
}

function enforceRouteAccess(match, isAuthenticated) {
	if (!match) {
		return {
			redirectTo: isAuthenticated ? '/' : '/login',
			match: null,
		};
	}

	if (match.route.access === 'protected' && !isAuthenticated) {
		return { redirectTo: '/login', match: null };
	}

	if (match.route.access === 'public-only' && isAuthenticated) {
		return { redirectTo: '/', match: null };
	}

	return { redirectTo: null, match };
}

export function createApp(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const fetchRef = options.fetchRef ?? (typeof fetch === 'function' ? fetch : null);

	if (!windowRef || !documentRef) {
		return null;
	}

	const mainContent = documentRef.getElementById('main-content');
	const setTimeoutRef =
		typeof options.setTimeoutRef === 'function'
			? options.setTimeoutRef
			: windowRef.setTimeout.bind(windowRef);

	const state = {
		isAuthenticated: false,
		activePath: normalizePathname(windowRef.location.pathname || '/'),
	};

	function renderRoute(match) {
		if (!mainContent) {
			return;
		}

		const isProtected = match.route.access === 'protected';
		const routeMarkup = renderTemplate(match);

		if (isProtected) {
			// Persistent Shell logic: only render shell if not already present
			const existingShell = mainContent.querySelector('[data-auth-shell]');
			if (existingShell) {
				const outlet = existingShell.querySelector('.app-shell__outlet');
				if (outlet) {
					outlet.innerHTML = routeMarkup;
				} else {
					// Fallback if DOM structure is compromised
					mainContent.innerHTML = renderAuthenticatedShell(routeMarkup);
				}
			} else {
				mainContent.innerHTML = renderAuthenticatedShell(routeMarkup);
			}
		} else {
			// Public routes: always full render
			mainContent.innerHTML = routeMarkup;
		}
	}

	function goTo(pathname, replace = false) {
		const normalizedTarget = normalizePathname(pathname);
		const currentPath = normalizePathname(windowRef.location.pathname || '/');

		if (replace) {
			windowRef.history.replaceState({}, '', normalizedTarget);
		} else if (normalizedTarget !== currentPath) {
			windowRef.history.pushState({}, '', normalizedTarget);
		}

		state.activePath = normalizedTarget;
		handleLocationChange();
	}

	function handleLocationChange() {
		const normalizedPath = normalizePathname(windowRef.location.pathname || '/');

		if (normalizedPath !== windowRef.location.pathname) {
			windowRef.history.replaceState({}, '', normalizedPath);
		}

		const match = matchRoute(normalizedPath);
		const access = enforceRouteAccess(match, state.isAuthenticated);

		if (access.redirectTo) {
			const normalizedRedirect = normalizePathname(access.redirectTo);
			if (normalizedRedirect !== normalizedPath) {
				goTo(normalizedRedirect, true);
				return;
			}
		}

		if (!access.match) {
			return;
		}

		state.activePath = normalizedPath;
		renderRoute(access.match);
	}

	async function performLogout() {
		if (typeof fetchRef === 'function') {
			try {
				await fetchRef('/api/v1/users/logout', {
					method: 'POST',
					credentials: 'include',
					headers: {
						Accept: 'application/json',
					},
				});
			} catch {
				// Continue local logout flow even on network failure.
			}
		}

		state.isAuthenticated = false;
		goTo('/login', true);
	}

	async function handleAuthForm(form) {
		const formData = new FormData(form);
		const data = Object.fromEntries(formData.entries());

		// Basic data normalization for age
		if (data.age) {
			data.age = Number.parseInt(data.age, 10);
		}

		const endpoint = form.id === 'login-form' ? '/api/v1/users/login' : '/api/v1/users/register';

		try {
			const response = await fetchRef(endpoint, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data),
				credentials: 'include',
			});

			if (response.ok) {
				state.isAuthenticated = true;
				goTo('/');
			} else {
				const error = await response.json();
				alert(error.error?.message || 'Authentication failed');
			}
		} catch (err) {
			console.error('Auth error:', err);
			alert('A connection error occurred. Please try again.');
		}
	}

	function onDocumentClick(event) {
		if (event.defaultPrevented || event.button !== 0) {
			return;
		}

		if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
			return;
		}

		const logoutButton = event.target?.closest?.('[data-action="logout"]');
		if (logoutButton) {
			event.preventDefault();
			void performLogout();
			return;
		}

		const anchor = event.target?.closest?.('a[data-link]');
		if (!anchor) {
			return;
		}

		const href = anchor.getAttribute('href');
		if (!href || href.startsWith('http') || href.startsWith('mailto:')) {
			return;
		}

		event.preventDefault();
		goTo(href, false);
	}

	function onDocumentSubmit(event) {
		const form = event.target?.closest?.('form');
		if (!form) {
			return;
		}

		// Only intercept critical authentication forms to prevent URL exposure
		const criticalForms = ['login-form', 'register-form'];
		if (criticalForms.includes(form.id)) {
			event.preventDefault();
			void handleAuthForm(form);
		}
	}

	function onPopState() {
		handleLocationChange();
	}

	async function boot() {
		const session = await resolveSession(fetchRef);
		state.isAuthenticated = session.isAuthenticated;

		handleLocationChange();
		hideLoadingOverlay(documentRef, setTimeoutRef);
	}

	function start() {
		documentRef.addEventListener('click', onDocumentClick);
		documentRef.addEventListener('submit', onDocumentSubmit);
		windowRef.addEventListener('popstate', onPopState);
		return boot();
	}

	function stop() {
		documentRef.removeEventListener('click', onDocumentClick);
		documentRef.removeEventListener('submit', onDocumentSubmit);
		windowRef.removeEventListener('popstate', onPopState);
	}

	return {
		boot: start,
		stop,
		navigate: (pathname, navigationOptions = {}) => {
			goTo(pathname, Boolean(navigationOptions.replace));
		},
		handleLocationChange,
		getState: () => ({ ...state }),
	};
}
