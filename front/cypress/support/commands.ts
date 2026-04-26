/// <reference types="cypress" />

declare global {
  namespace Cypress {
    interface Chainable {
      login(token?: string): Chainable<void>;
    }
  }
}

Cypress.Commands.add("login", (token = "fake-token") => {
  cy.setCookie("auth-token", token, { domain: "localhost" });
  // Sidebar UserMenu calls /api/users/me on mount; an unstubbed 401 triggers
  // a client-side redirect to /login via apiFetch. Tests can override this.
  cy.intercept("GET", "/api/users/me", {
    statusCode: 200,
    body: {
      success: true,
      user: { user_id: "test-user", email: "test@test.fr" },
    },
  });
});

export {};
