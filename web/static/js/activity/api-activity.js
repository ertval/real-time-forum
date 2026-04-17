// /static/js/activity/api-activity.js

import { API_BASE, toPositiveInt } from '../utils.js';

const DEFAULT_PAGE = 1;
const DEFAULT_PER_PAGE = 10;

export function getQueryState() {
	const params = new URLSearchParams(window.location.search);
	const rawStatus = (params.get('status') || 'all').toLowerCase();

	return {
		page: toPositiveInt(params.get('page')) || DEFAULT_PAGE,
		perPage: toPositiveInt(params.get('per_page')) || DEFAULT_PER_PAGE,
		status: rawStatus === 'draft' || rawStatus === 'published' ? rawStatus : 'all',
	};
}

export function setQueryState({ page, perPage, status }) {
	const params = new URLSearchParams();

	if (page > DEFAULT_PAGE) params.set('page', String(page));
	if (perPage !== DEFAULT_PER_PAGE) params.set('per_page', String(perPage));
	if (status && status !== 'all') params.set('status', status);

	const qs = params.toString();
	history.pushState(null, '', qs ? `?${qs}` : window.location.pathname);
}

export async function loadActivity({ page, perPage, status }) {
	const url = new URL(`${API_BASE}/users/activity`, window.location.origin);

	url.searchParams.set('page', String(page));
	url.searchParams.set('per_page', String(perPage));

	if (status && status !== 'all') {
		url.searchParams.set('status', status);
	}

	const res = await fetch(url.toString(), {
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});

	if (res.status === 401) {
		window.location.href = '/login';
		return null;
	}

	if (!res.ok) {
		throw new Error('Failed to load activity.');
	}

	return res.json();
}
