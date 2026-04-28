/// <reference types="cypress" />

type UserMode = "provider" | "customer";

declare global {
  // Cypress's TypeScript hook for custom commands is namespace augmentation.
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      login(token?: string): Chainable<void>;
      // Visits a URL with the user mode pre-seeded in localStorage so the
      // first paint already reflects it. ModeProvider reads localStorage
      // synchronously in its useState initializer.
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
});

Cypress.Commands.add("visitAs", (mode: UserMode, url: string) => {
  cy.visit(url, {
    onBeforeLoad(win) {
      win.localStorage.setItem("app.user-mode", mode);
    },
  });
});

export {};
