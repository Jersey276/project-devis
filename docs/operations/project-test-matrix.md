# Matrice De Tests - Projets

Objectif: definir la strategie TDD de la fonctionnalite de gestion de projets et fournir une matrice directement exploitable en tickets de sprint et en criteres de QA. Chaque ligne backend existante reference le test Go correspondant (colonne `Source`).

## Regle TDD

Pour toute evolution de la fonctionnalite projets:

1. le test est ecrit avant le code fonctionnel
2. le test doit echouer initialement (rouge)
3. le code minimal est implemente pour passer au vert
4. le refactor est autorise uniquement apres retour au vert

Aucun endpoint, regle metier ou migration ne doit etre ajoute sans test prealable.

## Perimetre fonctionnel couvert

- service dedie project (gRPC `:50061`)
- creation / lecture / mise a jour / suppression de projets
- rattachement / detachement de devis a un projet (contrainte : un devis appartient a un seul projet)
- liste paginee avec filtres search / status / client_id et tri configurable
- endpoint agrege `GET /api/projects/:id/detail` (fan-out gateway vers quote, schedule, invoice)

## Regles metier a verrouiller par les tests

- `user_id` et `name` sont obligatoires a la creation
- `client_id` est nullable (vide stocke comme NULL)
- statuts autorises : `active`, `completed`, `archived`
- un devis ne peut appartenir qu'a un seul projet : `UNIQUE(quote_id)` sur `project_quotes`
- `AddQuoteToProject` retourne `AlreadyExists` (1002) si le devis est deja dans un projet
- `DeleteProject` s'execute dans une transaction : suppression de `project_quotes` puis de `projects`
- le tri par colonne inconnue bascule sur `p.created_at DESC` (whitelist)
- la verification d'appartenance (ownership) se fait avant toute mutation

## Priorisation

- `P0`: bloque la livraison fonctionnelle ou metier
- `P1`: important, mais peut suivre le socle initial
- `P2`: finition ou robustesse etendue

## Matrice De Tests

### Backend metier - Creation de projet

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                             | Priorite | Source                                                          |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------------ | -------- | --------------------------------------------------------------- |
| PRJ-BE-001  | Unit/sqlmock | user_id et name valides, client_id vide    | `CreateProject`                              | INSERT OK, `project_id` UUID genere, `client_id` NULL        | P0       | `tests/project_test.go:TestCreateProject_Success`               |
| PRJ-BE-002  | Unit/sqlmock | user_id="" et name=""                      | `CreateProject`                              | `CodeInvalidInput`, ValidationErrors sur user_id et name     | P0       | `tests/project_test.go:TestCreateProject_MissingFields`         |
| PRJ-BE-003  | Unit/sqlmock | client_id non vide                         | `CreateProject`                              | INSERT avec client_id passe comme string (non NULL)          | P0       | `tests/project_test.go:TestCreateProject_WithClientID`          |

<!-- markdownlint-enable MD060 -->

### Backend metier - Lecture de projet

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                             | Priorite | Source                                                          |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------------ | -------- | --------------------------------------------------------------- |
| PRJ-BE-010  | Unit/sqlmock | Projet existant, ownership OK              | `GetProject`                                 | `Success: true`, tous les champs scannes correctement        | P0       | `tests/project_test.go:TestGetProject_Success`                  |
| PRJ-BE-011  | Unit/sqlmock | project_id inexistant (ErrNoRows)          | `GetProject`                                 | `Code: 1001 (NotFound)`                                      | P0       | `tests/project_test.go:TestGetProject_NotFound`                 |
| PRJ-BE-012  | Unit/sqlmock | project_id=""                              | `GetProject`                                 | `Code: 1003 (InvalidInput)`, aucune query DB                 | P0       | `tests/project_test.go:TestGetProject_MissingInput`             |

<!-- markdownlint-enable MD060 -->

