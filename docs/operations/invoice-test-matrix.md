# Matrice De Tests - Facturation

Objectif: definir la strategie TDD de la fonctionnalite de facturation (factures et avoirs) et fournir une matrice directement exploitable en tickets de sprint et en criteres de QA. Chaque ligne backend correspond a un test Go existant (colonne `Source`) ou a une case identifiee comme a couvrir.

## Regle TDD

Pour toute evolution de la fonctionnalite facturation:

1. le test est ecrit avant le code fonctionnel
2. le test doit echouer initialement (rouge)
3. le code minimal est implemente pour passer au vert
4. le refactor est autorise uniquement apres retour au vert

Aucun endpoint, regle metier, migration ou modification du format scelle ne doit etre ajoute sans test prealable.

## Perimetre fonctionnel couvert

- service dedie facturation (gRPC `:50059`)
- creation de facture depuis un devis valide ou depuis un echeancier valide
- cycle de vie facture: `DRAFT` -> `ISSUED` -> `PAID`, plus `CANCELLED`
- avoirs (credit notes) numerotes, total ou partiel par ligne
- snapshot legal des parties, lignes et ventilation TVA, gele a l'emission
- numerotation continue sans trou par (utilisateur, annee), formats `YYYY-NNNN` et `AV-YYYY-NNNN`
- scellement cryptographique des documents emis (chainage par emetteur, inalterabilite)
- conformite B2C / OSS: TVA du pays de destination, bascule automatique au seuil de 10 000 EUR

## Regles metier a verrouiller par les tests

- statuts facture: `DRAFT`, `ISSUED`, `PAID`, `CANCELLED`
- seule une facture `DRAFT` est modifiable ou supprimable; `ISSUED`/`PAID`/`CANCELLED` sont immuables
- la source doit etre eligible: devis valide, ou echeancier `VALID`
- pour une source echeancier, les mois deja factures (par des factures `ISSUED`/`PAID`) ne sont pas refacturables
- la TVA est arrondie une fois par taux (convention FR), pas par ligne
- en franchise (`vat_exempt`, art. 293 B CGI) aucune TVA n'est calculee
- numerotation continue sans trou: le numero n'est consomme qu'au commit (rollback = compteur intact), garantie meme en emission concurrente
- l'annee de numerotation est fixee en Europe/Paris
- un document emis est scelle: toute alteration de la facture, du snapshot de lignes ou du sceau est bloquee par trigger; seule la transition `ISSUED` -> `PAID` reste autorisee
- le hash de contenu est un format de fil gele: tout changement invalide les sceaux existants et exige une migration de re-scellement
- avoir: credit total du reste ou credit partiel des lignes selectionnees; aucune ligne ne peut etre creditee deux fois
- OSS applicable seulement pour un particulier (B2C) dans un pays UE hors FR, des lors que le vendeur a opte OU que le cumul annuel atteint 10 000 EUR

## Priorisation

- `P0`: bloque la livraison fonctionnelle ou metier
- `P1`: important, mais peut suivre le socle initial
- `P2`: finition ou robustesse etendue

## Matrice De Tests

### Backend metier - Calcul des totaux TVA

<!-- markdownlint-disable MD060 -->

| ID         | Niveau       | Preconditions                                  | Action                            | Resultat attendu                                                          | Priorite | Source                                              |
| ---------- | ------------ | ---------------------------------------------- | --------------------------------- | ------------------------------------------------------------------------- | -------- | --------------------------------------------------- |
| INV-BE-001 | Unit         | Une ligne a 100,00 EUR HT, taux 20%            | Calculer les totaux               | HT 10000, TVA 2000, TTC 12000, une ligne de ventilation `20`              | P0       | `compute_test.go:TestComputeTotals_SingleRate20`    |
| INV-BE-002 | Unit         | Plusieurs taux (0 / 5,5 / 10 / 20)             | Calculer les totaux               | Totaux corrects, ventilation triee par taux croissant                     | P0       | `compute_test.go:TestComputeTotals_MultiRate...`    |
| INV-BE-003 | Unit         | Deux lignes au meme taux 5,5%                  | Calculer les totaux               | HT agrege avant arrondi, TVA arrondie une fois par groupe                 | P0       | `compute_test.go:TestComputeTotals_RoundsPerRate...`|
| INV-BE-004 | Unit         | Lignes avec TVA mais `vat_exempt = true`       | Calculer les totaux               | TVA 0, TTC = HT, ventilation vide (franchise)                             | P0       | `compute_test.go:TestComputeTotals_VATExempt`       |
| INV-BE-005 | Unit         | Aucune ligne                                   | Calculer les totaux               | Tous les totaux a zero, ventilation vide                                  | P1       | `compute_test.go:TestComputeTotals_Empty`           |

