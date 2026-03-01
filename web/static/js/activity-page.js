// web/static/js/activity-page.js
import { API_BASE } from "./utils.js";

async function loadHeader() {
  const res = await fetch("/static/partials/header.html");
  if (!res.ok) throw new Error(`Header load failed: ${res.status}`);
  const html = await res.text();

  const host = document.getElementById("header");
  if (!host) throw new Error("Missing #header container");
  host.innerHTML = html;

  // Hide current page button.
  document.getElementById("my-activity-btn")?.remove();
}

function loadModuleScript(src, version = "") {
  const s = document.createElement("script");
  s.type = "module";
  if (version) {
    const sep = src.includes("?") ? "&" : "?";
    s.src = `${src}${sep}v=${encodeURIComponent(version)}`;
  } else {
    s.src = src;
  }
  document.body.appendChild(s);
}

async function requireAuth() {
  const meRes = await fetch(`${API_BASE}/users/me`, {
    credentials: "include",
  });

  if (meRes.status === 401) {
    window.location.href = "/login";
    return false;
  }

  if (!meRes.ok) {
    throw new Error("Auth check failed");
  }

  return true;
}

(async function bootstrap() {
  await loadHeader();

  const isAuthed = await requireAuth();
  if (!isAuthed) return;

  const version = String(Date.now());
  loadModuleScript("/static/js/header-loader.js", version);
  loadModuleScript("/static/js/activity.js", version);
})().catch((err) => {
  console.error("Activity page bootstrap failed:", err);
  alert("Failed to load activity page.");
});
