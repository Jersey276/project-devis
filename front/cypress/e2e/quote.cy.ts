import {
  line,
  type LineFixture,
  quote,
  tax,
  type TaxFixture,
} from "../support/fixtures";

function stubAvailableTaxes(taxes: TaxFixture[] = []) {
  cy.intercept("GET", "/api/users/taxes/available**", {
    statusCode: 200,
    body: { success: true, taxes },
  }).as("listAvailableTaxes");
}

describe("Quote", () => {
  describe("List", () => {
    it("renders quotes and maps state + archived_at to status", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes", {
        statusCode: 200,
        body: {
          success: true,
          quotes: [
            quote({ quote_id: "q-1", name: "Devis Alpha" }),
            quote({
              quote_id: "q-2",
              name: "Devis Beta",
              archived_at: "2026-04-01T00:00:00Z",
            }),
            quote({
              quote_id: "q-3",
              name: "Devis Gamma",
              state: "drop",
            }),
          ],
        },
      }).as("listQuotes");

      cy.visit("/quote");
      cy.wait("@listQuotes");

      cy.contains("td", "Devis Alpha").should("be.visible");
      cy.contains("td", "brouillon").should("be.visible");
      cy.contains("td", "Devis Beta").should("be.visible");
      cy.contains("td", "archivé").should("be.visible");
      cy.contains("td", "Devis Gamma").should("be.visible");
      cy.contains("td", "abandonné").should("be.visible");
    });

    it("shows the empty state when no quotes are returned", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes", {
        statusCode: 200,
        body: { success: true, quotes: [] },
      }).as("listQuotesEmpty");

      cy.visit("/quote");
      cy.wait("@listQuotesEmpty");
      cy.contains("Aucun devis pour le moment.").should("be.visible");
    });
  });

  describe("Create", () => {
    beforeEach(() => stubAvailableTaxes());

    function stubClientsAndAddresses() {
      cy.intercept("GET", "/api/users/clients**", {
        statusCode: 200,
        body: {
          success: true,
          clients: [
            {
              client_id: "c-1",
              user_id: "u-1",
              first_name: "Jean",
              last_name: "Dupont",
              email: "jean@example.com",
              phone: "",
              company: "Acme",
              siren: "",
              vat: "",
              archived: false,
            },
          ],
        },
      }).as("listClients");

      cy.intercept(
        "GET",
        "/api/users/addresses?owner_type=client&owner_id=c-1",
        {
          statusCode: 200,
          body: {
            success: true,
            addresses: [
              {
                id: 1,
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
              },
            ],
          },
        },
      ).as("listClientAddresses");
    }

    function fillStep1() {
      cy.get("input[name='name']").type("Nouveau devis");
      cy.get("input[name='client_id']").type("Jean");
      cy.contains("[data-slot='combobox-item']", "Jean Dupont").click({
        force: true,
      });
      cy.wait("@listClientAddresses");
      cy.get("input[name='address_id']").click();
      cy.contains("[data-slot='combobox-item']", "Siège").click({
        force: true,
      });
    }

    it("creates a quote and redirects to step 2", () => {
      cy.login();
      stubClientsAndAddresses();
      cy.intercept("POST", "/api/quotes", {
        statusCode: 201,
        body: { success: true, quote_id: "q-new" },
      }).as("createQuote");
      cy.intercept("GET", "/api/quotes/q-new", {
        statusCode: 200,
        body: {
          success: true,
          quote: quote({
            quote_id: "q-new",
            name: "Nouveau devis",
            client_id: "c-1",
            address_id: 1,
          }),
          lines: [],
        },
      }).as("getNewQuote");

      cy.visit("/quote/create");
      cy.wait("@listClients");
      fillStep1();
      cy.contains("button", "Suivant").click();

      cy.wait("@createQuote").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          name: "Nouveau devis",
          client_id: "c-1",
          address_id: 1,
        });
      });

      cy.url().should("include", "/quote/q-new");
      cy.url().should("include", "step=2");
      cy.wait("@getNewQuote");
      cy.get("[data-step-tab='1'][data-active='true']").should("exist");
    });

    it("surfaces field errors on 422", () => {
      cy.login();
      stubClientsAndAddresses();
      cy.intercept("POST", "/api/quotes", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "name", error_code: [1] }],
        },
      }).as("createQuoteInvalid");

      cy.visit("/quote/create");
      cy.wait("@listClients");
      fillStep1();
      cy.contains("button", "Suivant").click();
      cy.wait("@createQuoteInvalid");

      cy.get("input[name='name']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Ce champ est requis.");
      cy.url().should("include", "/quote/create");
    });

    it("blocks Suivant locally when client or address is missing", () => {
      cy.login();
      stubClientsAndAddresses();

      cy.visit("/quote/create");
      cy.wait("@listClients");

      cy.get("input[name='name']").type("Sans client");
      cy.contains("button", "Suivant").should("be.disabled");
    });
  });

  describe("Step 2 — lines", () => {
    function stubGet(lines: LineFixture[], taxes: TaxFixture[] = []) {
      cy.login();
      stubAvailableTaxes(taxes);
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote(), lines },
      }).as("getQuote");
    }

    it("renders existing lines on /quote/q-1?step=2", () => {
      stubGet([line()]);
      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");
      cy.get("[data-line-id='l-1']").should("exist");
      cy.get("[data-line-id='l-1'] input[name='line-name']").should(
        "have.value",
        "Design UI",
      );
      cy.get("[data-line-id='l-1'] input[name='line-unit-price']").should(
        "have.value",
        "80",
      );
    });

    it("adds a line via POST /lines", () => {
      stubGet([line()]);
      cy.intercept("POST", "/api/quotes/q-1/lines", {
        statusCode: 201,
        body: { success: true, line_id: "l-2" },
      }).as("createLine");

      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");
      cy.get("[aria-label='Ajouter une ligne']").click();

      cy.wait("@createLine").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          type: "simple",
          name: "",
          quantity: "1",
          unit: "",
          unit_price: 0,
          data: {},
          position: 1,
          tax_id: 0,
        });
      });
      cy.get("[data-line-id='l-2']").should("exist");
    });

    it("auto-saves a line edit (debounced) and converts price to cents", () => {
      stubGet([line()]);
      cy.intercept("PUT", "/api/quotes/q-1/lines/l-1", {
        statusCode: 200,
        body: { success: true },
      }).as("updateLine");

      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");

      cy.get("[data-line-id='l-1'] input[name='line-name']")
        .clear()
        .type("Design UI v2");
      cy.get(
        "[data-line-id='l-1'] [data-slot='line-save-indicator'][data-status='saving']",
      ).should("exist");

      cy.wait("@updateLine").then((interception) => {
        expect(interception.request.body).to.include({
          type: "simple",
          name: "Design UI v2",
          quantity: "10",
          unit_price: 8000,
          position: 0,
        });
      });

      cy.get(
        "[data-line-id='l-1'] [data-slot='line-save-indicator'][data-status='saved']",
      ).should("exist");
    });

    it("deletes a line via DELETE /lines/:id", () => {
      stubGet([
        line({ line_id: "l-1" }),
        line({ line_id: "l-2", position: 1 }),
      ]);
      cy.intercept("DELETE", "/api/quotes/q-1/lines/l-2", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteLine");

      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");

      cy.get("[data-line-id='l-2'] [aria-label='Supprimer la ligne']").click();
      cy.wait("@deleteLine");
      cy.get("[data-line-id='l-2']").should("not.exist");
    });

    it("shows an error indicator and toast on failed save", () => {
      stubGet([line()]);
      cy.intercept("PUT", "/api/quotes/q-1/lines/l-1", {
        statusCode: 500,
        body: { success: false, message: "Échec serveur." },
      }).as("updateLineFail");

      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");

      cy.get("[data-line-id='l-1'] input[name='line-name']")
        .clear()
        .type("Boom");
      cy.wait("@updateLineFail");

      cy.get(
        "[data-line-id='l-1'] [data-slot='line-save-indicator'][data-status='error']",
      ).should("exist");
      cy.get("[data-sonner-toaster]").should("contain", "Échec serveur.");
    });
  });

  describe("Step 2 — taxes", () => {
    function stubGet(lines: LineFixture[], taxes: TaxFixture[] = []) {
      cy.login();
      stubAvailableTaxes(taxes);
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote(), lines },
      }).as("getQuote");
    }

    it("disables the tax select and hides TTC when no tax is available", () => {
      stubGet([line()]);
      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");
      cy.wait("@listAvailableTaxes");

      cy.get("[data-line-id='l-1'] [data-slot='line-tax-cell'] input")
        .should("be.disabled");
      cy.get("[data-slot='total-ht']").should("contain", "800.00");
      cy.get("[data-slot='total-ttc']").should("not.exist");
      cy.get("[data-slot='total-tax-line']").should("not.exist");
    });

    it("preselects the default tax on a newly added line", () => {
      stubGet([line()], [tax({ id: 100, is_default: true })]);
      cy.intercept("POST", "/api/quotes/q-1/lines", {
        statusCode: 201,
        body: { success: true, line_id: "l-2" },
      }).as("createLine");

      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");
      cy.wait("@listAvailableTaxes");
      cy.get("[aria-label='Ajouter une ligne']").click();

      cy.wait("@createLine").then((interception) => {
        expect(interception.request.body).to.include({ tax_id: 100 });
      });
    });

    it("shows tax breakdown and TTC when a line uses a tax", () => {
      stubGet(
        [line({ tax_id: 100 })],
        [tax({ id: 100, name: "TVA 20", rate: "20.00", is_default: true })],
      );
      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");
      cy.wait("@listAvailableTaxes");

      cy.get("[data-slot='total-ht']").should("contain", "800.00");
      cy.get("[data-slot='total-tax-line'][data-tax-id='100']").should(
        "contain",
        "160.00",
      );
      cy.get("[data-slot='total-ttc']").should("contain", "960.00");
    });

    it("forwards orphan tax_ids to /available so superseded snapshots render", () => {
      // Line references the previous version (id=99, superseded). The
      // current version (id=100) is in the live list; the backend returns
      // both because the frontend forwarded include_ids=99.
      stubGet(
        [line({ tax_id: 99 })],
        [
          tax({ id: 100, name: "TVA 21", rate: "21.00", is_default: true, version: 2 }),
          tax({
            id: 99,
            name: "TVA 20",
            rate: "20.00",
            is_default: false,
            version: 1,
            superseded_at: "2026-04-01T00:00:00Z",
            superseded_by: 100,
          }),
        ],
      );
      cy.visit("/quote/q-1?step=2");
      cy.wait("@getQuote");
      cy.wait("@listAvailableTaxes").then((interception) => {
        expect(interception.request.url).to.include("include_ids=99");
      });

      // Breakdown uses the snapshot (20%, not the current 21%).
      cy.get("[data-slot='total-tax-line'][data-tax-id='99']").should(
        "contain",
        "160.00",
      );
      cy.get("[data-slot='total-ttc']").should("contain", "960.00");
    });
  });

  describe("Drop / Continue", () => {
    beforeEach(() => stubAvailableTaxes());

    it("drops a draft quote with confirmation and switches form to readonly", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: {
          success: true,
          quote: quote({ state: "draft" }),
          lines: [line()],
        },
      }).as("getQuote");
      cy.intercept("POST", "/api/quotes/q-1/drop", {
        statusCode: 200,
        body: { success: true },
      }).as("dropQuote");

      cy.visit("/quote/q-1");
      cy.wait("@getQuote");

      cy.get("[data-quote-state='draft']").should("exist");
      cy.contains("button", "Abandonner").click();
      cy.get("[data-slot='alert-dialog-content']").should("be.visible");
      cy.contains("button", "Confirmer").click();

      cy.wait("@dropQuote");
      cy.get("[data-quote-state='drop']").should("exist");
      cy.get("[data-slot='quote-state-badge']").should("contain", "Abandonné");
      cy.get("input[name='name']").should("be.disabled");
      cy.contains("button", "Continuer").should("be.visible");
      cy.contains("button", "Abandonner").should("not.exist");
    });

    it("cancels the drop confirmation without calling the API", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote({ state: "draft" }), lines: [] },
      }).as("getQuote");
      cy.intercept("POST", "/api/quotes/q-1/drop", cy.spy().as("dropSpy"));

      cy.visit("/quote/q-1");
      cy.wait("@getQuote");

      cy.contains("button", "Abandonner").click();
      cy.contains("button", "Annuler").click();
      cy.get("[data-slot='alert-dialog-content']").should("not.exist");
      cy.get("[data-quote-state='draft']").should("exist");
      cy.get("@dropSpy").should("not.have.been.called");
    });

    it("reactivates a Drop quote via Continuer", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote({ state: "drop" }), lines: [] },
      }).as("getQuote");
      cy.intercept("POST", "/api/quotes/q-1/continue", {
        statusCode: 200,
        body: { success: true },
      }).as("continueQuote");

      cy.visit("/quote/q-1");
      cy.wait("@getQuote");

      cy.get("[data-quote-state='drop']").should("exist");
      cy.get("input[name='name']").should("be.disabled");
      cy.contains("button", "Continuer").click();
      cy.wait("@continueQuote");

      cy.get("[data-quote-state='draft']").should("exist");
      cy.get("input[name='name']").should("not.be.disabled");
    });

    it("shows readonly UI and no Abandonner button for validated quotes", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: {
          success: true,
          quote: quote({ state: "validated" }),
          lines: [],
        },
      }).as("getQuote");

      cy.visit("/quote/q-1");
      cy.wait("@getQuote");

      cy.get("[data-quote-state='validated']").should("exist");
      cy.get("input[name='name']").should("be.disabled");
      cy.contains("button", "Abandonner").should("not.exist");
      cy.contains("button", "Continuer").should("not.exist");
    });
  });
});
