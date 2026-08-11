// Entry point for first-party browser JavaScript. esbuild bundles every module
// reachable from this file into static/js/dist.js, so this is the only script
// the browser loads besides the pinned vendor libraries.
//
// Sources live under assets/ rather than static/ because server.go embeds the
// whole static/ tree: anything placed there ships inside the binary. Keeping
// sources out means only the minified bundle is embedded and served.
import { userCard } from "./components/user-card.js";

// Alpine boots on DOMContentLoaded, so every Alpine.data() registration has to
// run first — on the alpine:init event. This bundle wins the race only because
// layout.templ loads it before the deferred Alpine script: deferred scripts run
// in document order, so moving this tag below Alpine silently breaks every
// component registered here.
document.addEventListener("alpine:init", () => {
  Alpine.data("userCard", userCard);
});
