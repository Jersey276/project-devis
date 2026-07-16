import {
  creditNoteDetails,
  creditNoteSummary,
  invoiceDetails,
  invoiceLine,
} from "../support/fixtures";

describe("Credit note", () => {
  describe("List", () => {
    it("renders credit notes returned by the API", () => {
      cy.login();
      cy.intercept("GET", "/api/credit-notes**", {
        statusCode: 200,
        body: {
          success: true,
          credit_notes: [
            creditNoteSummary({ credit_note_id: "cn-1", credit_note_number: "A-2026-0001", is_total: false }),
            creditNoteSummary({ credit_note_id: "cn-2", credit_note_number: "A-2026-0002", is_total: true }),
          ],
          total: 2,
        },
      }).as("listCreditNotes");

      cy.visit("/credit-note");
      cy.wait("@listCreditNotes");

      cy.contains("td", "A-2026-0001").should("be.visible");
      cy.contains("td", "Partiel").should("be.visible");
      cy.contains("td", "A-2026-0002").should("be.visible");
      cy.contains("td", "Total").should("be.visible");
    });
  });

  describe("Detail", () => {
    it("renders a credit note with negative totals", () => {
      cy.login();
      cy.intercept("GET", "/api/credit-notes/cn-1", {
        statusCode: 200,
        body: { success: true, credit_note: creditNoteDetails({ credit_note_id: "cn-1" }) },
      }).as("getCreditNote");

      cy.visit("/credit-note/cn-1");
      cy.wait("@getCreditNote");

      cy.contains("Avoir N° A-2026-0001").should("be.visible");
      cy.contains("En référence à la facture N° 2026-0001").should("be.visible");
      cy.contains("-120.00 €").should("be.visible");
    });
  });

  describe("Create from invoice", () => {
    function stubIssuedInvoice(over: Partial<Parameters<typeof invoiceDetails>[0]> = {}) {
      cy.intercept("GET", "/api/invoices/inv-1", {
        statusCode: 200,
        body: {
          success: true,
          invoice: invoiceDetails({
            invoice_id: "inv-1",
            status: "ISSUED",
            lines: [
              invoiceLine({ quote_line_id: "l-1", name: "Prestation A", line_ht_cents: 20000 }),
              invoiceLine({ quote_line_id: "l-2", name: "Prestation B", line_ht_cents: 10000 }),
            ],
            total_ht_cents: 30000,
            total_ttc_cents: 36000,
            ...over,
          }),
        },
      }).as("getInvoice");
      cy.intercept("GET", "/api/invoices/inv-1/lifecycle-events", {
        statusCode: 200,
        body: { success: true, events: [] },
      });
      cy.intercept("GET", "/api/credit-notes**invoice_id=inv-1**", {
        statusCode: 200,
        body: { success: true, credit_notes: [], total: 0 },
      });
    }

    it("creates a partial credit note for one selected line", () => {
      cy.login();
      stubIssuedInvoice();
      cy.intercept("POST", "/api/invoices/inv-1/credit-notes", (req) => {
        expect(req.body).to.deep.equal({ positions: [0], reason: "" });
        req.reply({ statusCode: 200, body: { success: true, credit_note_id: "cn-new" } });
      }).as("createCreditNote");
      cy.intercept("GET", "/api/credit-notes/cn-new", {
        statusCode: 200,
        body: { success: true, credit_note: creditNoteDetails({ credit_note_id: "cn-new", is_total: false }) },
      });

      cy.visit("/invoice/inv-1");
      cy.wait("@getInvoice");
      cy.contains("button", "Créer un avoir").click();
      cy.get("#cn-line-0").click({ force: true });
      cy.contains('[data-slot="dialog-content"] button', "Créer l'avoir").click();
      cy.wait("@createCreditNote");
      cy.url().should("include", "/credit-note/cn-new");
    });

    it("creates a total credit note via 'select all remaining'", () => {
      cy.login();
      stubIssuedInvoice();
      cy.intercept("POST", "/api/invoices/inv-1/credit-notes", (req) => {
        expect(req.body).to.deep.equal({ positions: [0, 1], reason: "Erreur de facturation" });
        req.reply({ statusCode: 200, body: { success: true, credit_note_id: "cn-new" } });
      }).as("createCreditNote");
      cy.intercept("GET", "/api/credit-notes/cn-new", {
        statusCode: 200,
        body: { success: true, credit_note: creditNoteDetails({ credit_note_id: "cn-new", is_total: true }) },
      });

      cy.visit("/invoice/inv-1");
      cy.wait("@getInvoice");
      cy.contains("button", "Créer un avoir").click();
      cy.contains("button", "Tout sélectionner").click();
      cy.get("#cn-reason").type("Erreur de facturation");
      cy.contains('[data-slot="dialog-content"] button', "Créer l'avoir").click();
      cy.wait("@createCreditNote");
    });

    it("disables an already-credited line", () => {
      cy.login();
      stubIssuedInvoice({ credited_positions: [0] });

      cy.visit("/invoice/inv-1");
      cy.wait("@getInvoice");
      cy.contains("button", "Créer un avoir").click();
      cy.get("#cn-line-0").should("be.disabled");
      cy.contains("(déjà avoirée)").should("be.visible");
    });
  });
});
