const USER = {
  user_id: "u1",
  email: "john@test.fr",
  phone: "0102030405",
  company: "Acme",
  siren: "123456789",
  siret: "12345678900012",
  vat: "FR12345",
};

const COUNTRIES = [
  { id: 1, code: "FR", name: "France" },
  { id: 2, code: "BE", name: "Belgique" },
];

const INITIAL_ADDRESSES = [
  {
    id: 10,
    owner_type: "user",
    owner_id: USER.user_id,
    name: "Principale",
    street: "12 Rue des Lilas",
    additional_street: "",
    city: "Paris",
    zip_code: "75001",
    country_id: 1,
    email: "",
    phone: "",
    archived: false,
  },
];

function stubProfile(addresses = INITIAL_ADDRESSES) {
  cy.login();
  cy.intercept("GET", "/api/users/me", {
    statusCode: 200,
    body: { success: true, user: USER },
  }).as("getMe");
  cy.intercept("GET", "/api/users/countries", {
    statusCode: 200,
    body: { success: true, countries: COUNTRIES },
  }).as("getCountries");
  // Addresses moved from /me/addresses to /addresses?owner_type=user&owner_id=…
  // when the polymorphic owner model landed. Match with a glob so query-arg
  // order doesn't matter.
  cy.intercept("GET", "/api/users/addresses?**", {
    statusCode: 200,
    body: { success: true, addresses },
  }).as("getAddresses");
  cy.intercept("GET", "/api/auth/oauth-identities", {
    statusCode: 200,
    body: { success: true, identities: [], has_password: true },
  }).as("getOAuthIdentities");
  cy.intercept("POST", "/api/auth/email/request-change", {
    statusCode: 200,
    body: { success: true },
  }).as("requestEmailChange");
  cy.intercept("GET", "/api/subscriptions/me", {
    statusCode: 200,
    body: {
      success: true,
      subscription: {
        subscription_id: "",
        user_id: USER.user_id,
        plan_id: 1,
        tier: "free",
        status: "active",
        current_period_start: "",
        current_period_end: null,
        cancel_at_period_end: false,
        stripe_subscription_id: null,
        updated_at: "",
      },
    },
  }).as("getMySubscription");
  cy.intercept("GET", "/api/plans", {
    statusCode: 200,
    body: {
      success: true,
      plans: [
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
      ],
    },
  }).as("getPlans");
}

