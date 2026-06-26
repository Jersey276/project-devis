describe("OAuth login", () => {
  const providers = ["google", "github", "microsoft"];

  it("renders an OAuth link per provider on the login page", () => {
    cy.visit("/login");
    providers.forEach((provider) => {
      cy.get(`a[href*='/api/auth/oauth/${provider}']`)
        .should("exist")
        .and("have.attr", "href")
        .and("include", "next=");
    });
  });

  it("renders an OAuth link per provider on the register page", () => {
    cy.visit("/register");
    providers.forEach((provider) => {
      cy.get(`a[href*='/api/auth/oauth/${provider}']`).should("exist");
    });
  });

  it("carries the next param through the OAuth link from login", () => {
    cy.visit("/login?next=%2Fquote%2F42");
    cy.get("a[href*='/api/auth/oauth/google']")
      .should("have.attr", "href")
      .and("include", encodeURIComponent("/quote/42"));
  });

  it("lands on the app after a successful OAuth callback", () => {
    // The gateway callback sets the auth cookie and 302s to the landing path.
    cy.intercept("GET", "/api/auth/oauth/google/callback*", {
      statusCode: 302,
      headers: {
        location: "/quote",
        "set-cookie": "auth-token=fake-token; Path=/",
      },
    }).as("oauthCallback");

    cy.visit("/api/auth/oauth/google/callback?code=abc&state=xyz", {
      failOnStatusCode: false,
    });
    cy.url().should("not.include", "/login");
  });

  it("shows an error toast when redirected back with oauth_error", () => {
    cy.visit("/login?oauth_error=provider");
    cy.contains("La connexion avec le fournisseur a échoué").should("be.visible");
  });
});

describe("OAuth account linking (profile)", () => {
  const USER = {
    user_id: "u1",
    email: "john@test.fr",
    phone: "",
    company: "Acme",
    siren: "",
    vat: "",
  };

  function stubMe() {
    cy.login();
    cy.intercept("GET", "/api/users/me", {
      statusCode: 200,
      body: { success: true, user: USER },
    }).as("getMe");
    cy.intercept("GET", "/api/users/addresses?**", {
      statusCode: 200,
      body: { success: true, addresses: [] },
    });
  }

  it("renders link buttons for unlinked providers and the linked list", () => {
    stubMe();
    cy.intercept("GET", "/api/auth/oauth-identities", {
      statusCode: 200,
      body: {
        success: true,
        has_password: true,
        identities: [{ provider: "google", email: "john@test.fr" }],
      },
    }).as("getIdentities");

    cy.visit("/profile?tab=compte");
    cy.wait("@getMe");
    cy.wait("@getIdentities");

    // Google is linked → shown in the list with a "Délier" button.
    cy.contains("Google").should("exist");
    cy.contains("button", "Délier").should("exist");

    // GitHub/Microsoft are not linked → link buttons present, Google link hidden.
    cy.get("a[href*='/api/auth/oauth-link/github']").should("exist");
    cy.get("a[href*='/api/auth/oauth-link/microsoft']").should("exist");
    cy.get("a[href*='/api/auth/oauth-link/google']").should("not.exist");
  });

  it("disables unlink for the only login method (no password)", () => {
    stubMe();
    cy.intercept("GET", "/api/auth/oauth-identities", {
      statusCode: 200,
      body: {
        success: true,
        has_password: false,
        identities: [{ provider: "google", email: "john@test.fr" }],
      },
    }).as("getIdentities");

    cy.visit("/profile?tab=compte");
    cy.wait("@getIdentities");

    cy.contains("button", "Délier").should("be.disabled");
    cy.contains("Définissez un mot de passe").should("be.visible");
  });

  it("unlinks a provider and removes it from the list", () => {
    stubMe();
    cy.intercept("GET", "/api/auth/oauth-identities", {
      statusCode: 200,
      body: {
        success: true,
        has_password: true,
        identities: [{ provider: "google", email: "john@test.fr" }],
      },
    }).as("getIdentities");
    cy.intercept("DELETE", "/api/auth/oauth-identities/google", {
      statusCode: 200,
      body: { success: true },
    }).as("unlink");

    cy.visit("/profile?tab=compte");
    cy.wait("@getIdentities");

    cy.contains("button", "Délier").click();
    cy.wait("@unlink");
    cy.get("[data-sonner-toaster]").should("contain", "Fournisseur dissocié.");
  });

  it("toasts the link outcome from the callback redirect", () => {
    stubMe();
    cy.intercept("GET", "/api/auth/oauth-identities", {
      statusCode: 200,
      body: { success: true, has_password: true, identities: [] },
    }).as("getIdentities");

    cy.visit("/profile?tab=compte&oauth_linked=google");
    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Fournisseur lié à votre compte.",
    );
  });

  it("toasts an error when the identity is already taken", () => {
    stubMe();
    cy.intercept("GET", "/api/auth/oauth-identities", {
      statusCode: 200,
      body: { success: true, has_password: true, identities: [] },
    }).as("getIdentities");

    cy.visit("/profile?tab=compte&oauth_error=identity_taken");
    cy.contains("déjà lié à un autre utilisateur").should("be.visible");
  });
});
