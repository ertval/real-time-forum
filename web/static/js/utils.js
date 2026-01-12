// web/static/js/utils.js

// --------------------------------------------------
// API
// --------------------------------------------------

export const API_BASE = "/api/v1";

// --------------------------------------------------
// DATE FORMATTER
// --------------------------------------------------

export function formatCreatedAt(iso) {
  if (!iso) return "";

  const d = new Date(iso);

  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");

  let hours = d.getHours();
  const minutes = String(d.getMinutes()).padStart(2, "0");

  const ampm = hours >= 12 ? "PM" : "AM";
  hours = hours % 12;
  hours = hours === 0 ? 12 : hours;
  hours = String(hours).padStart(2, "0");

  return `${year}-${month}-${day}, ${hours}:${minutes} ${ampm}`;
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

/* =========================
   Pagination (read only for now)
========================= */

export function getPaginationFromURL() {
  const params = new URLSearchParams(window.location.search);
  return {
    page: toPositiveInt(params.get("page")) || 1,
    perPage: toPositiveInt(params.get("per_page")) || 10,
  };
}

export function toPositiveInt(v) {
  const n = Number.parseInt(v, 10);
  return Number.isFinite(n) && n > 0 ? n : null;
}