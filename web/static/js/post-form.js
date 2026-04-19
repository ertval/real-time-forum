// web/static/js/post-form.js
import { API_BASE } from './utils.js';

export async function loadCategories() {
	try {
		const res = await fetch(`${API_BASE}/categories`, {
			credentials: 'include',
			headers: { Accept: 'application/json' },
		});
		if (!res.ok) return [];
		const payload = await res.json().catch(() => null);
		if (Array.isArray(payload?.data)) return payload.data;
		return Array.isArray(payload) ? payload : [];
	} catch {
		return [];
	}
}

export async function renderCategoryCheckboxes(host) {
	if (!(host instanceof HTMLElement)) return;
	const categories = await loadCategories();
	host.innerHTML = '';

	categories.forEach((category) => {
		const label = document.createElement('label');
		label.className = 'category-checkbox';

		const input = document.createElement('input');
		input.type = 'checkbox';
		input.value = String(category.id);

		const text = document.createElement('span');
		text.textContent = category.name;

		label.appendChild(input);
		label.appendChild(text);
		host.appendChild(label);
	});
}

export function getSelectedCategoryIds(host) {
	if (!(host instanceof HTMLElement)) return [];
	return Array.from(host.querySelectorAll("input[type='checkbox']:checked")).map((el) =>
		Number(el.value),
	);
}

export function setSelectedCategoryIds(host, ids = []) {
	if (!(host instanceof HTMLElement)) return;
	const selected = new Set((ids || []).map(Number));
	host.querySelectorAll("input[type='checkbox']").forEach((cb) => {
		cb.checked = selected.has(Number(cb.value));
	});
}

export function normalizeCategoryIDs(categories = []) {
	return categories
		.map((category) => Number(category?.id))
		.filter((id) => Number.isFinite(id) && id > 0)
		.sort((a, b) => a - b);
}

export function sameNumberSet(a = [], b = []) {
	if (a.length !== b.length) return false;
	for (let i = 0; i < a.length; i += 1) {
		if (a[i] !== b[i]) return false;
	}
	return true;
}
