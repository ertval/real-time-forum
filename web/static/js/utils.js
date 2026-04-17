// web/static/js/utils.js

/*------
  API
------*/

export const API_BASE = '/api/v1';
export const MAX_IMAGE_BYTES = 20 * 1024 * 1024;
export const IMAGE_ACCEPT_MIME_TYPES = ['image/jpeg', 'image/png', 'image/gif'];
export const IMAGE_ACCEPT_ATTR = IMAGE_ACCEPT_MIME_TYPES.join(',');

export function buildImageRequestOptions({
	method = 'POST',
	credentials = 'include',
	imageFile = null,
	buildMultipartBody = () => null,
	jsonBody = {},
	multipartHeaders = {},
	jsonHeaders = {},
} = {}) {
	const hasImage = !!imageFile;
	const options = { method, credentials };

	if (hasImage) {
		options.body = buildMultipartBody(imageFile);
		if (multipartHeaders && Object.keys(multipartHeaders).length > 0) {
			options.headers = multipartHeaders;
		}
		return options;
	}

	options.headers = {
		'Content-Type': 'application/json',
		...jsonHeaders,
	};
	options.body = JSON.stringify(jsonBody);
	return options;
}

export function buildPostMultipartFormData({
	title = '',
	body = '',
	categoryIds = [],
	imageFile = null,
	manual,
	imageURL = null,
} = {}) {
	const formData = new FormData();
	formData.append('title', title);
	formData.append('body', body);
	categoryIds.forEach((id) => formData.append('category_ids', String(id)));

	if (imageFile) {
		formData.append('image', imageFile);
	}

	if (typeof manual !== 'undefined') {
		formData.append('manual', String(Boolean(manual)));
	}

	if (typeof imageURL === 'string' && imageURL.trim()) {
		formData.append('image_url', imageURL);
	}

	return formData;
}

/*----------------
  DATE FORMATTER
----------------*/

export function formatCreatedAt(iso) {
	if (!iso) return '';

	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) return String(iso);

	const year = d.getFullYear();
	const month = String(d.getMonth() + 1).padStart(2, '0');
	const day = String(d.getDate()).padStart(2, '0');

	let hours = d.getHours();
	const minutes = String(d.getMinutes()).padStart(2, '0');

	const ampm = hours >= 12 ? 'PM' : 'AM';
	hours = hours % 12;
	hours = hours === 0 ? 12 : hours;
	hours = String(hours).padStart(2, '0');

	return `${year}-${month}-${day}, ${hours}:${minutes} ${ampm}`;
}

export function escapeHTML(str) {
	if (!str) return '';
	const div = document.createElement('div');
	div.textContent = str;
	return div.innerHTML;
}

/*------------------
  USERNAME HELPEPR
------------------*/

export function resolveUsername(obj) {
	return obj.username || obj.author || (obj.user_id ? `User ${obj.user_id}` : 'User');
}

/*-------------
  PAGINATION
-------------*/

export function getPaginationFromURL() {
	const params = new URLSearchParams(window.location.search);
	return {
		page: toPositiveInt(params.get('page')) || 1,
		perPage: toPositiveInt(params.get('per_page')) || 10,
	};
}

export function toPositiveInt(v) {
	const n = Number.parseInt(v, 10);
	return Number.isFinite(n) && n > 0 ? n : null;
}
