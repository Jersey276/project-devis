describe("Auth middleware - SSR refresh on missing access token", () => {
  it("redirects to /login with next= when no auth cookies are present", () => {
    cy.clearCookies();
    cy.visit("/quote", { failOnStatusCode: false });
    cy.url().should("include", "/login");
    cy.url().should("include", "next=%2Fquote");
  });

  it("preserves query string of the protected route in next=", () => {
    cy.clearCookies();
    cy.visit("/quote?filter=open", { failOnStatusCode: false });
    cy.url().should("include", "/login");
    cy.url().should("include", "next=%2Fquote%3Ffilter%3Dopen");
  });

  describe("after successful login", () => {
    beforeEach(() => {
      // Set-Cookie matches what the real gateway returns; without auth-token the
      // middleware would block the post-login navigation and bounce back to /login.
      cy.intercept("POST", "/api/auth/login", {
        statusCode: 200,
        headers: {
          "set-cookie": "auth-token=fake-token; Path=/",
        },
        body: { success: true, token: "fake-token" },
      }).as("login");
    });

    it("redirects back to next=", () => {
      cy.visit("/login?next=/clients");
      cy.fillLoginForm("test@test.fr", "test");

      cy.wait("@login").its("response.statusCode").should("eq", 200);
      cy.url().should("include", "/clients");
      cy.url().should("not.include", "/login");
    });

    it("ignores absolute external next= (open-redirect guard)", () => {
      cy.visit("/login?next=https://evil.example.com/path");
      cy.fillLoginForm("test@test.fr", "test");

      cy.wait("@login");
      cy.url().should("not.include", "evil.example.com");
      cy.location("pathname").should("eq", "/");
    });

    it("ignores protocol-relative next= (open-redirect guard)", () => {
      cy.visit("/login?next=//evil.example.com");
      cy.fillLoginForm("test@test.fr", "test");

      cy.wait("@login");
      cy.url().should("not.include", "evil.example.com");
      cy.location("pathname").should("eq", "/");
    });
  });
});
