const PLANS = [
  {
    plan_id: 1,
    name: "Free",
    tier: "free",
    price_cents: 0,
    billing_cycle: "none",
    features: {},
  },
  {
    plan_id: 2,
    name: "Pro",
    tier: "pro",
    price_cents: 900,
    billing_cycle: "monthly",
    features: {},
  },
  {
    plan_id: 3,
    name: "Enterprise",
    tier: "enterprise",
    price_cents: 4900,
    billing_cycle: "monthly",
    features: {},
  },
];

const FREE_SUB = {
  subscription_id: "",
  user_id: "u1",
  plan_id: 1,
  tier: "free",
  status: "active",
  current_period_start: "",
  current_period_end: null as string | null,
  cancel_at_period_end: false,
  stripe_subscription_id: null as string | null,
  updated_at: "",
};

const PRO_SUB = {
  subscription_id: "sub-1",
  user_id: "u1",
  plan_id: 2,
  tier: "pro",
  status: "active",
  current_period_start: "2026-01-01T00:00:00Z",
  current_period_end: "2026-07-01T00:00:00Z",
  cancel_at_period_end: false,
  stripe_subscription_id: "sub_stripe_123",
  updated_at: "2026-01-01T00:00:00Z",
};

function stubSubscriptionTab(subscription = FREE_SUB) {
  cy.login();
  cy.intercept("GET", "/api/users/me", {
    statusCode: 200,
    body: { success: true, user: { user_id: "u1", email: "test@test.fr" } },
  });
  cy.intercept("GET", "/api/subscriptions/me", {
    statusCode: 200,
    body: { success: true, subscription },
  }).as("getMySub");
  cy.intercept("GET", "/api/plans", {
    statusCode: 200,
    body: { success: true, plans: PLANS },
  }).as("getPlans");
  cy.intercept("GET", "/api/users/addresses?**", {
    statusCode: 200,
    body: { success: true, addresses: [] },
  });
  cy.intercept("GET", "/api/users/countries", {
    statusCode: 200,
    body: { success: true, countries: [] },
  });
}

describe("Profile — Subscription tab", () => {
  it("shows the Abonnement tab on the profile page", () => {
    stubSubscriptionTab();
    cy.visit("/profile");
    cy.contains("[role='tab']", "Abonnement").should("be.visible").click();
    cy.wait("@getMySub");
    cy.wait("@getPlans");
    cy.contains("Mon abonnement actuel").should("be.visible");
  });

  it("free user sees plan cards with S'abonner button for paid plans", () => {
    stubSubscriptionTab(FREE_SUB);
    cy.visit("/profile?tab=abonnement");
    cy.wait("@getMySub");
    cy.wait("@getPlans");
    cy.contains("Gratuit — pas de renouvellement").should("be.visible");
    cy.contains("button", "S'abonner").should("be.visible");
  });

  it("pro user sees renewal info and cancel button", () => {
    stubSubscriptionTab(PRO_SUB);
    cy.visit("/profile?tab=abonnement");
    cy.wait("@getMySub");
    cy.wait("@getPlans");
    cy.contains("Renouvellement le").should("be.visible");
    cy.contains("button", "Résilier l'abonnement").should("be.visible");
  });

  it("cancel button calls API and shows success toast", () => {
    stubSubscriptionTab(PRO_SUB);
    cy.intercept("POST", "/api/subscriptions/cancel", {
      statusCode: 200,
      body: { success: true },
    }).as("cancelSub");

    cy.visit("/profile?tab=abonnement");
    cy.wait("@getMySub");
    cy.wait("@getPlans");

    cy.contains("button", "Résilier l'abonnement").click();
    cy.wait("@cancelSub");
    cy.contains("Résiliation programmée.").should("be.visible");
  });

  it("S'abonner button calls payment-intent and opens dialog", () => {
    stubSubscriptionTab(FREE_SUB);
    cy.intercept("POST", "/api/subscriptions/payment-intent", {
      statusCode: 200,
      body: {
        success: true,
        client_secret: "pi_test_abc123_secret_def456xyz",
        stripe_subscription_id: "sub_test_123",
      },
    }).as("createIntent");

    cy.visit("/profile?tab=abonnement");
    cy.wait("@getMySub");
    cy.wait("@getPlans");

    cy.contains("button", "S'abonner").first().click();
    cy.wait("@createIntent");
    cy.contains("Passer au plan").should("be.visible");
  });
});
