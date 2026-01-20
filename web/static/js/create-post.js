// web/static/js/create-post.js

import { API_BASE } from "./utils.js";

/* =========================
   HELPERS
========================= */

function extractArray(payload) {
  if (Array.isArray(payload)) return payload;
  if (payload && Array.isArray(payload.data)) return payload.data;
  return [];
}

/* =========================
   LOAD CATEGORIES
========================= */

async function loadCategories() {
  try {
    const res = await fetch(`${API_BASE}/categories`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (!res.ok) return [];
    return await res.json();
  } catch {
    return [];
  }
}

async function populateCategories() {
  const select = document.getElementById("categorySelect");
  if (!select) return;

  const payload = await loadCategories();
  const categories = extractArray(payload);

  select.querySelectorAll("option:not(:first-child)").forEach(o => o.remove());

  categories.forEach(c => {
    const opt = document.createElement("option");
    opt.value = c.id;
    opt.textContent = c.name;
    select.appendChild(opt);
  });
}

/* =========================
   DRAFT AUTO-SAVE
========================= */

let draftTimer = null;
let autosaveEnabled = true;

function scheduleDraftSave() {
  if (!autosaveEnabled) return;
  clearTimeout(draftTimer);
  draftTimer = setTimeout(saveDraft, 1000);
}

async function saveDraft() {
  if (!autosaveEnabled) return;

  const title = document.getElementById("title")?.value.trim();
  const body  = document.getElementById("body")?.value.trim();

  if (!title && !body) return;

  try {
    await fetch(`${API_BASE}/posts/draft`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, body }),
    });
  } catch (err) {
    console.warn("Draft auto-save failed", err);
  }
}

// for tab close / refresh
function saveDraftSync() {
  if (!autosaveEnabled) return;

  const title = document.getElementById("title")?.value.trim();
  const body  = document.getElementById("body")?.value.trim();

  if (!title && !body) return;

  navigator.sendBeacon(
    `${API_BASE}/posts/draft`,
    JSON.stringify({ title, body })
  );
}

/* =========================
   RESTORE DRAFT
========================= */

async function restoreDraftIfExists() {
  try {
    const res = await fetch(`${API_BASE}/posts/draft`, {
      credentials: "include",
    });

    if (!res.ok || res.status === 401) return;

    const { data } = await res.json();
    if (!data) return;

    const ok = confirm("Unsaved draft found. Restore it?");
    if (!ok) {
      await fetch(`${API_BASE}/posts/draft`, {
        method: "DELETE",
        credentials: "include",
      });
      return;
    }

    document.getElementById("title").value = data.title || "";
    document.getElementById("body").value  = data.body || "";

  } catch (err) {
    console.warn("Draft restore failed", err);
  }
}

/* =========================
   CREATE POST
========================= */

document.addEventListener("DOMContentLoaded", async () => {
  const form = document.getElementById("create-post-form");
  if (!form) return;

  const titleInput = document.getElementById("title");
  const bodyInput  = document.getElementById("body");

  populateCategories();

  const user = await getCurrentUser();

  if (user) {
    restoreDraftIfExists();

    titleInput?.addEventListener("input", scheduleDraftSave);
    bodyInput?.addEventListener("input", scheduleDraftSave);

    window.addEventListener("beforeunload", saveDraftSync);
    window.addEventListener("pagehide", saveDraftSync);
    document.addEventListener("visibilitychange", () => {
      if (document.hidden) saveDraft();
    });
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    // 🔴 STOP AUTOSAVE COMPLETELY
    autosaveEnabled = false;
    clearTimeout(draftTimer);

    const title = titleInput.value.trim();
    const body  = bodyInput.value.trim();
    const categoryId = document.getElementById("categorySelect").value;
    const action = e.submitter?.value;

    if (!title) {
      alert("Title is required");
      return;
    }

    if (!categoryId && action !== "draft") {
      alert("Category is required");
      return;
    }

    const payload = {
      title,
      body,
      status: action === "draft" ? "draft" : "published",
      category_ids: categoryId ? [Number(categoryId)] : [],
    };

    try {
      const res = await fetch(`${API_BASE}/posts`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (res.status === 401) {
        alert("You must be logged in to create a post.");
        window.location.href = "/login";
        return;
      }

      const data = await res.json();

      if (!res.ok) {
        alert(data?.error?.message || "Failed to create post");
        return;
      }

      if (payload.status === "published") {
        // 🧹 CLEAN DRAFT & REDIRECT
        await fetch(`${API_BASE}/posts/draft`, {
          method: "DELETE",
          credentials: "include",
        });

        window.location.href = "/";
        return;
      }

      alert("Draft saved successfully");

    } catch (err) {
      console.error("Create post error:", err);
      alert("Unexpected error while creating post");
    }
  });
});

async function getCurrentUser() {
  try {
    const res = await fetch(`${API_BASE}/users/me`, {
      credentials: "include",
    });
    if (!res.ok) return null;
    const { data } = await res.json();
    return data;
  } catch {
    return null;
  }
}
