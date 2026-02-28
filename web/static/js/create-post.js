// web/static/js/create-post.js
import {
  API_BASE,
  buildImageRequestOptions,
  buildPostMultipartFormData,
  IMAGE_ACCEPT_ATTR,
} from "./utils.js";
import { Auth } from "./auth.js";
import { uiNotify, uiConfirm } from "./ui-messages.js";
import { playUpload } from "./sound-effects.js";
import { setupImagePicker } from "./image-picker.js";
import {
  renderCategoryCheckboxes,
  getSelectedCategoryIds,
  setSelectedCategoryIds,
} from "./post-form.js";

let currentDraftId = null;
let draftImageURL = null;

/*----------
  AUTOSAVE
----------*/

let draftTimer = null;
let autosaveEnabled = true;
let titleInputRef = null;
let bodyInputRef = null;
let imageInputRef = null;
let categoryCheckboxesRef = null;
let draftSaveInFlight = false;
let draftSaveQueued = false;

function getCategoryHost() {
  return categoryCheckboxesRef || document.getElementById("categoryCheckboxes");
}

function scheduleDraftSave() {
  if (!autosaveEnabled) return;

  const title = titleInputRef?.value.trim()
    ?? document.getElementById("title")?.value.trim();
  if (!title) return;

  clearTimeout(draftTimer);
  draftTimer = setTimeout(saveDraft, 1000);
}

async function saveDraft() {
  if (!autosaveEnabled) return;
  if (draftSaveInFlight) {
    draftSaveQueued = true;
    return;
  }

  const title = titleInputRef?.value.trim()
    ?? document.getElementById("title")?.value.trim();
  const body = bodyInputRef?.value.trim()
    ?? document.getElementById("body")?.value.trim();
  const categoryIds = getSelectedCategoryIds(getCategoryHost());
  const pendingImageFile = imageInputRef?.files?.[0]
    ?? document.getElementById("image")?.files?.[0];

  if (!title) return;
  if (!body && !draftImageURL) return;
  if (pendingImageFile) return;

  draftSaveInFlight = true;
  try {
    if (currentDraftId) {
      await fetch(`${API_BASE}/posts/draft/${currentDraftId}`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title,
          body,
          image_url: draftImageURL,
          category_ids: categoryIds,
        }),
      });
      return;
    }

    const res = await fetch(`${API_BASE}/posts/draft`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title,
        body,
        image_url: draftImageURL,
        category_ids: categoryIds,
      }),
    });

    if (!res.ok) return;

    const payload = await res.json().catch(() => null);
    currentDraftId = payload?.data?.id ?? null;
  } catch {
    // Best-effort autosave: network failures should not break user flow.
  } finally {
    draftSaveInFlight = false;
    if (draftSaveQueued) {
      draftSaveQueued = false;
      queueMicrotask(saveDraft);
    }
  }
}

async function refreshDraftState() {
  const data = await fetchDraftRecord();
  if (!data) return null;
  applyDraftState(data);
  return data;
}

/*---------------
  RESTORE DRAFT
---------------*/

async function restoreDraftIfExists() {
  try {
    const data = await fetchDraftRecord();
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

    applyDraftState(data);
    document.getElementById("title").value = data.title || "";
    document.getElementById("body").value = data.body || "";
    setSelectedCategoryIds(
      document.getElementById("categoryCheckboxes"),
      data.category_ids || []
    );

    uiNotify("Draft restored successfully.", { type: "success" });
  } catch {}
}

async function fetchDraftRecord() {
  try {
    const res = await fetch(`${API_BASE}/posts/draft`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });
    if (!res.ok) return null;
    const payload = await res.json().catch(() => null);
    return payload?.data || null;
  } catch {
    return null;
  }
}

function applyDraftState(data) {
  currentDraftId = data?.id ?? null;
  draftImageURL = data?.image_url || null;
}

/*-------
  MAIN
-------*/

