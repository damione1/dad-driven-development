/* Native buttons, keyboard navigation and focus restoration for the theme menu. */
(() => {
  const trigger = document.getElementById("menu-button");
  const close = document.getElementById("menu-close-button");
  const panel = document.getElementById("menu-wrapper");
  if (!trigger || !close || !panel) return;
  let previousOverflow;
  const setOpen = (open, restoreFocus = true) => {
    if (open === !panel.hidden) return;
    trigger.setAttribute("aria-expanded", String(open));
    panel.hidden = !open;
    panel.inert = !open;
    if (open) {
      previousOverflow = document.body.style.overflowY;
      document.body.style.overflowY = "hidden";
      close.focus();
    } else {
      document.body.style.overflowY = previousOverflow;
      if (restoreFocus) trigger.focus();
    }
  };
  trigger.addEventListener("click", () => setOpen(panel.hidden));
  close.addEventListener("click", () => setOpen(false));
  panel.addEventListener("click", (event) => {
    if (event.target === panel || event.target.closest("a")) setOpen(false);
  });
  panel.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
    } else if (event.key === "Tab") {
      const focusable = Array.from(panel.querySelectorAll('button:not([disabled]), a[href], input:not([disabled]), [tabindex="0"]')).filter(element => element.getClientRects().length);
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }
  });
  window.addEventListener("pagehide", () => setOpen(false, false));
  window.matchMedia("(min-width: 768px)").addEventListener("change", (event) => {
    if (event.matches) setOpen(false, false);
  });
})();
