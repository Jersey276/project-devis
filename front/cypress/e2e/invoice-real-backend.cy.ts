// This spec exercises the REAL backend (no cy.intercept mocking) end-to-end:
// register → create country → create client → create quote → validate →
// generate invoice. It only runs when CYPRESS_REAL_BACKEND=1 (set only in
// CI's e2e-integration job, which boots the full Docker stack); everywhere
// else it's a no-op describe.skip.
//
// It relies on the fact that the very first account ever registered against
// a fresh backend is auto-promoted to super_admin (backend/auth/actions/
// provision.go), which is the only way to reach /countries and create the
// country a quote's addresses require — there is no seed data and no other
// reachable way to create one. This means the test MUST run first, before
// any other spec that might register a user against the same database.

const realBackend = Cypress.env("realBackend") === true;

(realBackend ? describe : describe.skip)("Invoice — real backend", () => {
  it("creates and issues an invoice from a validated quote", () => {
    const stamp = Date.now();
    const email = `e2e-invoice-${stamp}@example.com`;
    const password = "password123456";

    // 1. Register — the first account on a fresh DB becomes super_admin.
    cy.visit("/register");
    cy.get("input[name='email']").type(email);
    cy.get("input[name='password']").type(password);
    cy.get("input[name='confirm-password']").type(password);
    cy.get("#cgv").click();
    cy.get("button[type='submit']").click();
    cy.url().should("include", "/login");

    // 2. Log in.
    cy.get("form[data-hydrated]");
    cy.get("input[name='email']").type(email);
    cy.get("input[name='password']").type(password);
    cy.get("button[type='submit']").click();
    cy.url().should("not.include", "/login");

    // 3. Create a country via the real admin flow (reachable only because
    // this account is super_admin).
    cy.visit("/countries");
    cy.contains("button", "Nouveau pays", { timeout: 10000 }).click();
    cy.get("[data-slot='dialog-content']").should("be.visible");
    cy.get("input[name='code']").type("FR");
    cy.get("input[name='name']").type("France");
    cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();
    cy.get("[data-slot='dialog-content']").should("not.exist");

    // 4. Create a client with an address using that country.
    cy.visit("/clients/create");
    cy.get("input[name='first_name']").type("Jean");
    cy.get("input[name='last_name']").type("Dupont");
    cy.get("input[name='street']").type("2 rue de Lyon");
    cy.get("input[name='city']").type("Lyon");
    cy.get("input[name='zip_code']").type("69001");
    cy.get("input[name='country_id']").click();
    cy.contains("[data-slot='combobox-item']", "France").click({ force: true });
    cy.contains("button", "Créer le compte client").click();
    cy.url().should("match", /\/clients\/[^/]+$/);

    // 5. Create a quote: name, own (issuer) address, client, client address.
    cy.visit("/quote/create");
    cy.get("input[name='name']").type("Devis e2e");

    cy.get("input[name='user_address_id']").click();
    cy.contains("button", "Nouvelle adresse").click();
    cy.get("[data-slot='dialog-content']").within(() => {
      cy.get("input[name='street']").type("1 rue de Paris");
      cy.get("input[name='city']").type("Paris");
      cy.get("input[name='zip_code']").type("75001");
      cy.get("input[name='country_id']").click();
    });
    cy.contains("[data-slot='combobox-item']", "France").click({ force: true });
    cy.contains("[data-slot='dialog-content'] button", "Enregistrer").click();
    cy.get("[data-slot='dialog-content']").should("not.exist");

    cy.get("input[name='client_id']").type("Jean");
    cy.contains("[data-slot='combobox-item']", "Jean Dupont").click({ force: true });
    cy.get("input[name='address_id']").click();
    cy.contains("[data-slot='combobox-item']", "2 rue de Lyon").click({ force: true });

    cy.contains("button", "Suivant").click();
    cy.url().should("include", "step=2");

    // 6. Add a single line, no tax (no tax rate exists — none is required).
    cy.get("[aria-label='Ajouter une ligne']").click();
    cy.contains("[role='menuitem']", "Ligne simple").click();
    cy.get("[data-line-id]")
      .first()
      .within(() => {
        cy.get("input[name='line-name']").type("Prestation e2e");
        cy.get("input[name='line-quantity']").clear().type("1");
        cy.get("input[name='line-unit-price']").clear().type("100");
      });
    // Blur the last field so the line save request fires before continuing.
    cy.get("body").click(0, 0);

    // 7. Validate the quote: draft → negociation → validated.
    cy.contains("button", "Changer l'état").click();
    cy.contains("[role='menuitem']", "Mettre en négociation").click();
    cy.get("[data-slot='alert-dialog-content']").should("be.visible");
    cy.contains("[data-slot='alert-dialog-content'] button", "Mettre en négociation").click();
    cy.get("[data-quote-state='negociation']").should("exist");

    cy.contains("button", "Changer l'état").click();
    cy.contains("[role='menuitem']", "Valider").click();
    cy.get("[data-slot='alert-dialog-content']").should("be.visible");
    cy.contains("[data-slot='alert-dialog-content'] button", "Valider").click();
    cy.get("[data-quote-state='validated']").should("exist");

    // 8. Generate the invoice and check it was issued with the right totals.
    cy.contains("button", "Générer une facture").click();
    cy.url({ timeout: 10000 }).should("include", "/invoice/");

    cy.contains(/Facture N° \d{4}-\d{4}/).should("be.visible");
    cy.contains("100.00 €").should("be.visible");
  });
});
