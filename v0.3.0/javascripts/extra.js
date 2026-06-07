// Extra JS for teststop docs

// Animate confidence bars when they come into view
document.addEventListener("DOMContentLoaded", () => {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const fill = entry.target.querySelector(".ts-confidence-bar-fill");
        if (fill) {
          const width = fill.dataset.width || "0%";
          fill.style.width = width;
        }
        observer.unobserve(entry.target);
      }
    });
  });

  document.querySelectorAll(".ts-confidence-bar").forEach((bar) => {
    const fill = bar.querySelector(".ts-confidence-bar-fill");
    if (fill) {
      const targetWidth = fill.style.width;
      fill.dataset.width = targetWidth;
      fill.style.width = "0%";
    }
    observer.observe(bar);
  });
});
