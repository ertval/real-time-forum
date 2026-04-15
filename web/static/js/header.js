// web/static/js/header.js

import { Auth } from "./auth.js";
import {
	initNotificationBell,
	startNotificationPolling,
	stopNotificationPolling,
} from "./notifications.js";

let pollingStarted = false;
let bellInitialized = false;

export async function initHeader() {
	setupForumLogo();

	const greetingEl = document.getElementById("greeting");
	const loginBtn = document.getElementById("login-btn");

	const createPostBtn = document.getElementById("create-post-btn");

	const authedOnlyEls = [document.getElementById("my-activity-btn")].filter(
		Boolean,
	);

	if (!greetingEl || !loginBtn) return;

	const isAuthed = Auth.isAuthenticated;
	const username = Auth.user?.username ?? "User";

	/* -------------------------
     NOTIFICATIONS
  -------------------------- */

	if (isAuthed) {
		if (!pollingStarted) {
			startNotificationPolling();
			pollingStarted = true;
		}

		if (!bellInitialized) {
			initNotificationBell();
			bellInitialized = true;
		}
	} else {
		if (pollingStarted) {
			stopNotificationPolling();
			pollingStarted = false;
		}

		bellInitialized = false;
	}

	/* -------------------------
     GREETING + BUTTONS
  -------------------------- */

	greetingEl.textContent = isAuthed
		? `Welcome back, ${username}`
		: "Welcome back, Guest";

	greetingEl.style.display = "block";
	loginBtn.style.display = isAuthed ? "none" : "inline-flex";

	setAuthedOnlyVisibility(authedOnlyEls, isAuthed);

	/* -------------------------
     CREATE POST
  -------------------------- */

	if (createPostBtn && !createPostBtn.dataset.bound) {
		createPostBtn.dataset.bound = "1";

		createPostBtn.addEventListener("click", async (e) => {
			e.preventDefault();

			const allowed = await Auth.requireOrPrompt();
			if (!allowed) return;

			window.location.assign("/create-post");
		});
	}
}

/* -------------------------
   AUTHEd VISIBILITY
-------------------------- */

function setAuthedOnlyVisibility(elements, isAuthed) {
	const displayValue = isAuthed ? "inline-flex" : "none";
	for (const el of elements) {
		el.style.display = displayValue;
	}
}

/* -------------------------
   FORUM LOGO
-------------------------- */

function setupForumLogo() {
	const forumTitle = document.getElementById("forum-title");
	if (!forumTitle) return;

	forumTitle.style.cursor = "pointer";

	if (!forumTitle.dataset.bound) {
		forumTitle.dataset.bound = "1";

		forumTitle.addEventListener("click", (e) => {
			e.preventDefault();
			window.location.assign("/");
		});
	}
}
