// web/static/js/ui-messages.js

let toastRoot = null;
let activeConfirm = null;

const MAX_TOASTS = 4;

/* --------------------
   TOAST HELPERS
-------------------- */

function ensureToastRoot() {
	if (toastRoot) return toastRoot;

	toastRoot = document.querySelector(".ui-toasts");
	if (!toastRoot) {
		toastRoot = document.createElement("div");
		toastRoot.className = "ui-toasts";
		document.body.appendChild(toastRoot);
	}
	return toastRoot;
}

function typeToTitle(type) {
	switch (type) {
		case "success":
			return "Success";
		case "warn":
			return "Warning";
		case "danger":
			return "Error";
		default:
			return "Info";
	}
}

function getTypeClass(type) {
	switch (type) {
		case "success":
			return "ui-type-success";
		case "warn":
			return "ui-type-warn";
		case "danger":
			return "ui-type-danger";
		default:
			return "ui-type-info";
	}
}

/**
 * Toast notification (non-blocking).
 * @param {string} message
 * @param {{type?: "info"|"success"|"warn"|"danger", timeoutMs?: number}} opts
 */
export function uiNotify(message, opts = {}) {
	const type = opts.type || "info";
	const timeoutMs = Number.isFinite(opts.timeoutMs) ? opts.timeoutMs : 2600;

	const root = ensureToastRoot();

	// limit toast stacking
	while (root.children.length >= MAX_TOASTS) {
		root.firstChild.remove();
	}

	const toast = document.createElement("div");
	toast.className = "ui-toast";
	toast.setAttribute("role", "status");
	toast.setAttribute("aria-live", "polite");

	const bar = document.createElement("div");
	bar.className = `ui-toast__bar ${getTypeClass(type)}`;

	const content = document.createElement("div");
	content.className = "ui-toast__content";

	if (opts.html === true) {
		content.innerHTML = message;
	} else {
		content.textContent = message;
	}

	const closeBtn = document.createElement("button");
	closeBtn.className = "ui-toast__close";
	closeBtn.type = "button";
	closeBtn.setAttribute("aria-label", "Close");
	closeBtn.textContent = "×";

	closeBtn.addEventListener("click", () => toast.remove());

	toast.appendChild(bar);
	toast.appendChild(content);
	toast.appendChild(closeBtn);
	root.appendChild(toast);

	if (timeoutMs > 0) {
		window.setTimeout(() => {
			if (toast.isConnected) toast.remove();
		}, timeoutMs);
	}
}

/* --------------------
   CONFIRM MODAL
-------------------- */

/**
 * Modal confirm (blocking-like via Promise).
 * @param {string} message
 * @param {{
 *  type?: "info"|"success"|"warn"|"danger",
 *  title?: string,
 *  okText?: string,
 *  cancelText?: string
 * }} opts
 * @returns {Promise<boolean>}
 */
export function uiConfirm(message, opts = {}) {
	if (activeConfirm) return activeConfirm;

	const type = opts.type || "warn";
	const title = opts.title || typeToTitle(type);
	const okText = opts.okText || "OK";
	const cancelText = opts.cancelText || "Cancel";

	activeConfirm = new Promise((resolve) => {
		const overlay = document.createElement("div");
		overlay.className = "ui-overlay";
		overlay.setAttribute("role", "presentation");

		const dialog = document.createElement("div");
		dialog.className = "ui-dialog";
		dialog.setAttribute("role", "dialog");
		dialog.setAttribute("aria-modal", "true");
		dialog.setAttribute("aria-label", title);

		const header = document.createElement("div");
		header.className = "ui-dialog__header";

		const badge = document.createElement("div");
		badge.className = `ui-badge ${getTypeClass(type)}`;

		const h = document.createElement("h2");
		h.className = "ui-dialog__title";
		h.textContent = title;

		header.appendChild(badge);
		header.appendChild(h);

		const body = document.createElement("div");
		body.className = "ui-dialog__body";
		body.textContent = message;

		const footer = document.createElement("div");
		footer.className = "ui-dialog__footer";

		const cancelBtn = document.createElement("button");
		cancelBtn.className = "ui-btn ui-btn--ghost";
		cancelBtn.type = "button";
		cancelBtn.textContent = cancelText;

		const okBtn = document.createElement("button");
		okBtn.className = "ui-btn ui-btn--primary";
		okBtn.type = "button";
		okBtn.textContent = okText;

		function cleanup(result) {
			document.removeEventListener("keydown", onKeydown);
			overlay.remove();
			activeConfirm = null;
			resolve(result);
		}

		function onKeydown(e) {
			if (!overlay.isConnected) return;
			if (e.key === "Escape") cleanup(false);
			if (e.key === "Enter") cleanup(true);
		}

		cancelBtn.addEventListener("click", () => cleanup(false));
		okBtn.addEventListener("click", () => cleanup(true));

		overlay.addEventListener("click", (e) => {
			if (e.target === overlay) cleanup(false);
		});

		footer.appendChild(cancelBtn);
		footer.appendChild(okBtn);

		dialog.appendChild(header);
		dialog.appendChild(body);
		dialog.appendChild(footer);

		overlay.appendChild(dialog);
		document.body.appendChild(overlay);

		document.addEventListener("keydown", onKeydown);

		// focus OK by default
		okBtn.focus();
	});

	return activeConfirm;
}
