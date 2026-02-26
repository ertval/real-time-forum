// web/static/js/edit-post.js
import {
  API_BASE,
  buildImageRequestOptions,
  buildPostMultipartFormData,
  IMAGE_ACCEPT_ATTR,
} from "./utils.js";
import { Auth } from "./auth.js";
import { uiNotify } from "./ui-messages.js";
import { playUpload } from "./sound-effects.js";
import { setupImagePicker } from "./image-picker.js";

const REDIRECT_DELAY_MS = 700;

function getPostIdFromURL() {
  const parts = window.location.pathname.split("/").filter(Boolean);
  const id = Number(parts[parts.length - 1]);
  return Number.isFinite(id) && id > 0 ? id : null;
}

function getReturnPath() {
  const raw = new URLSearchParams(window.location.search).get("next");
  if (!raw) return "/activity";
  if (!raw.startsWith("/") || raw.startsWith("//")) return "/activity";
  return raw;
}

function setSubmitDisabled(submitButton, disabled) {
  if (submitButton instanceof HTMLButtonElement) {
    submitButton.disabled = disabled;
  }
}

function redirectTo(path) {
  setTimeout(() => {
    window.location.href = path;
  }, REDIRECT_DELAY_MS);
}

async function loadCategories() {
  try {
    const res = await fetch(`${API_BASE}/categories`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });
    if (!res.ok) return [];
    const payload = await res.json().catch(() => null);
    if (Array.isArray(payload?.data)) return payload.data;
    return Array.isArray(payload) ? payload : [];
  } catch {
    return [];
  }
}

async function renderCategoryCheckboxes(host) {
  if (!(host instanceof HTMLElement)) return;
  const categories = await loadCategories();
  host.innerHTML = "";

  categories.forEach((category) => {
    const label = document.createElement("label");
    label.className = "category-checkbox";

    const input = document.createElement("input");
    input.type = "checkbox";
    input.value = String(category.id);

    const text = document.createElement("span");
    text.textContent = category.name;

    label.appendChild(input);
    label.appendChild(text);
    host.appendChild(label);
  });
}

function getSelectedCategoryIds(host) {
  if (!(host instanceof HTMLElement)) return [];
  return Array.from(
    host.querySelectorAll("input[type='checkbox']:checked")
  ).map((el) => Number(el.value));
}

function setSelectedCategoryIds(host, ids = []) {
  const selected = new Set((ids || []).map(Number));
  host
    .querySelectorAll("input[type='checkbox']")
    .forEach((cb) => {
      cb.checked = selected.has(Number(cb.value));
    });
}

function normalizeCategoryIDs(categories = []) {
  return categories
    .map((category) => Number(category?.id))
    .filter((id) => Number.isFinite(id) && id > 0)
    .sort((a, b) => a - b);
}

function sameNumberSet(a = [], b = []) {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i += 1) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

