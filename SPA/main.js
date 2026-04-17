const ROUTE_DEFINITIONS = [
	{ id: 'login', pattern: '/login', access: 'public-only', title: 'Login' },
	{
		id: 'register',
		pattern: '/register',
		access: 'public-only',
		title: 'Register',
	},
	{ id: 'feed', pattern: '/', access: 'protected', title: 'Feed' },
	{
		id: 'post-detail',
		pattern: '/post/:id',
		access: 'protected',
		title: 'Post Detail',
	},
	{
		id: 'create-post',
		pattern: '/create-post',
		access: 'protected',
		title: 'Create Post',
	},
	{
		id: 'edit-post',
		pattern: '/edit-post/:id',
		access: 'protected',
		title: 'Edit Post',
	},
	{
		id: 'activity',
		pattern: '/activity',
		access: 'protected',
		title: 'Activity',
	},
];

const COMPILED_ROUTES = ROUTE_DEFINITIONS.map((route) => ({
	...route,
	...compileRoutePattern(route.pattern),
}));

function compileRoutePattern(pattern) {
	if (pattern === '/') {
		return { regex: /^\/$/, paramKeys: [] };
	}

	const parts = pattern
		.split('/')
		.filter(Boolean)
		.map((segment) => {
			if (segment.startsWith(':')) {
				return '([^/]+)';
			}

			return escapeRegex(segment);
		});

	const paramKeys = pattern
		.split('/')
		.filter((segment) => segment.startsWith(':'))
		.map((segment) => segment.slice(1));

	return {
		regex: new RegExp(`^/${parts.join('/')}$`),
		paramKeys,
	};
}

function escapeRegex(value) {
	return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export function normalizePathname(pathname) {
	const raw = typeof pathname === 'string' && pathname.trim() ? pathname.trim() : '/';
	let normalized = raw.startsWith('/') ? raw : `/${raw}`;

	normalized = normalized.replace(/\/+/g, '/');

	if (normalized.length > 1 && normalized.endsWith('/')) {
		normalized = normalized.slice(0, -1);
	}

	const legacyPostMatch = /^\/view-post\/([^/]+)$/.exec(normalized);
	if (legacyPostMatch) {
		normalized = `/post/${legacyPostMatch[1]}`;
	}

	return normalized || '/';
}

export function matchRoute(pathname) {
	const targetPath = normalizePathname(pathname);

	for (const route of COMPILED_ROUTES) {
		const result = route.regex.exec(targetPath);
		if (!result) {
			continue;
		}

		const params = {};
		for (let i = 0; i < route.paramKeys.length; i += 1) {
			params[route.paramKeys[i]] = decodeURIComponent(result[i + 1]);
		}

		return { route, params, path: targetPath };
	}

	return null;
}

function escapeHTML(value) {
	return String(value)
		.replaceAll('&', '&amp;')
		.replaceAll('<', '&lt;')
		.replaceAll('>', '&gt;')
		.replaceAll('"', '&quot;')
		.replaceAll("'", '&#39;');
}

function renderPublicNavigation() {
	return `
		<nav aria-label="Authentication navigation">
			<a data-link href="/login">Login</a>
			<span aria-hidden="true"> | </span>
			<a data-link href="/register">Register</a>
		</nav>
	`;
}

function renderAppNavigation() {
	return `
		<nav aria-label="Forum navigation">
			<a data-link href="/">Feed</a>
			<span aria-hidden="true"> | </span>
			<a data-link href="/create-post">Create Post</a>
			<span aria-hidden="true"> | </span>
			<a data-link href="/activity">Activity</a>
		</nav>
	`;
}

function renderAuthenticatedShell(content) {
	return `
		<div class="app-shell" data-auth-shell>
			<header class="app-shell__header" aria-label="Forum header">
				<div class="app-shell__brand">Real-Time Forum</div>
				${renderAppNavigation()}
				<button class="app-shell__logout" type="button" data-action="logout">Logout</button>
			</header>
			<div class="app-shell__layout">
				<section class="app-shell__outlet" aria-label="Page content">${content}</section>
				<aside class="app-shell__chat" aria-label="Direct messages">
					<section class="chat-panel" data-chat-roster>
						<h2>Chat Roster</h2>
					</section>
					<section class="chat-panel" data-chat-active>
						<h2>Active Chat</h2>
					</section>
				</aside>
			</div>
		</div>
	`;
}

function renderTemplate(match) {
	const { id, title } = match.route;

	if (id === 'login') {
		return `
			<section data-screen="login" aria-labelledby="screen-login-title">
				<h1 id="screen-login-title">Login</h1>
				<p>Sign in to continue to the forum.</p>
				${renderPublicNavigation()}
			</section>
		`;
	}

	if (id === 'register') {
		return `
			<section data-screen="register" aria-labelledby="screen-register-title">
				<h1 id="screen-register-title">Register</h1>
				<p>Create an account to join the forum.</p>
				${renderPublicNavigation()}
			</section>
		`;
	}

	if (id === 'post-detail') {
		const postID = escapeHTML(match.params.id || '');
		return `
			<section data-screen="post-detail" aria-labelledby="screen-post-title">
				<h1 id="screen-post-title">Post Detail</h1>
				<p data-post-id="${postID}">Viewing post ${postID}</p>
			</section>
		`;
	}

	if (id === 'edit-post') {
		const postID = escapeHTML(match.params.id || '');
		return `
			<section data-screen="edit-post" aria-labelledby="screen-edit-post-title">
				<h1 id="screen-edit-post-title">Edit Post</h1>
				<p data-post-id="${postID}">Editing post ${postID}</p>
			</section>
		`;
	}

	if (id === 'create-post') {
		return `
			<section data-screen="create-post" aria-labelledby="screen-create-post-title">
				<h1 id="screen-create-post-title">Create Post</h1>
				<p>Compose and publish a new post.</p>
			</section>
		`;
	}

	if (id === 'activity') {
		return `
			<section data-screen="activity" aria-labelledby="screen-activity-title">
				<h1 id="screen-activity-title">Activity</h1>
				<p>Review recent activity and updates.</p>
			</section>
		`;
	}

	return `
		<section data-screen="feed" aria-labelledby="screen-feed-title">
			<h1 id="screen-feed-title">${escapeHTML(title)}</h1>
			<p>Browse the latest discussions.</p>
		</section>
	`;
}

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

		const routeMarkup = renderTemplate(match);
		mainContent.innerHTML =
			match.route.access === 'protected' ? renderAuthenticatedShell(routeMarkup) : routeMarkup;
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
		windowRef.addEventListener('popstate', onPopState);
		return boot();
	}

	function stop() {
		documentRef.removeEventListener('click', onDocumentClick);
		windowRef.removeEventListener('popstate', onPopState);
	}

	return {
		boot: start,
		stop,
		navigate: (pathname, options = {}) => {
			goTo(pathname, Boolean(options.replace));
		},
		handleLocationChange,
		getState: () => ({ ...state }),
	};
}

const app = createApp();

if (app) {
	if (document.readyState === 'loading') {
		document.addEventListener(
			'DOMContentLoaded',
			() => {
				void app.boot();
			},
			{ once: true },
		);
	} else {
		void app.boot();
	}
}