<!-- markdownlint-enable MD060 -->

### Backend metier - Eligibilite source et selection des mois

<!-- markdownlint-disable MD060 -->

| ID         | Niveau | Preconditions                                       | Action                          | Resultat attendu                                            | Priorite | Source                                                  |
| ---------- | ------ | --------------------------------------------------- | ------------------------------- | ----------------------------------------------------------- | -------- | ------------------------------------------------------- |
| INV-BE-010 | Unit   | Echeancier dans chaque statut                       | Tester l'eligibilite            | Seul `VALID` est facturable; `DRAFT`/`NEGOCIATE`/`DENIED`/"" refuses | P0 | `sources_test.go:TestScheduleEligible`                  |
| INV-BE-011 | Unit   | Echeancier 6 mois, aucun mois deja facture          | Selectionner les mois 1 et 2    | Selection valide (`Success`)                                | P0       | `sources_test.go:TestValidateMonthSelection_Valid`      |
| INV-BE-012 | Unit   | Selection de mois vide                               | Valider la selection            | `InvalidInput`                                              | P0       | `sources_test.go:TestValidateMonthSelection_Empty`      |
| INV-BE-013 | Unit   | Mois hors plage (7 sur 6, ou 0)                     | Valider la selection            | `InvalidInput`                                              | P0       | `sources_test.go:TestValidateMonthSelection_OutOfRange` |
| INV-BE-014 | Unit   | Mois en double dans la selection                    | Valider la selection            | `InvalidInput`                                              | P0       | `sources_test.go:TestValidateMonthSelection_Duplicates` |
| INV-BE-015 | Unit   | Mois 1 et 2 deja factures, selection 2 et 3         | Valider la selection            | `MonthsAlreadyBilled`                                       | P0       | `sources_test.go:TestValidateMonthSelection_AlreadyBilled` |
| INV-BE-016 | Unit   | Mois 1 et 2 deja factures, selection 3 et 4         | Valider la selection            | `Success` (selection disjointe)                            | P0       | `sources_test.go:TestValidateMonthSelection_Disjoint...`|

<!-- markdownlint-enable MD060 -->

### Backend metier - Numerotation (factures et avoirs)

<!-- markdownlint-disable MD060 -->

| ID         | Niveau         | Preconditions                          | Action                                  | Resultat attendu                                                       | Priorite | Source                                                       |
| ---------- | -------------- | -------------------------------------- | --------------------------------------- | --------------------------------------------------------------------- | -------- | ------------------------------------------------------------ |
| INV-BE-020 | Unit           | Annee/sequence donnees                 | Formater le numero de facture           | `YYYY-NNNN` zero-padde sur 4 chiffres                                  | P0       | `numbering_test.go:TestFormatInvoiceNumber`                  |
| INV-BE-021 | Unit (sqlmock) | Sequence inexistante pour (user,annee) | Allouer un numero                       | Premiere valeur = 1, numero `2026-0001`                               | P0       | `numbering_test.go:TestAllocateInvoiceNumber_FirstOfYear`    |
| INV-BE-022 | Unit (sqlmock) | Sequence deja a 6                      | Allouer un numero                       | Incremente a 7, numero `2026-0007`                                    | P0       | `numbering_test.go:TestAllocateInvoiceNumber_Increments...`  |
| INV-BE-023 | Integration DB | Base Postgres jetable                  | 50 allocations concurrentes             | 50 sequences 1..50 sans trou ni doublon (gap-free sous concurrence)   | P0       | `numbering_integration_test.go:..._Concurrent` (garde par `INVOICE_TEST_DATABASE_URL`) |
| INV-BE-024 | Unit           | Annee/sequence donnees                 | Formater le numero d'avoir              | `AV-YYYY-NNNN`                                                        | P0       | `credit_note_numbering_test.go:TestFormatCreditNoteNumber`   |
| INV-BE-025 | Unit (sqlmock) | Sequence avoir inexistante             | Allouer un numero d'avoir               | Premiere valeur = 1, numero `AV-2026-0001`                           | P0       | `credit_note_numbering_test.go:..._FirstOfYear`              |
| INV-BE-026 | Unit (sqlmock) | Sequence avoir a 6                     | Allouer un numero d'avoir               | Incremente a 7, numero `AV-2026-0007`                               | P0       | `credit_note_numbering_test.go:..._Increments`               |

