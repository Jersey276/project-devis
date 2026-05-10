/// <reference types="cypress" />

type UserMode = "provider" | "customer";

declare global {
  // Cypress's TypeScript hook for custom commands is namespace augmentation.
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      login(token?: string): Chainable<void>;
      // Visits a URL with the user mode pre-seeded as a cookie so SSR and
      // hydration agree on the mode. The server-side layout reads
      // `app.user-mode` and feeds it to ModeProvider as initialMode.
      visitAs(mode: UserMode, url: string): Chainable<void>;
    }
  }
}

Cypress.Commands.add("login", (token = "fake-token") => {
  cy.setCookie("auth-token", token, { domain: "localhost" });
  // Default-stub the auth-sensitive endpoints so an unstubbed call never
  // bubbles up as a 401 → /api/auth/refresh failure → window.location = /login
  // inside apiFetch. Individual tests can override these intercepts.
  cy.intercept("GET", "/api/users/me", {
    statusCode: 200,
    body: {
      success: true,
      user: { user_id: "test-user", email: "test@test.fr" },
    },
  });
  cy.intercept("POST", "/api/auth/refresh", {
    statusCode: 200,
    body: { success: true },
  });
  cy.intercept("GET", "/api/quotes**", {
    statusCode: 200,
    body: { success: true, quotes: [] },
  });
  cy.intercept("GET", /^\/api\/users\/clients(\?.*)?$/, {
    statusCode: 200,
    body: { success: true, clients: [] },
  });
  cy.intercept("GET", /^\/api\/users\/addresses(\?.*)?$/, {
    statusCode: 200,
    body: { success: true, addresses: [] },
  });
});

Cypress.Commands.add("visitAs", (mode: UserMode, url: string) => {
  cy.setCookie("app.user-mode", mode, { domain: "localhost" });
  cy.visit(url);
});

export {};
