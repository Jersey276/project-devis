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
    // as Firefox's "NetworkError when attempting to fetch resource". Unrelated
    // to the assertions under test.
    err.message.includes("NetworkError when attempting to fetch resource") ||
    // Stripe.js cannot load its external script in the Cypress sandbox. The
    // payment dialog guards against a null stripePromise, so this is harmless.
    err.message.includes("Failed to load Stripe.js") ||
    // Stripe Elements validates clientSecret format; in tests we pass a fake
    // secret that doesn't match the ${id}secret${secret} pattern.
    err.message.includes("Invalid value for elements()")
  ) {
    return false;
  }
});
