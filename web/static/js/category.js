import { API_BASE } from "./utils.js";

/* =========================
   URL HELPERS
========================= */

function getCategoryFromURL() {
  const params = new URLSearchParams(window.location.search);
  return params.get("category") || "";
}

function setCategoryToURL(categoryName) {
  const url = new URL(window.location.href);

  if (categoryName) {
    url.searchParams.set("category", categoryName);
  } else {
    url.searchParams.delete("category");
  }

  history.pushState({}, "", url);
}

/* =========================
   API
========================= */

async function loadCategories() {
  const res = await fetch(`${API_BASE}/categories`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  if (!res.ok) return [];

  const payload = await res.json();
  return Array.isArray(payload) ? payload : payload.data ?? [];
}

/* =========================
   INIT CATEGORY FILTER
========================= */

export async function initCategoryFilter(onChange) {
  const select = document.getElementById("categoryFilter");
  if (!select) return "";

  const categories = await loadCategories();

  // clear (keep "All")
  select.querySelectorAll("option:not(:first-child)").forEach(o => o.remove());

  const nameToId = new Map();
  const idToName = new Map();

  for (const c of categories) {
    nameToId.set(c.name, String(c.id));
    idToName.set(String(c.id), c.name);

    const opt = document.createElement("option");
    opt.value = String(c.id);
    opt.textContent = c.name;
    select.appendChild(opt);
  }

  let initialId = "";

  // apply URL → select
  const categoryName = getCategoryFromURL();
  if (categoryName && nameToId.has(categoryName)) {
    initialId = nameToId.get(categoryName);
    select.value = initialId;
  }

  // change handler
  select.addEventListener("change", () => {
    const id = select.value;
    const name = idToName.get(id) || "";

    setCategoryToURL(name);
    onChange(id);
  });

  // back / forward
  window.addEventListener("popstate", () => {
    const name = getCategoryFromURL();
    const id = nameToId.get(name) || "";

    select.value = id;
    onChange(id);
  });

  return initialId;
}
