# Matrice De Tests - Echeanciers

Objectif: definir la strategie TDD obligatoire pour la fonctionnalite d'echeancier, avec une matrice de tests directement exploitable en tickets de sprint et en criteres de QA.

## Regle TDD

Pour toute evolution de la fonctionnalite echeancier:

1. le test est ecrit avant le code fonctionnel
2. le test doit echouer initialement (rouge)
3. le code minimal est implemente pour passer au vert
4. le refactor est autorise uniquement apres retour au vert

Aucun endpoint, composant UI, migration ou regle metier ne doit etre ajoute sans test prealable.

## Perimetre fonctionnel couvert

- service dedie echeancier
- creation d'echeancier depuis un devis ou depuis un listing dedie
- tableau mensuel editable avec autosave
- controle des ecarts par ligne et au global
- validation d'un unique echeancier par devis
- passage automatique des autres echeanciers en `DENIED`
- export PDF paysage, avec logo et mentions legales

## Regles metier a verrouiller par les tests

- statuts echeancier: `DRAFT`, `NEGOCIATE`, `DENIED`, `VALID`
- `NEGOCIATE` est un statut informatif non finalisant
- un echeancier n'est validable que si toutes les lignes devis actives et valides sont exactement equilibrees
- un seul echeancier peut etre en `VALID` pour un devis
- a la validation d'un echeancier, tous les autres passent automatiquement en `DENIED`
- un echeancier `DENIED` conserve ses cellules en lecture seule
- les montants sont en euro, sans valeur negative, avec 2 decimales maximum
- les colonnes representent des mois/annees, avec mois affiches en francais
- l'export PDF est autorise meme si l'echeancier n'est pas `VALID`

## Priorisation

- `P0`: bloque la livraison fonctionnelle ou metier
- `P1`: important, mais peut suivre le socle initial
- `P2`: finition ou robustesse etendue

## Matrice De Tests

### Backend metier - Service echeancier

| ID         | Niveau       | Preconditions                                 | Action                                        | Resultat attendu                                                                            | Priorite |
| ---------- | ------------ | --------------------------------------------- | --------------------------------------------- | ------------------------------------------------------------------------------------------- | -------- |
| SCH-BE-001 | Unit/Service | Devis existant avec lignes actives et valides | Creer un echeancier                           | Echeancier cree en `DRAFT`, cellules generees a `0.00` pour chaque ligne x mois             | P0       |
| SCH-BE-002 | Unit/Service | Requete creation sans `quote_id`              | Creer                                         | Erreur validation input                                                                     | P0       |
| SCH-BE-003 | Unit/Service | Requete creation avec `duration_months = 0`   | Creer                                         | Erreur validation input                                                                     | P0       |
| SCH-BE-004 | Unit/Service | Requete creation avec `start_month` invalide  | Creer                                         | Erreur validation input                                                                     | P0       |
| SCH-BE-005 | Unit/Service | Echeancier `DRAFT`                            | Modifier une cellule avec montant negatif     | Rejet, aucune persistance                                                                   | P0       |
| SCH-BE-006 | Unit/Service | Echeancier `DRAFT`                            | Modifier une cellule avec plus de 2 decimales | Rejet, aucune persistance                                                                   | P0       |
| SCH-BE-007 | Unit/Service | Echeancier `DRAFT`                            | Modifier une cellule a `12.30`                | Persistance OK                                                                              | P0       |
| SCH-BE-008 | Unit/Service | Echeancier `DENIED`                           | Modifier cellule                              | Rejet, lecture seule                                                                        | P0       |
| SCH-BE-009 | Unit/Service | Echeancier `VALID`                            | Modifier cellule                              | Rejet, lecture seule                                                                        | P0       |
| SCH-BE-010 | Unit/Service | Echeancier avec au moins une ligne en ecart   | Valider                                       | Rejet validation                                                                            | P0       |
| SCH-BE-011 | Unit/Service | Toutes les lignes valides sont equilibrees    | Valider                                       | Passage en `VALID`                                                                          | P0       |
| SCH-BE-012 | Unit/Service | Plusieurs echeanciers pour un meme devis      | Valider un echeancier                         | Les autres passent automatiquement en `DENIED`                                              | P0       |
| SCH-BE-013 | Unit/Service | Echeancier en `NEGOCIATE`                     | Modifier cellule                              | Autorise car statut informatif non final                                                    | P1       |
| SCH-BE-014 | Unit/Service | Detail echeancier charge                      | Calculer les ecarts ligne                     | Couleur ligne: jaune si inferieur, vert si exact, rouge si superieur au montant ligne devis | P0       |
| SCH-BE-015 | Unit/Service | Detail echeancier charge                      | Calculer les totaux colonnes                  | Reference globale = total devis, code couleur coherent                                      | P0       |
| SCH-BE-016 | Unit/Service | Deux validations concurrentes                 | Valider en parallele                          | Un seul `VALID` final, operation atomique                                                   | P0       |
| SCH-BE-017 | Unit/Service | Echeancier `DRAFT`                            | Export PDF                                    | Autorise                                                                                    | P1       |
| SCH-BE-018 | Unit/Service | Echeancier `DENIED`                           | Export PDF                                    | Autorise                                                                                    | P1       |