### Backend metier - Mise a jour de projet

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                             | Priorite | Source                                                          |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------------ | -------- | --------------------------------------------------------------- |
| PRJ-BE-020  | Unit/sqlmock | Projet existant, 1 row affected            | `UpdateProject`                              | `Success: true`                                              | P0       | `tests/project_test.go:TestUpdateProject_Success`               |
| PRJ-BE-021  | Unit/sqlmock | project_id inexistant (0 rows affected)    | `UpdateProject`                              | `Code: 1001 (NotFound)`                                      | P0       | `tests/project_test.go:TestUpdateProject_NotFound`              |
| PRJ-BE-022  | Unit/sqlmock | name=""                                    | `UpdateProject`                              | `Code: 1003 (InvalidInput)`, aucune query DB                 | P0       | `tests/project_test.go:TestUpdateProject_MissingName`           |

<!-- markdownlint-enable MD060 -->

### Backend metier - Suppression de projet

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                             | Priorite | Source                                                          |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------------ | -------- | --------------------------------------------------------------- |
| PRJ-BE-030  | Unit/sqlmock | Projet existant, ownership OK              | `DeleteProject`                              | BEGIN + DELETE project_quotes + DELETE projects + COMMIT     | P0       | `tests/project_test.go:TestDeleteProject_Success`               |
| PRJ-BE-031  | Unit/sqlmock | Projet inexistant (EXISTS = false)         | `DeleteProject`                              | `Code: 1001 (NotFound)`, aucune transaction ouverte          | P0       | `tests/project_test.go:TestDeleteProject_NotFound`              |
| PRJ-BE-032  | Unit/sqlmock | project_id=""                              | `DeleteProject`                              | `Code: 1003 (InvalidInput)`, aucune query DB                 | P0       | `tests/project_test.go:TestDeleteProject_MissingInput`          |
| PRJ-BE-033  | Unit/sqlmock | DB retourne une erreur sur EXISTS          | `DeleteProject`                              | Erreur propagee, `Code: 2001 (InternalError)`                | P0       | `tests/project_test.go:TestDeleteProject_DBError`               |

<!-- markdownlint-enable MD060 -->

### Backend metier - Gestion des devis rattaches

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                             | Priorite | Source                                                          |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------------ | -------- | --------------------------------------------------------------- |
| PRJ-BE-040  | Unit/sqlmock | Projet existant, devis non encore rattache | `AddQuoteToProject`                          | INSERT ON CONFLICT, 1 row affected, `Success: true`          | P0       | `tests/project_quotes_test.go:TestAddQuoteToProject_Success`    |
| PRJ-BE-041  | Unit/sqlmock | Devis deja dans un projet (0 rows)         | `AddQuoteToProject`                          | `Code: 1002 (AlreadyExists)`                                 | P0       | `tests/project_quotes_test.go:TestAddQuoteToProject_AlreadyExists` |
| PRJ-BE-042  | Unit/sqlmock | Projet inexistant (EXISTS = false)         | `AddQuoteToProject`                          | `Code: 1001 (NotFound)`, pas d'INSERT                        | P0       | `tests/project_quotes_test.go:TestAddQuoteToProject_ProjectNotFound` |
| PRJ-BE-043  | Unit/sqlmock | quote_id=""                                | `AddQuoteToProject`                          | `Code: 1003 (InvalidInput)`, aucune query DB                 | P0       | `tests/project_quotes_test.go:TestAddQuoteToProject_MissingInput` |
| PRJ-BE-044  | Unit/sqlmock | Projet existant, devis rattache            | `RemoveQuoteFromProject`                     | DELETE OK, `Success: true`                                   | P0       | `tests/project_quotes_test.go:TestRemoveQuoteFromProject_Success` |
| PRJ-BE-045  | Unit/sqlmock | Projet inexistant                          | `RemoveQuoteFromProject`                     | `Code: 1001 (NotFound)`, pas de DELETE                       | P0       | `tests/project_quotes_test.go:TestRemoveQuoteFromProject_NotFound` |
| PRJ-BE-046  | Unit/sqlmock | 2 devis rattaches                          | `ListProjectQuoteIds`                        | Slice de 2 quote_ids, `Success: true`                        | P0       | `tests/project_quotes_test.go:TestListProjectQuoteIds_Success`  |
| PRJ-BE-047  | Unit/sqlmock | Aucun devis rattache                       | `ListProjectQuoteIds`                        | Slice vide (non nil), `Success: true`                        | P1       | `tests/project_quotes_test.go:TestListProjectQuoteIds_Empty`    |
| PRJ-BE-048  | Unit/sqlmock | project_id=""                              | `ListProjectQuoteIds`                        | `Code: 1003 (InvalidInput)`, aucune query DB                 | P0       | `tests/project_quotes_test.go:TestListProjectQuoteIds_MissingInput` |

