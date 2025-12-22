document.addEventListener(
  "DOMContentLoaded",
  () => {
    // Stagger card animation
    gsap.from(".glass-card, .card", {
      duration: 1,
      y: 30,
      opacity: 0,
      stagger: 0.2,
      ease: "power4.out",
    });

    // Input focus animation
    const inputs = document.querySelectorAll(
      ".input-group-custom, .form-control"
    );
    inputs.forEach((input) => {
      input.addEventListener("focusin", () => {
        gsap.to(input, {
          scale: 1.01,
          duration: 0.3,
          ease: "power2.out",
        });
      });
      input.addEventListener("focusout", () => {
        gsap.to(input, {
          scale: 1,
          duration: 0.3,
          ease: "power2.out",
        });
      });
    });
  }
);

const LinkAnimate = {
  // Success animation
  animateSuccess: function () {
    const target =
      document.querySelector(".card");
    if (target) {
      gsap.fromTo(
        target,
        { scale: 0.98, opacity: 0.1 },
        {
          scale: 1,
          opacity: 10,
          duration: 0.9,
          ease: "back.out(2)",

          clearProps: "all",
        }
      );
    }
  },

  // Falling animation
  animateTableRows: function () {
    const rows = document.querySelectorAll(
      "#linkList tr"
    );
    if (rows.length > 0) {
      gsap.from(rows, {
        duration: 0.4,
        opacity: 0,
        y: 10,
        stagger: 0.05,
        ease: "power2.out",
        clearProps: "all",
      });
    }
  },

  // Shake animation
  shakeElement: function (elementId) {
    const el =
      document.getElementById(elementId);
    if (el) {
      gsap.to(el, {
        duration: 0.1,
        x: 6,
        repeat: 5,
        yoyo: true,
        onComplete: () =>
          gsap.set(el, { x: 0 }),
      });
    }
  },

  // Bounce btn
  bounceButton: function (btnElement) {
    if (btnElement) {
      gsap.to(btnElement, {
        scale: 1.2,
        duration: 0.1,
        yoyo: true,
        repeat: 1,
        ease: "power1.inOut",
        clearProps: "all",
      });
    }
  },
};

globalThis.LinkAnimate = LinkAnimate;
globalThis.animateSuccess =
  LinkAnimate.animateSuccess;