<!-- markdownlint-enable MD060 -->

### Backend metier - Selection des lignes d'avoir

<!-- markdownlint-disable MD060 -->

| ID         | Niveau | Preconditions                                          | Action                                | Resultat attendu                                       | Priorite | Source                                                         |
| ---------- | ------ | ------------------------------------------------------ | ------------------------------------- | ------------------------------------------------------ | -------- | -------------------------------------------------------------- |
| INV-BE-030 | Unit   | Lignes 0,1,2 ; ligne 1 deja creditee ; demande vide    | Resoudre les positions a crediter     | Credit total du reste: lignes [0,2], `isTotal=true`    | P0       | `credit_note_select_test.go:..._TotalOfRemainder`              |
| INV-BE-031 | Unit   | Toutes les lignes deja creditees ; demande vide        | Resoudre les positions                | `CreditNoteNoLinesLeft`                                | P0       | `credit_note_select_test.go:..._NothingLeft`                   |
| INV-BE-032 | Unit   | Demande explicite [2,0], rien de credite               | Resoudre les positions                | Selection triee [0,2], `isTotal=false` (partiel)       | P0       | `credit_note_select_test.go:..._PartialValid`                  |
| INV-BE-033 | Unit   | Demande [0,1] couvrant toutes les lignes               | Resoudre les positions                | `Success`, `isTotal=true`                              | P1       | `credit_note_select_test.go:..._PartialIsTotalWhenAllSelected` |
| INV-BE-034 | Unit   | Demande d'une ligne deja creditee                      | Resoudre les positions                | `CreditNoteLineAlreadyCredited`                       | P0       | `credit_note_select_test.go:..._AlreadyCredited`               |
| INV-BE-035 | Unit   | Position demandee hors plage                           | Resoudre les positions                | `InvalidInput`                                        | P0       | `credit_note_select_test.go:..._OutOfRange`                    |
| INV-BE-036 | Unit   | Position en double dans la demande                     | Resoudre les positions                | `InvalidInput`                                        | P0       | `credit_note_select_test.go:..._Duplicate`                     |

<!-- markdownlint-enable MD060 -->

### Backend metier - Conformite B2C / OSS

<!-- markdownlint-disable MD060 -->

| ID         | Niveau         | Preconditions                                          | Action                              | Resultat attendu                                                       | Priorite | Source                                                       |
| ---------- | -------------- | ------------------------------------------------------ | ----------------------------------- | --------------------------------------------------------------------- | -------- | ------------------------------------------------------------ |
| INV-BE-040 | Unit           | Liste de taxes du pays destination                     | Choisir le taux standard            | Defaut s'il est marque, sinon le taux le plus eleve; vide -> aucun    | P0       | `oss_test.go:TestPickDestinationTax`                         |
| INV-BE-041 | Unit           | Combinaisons (opt-in, cumul, type client, pays)        | Decider l'application OSS           | Vrai seulement si B2C + UE hors FR + (opt-in OU cumul >= 10 000 EUR)   | P0       | `oss_threshold_test.go:TestOSSApplies`                       |
| INV-BE-042 | Integration DB | Factures `ISSUED`/`PAID` flaguees, autres exclues      | Calculer le cumul annuel            | Somme uniquement des factures assiette de l'annee Paris, hors brouillon courant | P0 | `oss_threshold_integration_test.go:..._SumsAssietteForYear`  |
| INV-BE-043 | Integration DB | Utilisateur sans facture qualifiante                   | Calculer le cumul annuel            | Retourne 0 (et non une erreur)                                        | P1       | `oss_threshold_integration_test.go:..._Empty`                |
| INV-BE-044 | Integration DB | Cumul juste sous puis au seuil                         | Calculer le cumul + decider OSS     | OSS faux juste sous le seuil sans opt-in, vrai une fois le seuil atteint | P0    | `oss_threshold_integration_test.go:..._ThresholdBoundary`    |

