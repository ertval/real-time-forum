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
   CREATE POST
========================= */

document.addEventListener("DOMContentLoaded", () => {
  const form = document.getElementById("create-post-form");
  if (!form) return;

  populateCategories();

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const title = document.getElementById("title").value.trim();
    const body = document.getElementById("body").value.trim();
    const categoryId = document.getElementById("categorySelect").value;
    const action = e.submitter?.value;

    if (!title) {
      alert("Title is required");
      return;
    }

    if (!categoryId) {
      alert("Category is required");
      return;
    }

    const payload = {
      title,
      body,
      status: action === "draft" ? "draft" : "published",
      category_ids: [Number(categoryId)],
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

      if (payload.status === "draft") {
        alert("Draft saved successfully");
      } else {
        window.location.href = "/";
      }

    } catch (err) {
      console.error("Create post error:", err);
      alert("Unexpected error while creating post");
    }
  });
});
