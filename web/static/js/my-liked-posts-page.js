import { API_BASE } from "./utils.js";

async function loadHeader() {
    const res = await fetch("/static/partials/header.html");
    if (!res.ok) throw new Error(`Header load failed: ${res.status}`);
    const html = await res.text();

    const host = document.getElementById("header");
    if (!host) throw new Error("Missing #header container");
    host.innerHTML = html;
    // Hide my-liked-posts button if on my-liked-posts page
    document.getElementById("my-liked-posts-btn")?.remove();
}

function loadModuleScript(src) {
    const s = document.createElement("script");
    s.type = "module";
    s.src = src;
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

    // Auth check
    const isAuthed = await requireAuth();
    if (!isAuthed) return;

    // header logic login/logout/create-post visibility
    loadModuleScript("/static/js/header-loader.js");

    // page logic
    loadModuleScript("/static/js/my-liked-posts.js");
})().catch((err) => {
    console.error("My Liked Posts page bootstrap failed:", err);
    alert("Failed to load My Liked Posts page.");
});