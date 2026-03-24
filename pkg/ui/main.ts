import "htmx.org";
import { animate, hover, press } from "motion";

document.addEventListener('htmx:load', function (evt) {
    console.log("HTMX loaded new content");

    // 1. Initial Entrance Animation 
    // This replaces `initial={{ opacity: 0, y: 20 }}` and `animate={{ opacity: 1, y: 0 }}`
    // Since HTMX swaps content, we run this whenever new buttons are loaded into the DOM.
    const buttons = document.querySelectorAll(".btn-solid-3d");
    const title = document.getElementById("hero-title");
    const subtitle = document.getElementById("hero-subtitle");

    if (buttons.length > 0) {
        animate(
            ".btn-solid-3d",
            { opacity: [0, 1], y: [20, 0] },
            { duration: 0.8, delay: 0.4 }
        );
    }

    if (title) {
        animate(
            title,
            { opacity: [0, 1], x: [-20, 0] },
            { duration: 0.8 }
        );
    }

    if (subtitle) {
        animate(
            subtitle,
            { opacity: [0, 1], x: [-20, 0] },
            { duration: 0.8, delay: 0.4 }
        );
    }

    // 2. Interactive Animations (Hover and Press/Tap) are now managed
    // via pure CSS in main.css to keep the box-shadow base totally anchored.
});
