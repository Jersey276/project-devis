describe("Schedule", () => {
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