document.addEventListener("DOMContentLoaded", async () => {
  const form = document.getElementById("create-post-form");
  if (!form) return;
  let isSubmitting = false;

  const titleInput = document.getElementById("title");
  const bodyInput = document.getElementById("body");
  const imageInput = document.getElementById("image");
  if (imageInput) {
    imageInput.setAttribute("accept", IMAGE_ACCEPT_ATTR);
  }
  const imageButton = document.getElementById("image-button");
  const imageName = document.getElementById("image-name");
  const imagePreview = document.getElementById("image-preview");
  const imageClear = document.getElementById("image-clear");
  const imagePreviewImg = imagePreview?.querySelector("img");
  const categoryCheckboxes = document.getElementById("categoryCheckboxes");
  titleInputRef = titleInput;
  bodyInputRef = bodyInput;
  imageInputRef = imageInput;
  categoryCheckboxesRef = categoryCheckboxes;
  const submitButtons = Array.from(form.querySelectorAll("button[type='submit']"));

  const setSubmitButtonsDisabled = (disabled) => {
    submitButtons.forEach(btn => {
      btn.disabled = disabled;
    });
  };

  const imagePicker = setupImagePicker({
    input: imageInput,
    triggerButton: imageButton,
    clearButton: imageClear,
    nameLabel: imageName,
    previewContainer: imagePreview,
    previewImage: imagePreviewImg,
    persistedUrl: draftImageURL,
    persistedLabel: "Saved draft image attached",
    onTooLarge: () => {
      uiNotify("Image must be 20MB or smaller.", { type: "danger" });
    },
    onClearPersisted: () => {
      draftImageURL = null;
      scheduleDraftSave();
    },
  });

  await renderCategoryCheckboxes(categoryCheckboxes);
  categoryCheckboxes?.addEventListener("change", scheduleDraftSave);

  await restoreDraftIfExists();
  imagePicker.setPersistedUrl(draftImageURL);

  titleInput.addEventListener("input", scheduleDraftSave);
  bodyInput.addEventListener("input", scheduleDraftSave);

  window.addEventListener("beforeunload", saveDraft);
  window.addEventListener("pagehide", saveDraft);

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (isSubmitting) return;
    isSubmitting = true;
    setSubmitButtonsDisabled(true);
    let shouldReleaseLock = true;

    try {
      const allowed = await Auth.requireOrPrompt();
      if (!allowed) return;

      const title = titleInput.value.trim();
      const body = bodyInput.value.trim();
      const categoryIds = getSelectedCategoryIds(categoryCheckboxes);
      const action = e.submitter?.value;

      if (!title) {
        uiNotify("Title is required.", { type: "warn" });
        return;
      }

      const selectedImageFile = imagePicker.getFile();
      const hasImage = !!selectedImageFile;
      const hasDraftImage = !!draftImageURL;

      if (!body && !hasImage && !hasDraftImage) {
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
        const res = await fetch(url, buildImageRequestOptions({
          method,
          imageFile: selectedImageFile,
          buildMultipartBody: imageFile =>
            buildPostMultipartFormData({
              title,
              body,
              categoryIds,
              imageFile,
              manual: true,
              imageURL: draftImageURL,
            }),
          jsonBody: {
            title,
            body,
            image_url: draftImageURL,
            category_ids: categoryIds,
            manual: true,
          },
        }));

        if (!res.ok) {
          const payload = await res.json().catch(() => null);
          uiNotify(payload?.error?.message || "Draft save failed.", {
            type: "danger",
          });
          return;
        }

        if (hasImage) {
          imagePicker.clearSelectedFile();
        }
        await refreshDraftState();
        imagePicker.setPersistedUrl(draftImageURL);

        // Manual draft save → play sound + redirect to My Posts
        playUpload();
        uiNotify("Draft saved successfully.", { type: "success" });

        shouldReleaseLock = false;
        setTimeout(() => {
          window.location.href = "/my-posts";
        }, 750);

        return;
      }

      autosaveEnabled = false;
      clearTimeout(draftTimer);

      const res = await fetch(`${API_BASE}/posts`, buildImageRequestOptions({
        method: "POST",
        imageFile: selectedImageFile,
        buildMultipartBody: imageFile =>
          buildPostMultipartFormData({
            title,
            body,
            categoryIds,
            imageFile,
          }),
        jsonBody: {
          title,
          body,
          image_url: draftImageURL,
          category_ids: categoryIds,
        },
        jsonHeaders: {
          Accept: "application/json",
        },
      }));

      if (!res.ok) {
        autosaveEnabled = true;
        const payload = await res.json().catch(() => null);
        uiNotify(payload?.error?.message || "Failed to publish post.", { type: "danger" });
        return;
      }

      if (currentDraftId) {
        await fetch(`${API_BASE}/posts/draft/${currentDraftId}`, {
          method: "DELETE",
          credentials: "include",
        });
      }

      // SOUND EFFECT for publish
      playUpload();

      shouldReleaseLock = false;
      setTimeout(() => {
        window.location.href = "/";
      }, 750);
    } finally {
      if (shouldReleaseLock) {
        isSubmitting = false;
        setSubmitButtonsDisabled(false);
      }
    }
  });
});
