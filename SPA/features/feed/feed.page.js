import { createPagination } from '/static/js/pagination.js';
import { loadPostCommentsPreview, renderPostCard } from '/static/js/posts.js';
import { initReactions } from '/static/js/reactions.js';
import { API_BASE } from '/static/js/utils.js';

const DEFAULT_PER_PAGE = 10;
const ALLOWED_PER_PAGE_VALUES = new Set([0, 5, 10, 20]);
const FEED_BOUND_ATTR = 'data-feed-bound';

function toPositiveInt(value) {
	const parsed = Number.parseInt(value ?? '', 10);
	return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function getQueryState(windowRef) {
	const params = new URLSearchParams(windowRef.location.search);
	const page = toPositiveInt(params.get('page')) ?? 1;
	const requestedPerPage = Number.parseInt(params.get('per_page') ?? '', 10);
	const perPage = ALLOWED_PER_PAGE_VALUES.has(requestedPerPage)
		? requestedPerPage
		: DEFAULT_PER_PAGE;
	const categoryId = toPositiveInt(params.get('category_id'));

	return {
		page,
		perPage,
		categoryId: categoryId ? String(categoryId) : '',
	};
}

function setQueryState(windowRef, { page, perPage, categoryId }) {
	const url = new URL(windowRef.location.href);
	url.search = '';

	if (page > 1) {
		url.searchParams.set('page', String(page));
	}

	if (perPage !== DEFAULT_PER_PAGE) {
		url.searchParams.set('per_page', String(perPage));
	}

	if (categoryId) {
		url.searchParams.set('category_id', String(categoryId));
	}

	windowRef.history.pushState({}, '', `${url.pathname}${url.search}`);
}

async function loadCategories(fetchRef) {
	const response = await fetchRef(`${API_BASE}/categories`, {
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});

	if (!response.ok) {
		return [];
	}

	const payload = await response.json();
	return Array.isArray(payload) ? payload : (payload.data ?? []);
}

async function loadPosts(fetchRef, windowRef, { page, perPage, categoryId }) {
	const url = new URL(`${API_BASE}/posts`, windowRef.location.origin);
	url.searchParams.set('page', String(page));

	if (perPage > 0) {
		url.searchParams.set('per_page', String(perPage));
	}

	if (categoryId) {
		url.searchParams.set('category_id', categoryId);
	}

	const response = await fetchRef(url.toString(), {
		credentials: 'include',
		headers: { Accept: 'application/json' },
	});

	if (!response.ok) {
		return null;
	}

	return response.json();
}

function syncCategoryOptions(documentRef, select, categories, selectedCategoryId) {
	if (!select || select.tagName !== 'SELECT') {
		return;
	}

	select.querySelectorAll('option:not(:first-child)').forEach((option) => {
		option.remove();
	});

	for (const category of categories) {
		const option = documentRef.createElement('option');
		option.value = String(category.id);
		option.textContent = category.name;
		select.appendChild(option);
	}

	select.value = selectedCategoryId;
}

async function renderPosts(elements, state, pager, fetchRef, windowRef) {
	const { postsOutput, postsEmpty, pagination } = elements;

	if (!postsOutput || !postsEmpty || !pagination) {
		return;
	}

	postsOutput.innerHTML = '';
	postsEmpty.hidden = true;
	pagination.hidden = true;

	const payload = await loadPosts(fetchRef, windowRef, state);
	const posts = Array.isArray(payload?.data) ? payload.data : [];

	if (posts.length === 0) {
		postsEmpty.hidden = false;
		pager.set(1, 1);
		return;
	}

	for (const post of posts) {
		const card = renderPostCard(post, {
			onNavigate: (targetPost) => {
				windowRef.history.pushState({}, '', `/post/${targetPost.id}`);
				windowRef.dispatchEvent(new PopStateEvent('popstate'));
			},
		});
		postsOutput.appendChild(card);
		await loadPostCommentsPreview(post.id, card);
	}

	if (state.perPage === 0) {
		return;
	}

	const paginationMeta = payload?.meta?.pagination;
	if (paginationMeta && paginationMeta.total_pages > 1) {
		pagination.hidden = false;
		pager.set(paginationMeta.page, paginationMeta.total_pages);
		return;
	}

	pager.set(1, 1);
}

export function initFeedPage(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const fetchRef = options.fetchRef ?? (typeof fetch === 'function' ? fetch.bind(windowRef) : null);

	if (
		!windowRef ||
		!documentRef ||
		typeof fetchRef !== 'function' ||
		typeof documentRef.querySelector !== 'function'
	) {
		return null;
	}

	const root = documentRef.querySelector('[data-screen="feed"]');
	if (!root) {
		return null;
	}

	if (root.getAttribute(FEED_BOUND_ATTR) === 'true') {
		return null;
	}

	root.setAttribute(FEED_BOUND_ATTR, 'true');

	const elements = {
		root,
		categoryFilter: root.querySelector('#categoryFilter'),
		perPageSelect: root.querySelector('#perPageSelect'),
		postsOutput: root.querySelector('#posts-output'),
		postsEmpty: root.querySelector('#posts-empty'),
		pagination: root.querySelector('#pagination'),
		prevPage: root.querySelector('#prevPage'),
		nextPage: root.querySelector('#nextPage'),
		pageNumbers: root.querySelector('#pageNumbers'),
	};

	const state = getQueryState(windowRef);

	if (elements.perPageSelect) {
		elements.perPageSelect.value = String(state.perPage);
	}

	initReactions();

	const pager = createPagination({
		prevBtn: elements.prevPage,
		nextBtn: elements.nextPage,
		numbersEl: elements.pageNumbers,
		onPageChange: (page) => {
			state.page = page;
			setQueryState(windowRef, state);
			void renderPosts(elements, state, pager, fetchRef, windowRef);
		},
	});

	elements.prevPage?.addEventListener('click', () => pager.prev());
	elements.nextPage?.addEventListener('click', () => pager.next());

	elements.perPageSelect?.addEventListener('change', () => {
		const nextPerPage = Number.parseInt(elements.perPageSelect.value, 10);
		state.perPage = ALLOWED_PER_PAGE_VALUES.has(nextPerPage) ? nextPerPage : DEFAULT_PER_PAGE;
		state.page = 1;
		setQueryState(windowRef, state);
		void renderPosts(elements, state, pager, fetchRef, windowRef);
	});

	elements.categoryFilter?.addEventListener('change', () => {
		state.categoryId = elements.categoryFilter.value || '';
		state.page = 1;
		setQueryState(windowRef, state);
		void renderPosts(elements, state, pager, fetchRef, windowRef);
	});

	void (async () => {
		const categories = await loadCategories(fetchRef);
		syncCategoryOptions(documentRef, elements.categoryFilter, categories, state.categoryId);
		await renderPosts(elements, state, pager, fetchRef, windowRef);
	})();

	return { state, elements };
}
