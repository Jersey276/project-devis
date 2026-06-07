const STATS = {
  success: true,
  total_active_subscriptions: 10,
  total_revenue_cents: 95000,
  plan_distribution: [
    { tier: "free", count: 5 },
    { tier: "pro", count: 4 },
    { tier: "enterprise", count: 1 },
  ],
  monthly_revenue: [
    { month: "2026-01-01", revenue_cents: 8100 },
    { month: "2026-02-01", revenue_cents: 9000 },
  ],
};

function stubAnalyticsAdmin() {
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
  cy.intercept("GET", "/api/subscriptions/admin/stats", {
    statusCode: 200,
    body: STATS,
  }).as("getStats");
}

describe("Analytics page", () => {
  it("admin sees the analytics page with title", () => {
    stubAnalyticsAdmin();
    cy.visit("/analytics");
    cy.wait("@getAuthMe");
    cy.wait("@getStats");
    cy.contains("Analytiques").should("be.visible");
  });

  it("shows total active subscriptions metric", () => {
    stubAnalyticsAdmin();
    cy.visit("/analytics");
    cy.wait("@getStats");
    cy.contains("10").should("be.visible");
    cy.contains("Abonnements actifs").should("be.visible");
  });

  it("shows total revenue formatted in euros", () => {
    stubAnalyticsAdmin();
    cy.visit("/analytics");
    cy.wait("@getStats");
    cy.contains("Revenu total").should("be.visible");
    cy.contains("950").should("be.visible");
  });

  it("renders plan distribution chart container", () => {
    stubAnalyticsAdmin();
    cy.visit("/analytics");
    cy.wait("@getStats");
    cy.contains("Répartition des plans").should("be.visible");
    cy.get(".recharts-wrapper").should("exist");
  });

  it("renders monthly revenue line chart", () => {
    stubAnalyticsAdmin();
    cy.visit("/analytics");
    cy.wait("@getStats");
    cy.contains("Revenu mensuel").should("be.visible");
    cy.get(".recharts-wrapper").should("have.length.at.least", 1);
  });

  it("non-admin sees access denied", () => {
    cy.login();
    cy.intercept("GET", "/api/auth/me", {
      statusCode: 200,
      body: {
        success: true,
        auth: {
          user_id: "user-1",
          email: "user@test.fr",
          role: "free_user",
          account_status: "active",
          subscription_tier: "free",
        },
      },
    }).as("getAuthMe");

    cy.visit("/analytics");
    cy.wait("@getAuthMe");
    cy.contains("Accès refusé").should("be.visible");
  });
});
