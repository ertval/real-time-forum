// web/static/js/image-picker.js
import { MAX_IMAGE_BYTES } from "./utils.js";

export function setupImagePicker({
  input,
  triggerButton,
  clearButton,
  nameLabel,
  previewContainer,
  previewImage,
  persistedUrl = null,
  persistedLabel = "",
  maxBytes = MAX_IMAGE_BYTES,
  onTooLarge = () => {},
  onClearPersisted = () => {},
} = {}) {
  let objectPreviewUrl = null;
  let currentPersistedUrl = normalizeURL(persistedUrl);

  const revokeObjectPreview = () => {
    if (!objectPreviewUrl) return;
    URL.revokeObjectURL(objectPreviewUrl);
    objectPreviewUrl = null;
  };

  const getFile = () => input?.files?.[0] || null;

  const updateUI = () => {
    const file = getFile();

    if (clearButton) {
      clearButton.hidden = !file && !currentPersistedUrl;
    }

    if (nameLabel) {
      if (file) {
        nameLabel.textContent = `Selected: ${file.name}`;
      } else if (currentPersistedUrl && persistedLabel) {
        nameLabel.textContent = persistedLabel;
      } else {
        nameLabel.textContent = "";
      }
    }

    if (!previewContainer || !previewImage) {
      revokeObjectPreview();
      return;
    }

    revokeObjectPreview();

    if (file) {
      objectPreviewUrl = URL.createObjectURL(file);
      previewImage.src = objectPreviewUrl;
      previewContainer.hidden = false;
      return;
    }

    if (currentPersistedUrl) {
      previewImage.src = currentPersistedUrl;
      previewContainer.hidden = false;
      return;
    }

    previewImage.removeAttribute("src");
    previewContainer.hidden = true;
  };

  const onInputChange = () => {
    const file = getFile();
    if (file && file.size > maxBytes) {
      if (input) input.value = "";
      onTooLarge(file, maxBytes);
    }
    updateUI();
  };

  const onTriggerClick = () => input?.click();

  const onClearClick = () => {
    const file = getFile();
    if (file && input) {
      input.value = "";
      updateUI();
      return;
    }

    if (currentPersistedUrl) {
      currentPersistedUrl = null;
      onClearPersisted();
      updateUI();
    }
  };

  triggerButton?.addEventListener("click", onTriggerClick);
  input?.addEventListener("change", onInputChange);
  clearButton?.addEventListener("click", onClearClick);

  updateUI();

  return {
    getFile,
    clearSelectedFile() {
      if (input) input.value = "";
      updateUI();
    },
    setPersistedUrl(url) {
      currentPersistedUrl = normalizeURL(url);
      updateUI();
    },
    clearPersistedUrl() {
      currentPersistedUrl = null;
      updateUI();
    },
    hasAnyImage() {
      return !!getFile() || !!currentPersistedUrl;
    },
    destroy() {
      triggerButton?.removeEventListener("click", onTriggerClick);
      input?.removeEventListener("change", onInputChange);
      clearButton?.removeEventListener("click", onClearClick);
      revokeObjectPreview();
    },
  };
}

function normalizeURL(value) {
  if (typeof value !== "string") return null;
  const trimmed = value.trim();
  return trimmed || null;
}
