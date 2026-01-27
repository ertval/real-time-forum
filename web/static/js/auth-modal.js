// web/static/js/auth-modal.js

let modalLoaded = false;
let modal = null;
let frame = null;
let isOpen = false;

export async function loadAuthModal() {
  if (modalLoaded) return;

  const res = await fetch("/static/partials/auth-modal.html");
  const html = await res.text();

  document.body.insertAdjacentHTML("beforeend", html);

  modal = document.getElementById("auth-modal");
  frame = document.getElementById("auth-frame");

  modal.querySelector(".auth-close").onclick =
  modal.querySelector(".auth-backdrop").onclick =
    closeAuthModal;

  modalLoaded = true;
}

export async function openAuthModal(path = "/login") {
  if (isOpen) return;

  if (!modalLoaded) {
    await loadAuthModal();
  }

  // load correct page (login / register)
  frame.src = path;

  modal.classList.remove("hidden");
  document.body.style.overflow = "hidden";

  isOpen = true;
}

export function closeAuthModal() {
  if (!modal) return;

  modal.classList.add("hidden");
  document.body.style.overflow = "";
  isOpen = false;
}
