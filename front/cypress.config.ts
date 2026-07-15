import { defineConfig } from "cypress";

export default defineConfig({
  e2e: {
    baseUrl: "http://localhost:3000",
    setupNodeEvents(_on, config) {
      // Opt-in flag for the handful of specs that exercise a real backend
      // (no cy.intercept mocking) instead of the default fully-mocked suite.
      // Only CI's e2e-integration job sets CYPRESS_REAL_BACKEND — everywhere
      // else (front-quality, local `npm run cy:open`) it's unset, so those
      // specs stay describe.skip'd.
      config.env.realBackend = process.env.CYPRESS_REAL_BACKEND === "1";
      return config;
    },
  },
});
