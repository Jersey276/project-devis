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

  function openStatusSelect() {
    cy.get("[data-slot='select-trigger']").click({ force: true });
  }

  function chooseStatusOption(label: string) {
    cy.contains("[data-slot='select-item']", label).click({ force: true });
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

  function detailsResponse(
    status = "DRAFT",
    planned = 1400,
    options?: { durationMonths?: number; lineCount?: number },
  ) {
    const durationMonths = options?.durationMonths ?? 3;
    const lineCount = options?.lineCount ?? 1;
    const lines = Array.from({ length: lineCount }, (_, index) => ({
      quote_line_id: `line-${index + 1}`,
      planned_cents: planned,
      expected_cents: 1500,
    }));

    const cells = Array.from({ length: lineCount }, (_, index) => ({
      quote_line_id: `line-${index + 1}`,
      month_index: 1,
      amount_cents: planned,
    }));

    return {
      success: true,
      schedule: {
        schedule_id: "sch-1",
        quote_id: "q-1",
        status,
        name: "Echeancier principal",
        start_month: "2026-06",
        duration_months: durationMonths,
        lines,
        cells,
        column_totals: [{ month_index: 1, amount_cents: planned * lineCount }],
        quote_total_cents: 1500 * lineCount,
        planned_total_cents: planned * lineCount,
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
      cy.contains("Brouillon").should("be.visible");
      cy.contains("td", "2026-06").should("be.visible");
      cy.contains("td", "3").should("be.visible");
    });

    it("changes schedule status from the list with confirmation for denied", () => {
      cy.login();

      let currentStatus = "DRAFT";
      cy.intercept("GET", "/api/schedules", (req) => {
        req.reply({
          statusCode: 200,
          body: {
            success: true,
            schedules: [
              {
                schedule_id: "sch-1",
                quote_id: "q-1",
                status: currentStatus,
                name: "Echeancier principal",
                start_month: "2026-06",
                duration_months: 3,
              },
            ],
          },
        });
      }).as("listSchedules");

      cy.intercept("PATCH", "/api/schedules/sch-1/status", (req) => {
        currentStatus = req.body.status;
        req.reply({
          statusCode: 200,
          body: { success: true, status: req.body.status },
        });
      }).as("updateScheduleStatus");

      cy.on("window:confirm", (message) => {
        expect(message).to.include("Confirmer le refus");
        return true;
      });

      cy.visit("/schedule");
      cy.wait("@listSchedules");

      cy.contains("td", "sch-1")
        .closest("tr")
        .within(() => {
          cy.get("[data-slot='select-trigger']").click({ force: true });
        });
      chooseStatusOption("Refusé");

      cy.wait("@updateScheduleStatus").then(({ request }) => {
        expect(request.body).to.deep.equal({ status: "DENIED" });
      });
      cy.wait("@listSchedules");
      cy.contains("td", "sch-1")
        .closest("tr")
        .contains("Refusé")
        .should("be.visible");
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
      cy.get("input[name='duration_months']")
        .click({ force: true })
        .clear({ force: true })
        .type("6", { force: true })
        .should("have.value", "6");
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
      cy.get("input[name='duration_months']")
        .click({ force: true })
        .clear({ force: true })
        .type("0", { force: true })
        .should("have.value", "0");
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
    it("renders details and applies balance row color", () => {
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
      cy.contains("Brouillon").should("be.visible");
      cy.contains("2026").should("be.visible");
      cy.contains("juin").should("be.visible");
      cy.contains("juillet").should("be.visible");
      cy.contains("line-1").should("be.visible");
      cy.contains("td", "line-1")
        .closest("tr")
        .should("have.class", "bg-amber-50/60");
      cy.get("[data-testid='footer-month-total-1']").should(
        "contain",
        "14.00 €",
      );
      cy.contains("14.00 €").should("be.visible");
      cy.contains("15.00 €").should("be.visible");
    });

    it("saves edited cell on blur", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400),
      }).as("getSchedule");

      cy.intercept("PATCH", "/api/schedules/sch-1/cells", {
        statusCode: 200,
        body: { success: true },
      }).as("patchCellInline");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      cy.get("input[name='cell-line-1-m1']").clear().type("12.34").blur();

      cy.wait("@patchCellInline").then(({ request }) => {
        expect(request.body).to.deep.equal({
          quote_line_id: "line-1",
          month_index: 1,
          amount_eur: "12.34",
        });
      });

      cy.get("@getSchedule.all").should("have.length", 1);
      cy.get("[data-testid='line-remaining-line-1']").should(
        "contain",
        "2.66 €",
      );
      cy.get("[data-testid='footer-month-total-1']").should(
        "contain",
        "12.34 €",
      );
    });

    it("saves edited cell on Enter", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400),
      }).as("getSchedule");

      cy.intercept("PATCH", "/api/schedules/sch-1/cells", {
        statusCode: 200,
        body: { success: true },
      }).as("patchCellInlineOnEnter");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      cy.get("input[name='cell-line-1-m1']").clear().type("11.11{enter}");

      cy.wait("@patchCellInlineOnEnter").then(({ request }) => {
        expect(request.body).to.deep.equal({
          quote_line_id: "line-1",
          month_index: 1,
          amount_eur: "11.11",
        });
      });
    });

    it("navigates between cells with arrows and tab keys", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400, {
          lineCount: 2,
          durationMonths: 3,
        }),
      }).as("getSchedule");

      cy.intercept("PATCH", "/api/schedules/sch-1/cells", {
        statusCode: 200,
        body: { success: true },
      }).as("patchCellNav");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      cy.get("input[name='cell-line-1-m1']").focus().type("{rightarrow}");
      cy.focused().should("have.attr", "name", "cell-line-1-m2");

      cy.focused().type("{downarrow}");
      cy.focused().should("have.attr", "name", "cell-line-2-m2");

      cy.focused().type("{leftarrow}");
      cy.focused().should("have.attr", "name", "cell-line-2-m1");

      cy.focused().type("{uparrow}");
      cy.focused().should("have.attr", "name", "cell-line-1-m1");

      cy.focused().trigger("keydown", { key: "Tab" });
      cy.focused().should("have.attr", "name", "cell-line-1-m2");

      cy.focused().trigger("keydown", { key: "Tab", shiftKey: true });
      cy.focused().should("have.attr", "name", "cell-line-1-m1");

      cy.get("@patchCellNav.all").should("have.length", 0);
    });

    it("keeps horizontal overflow inside table container", () => {
      cy.login();
      cy.viewport(1024, 768);

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400, { durationMonths: 18 }),
      }).as("getScheduleWide");

      cy.visit("/schedule/sch-1");
      cy.wait("@getScheduleWide");

      cy.get("[data-slot='table-container']").then(($el) => {
        const container = $el[0];
        expect(container.scrollWidth).to.be.greaterThan(container.clientWidth);
      });

      cy.get("[data-slot='table-container']").scrollTo("right");
      cy.get("[data-testid='footer-month-total-18']").should("be.visible");
    });

    it("rejects invalid cell value and does not send PATCH", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400),
      }).as("getSchedule");

      cy.intercept("PATCH", "/api/schedules/sch-1/cells", {
        statusCode: 200,
        body: { success: true },
      }).as("patchCellInvalid");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      cy.get("input[name='cell-line-1-m1']").clear().type("-12").blur();

      cy.get("@patchCellInvalid.all").should("have.length", 0);
      cy.get("input[name='cell-line-1-m1']").should(
        "have.attr",
        "aria-invalid",
        "true",
      );
      cy.get("[data-testid='cell-error-line-1-m1']")
        .should("have.attr", "aria-label")
        .and("include", "Montant invalide");
    });

    it("changes status to VALID from details then refreshes", () => {
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

      cy.intercept("PATCH", "/api/schedules/sch-1/status", (req) => {
        scheduleValidated = true;
        req.reply({
          statusCode: 200,
          body: { success: true, status: "VALID" },
        });
      }).as("updateScheduleStatus");

      cy.on("window:confirm", (message) => {
        expect(message).to.include("Confirmer la validation");
        return true;
      });

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");
      cy.contains("line-1").should("be.visible");

      openStatusSelect();
      chooseStatusOption("Validé");

      cy.wait("@updateScheduleStatus").then(({ request }) => {
        expect(request.body).to.deep.equal({ status: "VALID" });
      });
      cy.wait("@getScheduleAfterValidate");
      cy.get("input[name='cell-line-1-m1']").should("be.disabled");
      cy.contains("Validé").should("be.visible");
    });

    it("shows explicit message when status update to VALID is refused", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1200),
      }).as("getSchedule");

      cy.intercept("PATCH", "/api/schedules/sch-1/status", {
        statusCode: 422,
        body: { success: false, message: "L'échéancier n'est pas équilibré." },
      }).as("updateScheduleStatusRejected");

      cy.on("window:confirm", (message) => {
        expect(message).to.include("Confirmer la validation");
        return true;
      });

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      openStatusSelect();
      chooseStatusOption("Validé");

      cy.wait("@updateScheduleStatusRejected");
      cy.contains("n'est pas équilibré").should("be.visible");
      cy.get("@getSchedule.all").should("have.length", 1);
    });

    it("shows other schedules as DENIED in list after validation", () => {
      cy.login();

      let validated = false;
      cy.intercept("GET", "/api/schedules/sch-1", (req) => {
        req.reply({
          statusCode: 200,
          body: detailsResponse(validated ? "VALID" : "DRAFT", 1500),
        });
      }).as("getSchedule");

      cy.intercept("PATCH", "/api/schedules/sch-1/status", (req) => {
        validated = true;
        req.reply({
          statusCode: 200,
          body: { success: true, status: "VALID" },
        });
      }).as("updateScheduleStatus");

      cy.on("window:confirm", () => true);

      cy.intercept("GET", "/api/schedules", {
        statusCode: 200,
        body: {
          success: true,
          schedules: [
            {
              schedule_id: "sch-1",
              quote_id: "q-1",
              status: "VALID",
              name: "Echeancier principal",
              start_month: "2026-06",
              duration_months: 3,
            },
            {
              schedule_id: "sch-2",
              quote_id: "q-1",
              status: "DENIED",
              name: "Echeancier alternatif",
              start_month: "2026-06",
              duration_months: 3,
            },
          ],
        },
      }).as("listSchedulesAfterValidate");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");
      openStatusSelect();
      chooseStatusOption("Validé");
      cy.wait("@updateScheduleStatus");

      cy.visit("/schedule");
      cy.wait("@listSchedulesAfterValidate");
      cy.contains("td", "sch-2")
        .closest("tr")
        .contains("Refusé")
        .should("be.visible");
    });

    it("locks cell editing when schedule is already VALID", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("VALID", 1500),
      }).as("getScheduleValid");

      cy.intercept("PATCH", "/api/schedules/sch-1/cells", {
        statusCode: 200,
        body: { success: true },
      }).as("patchCellWhileValid");

      cy.visit("/schedule/sch-1");
      cy.wait("@getScheduleValid");

      cy.get("input[name='cell-line-1-m1']").should("be.disabled");
      cy.contains("Validé").should("be.visible");
      cy.get("@patchCellWhileValid.all").should("have.length", 0);
    });

    it("locks cell editing when schedule is DENIED", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DENIED", 1200),
      }).as("getScheduleDenied");

      cy.intercept("PATCH", "/api/schedules/sch-1/cells", {
        statusCode: 200,
        body: { success: true },
      }).as("patchCellWhileDenied");

      cy.visit("/schedule/sch-1");
      cy.wait("@getScheduleDenied");

      cy.get("input[name='cell-line-1-m1']").should("be.disabled");
      cy.contains("Refusé").should("be.visible");
      cy.get("@patchCellWhileDenied.all").should("have.length", 0);
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

    it("exports schedule as PDF", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400),
      }).as("getSchedule");

      cy.intercept("GET", "/api/export/schedules/sch-1", {
        statusCode: 200,
        headers: {
          "content-type": "application/pdf",
          "content-disposition":
            'attachment; filename="echeancier-Echeancier-principal.pdf"',
        },
        body: "fake-pdf",
      }).as("exportSchedulePdf");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      cy.contains("button", "Exporter PDF").click();

      cy.wait("@exportSchedulePdf").then((interception) => {
        expect(interception.request.method).to.equal("GET");
        expect(
          interception.response?.headers?.["content-disposition"],
        ).to.include("echeancier-Echeancier-principal.pdf");
      });
      cy.contains("Export PDF impossible.").should("not.exist");
    });

    it("shows error when schedule PDF export fails", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("DRAFT", 1400),
      }).as("getSchedule");

      cy.intercept("GET", "/api/export/schedules/sch-1", {
        statusCode: 500,
        body: { success: false, message: "Une erreur interne est survenue." },
      }).as("exportSchedulePdfError");

      cy.visit("/schedule/sch-1");
      cy.wait("@getSchedule");

      cy.contains("button", "Exporter PDF").click();

      cy.wait("@exportSchedulePdfError")
        .its("request.method")
        .should("eq", "GET");
      cy.contains("Export PDF impossible.").should("be.visible");
    });

    it("exports schedule as PDF when schedule is VALID", () => {
      cy.login();

      cy.intercept("GET", "/api/schedules/sch-1", {
        statusCode: 200,
        body: detailsResponse("VALID", 1500),
      }).as("getScheduleValid");

      cy.intercept("GET", "/api/export/schedules/sch-1", {
        statusCode: 200,
        headers: {
          "content-type": "application/pdf",
          "content-disposition":
            'attachment; filename="echeancier-Echeancier-principal.pdf"',
        },
        body: "fake-pdf",
      }).as("exportSchedulePdfValid");

      cy.visit("/schedule/sch-1");
      cy.wait("@getScheduleValid");

      cy.contains("button", "Exporter PDF").should("be.enabled").click();

      cy.wait("@exportSchedulePdfValid")
        .its("request.method")
        .should("eq", "GET");
      cy.contains("Export PDF impossible.").should("not.exist");
    });
  });
});
