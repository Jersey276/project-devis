describe("Taxes page", () => {
  const GROUPS = [
    { id: 10, name: "Union européenne", countries: [] },
    { id: 11, name: "Mercosur", countries: [] },
  ];

  const TAXES = [
    {
      id: 100,
      name: "TVA 20",
      rate: "20.00",
      country_group_id: 10,
      is_default: false,
    },
  ];

  function stubTaxesPage(opts?: {
    taxes?: typeof TAXES;
    groups?: typeof GROUPS;
  }) {
    cy.login();
    cy.intercept("GET", "/api/users/taxes", {
      statusCode: 200,
      body: { success: true, taxes: opts?.taxes ?? TAXES },
    }).as("getTaxes");
    cy.intercept("GET", "/api/users/country-groups", {
      statusCode: 200,
      body: { success: true, country_groups: opts?.groups ?? GROUPS },
    }).as("getCountryGroups");
  }

  describe("Structure", () => {
    beforeEach(() => stubTaxesPage());

    it("lists taxes and joins group name", () => {
      cy.visit("/taxes");
      cy.wait("@getTaxes");
      cy.wait("@getCountryGroups");

      cy.contains("td", "TVA 20").should("be.visible");
      cy.contains("td", "20.00 %").should("be.visible");
      cy.contains("td", "Union européenne").should("be.visible");
    });
  });

  describe("Create", () => {
    beforeEach(() => stubTaxesPage());

    it("creates a new tax (success)", () => {
      cy.intercept("POST", "/api/users/taxes", {
        statusCode: 201,
        body: { success: true, tax_id: 101 },
      }).as("createTax");

      cy.visit("/taxes");
      cy.wait("@getTaxes");
      cy.wait("@getCountryGroups");

      cy.contains("button", "Nouvelle taxe").click();
      cy.get("[data-slot='dialog-content']").should("be.visible");

      cy.get("input[name='name']").type("TVA réduite");
      cy.get("input[name='rate']").type("5.50");
      cy.get("input[name='country_group_id']").type("Mercosur");
      cy.contains("[data-slot='combobox-item']", "Mercosur").click({ force: true });

      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@createTax").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          name: "TVA réduite",
          rate: "5.50",
          country_group_id: 11,
          is_default: false,
        });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Taxe ajoutée.");
      cy.get("[data-slot='dialog-content']").should("not.exist");
    });

    it("shows inline error on 422 for rate", () => {
      cy.intercept("POST", "/api/users/taxes", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "rate", error_code: [2] }],
        },
      }).as("createTaxInvalid");

      cy.visit("/taxes");
      cy.wait("@getTaxes");
      cy.wait("@getCountryGroups");

      cy.contains("button", "Nouvelle taxe").click();
      cy.get("input[name='name']").type("Mauvais taux");
      cy.get("input[name='rate']").type("abc");
      cy.get("input[name='country_group_id']").type("Union");
      cy.contains("[data-slot='combobox-item']", "Union européenne").click({ force: true });
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@createTaxInvalid");

      cy.get("input[name='rate']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Format invalide.");
      cy.get("[data-slot='dialog-content']").should("be.visible");
    });
  });

  describe("Default toggle", () => {
    beforeEach(() => stubTaxesPage());

    it("sends is_default=true when the checkbox is ticked", () => {
      cy.intercept("POST", "/api/users/taxes", {
        statusCode: 201,
        body: { success: true, tax_id: 102 },
      }).as("createTaxDefault");

      cy.visit("/taxes");
      cy.wait("@getTaxes");
      cy.wait("@getCountryGroups");

      cy.contains("button", "Nouvelle taxe").click();
      cy.get("input[name='name']").type("TVA défaut");
      cy.get("input[name='rate']").type("20.00");
      cy.get("input[name='country_group_id']").type("Union");
      cy.contains("[data-slot='combobox-item']", "Union européenne").click({
        force: true,
      });
      cy.get("button#tax_is_default").click();
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@createTaxDefault").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          name: "TVA défaut",
          rate: "20.00",
          country_group_id: 10,
          is_default: true,
        });
      });
    });
  });

  describe("Edit", () => {
    beforeEach(() => stubTaxesPage());

    it("disables the group field and only sends name + rate", () => {
      cy.intercept("PUT", "/api/users/taxes/100", {
        statusCode: 200,
        body: { success: true },
      }).as("updateTax");

      cy.visit("/taxes");
      cy.wait("@getTaxes");
      cy.wait("@getCountryGroups");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Modifier").click();

      cy.get("input[name='name']").should("have.value", "TVA 20");
      cy.get("input[name='rate']").should("have.value", "20.00");
      cy.get("input[name='country_group_id']").should("be.disabled");

      cy.get("input[name='rate']").clear().type("21.00");
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@updateTax").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          name: "TVA 20",
          rate: "21.00",
          is_default: false,
        });
        expect(interception.request.body).not.to.have.property(
          "country_group_id",
        );
      });
      cy.get("[data-sonner-toaster]").should("contain", "Taxe mise à jour.");
    });
  });

  describe("Delete", () => {
    beforeEach(() => stubTaxesPage());

    it("deletes a tax (success)", () => {
      cy.intercept("DELETE", "/api/users/taxes/100", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteTax");

      cy.visit("/taxes");
      cy.wait("@getTaxes");
      cy.wait("@getCountryGroups");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteTax");
      cy.get("[data-sonner-toaster]").should("contain", "Taxe supprimée.");
    });

    it("shows error toast on delete failure", () => {
      cy.intercept("DELETE", "/api/users/taxes/100", {
        statusCode: 500,
        body: { success: false, message: "Échec serveur." },
      }).as("deleteTaxFail");

      cy.visit("/taxes");
      cy.wait("@getTaxes");
      cy.wait("@getCountryGroups");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteTaxFail");
      cy.get("[data-sonner-toaster]").should("contain", "Échec serveur.");
      cy.contains("td", "TVA 20").should("be.visible");
    });
  });
});
