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
				<div class="app-shell__brand">
					<span style="color: var(--accent-primary); margin-right: 4px;">F</span>orum
				</div>
				${renderAppNavigation()}
				<div style="display: flex; align-items: center; gap: 1rem;">
					<button class="app-shell__logout" type="button" data-action="logout">Logout</button>
				</div>
			</header>
			<div class="app-shell__layout">
				<main class="app-shell__outlet" aria-label="Page content">
					${content}
				</main>
				<aside class="app-shell__chat" aria-label="Direct messages">
					<section class="chat-panel" data-chat-roster>
						<h2 style="font-size: 0.875rem; color: var(--text-muted); text-transform: uppercase; margin-bottom: 1rem;">Chat</h2>
						<div id="roster-list" style="display: flex; flex-direction: column; gap: 0.5rem;">
							<!-- Dynamic roster -->
							<p style="font-size: 0.875rem; color: var(--text-muted);">No active chats</p>
						</div>
					</section>
					<section class="chat-panel" data-chat-active style="margin-top: 1rem;">
						<h2 style="font-size: 0.875rem; color: var(--text-muted); text-transform: uppercase;">Active</h2>
						<div style="height: 200px; display: flex; align-items: center; justify-content: center; color: var(--text-muted); font-size: 0.875rem;">
							Select a user to chat
						</div>
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
			<div class="auth-page" data-screen="login">
				<div class="auth-form-container">
					<div class="auth-form-card">
						<div class="auth-header">
							<div class="auth-logo">
								<div class="auth-logo-icon">F</div>
								Real-Time Forum
							</div>
							<h1 class="auth-title">Welcome Back!</h1>
							<p class="auth-subtitle">Sign in to access your dashboard and continue participating in the community.</p>
						</div>
						
						<form class="auth-form" id="login-form">
							<div class="form-group">
								<label class="form-label" for="login-identifier">Username or Email</label>
								<div class="input-wrapper">
									<span class="input-icon">
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
									</span>
									<input class="auth-input" type="text" id="login-identifier" name="identifier" placeholder="Enter your username or email" required autocomplete="username">
								</div>
							</div>
							<div class="form-group">
								<label class="form-label" for="login-password">Password</label>
								<div class="input-wrapper">
									<span class="input-icon">
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
									</span>
									<input class="auth-input" type="password" id="login-password" name="password" placeholder="Enter your password" required autocomplete="current-password">
								</div>
							</div>
							<div class="form-footer">
								<a href="#" class="forgot-link">Forgot Password?</a>
							</div>
							<button type="submit" class="auth-button">Sign In</button>
						</form>

						<div class="separator">OR</div>

						<div class="social-buttons">
							<button class="social-button" type="button">
								<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>
								Continue with Google
							</button>
						</div>

						<p class="switch-auth">
							Don't have an account? <a data-link href="/register">Sign Up</a>
						</p>
					</div>
				</div>
				<div class="auth-hero">
					<div class="hero-content">
						<h2 class="hero-title">The heart of community.</h2>
						<div class="testimonial">
							<p class="testimonial-text">"The real-time updates make it feel like a living conversation. It's transformed how we collaborate."</p>
							<div class="testimonial-author">
								<img class="author-avatar" src="https://i.pravatar.cc/100?u=michael" alt="Michael Carter">
								<div class="author-info">
									<h4>Michael Carter</h4>
									<p>Lead Developer at DevCore</p>
								</div>
							</div>
						</div>
					</div>
					<div class="hero-footer">
						<div class="footer-label">Powering 1K+ Communities</div>
						<div class="brand-logos">
							<span style="font-weight: 800; letter-spacing: -1px;">DISCORD</span>
							<span style="font-weight: 800; letter-spacing: -1px;">SLACK</span>
							<span style="font-weight: 800; letter-spacing: -1px;">REDDIT</span>
						</div>
					</div>
				</div>
			</div>
		`;
	}

	if (id === 'register') {
		return `
			<div class="auth-page" data-screen="register">
				<div class="auth-form-container">
					<div class="auth-form-card" style="max-width: 500px;">
						<div class="auth-header">
							<div class="auth-logo">
								<div class="auth-logo-icon">F</div>
								Real-Time Forum
							</div>
							<h1 class="auth-title">Create Account</h1>
							<p class="auth-subtitle">Join our community and start sharing your thoughts today.</p>
						</div>
						
						<form class="auth-form" id="register-form">
							<div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-md);">
								<div class="form-group">
									<label class="form-label" for="reg-first-name">First Name</label>
									<input class="auth-input" style="padding-left: 1rem;" type="text" id="reg-first-name" name="first_name" placeholder="John" required>
								</div>
								<div class="form-group">
									<label class="form-label" for="reg-last-name">Last Name</label>
									<input class="auth-input" style="padding-left: 1rem;" type="text" id="reg-last-name" name="last_name" placeholder="Doe" required>
								</div>
							</div>

							<div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-md);">
								<div class="form-group">
									<label class="form-label" for="reg-age">Age</label>
									<input class="auth-input" style="padding-left: 1rem;" type="number" id="reg-age" name="age" min="13" placeholder="25" required>
								</div>
								<div class="form-group">
									<label class="form-label" for="reg-gender">Gender</label>
									<select class="auth-input" style="padding-left: 1rem;" id="reg-gender" name="gender" required>
										<option value="" disabled selected>Select</option>
										<option value="male">Male</option>
										<option value="female">Female</option>
										<option value="other">Other</option>
										<option value="prefer-not-to-say">Prefer not to say</option>
									</select>
								</div>
							</div>

							<div class="form-group">
								<label class="form-label" for="reg-username">Username</label>
								<div class="input-wrapper">
									<span class="input-icon">
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
									</span>
									<input class="auth-input" type="text" id="reg-username" name="username" placeholder="johndoe" required autocomplete="username">
								</div>
							</div>
							
							<div class="form-group">
								<label class="form-label" for="reg-email">Email Address</label>
								<div class="input-wrapper">
									<span class="input-icon">
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>
									</span>
									<input class="auth-input" type="email" id="reg-email" name="email" placeholder="john@example.com" required autocomplete="email">
								</div>
							</div>

							<div class="form-group">
								<label class="form-label" for="reg-password">Password</label>
								<div class="input-wrapper">
									<span class="input-icon">
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
									</span>
									<input class="auth-input" type="password" id="reg-password" name="password" placeholder="Min. 8 characters" required autocomplete="new-password">
								</div>
							</div>

							<button type="submit" class="auth-button">Create Account</button>
						</form>

						<p class="switch-auth">
							Already have an account? <a data-link href="/login">Sign In</a>
						</p>
					</div>
				</div>
				<div class="auth-hero">
					<div class="hero-content">
						<h2 class="hero-title">Start your journey.</h2>
						<div class="testimonial">
							<p class="testimonial-text">"Creating an account was the best thing I did for my professional network. The discussions are top-tier."</p>
							<div class="testimonial-author">
								<img class="author-avatar" src="https://i.pravatar.cc/100?u=sarah" alt="Sarah J.">
								<div class="author-info">
									<h4>Sarah Jenkins</h4>
									<p>Community Member</p>
								</div>
							</div>
						</div>
					</div>
					<div class="hero-footer">
						<div class="footer-label">Join 50K+ others</div>
					</div>
				</div>
			</div>
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