document.addEventListener("DOMContentLoaded", async () => {
  const form = document.getElementById("edit-post-form");
  const titleInput = document.getElementById("title");
  const bodyInput = document.getElementById("body");
  const imageInput = document.getElementById("image");
  const imageButton = document.getElementById("image-button");
  const imageName = document.getElementById("image-name");
  const imagePreview = document.getElementById("image-preview");
  const imageClear = document.getElementById("image-clear");
  const imagePreviewImg = imagePreview?.querySelector("img");
  const categoryCheckboxes = document.getElementById("categoryCheckboxes");
  const backLink = document.querySelector(".create-post-back a");

  if (
    !(form instanceof HTMLFormElement) ||
    !(titleInput instanceof HTMLInputElement) ||
    !(bodyInput instanceof HTMLTextAreaElement) ||
    !(imageInput instanceof HTMLInputElement) ||
    !(imageButton instanceof HTMLButtonElement) ||
    !(imageName instanceof HTMLElement) ||
    !(imagePreview instanceof HTMLElement) ||
    !(imageClear instanceof HTMLButtonElement) ||
    !(imagePreviewImg instanceof HTMLImageElement) ||
    !(categoryCheckboxes instanceof HTMLElement)
  ) {
    return;
  }

  imageInput.setAttribute("accept", IMAGE_ACCEPT_ATTR);

  const submitButton = form.querySelector("button[type='submit']");
  const returnPath = getReturnPath();
  if (backLink instanceof HTMLAnchorElement) {
    backLink.href = returnPath;
  }

  const postID = getPostIdFromURL();
  if (!postID) {
    uiNotify("Invalid post ID.", { type: "danger" });
    setSubmitDisabled(submitButton, true);
    return;
  }

  await Auth.init();
  if (!Auth.isAuthenticated) {
    window.location.href = "/login";
    return;
  }

  setSubmitDisabled(submitButton, true);

  let originalTitle = "";
  let originalBody = "";
  let originalCategoryIDs = [];
  let originalImageURL = null;
  let currentImageURL = null;
  let removeImage = false;

  const imagePicker = setupImagePicker({
    input: imageInput,
    triggerButton: imageButton,
    clearButton: imageClear,
    nameLabel: imageName,
    previewContainer: imagePreview,
    previewImage: imagePreviewImg,
    persistedUrl: null,
    persistedLabel: "Current post image attached",
    onTooLarge: () => {
      uiNotify("Image must be 20MB or smaller.", { type: "danger" });
    },
    onClearPersisted: () => {
      currentImageURL = null;
      removeImage = true;
    },
  });

  imageInput.addEventListener("change", () => {
    if (imagePicker.getFile()) {
      removeImage = false;
    }
  });

  try {
    await renderCategoryCheckboxes(categoryCheckboxes);

    const res = await fetch(`${API_BASE}/posts/${postID}`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });

    if (res.status === 404) {
      uiNotify("Post not found.", { type: "warn" });
      redirectTo(returnPath);
      return;
    }

    if (!res.ok) {
      uiNotify("Failed to load post.", { type: "danger" });
      return;
    }

    const payload = await res.json().catch(() => null);
    const post = payload?.data;

    if (!post || Number(post.id) !== postID) {
      uiNotify("Failed to load post.", { type: "danger" });
      return;
    }

    if (Number(post.author_id) !== Number(Auth.user?.id)) {
      uiNotify("You can only edit your own posts.", { type: "warn" });
      redirectTo(returnPath);
      return;
    }

    originalTitle = String(post.title ?? "");
    originalBody = String(post.body ?? "");
    originalCategoryIDs = normalizeCategoryIDs(post.categories || []);
    originalImageURL =
      typeof post.image_url === "string" && post.image_url.trim()
        ? post.image_url.trim()
        : null;

    currentImageURL = originalImageURL;
    removeImage = false;

    titleInput.value = originalTitle;
    bodyInput.value = originalBody;
    setSelectedCategoryIds(categoryCheckboxes, originalCategoryIDs);
    imagePicker.setPersistedUrl(currentImageURL);
  } catch {
    uiNotify("Failed to load post.", { type: "danger" });
    return;
  } finally {
    setSubmitDisabled(submitButton, false);
  }

  let isSubmitting = false;

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (isSubmitting) return;

    const allowed = await Auth.requireOrPrompt();
    if (!allowed) return;

    const title = titleInput.value.trim();
    const body = bodyInput.value.trim();
    const categoryIDs = getSelectedCategoryIds(categoryCheckboxes);
    const sortedCategoryIDs = [...categoryIDs].sort((a, b) => a - b);
    const selectedImageFile = imagePicker.getFile();
    const hasSelectedImage = !!selectedImageFile;
    const hasCurrentImage = !!currentImageURL;

    if (!title) {
      uiNotify("Title is required.", { type: "warn" });
      return;
    }

    if (!body && !hasSelectedImage && !hasCurrentImage) {
      uiNotify("Post body is required.", { type: "warn" });
      return;
    }

    if (categoryIDs.length === 0) {
      uiNotify("Select at least one category.", { type: "warn" });
      return;
    }

    const titleChanged = title !== originalTitle;
    const bodyChanged = body !== originalBody;
    const categoriesChanged = !sameNumberSet(sortedCategoryIDs, originalCategoryIDs);
    const imageChanged = hasSelectedImage || (removeImage && !!originalImageURL);

    if (!titleChanged && !bodyChanged && !categoriesChanged && !imageChanged) {
      uiNotify("No changes to update.", { type: "info" });
      return;
    }

    isSubmitting = true;
    setSubmitDisabled(submitButton, true);

    try {
      const res = await fetch(`${API_BASE}/posts/${postID}`, buildImageRequestOptions({
        method: "PATCH",
        imageFile: selectedImageFile,
        buildMultipartBody: (imageFile) => {
          const formData = buildPostMultipartFormData({
            title,
            body,
            categoryIds: categoryIDs,
            imageFile,
          });
          if (removeImage && originalImageURL) {
            formData.append("remove_image", "true");
          }
          return formData;
        },
        jsonBody: {
          title,
          body,
          category_ids: categoryIDs,
          remove_image: removeImage && !!originalImageURL,
        },
        jsonHeaders: {
          Accept: "application/json",
        },
      }));

      if (res.status === 401) {
        uiNotify("You must be logged in.", { type: "warn" });
        window.location.href = "/login";
        return;
      }

      if (res.status === 403) {
        uiNotify("You can only edit your own posts.", { type: "warn" });
        redirectTo(returnPath);
        return;
      }

      if (res.status === 404) {
        uiNotify("Post not found.", { type: "warn" });
        redirectTo(returnPath);
        return;
      }

      if (!res.ok) {
        const payload = await res.json().catch(() => null);
        uiNotify(payload?.error?.message || "Failed to update post.", {
          type: "danger",
        });
        return;
      }

      playUpload();
      uiNotify("Post updated.", { type: "success" });
      redirectTo(returnPath);
    } catch {
      uiNotify("Failed to update post.", { type: "danger" });
    } finally {
      isSubmitting = false;
      setSubmitDisabled(submitButton, false);
    }
  });
});
