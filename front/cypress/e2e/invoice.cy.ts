import {
  invoiceDetails,
  invoiceSummary,
  quote,
  type InvoiceDetailsFixture,
} from "../support/fixtures";

function stubOssAndReporting() {
  cy.intercept("GET", "/api/invoices/oss-status", {
    statusCode: 200,
    body: {
      year: 2026,
      cumulative_ht_cents: 0,
      threshold_cents: 1000000,
      oss_enabled: false,
      oss_active: false,
      prior_year_over_threshold: false,
      prior_year_cumulative_ht_cents: 0,
    },
  });
  cy.intercept("GET", "/api/invoices/reports**", {
    statusCode: 200,
    body: { success: true, reports: [] },
  });
}

describe("Invoice", () => {
  describe("List", () => {
    it("renders invoices returned by the API", () => {
      cy.login();
      stubOssAndReporting();
      cy.intercept("GET", "/api/invoices**", {
        statusCode: 200,
        body: {
          success: true,
          invoices: [
            invoiceSummary({ invoice_id: "inv-1", invoice_number: "2026-0001", status: "ISSUED" }),
            invoiceSummary({ invoice_id: "inv-2", invoice_number: "", status: "DRAFT", total_ttc_cents: 5000, total_ht_cents: 5000 }),
          ],
          total: 2,
        },
      }).as("listInvoices");

      cy.visit("/invoice");
      cy.wait("@listInvoices");

      cy.contains("td", "2026-0001").should("be.visible");
      cy.contains("td", "Émise").should("be.visible");
      cy.contains("td", "Brouillon").should("be.visible");
    });

    it("filters by status", () => {
      cy.login();
      stubOssAndReporting();
      cy.intercept("GET", "/api/invoices**", {
        statusCode: 200,
        body: { success: true, invoices: [], total: 0 },
      }).as("listInvoicesDefault");

      cy.visit("/invoice");
      cy.wait("@listInvoicesDefault");

      cy.intercept("GET", "/api/invoices**statuses=ISSUED**", (req) => {
        expect(req.url).to.include("statuses=ISSUED");
        req.reply({ statusCode: 200, body: { success: true, invoices: [], total: 0 } });
      }).as("listInvoicesFiltered");

      cy.contains("button", "Filtres").click();
      cy.get('input[placeholder="Filtrer par statut"]').click();
      cy.contains("[data-slot='combobox-item']", "Émise").click({ force: true });
      cy.wait("@listInvoicesFiltered");
    });
  });

  describe("Detail", () => {
    // invoice-detail.tsx unconditionally renders LinkedCreditNotes and (for
    // ISSUED/PAID invoices) the lifecycle timeline, so every visit to
    // /invoice/:id fires both — stub both here so a missed one doesn't fall
    // through to a real 401 and redirect to /login mid-test.
    function stubInvoiceDetail(invoiceId: string, over: Partial<InvoiceDetailsFixture> = {}) {
      cy.intercept("GET", `/api/invoices/${invoiceId}`, {
        statusCode: 200,
        body: { success: true, invoice: invoiceDetails({ invoice_id: invoiceId, ...over }) },
      }).as("getInvoice");
      cy.intercept("GET", `/api/invoices/${invoiceId}/lifecycle-events`, {
        statusCode: 200,
        body: { success: true, events: [] },
      });
      cy.intercept("GET", `/api/credit-notes**invoice_id=${invoiceId}**`, {
        statusCode: 200,
        body: { success: true, credit_notes: [], total: 0 },
      });
    }

    it("renders an ISSUED invoice with totals and lines", () => {
      cy.login();
      stubInvoiceDetail("inv-1");

      cy.visit("/invoice/inv-1");
      cy.wait("@getInvoice");

      cy.contains("Facture N° 2026-0001").should("be.visible");
      cy.contains("Acme SARL").should("be.visible");
      cy.contains("Jean Dupont").should("be.visible");
      cy.contains("Design UI").should("be.visible");
      cy.contains("120.00 €").should("be.visible");
    });

    it("marks an ISSUED invoice as paid", () => {
      cy.login();
      stubInvoiceDetail("inv-1", { status: "ISSUED" });
      cy.intercept("POST", "/api/invoices/inv-1/paid", {
        statusCode: 200,
        body: { success: true },
      }).as("markPaid");

      cy.visit("/invoice/inv-1");
      cy.wait("@getInvoice");
      cy.contains("button", "Marquer payée").click();
      cy.wait("@markPaid");
    });

    it("deletes a DRAFT invoice after confirmation", () => {
      cy.login();
      stubInvoiceDetail("inv-2", { status: "DRAFT", invoice_number: "" });
      cy.intercept("DELETE", "/api/invoices/inv-2", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteDraft");

      cy.visit("/invoice/inv-2");
      cy.wait("@getInvoice");
      cy.contains("button", "Supprimer le brouillon").click();
      cy.get('[data-slot="alert-dialog-title"]').should(
        "contain.text",
        "Supprimer ce brouillon ?",
      );
      cy.get('[data-slot="alert-dialog-action"]').click();
      cy.wait("@deleteDraft");
    });

    it("shows a toast when the PDF export fails", () => {
      cy.login();
      stubInvoiceDetail("inv-1");
      cy.intercept("GET", "/api/export/invoices/inv-1", {
        statusCode: 500,
        body: { success: false },
      }).as("exportPdf");

      cy.visit("/invoice/inv-1");
      cy.wait("@getInvoice");
      cy.contains("button", "Télécharger le PDF").click();
      cy.wait("@exportPdf");
      cy.contains("Export PDF impossible.").should("be.visible");
    });
  });

  describe("Customer mode hides provider-only actions", () => {
    it("hides Mark Paid, Delete Draft and Create Credit Note", () => {
      cy.login();
      cy.intercept("GET", "/api/invoices/inv-1", {
        statusCode: 200,
        body: { success: true, invoice: invoiceDetails({ invoice_id: "inv-1", status: "ISSUED" }) },
      }).as("getInvoice");
      // LifecycleTimeline is gated on !isCustomer, so no lifecycle-events call
      // is expected here — only LinkedCreditNotes, which always renders.
      cy.intercept("GET", "/api/credit-notes**invoice_id=inv-1**", {
        statusCode: 200,
        body: { success: true, credit_notes: [], total: 0 },
      });

      cy.visitAs("customer", "/invoice/inv-1");
      cy.wait("@getInvoice");

      cy.contains("button", "Marquer payée").should("not.exist");
      cy.contains("button", "Supprimer le brouillon").should("not.exist");
      cy.contains("button", "Créer un avoir").should("not.exist");
      cy.contains("button", "Déposer sur la plateforme").should("not.exist");
      cy.contains("button", "Faire évoluer le statut").should("not.exist");
    });
  });

  describe("Generate from quote", () => {
    it("shows the button for a validated quote with no schedule, and generates on click", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote({ quote_id: "q-1", state: "validated" }) },
      }).as("getQuote");
      cy.intercept("GET", "/api/quotes/q-1/lines**", {
        statusCode: 200,
        body: { success: true, lines: [] },
      });
      cy.intercept("GET", "/api/schedules?quote_id=q-1", {
        statusCode: 200,
        body: { success: true, schedules: [] },
      }).as("listSchedules");
      cy.intercept("POST", "/api/invoices/from-quote", {
        statusCode: 200,
        body: { success: true, invoice_id: "inv-new" },
      }).as("createFromQuote");
      cy.intercept("GET", "/api/invoices/inv-new", {
        statusCode: 200,
        body: { success: true, invoice: invoiceDetails({ invoice_id: "inv-new" }) },
      }).as("getNewInvoice");
      cy.intercept("GET", "/api/invoices/inv-new/lifecycle-events", {
        statusCode: 200,
        body: { success: true, events: [] },
      });
      cy.intercept("GET", "/api/credit-notes**invoice_id=inv-new**", {
        statusCode: 200,
        body: { success: true, credit_notes: [], total: 0 },
      });

      cy.visit("/quote/q-1");
      cy.wait(["@getQuote", "@listSchedules"]);

      cy.contains("button", "Générer une facture").click();
      cy.wait("@createFromQuote");
      cy.url().should("include", "/invoice/inv-new");
    });

    it("hides the button when the quote already has a schedule", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote({ quote_id: "q-1", state: "validated" }) },
      }).as("getQuote");
      cy.intercept("GET", "/api/quotes/q-1/lines**", {
        statusCode: 200,
        body: { success: true, lines: [] },
      });
      cy.intercept("GET", "/api/schedules?quote_id=q-1", {
        statusCode: 200,
        body: { success: true, schedules: [{ schedule_id: "sched-1" }] },
      }).as("listSchedules");

      cy.visit("/quote/q-1");
      cy.wait(["@getQuote", "@listSchedules"]);

      cy.contains("button", "Générer une facture").should("not.exist");
    });

    it("shows an error toast when generation fails", () => {
      cy.login();
      cy.intercept("GET", "/api/quotes/q-1", {
        statusCode: 200,
        body: { success: true, quote: quote({ quote_id: "q-1", state: "validated" }) },
      }).as("getQuote");
      cy.intercept("GET", "/api/quotes/q-1/lines**", {
        statusCode: 200,
        body: { success: true, lines: [] },
      });
      cy.intercept("GET", "/api/schedules?quote_id=q-1", {
        statusCode: 200,
        body: { success: true, schedules: [] },
      }).as("listSchedules");
      cy.intercept("POST", "/api/invoices/from-quote", {
        statusCode: 500,
        body: { success: false },
      }).as("createFromQuote");

      cy.visit("/quote/q-1");
      cy.wait(["@getQuote", "@listSchedules"]);
      cy.contains("button", "Générer une facture").click();
      cy.wait("@createFromQuote");
      cy.get("[data-sonner-toaster]").should("contain", "La génération de la facture a échoué.");
    });
  });

  describe("Generate from schedule", () => {
    it("bills selected months via the dialog", () => {
      cy.login();
      cy.intercept("GET", "/api/schedules/sched-1", {
        statusCode: 200,
        body: {
          success: true,
          schedule: {
            schedule_id: "sched-1",
            quote_id: "q-1",
            status: "VALID",
            name: "Échéancier",
            start_month: "2026-01",
            duration_months: 3,
            lines: [],
            column_totals: [
              { month_index: 1, amount_cents: 3000 },
              { month_index: 2, amount_cents: 3500 },
              { month_index: 3, amount_cents: 3500 },
            ],
            quote_total_cents: 10000,
            planned_total_cents: 10000,
          },
        },
      }).as("getSchedule");
      cy.intercept("POST", "/api/invoices/from-schedule", (req) => {
        expect(req.body).to.deep.equal({
          schedule_id: "sched-1",
          month_indexes: [1, 2],
          sale_date: "",
          due_in_days: 0,
          issue_now: true,
        });
        req.reply({ statusCode: 200, body: { success: true, invoice_id: "inv-new" } });
      }).as("createFromSchedule");

      cy.visit("/schedule/sched-1");
      cy.wait("@getSchedule");

      cy.contains("button", "Générer une facture").click();
      cy.get("#inv-month-1").click({ force: true });
      cy.get("#inv-month-2").click({ force: true });
      cy.contains('[data-slot="dialog-content"] button', "Générer").click();
      cy.wait("@createFromSchedule");
    });
  });
});
