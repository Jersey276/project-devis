describe("Templates", () => {
  const TEMPLATE_LINE = {
    template_id: "tpl-1",
    user_id: "u-1",
    template_type: "quote_line",
    target_resource: "quote",
    name: "Template Prestation",
    archived_at: null,
    payload_version: 1,
    payload: {},
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };

  const TEMPLATE_ARCHIVED = {
    ...TEMPLATE_LINE,
    template_id: "tpl-arch",
    name: "Template archivé",
    archived_at: "2026-05-01T00:00:00Z",
  };

  function stubTemplates(opts?: {
    lineTemplates?: typeof TEMPLATE_LINE[];
    quoteTemplates?: typeof TEMPLATE_LINE[];
  }) {
    cy.login();
    cy.intercept("GET", "/api/templates?type=quote_line", {
      statusCode: 200,
      body: { success: true, templates: opts?.lineTemplates ?? [TEMPLATE_LINE] },
    }).as("getLineTemplates");
    cy.intercept("GET", "/api/templates?type=quote_document", {
      statusCode: 200,
      body: { success: true, templates: opts?.quoteTemplates ?? [] },
    }).as("getDocTemplates");
  }

  describe("Archive filter", () => {
    it("excludes archived templates by default", () => {
      cy.login();
      cy.intercept("GET", "/api/templates**", (req) => {
        expect(req.url).to.not.include("archived=true");
        req.reply({ statusCode: 200, body: { success: true, templates: [] } });
      }).as("listDefault");

      cy.visit("/templates");
      cy.wait("@listDefault");
    });

    it("shows archived templates with a badge when the toggle is checked", () => {
      stubTemplates({ lineTemplates: [] });

      cy.intercept("GET", "/api/templates?type=quote_line&archived=true", {
        statusCode: 200,
        body: { success: true, templates: [TEMPLATE_ARCHIVED] },
      }).as("getLineArchived");
      cy.intercept("GET", "/api/templates?type=quote_document&archived=true", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getDocArchived");

      cy.visit("/templates");
      cy.wait("@getLineTemplates");

      cy.get("label[for='template-archived']").click();
      cy.wait("@getLineArchived");

      cy.contains("td", "Template archivé").should("be.visible");
      cy.contains("Archivé").should("be.visible");
    });
  });

  describe("Archive action", () => {
    it("archives a template via the row action menu", () => {
      stubTemplates();

      cy.intercept("POST", "/api/templates/tpl-1/archive", {
        statusCode: 200,
        body: { success: true },
      }).as("archiveTemplate");

      cy.visit("/templates");
      cy.wait("@getLineTemplates");

      cy.contains("td", "Template Prestation")
        .closest("tr")
        .within(() => {
          cy.get("button[aria-label]").click();
        });
      cy.contains("[role='menuitem']", "Archiver").click();

      cy.wait("@archiveTemplate");
      cy.get("[data-sonner-toaster]").should("contain", "Template archivé.");
      cy.contains("td", "Template Prestation").should("not.exist");
    });
  });

  describe("Restore action", () => {
    it("restores an archived template via the row action menu", () => {
      cy.login();
      cy.intercept("GET", "/api/templates?type=quote_line", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getLineTemplates");
      cy.intercept("GET", "/api/templates?type=quote_document", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getDocTemplates");

      cy.intercept("GET", "/api/templates?type=quote_line&archived=true", {
        statusCode: 200,
        body: { success: true, templates: [TEMPLATE_ARCHIVED] },
      }).as("getLineArchived");
      cy.intercept("GET", "/api/templates?type=quote_document&archived=true", {
        statusCode: 200,
        body: { success: true, templates: [] },
      }).as("getDocArchived");

      cy.intercept("POST", "/api/templates/tpl-arch/restore", {
        statusCode: 200,
        body: { success: true },
      }).as("restoreTemplate");

      cy.visit("/templates");
      cy.wait("@getLineTemplates");

      // Enable archived filter
      cy.get("label[for='template-archived']").click();
      cy.wait("@getLineArchived");

      cy.contains("td", "Template archivé")
        .closest("tr")
        .within(() => {
          cy.get("button[aria-label]").click();
        });
      cy.contains("[role='menuitem']", "Restaurer").click();

      cy.wait("@restoreTemplate");
      cy.get("[data-sonner-toaster]").should("contain", "Template restauré.");
      cy.contains("td", "Template archivé").should("not.exist");
    });
  });
});
