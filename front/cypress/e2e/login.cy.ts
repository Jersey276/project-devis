describe("Login", () => {
  it("redirects to login page when not authenticated", () => {
    cy.visit("/");
    cy.url().should("include", "/login");
  });

  it("has a login form", () => {
    cy.visit("/login");
    cy.get("form").should("exist");
    cy.get("input[name='email']").should("exist");
    cy.get("input[name='password']").should("exist");
    cy.get("button[type='submit']").should("exist");
  });

  it("redirects to home on successful login", () => {
    // intercept MUST be set up before the action that triggers the request
    cy.intercept("POST", "/api/auth/login", {
      statusCode: 200,
      body: { success: true, token: "fake-token" },
    }).as("loginRequestSuccess");

    cy.visit("/login");
    cy.get("input[name='email']").type("test@test.fr");
    cy.get("input[name='password']").type("test");
    cy.get("button[type='submit']").click();

    cy.wait("@loginRequestSuccess")
      .its("response.statusCode")
      .should("eq", 200);
    cy.url().should("not.include", "/login");
  });

  it("shows error toast on invalid credentials", () => {
    cy.intercept("POST", "/api/auth/login", {
      statusCode: 401,
      body: { success: false, message: "Invalid credentials" },
    }).as("loginRequestInvalidCredentials");

    cy.visit("/login");
    cy.get("input[name='email']").type("wrong@test.fr");
    cy.get("input[name='password']").type("wrongpassword");
    cy.get("button[type='submit']").click();

    cy.wait("@loginRequestInvalidCredentials")
      .its("response.statusCode")
      .should("eq", 401);
    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Login failed. Please check your credentials and try again.",
    );
  });

  it("shows error toast when server is unavailable", () => {
    cy.intercept("POST", "/api/auth/login", {
      forceNetworkError: true,
    }).as("loginRequestServerUnavailable");

    cy.visit("/login");
    cy.get("input[name='email']").type("test@test.fr");
    cy.get("input[name='password']").type("test");
    cy.get("button[type='submit']").click();

    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Login failed. Please check your credentials and try again.",
    );
  });
});