### Gateway - Contrat API HTTP

| ID         | Niveau     | Preconditions                         | Action                              | Resultat attendu                                                       | Priorite |
| ---------- | ---------- | ------------------------------------- | ----------------------------------- | ---------------------------------------------------------------------- | -------- |
| SCH-GW-001 | Controller | Payload creation valide               | `POST /api/schedules`               | `201` + `success` + `schedule_id`                                      | P0       |
| SCH-GW-002 | Controller | Payload invalide                      | `POST /api/schedules`               | `400` + message d'erreur standard                                      | P0       |
| SCH-GW-003 | Controller | `schedule_id` existant                | `GET /api/schedules/:id`            | `200` + structure complete: header, lignes, colonnes, totaux, couleurs | P0       |
| SCH-GW-004 | Controller | Update cellule invalide               | `PATCH /api/schedules/:id/cells`    | `400`                                                                  | P0       |
| SCH-GW-005 | Controller | Echeancier finalise                   | `PATCH /api/schedules/:id/cells`    | `409` ou code metier equivalent                                        | P0       |
| SCH-GW-006 | Controller | Regles de validation non respectees   | `POST /api/schedules/:id/validate`  | `409` + code metier                                                    | P0       |
| SCH-GW-007 | Controller | Regles de validation respectees       | `POST /api/schedules/:id/validate`  | `200` + `success`                                                      | P0       |
| SCH-GW-008 | Controller | Echeancier dans n'importe quel statut | `GET /api/schedules/:id/export/pdf` | `200` + `application/pdf`                                              | P1       |

### Front unitaire - Composants et comportements

| ID         | Niveau    | Preconditions                              | Action                           | Resultat attendu                             | Priorite |
| ---------- | --------- | ------------------------------------------ | -------------------------------- | -------------------------------------------- | -------- |
| SCH-FE-001 | Component | Modale creation ouverte                    | Saisie incomplete puis submit    | Messages d'erreur sur champs requis          | P0       |
| SCH-FE-002 | Component | Modale creation depuis listing echeanciers | Ouvrir                           | Champ selection devis visible et obligatoire | P0       |
| SCH-FE-003 | Component | Detail echeancier charge                   | Rendu tableau                    | Entete annee + mois en francais              | P0       |
| SCH-FE-004 | Component | Cellule editable                           | Modifier puis blur               | Autosave declenche                           | P0       |
| SCH-FE-005 | Component | Cellule editable                           | Modifier puis appuyer sur Entree | Autosave declenche                           | P0       |
| SCH-FE-006 | Component | Ligne insuffisamment repartie              | Rendu                            | Etat visuel jaune                            | P0       |
| SCH-FE-007 | Component | Ligne exactement repartie                  | Rendu                            | Etat visuel vert                             | P0       |
| SCH-FE-008 | Component | Ligne en surplus                           | Rendu                            | Etat visuel rouge                            | P0       |
| SCH-FE-009 | Component | Ligne totaux colonnes                      | Interaction souris/clavier       | Non editable                                 | P0       |
| SCH-FE-010 | Component | Au moins un ecart existe                   | Affichage actions                | Bouton valider desactive                     | P0       |
| SCH-FE-011 | Component | Echeancier `DENIED` ou `VALID`             | Rendu detail                     | Toutes les cellules sont en lecture seule    | P0       |

