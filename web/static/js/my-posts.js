import { API_BASE, getPaginationFromURL } from "./utils.js";
import { renderPostCard, loadPostCommentsPreview, initReactions } from "./posts.js";

document.addEventListener("DOMContentLoaded", () => {
    boot().catch((err) => {
        console.error("My Posts boot failed:", err);
        showMessage("Failed to load your posts.");
    });
});

async function boot() {
    const output = document.getElementById("posts-output");
    const empty = document.getElementById("posts-empty");
    if (!output) return;

    // reset UI
    output.innerHTML = "";
    if (empty) empty.hidden = true;

    const { page, perPage } = getPaginationFromURL();

    const { posts } = await fetchMyPosts({ page, perPage });

    if (!posts.length) {
        if (empty) empty.hidden = false;
        return;
    }

    // Render post cards
    const fragment = document.createDocumentFragment();
    const articles = [];

    for (const post of posts) {
        const article = renderPostCard(post, { clickable: true });
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
}

/* =========================
   API
========================= */

async function fetchMyPosts({ page, perPage }) {
    const url = new URL(`${API_BASE}/posts/mine`, window.location.origin);
    url.searchParams.set("page", String(page));
    url.searchParams.set("per_page", String(perPage));

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