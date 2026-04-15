//web/static/js/logout.js
(() => {
	const btn = document.querySelector("[data-logout-button]");

	// optional: error surface (same approach as register.js)
	let errorEl = document.querySelector(".auth-error");
	if (!errorEl) {
		errorEl = document.createElement("div");
		errorEl.className = "auth-error";
		errorEl.setAttribute("role", "alert");
		// append somewhere sensible
		document.body.prepend(errorEl);
	}

	function showError(msg) {
		errorEl.textContent = msg || "";
		errorEl.style.display = msg ? "block" : "none";
	}

	async function doLogout(logoutUrl) {
		showError("");

		const res = await fetch(logoutUrl, {
			method: "POST",
			headers: { Accept: "application/json" },
			credentials: "include",
		});

		// Case where cookie has expired but frontend still thinks user is logged in
		if (res.ok || res.status === 401) {
			window.location.assign("/login");
			return;
		}

		const data = await res.json().catch(() => null);
		showError(data?.error?.message || "Logout failed.");
	}

	if (btn) {
		const logoutUrl =
			btn.getAttribute("data-logout-url") || "/api/v1/users/logout";

		btn.addEventListener("click", async (e) => {
			e.preventDefault();
			btn.disabled = true;

			try {
				await doLogout(logoutUrl);
			} catch {
				showError("Network error. Please try again.");
			} finally {
				btn.disabled = false;
			}
		});
	}
})();
