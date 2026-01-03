// web/static/js/header.js


/**
 * Entry point
 * Called AFTER header.html is injected into DOM
 */
initHeader();

async function initHeader() {
  setupForumLogo();

  const greetingEl = document.getElementById("greeting");
  const loginBtn = document.getElementById("login-btn");
  const logoutBtn = document.getElementById("logout-btn");

  if (!greetingEl || !loginBtn || !logoutBtn) {
    console.warn("Header elements not found");
    return;
  }

  await loadGreeting(greetingEl, loginBtn, logoutBtn);
  bindLogout(logoutBtn);
}

// --------------------------------------------------
// FORUM LOGO → /home
// --------------------------------------------------

function setupForumLogo() {
  const forumTitle = document.getElementById("forum-title");
  if (!forumTitle) return;

  forumTitle.style.cursor = "pointer";
  forumTitle.addEventListener("click", () => {
    window.location.assign("/home");
  });
}

// --------------------------------------------------
// GREETING + AUTH STATE
// --------------------------------------------------

async function loadGreeting(greetingEl, loginBtn, logoutBtn) {
  try {
    const res = await fetch(`${API_BASE}/users/me`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    // Guest
    if (!res.ok) {
      greetingEl.textContent = "Hello, Guest";
      greetingEl.hidden = false;

      loginBtn.style.display = "inline-flex";
      logoutBtn.style.display = "none";
      return;
    }

    // Logged in
    const payload = await res.json();
    const username = payload?.data?.username;

    greetingEl.textContent = username
      ? `Hello, ${username}`
      : "Hello";

    greetingEl.hidden = false;
    loginBtn.style.display = "none";
    logoutBtn.style.display = "inline-flex";
  } catch {
    greetingEl.textContent = "Hello, Guest";
    greetingEl.hidden = false;

    loginBtn.style.display = "inline-flex";
    logoutBtn.style.display = "none";
  }
}

// --------------------------------------------------
// LOGOUT
// --------------------------------------------------

function bindLogout(logoutBtn) {
  logoutBtn.addEventListener("click", async (e) => {
    e.preventDefault();
    logoutBtn.disabled = true;

    try {
      await fetch(`${API_BASE}/users/logout`, {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      window.location.assign("/login");
    } catch (err) {
      console.error("Logout failed:", err);
      alert("Logout failed. Please try again.");
      logoutBtn.disabled = false;
    }
  });
}
