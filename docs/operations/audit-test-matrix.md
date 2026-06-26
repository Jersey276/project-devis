# Matrice De Tests - Audit

Objectif: definir la strategie TDD de la fonctionnalite de journalisation d'activite et fournir une matrice directement exploitable en tickets de sprint et en criteres de QA. Chaque ligne backend existante reference le test Go correspondant (colonne `Source`).

## Regle TDD

Pour toute evolution de la fonctionnalite audit:

1. le test est ecrit avant le code fonctionnel
2. le test doit echouer initialement (rouge)
3. le code minimal est implemente pour passer au vert
4. le refactor est autorise uniquement apres retour au vert

Aucun endpoint, regle metier ou modification du schema `activity_logs` ne doit etre ajoute sans test prealable.

## Perimetre fonctionnel couvert

- service dedie audit (gRPC `:50057`)
- enregistrement de chaque requete HTTP du gateway
- consultation et filtrage pagine du journal
- statistiques d'activite sur 6 mois glissants
- export CSV envoye par email (service email `:50058`)
- purge automatique des entrees de plus de 6 mois (worker 24 h)

## Regles metier a verrouiller par les tests

- `user_id` et `req_body` vides sont stockes NULL (requetes non authentifiees)
- le `pageSize` est clampe entre 1 et 200 ; valeur par defaut = 50
- les filtres `user_id` et `url_contains` combines produisent un `OR`, pas un `AND`
- les statistiques portent sur une fenetre glissante de 6 mois, groupees par (jour UTC, resp_status)
- la purge utilise une connexion DB DELETE-only distincte (`PURGE_DB_USER`) ; si absente, la purge est ignoree silencieusement
- l'export CSV ne se produit que si le service email est joignable ; sinon `CodeInternalError`

## Priorisation

- `P0`: bloque la livraison fonctionnelle ou metier
- `P1`: important, mais peut suivre le socle initial
- `P2`: finition ou robustesse etendue

## Matrice De Tests

### Backend metier - Enregistrement d'activite

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                        | Priorite | Source                                                        |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------- | -------- | ------------------------------------------------------------- |
| AUD-BE-001  | Unit/sqlmock | user_id vide, req_body vide                | `LogActivity`                                | INSERT OK ; `user_id` et `req_body` passes comme NULL   | P0       | `tests/log_activity_test.go:TestLogActivity_Success`          |
| AUD-BE-002  | Unit/sqlmock | user_id non vide, req_body non vide        | `LogActivity`                                | INSERT OK ; champs passes comme strings non NULL        | P0       | `tests/log_activity_test.go:TestLogActivity_WithUserID`       |
| AUD-BE-003  | Unit/sqlmock | DB retourne une erreur                     | `LogActivity`                                | `Success: false, Code: 1 (InternalError)`               | P0       | `tests/log_activity_test.go:TestLogActivity_DBError`          |

<!-- markdownlint-enable MD060 -->

### Backend metier - Consultation d'une entree

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                        | Priorite | Source                                                        |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------- | -------- | ------------------------------------------------------------- |
| AUD-BE-010  | Unit/sqlmock | Entree existante (id=1)                    | `GetActivityLog`                             | `Success: true`, champs scan corrects                   | P0       | `tests/get_log_test.go:TestGetActivityLog_Success`            |
| AUD-BE-011  | Unit/sqlmock | id inexistant (ErrNoRows)                  | `GetActivityLog`                             | `Success: false, Code: 3 (NotFound)`                    | P0       | `tests/get_log_test.go:TestGetActivityLog_NotFound`           |
| AUD-BE-012  | Unit/sqlmock | DB retourne une erreur                     | `GetActivityLog`                             | `Success: false, Code: 1 (InternalError)`               | P0       | `tests/get_log_test.go:TestGetActivityLog_DBError`            |

<!-- markdownlint-enable MD060 -->

