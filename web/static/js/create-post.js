// web/static/js/create-post.js
import { API_BASE } from "./utils.js";
import { Auth } from "./auth.js";
import { uiNotify, uiConfirm } from "./ui-messages.js";

let currentDraftId = null;

/*-----------------
  LOAD CATEGORIES
-----------------*/

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

/*----------
  AUTOSAVE
----------*/

let draftTimer = null;
let autosaveEnabled = true;

function scheduleDraftSave() {
  if (!autosaveEnabled) return;

  const title = document.getElementById("title")?.value.trim();
  if (!title) return;

  clearTimeout(draftTimer);
  draftTimer = setTimeout(saveDraft, 1000);
}

async function saveDraft() {
  if (!autosaveEnabled) return;

  const title = document.getElementById("title")?.value.trim();
  const body = document.getElementById("body")?.value.trim();
  const categoryIds = getSelectedCategoryIds();

  if (!title || !body) return;

  if (currentDraftId) {
    await fetch(`${API_BASE}/posts/draft/${currentDraftId}`, {
      method: "PUT",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, body, category_ids: categoryIds }),
    });
    return;
  }

  const res = await fetch(`${API_BASE}/posts/draft`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title, body, category_ids: categoryIds }),
  });

  if (!res.ok) return;

  const payload = await res.json().catch(() => null);
  currentDraftId = payload?.data?.id ?? null;
}

/*---------------
  RESTORE DRAFT
---------------*/

async function restoreDraftIfExists() {
  try {
    const res = await fetch(`${API_BASE}/posts/draft`, {
      credentials: "include",
    });

    if (!res.ok || res.status === 401) return;

    const { data } = await res.json();
    if (!data) return;

    const restore = await uiConfirm(
      "Draft found. Do you want to restore it?",
      {
        type: "warn",
        title: "Draft detected",
        okText: "Restore",
        cancelText: "Cancel",
      }
    );

    if (!restore) return;

    currentDraftId = data.id;
    document.getElementById("title").value = data.title || "";
    document.getElementById("body").value = data.body || "";
    setSelectedCategoryIds(data.category_ids || []);

    uiNotify("Draft restored successfully.", { type: "success" });
  } catch {}
}

/*-------
  MAIN
-------*/

document.addEventListener("DOMContentLoaded", async () => {
  const form = document.getElementById("create-post-form");
  if (!form) return;

  const titleInput = document.getElementById("title");
  const bodyInput = document.getElementById("body");
  const imageInput = document.getElementById("image");
  const imageButton = document.getElementById("image-button");
  const imageName = document.getElementById("image-name");

  await renderCategoryCheckboxes();
  document
    .getElementById("categoryCheckboxes")
    ?.addEventListener("change", scheduleDraftSave);

  await restoreDraftIfExists();

  titleInput.addEventListener("input", scheduleDraftSave);
  bodyInput.addEventListener("input", scheduleDraftSave);

  if (imageButton && imageInput) {
    imageButton.addEventListener("click", () => {
      imageInput.click();
    });
  }

  if (imageInput && imageName) {
    imageInput.addEventListener("change", () => {
      const file = imageInput.files && imageInput.files[0];
      imageName.textContent = file ? `Selected: ${file.name}` : "";
    });
  }

  window.addEventListener("beforeunload", saveDraft);
  window.addEventListener("pagehide", saveDraft);

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const allowed = await Auth.requireOrPrompt();
    if (!allowed) return;

    const title = titleInput.value.trim();
    const body = bodyInput.value.trim();
    const categoryIds = getSelectedCategoryIds();
    const action = e.submitter?.value;

    if (!title) {
      uiNotify("Title is required.", { type: "warn" });
      return;
    }

    if (!body) {
      uiNotify("Post body is required.", { type: "warn" });
      return;
    }

    // Check if at least one category is selected for both actions
    if (categoryIds.length === 0) {
      uiNotify("Select at least one category.", { type: "warn" });
      return;
    }

    if (action === "draft") {
      const url = currentDraftId
        ? `${API_BASE}/posts/draft/${currentDraftId}`
        : `${API_BASE}/posts/draft`;

      const method = currentDraftId ? "PUT" : "POST";

      const res = await fetch(url, {
        method,
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title,
          body,
          category_ids: categoryIds,
          manual: true,
        }),
      });

      if (!res.ok) {
        const payload = await res.json().catch(() => null);
        uiNotify(payload?.error?.message || "Draft save failed.", {
          type: "danger",
        });
        return;
      }

      if (!currentDraftId) {
        const payload = await res.json().catch(() => null);
        currentDraftId = payload?.data?.id ?? null;
      }

      uiNotify("Draft saved successfully.", { type: "success" });
      return;
    }

    autosaveEnabled = false;
    clearTimeout(draftTimer);

    const hasImage =
      imageInput &&
      imageInput.files &&
      imageInput.files.length > 0 &&
      imageInput.files[0];

    const res = await fetch(`${API_BASE}/posts`, hasImage
      ? {
          method: "POST",
          credentials: "include",
          body: buildMultipartPayload(title, body, categoryIds, imageInput.files[0]),
        }
      : {
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
      uiNotify("Failed to publish post.", { type: "danger" });
      return;
    }

    if (currentDraftId) {
      await fetch(`${API_BASE}/posts/draft/${currentDraftId}`, {
        method: "DELETE",
        credentials: "include",
      });
    }

    window.location.href = "/";
  });
});

function setSelectedCategoryIds(ids = []) {
  const set = new Set((ids || []).map(Number));
  document
    .querySelectorAll("#categoryCheckboxes input[type='checkbox']")
    .forEach((cb) => {
      cb.checked = set.has(Number(cb.value));
    });
}

function buildMultipartPayload(title, body, categoryIds, imageFile) {
  const formData = new FormData();
  formData.append("title", title);
  formData.append("body", body);
  categoryIds.forEach((id) => formData.append("category_ids", String(id)));
  if (imageFile) {
    formData.append("image", imageFile);
  }
  return formData;
}
