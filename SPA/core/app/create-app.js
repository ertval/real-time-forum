// SPA/core/app/create-app.js

import { renderAuthenticatedShell } from '../../features/shell/shell.views.js';
import { renderTemplate } from '../router/render-template.js';
import { matchRoute, normalizePathname } from '../router/routes.js';

const AUTH_SHELL_SELECTOR = '[data-auth-shell]';
const AUTH_OUTLET_SELECTOR = '.app-shell__outlet';
const AUTH_OUTLET_OPENING_TAG = '<main class="app-shell__outlet" aria-label="Page content">';

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

function replaceProtectedOutletMarkup(shellMarkup, outletMarkup) {
	const outletStartMarkerIndex = shellMarkup.indexOf(AUTH_OUTLET_OPENING_TAG);
	if (outletStartMarkerIndex === -1) {
		return null;
	}

	const outletContentStart = outletStartMarkerIndex + AUTH_OUTLET_OPENING_TAG.length;
	const outletEndMarkerIndex = shellMarkup.indexOf('</main>', outletContentStart);
	if (outletEndMarkerIndex === -1) {
		return null;
	}

	return `${shellMarkup.slice(0, outletContentStart)}${outletMarkup}${shellMarkup.slice(outletEndMarkerIndex)}`;
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
		authShellRoot: null,
		authShellOutlet: null,
	};

	function cacheAuthShellNodes() {
		if (typeof mainContent.querySelector !== 'function') {
			state.authShellRoot = null;
			state.authShellOutlet = null;
			return;
		}

		state.authShellRoot = mainContent.querySelector(AUTH_SHELL_SELECTOR);
		state.authShellOutlet = state.authShellRoot?.querySelector?.(AUTH_OUTLET_SELECTOR) ?? null;
	}

	function renderProtectedRoute(routeMarkup) {
		if (typeof mainContent.querySelector === 'function') {
			if (!state.authShellRoot || !state.authShellOutlet) {
				cacheAuthShellNodes();
			}

			if (!state.authShellRoot || !state.authShellOutlet) {
				mainContent.innerHTML = renderAuthenticatedShell(routeMarkup);
				cacheAuthShellNodes();
				return;
			}

			state.authShellOutlet.innerHTML = routeMarkup;
			return;
		}

		const currentMarkup = typeof mainContent.innerHTML === 'string' ? mainContent.innerHTML : '';
		if (!currentMarkup.includes('data-auth-shell')) {
			mainContent.innerHTML = renderAuthenticatedShell(routeMarkup);
			return;
		}

		const updatedMarkup = replaceProtectedOutletMarkup(currentMarkup, routeMarkup);
		mainContent.innerHTML = updatedMarkup ?? renderAuthenticatedShell(routeMarkup);
	}

	function renderPublicRoute(routeMarkup) {
		state.authShellRoot = null;
		state.authShellOutlet = null;
		mainContent.innerHTML = routeMarkup;
	}

	function renderRoute(match) {
		if (!mainContent) {
			return;
		}

		const routeMarkup = renderTemplate(match);

		if (match.route.access === 'protected') {
			renderProtectedRoute(routeMarkup);
			return;
		}

		renderPublicRoute(routeMarkup);
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

	function onDocumentClick(event) {
		if (event.defaultPrevented || event.button !== 0) {
			return;
		}

		if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
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