<!-- markdownlint-enable MD060 -->

### Backend metier - Scellement et inalterabilite

<!-- markdownlint-disable MD060 -->

| ID         | Niveau         | Preconditions                                       | Action                                          | Resultat attendu                                                                     | Priorite | Source                                                  |
| ---------- | -------------- | --------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------ | -------- | ------------------------------------------------------- |
| INV-BE-050 | Unit           | Document de reference fige                           | Calculer le hash de contenu                     | Hash egal a la valeur golden (format de fil gele)                                    | P0       | `seal_test.go:TestComputeContentHash_Golden`            |
| INV-BE-051 | Unit           | Hash de contenu du document golden                  | Calculer le hash de chaine genesis              | Hash de chaine egal a la valeur golden                                               | P0       | `seal_test.go:TestComputeChainHash_GenesisGolden`       |
| INV-BE-052 | Unit           | Meme document appele deux fois                       | Recalculer le hash                              | Hash stable (deterministe)                                                           | P0       | `seal_test.go:TestComputeContentHash_Stable`            |
| INV-BE-053 | Unit           | Mutation de chaque champ legal (montant, numero...) | Recalculer le hash                              | Le hash change pour chaque champ mute (sensibilite complete)                         | P0       | `seal_test.go:TestComputeContentHash_FieldSensitivity`  |
| INV-BE-054 | Unit           | Variation de prevHash / index / contenu             | Recalculer le hash de chaine                    | Le hash change pour chacune des trois entrees                                        | P0       | `seal_test.go:TestComputeChainHash_Sensitivity`         |
| INV-BE-055 | Unit           | Meme instant exprime en UTC et en Europe/Paris      | Calculer le hash de contenu                     | Hash identique (independance fuseau)                                                 | P1       | `seal_test.go:TestComputeContentHash_TimezoneIndependent` |
| INV-BE-056 | Integration DB | Deux factures emises non scellees                   | Backfill puis verification de chaine            | 2 sceaux crees, chaine valide (checked=2), backfill idempotent (re-run = 2 sceaux)   | P0       | `seal_integration_test.go:TestSeal_BackfillAndVerify`   |
| INV-BE-057 | Integration DB | Facture scellee                                     | UPDATE/DELETE facture, snapshot ligne ou sceau  | Chaque tentative d'alteration est bloquee par trigger                                | P0       | `seal_integration_test.go:TestSeal_TriggerBlocksTamper` |
| INV-BE-058 | Integration DB | Facture scellee `ISSUED`                            | Marquer la facture payee                        | Transition `ISSUED` -> `PAID` autorisee malgre le trigger                            | P0       | `seal_integration_test.go:TestSeal_MarkInvoicePaidStillAllowed` |

<!-- markdownlint-enable MD060 -->

### Backend metier - Suppression de brouillon et lecture du snapshot

<!-- markdownlint-disable MD060 -->

