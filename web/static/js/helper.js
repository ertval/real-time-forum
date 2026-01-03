// web/static/js/helper.js

const API_BASE = "/api/v1";

/**
 * Format ISO date string to:
 * YYYY-MM-DD, HH:MM AM/PM
 */
function formatCreatedAt(iso) {
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
