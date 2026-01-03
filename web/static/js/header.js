// web/static/js/header.js

/**
 * Entry point
 * Called AFTER header.html is injected into DOM
 */
initHeader();

async function initHeader() {
  const greetingEl = document.getElementById("greeting");
  const loginBtn = document.getElementById("login-btn");
  const logoutBtn = document.getElementById("logout-btn");

  // Safety check
  if (!greetingEl || !loginBtn || !logoutBtn) {
    console.warn("Header elements not found");
    return;
  }

  await loadGreeting(greetingEl, loginBtn, logoutBtn);
  bindLogout(logoutBtn);
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
  } catch (err) {
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
  logoutBtn.addEventListener("click", async () => {
    logoutBtn.disabled = true;

    try {
      await fetch(`${API_BASE}/users/logout`, {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      });

      // Always redirect – even if cookie already expired
      window.location.assign("/login");
    } catch (err) {
      console.error("Logout failed:", err);
      alert("Logout failed. Please try again.");
      logoutBtn.disabled = false;
    }
  });
}
