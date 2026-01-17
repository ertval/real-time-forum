// /web/static/js/pagination.js

export function createPagination({
  prevBtn,
  nextBtn,
  numbersEl,
  onPageChange,
  maxVisible = 5,
}) {
  let currentPage = 1;
  let totalPages = 1;

  function goToPage(page) {
    if (page < 1 || page > totalPages || page === currentPage) return;
    currentPage = page;
    onPageChange(page);
    render();
  }

  function addPageButton(page) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "pagination-btn";
    btn.textContent = String(page);

    if (page === currentPage) {
      btn.classList.add("is-active");
      btn.disabled = true;
    }

    btn.addEventListener("click", () => goToPage(page));
    numbersEl.appendChild(btn);
  }

  function addEllipsis() {
    const span = document.createElement("span");
    span.className = "pagination-ellipsis";
    span.textContent = "…";
    numbersEl.appendChild(span);
  }

  function render() {
    numbersEl.innerHTML = "";

    // Prev/Next buttons state
    prevBtn.disabled = currentPage <= 1;
    nextBtn.disabled = currentPage >= totalPages;

    // If few pages, show all
    if (totalPages <= maxVisible) {
      for (let p = 1; p <= totalPages; p++) addPageButton(p);
      return;
    }

    // We want:
    // 1 … [window] … totalPages
    // window size ~ maxVisible (but we also show 1 and totalPages always)
    const windowSize = Math.max(1, maxVisible - 2); // excluding first+last
    let start = currentPage - Math.floor(windowSize / 2);
    let end = start + windowSize - 1;

    // Clamp window into [2, totalPages-1]
    if (start < 2) {
      start = 2;
      end = start + windowSize - 1;
    }
    if (end > totalPages - 1) {
      end = totalPages - 1;
      start = end - windowSize + 1;
      if (start < 2) start = 2;
    }

    // First page
    addPageButton(1);

    // Left ellipsis (if gap)
    if (start > 2) addEllipsis();

    // Window pages
    for (let p = start; p <= end; p++) addPageButton(p);

    // Right ellipsis (if gap)
    if (end < totalPages - 1) addEllipsis();

    // Last page
    addPageButton(totalPages);
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
