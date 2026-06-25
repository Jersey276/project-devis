import { line, quote, type LineFixture } from "../support/fixtures";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function stubMe(company = "Acme SARL") {
  cy.intercept("GET", "/api/users/me", {
    statusCode: 200,
    body: {
      success: true,
      user: { user_id: "u-1", email: "user@test.fr", company },
    },
  }).as("getMe");
}

function stubQuote(lines: LineFixture[] = []) {
  cy.intercept("GET", "/api/quotes/q-1", {
    statusCode: 200,
    body: {
      success: true,
      quote: quote({ quote_id: "q-1", name: "Devis Alpha" }),
      lines,
    },
  }).as("getQuote");
}

function stubComments(lineId: string, comments: object[] = []) {
  cy.intercept("GET", `/api/quotes/q-1/lines/${lineId}/comments`, {
    statusCode: 200,
    body: { success: true, comments },
  }).as("listComments");
}

function stubTaxes() {
  cy.intercept("GET", "/api/users/taxes/available**", {
    statusCode: 200,
    body: { success: true, taxes: [] },
  }).as("listTaxes");
}

function stubAddresses() {
  cy.intercept("GET", /\/api\/users\/addresses/, {
    statusCode: 200,
    body: { success: true, addresses: [] },
  }).as("listAddresses");
}

function visitQuoteEdit() {
  cy.visit("/quote/q-1");
  cy.wait("@getQuote");
}

function goToItemsStep() {
  cy.get("[data-step-tab='1']").click();
}

