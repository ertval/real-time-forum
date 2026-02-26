// /web/static/js/notifications.js

import { API_BASE } from "./utils.js";
import { uiNotify } from "./ui-messages.js";
import { playNotification } from "./sound-effects.js";

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
    const postId = n.post_id;
    const commentId = n.comment_id;

    const div = document.createElement("div");
    div.className = "notification-item";
    div.style.cursor = "pointer";

    if (!n.is_read) div.classList.add("unread");

    // HTML rendering enabled
    div.innerHTML = buildMessage(n);

    /* -------------------------
       CLICK REDIRECT + MARK READ
    -------------------------- */
    div.addEventListener("click", async () => {
      await markOneAsRead(n.id);

      if (postId) {
        if (n.type === "comment") {
          // New comment notification → highlight last comment
          window.location.href = `/view-post/${postId}?highlight=last`;
          return;
        }

        if (commentId) {
          // Comment reaction → highlight specific comment
          window.location.href = `/view-post/${postId}?highlight=${commentId}`;
          return;
        }

        // Post like/dislike
        window.location.href = `/view-post/${postId}`;
      }
    });

    list.appendChild(div);
  });
}

/* -------------------------
   MARK ONE AS READ
-------------------------- */
async function markOneAsRead(id) {
  try {
    await fetch(`${API_BASE}/notifications/${id}/read`, {
      method: "PATCH",
      credentials: "include",
    });
  } catch (err) {
    console.error("mark-one error:", err);
  }
}

/* -------------------------
   TOAST LOGIC
-------------------------- */
function processNewNotifications(notifications) {
  notifications.forEach(n => {

    if (!initialized) {
      seenNotificationIds.add(n.id);
      return;
    }

    if (!seenNotificationIds.has(n.id)) {
      seenNotificationIds.add(n.id);

      if (!n.is_read) {
        uiNotify(buildMessage(n), { type: "info", html: true });
        playNotification(); // Play sound only for NEW unread notifications
      }
    }
  });
}

/* -------------------------
   MESSAGE BUILDER
-------------------------- */
function buildMessage(n) {
  const truncate = (str, len = 20) => {
    if (!str) return "";
    return str.length > len ? str.slice(0, len) + "…" : str;
  };

  const title = `<strong>${truncate(n.post_title)}</strong>`;
  const excerpt = `<em>${truncate(n.comment_excerpt)}</em>`;

  switch (n.type) {
    case "post_like":
      return `${n.actor_username} liked your post: ${title} 👍`;

    case "post_dislike":
      return `${n.actor_username} disliked your post: ${title} 👎`;

    case "comment":
      return `${n.actor_username} commented ${excerpt} on ${title} 💬`;

    case "comment_like":
      return `${n.actor_username} liked your comment: ${excerpt} 👍`;

    case "comment_dislike":
      return `${n.actor_username} disliked your comment: ${excerpt} 👎`;

    default:
      return "New notification 🔔";
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