describe("Profile page", () => {
  describe("structure", () => {
    beforeEach(() => stubProfile());

    it("shows the four tabs", () => {
      cy.visit("/profile");
      cy.wait("@getMe");
      cy.get("[role='tablist']").should("exist");
      cy.get("[role='tab']").should("have.length", 4);
      cy.contains("[role='tab']", "Information").should("exist");
      cy.contains("[role='tab']", "Adresses").should("exist");
      cy.contains("[role='tab']", "Connexion").should("exist");
      cy.contains("[role='tab']", "Abonnement").should("exist");
    });

    it("switches tabs on click", () => {
      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.contains("button", "Ajouter une adresse").should("be.visible");
      cy.contains("[role='tab']", "Connexion").click();
      cy.contains("Adresse email").should("be.visible");
    });
  });

  describe("Information tab", () => {
    beforeEach(() => stubProfile());

    it("prefills fields from /api/users/me", () => {
      cy.visit("/profile");
      cy.wait("@getMe");
      cy.get("input[name='email']").should("have.value", USER.email);
      cy.get("input[name='email']").should("have.attr", "readonly");
      cy.get("input[name='phone']").should("have.value", USER.phone);
      cy.get("input[name='company']").should("have.value", USER.company);
      cy.get("input[name='siren']").should("have.value", USER.siren);
      cy.get("input[name='vat']").should("have.value", USER.vat);
    });

    it("submits update and shows success toast", () => {
      cy.intercept("PUT", "/api/users/me", {
        statusCode: 200,
        body: { success: true },
      }).as("updateMe");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.get("input[name='phone']").clear().type("0606060606");
      cy.contains("button", "Enregistrer").click();

      cy.wait("@updateMe").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          first_name: "",
          last_name: "",
          phone: "0606060606",
          company: USER.company,
          siren: USER.siren,
          siret: USER.siret,
          vat: USER.vat,
          oss_enabled: false,
          iban: "",
          bic: "",
        });
      });
      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Informations mises à jour.",
      );
    });

    it("shows inline field error on 422", () => {
      cy.intercept("PUT", "/api/users/me", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "phone", error_code: [2] }],
        },
      }).as("updateMeInvalid");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.get("input[name='phone']").clear().type("bad");
      cy.contains("button", "Enregistrer").click();
      cy.wait("@updateMeInvalid");

      cy.get("input[name='phone']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Format invalide.");
    });

    it("shows error toast on 500", () => {
      cy.intercept("PUT", "/api/users/me", {
        statusCode: 500,
        body: { success: false, message: "Erreur serveur." },
      }).as("updateMeError");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("button", "Enregistrer").click();
      cy.wait("@updateMeError");

      cy.get("[data-sonner-toaster]").should("contain", "Erreur serveur.");
    });

    it("shows error toast on network failure", () => {
      cy.intercept("PUT", "/api/users/me", { forceNetworkError: true }).as(
        "updateMeNetwork",
      );

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("button", "Enregistrer").click();

      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Une erreur est survenue.",
      );
    });
  });

  describe("Addresses tab", () => {
    beforeEach(() => stubProfile());

    it("displays existing addresses", () => {
      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.wait("@getAddresses");

      cy.contains("Principale").should("be.visible");
      cy.contains("12 Rue des Lilas").should("be.visible");
      cy.contains("75001 Paris").should("be.visible");
      cy.contains("France").should("be.visible");
    });

    it("creates a new address (success)", () => {
      cy.intercept("POST", "/api/users/addresses", {
        statusCode: 201,
        body: { success: true, address_id: 11 },
      }).as("createAddress");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.wait("@getAddresses");

      cy.contains("button", "Ajouter une adresse").click();
      cy.get("[data-slot='dialog-content']").should("be.visible");
      // AddressForm fires /api/users/countries on mount; wait for it so the
      // combobox is populated before we start typing.
      cy.wait("@getCountries");
      cy.get("input[name='name']").should("be.visible");

      cy.get("input[name='name']").type("Bureau");
      cy.get("input[name='street']").type("3 Rue de Rivoli");
      cy.get("input[name='city']").type("Lyon");
      cy.get("input[name='zip_code']").type("69002");
      cy.get("input[name='country_id']").type("Bel");
      cy.contains("[data-slot='combobox-item']", "Belgique").click({
        force: true,
      });

      cy.contains("[data-slot='dialog-footer'] button", "Enregistrer").click();

      cy.wait("@createAddress").then((interception) => {
        expect(interception.request.body).to.include({
          owner_type: "user",
          owner_id: USER.user_id,
          name: "Bureau",
          street: "3 Rue de Rivoli",
          city: "Lyon",
          zip_code: "69002",
          country_id: 2,
        });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Adresse ajoutée.");
      cy.get("[data-slot='dialog-content']").should("not.exist");
    });

    it("shows inline errors on 422 when creating", () => {
      cy.intercept("POST", "/api/users/addresses", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "name", error_code: [1] }],
        },
      }).as("createAddressInvalid");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.wait("@getAddresses");

      cy.contains("button", "Ajouter une adresse").click();
      cy.contains("[data-slot='dialog-footer'] button", "Enregistrer").click();
      cy.wait("@createAddressInvalid");

      cy.get("input[name='name']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Ce champ est requis.");
      cy.get("[data-slot='dialog-content']").should("be.visible");
    });

    it("edits an existing address", () => {
      cy.intercept("PUT", "/api/users/addresses/10", {
        statusCode: 200,
        body: { success: true },
      }).as("updateAddress");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.wait("@getAddresses");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Modifier").click();
      cy.get("input[name='city']").should("have.value", "Paris");
      cy.get("input[name='city']").clear().type("Marseille");
      cy.contains("[data-slot='dialog-footer'] button", "Enregistrer").click();

      cy.wait("@updateAddress").then((interception) => {
        expect(interception.request.body).to.include({
          owner_type: "user",
          owner_id: USER.user_id,
          city: "Marseille",
        });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Adresse mise à jour.");
    });

    it("cancels deletion (no API call)", () => {
      cy.intercept("DELETE", "/api/users/addresses/10?**", () => {
        throw new Error("DELETE should not be called when canceling");
      }).as("deleteAddressCancel");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.wait("@getAddresses");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.get("[data-slot='alert-dialog-content']").should("be.visible");
      cy.contains("[data-slot='alert-dialog-cancel']", "Annuler").click();
      cy.get("[data-slot='alert-dialog-content']").should("not.exist");
      cy.contains("Principale").should("be.visible");
    });

    it("deletes an address (success)", () => {
      cy.intercept("DELETE", "/api/users/addresses/10?**", {
        statusCode: 200,
        body: { success: true },
      }).as("deleteAddress");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.wait("@getAddresses");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();

      // Post-delete reload must return an empty list — register only now so the
      // initial GET above still resolves with INITIAL_ADDRESSES (LIFO matching).
      cy.intercept("GET", "/api/users/addresses?**", {
        statusCode: 200,
        body: { success: true, addresses: [] },
      });
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteAddress");
      cy.get("[data-sonner-toaster]").should("contain", "Adresse supprimée.");
      cy.contains("Aucune adresse pour le moment.").should("be.visible");
    });

    it("shows error toast on delete failure", () => {
      cy.intercept("DELETE", "/api/users/addresses/10?**", {
        statusCode: 500,
        body: { success: false, message: "Échec serveur." },
      }).as("deleteAddressFail");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Adresses").click();
      cy.wait("@getAddresses");

      cy.get("table [data-slot='dropdown-menu-trigger']").first().click();
      cy.contains("Supprimer").click();
      cy.contains("[data-slot='alert-dialog-action']", "Supprimer").click();

      cy.wait("@deleteAddressFail");
      cy.get("[data-sonner-toaster]").should("contain", "Échec serveur.");
      cy.contains("Principale").should("be.visible");
    });
  });

  describe("Connection tab", () => {
    beforeEach(() => stubProfile());

    it("shows email section with current email read-only", () => {
      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.contains("Adresse email").should("be.visible");
      cy.get("input[name='current_email']").should("have.value", USER.email);
      cy.get("input[name='current_email']").should("have.attr", "readonly");
    });

    it("sends email change request and shows confirmation notice", () => {
      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.get("input[name='new_email']").type("new@example.com");
      cy.contains("button", "Envoyer le lien de confirmation").click();

      cy.wait("@requestEmailChange").then((interception) => {
        expect(interception.request.body).to.deep.equal({ new_email: "new@example.com" });
      });
      cy.get("[data-sonner-toaster]").should("contain", "Lien de confirmation envoyé.");
      cy.contains("new@example.com").should("be.visible");
    });

    it("shows error toast when email change request fails", () => {
      cy.intercept("POST", "/api/auth/email/request-change", {
        statusCode: 409,
        body: { success: false, message: "Cette adresse email est déjà utilisée." },
      }).as("requestEmailChangeFail");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.get("input[name='new_email']").type("taken@example.com");
      cy.contains("button", "Envoyer le lien de confirmation").click();

      cy.wait("@requestEmailChangeFail");
      cy.get("[data-sonner-toaster]").should("contain", "Cette adresse email est déjà utilisée.");
    });

    it("validates password confirmation client-side without calling API", () => {
      cy.intercept("POST", "/api/auth/password/update", () => {
        throw new Error("API must not be called on confirm mismatch");
      }).as("updatePasswordMismatch");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.get("input[name='old_password']").type("currentPass1");
      cy.get("input[name='new_password']").type("newPassword1");
      cy.get("input[name='confirm_password']").type("different");
      cy.contains("button", "Mettre à jour le mot de passe").click();

      cy.get("input[name='confirm_password']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Les mots de passe ne correspondent pas.");
    });

    it("submits password update successfully", () => {
      cy.intercept("POST", "/api/auth/password/update", {
        statusCode: 200,
        body: { success: true, message: "Mot de passe mis à jour." },
      }).as("updatePassword");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.get("input[name='old_password']").type("currentPass1");
      cy.get("input[name='new_password']").type("newPassword1");
      cy.get("input[name='confirm_password']").type("newPassword1");
      cy.contains("button", "Mettre à jour le mot de passe").click();

      cy.wait("@updatePassword").then((interception) => {
        expect(interception.request.body).to.deep.equal({
          old_password: "currentPass1",
          new_password: "newPassword1",
        });
      });
      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Mot de passe mis à jour.",
      );
      cy.get("input[name='old_password']").should("have.value", "");
    });

    it("shows error toast on invalid current password (401)", () => {
      cy.intercept("POST", "/api/auth/password/update", {
        statusCode: 401,
        body: {
          success: false,
          code: 1003,
          message: "Adresse email ou mot de passe incorrect.",
        },
      }).as("updatePasswordWrong");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.get("input[name='old_password']").type("wrongPass1");
      cy.get("input[name='new_password']").type("newPassword1");
      cy.get("input[name='confirm_password']").type("newPassword1");
      cy.contains("button", "Mettre à jour le mot de passe").click();

      cy.wait("@updatePasswordWrong");
      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Adresse email ou mot de passe incorrect.",
      );
    });

    it("shows error toast when backend returns 501 not implemented", () => {
      cy.intercept("POST", "/api/auth/password/update", {
        statusCode: 501,
        body: {
          success: false,
          code: 2003,
          message: "Cette fonctionnalité n'est pas encore disponible.",
        },
      }).as("updatePasswordNotImpl");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.get("input[name='old_password']").type("currentPass1");
      cy.get("input[name='new_password']").type("newPassword1");
      cy.get("input[name='confirm_password']").type("newPassword1");
      cy.contains("button", "Mettre à jour le mot de passe").click();

      cy.wait("@updatePasswordNotImpl");
      cy.get("[data-sonner-toaster]").should(
        "contain",
        "Cette fonctionnalité n'est pas encore disponible.",
      );
    });

    it("shows inline error for new_password too short (422)", () => {
      cy.intercept("POST", "/api/auth/password/update", {
        statusCode: 422,
        body: {
          success: false,
          field_errors: [{ field: "new_password", error_code: [3] }],
        },
      }).as("updatePasswordShort");

      cy.visit("/profile");
      cy.wait("@getMe");
      cy.contains("[role='tab']", "Connexion").click();

      cy.get("input[name='old_password']").type("currentPass1");
      cy.get("input[name='new_password']").type("short1");
      cy.get("input[name='confirm_password']").type("short1");
      cy.contains("button", "Mettre à jour le mot de passe").click();

      cy.wait("@updatePasswordShort");
      cy.get("input[name='new_password']")
        .closest("[data-slot='field']")
        .find("[data-slot='field-error']")
        .should("contain", "Trop court (12 caractères minimum).");
    });
  });
});
