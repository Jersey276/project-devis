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
  pour l'OSS, **ou** que le cumul annuel de ces ventes atteint le seuil legal de
  10 000 EUR, **ou** que ce seuil a ete franchi l'annee precedente (regle annee
  N-1, art. 259 D du CGI : TVA destination des le 1er euro). La bascule
  origine -> destination est automatique ; voir
  `docs/adr/0002-oss-seuil-bascule-automatique.md`.
  - Les avoirs sont deduits de l'assiette (annee courante ET N-1) : le cumul
    reflete le chiffre d'affaires net des ventes a distance B2C intra-UE.
  - `GetOSSThresholdStatus` expose `prior_year_over_threshold` (+ cumul N-1) ; la
    banniere front affiche une mention dediee quand l'OSS est actif via la regle N-1.
  - Simplification connue restante (C4) : la facture qui franchit le seuil en cours
    d'annee peut rester en TVA origine (bascule effective sur la suivante).
- **Coordonnees de paiement (BG-16)** : l'IBAN et le BIC de l'emetteur (saisis sur
  son profil) sont geles dans le snapshot a l'emission (`issuer_iban` /
  `issuer_bic` sur `invoice_party_snapshots` et `credit_note_party_snapshots`,
  migration `000010`). Le service export les emet dans le XML CII comme groupe
  BG-16 (moyen de paiement BT-81 = `30` virement, IBAN BT-84, BIC BT-86) ; un avoir
  herite des coordonnees de la facture origine. Le groupe est omis si aucun IBAN
  n'est renseigne (le document reste valide EN 16931, notamment pour les factures
  en franchise / B2C).
- **SIRET de routage (B4)** : le SIRET (14 chiffres = SIREN + NIC etablissement)
  de l'emetteur et du destinataire est la cle de routage du destinataire dans
  l'annuaire DGFiP. Saisi sur le profil emetteur et la fiche client (validation :
  14 chiffres, doit commencer par le SIREN s'il est renseigne), il est gele dans le
  snapshot a l'emission (`issuer_siret` / `client_siret` sur
  `invoice_party_snapshots` et `credit_note_party_snapshots`, migration `000012`) ;
  un avoir herite du SIRET de la facture origine. Le service export l'emet comme
  identifiant d'immatriculation legale (BT-30 / BT-47) dans
  `SpecifiedLegalOrganization/ID` avec le schema ISO 6523 `0009` (registre SIRET),
  **de preference au SIREN** (schema `0002`) : un seul des deux est emis par partie.
  Le groupe est omis si ni SIRET ni SIREN ne sont renseignes (valide EN 16931, B2C).
  Le lookup annuaire lui-meme (resolution SIRET -> plateforme destinataire) releve
  de l'integration PDP/PA (B6).
- **Depot plateforme (B6, 1ere iteration)** : abstraction neutre `PDPClient`
  (package `backend/invoice/pdp/`) vers une Plateforme Agreee (PA, ex-PDP), avec
  un adaptateur **no-op** par defaut (aucune PA contractee : voir la decision de
  cadrage) et un mock pour les tests. Le RPC `DepositInvoice` depose une facture
  emise puis fait avancer le cycle de vie e-invoicing a `DEPOSITED` **via les memes
  gardes** que la transition manuelle (`applyLifecycleTransition` : verrou de ligne,
  garde `ISSUED|PAID`, table de transitions, journal append-only). Le handle
  plateforme est gele dans `pdp_submission_id` (migration `000013`, nullable). La
  meme migration **corrige un bug latent** : le trigger d'inalterabilite (`000003`)
  n'autorisait sur facture scellee que la transition `ISSUED -> PAID` ; il accepte
  desormais aussi une MAJ ne touchant que `lifecycle_status` / `pdp_submission_id`
  / `updated_at`, toutes les colonnes legales/financieres restant gelees (le
  document legal et son scellement chaine restent inalterables). Gateway :
  `POST /api/invoices/:id/deposit`. Front : bouton « Deposer sur la plateforme »
  (statut e-invoicing `NONE`), le reste des transitions restant manuel.
