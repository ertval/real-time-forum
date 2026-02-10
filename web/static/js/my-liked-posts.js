//web/static/js/my-liked-posts.js
import { API_BASE, getPaginationFromURL } from "./utils.js";
import { renderPostCard, loadPostCommentsPreview } from "./posts.js";
import { initReactions } from "./reactions.js";

function start() {
    boot().catch((err) => {
        console.error("My Liked Posts boot failed:", err);
        showMessage("Failed to load your liked posts.");
    });
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", start);
} else {
    start();
}

async function boot() {
    const output = document.getElementById("posts-output");
    const empty = document.getElementById("posts-empty");
    if (!output) return;

    output.innerHTML = "";
    if (empty) empty.hidden = true;

    const { page, perPage } = getPaginationFromURL();

    const { posts } = await fetchLikedPosts({ page, perPage });

    if (!posts.length) {
        if (empty) empty.hidden = false;
        return;
    }

    const fragment = document.createDocumentFragment();
    const articles = [];

    for (const post of posts) {
        // IMPORTANT: disable owner-only controls
        const article = renderPostCard(post, {
            clickable: true,
            showStatusToggle: false,
            showDelete: false,
        });
        fragment.appendChild(article);
        articles.push(article);
    }

    output.appendChild(fragment);

    await runWithConcurrencyLimit(
        articles.map((article) => async () => {
            const postId = article.dataset.postId;
            if (!postId) return;
            await loadPostCommentsPreview(postId, article);
        }),
        4
    );

    initReactions();
}

/*-----
  API
-----*/

async function fetchLikedPosts({ page, perPage }) {
    const url = new URL(`${API_BASE}/posts/liked`, window.location.origin);
    url.searchParams.set("page", String(page));
    url.searchParams.set("per_page", String(perPage));

    const res = await fetch(url.toString(), {
        credentials: "include",
        headers: { Accept: "application/json" },
    });

    if (res.status === 401) {
        showMessage("You must be logged in to view liked posts.");
        return { posts: [], meta: null };
    }

    if (!res.ok) {
        showMessage(`Failed to load liked posts (${res.status}).`);
        return { posts: [], meta: null };
    }

    const payload = await res.json();
    const posts = Array.isArray(payload?.data) ? payload.data : [];
    const meta = payload?.meta ?? null;

    return { posts, meta };
}

/*------------
  UX HELPERS
------------*/

function showMessage(text) {
    const empty = document.getElementById("posts-empty");
    if (empty) {
        empty.textContent = text;
        empty.hidden = false;
    } else {
        alert(text);
    }
}

/*-------------------
  CONCURENCY HELPER
-------------------*/

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