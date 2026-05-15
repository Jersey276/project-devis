import { line, quote } from "../support/fixtures";

describe("Quote — customer mode", () => {
  describe("Sidebar", () => {
    it("hides provider-only entries when in customer mode", () => {
      cy.login();
      cy.visitAs("customer", "/quote");

      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Devis").should("exist");
        cy.contains("Factures").should("not.exist");
        cy.contains("Clients").should("not.exist");
        cy.contains("Pays").should("not.exist");
        cy.contains("Taxes").should("not.exist");
        cy.contains("Test").should("not.exist");
      });
    });

    it("toggles between provider and customer from the sidebar mode button", () => {
      cy.login();
      cy.visit("/quote");

      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Factures").should("exist");
      });

      cy.get("[data-slot='mode-toggle']").click();

      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Factures").should("not.exist");
        cy.contains("Devis").should("exist");
      });
      cy.get("[data-slot='mode-toggle']").should(
        "have.attr",
        "data-active",
        "true",
      );

      cy.get("[data-slot='mode-toggle']").click();

      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Factures").should("exist");
      });
      cy.get("[data-slot='mode-toggle']").should("not.have.attr", "data-active");
    });
  });

  describe("List", () => {
    it("renders an empty list and skips the /api/quotes fetch", () => {
      cy.login();
      // Stubbed but should never fire — we assert below that the call count is 0.
      cy.intercept("GET", "/api/quotes", {
        statusCode: 200,
        body: { success: true, quotes: [quote()] },
      }).as("listQuotes");

      cy.visitAs("customer", "/quote");

      cy.contains("Aucun devis pour le moment.").should("be.visible");
      cy.contains("button", "Nouveau devis").should("not.exist");
      cy.contains("a", "Nouveau devis").should("not.exist");

      // The empty-state visibility above guarantees the page has rendered;
      // QuoteListTable's effect skips the fetch synchronously in customer mode.
      cy.get("@listQuotes.all").should("have.length", 0);
    });
  });

  describe("Detail", () => {
    it("renders the form read-only and hides drop/continue actions", () => {
      cy.login();
      cy.intercept("GET", "/api/users/taxes/available", {
        statusCode: 200,
        body: { success: true, taxes: [] },
      }).as("listAvailableTaxes");
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote(), lines: [line()] },
      }).as("getQuote");

      // Step 1 — basic info: project name input is disabled.
      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");
      cy.get("input[name='name']").should("be.disabled");

      // No drop/continue actions in the header regardless of step.
      cy.contains("button", "Abandonner").should("not.exist");
      cy.contains("button", "Continuer").should("not.exist");

      // Step 2 — lines: every line input is disabled, no add/remove buttons.
      cy.get("[data-step-tab='1']").click();
      cy.get("[data-line-id='l-1'] input[name='line-name']").should(
        "be.disabled",
      );
      cy.get("[data-line-id='l-1'] input[name='line-quantity']").should(
        "be.disabled",
      );
      cy.get("[data-line-id='l-1'] input[name='line-unit-price']").should(
        "be.disabled",
      );
      cy.get("[aria-label='Ajouter une ligne']").should("not.exist");
      cy.get("[aria-label='Supprimer la ligne']").should("not.exist");
    });
  });

  describe("/quote/create", () => {
    it("redirects back to /quote in customer mode", () => {
      cy.login();
      cy.visitAs("customer", "/quote/create");
      cy.location("pathname").should("eq", "/quote");
    });
  });

  describe("Provider regression", () => {
    it("keeps the create button and fetches /api/quotes in provider mode", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes", {
        statusCode: 200,
        body: { success: true, quotes: [quote()] },
      }).as("listQuotes");

      cy.visit("/quote");
      cy.wait("@listQuotes");

      cy.contains("a", "Nouveau devis").should("be.visible");
      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Factures").should("exist");
      });
    });
  });
});
