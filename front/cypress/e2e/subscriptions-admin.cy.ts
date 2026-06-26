describe("Admin Subscriptions page", () => {
  const PLANS = [
    {
      plan_id: 1,
      name: "Free",
      tier: "free",
      price_cents: 0,
      billing_cycle: "none",
      features: { max_schedules: 3, max_templates: 2, max_projects: 0, max_linked_clients: 0, fees_catalog: 0, b2b_invoicing: 0 },
    },
    {
      plan_id: 2,
      name: "Pro",
      tier: "pro",
      price_cents: 900,
      billing_cycle: "monthly",
      features: { max_schedules: -1, max_templates: -1, max_projects: 10, max_linked_clients: 5, fees_catalog: 1, b2b_invoicing: 1 },
    },
    {
      plan_id: 3,
      name: "Enterprise",
      tier: "enterprise",
      price_cents: 4900,
      billing_cycle: "monthly",
      features: { max_schedules: -1, max_templates: -1, max_projects: -1, max_linked_clients: -1, fees_catalog: 1, b2b_invoicing: 1 },
    },
  ];

  const SUBSCRIPTIONS = [
    {
      subscription_id: "sub-1",
      user_id: "user-abc-123",
      plan_id: 2,
      tier: "pro",
      status: "active",
      current_period_start: "2026-01-01T00:00:00Z",
      current_period_end: null,
      updated_at: "2026-05-01T00:00:00Z",
    },
    {
      subscription_id: "sub-2",
      user_id: "user-def-456",
      plan_id: 1,
      tier: "free",
      status: "active",
      current_period_start: "2026-01-01T00:00:00Z",
      current_period_end: null,
      updated_at: "2026-04-01T00:00:00Z",
    },
  ];

  function stubAdminSubscriptionsPage() {
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

    cy.intercept("GET", "/api/subscriptions/admin", {
      statusCode: 200,
      body: { success: true, subscriptions: SUBSCRIPTIONS, total: 2 },
    }).as("getSubscriptions");

    cy.intercept("GET", "/api/plans", {
      statusCode: 200,
      body: { success: true, plans: PLANS },
    }).as("getPlans");
    cy.intercept("GET", "/api/plans?include_inactive=true", {
      statusCode: 200,
      body: { success: true, plans: PLANS },
    }).as("getAllPlans");
  }

  it("renders the subscriptions table with user rows", () => {
    stubAdminSubscriptionsPage();

    cy.visit("/subscriptions");
    cy.wait("@getAuthMe");
    cy.wait("@getSubscriptions");

    cy.contains("td", "user-abc-123").should("be.visible");
    cy.contains("td", "user-def-456").should("be.visible");
    cy.contains("Pro").should("be.visible");
    cy.contains("Gratuit").should("be.visible");
  });

  it("shows empty state when no subscriptions", () => {
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
    cy.intercept("GET", "/api/subscriptions/admin", {
      statusCode: 200,
      body: { success: true, subscriptions: [], total: 0 },
    }).as("getSubscriptionsEmpty");
    cy.intercept("GET", "/api/plans", {
      statusCode: 200,
      body: { success: true, plans: PLANS },
    }).as("getPlans");
    cy.intercept("GET", "/api/plans?include_inactive=true", {
      statusCode: 200,
      body: { success: true, plans: PLANS },
    });

    cy.visit("/subscriptions");
    cy.wait("@getSubscriptionsEmpty");

    cy.contains("Aucun abonnement pour le moment.").should("be.visible");
  });

  it("opens change plan dialog and assigns a plan", () => {
    stubAdminSubscriptionsPage();

    let assigned = false;
    cy.intercept("POST", "/api/subscriptions/admin/*/plan", (req) => {
      assigned = true;
      req.reply({ statusCode: 200, body: { success: true } });
    }).as("assignPlan");

    cy.intercept("GET", "/api/subscriptions/admin", (req) => {
      req.reply({
        statusCode: 200,
        body: {
          success: true,
          subscriptions: assigned
            ? [
                { ...SUBSCRIPTIONS[0], tier: "enterprise", plan_id: 3 },
                SUBSCRIPTIONS[1],
              ]
            : SUBSCRIPTIONS,
          total: 2,
        },
      });
    }).as("getSubscriptionsReload");

    cy.visit("/subscriptions");
    cy.wait("@getAuthMe");
    cy.wait("@getSubscriptionsReload");
    cy.wait("@getPlans");

    cy.contains("td", "user-abc-123")
      .closest("tr")
      .within(() => {
        cy.get("button.h-8.w-8").click({ force: true });
      });

    cy.contains("Changer de plan").click({ force: true });
    cy.contains("Modifier le plan").should("be.visible");

    cy.get("#assign_plan_select").click({ force: true });
    cy.contains("[data-slot='select-item']", "Enterprise").click({
      force: true,
    });

    cy.contains("button", "Confirmer").click();

    cy.wait("@assignPlan").then(({ request }) => {
      expect(request.body).to.deep.equal({ plan_id: 3 });
    });
    cy.wait("@getSubscriptionsReload");
    cy.contains("Enterprise").should("be.visible");
  });

  it("opens edit plan dialog and shows new feature fields", () => {
    stubAdminSubscriptionsPage();

    cy.visit("/subscriptions");
    cy.wait("@getAuthMe");
    cy.wait("@getAllPlans");

    // Open actions menu on the Pro plan row
    cy.contains("td", "Pro")
      .closest("tr")
      .within(() => {
        cy.get("button.h-8.w-8").click({ force: true });
      });

    cy.contains("Modifier le plan").click({ force: true });
    cy.contains("Projets max").scrollIntoView().should("be.visible");
    cy.contains("Clients liés max").scrollIntoView().should("be.visible");
    cy.contains("Catalogue de frais").scrollIntoView().should("be.visible");
    cy.contains("Facturation B2B").scrollIntoView().should("be.visible");
  });

  it("blocks access for non-admin users", () => {
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

    cy.visit("/subscriptions");
    cy.wait("@getAuthMe");

    cy.contains("Accès refusé").should("be.visible");
  });
});