### E2E - Parcours metier Cypress

| ID          | Niveau | Preconditions                         | Action                                      | Resultat attendu                                         | Priorite |
| ----------- | ------ | ------------------------------------- | ------------------------------------------- | -------------------------------------------------------- | -------- |
| SCH-E2E-001 | E2E    | Utilisateur connecte + devis existant | Creer un echeancier depuis le listing devis | Redirection vers page echeancier, grille initialisee     | P0       |
| SCH-E2E-002 | E2E    | Listing echeanciers accessible        | Creer via modale                            | Selection devis obligatoire, creation OK                 | P0       |
| SCH-E2E-003 | E2E    | Echeancier `DRAFT`                    | Editer cellules puis blur                   | Sauvegarde persistante                                   | P0       |
| SCH-E2E-004 | E2E    | Echeancier avec ecarts                | Tenter validation                           | Refus + message explicite                                | P0       |
| SCH-E2E-005 | E2E    | Echeancier integralement equilibre    | Valider                                     | Statut `VALID` + autres echeanciers du devis en `DENIED` | P0       |
| SCH-E2E-006 | E2E    | Echeancier `DENIED`                   | Essayer de modifier une cellule             | Modification impossible                                  | P0       |
| SCH-E2E-007 | E2E    | Echeancier non `VALID`                | Export PDF                                  | Telechargement OK                                        | P1       |
| SCH-E2E-008 | E2E    | Echeancier `VALID`                    | Export PDF                                  | Telechargement OK                                        | P1       |

### PDF - Validation de rendu

| ID          | Niveau      | Preconditions                         | Action      | Resultat attendu                | Priorite |
| ----------- | ----------- | ------------------------------------- | ----------- | ------------------------------- | -------- |
| SCH-PDF-001 | Integration | Donnees standard                      | Generer PDF | Orientation paysage             | P1       |
| SCH-PDF-002 | Integration | Branding configure                    | Generer PDF | Logo present                    | P1       |
| SCH-PDF-003 | Integration | Mentions legales configurees          | Generer PDF | Mentions presentes              | P1       |
| SCH-PDF-004 | Integration | Duree longue avec nombreuses colonnes | Generer PDF | Mise en page lisible et paginee | P2       |

## Ordre TDD recommande

1. backend metier: `SCH-BE-001` a `SCH-BE-012`
2. gateway HTTP: `SCH-GW-001` a `SCH-GW-007`
3. front unitaire: `SCH-FE-001` a `SCH-FE-011`
4. e2e metier: `SCH-E2E-001` a `SCH-E2E-006`
5. export PDF et finition: `SCH-PDF-001` a `SCH-PDF-004`, puis `SCH-E2E-007` et `SCH-E2E-008`

## Recommandation d'outillage

### Backend

- conserver le style actuel des tests Go avec `sqlmock`
- separer les tests repository/action des tests gateway

### Front

Le depot dispose deja de Cypress, mais pas encore de framework de tests unitaires front declare.

Recommendation:

- ajouter `vitest`
- ajouter `@testing-library/react`
- ajouter `@testing-library/user-event`
- ajouter `jsdom`

Objectif:

- tester la modale, la grille, l'autosave, les etats visuels et le verrouillage en lecture seule sans attendre l'E2E

## Definition Of Done d'un ticket TDD

Un ticket de la matrice est considere termine seulement si:

1. le test a ete ecrit avant le code metier ou UI correspondant
2. le test a echoue initialement
3. le code est passe au vert
4. le refactor n'a pas casse le test
5. la CI couvre l'execution du test ajoute

## Usage recommande en sprint

Chaque ligne de matrice peut etre convertie en ticket avec le format suivant:

- ID de test
- niveau de test
- preconditions
- action
- resultat attendu
- priorite
- statut initial: `Rouge`

## Documents a mettre a jour lors de l'implementation

Si la fonctionnalite echeancier est engagee, mettre a jour au minimum:

- `docs/architecture-overview.md`
- `docs/contracts/http-gateway.md`
- `docs/contracts/grpc-services.md`
- `docs/services/gateway.md`
- `docs/services/quote.md` si un couplage quote est ajoute
- futur document service echeancier
- `docs/operations/runbook.md` si impact migrations, export ou rollback
