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
		try {
			for (let i = 0; i < route.paramKeys.length; i += 1) {
				params[route.paramKeys[i]] = decodeURIComponent(result[i + 1]);
			}
		} catch {
			continue;
		}

		return { route, params, path: targetPath };
	}

	return null;
}
