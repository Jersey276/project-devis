describe("Fees page", () => {
  const FEES = [
    {
      fee_id: "fee-1",
      category: "service",
      name: "Livraison",
      unit: "h",
      unit_price: 5000,
      tax_id: null,
      archived: false,
    },
  ];

  const TAXES = [
    {
      id: 100,
      name: "TVA 20",
      rate: "20.00",
      country_group_id: 10,
      is_default: false,
      version: 1,
    },
  ];

  // cy.login() already stubs /api/auth/me with a pro tier, which clears the
  // SubscriptionGuard. The fee catalog and the dialog's tax dropdown are stubbed
  // here so no real backend is needed.
  function stubFeesPage(opts?: { fees?: typeof FEES }) {
    cy.login();
    cy.intercept("GET", "/api/fees**", {
      statusCode: 200,
      body: { success: true, fees: opts?.fees ?? FEES },
    }).as("getFees");
    cy.intercept("GET", "/api/users/taxes/available**", {
      statusCode: 200,
      body: { success: true, taxes: TAXES },
    }).as("getTaxes");
  }

  describe("Premium gating", () => {
    function stubTier(tier: "free" | "pro") {
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

    it("shows the upgrade message for free users", () => {
      cy.login();
      stubTier("free");

      cy.visit("/fees");
      cy.wait("@getAuthMe");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("be.visible");
    });

    it("renders the catalog for pro users", () => {
      cy.login();
      stubTier("pro");
      cy.intercept("GET", "/api/fees**", {
        statusCode: 200,
        body: { success: true, fees: [] },
      }).as("getFees");

      cy.visit("/fees");
      cy.wait("@getAuthMe");

      cy.contains(
        "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
      ).should("not.exist");
    });
  });

  describe("Structure", () => {
    beforeEach(() => stubFeesPage());

    it("lists fees with category label and formatted price", () => {
      cy.visit("/fees");
      cy.wait("@getFees");

      cy.contains("td", "Livraison").should("be.visible");
      cy.contains("td", "Prestation").should("be.visible");
      cy.contains("td", "h").should("be.visible");
      cy.contains("td", "50.00 €").should("be.visible");
    });

    it("shows the empty state when there is no fee", () => {
      stubFeesPage({ fees: [] });
      cy.visit("/fees");
      cy.wait("@getFees");

      cy.contains("Aucun frais pour le moment.").should("be.visible");
    });
  });

  describe("Create", () => {
    beforeEach(() => stubFeesPage());

    it("creates a fee, converting the euro price to cents", () => {
      cy.intercept("POST", "/api/fees", {
        statusCode: 201,
        body: { success: true, fee_id: "fee-2" },
      }).as("createFee");

      cy.visit("/fees");
      cy.wait("@getFees");

      cy.contains("button", "Nouveau frais").click();
      cy.get("[data-slot='dialog-content']").should("be.visible");

      // Category defaults to "fixed"; pick "Prestation" (service).
      cy.get("[data-slot='select-trigger']").click();
      cy.contains("[data-slot='select-item']", "Prestation").click();

      cy.get("input[name='name']").type("Frais de dossier");
      cy.get("input[name='unit_price']").type("12.50");
      cy.get("input[name='unit']").type("forfait");

      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@createFee").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          category: "service",
          name: "Frais de dossier",
          unit: "forfait",
          unit_price: 1250,
          tax_id: 0,
        });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Frais créé.");
      cy.get("[data-slot='dialog-content']").should("not.exist");
    });

    it("shows an inline error on 422 for the price", () => {
      cy.intercept("POST", "/api/fees", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "unit_price", error_code: [2] }],
        },
      }).as("createFeeInvalid");

      cy.visit("/fees");
      cy.wait("@getFees");

      cy.contains("button", "Nouveau frais").click();
      cy.get("input[name='name']").type("Mauvais prix");
      cy.get("input[name='unit_price']").type("5");
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@createFeeInvalid");
      cy.get("[data-slot='dialog-content']").should("be.visible");
    });
  });

  describe("Edit", () => {
    beforeEach(() => stubFeesPage());

    it("pre-fills the dialog and sends the full payload", () => {
      cy.intercept("PUT", "/api/fees/fee-1", {
        statusCode: 200,
        body: { success: true },
      }).as("updateFee");

      cy.visit("/fees");
      cy.wait("@getFees");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Modifier").click();

      cy.get("input[name='name']").should("have.value", "Livraison");
      cy.get("input[name='unit_price']").should("have.value", "50");
      cy.get("input[name='unit']").should("have.value", "h");

      cy.get("input[name='unit_price']").clear().type("60");
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@updateFee").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          category: "service",
          name: "Livraison",
          unit: "h",
          unit_price: 6000,
          tax_id: 0,
        });
      });
      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Frais mis à jour. Les devis non validés ont été actualisés.",
      );
    });
  });

  describe("Delete", () => {
    beforeEach(() => stubFeesPage());

    it("archives a fee (success)", () => {
      cy.intercept("DELETE", "/api/fees/fee-1", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteFee");

      cy.visit("/fees");
      cy.wait("@getFees");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteFee");
      cy.get("[data-sonner-toaster]").should("contain", "Frais supprimé.");
    });

    it("shows an error toast on delete failure", () => {
      cy.intercept("DELETE", "/api/fees/fee-1", {
        statusCode: 500,
        body: { success: false, message: "Échec serveur." },
      }).as("deleteFeeFail");

      cy.visit("/fees");
      cy.wait("@getFees");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteFeeFail");
      cy.get("[data-sonner-toaster]").should("contain", "Échec serveur.");
      cy.contains("td", "Livraison").should("be.visible");
    });
  });
});
