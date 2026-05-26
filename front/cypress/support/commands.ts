/// <reference types="cypress" />

import { AUTH_TOKEN_COOKIE } from "../../lib/auth-constants";

type UserMode = "provider" | "customer";

declare global {
  // Cypress's TypeScript hook for custom commands is namespace augmentation.
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      login(token?: string): Chainable<void>;
      // Mode is currently inert: ModeProvider hard-codes "provider" while
      // customer mode is disabled at the UI level. The signature is kept so
      // call sites stay stable when the toggle is restored.
      visitAs(mode: UserMode, url: string): Chainable<void>;
      fillLoginForm(email: string, password: string): Chainable<void>;
    }
  }
}

Cypress.Commands.add("login", (token = "fake-token") => {
  cy.setCookie(AUTH_TOKEN_COOKIE, token, { domain: "localhost" });
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

Cypress.Commands.add("visitAs", (_mode: UserMode, url: string) => {
  cy.visit(url);
});

Cypress.Commands.add("fillLoginForm", (email: string, password: string) => {
  cy.get("input[name='email']").type(email);
  cy.get("input[name='password']").type(password);
  cy.get("button[type='submit']").click();
});

export {};