| ID         | Niveau         | Preconditions                                  | Action                          | Resultat attendu                                                          | Priorite | Source                                                       |
| ---------- | -------------- | ---------------------------------------------- | ------------------------------- | ------------------------------------------------------------------------- | -------- | ------------------------------------------------------------ |
| INV-BE-060 | Integration DB | Facture `DRAFT`                                | Supprimer le brouillon          | Suppression OK, la facture n'existe plus                                  | P0       | `delete_draft_integration_test.go:..._RemovesDraft`          |
| INV-BE-061 | Integration DB | Facture `ISSUED` scellee                       | Supprimer le brouillon          | Refuse `InvoiceFinalized`, la facture reste presente (immuable)           | P0       | `delete_draft_integration_test.go:..._RefusesIssued`         |
| INV-BE-062 | Integration DB | Identifiant inconnu                            | Supprimer le brouillon          | `NotFound`                                                               | P1       | `delete_draft_integration_test.go:..._NotFound`              |
| INV-BE-063 | Integration DB | Snapshot B2B avec SIREN/VAT et codes pays      | Lire la facture (`GetInvoice`)  | SIREN, VAT, type client, country id et codes pays emetteur/client exposes | P0       | `party_snapshot_integration_test.go:..._ExposesClientTaxIds` |
| INV-BE-064 | Integration DB | Snapshot legacy/B2C sans identifiants fiscaux  | Lire la facture                 | Identifiants vides (chaine ""), jamais d'erreur de scan                   | P1       | `party_snapshot_integration_test.go:..._EmptyClientTaxIds`   |
| INV-BE-065 | Integration DB | Snapshot avec `oss_applied=true`, client DE    | Lire la facture                 | `oss_applied` expose vrai, code pays client DE et emetteur FR             | P0       | `party_snapshot_integration_test.go:..._OssApplied`          |

<!-- markdownlint-enable MD060 -->

### Gateway - Contrat API HTTP (a couvrir)

Ces lignes ne sont pas encore couvertes par des tests automatises dans le service invoice; elles cadrent le contrat attendu cote gateway (`/api/invoices/*`, gate `authz.ResourceSubscriptionInvoices`).

| ID         | Niveau     | Preconditions                          | Action                                   | Resultat attendu                                       | Priorite | Statut    |
| ---------- | ---------- | -------------------------------------- | ---------------------------------------- | ------------------------------------------------------ | -------- | --------- |
| INV-GW-001 | Controller | Source devis valide                    | `POST /api/invoices`                     | `201` + `success` + `invoice_id`                       | P0       | A couvrir |
| INV-GW-002 | Controller | Payload invalide                       | `POST /api/invoices`                     | `400` + message d'erreur standard                      | P0       | A couvrir |
| INV-GW-003 | Controller | Facture existante                      | `GET /api/invoices/:id`                  | `200` + details (parties, lignes, ventilation, totaux) | P0       | A couvrir |
| INV-GW-004 | Controller | Facture `DRAFT`                        | `POST /api/invoices/:id/issue`           | `200` + numero legal attribue                          | P0       | A couvrir |
| INV-GW-005 | Controller | Facture deja `ISSUED`                  | `POST /api/invoices/:id/issue`           | Idempotent: meme numero, pas de doublon                | P0       | A couvrir |
| INV-GW-006 | Controller | OSS applicable, pays destination sans taux configure | `POST /api/invoices/:id/issue` | `OSSDestinationTaxMissing` (4010), emission bloquee   | P0       | A couvrir |
| INV-GW-007 | Controller | Facture `ISSUED`/`PAID`                | `POST /api/credit-notes`                 | `201` + numero `AV-YYYY-NNNN`                          | P0       | A couvrir |
| INV-GW-008 | Controller | Facture `DRAFT`                        | `DELETE /api/invoices/:id`               | `200`; sur une facture emise -> code metier immuable   | P0       | A couvrir |
| INV-GW-009 | Controller | Facture/avoir emis                     | `GET /api/export/invoices/:id`           | `200` + `application/pdf` (PDF + Factur-X)             | P1       | A couvrir |

### E2E - Parcours metier Cypress (a couvrir)

