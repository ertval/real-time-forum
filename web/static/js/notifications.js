// web/static/js/notifications.js

import { API_BASE } from "./utils.js";
import { uiNotify } from "./ui-messages.js";

let seenNotificationIds = new Set();
let pollInterval = null;
let initialized = false;

/* -------------------------
   START POLLING
-------------------------- */

export function startNotificationPolling() {
  if (pollInterval) return;

  resetState();

  fetchNotifications();
  pollInterval = setInterval(fetchNotifications, 5000);
}

/* -------------------------
   STOP POLLING
-------------------------- */

export function stopNotificationPolling() {
  if (!pollInterval) return;

  clearInterval(pollInterval);
  pollInterval = null;

  resetState();
}

/* -------------------------
   RESET STATE
-------------------------- */

function resetState() {
  seenNotificationIds.clear();
  initialized = false;
}

/* -------------------------
   FETCH
-------------------------- */

async function fetchNotifications() {
  try {
    const res = await fetch(`${API_BASE}/notifications`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (!res.ok) return;

    const json = await res.json();

    const notifications = json.data?.notifications || [];
    const unreadCount = json.data?.unread_count ?? 0;

    updateBadge(unreadCount);
    renderDropdown(notifications);
    processNewNotifications(notifications);

    initialized = true;

  } catch (err) {
    console.error("notifications error:", err);
  }
}

/* -------------------------
   BADGE
-------------------------- */

function updateBadge(count) {
  const badge = document.getElementById("notification-badge");
  if (!badge) return;

  if (count > 0) {
    badge.textContent = count;
    badge.classList.remove("hidden");
  } else {
    badge.classList.add("hidden");
  }
}

/* -------------------------
   DROPDOWN
-------------------------- */

function renderDropdown(notifications) {
  const list = document.getElementById("notification-list");
  if (!list) return;

  list.innerHTML = "";

  notifications.forEach(n => {
    const div = document.createElement("div");
    div.className = "notification-item";

    if (!n.is_read) {
      div.classList.add("unread");
    }

    div.textContent = buildMessage(n);
    list.appendChild(div);
  });
}

/* -------------------------
   TOAST LOGIC
-------------------------- */

function processNewNotifications(notifications) {
  notifications.forEach(n => {

    // First load → just register existing IDs
    if (!initialized) {
      seenNotificationIds.add(n.id);
      return;
    }

    if (!seenNotificationIds.has(n.id)) {
      seenNotificationIds.add(n.id);

      if (!n.is_read) {
        uiNotify(buildMessage(n), { type: "info" });
      }
    }
  });
}

/* -------------------------
   MESSAGE BUILDER
-------------------------- */

function buildMessage(n) {
  switch (n.type) {
    case "post_like":
      return "Someone liked your post ❤️"

    case "post_dislike":
      return "Someone disliked your post"

    case "comment":
      return "New comment on your post 💬"

    case "comment_like":
      return "Someone liked your comment ❤️"

    case "comment_dislike":
      return "Someone disliked your comment"

    default:
      return "New notification"
  }
}

/* -------------------------
   BELL CLICK HANDLER
-------------------------- */

export function initNotificationBell() {
  const bell = document.getElementById("notification-bell");
  const dropdown = document.getElementById("notification-dropdown");

  if (!bell || !dropdown) return;

  bell.addEventListener("click", async (e) => {
    e.stopPropagation();
    dropdown.classList.toggle("hidden");

    if (!dropdown.classList.contains("hidden")) {
      await markAllAsRead();
      updateBadge(0);
    }
  });

  // CLICK OUTSIDE
  document.addEventListener("click", (e) => {
    if (!dropdown.contains(e.target) && !bell.contains(e.target)) {
      dropdown.classList.add("hidden");
    }
  });
}

async function markAllAsRead() {
  try {
    await fetch(`${API_BASE}/notifications/read-all`, {
      method: "PATCH",
      credentials: "include",
    });
  } catch (err) {
    console.error("mark-all error:", err);
  }
}