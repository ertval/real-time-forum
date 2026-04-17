// /web/static/js/pagination.js

export function createPagination({ prevBtn, nextBtn, numbersEl, onPageChange, maxVisible = 5 }) {
	let currentPage = 1;
	let totalPages = 1;

	function goToPage(page) {
		if (page < 1 || page > totalPages || page === currentPage) return;
		currentPage = page;
		onPageChange(page);
		render();
	}

	function addPageButton(page) {
		const btn = document.createElement('button');
		btn.type = 'button';
		btn.className = 'pagination-btn';
		btn.textContent = String(page);

		if (page === currentPage) {
			btn.classList.add('is-active');
			btn.disabled = true;
		}

		btn.addEventListener('click', () => goToPage(page));
		numbersEl.appendChild(btn);
	}

	function addEllipsis() {
		const span = document.createElement('span');
		span.className = 'pagination-ellipsis';
		span.textContent = '…';
		numbersEl.appendChild(span);
	}

	function renderAllPages() {
		for (let p = 1; p <= totalPages; p++) addPageButton(p);
	}

	function renderPagedWindow() {
		const windowSize = Math.max(1, maxVisible - 2);
		let start = currentPage - Math.floor(windowSize / 2);
		let end = start + windowSize - 1;

		if (start < 2) {
			start = 2;
			end = start + windowSize - 1;
		}

		if (end > totalPages - 1) {
			end = totalPages - 1;
			start = end - windowSize + 1;
			if (start < 2) start = 2;
		}

		addPageButton(1);

		if (start > 2) addEllipsis();

		for (let p = start; p <= end; p++) addPageButton(p);

		if (end < totalPages - 1) addEllipsis();

		addPageButton(totalPages);
	}

	function render() {
		numbersEl.innerHTML = '';

		prevBtn.disabled = currentPage <= 1;
		nextBtn.disabled = currentPage >= totalPages;

		if (totalPages <= maxVisible) {
			renderAllPages();
			return;
		}

		renderPagedWindow();
	}

	return {
		set(page, total) {
			currentPage = page;
			totalPages = total;
			render();
		},
		next() {
			if (currentPage < totalPages) goToPage(currentPage + 1);
		},
		prev() {
			if (currentPage > 1) goToPage(currentPage - 1);
		},
	};
}