<!-- markdownlint-enable MD060 -->

### Backend metier - Liste paginee

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                             | Priorite | Source                                                          |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------------ | -------- | --------------------------------------------------------------- |
| PRJ-BE-050  | Unit/sqlmock | 2 projets, page=1, pageSize=20             | `ListProjects`                               | COUNT puis SELECT ; total=2, 2 projets retournes             | P0       | `tests/list_projects_test.go:TestListProjects_Success`          |
| PRJ-BE-051  | Unit/sqlmock | user_id=""                                 | `ListProjects`                               | `Code: 1003 (InvalidInput)`, aucune query DB                 | P0       | `tests/list_projects_test.go:TestListProjects_MissingUserID`    |
| PRJ-BE-052  | Unit/sqlmock | Filtre search="Matching"                   | `ListProjects`                               | Clause `ILIKE` presente, total=1                             | P0       | `tests/list_projects_test.go:TestListProjects_WithSearch`       |
| PRJ-BE-053  | Unit/sqlmock | sort_by="name", sort_direction="ASC"       | `ListProjects`                               | ORDER BY `p.name ASC`, query s'execute                       | P1       | `tests/list_projects_test.go:TestListProjects_SortByName`       |
| PRJ-BE-054  | Unit/sqlmock | sort_by="unknown_field"                    | `ListProjects`                               | Fallback sur `p.created_at DESC`, pas d'injection SQL        | P0       | `tests/list_projects_test.go:TestListProjects_SortByUnknown`    |
| PRJ-BE-055  | Unit/sqlmock | DB COUNT retourne une erreur               | `ListProjects`                               | Erreur propagee, `Code: 2001 (InternalError)`                | P0       | `tests/list_projects_test.go:TestListProjects_DBError`          |

<!-- markdownlint-enable MD060 -->

### Gateway - Contrat API HTTP (a couvrir)

