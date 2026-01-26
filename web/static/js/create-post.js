// web/static/js/create-post.js
import { API_BASE } from "./utils.js";
import { Auth } from "./auth.js";

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

async function renderCategoryCheckboxes() {
  const host = document.getElementById("categoryCheckboxes");
  if (!host) return;

  const payload = await loadCategories();
  const categories = Array.isArray(payload?.data) ? payload.data : payload;

  host.innerHTML = "";

  categories.forEach((c) => {
    const label = document.createElement("label");
    label.className = "category-checkbox";

    const input = document.createElement("input");
    input.type = "checkbox";
    input.value = c.id;

    const span = document.createElement("span");
    span.textContent = c.name;

    label.appendChild(input);
    label.appendChild(span);
    host.appendChild(label);
  });
}

function getSelectedCategoryIds() {
  return Array.from(
    document.querySelectorAll(
      "#categoryCheckboxes input[type='checkbox']:checked"
    )
  ).map((el) => Number(el.value));
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

function saveDraft() {
  if (!autosaveEnabled) return;

  const title = document.getElementById("title")?.value.trim();
  const body = document.getElementById("body")?.value.trim();

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
    document.getElementById("body").value = data.body || "";
  } catch {}
}

/* =========================
   MAIN
========================= */

document.addEventListener("DOMContentLoaded", async () => {
  const form = document.getElementById("create-post-form");
  if (!form) return;

  const titleInput = document.getElementById("title");
  const bodyInput = document.getElementById("body");

  await renderCategoryCheckboxes();
  await restoreDraftIfExists();

  titleInput.addEventListener("input", scheduleDraftSave);
  bodyInput.addEventListener("input", scheduleDraftSave);

  window.addEventListener("beforeunload", saveDraft);
  window.addEventListener("pagehide", saveDraft);

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    // 🔐 AUTH GUARD (modal for guests)
    const allowed = await Auth.requireOrPrompt();
    if (!allowed) return;

    const title = titleInput.value.trim();
    const body = bodyInput.value.trim();
    const categoryIds = getSelectedCategoryIds();
    const action = e.submitter?.value;

    if (!title) {
      alert("Title is required");
      return;
    }

    if (action === "draft") {
      await fetch(`${API_BASE}/posts/draft`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, body }),
      });

      localStorage.setItem(MANUAL_DRAFT_KEY, "1");
      alert("Draft saved successfully");
      return;
    }

    if (categoryIds.length === 0) {
      alert("Select at least one category");
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
        category_ids: categoryIds,
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
