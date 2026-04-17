// /static/js/activity/bootstrap-activity.js

import { Auth } from "../auth.js";
import { createPagination } from "../pagination.js";
import { uiNotify } from "../ui-messages.js";

import { getQueryState, setQueryState } from "./api-activity.js";
import { initDeleteComment, initEditComment } from "./comments-activity.js";
import {
	initDeletePost,
	initEditPostNavigation,
	initStatusToggle,
} from "./posts-activity.js";
import { renderActivity } from "./render-activity.js";
import { initSectionToggles, openSectionByHash } from "./sections-activity.js";
import { activityState } from "./state-activity.js";

export async function startActivityPage() {
	initSectionToggles();

	await Auth.init();
	activityState.activityUserID = Number(Auth.user?.id) || 0;

	const prevBtn = document.getElementById("activity-prev-page");
	const nextBtn = document.getElementById("activity-next-page");
	const numbersEl = document.getElementById("activity-page-numbers");

	if (!prevBtn || !nextBtn || !numbersEl) return;

	const state = getQueryState();

	const pager = createPagination({
		prevBtn,
		nextBtn,
		numbersEl,
		onPageChange: async (page) => {
			state.page = page;
			setQueryState(state);
			await refresh();
		},
	});

	const refresh = async () => {
		try {
			await renderActivity(state, pager);
			openSectionByHash();
		} catch (err) {
			console.error(err);
			uiNotify("Failed to load activity.", { type: "danger" });
		}
	};

	await refresh();

	initEditPostNavigation();
	initDeletePost(refresh);
	initStatusToggle(refresh);
	initEditComment(refresh);
	initDeleteComment(refresh);
}

if (document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", startActivityPage);
} else {
	startActivityPage();
}