const COMMENT_FIXTURE = {
  comment_id: "cmt-1",
  line_id: "l-1",
  quote_id: "q-1",
  author_id: "u-1",
  author_name: "Acme SARL",
  body: "Premier commentaire",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

const OTHER_COMMENT = {
  comment_id: "cmt-2",
  line_id: "l-1",
  quote_id: "q-1",
  author_id: "u-other",
  author_name: "Bob",
  body: "Commentaire de Bob",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

// ─── Suite ────────────────────────────────────────────────────────────────────

describe("Commentaires de devis", () => {
  const lineFixture = line({ line_id: "l-1", quote_id: "q-1", name: "Design UI" });

  beforeEach(() => {
    cy.login();
    stubMe();
    stubTaxes();
    stubAddresses();
  });

  // ── Ouverture via icône sur une ligne ─────────────────────────────────────

  describe("Ouverture depuis une ligne", () => {
    it("ouvre la sidebar avec le nom de la ligne dans le titre", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", []);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.contains("Commentaires — Design UI").should("be.visible");
    });

    it("affiche l'état vide quand la ligne n'a pas de commentaires", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", []);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.contains("Aucun commentaire pour cette ligne.").should("be.visible");
    });

    it("liste les commentaires existants", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE, OTHER_COMMENT]);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.contains("Acme SARL").should("be.visible");
      cy.contains("Premier commentaire").should("be.visible");
      cy.contains("Bob").should("be.visible");
      cy.contains("Commentaire de Bob").should("be.visible");
    });
  });

  // ── Créer un commentaire ──────────────────────────────────────────────────

  describe("Création", () => {
    it("envoie un commentaire et l'affiche dans la liste", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", []);
      visitQuoteEdit();
      goToItemsStep();

      const newComment = { ...COMMENT_FIXTURE, body: "Nouveau message" };
      cy.intercept("POST", "/api/quotes/q-1/lines/l-1/comments", {
        statusCode: 201,
        body: { success: true, comment: newComment },
      }).as("createComment");

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get("textarea[placeholder='Écrire un commentaire…']").type("Nouveau message");
      cy.contains("button", "Envoyer").click();
      cy.wait("@createComment");

      cy.contains("Nouveau message").should("be.visible");
    });

    it("vide le champ après envoi", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", []);
      visitQuoteEdit();
      goToItemsStep();

      cy.intercept("POST", "/api/quotes/q-1/lines/l-1/comments", {
        statusCode: 201,
        body: { success: true, comment: COMMENT_FIXTURE },
      }).as("createComment");

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get("textarea[placeholder='Écrire un commentaire…']").type("Test");
      cy.contains("button", "Envoyer").click();
      cy.wait("@createComment");

      cy.get("textarea[placeholder='Écrire un commentaire…']").should("have.value", "");
    });

    it("désactive le bouton Envoyer quand le champ est vide", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", []);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.contains("button", "Envoyer").should("be.disabled");
    });

    it("envoie le commentaire avec Ctrl+Entrée", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", []);
      visitQuoteEdit();
      goToItemsStep();

      cy.intercept("POST", "/api/quotes/q-1/lines/l-1/comments", {
        statusCode: 201,
        body: { success: true, comment: COMMENT_FIXTURE },
      }).as("createCommentCtrl");

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get("textarea[placeholder='Écrire un commentaire…']")
        .type("Message clavier{ctrl+enter}");
      cy.wait("@createCommentCtrl");
    });
  });

  // ── Modifier un commentaire (auteur uniquement) ───────────────────────────

  describe("Modification", () => {
    it("affiche les boutons modifier/supprimer uniquement pour ses propres commentaires", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE, OTHER_COMMENT]);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      // propre commentaire : boutons visibles
      cy.get(`[aria-label="Modifier"]`).should("have.length", 1);
      cy.get(`[aria-label="Supprimer"]`).should("have.length", 1);
    });

    it("passe en mode édition inline au clic sur Modifier", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE]);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get(`[aria-label="Modifier"]`).click();
      cy.get("textarea").first().should("have.value", "Premier commentaire");
    });

    it("sauvegarde la modification et met à jour la liste", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE]);
      visitQuoteEdit();
      goToItemsStep();

      const updated = { ...COMMENT_FIXTURE, body: "Commentaire modifié" };
      cy.intercept("PUT", "/api/quotes/q-1/lines/l-1/comments/cmt-1", {
        statusCode: 200,
        body: { success: true, comment: updated },
      }).as("updateComment");

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get(`[aria-label="Modifier"]`).click();
      cy.get("textarea").first().clear().type("Commentaire modifié");
      cy.get("[data-testid='comment-save']").click();
      cy.wait("@updateComment");

      cy.contains("Commentaire modifié").should("be.visible");
      // le mode édition est fermé
      cy.get("[data-testid='comment-save']").should("not.exist");
    });

    it("annule l'édition au clic sur Annuler", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE]);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get(`[aria-label="Modifier"]`).click();
      cy.get("textarea").first().clear().type("Modification annulée");
      cy.get("[data-testid='comment-cancel']").click();

      cy.contains("Premier commentaire").should("be.visible");
      cy.get("[data-testid='comment-save']").should("not.exist");
    });
  });

  // ── Supprimer un commentaire ──────────────────────────────────────────────

  describe("Suppression", () => {
    it("demande confirmation avant de supprimer", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE]);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get(`[aria-label="Supprimer"]`).click();
      cy.contains("Supprimer ce commentaire ?").should("be.visible");
    });

    it("supprime le commentaire après confirmation", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE]);
      visitQuoteEdit();
      goToItemsStep();

      cy.intercept("DELETE", "/api/quotes/q-1/lines/l-1/comments/cmt-1", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteComment");

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get(`[aria-label="Supprimer"]`).click();
      cy.contains("Supprimer ce commentaire ?").should("be.visible");
      // Le bouton de confirmation dans l'AlertDialog
      cy.get("[role='alertdialog']").within(() => {
        cy.contains("button", "Supprimer").click();
      });
      cy.wait("@deleteComment");

      cy.contains("Premier commentaire").should("not.exist");
      cy.contains("Aucun commentaire pour cette ligne.").should("be.visible");
    });

    it("garde le commentaire si on clique sur Annuler dans la dialog", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", [COMMENT_FIXTURE]);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");

      cy.get(`[aria-label="Supprimer"]`).click();
      cy.get("[role='alertdialog']").within(() => {
        cy.contains("button", "Annuler").click();
      });

      cy.contains("Premier commentaire").should("be.visible");
    });
  });

  // ── Fermeture de la sidebar ───────────────────────────────────────────────

  describe("Fermeture", () => {
    it("ferme la sidebar au clic sur la croix", () => {
      stubQuote([lineFixture]);
      stubComments("l-1", []);
      visitQuoteEdit();
      goToItemsStep();

      cy.get(`[aria-label="Commentaires de la ligne"]`).first().click();
      cy.wait("@listComments");
      cy.contains("Commentaires — Design UI").should("be.visible");

      cy.get("[data-slot='sheet-close']").click();
      cy.contains("Commentaires — Design UI").should("not.exist");
    });
  });
});
