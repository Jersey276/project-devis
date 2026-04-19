describe("Register", () => {
  it("has a register form", () => {
    cy.visit("/register");
    cy.get("form").should("exist");
    cy.get("input[name='name']").should("exist");
    cy.get("input[name='email']").should("exist");
    cy.get("input[name='password']").should("exist");
    cy.get("input[name='confirm-password']").should("exist");
    cy.get("button[type='submit']").should("exist");
  });

  it("redirects to login on successful registration", () => {
    cy.intercept("POST", "/api/auth/register", {
      statusCode: 200,
      body: { success: true },
    }).as("registerSuccess");

    cy.visit("/register");
    cy.get("input[name='name']").type("John Doe");
    cy.get("input[name='email']").type("john@test.fr");
    cy.get("input[name='password']").type("password123");
    cy.get("input[name='confirm-password']").type("password123");
    cy.get("button[type='submit']").click();

    cy.wait("@registerSuccess").its("response.statusCode").should("eq", 200);
    cy.url().should("include", "/login");
  });

  it("shows success toast on successful registration", () => {
    cy.intercept("POST", "/api/auth/register", {
      statusCode: 200,
      body: { success: true },
    }).as("registerSuccessToast");

    cy.visit("/register");
    cy.get("input[name='name']").type("John Doe");
    cy.get("input[name='email']").type("john@test.fr");
    cy.get("input[name='password']").type("password123");
    cy.get("input[name='confirm-password']").type("password123");
    cy.get("button[type='submit']").click();

    cy.wait("@registerSuccessToast");
    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Registration successful! Please log in."
    );
  });

  it("shows error toast on failed registration", () => {
    cy.intercept("POST", "/api/auth/register", {
      statusCode: 400,
      body: { success: false, message: "Email already exists" },
    }).as("registerFailure");

    cy.visit("/register");
    cy.get("input[name='name']").type("John Doe");
    cy.get("input[name='email']").type("existing@test.fr");
    cy.get("input[name='password']").type("password123");
    cy.get("input[name='confirm-password']").type("password123");
    cy.get("button[type='submit']").click();

    cy.wait("@registerFailure").its("response.statusCode").should("eq", 400);
    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Registration failed. Please try again."
    );
  });

  it("shows error toast when server is unavailable", () => {
    cy.intercept("POST", "/api/auth/register", {
      forceNetworkError: true,
    }).as("registerNetworkError");

    cy.visit("/register");
    cy.get("input[name='name']").type("John Doe");
    cy.get("input[name='email']").type("john@test.fr");
    cy.get("input[name='password']").type("password123");
    cy.get("input[name='confirm-password']").type("password123");
    cy.get("button[type='submit']").click();

    cy.get("[data-sonner-toaster]").should(
      "contain",
      "Registration failed. Please try again."
    );
  });

  describe("form validation", () => {
    it("shows confirm password mismatch error without calling the API", () => {
      cy.intercept("POST", "/api/auth/register", (_req) => {
        throw new Error("API should not be called on confirm password mismatch");
      }).as("registerMismatch");

      cy.visit("/register");
      cy.get("input[name='name']").type("John Doe");
      cy.get("input[name='email']").type("john@test.fr");
      cy.get("input[name='password']").type("password123");
      cy.get("input[name='confirm-password']").type("different");
      cy.get("button[type='submit']").click();

      cy.get("input[name='confirm-password']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Passwords do not match.");
    });

    it("shows inline field error when email is already in use", () => {
      cy.intercept("POST", "/api/auth/register", {
        statusCode: 422,
        body: {
          success: false,
          code: 1001,
          field_errors: [{ field: "email", error_code: [4] }],
        },
      }).as("registerEmailInUse");

      cy.visit("/register");
      cy.get("input[name='name']").type("John Doe");
      cy.get("input[name='email']").type("existing@test.fr");
      cy.get("input[name='password']").type("password123");
      cy.get("input[name='confirm-password']").type("password123");
      cy.get("button[type='submit']").click();

      cy.wait("@registerEmailInUse");
      cy.get("input[name='email']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "This email address is already in use.");
    });

    it("shows inline field errors for required, invalid format, and too short", () => {
      cy.intercept("POST", "/api/auth/register", {
        statusCode: 422,
        body: {
          success: false,
          code: 2002,
          field_errors: [
            { field: "name", error_code: [1] },
            { field: "email", error_code: [2] },
            { field: "password", error_code: [3] },
          ],
        },
      }).as("registerValidationErrors");

      cy.visit("/register");
      cy.get("input[name='name']").type("x");
      cy.get("input[name='email']").type("bad");
      cy.get("input[name='password']").type("short");
      cy.get("input[name='confirm-password']").type("short");
      cy.get("button[type='submit']").click();

      cy.wait("@registerValidationErrors");

      cy.get("input[name='name']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "This field is required.");

      cy.get("input[name='email']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Invalid format.");

      cy.get("input[name='password']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Too short (minimum 8 characters).");
    });

    it("clears field errors when submitting again", () => {
      cy.intercept("POST", "/api/auth/register", {
        statusCode: 422,
        body: {
          success: false,
          code: 1001,
          field_errors: [{ field: "email", error_code: [4] }],
        },
      }).as("registerClearErrorsFirst");

      cy.visit("/register");
      cy.get("input[name='name']").type("John Doe");
      cy.get("input[name='email']").type("existing@test.fr");
      cy.get("input[name='password']").type("password123");
      cy.get("input[name='confirm-password']").type("password123");
      cy.get("button[type='submit']").click();

      cy.wait("@registerClearErrorsFirst");
      cy.get("input[name='email']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("exist");

      cy.intercept("POST", "/api/auth/register", {
        statusCode: 200,
        body: { success: true },
      }).as("registerClearErrorsSecond");

      cy.get("input[name='email']").clear().type("new@test.fr");
      cy.get("button[type='submit']").click();

      cy.wait("@registerClearErrorsSecond");
      cy.get("input[name='email']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("not.exist");
    });
  });
});
