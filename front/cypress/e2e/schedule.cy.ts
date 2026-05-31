describe("Schedule", () => {
  function selectStartMonth(year: string, monthLabel: string) {
    cy.get("#schedule-start-month").click();
    cy.get("[data-slot='schedule-start-year-trigger']").click();
    cy.contains("[data-slot='select-item']", year).click({ force: true });
    cy.get("[data-slot='schedule-start-month-trigger']").click();
    cy.contains("[data-slot='select-item']", monthLabel).click({
      force: true,
    });
    cy.contains("button", "Valider").click();
  }

  function listResponse() {
    return {
      success: true,
      schedules: [
        {
          schedule_id: "sch-1",
          quote_id: "q-1",
          status: "DRAFT",
          name: "Echeancier principal",
          start_month: "2026-06",
          duration_months: 3,
        },
      ],
    };
  }

  function detailsResponse(status = "DRAFT", planned = 1400) {
    return {
      success: true,
      schedule: {
        schedule_id: "sch-1",
        quote_id: "q-1",
        status,
        name: "Echeancier principal",
        start_month: "2026-06",
        duration_months: 3,
        lines: [
          {
            quote_line_id: "line-1",
            planned_cents: planned,
            expected_cents: 1500,
          },
        ],
        column_totals: [{ month_index: 1, amount_cents: planned }],
        quote_total_cents: 1500,
        planned_total_cents: planned,
      },
    };
  }

  describe("List", () => {
    it("renders schedules list from API", () => {
      cy.login();
      cy.intercept("GET", "/api/schedules", {
        statusCode: 200,
        body: listResponse(),
      }).as("listSchedules");

      cy.visit("/schedule");
      cy.wait("@listSchedules");

      cy.contains("td", "sch-1").should("be.visible");
      cy.contains("td", "Echeancier principal").should("be.visible");
      cy.contains("td", "q-1").should("be.visible");
      cy.contains("td", "DRAFT").should("be.visible");
      cy.contains("td", "2026-06").should("be.visible");
      cy.contains("td", "3").should("be.visible");
    });

    it("shows empty state when API returns no schedules", () => {
      cy.login();
      cy.intercept("GET", "/api/schedules", {
        statusCode: 200,
        body: { success: true, schedules: [] },
      }).as("listSchedulesEmpty");

      cy.visit("/schedule");
      cy.wait("@listSchedulesEmpty");

      cy.contains("Aucun échéancier.").should("be.visible");
    });

    it("creates a schedule then refreshes list", () => {
      cy.login();

      let created = false;
      cy.intercept("GET", "/api/schedules", (req) => {
        req.reply({
          statusCode: 200,
          body: created
            ? {
                success: true,
                schedules: [
                  {
                    schedule_id: "sch-new",
                    quote_id: "q-9",
                    status: "DRAFT",
                    name: "Nouveau planning",
                    start_month: "2026-10",
                    duration_months: 6,
                  },
                ],
              }
            : { success: true, schedules: [] },
        });
      }).as("listSchedules");

      cy.intercept("POST", "/api/schedules", (req) => {
        created = true;
        req.reply({
          statusCode: 201,
          body: { success: true, schedule_id: "sch-new" },
        });
      }).as("createSchedule");

      cy.intercept("GET", "/api/quotes", {
        statusCode: 200,
        body: {
          success: true,
          quotes: [
            {
              quote_id: "q-9",
              user_id: "u-1",
              name: "Devis Alpha",
              archived_at: null,
              state: "draft",
              client_id: "c-9",
              address_id: 1,
              user_address_id: 1,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }).as("listQuotes");

      cy.intercept("GET", "/api/users/clients", {
        statusCode: 200,
        body: {
          success: true,
          clients: [
            {
              client_id: "c-9",
              user_id: "u-1",
              first_name: "Jean",
              last_name: "Dupont",
              email: "jean@example.com",
              phone: "",
              company: "",
              siren: "",
              vat: "",
              archived: false,
            },
          ],
        },
      }).as("listClients");

      cy.visit("/schedule");
      cy.wait("@listSchedules");

      cy.contains("button", "Nouvel échéancier").click();
      cy.wait("@listQuotes");
      cy.wait("@listClients");
      cy.get("input[name='quote_id']").click();
      cy.contains(
        "[data-slot='combobox-item']",
        "Devis Alpha (Jean Dupont)",
      ).click({ force: true });
      cy.get("input[name='name']").type("Nouveau planning");
      selectStartMonth("2026", "Octobre");
      cy.get("input[name='duration_months']").type("6");
      cy.contains("button", "Créer").click();

      cy.wait("@createSchedule").then(({ request }) => {
        expect(request.body).to.deep.equal({
          quote_id: "q-9",
          name: "Nouveau planning",
          start_month: "2026-10",
          duration_months: 6,
        });
      });
      cy.wait("@listSchedules");
      cy.contains("td", "sch-new").should("be.visible");
    });

    it("shows validation message on create failure", () => {
      cy.login();
      cy.intercept("GET", "/api/schedules", {
        statusCode: 200,
        body: { success: true, schedules: [] },
      }).as("listSchedules");
      cy.intercept("POST", "/api/schedules", {
        statusCode: 422,
        body: { success: false, message: "Données invalides." },
      }).as("createScheduleInvalid");
      cy.intercept("GET", "/api/quotes", {
        statusCode: 200,
        body: {
          success: true,
          quotes: [
            {
              quote_id: "q-1",
              user_id: "u-1",
              name: "Devis Beta",
              archived_at: null,
              state: "draft",
              client_id: "c-1",
              address_id: 1,
              user_address_id: 1,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }).as("listQuotes");
      cy.intercept("GET", "/api/users/clients", {
        statusCode: 200,
        body: {
          success: true,
          clients: [
            {
              client_id: "c-1",
              user_id: "u-1",
              first_name: "Marie",
              last_name: "Martin",
              email: "marie@example.com",
              phone: "",
              company: "",
              siren: "",
              vat: "",
              archived: false,
            },
          ],
        },
      }).as("listClients");

      cy.visit("/schedule");
      cy.wait("@listSchedules");

      cy.contains("button", "Nouvel échéancier").click();
      cy.wait("@listQuotes");
      cy.wait("@listClients");
      cy.get("input[name='quote_id']").click();
      cy.contains(
        "[data-slot='combobox-item']",
        "Devis Beta (Marie Martin)",
      ).click({ force: true });
      cy.get("input[name='name']").type("Bad");
      selectStartMonth("2026", "Septembre");
      cy.get("input[name='duration_months']").type("0");
      cy.contains("button", "Créer").click();

      cy.wait("@createScheduleInvalid");
      cy.contains("Données invalides.").should("be.visible");
    });

    it("sorts quote options by quote name and excludes quotes already VALID", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules", {
        statusCode: 200,
        body: {
          success: true,
          schedules: [
            {
              schedule_id: "sch-valid",
              quote_id: "q-valid",
              status: "VALID",
              name: "Déjà validé",
              start_month: "2026-06",
              duration_months: 3,
            },
          ],
        },
      }).as("listSchedules");

      cy.intercept("GET", "/api/quotes", {
        statusCode: 200,
        body: {
          success: true,
          quotes: [
            {
              quote_id: "q-beta",
              user_id: "u-1",
              name: "Beta",
              archived_at: null,
              state: "draft",
              client_id: "c-1",
              address_id: 1,
              user_address_id: 1,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              quote_id: "q-valid",
              user_id: "u-1",
              name: "Gamma",
              archived_at: null,
              state: "draft",
              client_id: "c-2",
              address_id: 1,
              user_address_id: 1,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              quote_id: "q-alpha",
              user_id: "u-1",
              name: "Alpha",
              archived_at: null,
              state: "draft",
              client_id: "c-1",
              address_id: 1,
              user_address_id: 1,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }).as("listQuotes");

      cy.intercept("GET", "/api/users/clients", {
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
              company: "",
              siren: "",
              vat: "",
              archived: false,
            },
            {
              client_id: "c-2",
              user_id: "u-1",
              first_name: "Marie",
              last_name: "Martin",
              email: "marie@example.com",
              phone: "",
              company: "",
              siren: "",
              vat: "",
              archived: false,
            },
          ],
        },
      }).as("listClients");

      cy.visit("/schedule");
      cy.wait("@listSchedules");

      cy.contains("button", "Nouvel échéancier").click();
      cy.wait("@listQuotes");
      cy.wait("@listClients");
      cy.get("input[name='quote_id']").click();

      cy.get("[data-slot='combobox-item']").should("have.length", 2);
      cy.get("[data-slot='combobox-item']").eq(0).should("contain", "Alpha");
      cy.get("[data-slot='combobox-item']").eq(1).should("contain", "Beta");
      cy.contains("[data-slot='combobox-item']", "Gamma").should("not.exist");
    });
  });

  describe("Details", () => {
    it("renders details and computed balance state", () => {
      cy.login();
      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400),
      }).as("getSchedule");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      cy.contains("Échéancier sch-1").should("be.visible");
      cy.contains("Nom:").should("be.visible");
      cy.contains("Statut:").should("be.visible");
      cy.contains("DRAFT").should("be.visible");
      cy.contains("line-1").should("be.visible");
      cy.contains("under").should("be.visible");
      cy.contains("14.00 €").should("be.visible");
      cy.contains("15.00 €").should("be.visible");
    });

    it("updates first cell then refreshes details", () => {
      cy.login();

      let cellUpdated = false;
      cy.intercept("GET", "/api/schedules/sch-1", (req) => {
        if (cellUpdated) {
          req.alias = "getScheduleAfterUpdate";
        }
        req.reply({
          statusCode: 200,
          body: detailsResponse("DRAFT", cellUpdated ? 0 : 1400),
        });
      }).as("getSchedule");

      cy.intercept("PATCH", "/api/schedules/sch-1/cells", (req) => {
        cellUpdated = true;
        req.reply({ statusCode: 200, body: { success: true } });
      }).as("patchCell");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");
      cy.contains("line-1").should("be.visible");

      cy.contains("button", "Cellule #1 → 0,00")
        .should("be.visible")
        .as("updateFirstCellButton");
      cy.get("@updateFirstCellButton").click({ force: true });

      cy.wait("@patchCell").then(({ request }) => {
        expect(request.body).to.deep.equal({
          quote_line_id: "line-1",
          month_index: 1,
          amount_eur: "0.00",
        });
      });
      cy.wait("@getScheduleAfterUpdate");
    });

    it("validates schedule then refreshes status", () => {
      cy.login();

      let scheduleValidated = false;
      cy.intercept("GET", "/api/schedules/sch-1", (req) => {
        if (scheduleValidated) {
          req.alias = "getScheduleAfterValidate";
        }
        req.reply({
          statusCode: 200,
          body: detailsResponse(scheduleValidated ? "VALID" : "DRAFT", 1500),
        });
      }).as("getSchedule");

      cy.intercept("POST", "/api/schedules/sch-1/validate", (req) => {
        scheduleValidated = true;
        req.reply({ statusCode: 200, body: { success: true } });
      }).as("validateSchedule");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");
      cy.contains("line-1").should("be.visible");

      cy.contains("button", "Valider l'échéancier")
        .should("be.visible")
        .as("validateScheduleButton");
      cy.get("@validateScheduleButton").click({ force: true });

      cy.wait("@validateSchedule");
      cy.wait("@getScheduleAfterValidate");
    });

    it("surfaces API error on details load", () => {
      cy.login();
      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 404,
        body: { success: false, message: "Echeancier introuvable." },
      }).as("getScheduleNotFound");

      cy.visit("/schedule/sch-1");
      cy.wait("@getScheduleNotFound");

      cy.contains("introuvable").should("be.visible");
    });
  });
});