| ID          | Niveau     | Preconditions                          | Action                                   | Resultat attendu                                              | Priorite | Statut    |
| ----------- | ---------- | -------------------------------------- | ---------------------------------------- | ------------------------------------------------------------- | -------- | --------- |
| PRJ-GW-001  | Controller | Payload valide                         | `POST /api/projects`                     | `201` + `project_id`                                          | P0       | A couvrir |
| PRJ-GW-002  | Controller | Payload invalide (name manquant)       | `POST /api/projects`                     | `400` + message d'erreur                                      | P0       | A couvrir |
| PRJ-GW-003  | Controller | project_id existant                    | `GET /api/projects/:id`                  | `200` + structure complete                                     | P0       | A couvrir |
| PRJ-GW-004  | Controller | project_id inexistant                  | `GET /api/projects/:id`                  | `404`                                                         | P0       | A couvrir |
| PRJ-GW-005  | Controller | Payload valide                         | `PUT /api/projects/:id`                  | `200` + `success`                                             | P0       | A couvrir |
| PRJ-GW-006  | Controller | project_id existant                    | `DELETE /api/projects/:id`               | `200` ; project_quotes + projects supprimes en transaction    | P0       | A couvrir |
| PRJ-GW-007  | Controller | Utilisateur connecte                   | `GET /api/projects`                      | `200` + liste paginee + total                                 | P0       | A couvrir |
| PRJ-GW-008  | Controller | Filtres query params (search, status)  | `GET /api/projects?search=x`             | `200` + resultats filtres                                     | P0       | A couvrir |
| PRJ-GW-009  | Controller | project_id + quote_id valides          | `POST /api/projects/:id/quotes`          | `200` + `success`                                             | P0       | A couvrir |
| PRJ-GW-010  | Controller | Devis deja dans un projet              | `POST /api/projects/:id/quotes`          | `409 Conflict`                                                | P0       | A couvrir |
| PRJ-GW-011  | Controller | Liaison existante                      | `DELETE /api/projects/:id/quotes/:qid`   | `200` + `success`                                             | P0       | A couvrir |
| PRJ-GW-012  | Controller | project_id existant                    | `GET /api/projects/:id/detail`           | `200` + quotes + schedules + invoices groupes + totaux HT     | P0       | A couvrir |

### E2E - Parcours metier Cypress (a couvrir)

| ID           | Niveau | Preconditions                          | Action                                         | Resultat attendu                                           | Priorite | Statut    |
| ------------ | ------ | -------------------------------------- | ---------------------------------------------- | ---------------------------------------------------------- | -------- | --------- |
| PRJ-E2E-001  | E2E    | Utilisateur connecte                   | Creer un projet                                | Redirection vers la page projet                            | P0       | A couvrir |
| PRJ-E2E-002  | E2E    | Projet existant + devis existant       | Rattacher un devis au projet                   | Devis appara?t dans la liste du projet                     | P0       | A couvrir |
| PRJ-E2E-003  | E2E    | Devis deja rattache                    | Rattacher le meme devis a un autre projet      | Message d'erreur `AlreadyExists`                           | P0       | A couvrir |
| PRJ-E2E-004  | E2E    | Projet avec devis                      | Acceder a la vue detail                        | Quotes + schedules + invoices affiches, totaux coherents   | P0       | A couvrir |
| PRJ-E2E-005  | E2E    | Projet existant                        | Mettre a jour le nom                           | Modification visible sans rechargement                     | P1       | A couvrir |
| PRJ-E2E-006  | E2E    | Projet existant                        | Supprimer le projet                            | Projet absent du listing, devis liberes                    | P0       | A couvrir |

## Ordre TDD recommande

1. backend deja couvert : `PRJ-BE-001` a `PRJ-BE-055`
2. gateway HTTP : `PRJ-GW-001` a `PRJ-GW-012`
3. e2e metier : `PRJ-E2E-001` a `PRJ-E2E-006`

## Recommandation d'outillage

### Backend

- conserver `sqlmock` pour tous les tests DB
- `DeleteProject` : utiliser `mock.ExpectBegin()` + `mock.ExpectCommit()` pour valider la transaction
- l'endpoint agrege `GET /api/projects/:id/detail` est logique gateway ; tester le fan-out avec des stubs gRPC

### Front

Strategie E2E uniquement via Cypress (toujours `--browser firefox` en local), alignee sur la strategie etablie pour les autres fonctionnalites.

## Definition Of Done d'un ticket TDD

Un ticket de la matrice est considere termine seulement si:

1. le test a ete ecrit avant le code metier correspondant
2. le test a echoue initialement
3. le code est passe au vert
4. le refactor n'a pas casse le test
5. la CI couvre l'execution du test ajoute

## Documents a mettre a jour lors de l'evolution

- `docs/services/project.md`
- `docs/contracts/http-gateway.md` pour toute nouvelle route `/api/projects/*`
- `docs/ERROR_CODES.md` pour tout nouveau code metier
