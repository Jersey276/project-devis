/// <reference types="cypress" />

import { AUTH_TOKEN_COOKIE } from "../../lib/auth-constants";

type UserMode = "provider" | "customer";

type LoginOptions = {
  tier?: "free" | "pro" | "enterprise";
  role?: string;
};


declare global {
  // Cypress's TypeScript hook for custom commands is namespace augmentation.
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      login(opts?: LoginOptions | string): Chainable<void>;
      visitAs(mode: UserMode, url: string): Chainable<void>;
      fillLoginForm(email: string, password: string): Chainable<void>;
    }
  }
}

Cypress.Commands.add("login", (opts: LoginOptions | string = {}) => {
  const tier = typeof opts === "string" ? "pro" : (opts.tier ?? "pro");
  const role = typeof opts === "string" ? "admin" : (opts.role ?? "admin");

  const authPayload = {
    user_id: "test-user",
    email: "test@test.fr",
    role,
    account_status: "active",
    subscription_tier: tier,
  };

  cy.setCookie(AUTH_TOKEN_COOKIE, "fake-token", { domain: "localhost" });
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
  cy.intercept("GET", "/api/auth/me", {
    statusCode: 200,
    body: { success: true, auth: authPayload },
  });
  cy.intercept("POST", "/api/auth/refresh", {
    statusCode: 200,
    body: { success: true },
  });
  cy.intercept("GET", "/api/quotes**", {
    statusCode: 200,
    body: { success: true, quotes: [] },
  });
  cy.intercept("GET", "/api/users/taxes/available**", {
    statusCode: 200,
    body: { success: true, taxes: [] },
  });
  cy.intercept("GET", /^\/api\/users\/clients(\?.*)?$/, {
    statusCode: 200,
    body: { success: true, clients: [] },
  });
  cy.intercept("GET", "/api/users/clients/me", {
    statusCode: 200,
    body: { success: true, clients: [] },
  });
  cy.intercept("GET", /\/api\/users\/clients\/me\/addresses/, {
    statusCode: 200,
    body: { success: true, addresses: [] },
  });
  cy.intercept("GET", /^\/api\/users\/addresses(\?.*)?$/, {
    statusCode: 200,
    body: { success: true, addresses: [] },
  });
});

Cypress.Commands.add("visitAs", (mode: UserMode, url: string) => {
  cy.setCookie("user-mode", mode, { domain: "localhost" });
  cy.visit(url);
});

Cypress.Commands.add("fillLoginForm", (email: string, password: string) => {
  // data-hydrated is set by the login form after React mounts, ensuring the
  // onSubmit handler is attached before we click submit.
  cy.get("form[data-hydrated]");
  cy.get("input[name='email']").type(email);
  cy.get("input[name='password']").type(password);
  cy.get("button[type='submit']").click();
});

export {};
