// web/static/js/header.js

import { Auth } from "./auth.js";

export async function initHeader() {
  await Auth.init();

  setupForumLogo();

  const greetingEl = document.getElementById("greeting");
  const loginBtn = document.getElementById("login-btn");
  const logoutBtn = document.getElementById("logout-btn");
  const myPostsBtn = document.getElementById("my-posts-btn");

  if (!greetingEl || !loginBtn || !logoutBtn) return;

  if (!Auth.isAuthenticated) {
    greetingEl.textContent = "Hello, Guest";
    loginBtn.style.display = "inline-flex";
    logoutBtn.style.display = "none";
    if (myPostsBtn) myPostsBtn.style.display = "none";
    return;
  }

  greetingEl.textContent = `Hello, ${Auth.user.username}`;
  loginBtn.style.display = "none";
  logoutBtn.style.display = "inline-flex";
  if (myPostsBtn) myPostsBtn.style.display = "inline-flex";

  bindLogout(logoutBtn);
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
