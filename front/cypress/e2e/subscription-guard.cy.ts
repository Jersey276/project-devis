describe("SubscriptionGuard", () => {
  describe("Templates page", () => {
    it("shows upgrade message for free users", () => {
      cy.login({ tier: "free" });

      cy.visit("/templates");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("be.visible");
    });

    it("renders templates for pro users", () => {
      cy.login({ tier: "pro" });

      cy.intercept("GET", "/api/templates**", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getTemplates");

      cy.visit("/templates");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("not.exist");
    });

    it("renders templates for enterprise users", () => {
      cy.login({ tier: "enterprise" });

      cy.intercept("GET", "/api/templates**", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getTemplates");

      cy.visit("/templates");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("not.exist");
    });
  });

  describe("Schedule page", () => {
    it("shows upgrade message for free users", () => {
      cy.login({ tier: "free" });

      cy.visit("/schedule");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("be.visible");
    });

    it("renders schedule list for pro users", () => {
      cy.login({ tier: "pro" });

      cy.intercept("GET", "/api/schedules**", {
        statusCode: 200,
        body: { success: true, schedules: [] },
      }).as("getSchedules");

      cy.visit("/schedule");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("not.exist");
    });
  });
});
