describe("Quote", () => {
  type QuoteFixture = {
    quote_id: string;
    user_id: string;
    name: string;
    archived_at: string | null;
    created_at: string;
    updated_at: string;
  };

  type LineFixture = {
    line_id: string;
    quote_id: string;
    type: "simple";
    name: string;
    quantity: string;
    unit: string;
    unit_price: number;
    data: Record<string, unknown>;
    position: number;
  };

  function quote(over: Partial<QuoteFixture> = {}): QuoteFixture {
    return {
      quote_id: "q-1",
      user_id: "u-1",
      name: "Devis Alpha",
      archived_at: null,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
      ...over,
    };
  }

  function line(over: Partial<LineFixture> = {}): LineFixture {
    return {
      line_id: "l-1",
      quote_id: "q-1",
      type: "simple",
      name: "Design UI",
      quantity: "10",
      unit: "",
      unit_price: 8000,
      data: {},
      position: 0,
      ...over,
    };
  }

  describe("List", () => {
    it("renders quotes and maps archived_at to status", () => {
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
          ],
        },
      }).as("listQuotes");

      cy.visit("/quote");
      cy.wait("@listQuotes");

      cy.contains("td", "Devis Alpha").should("be.visible");
      cy.contains("td", "brouillon").should("be.visible");
      cy.contains("td", "Devis Beta").should("be.visible");
      cy.contains("td", "archivé").should("be.visible");
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
    it("creates a quote and redirects to step 2", () => {
      cy.login();
      cy.intercept("POST", "/api/quotes", {
        statusCode: 201,
        body: { success: true, quote_id: "q-new" },
      }).as("createQuote");
      cy.intercept("GET", "/api/quotes/q-new", {
        statusCode: 200,
        body: {
          success: true,
          quote: quote({ quote_id: "q-new", name: "Nouveau devis" }),
          lines: [],
        },
      }).as("getNewQuote");

      cy.visit("/quote/create");
      cy.get("input[name='name']").type("Nouveau devis");
      cy.contains("button", "Suivant").click();

      cy.wait("@createQuote").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          name: "Nouveau devis",
        });
      });

      cy.url().should("include", "/quote/q-new");
      cy.url().should("include", "step=2");
      cy.wait("@getNewQuote");
      cy.get("[data-step-tab='1'][data-active='true']").should("exist");
    });

    it("surfaces field errors on 422", () => {
      cy.login();
      cy.intercept("POST", "/api/quotes", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "name", error_code: [1] }],
        },
      }).as("createQuoteInvalid");

      cy.visit("/quote/create");
      cy.get("input[name='name']").type("X");
      cy.contains("button", "Suivant").click();
      cy.wait("@createQuoteInvalid");

      cy.get("input[name='name']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Ce champ est requis.");
      cy.url().should("include", "/quote/create");
    });
  });

  describe("Step 2 — lines", () => {
    function stubGet(lines: LineFixture[]) {
      cy.login();
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
      stubGet([line({ line_id: "l-1" }), line({ line_id: "l-2", position: 1 })]);
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
});