| ID          | Niveau | Preconditions                          | Action                                   | Resultat attendu                                  | Priorite | Statut    |
| ----------- | ------ | -------------------------------------- | ---------------------------------------- | ------------------------------------------------- | -------- | --------- |
| INV-E2E-001 | E2E    | Devis valide                           | Creer une facture depuis le devis        | Redirection vers la facture en `DRAFT`            | P0       | A couvrir |
| INV-E2E-002 | E2E    | Echeancier `VALID`                     | Facturer des mois selectionnes           | Facture `DRAFT` avec les seuls mois choisis       | P0       | A couvrir |
| INV-E2E-003 | E2E    | Facture `DRAFT`                        | Emettre la facture                       | Numero attribue, document scelle, devient immuable | P0       | A couvrir |
| INV-E2E-004 | E2E    | Facture `ISSUED`                       | Creer un avoir partiel                   | Avoir numerote, lignes creditees grisees          | P0       | A couvrir |
| INV-E2E-005 | E2E    | Vente B2C UE hors FR au-dela du seuil  | Emettre la facture                       | TVA du pays de destination + mention OSS          | P1       | A couvrir |
| INV-E2E-006 | E2E    | Facture ou avoir emis                  | Telecharger le PDF                       | Telechargement OK                                 | P1       | A couvrir |

### PDF / Factur-X - Validation de rendu

Le service couvre deja la validation du XML Factur-X genere contre le XSD officiel (voir `feat/invoice`, commit `1ecefa7`), cote service export.

| ID          | Niveau      | Preconditions                          | Action                | Resultat attendu                                           | Priorite | Statut                     |
| ----------- | ----------- | -------------------------------------- | --------------------- | ---------------------------------------------------------- | -------- | -------------------------- |
| INV-PDF-001 | Integration | Facture standard                       | Generer PDF/Factur-X  | XML CII valide contre le XSD officiel                      | P0       | Couvert (service `export`) |
| INV-PDF-002 | Integration | Avoir                                  | Generer PDF/Factur-X  | XML CII valide contre le XSD officiel                      | P0       | Couvert (service `export`) |
| INV-PDF-003 | Integration | Vente OSS B2C                          | Generer Factur-X      | Pays acheteur et mention OSS coherents dans le XML         | P1       | A couvrir                  |
| INV-PDF-004 | Integration | Mentions legales / franchise TVA       | Generer PDF           | Mentions presentes (art. 293 B si exonere)                 | P1       | A couvrir                  |
| INV-PDF-005 | Integration | Emetteur avec IBAN/BIC                  | Generer Factur-X      | Groupe BG-16 (BT-81=30, IBAN, BIC) valide XSD/Schematron   | P1       | Couvert (service `export`) |

## Ordre TDD recommande

1. backend metier deja couvert: calculs (`INV-BE-001` a `005`), eligibilite (`010` a `016`), numerotation (`020` a `026`), avoirs (`030` a `036`), OSS (`040` a `044`), scellement (`050` a `058`), brouillon/snapshot (`060` a `065`)
2. gateway HTTP: `INV-GW-001` a `INV-GW-009`
3. e2e metier: `INV-E2E-001` a `INV-E2E-006`
4. rendu PDF/Factur-X restant: `INV-PDF-003` et `INV-PDF-004`

## Recommandation d'outillage

### Backend

- logique pure en TDD via les shims `*_export_test_shim.go` (calculs, OSS, selection d'avoir, formatage des numeros)
- tests DB (numerotation concurrente, scellement, snapshot, cumul OSS, suppression de brouillon) gardes par `INVOICE_TEST_DATABASE_URL`; chaque test s'execute dans un schema jetable
- conserver le style `sqlmock` pour les tests d'allocation hors base reelle

### Front

Strategie E2E uniquement via Cypress (toujours `--browser firefox` en local), alignee sur la fonctionnalite echeancier.

## Definition Of Done d'un ticket TDD

Un ticket de la matrice est considere termine seulement si:

1. le test a ete ecrit avant le code metier ou UI correspondant
2. le test a echoue initialement
3. le code est passe au vert
4. le refactor n'a pas casse le test
5. la CI couvre l'execution du test ajoute

## Documents a mettre a jour lors de l'implementation

- `docs/services/invoice.md`
- `docs/adr/0002-oss-seuil-bascule-automatique.md` si la regle OSS evolue
- `docs/contracts/http-gateway.md` et `docs/services/gateway.md` lors de l'ajout des routes `/api/invoices/*`
- `docs/contracts/grpc-services.md` si le contrat gRPC change
- `docs/ERROR_CODES.md` pour tout nouveau code metier
