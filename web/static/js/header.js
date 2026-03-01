// web/static/js/header.js

import { Auth } from "./auth.js";
import {
  startNotificationPolling,
  stopNotificationPolling,
  initNotificationBell,
} from "./notifications.js";

let pollingStarted = false;
let bellInitialized = false;

export async function initHeader() {
  await Auth.init();

  setupForumLogo();

  const greetingEl = document.getElementById("greeting");
  const loginBtn = document.getElementById("login-btn");
  const logoutBtn = document.getElementById("logout-btn");
  const createPostBtn = document.getElementById("create-post-btn");

  const authedOnlyEls = [
    document.getElementById("my-posts-btn"),
    document.getElementById("my-liked-posts-btn"),
    document.getElementById("my-activity-btn"),
  ].filter(Boolean);

  if (!greetingEl || !loginBtn || !logoutBtn) return;

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
  }

  /* -------------------------
     GREETING + BUTTONS
  -------------------------- */

  greetingEl.textContent = isAuthed
    ? `Hello, ${username}`
    : "Hello, Guest";

  greetingEl.style.display = "block";
  loginBtn.style.display = isAuthed ? "none" : "inline-flex";
  logoutBtn.style.display = isAuthed ? "inline-flex" : "none";

  setAuthedOnlyVisibility(authedOnlyEls, isAuthed);

  /* -------------------------
     LOGOUT
  -------------------------- */

  if (isAuthed) {
    bindLogout(logoutBtn);
  }

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

/* -------------------------
   LOGOUT
-------------------------- */

function bindLogout(logoutBtn) {
  if (!logoutBtn || logoutBtn.dataset.bound) return;

  logoutBtn.dataset.bound = "1";

  logoutBtn.addEventListener("click", async (e) => {
    e.preventDefault();

    await fetch("/api/v1/users/logout", {
      method: "POST",
      credentials: "include",
    });

    stopNotificationPolling();
    pollingStarted = false;

    Auth.checked = false;
    Auth.isAuthenticated = false;
    Auth.user = null;

    window.location.reload();
  });
}
