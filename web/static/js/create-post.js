// web/static/js/create-post.js
import { API_BASE } from "./utils.js";

const MANUAL_DRAFT_KEY = "manual_draft_saved";

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
  const categories = Array.isArray(payload?.data) ? payload.data : payload;

  select.querySelectorAll("option:not(:first-child)").forEach(o => o.remove());

  categories.forEach(c => {
    const opt = document.createElement("option");
    opt.value = c.id;
    opt.textContent = c.name;
    select.appendChild(opt);
  });
}

/* =========================
   AUTOSAVE
========================= */

let draftTimer = null;
let autosaveEnabled = true;

function scheduleDraftSave() {
  if (!autosaveEnabled) return;

  const title = document.getElementById("title")?.value.trim();
  if (!title) return; 

  localStorage.removeItem(MANUAL_DRAFT_KEY);

  clearTimeout(draftTimer);
  draftTimer = setTimeout(saveDraft, 1000);
}

function saveDraftSync() {
  if (!autosaveEnabled) return;

  const title = document.getElementById("title")?.value.trim();
  const body  = document.getElementById("body")?.value.trim();

  if (!title) return; 

  navigator.sendBeacon(
    `${API_BASE}/posts/draft`,
    JSON.stringify({ title, body })
  );
}

/* =========================
   RESTORE DRAFT
========================= */

async function restoreDraftIfExists() {
  // manual draft → no restore prompt
  if (localStorage.getItem(MANUAL_DRAFT_KEY)) return;

  try {
    const res = await fetch(`${API_BASE}/posts/draft`, {
      credentials: "include",
    });

    if (!res.ok || res.status === 401) return;

    const { data } = await res.json();
    if (!data) return;

    const ok = confirm("Unsaved draft found. Restore it?");
    if (!ok) return;

    document.getElementById("title").value = data.title || "";
    document.getElementById("body").value  = data.body || "";
  } catch {
    /* silent */
  }
}

/* =========================
   MAIN
========================= */

document.addEventListener("DOMContentLoaded", async () => {
  const form = document.getElementById("create-post-form");
  if (!form) return;

  const titleInput = document.getElementById("title");
  const bodyInput  = document.getElementById("body");

  await populateCategories();

  const user = await getCurrentUser();
  if (!user) return;

  await restoreDraftIfExists();

  titleInput.addEventListener("input", scheduleDraftSave);
  bodyInput.addEventListener("input", scheduleDraftSave);

  window.addEventListener("beforeunload", saveDraftSync);
  window.addEventListener("pagehide", saveDraftSync);

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const title = titleInput.value.trim();
    const body  = bodyInput.value.trim();
    const categoryId = document.getElementById("categorySelect").value;
    const action = e.submitter?.value;

    if (!title) {
      alert("Title is required");
      return;
    }

    /* =========================
       MANUAL SAVE DRAFT
    ========================= */
    if (action === "draft") {
      const res = await fetch(`${API_BASE}/posts/draft`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, body }),
      });

      if (!res.ok) {
        alert("Failed to save draft");
        return;
      }

      localStorage.setItem(MANUAL_DRAFT_KEY, "1");
      alert("Draft saved successfully");
      return;
    }

    /* =========================
       PUBLISH
    ========================= */
    if (!categoryId) {
      alert("Category is required");
      return;
    }

    autosaveEnabled = false;
    clearTimeout(draftTimer);

    const res = await fetch(`${API_BASE}/posts`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({
        title,
        body,
        category_ids: [Number(categoryId)],
      }),
    });

    if (!res.ok) {
      autosaveEnabled = true;
      alert("Failed to publish post");
      return;
    }

    await fetch(`${API_BASE}/posts/draft`, {
      method: "DELETE",
      credentials: "include",
    });

    localStorage.removeItem(MANUAL_DRAFT_KEY);
    window.location.href = "/";
  });
});

async function getCurrentUser() {
  try {
    const res = await fetch(`${API_BASE}/users/me`, { credentials: "include" });
    if (!res.ok) return null;
    const { data } = await res.json();
    return data;
  } catch {
    return null;
  }
}
