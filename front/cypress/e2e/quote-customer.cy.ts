import { line, quote, type LineFixture } from "../support/fixtures";

// Stub GET /api/quotes (customer mode sends X-Client-Mode: customer header).
function stubCustomerQuotesList() {
  cy.intercept("GET", /^\/api\/quotes(\?.*)?$/, {
    statusCode: 200,
    body: { success: true, quotes: [] },
  }).as("listMyQuotes");
}

describe("Quote — customer mode", () => {
  describe("Sidebar", () => {
    it("hides provider-only entries when in customer mode", () => {
      cy.login();
      cy.intercept("GET", /\/api\/users\/clients\/_/, {
        statusCode: 200,
        body: { success: true, clients: [{ client_id: "c-1", first_name: "Jean", last_name: "Dupont" }] },
      });
      stubCustomerQuotesList();
      cy.visitAs("customer", "/quote");
      cy.wait("@listMyQuotes");

      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Devis").should("exist");
        cy.contains("Factures").should("exist");
        cy.contains("Clients").should("not.exist");
        cy.contains("Pays").should("not.exist");
        cy.contains("Taxes").should("not.exist");
        cy.contains("Test").should("not.exist");
      });
    });

    it("mode-toggle button exists and carries the correct data-active state per mode", () => {
      cy.login();

      // Provider mode: toggle button not active, sidebar shows provider items.
      cy.visit("/quote");
      cy.get("[data-slot='mode-toggle']").should("not.have.attr", "data-active");
      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Clients").should("exist");
        cy.contains("Mon profil").should("not.exist");
      });

      // Customer mode: toggle button active, sidebar shows customer items.
      cy.intercept("GET", /\/api\/users\/clients\/_/, {
        statusCode: 200,
        body: { success: true, clients: [{ client_id: "c-1", first_name: "Jean", last_name: "Dupont" }] },
      });
      cy.visitAs("customer", "/quote");
      cy.get("[data-slot='mode-toggle']").should("have.attr", "data-active", "true");
      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Factures").should("exist");
        cy.contains("Mon profil").should("exist");
      });
    });
  });

  describe("List", () => {
    it("renders an empty list and skips the /api/quotes fetch for provider", () => {
      cy.login();
      cy.intercept("GET", /\/api\/users\/clients\/_/, {
        statusCode: 200,
        body: { success: true, clients: [{ client_id: "c-1", first_name: "Jean", last_name: "Dupont" }] },
      });
      stubCustomerQuotesList();

      cy.visitAs("customer", "/quote");
      cy.wait("@listMyQuotes");

      cy.contains("Aucun devis pour le moment.").should("be.visible");
      cy.contains("button", "Nouveau devis").should("not.exist");
      cy.contains("a", "Nouveau devis").should("not.exist");
    });
  });

  describe("Detail", () => {
    it("renders the form read-only and hides drop/continue actions", () => {
      cy.login();
      cy.intercept("GET", /\/api\/quotes\/q-1/, {
        statusCode: 200,
        body: { success: true, quote: quote(), lines: [line()] },
      }).as("getQuote");
      cy.intercept("GET", "/api/users/taxes/available**", {
        statusCode: 200,
        body: { success: true, taxes: [] },
      });
      cy.intercept("GET", /\/api\/users\/addresses/, {
        statusCode: 200,
        body: { success: true, addresses: [] },
      });

      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");
      cy.get("input[name='name']").should("be.disabled");
      cy.contains("button", "Changer l'état").should("not.exist");

      cy.get("[data-step-tab='1']").click();
      cy.get("[data-line-id='l-1'] input[name='line-name']").should("be.disabled");
      cy.get("[data-line-id='l-1'] input[name='line-quantity']").should("be.disabled");
      cy.get("[data-line-id='l-1'] input[name='line-unit-price']").should("be.disabled");
      cy.get("[aria-label='Ajouter une ligne']").should("not.exist");
      cy.get("[aria-label='Supprimer la ligne']").should("not.exist");
    });
  });

  describe("Adresse de livraison", () => {
    const lineFixture: LineFixture = line({ line_id: "l-1", quote_id: "q-1" });
    const address = {
      id: 10,
      owner_type: "client",
      owner_id: "c-1",
      name: "Siège",
      street: "10 rue de Paris",
      additional_street: "",
      city: "Paris",
      zip_code: "75001",
      country_id: 1,
      email: "",
      phone: "",
      archived: false,
    };

    beforeEach(() => {
      cy.login();
      cy.intercept("GET", /\/api\/quotes\/q-1/, {
        statusCode: 200,
        body: {
          success: true,
          quote: quote({ quote_id: "q-1", client_id: "c-1", address_id: 10 }),
          lines: [lineFixture],
        },
      }).as("getQuote");
      cy.intercept("GET", "/api/users/taxes/available**", {
        statusCode: 200,
        body: { success: true, taxes: [] },
      });
      cy.intercept("GET", /\/api\/users\/clients\/_/, {
        statusCode: 200,
        body: { success: true, clients: [{ client_id: "c-1", first_name: "Jean", last_name: "Dupont" }] },
      });
      cy.intercept("GET", /\/api\/users\/addresses(\?.*)?$/, {
        statusCode: 200,
        body: { success: true, addresses: [address] },
      }).as("listClientAddresses");
    });

    it("le champ adresse de livraison est modifiable en mode client", () => {
      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");
      cy.wait("@listClientAddresses");

      cy.get("input[name='address_id']").should("not.be.disabled");
    });

    it("le champ nom du projet reste disabled en mode client", () => {
      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");

      cy.get("input[name='name']").should("be.disabled");
    });

    it("enregistre la nouvelle adresse via PUT /api/quotes/q-1", () => {
      cy.intercept("PUT", /\/api\/quotes\/q-1/, {
        statusCode: 200,
        body: { success: true },
      }).as("updateAddress");

      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");
      cy.wait("@listClientAddresses");

      cy.get("input[name='address_id']").click();
      cy.contains("[data-slot='combobox-item']", "Siège").click({ force: true });
      cy.wait("@updateAddress");
    });
  });

  describe("Commentaires en mode client", () => {
    const lineFixture: LineFixture = line({ line_id: "l-1", quote_id: "q-1", name: "Design UI" });

    beforeEach(() => {
      cy.login();
      cy.intercept("GET", /\/api\/quotes\/q-1/, {
        statusCode: 200,
        body: {
          success: true,
          quote: quote({ quote_id: "q-1" }),
          lines: [lineFixture],
        },
      }).as("getQuote");
      cy.intercept("GET", "/api/users/taxes/available**", {
        statusCode: 200,
        body: { success: true, taxes: [] },
      });
      cy.intercept("GET", /\/api\/users\/addresses/, {
        statusCode: 200,
        body: { success: true, addresses: [] },
      });
      cy.intercept("GET", /\/api\/users\/clients\/_/, {
        statusCode: 200,
        body: {
          success: true,
          clients: [{ client_id: "c-1", first_name: "Jean", last_name: "Dupont" }],
        },
      });
    });

    it("le bouton Commentaires par ligne est visible en mode client", () => {
      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");

      cy.get("[data-step-tab='1']").click();
      cy.get("[aria-label='Commentaires de la ligne']").should("be.visible");
    });

    it("ouvre la sidebar de commentaires depuis une ligne en mode client", () => {
      cy.intercept("GET", "/api/quotes/q-1/lines/l-1/comments", {
        statusCode: 200,
        body: { success: true, comments: [] },
      }).as("listComments");

      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");

      cy.get("[data-step-tab='1']").click();
      cy.get("[aria-label='Commentaires de la ligne']").first().click();
      cy.wait("@listComments");

      cy.contains("Commentaires — Design UI").should("be.visible");
    });

    it("un client peut poster un commentaire", () => {
      const newComment = {
        comment_id: "cmt-1",
        line_id: "l-1",
        quote_id: "q-1",
        author_id: "test-user",
        author_name: "Jean Dupont",
        body: "Question sur ce poste",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      cy.intercept("GET", "/api/quotes/q-1/lines/l-1/comments", {
        statusCode: 200,
        body: { success: true, comments: [] },
      }).as("listComments");
      cy.intercept("POST", "/api/quotes/q-1/lines/l-1/comments", {
        statusCode: 201,
        body: { success: true, comment: newComment },
      }).as("createComment");

      cy.visitAs("customer", "/quote/q-1");
      cy.wait("@getQuote");

      cy.get("[data-step-tab='1']").click();
      cy.get("[aria-label='Commentaires de la ligne']").first().click();
      cy.wait("@listComments");

      cy.get("textarea[placeholder='Écrire un commentaire…']").type("Question sur ce poste");
      cy.contains("button", "Envoyer").click();
      cy.wait("@createComment");

      cy.contains("Question sur ce poste").should("be.visible");
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
      cy.intercept("GET", "/api/quotes**", {
        statusCode: 200,
        body: { success: true, quotes: [quote()] },
      }).as("listQuotes");

      cy.visit("/quote");
      cy.wait("@listQuotes");

      cy.contains("button", "Nouveau devis").should("be.visible");
      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Clients").should("exist");
      });
    });
  });
});
