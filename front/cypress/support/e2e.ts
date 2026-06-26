import "./commands";

// Ignore React's hydration-mismatch error so tests are not failed by it.
// It originates from the shadcn Sidebar's `<SidebarMenuButton asChild><Link/></SidebarMenuButton>`
// composition under React 19 + Next 16 and is unrelated to the behavior under test.
// Other uncaught exceptions still fail tests.
Cypress.on("uncaught:exception", (err) => {
  if (
    err.message.includes("Hydration failed") ||
    err.message.includes("server rendered HTML didn't match the client") ||
    // Radix UI sometimes calls setPointerCapture with a synthesized pointer id
    // that the browser rejects under Cypress. Harmless — ignore.
    err.message.includes("Invalid pointer id") ||
    // Background fetches that race with navigation (Next.js RSC prefetch, HMR
    // ping, Link hover-prefetch) can reject after the page is gone, surfacing
    // as Firefox's "NetworkError when attempting to fetch resource" or
    // "AbortError: The operation was aborted." Unrelated to the assertions
    // under test.
    err.message.includes("NetworkError when attempting to fetch resource") ||
    err.message.includes("The operation was aborted") ||
    // Chrome's variant of the same abort signal.
    err.message.includes("signal is aborted without reason") ||
    // Stripe.js cannot load its external script in the Cypress sandbox. The
    // payment dialog guards against a null stripePromise, so this is harmless.
    err.message.includes("Failed to load Stripe.js") ||
    // Stripe Elements validates clientSecret format; in tests we pass a fake
    // secret that doesn't match the ${id}secret${secret} pattern.
    err.message.includes("Invalid value for elements()") ||
    // Edge's HTTP cache can serve a stale RSC chunk after a page navigation,
    // causing Next.js to throw this during module factory lookup. Harmless —
    // the page re-renders correctly on the next tick.
    err.message.includes("module factory is not available")
  ) {
    return false;
  }
});
