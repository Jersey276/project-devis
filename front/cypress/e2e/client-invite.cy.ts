import { client } from "../support/fixtures";

const clientWithEmail = client({ email: "jean@example.com" });
const clientWithoutEmail = client({ email: "", client_id: "c-2" });
const clientLinked = client({ linked_user_id: "u-linked" });

describe("Client invitation", () => {
  describe("Clients table — row action", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", /^\/api\/users\/clients(\?.*)?$/, {
        statusCode: 200,
        body: { success: true, clients: [clientWithEmail], total: 1 },
      }).as("listClients");
    });

    it("opens invite dialog and sends invitation", () => {
      cy.intercept("POST", "/api/auth/invite/client", {
        statusCode: 200,
        body: { success: true },
      }).as("sendInvite");

      cy.visit("/clients");
      cy.wait("@listClients");

      // Open row actions dropdown (EllipsisVertical button) on the first row
      cy.contains("tr", clientWithEmail.first_name)
        .find("button")
        .last()
        .click();
      cy.contains("Inviter").click();

      // AlertDialog should appear
      cy.contains("Envoyer une invitation").should("be.visible");
      cy.contains(clientWithEmail.email).should("be.visible");

      cy.contains("button", "Envoyer").click();
      cy.wait("@sendInvite");

      cy.contains("Invitation envoyée.").should("be.visible");
    });

    it("dismisses invite dialog on cancel", () => {
      cy.visit("/clients");
      cy.wait("@listClients");

      cy.contains("tr", clientWithEmail.first_name)
        .find("button")
        .last()
        .click();
      cy.contains("Inviter").click();
      cy.contains("Envoyer une invitation").should("be.visible");
      cy.contains("button", "Annuler").click();
      cy.contains("Envoyer une invitation").should("not.exist");
    });
  });

  describe("/accept-invitation — unauthenticated user", () => {
    beforeEach(() => {
      // User not logged in: override auth/me stub set by cy.login()
      cy.intercept("GET", "/api/auth/me", {
        statusCode: 401,
        body: { success: false },
      }).as("authMe");
    });

    it("shows invalid link message when token is missing", () => {
      cy.visit("/accept-invitation");
      cy.contains("Lien invalide").should("be.visible");
    });

    it("shows login and register tabs when not logged in", () => {
      cy.visit("/accept-invitation?token=abc123");
      cy.wait("@authMe");
      cy.contains("Rejoindre votre espace client").should("be.visible");
      cy.contains("J'ai déjà un compte").should("be.visible");
      cy.contains("Créer un compte").should("be.visible");
    });

    it("registers and links via accept endpoint", () => {
      cy.intercept("POST", "/api/auth/invite/accept", {
        statusCode: 200,
        body: { success: true, token: "new-token", refresh_token: "refresh" },
      }).as("acceptNew");
      cy.intercept("GET", "/api/users/clients/me", {
        statusCode: 200,
        body: { success: true, clients: [clientLinked] },
      }).as("getMyClient");
      cy.intercept("GET", /^\/api\/users\/addresses(\?.*)?$/, {
        statusCode: 200,
        body: { success: true, addresses: [] },
      });

      cy.visit("/accept-invitation?token=abc123");
      cy.wait("@authMe");

      cy.contains("Créer un compte").click();
      cy.get("#reg-email").type("new@example.com");
      cy.get("#reg-password").type("Password1!");
      cy.get("#reg-confirm").type("Password1!");
      cy.contains("button", "Créer mon compte").click();

      cy.wait("@acceptNew");
      cy.location("pathname").should("eq", "/client-profile");
    });
  });

  describe("/accept-invitation — authenticated user", () => {
    beforeEach(() => {
      cy.login();
    });

    it("shows link button for already-logged-in user", () => {
      cy.visit("/accept-invitation?token=abc123");
      cy.contains("Lier ce compte").should("be.visible");
    });

    it("links the existing account", () => {
      cy.intercept("POST", "/api/auth/invite/accept-linked", {
        statusCode: 200,
        body: { success: true },
      }).as("acceptLinked");
      cy.intercept("GET", "/api/users/clients/me", {
        statusCode: 200,
        body: { success: true, clients: [clientLinked] },
      }).as("getMyClient");
      cy.intercept("GET", /^\/api\/users\/addresses(\?.*)?$/, {
        statusCode: 200,
        body: { success: true, addresses: [] },
      });

      cy.visit("/accept-invitation?token=abc123");
      cy.contains("button", "Lier ce compte").click();

      cy.wait("@acceptLinked");
      cy.location("pathname").should("eq", "/client-profile");
    });
  });

  describe("/client-profile — customer mode", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/users/clients/me", {
        statusCode: 200,
        body: { success: true, clients: [clientLinked] },
      }).as("getMyClient");
      cy.intercept("GET", /^\/api\/users\/addresses(\?.*)?$/, {
        statusCode: 200,
        body: { success: true, addresses: [] },
      });
    });

    it("shows client profile in customer mode", () => {
      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");
      cy.contains("Mon profil client").should("be.visible");
      cy.contains("Informations").should("be.visible");
      cy.contains("Adresses").should("be.visible");
    });

    it("saves client profile", () => {
      cy.intercept("PUT", "/api/users/clients/me", {
        statusCode: 200,
        body: { success: true, clients: [clientLinked] },
      }).as("updateMyClient");

      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");
      cy.contains("button", "Enregistrer").click();
      cy.wait("@updateMyClient");
      cy.contains("Profil mis à jour.").should("be.visible");
    });

    it("shows Mon profil in sidebar in customer mode", () => {
      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");
      cy.get("[data-sidebar='sidebar']").within(() => {
        cy.contains("Mon profil").should("be.visible");
        cy.contains("Clients").should("not.exist");
      });
    });

    it("shows no link message when client is not linked", () => {
      cy.intercept("GET", "/api/users/clients/me", {
        statusCode: 404,
        body: { success: false },
      }).as("getMyClientNotFound");

      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClientNotFound");
      cy.contains("Votre compte n'est pas encore lié").should("be.visible");
    });
  });
});
