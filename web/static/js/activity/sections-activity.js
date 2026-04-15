// /static/js/activity/sections-activity.js

let sectionToggleBound = false;

export function initSectionToggles() {
	if (sectionToggleBound) return;
	sectionToggleBound = true;

	const sections = document.querySelectorAll(".activity-section");

	for (const section of sections) {
		const head = section.querySelector(".activity-section-head");
		const toggle = section.querySelector("[data-activity-toggle]");

		if (!head || !toggle) continue;

		const controls = toggle.getAttribute("aria-controls");
		const content = controls ? document.getElementById(controls) : null;
		if (!content) continue;

		const isExpanded = toggle.getAttribute("aria-expanded") === "true";
		content.hidden = !isExpanded;

		const toggleSection = () => {
			const expanded = toggle.getAttribute("aria-expanded") === "true";
			const nextExpanded = !expanded;
			toggle.setAttribute("aria-expanded", String(nextExpanded));
			content.hidden = !nextExpanded;
		};

		head.addEventListener("click", (e) => {
			if (e.target.closest("[data-activity-toggle]")) return;
			toggleSection();
		});

		toggle.addEventListener("click", (e) => {
			e.stopPropagation();
			toggleSection();
		});
	}
}

export function openSectionByHash() {
	const hash = window.location.hash.replace("#", "");
	if (!hash) return;

	document.querySelectorAll("[data-activity-toggle]").forEach((toggle) => {
		const controls = toggle.getAttribute("aria-controls");
		const content = controls ? document.getElementById(controls) : null;
		if (!content) return;
		toggle.setAttribute("aria-expanded", "false");
		content.hidden = true;
	});

	const map = {
		created: "created-section-content",
		comments: "comments-section-content",
		liked: "liked-section-content",
		disliked: "disliked-section-content",
	};

	const targetId = map[hash];
	if (!targetId) return;

	const content = document.getElementById(targetId);
	if (!content) return;

	const toggle = document.querySelector(`[aria-controls="${targetId}"]`);

	if (toggle) toggle.setAttribute("aria-expanded", "true");
	content.hidden = false;

	content.scrollIntoView({
		behavior: "smooth",
		block: "start",
	});
}
