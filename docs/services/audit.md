# Service Audit

## Rôle

Tracer et exposer l'activité HTTP du gateway :

- enregistrement de chaque requête (méthode, URL, durée, statut)
- consultation et filtrage du journal
- statistiques d'activité sur 6 mois glissants
- export CSV envoyé par email
- purge automatique des entrées de plus de 6 mois

## Point d'entrée

- `backend/audit/main.go`

## Pattern de démarrage

1. connexion DB principale
2. connexion DB purge (utilisateur DELETE-only, optionnel)
3. migrations embed
4. démarrage du worker de purge (goroutine, toutes les 24 h)
5. exposition gRPC sur `:50057`

## Dossiers clés

- `backend/audit/actions/` — handlers gRPC + helpers `buildFilters`, `runPurge`
- `backend/audit/services/grpc/audit.proto` — définition des RPC
- `backend/audit/migrations/` — migrations SQL

## Modèle de données

### Table `activity_logs`

| Colonne      | Type      | Notes                                 |
| ------------ | --------- | ------------------------------------- |
| `id`         | BIGSERIAL | clé primaire auto-incrémentée         |
| `user_id`    | TEXT      | nullable (requêtes non authentifiées) |
| `method`     | TEXT      | verbe HTTP                            |
| `url`        | TEXT      | chemin complet                        |
| `duration_ms`| INT       | durée de traitement                   |
| `req_body`   | TEXT      | nullable                              |
| `resp_body`  | TEXT      |                                       |
| `resp_status`| INT       | code HTTP de réponse                  |
| `created_at` | TIMESTAMP | horodatage UTC                        |

## RPCs

| RPC                  | Description                                                                  |
| -------------------- | ---------------------------------------------------------------------------- |
| `LogActivity`        | Insère une entrée ; `user_id` et `req_body` vides deviennent NULL            |
| `GetActivityLog`     | Récupère une entrée par `id` (int64)                                         |
| `ListActivityLogs`   | Liste paginée (page défaut 1, pageSize clampé 1–200 → 50) avec filtres       |
| `GetActivityStats`   | Agrégat `(date, resp_status, count)` sur 6 mois glissants                   |
| `ExportActivityLogs` | Génère un CSV et l'envoie par email via le service email (`:50058`)          |

### Filtres de `ListActivityLogs` / `ExportActivityLogs`

| Champ          | Comportement                                                      |
| -------------- | ----------------------------------------------------------------- |
| `user_id`      | égalité exacte                                                    |
| `url_contains` | `ILIKE '%…%'`                                                     |
| `user_id` + `url_contains` | clause `OR` (l'un ou l'autre)                        |
| `resp_statuses`| `IN (…)`                                                          |
| `date_from`    | `created_at >=`                                                   |
| `date_to`      | `created_at <=`                                                   |

## Codes métier

Définis dans `backend/audit/actions/codes.go`.

| Code | Constante         | Signification                    |
| ---- | ----------------- | -------------------------------- |
| `0`  | `CodeSuccess`     | Succès                           |
| `1`  | `CodeInternalError` | Erreur interne                 |
| `2`  | `CodeInvalidInput`  | Champ requis manquant           |
| `3`  | `CodeNotFound`    | Entrée introuvable               |

## Worker de purge

`StartPurgeWorker(purgeDB)` démarre une goroutine qui exécute `runPurge` toutes les 24 h.

- utilise une connexion DB dédiée avec droits DELETE-only (`PURGE_DB_USER`)
- si `purgeDB` est nil, la purge est silencieusement ignorée (le service reste fonctionnel)
- supprime les lignes dont `created_at < now() - INTERVAL '6 months'`

## Export CSV

`ExportActivityLogs` applique les mêmes filtres que `ListActivityLogs`, génère un CSV en mémoire, puis délègue l'envoi au service email via gRPC. Si le service email est indisponible, le RPC retourne `CodeInternalError` sans crash.

## Ports

| Contexte          |       Port | Direction   | Note                             |
| ----------------- | ---------: | ----------- | -------------------------------- |
| Processus audit   |      50057 | écoute gRPC | défaut hardcodé dans `main.go`   |
| Docker local      | non publié | interne     | atteint via `devis-audit:50057`  |

## Variables d'environnement

| Variable              | Usage                              | Définie local | Définie prod |
| --------------------- | ---------------------------------- | ------------- | ------------ |
| `DB_HOST`             | connexion DB principale            | oui           | oui          |
| `DB_PORT`             | connexion DB principale            | oui           | oui          |
| `DB_USER`             | connexion DB principale            | oui (`devis-audit`) | oui    |
| `DB_PASSWORD_FILE`    | secret DB                          | oui           | oui          |
| `DB_NAME`             | base cible                         | oui (`audit`) | oui          |
| `PURGE_DB_USER`       | utilisateur DELETE-only pour purge | optionnel     | optionnel    |
| `PURGE_DB_PASSWORD_FILE` | secret purge DB                 | optionnel     | optionnel    |
| `EMAIL_SERVICE_ADDRESS` | adresse gRPC du service email    | non (défaut `localhost:50058`) | oui |
