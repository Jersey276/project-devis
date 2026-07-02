describe("Login", () => {
  it("redirects to login page with next= when not authenticated", () => {
    cy.clearCookies();
    cy.visit("/", { failOnStatusCode: false });
    cy.url().should("include", "/login");
    cy.url().should("include", "next=%2F");
  });

  it("has a login form", () => {
    cy.visit("/login");
    cy.get("form").should("exist");
    cy.get("input[name='email']").should("exist");
    cy.get("input[name='password']").should("exist");
    cy.get("input[name='remember_me']").should("exist");
    cy.get("button[type='submit']").should("exist");
  });

  it("redirects to home on successful login", () => {
    // Set-Cookie matches what the real gateway returns; without auth-token the
    // middleware would block the post-login navigation and bounce back to /login.
    cy.intercept("POST", "/api/auth/login", {
      statusCode: 200,
      headers: {
        "set-cookie": "auth-token=fake-token; Path=/",
      },
      body: { success: true, token: "fake-token" },
    }).as("loginRequestSuccess");

    cy.visit("/login");
    cy.fillLoginForm("test@test.fr", "test");

    cy.wait("@loginRequestSuccess")
      .its("response.statusCode")
      .should("eq", 200);
    cy.url().should("not.include", "/login");
  });

  it("sends remember_me when the checkbox is checked", () => {
    cy.intercept("POST", "/api/auth/login", {
      statusCode: 200,
      headers: {
        "set-cookie": "auth-token=fake-token; Path=/",
      },
      body: { success: true, token: "fake-token" },
    }).as("loginRequestRememberMe");

    cy.visit("/login");
    cy.get("input[name='email']").type("test@test.fr");
    cy.get("input[name='password']").type("test");
    cy.get("input[name='remember_me']").check({ force: true });
    cy.get("button[type='submit']").click();

    cy.wait("@loginRequestRememberMe")
      .its("request.body")
      .should("deep.include", { remember_me: true });
  });

  it("sends remember_me=false when the checkbox is left unchecked", () => {
    cy.intercept("POST", "/api/auth/login", {
      statusCode: 200,
      headers: {
        "set-cookie": "auth-token=fake-token; Path=/",
      },
      body: { success: true, token: "fake-token" },
    }).as("loginRequestNoRememberMe");

    cy.visit("/login");
    cy.fillLoginForm("test@test.fr", "test");

    cy.wait("@loginRequestNoRememberMe")
      .its("request.body")
      .should("deep.include", { remember_me: false });
  });

  it("shows error toast on invalid credentials", () => {
    cy.intercept("POST", "/api/auth/login", {
      statusCode: 401,
      body: { success: false, message: "Invalid credentials" },
    }).as("loginRequestInvalidCredentials");

    cy.visit("/login");
    cy.fillLoginForm("wrong@test.fr", "wrongpassword");

    cy.wait("@loginRequestInvalidCredentials")
      .its("response.statusCode")
      .should("eq", 401);
    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Échec de la connexion. Vérifiez vos identifiants et réessayez.",
    );
  });

  it("shows error toast when server is unavailable", () => {
    cy.intercept("POST", "/api/auth/login", {
      forceNetworkError: true,
    }).as("loginRequestServerUnavailable");

    cy.visit("/login");
    cy.fillLoginForm("test@test.fr", "test");

    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Échec de la connexion. Vérifiez vos identifiants et réessayez.",
    );
  });

  it("shows success toast when arriving with ?deleted=true", () => {
    cy.visit("/login?deleted=true");
    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Votre compte a été supprimé avec succès.",
    );
  });
});
