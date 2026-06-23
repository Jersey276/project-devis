# Service Project

## Rôle

Organiser les objets métier par projet :

- création / mise à jour / suppression de projets
- rattachement de devis à un projet (contrainte : un devis = un seul projet)
- agrégation détail projet (devis + échéanciers + factures) dans le gateway

## Point d'entrée

- `backend/project/main.go`

## Pattern de démarrage

1. connexion DB
2. migrations embed
3. exposition gRPC sur `:50061`

## Dossiers clés

- `backend/project/actions/project/` — handlers par opération
- `backend/project/services/grpc/project.proto` — définition des RPC
- `backend/project/migrations/` — migrations SQL

## Modèle de données

### Table `projects`

| Colonne       | Type      | Notes                              |
| ------------- | --------- | ---------------------------------- |
| `project_id`  | TEXT PK   | UUID généré à la création          |
| `user_id`     | TEXT      | propriétaire (indexé)              |
| `name`        | TEXT      | requis                             |
| `client_id`   | TEXT      | nullable                           |
| `status`      | TEXT      | `active` (défaut) / `completed` / `archived` |
| `archived_at` | TIMESTAMP | nullable                           |
| `created_at`  | TIMESTAMP |                                    |
| `updated_at`  | TIMESTAMP |                                    |

### Table `project_quotes`

| Colonne      | Type      | Notes                                       |
| ------------ | --------- | ------------------------------------------- |
| `project_id` | TEXT      | FK logique vers `projects.project_id`       |
| `quote_id`   | TEXT UNIQ | un devis ne peut appartenir qu'à un projet  |
| `added_at`   | TIMESTAMP |                                             |

Clé primaire composite `(project_id, quote_id)`.

## RPCs

| RPC                     | Description                                          |
| ----------------------- | ---------------------------------------------------- |
| `CreateProject`         | Crée un projet, retourne `project_id`                |
| `GetProject`            | Récupère un projet par ID (vérifie ownership)        |
| `ListProjects`          | Liste paginée avec filtres search / status / client  |
| `UpdateProject`         | Met à jour name / client_id / status                 |
| `DeleteProject`         | Supprime project_quotes puis project (transaction)   |
| `AddQuoteToProject`     | Insère dans project_quotes ; retourne AlreadyExists si le devis appartient déjà à un autre projet |
| `RemoveQuoteFromProject`| Supprime la liaison                                  |
| `ListProjectQuoteIds`   | Retourne les quote_ids liés (utilisé par le gateway pour le fan-out) |

## Codes métier

Définis dans `backend/project/actions/codes/codes.go`.

| Code   | Constante       | Signification                                   |
| ------ | --------------- | ----------------------------------------------- |
| `0`    | `Success`       | Succès                                          |
| `1001` | `NotFound`      | Projet introuvable ou n'appartient pas à l'user |
| `1002` | `AlreadyExists` | Le devis est déjà rattaché à un autre projet    |
| `1003` | `InvalidInput`  | Champ requis manquant                           |
| `2001` | `InternalError` | Erreur interne                                  |

## Intégration gateway

- routes : `backend/gateway/controllers/projects.go`
- endpoint agrégé `GET /api/projects/:id/detail` :
  1. fan-out parallèle vers `GetProject` + `ListProjectQuoteIds`
  2. fan-out parallèle vers `ListQuotes(quote_ids)`, `ListSchedules(quote_ids)`, `ListInvoices(quote_ids)`
  3. retourne quotes + schedules + invoices groupés par quote_id, plus `total_ht_cents` et `collected_ht_cents`

## Ports

| Contexte           |       Port | Direction   | Note                               |
| ------------------ | ---------: | ----------- | ---------------------------------- |
| Processus project  |      50061 | écoute gRPC | défaut hardcodé dans `main.go`     |
| Docker local       | non publié | interne     | atteint via `devis-project:50061`  |

## Variables d'environnement

| Variable           | Usage        | Définie local | Définie prod |
| ------------------ | ------------ | ------------- | ------------ |
| `DB_HOST`          | connexion DB | oui           | oui          |
| `DB_PORT`          | connexion DB | oui           | oui          |
| `DB_USER`          | connexion DB | oui (`devis-project`) | oui |
| `DB_PASSWORD_FILE` | secret DB    | oui           | oui          |
| `DB_NAME`          | base cible   | oui (`project`) | oui        |
