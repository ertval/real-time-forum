// web/static/js/utils.js

// --------------------------------------------------
// API
// --------------------------------------------------

export const API_BASE = "/api/v1";

// --------------------------------------------------
// DATE FORMATTER
// --------------------------------------------------

export function formatCreatedAt(dateStr) {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  return d.toLocaleString();
}

// --------------------------------------------------
// SAFE TEXT
// --------------------------------------------------

export function escapeHTML(str) {
  if (!str) return "";
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}

// --------------------------------------------------
// USERNAME HELPER
// --------------------------------------------------

export function resolveUsername(obj) {
  return (
    obj.username ||
    obj.author ||
    (obj.user_id ? `User ${obj.user_id}` : "User")
  );
}
