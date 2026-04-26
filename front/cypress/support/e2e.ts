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
    err.message.includes("Invalid pointer id")
  ) {
    return false;
  }
});