- **Reconciliation des statuts (B6, 2e iteration)** : un worker de fond
  (`PollPDPStatuses` / `StartPDPPoller`, `backend/invoice/actions/pdp_poll.go`)
  reconcilie le cycle de vie local avec la PA. A chaque balayage il selectionne
  les factures portant un `pdp_submission_id` et un statut non terminal
  (`DEPOSITED|RECEIVED|APPROVED`), interroge `FetchStatus`, et — si le statut
  plateforme est plus avance — fait progresser le cycle **un cran a la fois**
  (le flux B3 interdit les sauts : un saut PA `DEPOSITED -> APPROVED` est
  decompose en `RECEIVED` puis `APPROVED`). Chaque cran passe par
  `applyLifecycleTransition` dans sa propre transaction (gardes + journal
  append-only), donc l'avancement est idempotent et un echec tardif ne defait pas
  les crans deja commit. `REJECTED` est applique directement depuis tout etat
  actif ; un statut `UNKNOWN` (defaut de l'adaptateur no-op) ne touche a rien — le
  worker est **inerte en production** tant qu'aucune PA n'est branchee. Active par
  `PDP_POLL_INTERVAL` (duree Go, ex. `30s` ; vide = desactive), chaque balayage
  borne par un timeout.
- **Lookup annuaire (B6, 3e iteration)** : une seconde abstraction neutre
  `Directory` (`backend/invoice/pdp/`) resout le destinataire (SIRET) vers sa
  plateforme de reception via l'annuaire central e-invoicing (DGFiP/AIFE), avec un
  adaptateur **no-op** (resout tout le monde, handle vide) et un mock. Avant tout
  appel a la PA, `DepositInvoice` resout le `client_siret` **gele du snapshot
  legal** (B4) : un destinataire introuvable (`ErrRecipientNotFound`) **bloque le
  depot** (`RecipientNotInDirectory`, 4014) — on ne route jamais vers un
  destinataire que l'annuaire ne place pas, et la PA n'est pas appelee. Le handle
  de routage retourne est gele dans `recipient_routing_id` (migration `000014`,
  nullable) dans la **meme transaction** que le `pdp_submission_id`. La migration
  `000014` etend le trigger d'inalterabilite (`000013`) pour traiter
  `recipient_routing_id` comme metadonnee e-invoicing operationnelle (mutable sur
  facture scellee, colonnes legales/financieres toujours gelees). Le no-op resolvant
  tout le monde, le depot (et son endpoint `POST /api/invoices/:id/deposit`) reste
  inchange tant qu'aucun annuaire reel n'est branche.
- **Adaptateur PA reel Iopole (B6, derniere iteration)** : implementation concrete
  des seams `pdp.Client` / `pdp.Directory` contre l'API Iopole (Plateforme Agreee),
  dans le sous-package `backend/invoice/pdp/iopole/` (transport HTTP + OAuth2 isoles
  hors du package `pdp` neutre, qui reste sans I/O). Selection par variable
  d'environnement : **`PDP_PROVIDER=iopole`** branche l'adaptateur reel ; toute autre
  valeur (dont l'absence) garde les adaptateurs no-op — la production reste donc
  inerte tant qu'on ne configure pas explicitement le fournisseur. Auth = OAuth2
  client_credentials (Keycloak, `IOPOLE_TOKEN_URL`), header tenant `customer-id`
  (`IOPOLE_CUSTOMER_ID`). Depot = `POST /v1/invoice` en `multipart/form-data` (champ
  `file`), ou l'on envoie le **PDF/A-3 Factur-X** obtenu en reutilisant le pipeline
  EN16931 valide : `DepositInvoice` (via l'adaptateur) appelle
  `export.ExportInvoice(facturx=true)` (client gRPC `services/exportgrpc/`,
  `EXPORT_SERVICE_ADDRESS`) plutot que de regenerer le document. L'`id` retourne par
  Iopole devient le `pdp_submission_id` gele. Statut = `GET
  /v1/invoice/{id}/status-history` : on prend l'item le plus recent et on mappe son
  `status.code` (15 valeurs Iopole) vers le vocabulaire `PlatformStatus`
  (`SUBMITTED|ISSUED -> SUBMITTED`, `RECEIVED|MADE_AVAILABLE|IN_HAND -> RECEIVED`,
  `APPROVED|PARTIALLY_APPROVED|COMPLETED -> APPROVED`, `PAYMENT_SENT|PAYMENT_RECEIVED
  -> COLLECTED`, `REJECTED|REFUSED|UNACCEPTABLE -> REJECTED`, `DISPUTED|SUSPENDED` et
  inconnu -> `UNKNOWN` = aucune transition). Annuaire = `GET /v1/directory/french?q=
  {siret}` : `data` vide -> `ErrRecipientNotFound` (donc 4014). Note dependance :
  `export` est deja client de `invoice` (`export -> invoice.GetInvoice`) ; le
  back-call `invoice -> export.ExportInvoice` est un appel reseau distinct et
  non-reentrant (pas de cycle a la compilation, pas de deadlock).
- **Cadre e-invoicing FR** : le Factur-X B2C/OSS genere est coherent mais
  facultatif (l'obligation PPF/PDP vise le B2B domestique ; le transfrontalier
  releve de l'e-reporting).

## Codes d'erreur specifiques

Voir `backend/invoice/actions/codes/codes.go` et `docs/ERROR_CODES.md`. Notamment
`OSSDestinationTaxMissing` (4010) : l'OSS s'applique mais aucun taux n'est
configure pour le pays du client, l'emission est bloquee.
`PDPSubmissionFailed` (4013, B6) : le depot sur la plateforme a echoue (l'appel
PA renvoie une erreur ou un statut inattendu) ; le cycle de vie reste inchange.
La gateway le mappe en `502 Bad Gateway`.

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
