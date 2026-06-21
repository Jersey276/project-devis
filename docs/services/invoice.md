# Service Invoice

## Role

Generer et gerer les factures et avoirs a partir d'un devis valide ou d'un
echeancier valide:

- factures (cycle DRAFT -> ISSUED -> PAID, + CANCELLED)
- avoirs (credit notes) numerotes
- snapshot legal des parties, lignes et ventilation TVA a l'emission
- numerotation continue par (utilisateur, annee)
- scellement cryptographique des documents emis (inalterabilite)

## Point d'entree

- `backend/invoice/main.go`

## Pattern de demarrage

1. connexion DB
2. migrations embed
3. exposition gRPC sur `:50059`

## Dossiers cles

- `backend/invoice/actions/`
- `backend/invoice/services/`
- `backend/invoice/migrations/`

## Consommateurs

- gateway (`/api/invoices/*`, gated par `authz.ResourceSubscriptionInvoices`)
- export (`ExportInvoice` / `ExportCreditNote`, assemblage PDF et Factur-X)

## Dependances amont

- quote (lignes et donnees du devis source)
- users (mentions emetteur/client, taxes, pays)
- schedule (cellules par ligne/mois pour les factures issues d'un echeancier)

## Particularites

- **Stockage hybride** : references conservees pour les montants deja geles par la
  validation du devis/echeancier ; snapshot a l'emission du bloc non gele en amont
  (mentions des parties, taux de TVA par ligne).
- **Numerotation** continue sans trou, format `YYYY-NNNN`, annee fixee en
  Europe/Paris, attribuee uniquement a l'emission.
- **Conformite B2C / OSS** : pour une vente a distance B2C intra-UE (hors FR), la
  TVA appliquee est celle du pays de destination des lors que le vendeur a opte
  pour l'OSS **ou** que le cumul annuel de ces ventes atteint le seuil legal de
  10 000 EUR (art. 259 D du CGI). La bascule origine -> destination est automatique
  une fois le seuil atteint ; voir `docs/adr/0002-oss-seuil-bascule-automatique.md`.
  - Les avoirs sont deduits de l'assiette : le cumul reflete le chiffre
    d'affaires net des ventes a distance B2C intra-UE.
  - Simplifications connues : la facture qui franchit le seuil peut rester en TVA
    origine (bascule effective sur la suivante) ; la regle annee N-1 n'est pas
    encore implementee.
- **Cadre e-invoicing FR** : le Factur-X B2C/OSS genere est coherent mais
  facultatif (l'obligation PPF/PDP vise le B2B domestique ; le transfrontalier
  releve de l'e-reporting).

## Codes d'erreur specifiques

Voir `backend/invoice/actions/codes/codes.go` et `docs/ERROR_CODES.md`. Notamment
`OSSDestinationTaxMissing` (4010) : l'OSS s'applique mais aucun taux n'est
configure pour le pays du client, l'emission est bloquee.

## Ports

Le service n'est pas encore declare dans `docker-compose.prod.yml` (branche
`feat/invoice` en cours). La colonne production ci-dessous decrit la cible.

| Contexte           |       Port | Direction   | Note                              |
| ------------------ | ---------: | ----------- | --------------------------------- |
| Processus invoice  |      50059 | ecoute gRPC | flag `-port` (defaut 50059)       |
| Docker local       | non publie | interne     | atteint via `devis-invoice:50059` |
| Docker production  | non publie | interne     | cible (non encore declare)        |

## Variables d'environnement (vue exhaustive)

### Variables declarees par le service (`services/env.go`)

La colonne prod n'est pas encore renseignee (service absent du compose prod).

| Variable                   | Usage                       | Definie local |
| -------------------------- | --------------------------- | ------------- |
| `ENV`                      | convention environnement    | non           |
| `POSTGRES_USER`            | compat legacy               | non           |
| `POSTGRES_PASSWORD`        | compat legacy               | non           |
| `POSTGRES_DB`              | compat legacy               | non           |
| `POSTGRES_DB_ADDRESS`      | compat legacy               | non           |
| `POSTGRES_DB_PORT`         | compat legacy               | non           |
| `DB_HOST`                  | connexion DB                | oui           |
| `DB_PORT`                  | connexion DB                | oui           |
| `DB_USER`                  | connexion DB                | oui           |
| `DB_PASSWORD`              | fallback secret direct      | non           |
| `DB_PASSWORD_FILE`         | secret DB via fichier       | oui           |
| `DB_NAME`                  | base cible                  | oui           |
| `QUOTE_SERVICE_ADDRESS`    | client gRPC quote           | oui           |
| `USER_SERVICE_ADDRESS`     | client gRPC users           | oui           |
| `SCHEDULE_SERVICE_ADDRESS` | client gRPC schedule        | oui           |

## Tests

- Logique pure en TDD dans `actions/` via shims `*_export_test_shim.go`.
- Tests d'integration (numerotation concurrente, scellement, snapshot, cumul OSS)
  gardes par `INVOICE_TEST_DATABASE_URL` ; chaque test s'execute dans un schema
  jetable.
