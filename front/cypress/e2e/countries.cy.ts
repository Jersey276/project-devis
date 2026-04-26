describe("Countries page", () => {
  const COUNTRIES = [
    { id: 1, code: "FR", name: "France" },
    { id: 2, code: "BE", name: "Belgique" },
  ];

  const GROUPS = [
    {
      id: 10,
      name: "Union européenne",
      countries: [{ id: 1, code: "FR", name: "France" }],
    },
  ];

  function stubCountriesPage(opts?: {
    countries?: typeof COUNTRIES;
    groups?: typeof GROUPS;
  }) {
    cy.login();
    cy.intercept("GET", "/api/users/countries", {
      statusCode: 200,
      body: { success: true, countries: opts?.countries ?? COUNTRIES },
    }).as("getCountries");
    cy.intercept("GET", "/api/users/country-groups", {
      statusCode: 200,
      body: { success: true, country_groups: opts?.groups ?? GROUPS },
    }).as("getCountryGroups");
  }

  describe("Structure", () => {
    beforeEach(() => stubCountriesPage());

    it("shows the two sub-tabs", () => {
      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.get("[role='tablist']").should("exist");
      cy.get("[role='tab']").should("have.length", 2);
      cy.contains("[role='tab']", "Pays").should("exist");
      cy.contains("[role='tab']", "Groupes de pays").should("exist");
    });

    it("switches tabs on click", () => {
      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("[role='tab']", "Groupes de pays").click();
      cy.wait("@getCountryGroups");
      cy.contains("button", "Nouveau groupe").should("be.visible");
    });
  });

  describe("Sidebar", () => {
    beforeEach(() => stubCountriesPage());

    it("exposes Pays and Taxes entries", () => {
      cy.intercept("GET", "/api/users/me", {
        statusCode: 200,
        body: { success: true, user: { user_id: "u1", email: "x@y.z" } },
      });
      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("a", "Pays").should("have.attr", "href", "/countries");
      cy.contains("a", "Taxes").should("have.attr", "href", "/taxes");
    });
  });

  describe("Pays tab", () => {
    beforeEach(() => stubCountriesPage());

    it("lists countries from the API", () => {
      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("td", "FR").should("be.visible");
      cy.contains("td", "France").should("be.visible");
      cy.contains("td", "BE").should("be.visible");
      cy.contains("td", "Belgique").should("be.visible");
    });

    it("creates a new country (success)", () => {
      cy.intercept("POST", "/api/users/countries", {
        statusCode: 201,
        body: { success: true, country_id: 3 },
      }).as("createCountry");

      cy.visit("/countries");
      cy.wait("@getCountries");

      cy.contains("button", "Nouveau pays").click();
      cy.get("[data-slot='dialog-content']").should("be.visible");
      cy.get("input[name='code']").type("ES");
      cy.get("input[name='name']").type("Espagne");
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@createCountry").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          code: "ES",
          name: "Espagne",
        });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Pays ajouté.");
      cy.get("[data-slot='dialog-content']").should("not.exist");
    });

    it("shows inline errors on 422 when creating", () => {
      cy.intercept("POST", "/api/users/countries", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "code", error_code: [1] }],
        },
      }).as("createCountryInvalid");

      cy.visit("/countries");
      cy.wait("@getCountries");

      cy.contains("button", "Nouveau pays").click();
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();
      cy.wait("@createCountryInvalid");

      cy.get("input[name='code']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Ce champ est requis.");
      cy.get("[data-slot='dialog-content']").should("be.visible");
    });

    it("edits an existing country", () => {
      cy.intercept("PUT", "/api/users/countries/1", {
        statusCode: 200,
        body: { success: true },
      }).as("updateCountry");

      cy.visit("/countries");
      cy.wait("@getCountries");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Modifier").click();
      cy.get("input[name='name']").should("have.value", "France");
      cy.get("input[name='name']").clear().type("République française");
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@updateCountry").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          code: "FR",
          name: "République française",
        });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Pays mis à jour.");
    });

    it("cancels deletion (no API call)", () => {
      cy.intercept("DELETE", "/api/users/countries/1", () => {
        throw new Error("DELETE should not be called when canceling");
      });

      cy.visit("/countries");
      cy.wait("@getCountries");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.get("[data-slot='alert-dialog-content']").should("be.visible");
      cy.contains("[data-slot='alert-dialog-cancel']", "Annuler").click();
      cy.get("[data-slot='alert-dialog-content']").should("not.exist");
      cy.contains("td", "France").should("be.visible");
    });

    it("deletes a country (success)", () => {
      cy.intercept("DELETE", "/api/users/countries/1", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteCountry");

      cy.visit("/countries");
      cy.wait("@getCountries");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteCountry");
      cy.get("[data-sonner-toaster]").should("contain", "Pays supprimé.");
    });

    it("shows error toast on delete failure", () => {
      cy.intercept("DELETE", "/api/users/countries/1", {
        statusCode: 500,
        body: { success: false, message: "Échec serveur." },
      }).as("deleteCountryFail");

      cy.visit("/countries");
      cy.wait("@getCountries");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteCountryFail");
      cy.get("[data-sonner-toaster]").should("contain", "Échec serveur.");
      cy.contains("td", "France").should("be.visible");
    });
  });

  describe("Groupes de pays tab", () => {
    beforeEach(() => stubCountriesPage());

    it("lists groups with members", () => {
      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("[role='tab']", "Groupes de pays").click();
      cy.wait("@getCountryGroups");

      cy.contains("td", "Union européenne").should("be.visible");
      cy.contains("td", "France").should("be.visible");
    });

    it("creates a new group", () => {
      cy.intercept("POST", "/api/users/country-groups", {
        statusCode: 201,
        body: { success: true, country_group_id: 11 },
      }).as("createGroup");

      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("[role='tab']", "Groupes de pays").click();
      cy.wait("@getCountryGroups");

      cy.contains("button", "Nouveau groupe").click();
      cy.get("[data-slot='dialog-content']").should("be.visible");
      cy.get("input[name='name']").type("Mercosur");
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@createGroup").then((interception) => {
        expect(interception.request.body).to.deep.equal({ name: "Mercosur" });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Groupe ajouté.");
    });

    it("renames a group", () => {
      cy.intercept("PUT", "/api/users/country-groups/10", {
        statusCode: 200,
        body: { success: true },
      }).as("updateGroup");

      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("[role='tab']", "Groupes de pays").click();
      cy.wait("@getCountryGroups");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Modifier").click();

      cy.get("input[name='name']").should("have.value", "Union européenne");
      cy.get("input[name='name']").clear().type("UE");
      cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();

      cy.wait("@updateGroup").then((interception) => {
        expect(interception.request.body).to.deep.equal({ name: "UE" });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Groupe mis à jour.");
    });

    it("attaches a country to a group", () => {
      cy.intercept("POST", "/api/users/country-groups/10/countries/2", {
        statusCode: 200,
        body: { success: true },
      }).as("attach");
      cy.intercept("GET", "/api/users/country-groups/10", {
        statusCode: 200,
        body: {
          success: true,
          country_group: {
            id: 10,
            name: "Union européenne",
            countries: [
              { id: 1, code: "FR", name: "France" },
              { id: 2, code: "BE", name: "Belgique" },
            ],
          },
        },
      }).as("refreshGroup");

      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("[role='tab']", "Groupes de pays").click();
      cy.wait("@getCountryGroups");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Modifier").click();

      cy.get("[data-slot='group-members']").within(() => {
        cy.get("input[name='attach_country_id']").type("Belg");
      });
      cy.contains("[data-slot='combobox-item']", "Belgique").click({ force: true });
      cy.contains("[data-slot='group-members'] button", "Ajouter").click();

      cy.wait("@attach");
      cy.wait("@refreshGroup");
      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Pays ajouté au groupe.",
      );
      cy.get("[data-slot='group-member']").should("have.length", 2);
    });

    it("detaches a country from a group", () => {
      cy.intercept("DELETE", "/api/users/country-groups/10/countries/1", {
        statusCode: 200,
        body: { success: true },
      }).as("detach");
      cy.intercept("GET", "/api/users/country-groups/10", {
        statusCode: 200,
        body: {
          success: true,
          country_group: {
            id: 10,
            name: "Union européenne",
            countries: [],
          },
        },
      }).as("refreshGroup");

      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("[role='tab']", "Groupes de pays").click();
      cy.wait("@getCountryGroups");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Modifier").click();

      cy.get("[data-slot='group-member-remove']").first().click();

      cy.wait("@detach");
      cy.wait("@refreshGroup");
      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Pays retiré du groupe.",
      );
      cy.get("[data-slot='group-member']").should("have.length", 0);
    });

    it("deletes a group", () => {
      cy.intercept("DELETE", "/api/users/country-groups/10", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteGroup");

      cy.visit("/countries");
      cy.wait("@getCountries");
      cy.contains("[role='tab']", "Groupes de pays").click();
      cy.wait("@getCountryGroups");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteGroup");
      cy.get("[data-sonner-toaster]").should("contain", "Groupe supprimé.");
    });
  });
});
