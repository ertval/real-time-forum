function buildStatusFilter() {
    const label = document.createElement("label");
    label.className = "muted";
    label.setAttribute("for", "status-filter");
    label.textContent = "Status";

    const select = document.createElement("select");
    select.id = "status-filter";
    select.className = "input status-filter";
    select.innerHTML = `
    <option value="all">All</option>
    <option value="draft">Draft</option>
    <option value="published">Published</option>
  `;

    return { label, select };
}

async function loadHeader() {
    const res = await fetch("/static/partials/header.html");
    if (!res.ok) throw new Error(`Header load failed: ${res.status}`);
    const html = await res.text();

    const host = document.getElementById("header");
    if (!host) throw new Error("Missing #header container");
    host.innerHTML = html;

    const actions = document.querySelector(".page-actions");
    if (!actions) throw new Error("Header missing .page-actions");

    const { label, select } = buildStatusFilter();

    // Insert BEFORE Create Post button if it can be found
    const createBtn =
        actions.querySelector('#create-post-btn') ||
        actions.querySelector('a[href="/create-post"]') ||
        actions.querySelector('a[href="/create-post/"]');

    if (createBtn) {
        actions.insertBefore(label, createBtn);
        actions.insertBefore(select, createBtn);
    } else {
        actions.appendChild(label);
        actions.appendChild(select);
    }
}

function loadModuleScript(src) {
    const s = document.createElement("script");
    s.type = "module";
    s.src = src;
    document.body.appendChild(s);
}

(async function bootstrap() {
    await loadHeader();

    // header logic login/logout/create-post visibility
    loadModuleScript("/static/js/header-loader.js");

    // page logic expects status-filter to exist
    loadModuleScript("/static/js/my-posts.js");
})().catch((err) => {
    console.error("My Posts page bootstrap failed:", err);
    alert("Failed to load My Posts page.");
});