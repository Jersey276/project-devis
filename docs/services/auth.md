# Service Auth

## Role

Gerer l'authentification et la session:

- inscription
- connexion
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

## Contrats et erreurs

- codes metier auth documentes dans `docs/ERROR_CODES.md`
- mapping HTTP applique dans `backend/gateway/controllers/auth.go`

## Tests

- tests legacy potentiellement instables/hang sur certains scenarios
- recommandation actuelle pour verification rapide:
  - `go test . ./actions ./services`
