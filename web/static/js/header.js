// web/static/js/header.js

import { Auth } from "./auth.js";

export async function initHeader() {
  await Auth.init();

  setupForumLogo();

  const greetingEl = document.getElementById("greeting");
  const loginBtn = document.getElementById("login-btn");
  const logoutBtn = document.getElementById("logout-btn");

  const authedOnlyEls = [
    document.getElementById("my-posts-btn"),
    document.getElementById("my-liked-posts-btn"),
  ].filter(Boolean);

  if (!greetingEl || !loginBtn || !logoutBtn) return;

  const isAuthed = Auth.isAuthenticated;
  //fallback if auth.user is nullified
  const username = Auth.user?.username ?? "User";


  greetingEl.textContent = isAuthed ? `Hello, ${username}` : "Hello, Guest";
  loginBtn.style.display = isAuthed ? "none" : "inline-flex";
  logoutBtn.style.display = isAuthed ? "inline-flex" : "none";

  setAuthedOnlyVisibility(authedOnlyEls, isAuthed)

  if (isAuthed) {
    bindLogout(logoutBtn);
  }
}

function setAuthedOnlyVisibility(elements, isAuthed) {
  const displayValue = isAuthed ? "inline-flex" : "none";
  for (const el of elements) {
    el.style.display = displayValue;
  }
}

// --------------------------------------------------
// FORUM LOGO → /
// --------------------------------------------------

function setupForumLogo() {
  const forumTitle = document.getElementById("forum-title");
  if (!forumTitle) return;

  forumTitle.style.cursor = "pointer";
  forumTitle.addEventListener("click", () => {
    window.location.assign("/");
  });
}

// --------------------------------------------------
// LOGOUT
// --------------------------------------------------

function bindLogout(logoutBtn) {
  if (logoutBtn.dataset.bound) return;
  logoutBtn.dataset.bound = "1";

  logoutBtn.addEventListener("click", async (e) => {
    e.preventDefault();

    await fetch("/api/v1/users/logout", {
      method: "POST",
      credentials: "include",
    });

    window.location.assign("/login");
  });
}
