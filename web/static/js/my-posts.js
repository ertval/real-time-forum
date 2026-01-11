import { API_BASE, getPaginationFromURL } from "./utils.js";
import { renderPostCard, loadPostCommentsPreview, initReactions } from "./posts.js";

let statusToggleBound = false;
let deleteBound = false;

function start() {
    initStatusFilterUI();
    initStatusToggle();
    initDeletePost(); // if you added delete
    boot().catch((err) => {
        console.error("My Posts boot failed:", err);
        showMessage("Failed to load your posts.");
    });
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", start);
} else {
    start(); // DOM already loaded (your current situation)
}

function getStatusFilterFromURL() {
    const params = new URLSearchParams(window.location.search);
    const v = (params.get("status") || "all").toLowerCase();
    return v === "draft" || v === "published" ? v : "all";
}

function setStatusFilterInURL(status) {
    const url = new URL(window.location.href);
    const params = url.searchParams;

    // changing filter resets pagination
    params.set("page", "1");

    if (!status || status === "all") {
        params.delete("status");
    } else {
        params.set("status", status);
    }

    history.replaceState({}, "", url.toString());
}

function initStatusFilterUI() {
    const select = document.getElementById("status-filter");
    if (!select) return;

    select.value = getStatusFilterFromURL();

    // on change updates URL and re-runs boot()
    select.addEventListener("change", () => {
        const status = select.value;
        setStatusFilterInURL(status);
        boot().catch((err) => {
            console.error("My Posts boot failed:", err);
            showMessage("Failed to load your posts.");
        });
    });
}

async function boot() {
    const output = document.getElementById("posts-output");
    const empty = document.getElementById("posts-empty");
    if (!output) return;

    // reset UI
    output.innerHTML = "";
    if (empty) empty.hidden = true;

    const { page, perPage } = getPaginationFromURL();
    const status = getStatusFilterFromURL();

    const { posts } = await fetchMyPosts({ page, perPage, status });

    if (!posts.length) {
        if (empty) empty.hidden = false;
        return;
    }

    // Render post cards
    const fragment = document.createDocumentFragment();
    const articles = [];

    for (const post of posts) {
        const article = renderPostCard(post, { clickable: true, showStatusToggle: true, showDelete: true });
        fragment.appendChild(article);
        articles.push(article);
    }

    output.appendChild(fragment);

    // Load comments preview
    await runWithConcurrencyLimit(
        articles.map((article) => async () => {
            const postId = article.dataset.postId;
            if (!postId) return;
            await loadPostCommentsPreview(postId, article);
        }),
        4
    );

    // Bind like/dislike click handler
    initReactions();

    // Bind draft/publish toggle handler (my-posts only)
    initStatusToggle();
}

function initStatusToggle() {
    if (statusToggleBound) return;
    statusToggleBound = true;

    document.addEventListener("click", async (e) => {
        const btn = e.target.closest(".post-status-toggle");
        if (!btn) return;

        e.preventDefault();
        e.stopPropagation();

        const postId = btn.dataset.postId;
        const currentStatus = btn.dataset.currentStatus;
        if (!postId || !currentStatus) return;

        // determine next status
        const nextStatus = currentStatus === "draft" ? "published" : "draft";

        // disable button while request runs
        const originalText = btn.textContent;
        btn.disabled = true;

        try {
            const res = await fetch(`${API_BASE}/posts/${postId}`, {
                method: "PATCH",
                credentials: "include",
                headers: {
                    "Content-Type": "application/json",
                    Accept: "application/json",
                },
                body: JSON.stringify({ status: nextStatus }),
            });

            if (res.status === 401) {
                alert("You must be logged in.");
                return;
            }

            if (!res.ok) {
                const text = await res.text().catch(() => "");
                console.error("Status update failed:", res.status, text);
                alert("Failed to update post status.");
                return;
            }
            // dynamically updates on draft/publish button press
            await boot();
        } catch (err) {
            console.error("Status update request failed:", err);
            alert("Failed to update post status.");
        }
    }, true);
}

function initDeletePost() {
    if (deleteBound) return;
    deleteBound = true;

    document.addEventListener(
        "click",
        async (e) => {
            const btn = e.target.closest(".post-delete");
            if (!btn) return;

            e.preventDefault();
            e.stopPropagation();

            const postId = btn.dataset.postId;
            if (!postId) return;

            const ok = confirm("Delete this post? This cannot be undone.");
            if (!ok) return;

            btn.disabled = true;

            try {
                const res = await fetch(`${API_BASE}/posts/${postId}`, {
                    method: "DELETE",
                    credentials: "include",
                    headers: { Accept: "application/json" },
                });

                if (res.status === 401) {
                    alert("You must be logged in.");
                    return;
                }

                if (res.status === 404) {
                    alert("Post not found (it may have already been deleted).");
                    await boot();
                    return;
                }

                if (!res.ok) {
                    const text = await res.text().catch(() => "");
                    console.error("Delete failed:", res.status, text);
                    alert("Failed to delete post.");
                    return;
                }

                // re-render so it disappears and filter stays correct
                await boot();
            } catch (err) {
                console.error("Delete request failed:", err);
                alert("Failed to delete post.");
            } finally {
                if (document.contains(btn)) btn.disabled = false;
            }
        },
        true
    );
}

/* =========================
   API
========================= */

async function fetchMyPosts({ page, perPage, status }) {
    const url = new URL(`${API_BASE}/posts/mine`, window.location.origin);
    url.searchParams.set("page", String(page));
    url.searchParams.set("per_page", String(perPage));

    if (status && status !== "all") {
        url.searchParams.set("status", status);
    }

    const res = await fetch(url.toString(), {
        credentials: "include",
        headers: { Accept: "application/json" },
    });

    if (res.status === 401) {
        showMessage("You must be logged in to view your posts.");
        return { posts: [], meta: null };
    }

    if (!res.ok) {
        showMessage(`Failed to load posts (${res.status}).`);
        return { posts: [], meta: null };
    }

    const payload = await res.json();

    const posts = Array.isArray(payload?.data) ? payload.data : [];
    const meta = payload?.meta ?? null;

    return { posts, meta };
}

/* =========================
   UX helpers
========================= */

function showMessage(text) {
    const empty = document.getElementById("posts-empty");
    if (empty) {
        empty.textContent = text;
        empty.hidden = false;
    } else {
        alert(text);
    }
}

/* =========================
   Concurrency helper
========================= */

async function runWithConcurrencyLimit(tasks, limit = 4) {
    const queue = tasks.slice();

    const workers = Array.from({ length: limit }, async () => {
        while (queue.length) {
            const task = queue.shift();
            if (!task) return;
            try {
                await task();
            } catch (e) {
                console.error("Task failed:", e);
            }
        }
    });

    await Promise.all(workers);
}