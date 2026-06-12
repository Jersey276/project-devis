# Service Auth

## Role

Gerer l'authentification et la session:

- inscription
- connexion
- reset password (demande + confirmation)
- refresh token
- deconnexion
- operations de compte (email/password)

## Point d'entree

- `backend/auth/main.go`

## Pattern de demarrage

1. connexion DB
2. migrations embed
3. initialisation client users gRPC
4. exposition serveur gRPC sur `:50051`

## Dossiers cles

- `backend/auth/actions/`: handlers RPC
- `backend/auth/services/`: DB, env, jwt, refresh, migrate
- `backend/auth/migrations/`: schema auth

## Variables d'environnement importantes

- DB: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD(_FILE)`, `DB_NAME`
- inter-service: `USER_SERVICE_ADDRESS`
- securite: `APP_KEY`

## Ports

| Contexte          |       Port | Direction   | Note                           |
| ----------------- | ---------: | ----------- | ------------------------------ |
| Processus auth    |      50051 | ecoute gRPC | flag `-port` (defaut 50051)    |
| Docker local      | non publie | interne     | atteint via `devis-auth:50051` |
| Docker production | non publie | interne     | atteint via `devis-auth:50051` |

## Variables d'environnement (vue exhaustive)

### Variables declarees par le service (`services/env.go`)

| Variable                  | Usage                    | Definie local | Definie prod    |
| ------------------------- | ------------------------ | ------------- | --------------- |
| `ENV`                     | convention environnement | non           | non             |
| `API_HOST`                | reserve                  | non           | non             |
| `API_PORT`                | reserve                  | non           | non             |
| `APP_KEY`                 | signature/validation JWT | non (compose) | non (compose)   |
| `POSTGRES_USER`           | compat legacy            | non           | non             |
| `POSTGRES_PASSWORD`       | compat legacy            | non           | non             |
| `POSTGRES_DB`             | compat legacy            | non           | non             |
| `POSTGRES_DB_ADDRESS`     | compat legacy            | non           | non             |
| `POSTGRES_DB_PORT`        | compat legacy            | non           | non             |
| `DB_HOST`                 | connexion DB             | oui           | oui             |
| `DB_PORT`                 | connexion DB             | oui           | oui             |
| `DB_USER`                 | connexion DB             | oui           | oui             |
| `DB_PASSWORD`             | fallback secret direct   | non           | non             |
| `DB_PASSWORD_FILE`        | secret DB via fichier    | oui           | oui             |
| `DB_NAME`                 | base cible               | oui           | oui             |
| `USER_SERVICE_ADDRESS`    | client gRPC users        | oui           | oui             |
| `SMTP_HOST`               | hote SMTP                | oui (mailpit) | oui (si active) |
| `SMTP_PORT`               | port SMTP                | oui           | oui             |
| `SMTP_USER`               | auth SMTP login          | non           | oui (si requis) |
| `SMTP_PASSWORD`           | auth SMTP secret         | non           | oui (si requis) |
| `SMTP_FROM`               | expediteur email         | oui           | oui             |
| `RESET_PASSWORD_BASE_URL` | URL front lien reset     | oui           | oui             |

## Contrats et erreurs

- codes metier auth documentes dans `docs/ERROR_CODES.md`
- mapping HTTP applique dans `backend/gateway/controllers/auth.go`

## Session invalidation stricte

Le service auth applique une invalidation stricte des sessions d'access token:

- un champ `session_version` est stocke dans la table `auth`
- la claim `session_version` est incluse dans l'access token JWT
- l'RPC `IntrospectToken` verifie token + etat courant en base
- si la version JWT est obsolete, le service renvoie `CodeSessionInvalidated (1008)`

Actions qui invalident immediatement les sessions actives:

- changement de mot de passe
- confirmation de reset mot de passe

Dans ces cas, `session_version` est incremente et les anciens access tokens deviennent invalides.

## Tests

- tests legacy potentiellement instables/hang sur certains scenarios
- recommandation actuelle pour verification rapide:
  - `go test . ./actions ./services`
