import { client, type ClientFixture } from "../support/fixtures";

const COUNTRIES = [{ id: 1, code: "FR", name: "France" }];

function stubCountries() {
  cy.intercept("GET", "/api/users/countries", {
    statusCode: 200,
    body: { success: true, countries: COUNTRIES },
  }).as("getCountries");
}

function stubList(clients: ClientFixture[]) {
  cy.intercept("GET", "/api/users/clients", {
    statusCode: 200,
    body: { success: true, clients },
  }).as("listClients");
}

function stubGet(c: ClientFixture) {
  cy.intercept("GET", `/api/users/clients/${c.client_id}`, {
    statusCode: 200,
    body: { success: true, client: c },
  }).as("getClient");
}

describe("Clients", () => {
  describe("List", () => {
    it("renders clients returned by the API", () => {
      cy.login();
      stubList([
        client({ client_id: "c-1", first_name: "Jean", last_name: "Dupont" }),
        client({
          client_id: "c-2",
          first_name: "Marie",
          last_name: "Martin",
          email: "marie@example.com",
          company: "Beta",
        }),
      ]);

      cy.visit("/clients");
      cy.wait("@listClients");

      cy.contains("td", "Jean").should("be.visible");
      cy.contains("td", "Marie").should("be.visible");
      cy.contains("td", "marie@example.com").should("be.visible");
      cy.contains("td", "Beta").should("be.visible");
    });

    it("links 'Nouveau client' to the create page", () => {
      cy.login();
      stubList([]);
      cy.visit("/clients");
      cy.contains("a", "Nouveau client").should(
        "have.attr",
        "href",
        "/clients/create",
      );
    });
  });

  describe("Create", () => {
    beforeEach(() => {
      cy.login();
      stubCountries();
    });

    it("creates a client and an optional address, then redirects to the profile", () => {
      cy.intercept("POST", "/api/users/clients", {
        statusCode: 201,
        body: { success: true, client_id: "c-new" },
      }).as("createClient");
      cy.intercept("POST", "/api/users/addresses", {
        statusCode: 201,
        body: { success: true, address_id: 99 },
      }).as("createAddress");
      stubGet(client({ client_id: "c-new", first_name: "Alice", last_name: "Martin" }));

      cy.visit("/clients/create");
      cy.get("input[name='first_name']").type("Alice");
      cy.get("input[name='last_name']").type("Martin");
      cy.get("input[name='email']").type("alice@example.com");

      // Fill address (optional, but exercise the dual-create path)
      cy.get("input[name='street']").type("12 rue de Rivoli");
      cy.get("input[name='city']").type("Paris");
      cy.get("input[name='zip_code']").type("75001");
      cy.wait("@getCountries");
      cy.get("input[name='country_id']").click();
      cy.contains("[data-slot='combobox-item']", "France").click({ force: true });

      cy.contains("button", "Créer le compte client").click();

      cy.wait("@createClient").then(({ request }) => {
        expect(request.body).to.include({
          first_name: "Alice",
          last_name: "Martin",
          email: "alice@example.com",
        });
      });
      cy.wait("@createAddress").then(({ request }) => {
        expect(request.body).to.include({
          owner_type: "client",
          owner_id: "c-new",
          street: "12 rue de Rivoli",
          city: "Paris",
          zip_code: "75001",
          country_id: 1,
        });
      });
      cy.url().should("match", /\/clients\/c-new$/);
    });

    it("surfaces 422 field errors and stays on the create page", () => {
      cy.intercept("POST", "/api/users/clients", {
        statusCode: 422,
        body: {
          success: false,
          // Gateway shape: { field, error_code: number[] }; codes are mapped
          // to messages on the client (see FIELD_VALIDATION_MESSAGES).
          // 2 → "Format invalide."
          field_errors: [{ field: "email", error_code: [2] }],
        },
      }).as("createInvalid");

      cy.visit("/clients/create");
      cy.get("input[name='first_name']").type("Bob");
      cy.get("input[name='last_name']").type("Smith");
      // Use a syntactically valid value so the browser's HTML5 type=email
      // check doesn't block submit; the 422 is delivered by the stub.
      cy.get("input[name='email']").type("bob@example.com");
      cy.contains("button", "Créer le compte client").click();
      cy.wait("@createInvalid");

      cy.contains("Format invalide.").should("be.visible");
      cy.url().should("include", "/clients/create");
    });
  });

  describe("Profile", () => {
    beforeEach(() => {
      cy.login();
      stubCountries();
    });

    it("renders client info from the API", () => {
      const c = client({
        client_id: "c-1",
        first_name: "Jean",
        last_name: "Dupont",
        email: "jean@example.com",
        phone: "0612345678",
        company: "Acme",
        siren: "123456789",
        vat: "FR12345",
      });
      stubGet(c);

      cy.visit("/clients/c-1");
      cy.wait("@getClient");

      cy.contains("Profil client").should("be.visible");
      cy.contains("Jean").should("be.visible");
      cy.contains("Dupont").should("be.visible");
      cy.contains("jean@example.com").should("be.visible");
      cy.contains("0612345678").should("be.visible");
      cy.contains("Acme").should("be.visible");
      cy.contains("123456789").should("be.visible");
    });

    it("links 'Modifier' to the edit page", () => {
      stubGet(client());
      cy.visit("/clients/c-1");
      cy.wait("@getClient");
      cy.contains("a", "Modifier").should(
        "have.attr",
        "href",
        "/clients/c-1/edit",
      );
    });

    it("archives the client from the profile and redirects to the list", () => {
      stubGet(client());
      stubList([]);
      cy.intercept("DELETE", "/api/users/clients/c-1", {
        statusCode: 200,
        body: { success: true },
      }).as("archiveClient");

      cy.visit("/clients/c-1");
      cy.wait("@getClient");

      cy.contains("button", "Supprimer").click();
      // Confirm in the AlertDialog
      cy.get("[role='alertdialog']").contains("button", "Supprimer").click();

      cy.wait("@archiveClient");
      cy.url().should("match", /\/clients$/);
    });
  });

  describe("Edit", () => {
    beforeEach(() => {
      cy.login();
      stubCountries();
    });

    it("loads existing values, sends updateClient, redirects to profile", () => {
      const c = client({
        client_id: "c-1",
        first_name: "Jean",
        last_name: "Dupont",
        email: "jean@example.com",
        company: "Acme",
      });
      stubGet(c);
      cy.intercept("PUT", "/api/users/clients/c-1", {
        statusCode: 200,
        body: { success: true },
      }).as("updateClient");

      cy.visit("/clients/c-1/edit");
      cy.wait("@getClient");

      cy.get("input[name='first_name']").should("have.value", "Jean");
      cy.get("input[name='last_name']").should("have.value", "Dupont");
      cy.get("input[name='email']").should("have.value", "jean@example.com");
      cy.get("input[name='company']").should("have.value", "Acme");

      // The edit form must NOT render the embedded address section.
      cy.contains("Adresse principale").should("not.exist");

      cy.get("input[name='first_name']").clear().type("Jeanne");
      cy.get("input[name='company']").clear().type("Acme SA");

      cy.contains("button", "Enregistrer").click();

      cy.wait("@updateClient").then(({ request }) => {
        expect(request.body).to.include({
          first_name: "Jeanne",
          last_name: "Dupont",
          email: "jean@example.com",
          company: "Acme SA",
        });
      });
      cy.url().should("match", /\/clients\/c-1$/);
    });

    it("surfaces 422 field errors and stays on the edit page", () => {
      stubGet(client());
      cy.intercept("PUT", "/api/users/clients/c-1", {
        statusCode: 422,
        body: {
          success: false,
          // 2 → "Format invalide." (see FIELD_VALIDATION_MESSAGES).
          field_errors: [{ field: "email", error_code: [2] }],
        },
      }).as("updateInvalid");

      cy.visit("/clients/c-1/edit");
      cy.wait("@getClient");
      // Syntactically valid; 422 comes from the stub regardless.
      cy.get("input[name='email']").clear().type("changed@example.com");
      cy.contains("button", "Enregistrer").click();
      cy.wait("@updateInvalid");

      cy.contains("Format invalide.").should("be.visible");
      cy.url().should("include", "/clients/c-1/edit");
    });

    it("redirects to the list when the client cannot be loaded", () => {
      cy.intercept("GET", "/api/users/clients/ghost", {
        statusCode: 404,
        body: { success: false, message: "Client introuvable." },
      }).as("getGhost");

      cy.visit("/clients/ghost/edit");
      cy.wait("@getGhost");
      cy.url().should("match", /\/clients$/);
    });
  });

  describe("Archive from list", () => {
    it("archives a client and reloads the list", () => {
      cy.login();
      // First load: one client.
      cy.intercept("GET", "/api/users/clients", (req) => {
        req.reply({
          statusCode: 200,
          body: {
            success: true,
            clients: [client({ client_id: "c-1", first_name: "Jean" })],
          },
        });
      }).as("listFirst");

      cy.visit("/clients");
      cy.wait("@listFirst");
      cy.contains("td", "Jean").should("be.visible");

      // Re-stub: after archive, the list comes back empty.
      cy.intercept("DELETE", "/api/users/clients/c-1", {
        statusCode: 200,
        body: { success: true },
      }).as("archive");
      cy.intercept("GET", "/api/users/clients", {
        statusCode: 200,
        body: { success: true, clients: [] },
      }).as("listAfter");

      // Open the row action menu and click Supprimer.
      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("[role='menuitem']", "Supprimer").click();

      cy.wait("@archive");
      cy.wait("@listAfter");
      cy.contains("td", "Jean").should("not.exist");
    });
  });
});
