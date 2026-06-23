describe("SubscriptionGuard", () => {
  function stubAuthWithTier(tier: "free" | "pro" | "enterprise") {
    cy.intercept("GET", "/api/auth/me", {
      statusCode: 200,
      body: {
        success: true,
        auth: {
          user_id: "test-user",
          email: "test@test.fr",
          role: "free_user",
          account_status: "active",
          subscription_tier: tier,
        },
      },
    }).as("getAuthMe");
  }

  describe("Templates page", () => {
    it("shows upgrade message for free users", () => {
      cy.login();
      stubAuthWithTier("free");

      cy.visit("/templates");
      cy.wait("@getAuthMe");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("be.visible");
    });

    it("renders templates for pro users", () => {
      cy.login();
      stubAuthWithTier("pro");

      cy.intercept("GET", "/api/templates**", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getTemplates");

      cy.visit("/templates");
      cy.wait("@getAuthMe");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("not.exist");
    });

    it("renders templates for enterprise users", () => {
      cy.login();
      stubAuthWithTier("enterprise");

      cy.intercept("GET", "/api/templates**", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getTemplates");

      cy.visit("/templates");
      cy.wait("@getAuthMe");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("not.exist");
    });
  });

  describe("Schedule page", () => {
    it("shows upgrade message for free users", () => {
      cy.login();
      stubAuthWithTier("free");

      cy.visit("/schedule");
      cy.wait("@getAuthMe");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("be.visible");
    });

    it("renders schedule list for pro users", () => {
      cy.login();
      stubAuthWithTier("pro");

      cy.intercept("GET", "/api/schedules**", {
        statusCode: 200,
        body: { success: true, schedules: [] },
      }).as("getSchedules");

      cy.visit("/schedule");
      cy.wait("@getAuthMe");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("not.exist");
    });
  });
});
