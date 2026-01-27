// web/static/js/password-toggle.js
export function initPasswordToggles(root = document) {
  root.querySelectorAll(".password-toggle").forEach((btn) => {
    if (btn.dataset.bound) return;
    btn.dataset.bound = "1";

    btn.addEventListener("click", () => {
      const input = btn
        .closest(".password-input-wrap")
        ?.querySelector("input");

      if (!input) return;

      const isPassword = input.type === "password";
      input.type = isPassword ? "text" : "password";

      btn.setAttribute(
        "aria-label",
        isPassword ? "Hide password" : "Show password"
      );
    });
  });
}
