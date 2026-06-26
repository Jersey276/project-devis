import { client } from "../support/fixtures";

const clientWithEmail = client({ email: "jean@example.com" });
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
      cy.intercept("POST", "/api/auth/invite/accept", (req) => {
        req.reply({
          statusCode: 200,
          body: { success: true, token: "new-token", refresh_token: "refresh" },
          headers: { "Set-Cookie": "auth-token=new-token; Path=/; HttpOnly" },
        });
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

    it("se connecte et lie via l'onglet 'J'ai déjà un compte'", () => {
      cy.intercept("POST", "/api/auth/login", {
        statusCode: 200,
        body: { success: true },
      }).as("login");
      cy.intercept("POST", "/api/auth/invite/accept-linked", (req) => {
        req.reply({
          statusCode: 200,
          body: { success: true },
          headers: { "Set-Cookie": "auth-token=new-token; Path=/; HttpOnly" },
        });
      }).as("acceptLogin");
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

      cy.contains("J'ai déjà un compte").click();
      cy.get("#login-email").type("existing@example.com");
      cy.get("#login-password").type("Password1!");
      cy.contains("button", "Se connecter et lier").click();

      cy.wait("@login");
      cy.wait("@acceptLogin");
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
      cy.intercept("PUT", /\/api\/users\/clients\/me/, {
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

    it("affiche les erreurs de validation 422 lors de la sauvegarde", () => {
      cy.intercept("PUT", /\/api\/users\/clients\/me/, {
        statusCode: 422,
        body: {
          success: false,
          errors: [
            { field: "first_name", message: "Prénom requis." },
          ],
        },
      }).as("updateMyClientError");

      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");
      cy.contains("button", "Enregistrer").click();
      cy.wait("@updateMyClientError");

      cy.contains("Prénom requis.").should("be.visible");
    });
  });

  describe("/client-profile — onglet Adresses", () => {
    const address = {
      id: 5,
      owner_type: "client",
      owner_id: clientLinked.client_id,
      name: "Bureau",
      street: "5 avenue Foch",
      additional_street: "",
      city: "Lyon",
      zip_code: "69001",
      country_id: 1,
      email: "",
      phone: "",
      archived: false,
    };

    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/users/clients/me", {
        statusCode: 200,
        body: { success: true, clients: [clientLinked] },
      }).as("getMyClient");
      cy.intercept("GET", /^\/api\/users\/addresses(\?.*)?$/, {
        statusCode: 200,
        body: { success: true, addresses: [address] },
      }).as("listAddresses");
      cy.intercept("GET", "/api/users/countries**", {
        statusCode: 200,
        body: { success: true, countries: [{ id: 1, code: "FR", name: "France" }] },
      });
    });

    it("affiche la liste des adresses dans l'onglet Adresses", () => {
      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");

      cy.contains("Adresses").click();
      cy.wait("@listAddresses");

      cy.contains("5 avenue Foch").should("be.visible");
    });

    it("ouvre la dialog d'ajout d'adresse", () => {
      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");

      cy.contains("Adresses").click();
      cy.wait("@listAddresses");

      cy.contains("button", "Ajouter une adresse").click();
      cy.contains("Nouvelle adresse").should("be.visible");
    });

    it("ajoute une adresse et rafraîchit la liste", () => {
      const newAddress = { ...address, id: 6, name: "Domicile", street: "12 rue Molière" };
      cy.intercept("POST", /\/api\/users\/addresses/, {
        statusCode: 201,
        body: { success: true, address: newAddress },
      }).as("createAddress");

      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");

      cy.contains("Adresses").click();
      cy.wait("@listAddresses");

      cy.contains("button", "Ajouter une adresse").click();
      cy.get("input[name='name']").type("Domicile");
      cy.get("input[name='street']").type("12 rue Molière");
      cy.get("input[name='city']").type("Paris");
      cy.get("input[name='zip_code']").type("75001");
      cy.contains("button", "Enregistrer").click();

      cy.wait("@createAddress");
      cy.contains("Adresse ajoutée.").should("be.visible");
    });

    it("supprime une adresse après confirmation", () => {
      cy.intercept("DELETE", /\/api\/users\/addresses\/5/, {
        statusCode: 200,
        body: { success: true },
      }).as("deleteAddress");

      cy.visitAs("customer", "/client-profile");
      cy.wait("@getMyClient");

      cy.contains("Adresses").click();
      cy.wait("@listAddresses");

      cy.contains("tr", "5 avenue Foch").within(() => {
        cy.get("button[aria-label='Actions']").click();
      });
      cy.contains("Supprimer").click();
      cy.contains("Supprimer cette adresse ?").should("be.visible");
      cy.get("[role='alertdialog']").within(() => {
        cy.contains("button", "Supprimer").click();
      });

      cy.wait("@deleteAddress");
      cy.contains("Adresse supprimée.").should("be.visible");
    });
  });
});