### Backend metier - Liste paginee et filtres

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                        | Priorite | Source                                                        |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------- | -------- | ------------------------------------------------------------- |
| AUD-BE-020  | Unit/sqlmock | Aucun filtre, page=1, pageSize=50          | `ListActivityLogs`                           | COUNT puis SELECT ; total et logs corrects              | P0       | `tests/list_logs_test.go:TestListActivityLogs_NoFilters`      |
| AUD-BE-021  | Unit/sqlmock | Filtre user_id="user-x"                    | `ListActivityLogs`                           | Clause `WHERE user_id = $N` presente                    | P0       | `tests/list_logs_test.go:TestListActivityLogs_FilterByUserID` |
| AUD-BE-022  | Unit/sqlmock | Filtre resp_statuses=[500,502]             | `ListActivityLogs`                           | Clause `WHERE resp_status IN ($N,$M)` presente          | P0       | `tests/list_logs_test.go:TestListActivityLogs_FilterByStatus` |
| AUD-BE-023  | Unit/sqlmock | pageSize=500                               | `ListActivityLogs`                           | Clampe a 50, query s'execute sans erreur                | P0       | `tests/list_logs_test.go:TestListActivityLogs_PageSizeClamped`|
| AUD-BE-024  | Unit/sqlmock | DB COUNT retourne une erreur               | `ListActivityLogs`                           | `Success: false, Code: 1 (InternalError)`               | P0       | `tests/list_logs_test.go:TestListActivityLogs_DBError`        |
| AUD-BE-025  | Unit (pur)   | Filtre nil                                 | `buildFilters`                               | Clause vide, args nil                                   | P0       | `actions/build_filters_test.go:TestBuildFilters_Nil`          |
| AUD-BE-026  | Unit (pur)   | user_id seul                               | `buildFilters`                               | `WHERE user_id = $1`                                    | P0       | `actions/build_filters_test.go:TestBuildFilters_UserID`       |
| AUD-BE-027  | Unit (pur)   | url_contains seul                          | `buildFilters`                               | `WHERE url ILIKE $1`                                    | P0       | `actions/build_filters_test.go:TestBuildFilters_URLContains`  |
| AUD-BE-028  | Unit (pur)   | user_id + url_contains                     | `buildFilters`                               | Clause `OR` (pas `AND`)                                 | P0       | `actions/build_filters_test.go:TestBuildFilters_UserIDAndURL` |
| AUD-BE-029  | Unit (pur)   | resp_statuses=[500,502]                    | `buildFilters`                               | `WHERE resp_status IN ($1,$2)`, 2 args                  | P0       | `actions/build_filters_test.go:TestBuildFilters_RespStatuses` |
| AUD-BE-030  | Unit (pur)   | date_from + date_to                        | `buildFilters`                               | Clauses `>= $N AND <= $M`                               | P0       | `actions/build_filters_test.go:TestBuildFilters_DateRange`    |
| AUD-BE-031  | Unit (pur)   | Tous les filtres combines                  | `buildFilters`                               | Prefixe `WHERE`, au moins 4 args                        | P0       | `actions/build_filters_test.go:TestBuildFilters_AllFilters`   |

<!-- markdownlint-enable MD060 -->

### Backend metier - Statistiques

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                        | Priorite | Source                                                        |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------- | -------- | ------------------------------------------------------------- |
| AUD-BE-040  | Unit/sqlmock | 2 lignes (200 + 500) pour une meme date    | `GetActivityStats`                           | `Success: true`, 2 entrees `StatusCount` correctes      | P0       | `tests/get_stats_test.go:TestGetActivityStats_Success`        |
| AUD-BE-041  | Unit/sqlmock | Aucune entree dans la fenetre              | `GetActivityStats`                           | `Success: true`, slice vide                             | P1       | `tests/get_stats_test.go:TestGetActivityStats_Empty`          |
| AUD-BE-042  | Unit/sqlmock | DB retourne une erreur                     | `GetActivityStats`                           | `Success: false, Code: 1 (InternalError)`               | P0       | `tests/get_stats_test.go:TestGetActivityStats_DBError`        |

<!-- markdownlint-enable MD060 -->

### Backend metier - Purge

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                        | Priorite | Source                                                        |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------- | -------- | ------------------------------------------------------------- |
| AUD-BE-050  | Unit/sqlmock | 3 entrees eligibles                        | `runPurge`                                   | DELETE OK, retourne 3                                   | P0       | `actions/purge_test.go:TestRunPurge_Success`                  |
| AUD-BE-051  | Unit/sqlmock | Aucune entree eligible                     | `runPurge`                                   | DELETE OK, retourne 0                                   | P1       | `actions/purge_test.go:TestRunPurge_NoRows`                   |
| AUD-BE-052  | Unit/sqlmock | DB retourne une erreur                     | `runPurge`                                   | Erreur propagee                                         | P0       | `actions/purge_test.go:TestRunPurge_DBError`                  |
| AUD-BE-053  | Unit (pur)   | `purgeDB = nil`                            | `StartPurgeWorker`                           | Log informatif, pas de panique, pas de goroutine lancee | P1       | (a couvrir)                                                   |

