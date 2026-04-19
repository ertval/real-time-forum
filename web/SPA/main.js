/**
 * Real-Time Forum
 * Main Entry Point
 */

// Initial App Boot Logic
async function boot() {
	console.log("RTF SPA booting...");

	try {
		// Simulate some initial work
		await new Promise((resolve) => setTimeout(resolve, 500));

		// Hide loading overlay
		const overlay = document.getElementById("loading-overlay");
		if (overlay) {
			overlay.style.opacity = "0";
			setTimeout(() => overlay.remove(), 300);
		}

		// Initialize standard shell elements if needed
		renderWelcome();
	} catch (error) {
		console.error("App boot failed:", error);
	}
}

function renderWelcome() {
	const mainContent = document.getElementById("main-content");
	if (mainContent) {
		mainContent.innerHTML = `
      <section style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; text-align: center;">
        <h1 style="font-size: 3rem; margin-bottom: 1rem; background: var(--accent-gradient); -webkit-background-clip: text; -webkit-text-fill-color: transparent;">
          Real-Time Forum
        </h1>
        <p style="color: var(--text-secondary); max-width: 600px; font-size: 1.25rem;">
          Welcome to the new era of community interaction. 
          The SPA shell is now active.
        </p>
      </section>
    `;
	}
}

// Start the app when DOM is ready
if (document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", boot);
} else {
	boot();
}
