const API_BASE = "/api/v1";

document.addEventListener("DOMContentLoaded", () => {
  setupForumLogo();
  loadGreeting();
});

function setupForumLogo() {
  const forumTitle = document.getElementById("forum-title");
  if (!forumTitle) return;

  forumTitle.style.cursor = "pointer";
  forumTitle.addEventListener("click", () => {
    window.location.assign("/home");
  });
}

async function loadGreeting() {
  const greetingEl = document.getElementById("greeting");
  if (!greetingEl) return;

  try {
    const res = await fetch(`${API_BASE}/users/me`, {
      credentials: "include",
    });

    if (res.ok) {
      const payload = await res.json();
      const username = payload?.data?.username;
      greetingEl.textContent = username
        ? `Hello, ${username}`
        : "Hello";
    } else {
      greetingEl.textContent = "Hello";
    }
  } catch {
    greetingEl.textContent = "Hello";
  }

  greetingEl.hidden = false;
}