<!-- markdownlint-enable MD060 -->

### Backend metier - Export CSV (a couvrir)

<!-- markdownlint-disable MD060 -->

| ID          | Niveau       | Preconditions                              | Action                                       | Resultat attendu                                        | Priorite | Statut    |
| ----------- | ------------ | ------------------------------------------ | -------------------------------------------- | ------------------------------------------------------- | -------- | --------- |
| AUD-BE-060  | Unit/sqlmock | recipient_email vide                       | `ExportActivityLogs`                         | `CodeInvalidInput`, aucune query DB ni appel email      | P0       | A couvrir |
| AUD-BE-061  | Unit/mock    | DB OK, service email joignable             | `ExportActivityLogs`                         | CSV genere + email envoye, `Success: true`              | P0       | A couvrir |
| AUD-BE-062  | Unit/mock    | DB OK, service email indisponible          | `ExportActivityLogs`                         | `CodeInternalError`, pas de panique                     | P0       | A couvrir |

<!-- markdownlint-enable MD060 -->

### Gateway - Contrat API HTTP (a couvrir)

| ID          | Niveau     | Preconditions                          | Action                              | Resultat attendu                                  | Priorite | Statut    |
| ----------- | ---------- | -------------------------------------- | ----------------------------------- | ------------------------------------------------- | -------- | --------- |
| AUD-GW-001  | Controller | Role admin                             | `GET /api/audit/logs`               | `200` + liste paginee + total                     | P0       | A couvrir |
| AUD-GW-002  | Controller | Filtres query params                   | `GET /api/audit/logs?user_id=x`     | `200` + resultats filtres                         | P0       | A couvrir |
| AUD-GW-003  | Controller | id existant                            | `GET /api/audit/logs/:id`           | `200` + detail complet                            | P0       | A couvrir |
| AUD-GW-004  | Controller | id inexistant                          | `GET /api/audit/logs/:id`           | `404`                                             | P0       | A couvrir |
| AUD-GW-005  | Controller | Role admin                             | `GET /api/audit/stats`              | `200` + tableau de StatusCount                    | P1       | A couvrir |
| AUD-GW-006  | Controller | recipient_email valide, role admin     | `POST /api/audit/export`            | `200` + `success` (email parti)                   | P1       | A couvrir |
| AUD-GW-007  | Controller | Role non-admin                         | N'importe quelle route audit        | `403`                                             | P0       | A couvrir |

## Ordre TDD recommande

1. backend deja couvert : `AUD-BE-001` a `AUD-BE-052`
2. gateway HTTP : `AUD-GW-001` a `AUD-GW-007`
3. export CSV : `AUD-BE-060` a `AUD-BE-062` (necessite un mock email injectables)
4. purge nil : `AUD-BE-053`

## Recommandation d'outillage

### Backend

- conserver `sqlmock` pour tous les tests DB
- `buildFilters` et `runPurge` sont testes en in-package (`package actions`) pour acceder aux fonctions non exportees
- `ExportActivityLogs` necessite de rendre `emailClient` injectable (interface) pour etre testable sans service email reel

### Front

Strategie E2E uniquement via Cypress (toujours `--browser firefox` en local). Les vues admin du journal ne font pas partie du parcours utilisateur courant ; les couvrir en E2E quand elles seront exposees dans le gateway.

## Definition Of Done d'un ticket TDD

Un ticket de la matrice est considere termine seulement si:

1. le test a ete ecrit avant le code metier correspondant
2. le test a echoue initialement
3. le code est passe au vert
4. le refactor n'a pas casse le test
5. la CI couvre l'execution du test ajoute

## Documents a mettre a jour lors de l'evolution

- `docs/services/audit.md`
- `docs/contracts/http-gateway.md` lors de l'ajout des routes `/api/audit/*`
- `docs/ERROR_CODES.md` pour tout nouveau code metier
