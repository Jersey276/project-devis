const LOGS_PAGE_1 = {
  success: true,
  total: 3,
  logs: [
    { id: 1, user_id: "user-1", method: "GET", url: "/api/quotes", duration_ms: 45, resp_status: 200, created_at: "2026-06-22T10:00:00Z" },
    { id: 2, user_id: "user-2", method: "POST", url: "/api/quotes", duration_ms: 120, resp_status: 201, created_at: "2026-06-22T10:01:00Z" },
    { id: 3, user_id: "user-1", method: "DELETE", url: "/api/quotes/5", duration_ms: 30, resp_status: 204, created_at: "2026-06-22T10:02:00Z" },
  ],
};

const LOGS_FILTERED_204 = {
  success: true,
  total: 1,
  logs: [
    { id: 3, user_id: "user-1", method: "DELETE", url: "/api/quotes/5", duration_ms: 30, resp_status: 204, created_at: "2026-06-22T10:02:00Z" },
  ],
};

const LOG_DETAIL = {
  success: true,
  log: {
    id: 1,
    user_id: "user-1",
    method: "GET",
    url: "/api/quotes",
    duration_ms: 45,
    req_body: "",
    resp_body: '{"success":true,"quotes":[]}',
    resp_status: 200,
    created_at: "2026-06-22T10:00:00Z",
  },
};

const LOGS_STATS = {
  success: true,
  stats: [
    { date: "2026-06-01", resp_status: 200, count: 50 },
    { date: "2026-06-01", resp_status: 204, count: 10 },
  ],
};

function stubSuperAdmin() {
  cy.login();
  cy.intercept("GET", "/api/auth/me", {
    statusCode: 200,
    body: {
      success: true,
      auth: {
        user_id: "admin-1",
        email: "admin@test.fr",
        role: "super_admin",
        account_status: "active",
        subscription_tier: "free",
      },
    },
  }).as("getAuthMe");
  cy.intercept("GET", "/api/logs/stats", { statusCode: 200, body: LOGS_STATS }).as("getStats");
}

describe("Logs — page principale", () => {
  it("affiche le tableau avec les logs", () => {
    stubSuperAdmin();
    cy.intercept("GET", "/api/logs?*", { statusCode: 200, body: LOGS_PAGE_1 }).as("getLogs");

    cy.visit("/logs");
    cy.wait("@getLogs");

    cy.contains("Journal d'activité").should("be.visible");
    cy.contains("/api/quotes").should("be.visible");
    cy.contains("200").should("be.visible");
    cy.contains("201").should("be.visible");
    cy.contains("204").should("be.visible");
  });

  it("non-super-admin voit accès refusé", () => {
    cy.login();
    cy.intercept("GET", "/api/auth/me", {
      statusCode: 200,
      body: {
        success: true,
        auth: {
          user_id: "user-1",
          email: "user@test.fr",
          role: "admin",
          account_status: "active",
          subscription_tier: "pro",
        },
      },
    }).as("getAuthMe");

    cy.visit("/logs");
    cy.wait("@getAuthMe");
    cy.contains("Accès refusé").should("be.visible");
  });
});

describe("Logs — filtre statut HTTP", () => {
  beforeEach(() => {
    stubSuperAdmin();
    cy.intercept("GET", "/api/logs?*", { statusCode: 200, body: LOGS_PAGE_1 }).as("getLogs");
    cy.visit("/logs");
    cy.wait("@getLogs");
  });

  it("envoie resp_statuses dans la requête quand un statut est sélectionné", () => {
    cy.intercept("GET", "/api/logs?*resp_statuses=204*", { statusCode: 200, body: LOGS_FILTERED_204 }).as("getLogsFiltered");

    cy.contains("Filtres").click();
    cy.get('[placeholder="Sélectionner des statuts…"]').click();
    cy.get('[data-slot="combobox-item"]').contains("204").click();
    cy.get("body").type("{esc}");

    cy.wait("@getLogsFiltered");
  });

  it("n'affiche que les lignes correspondant au statut filtré", () => {
    cy.intercept("GET", "/api/logs?*resp_statuses=204*", { statusCode: 200, body: LOGS_FILTERED_204 }).as("getLogsFiltered");

    cy.contains("Filtres").click();
    cy.get('[placeholder="Sélectionner des statuts…"]').click();
    cy.get('[data-slot="combobox-item"]').contains("204").click();
    cy.get("body").type("{esc}");

    cy.wait("@getLogsFiltered");

    cy.contains("204").should("be.visible");
    cy.contains("200").should("not.exist");
    cy.contains("201").should("not.exist");
  });

  it("remet à zéro le filtre statut après reset", () => {
    cy.intercept("GET", "/api/logs?*resp_statuses=204*", { statusCode: 200, body: LOGS_FILTERED_204 }).as("getLogsFiltered");
    cy.intercept("GET", "/api/logs?page=1&page_size=50", { statusCode: 200, body: LOGS_PAGE_1 }).as("getLogsReset");

    cy.contains("Filtres").click();
    cy.get('[placeholder="Sélectionner des statuts…"]').click();
    cy.get('[data-slot="combobox-item"]').contains("204").click();
    cy.get("body").type("{esc}");
    cy.wait("@getLogsFiltered");

    cy.contains("Réinitialiser").click();
    cy.wait("@getLogsReset");
    cy.get("body").should("not.have.attr", "data-scroll-locked");

    cy.contains("200").should("be.visible");
    cy.contains("201").should("be.visible");
    cy.contains("204").should("be.visible");
  });
});

describe("Logs — détail d'une ligne", () => {
  it("affiche le corps de réponse au clic sur une ligne", () => {
    stubSuperAdmin();
    cy.intercept("GET", "/api/logs?*", { statusCode: 200, body: LOGS_PAGE_1 }).as("getLogs");
    cy.intercept("GET", "/api/logs/1", { statusCode: 200, body: LOG_DETAIL }).as("getLogDetail");

    cy.visit("/logs");
    cy.wait("@getLogs");

    cy.contains("/api/quotes").click();
    cy.wait("@getLogDetail");

    cy.contains("Corps de la réponse").should("be.visible");
    cy.contains("quotes").should("be.visible");
  });
});
