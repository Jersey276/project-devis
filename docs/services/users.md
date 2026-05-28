# Service Users

## Role

Fournir les donnees utilisateur et referentiels associes:

- utilisateur courant
- clients
- adresses
- pays et groupes de pays
- taxes

## Point d'entree

- `backend/users/main.go`

## Pattern de demarrage

1. connexion DB
2. migrations embed
3. exposition gRPC sur `:50052`

## Dossiers cles

- `backend/users/actions/`
- `backend/users/services/`
- `backend/users/migrations/`

## Consommateurs

- gateway (`/api/users/*`)
- auth (besoins identite)
- export (assemblage PDF)
- quote (recuperation taxes utilisateur)

## Particularites

- Le gateway applique des validations d'entree (owner_type, URL logo, IDs)
- La route taxes disponible sert a calculer les totaux TTC cote gateway

## Variables

- DB standard (`DB_*`)

## Ports

| Contexte          |       Port | Direction   | Note                           |
| ----------------- | ---------: | ----------- | ------------------------------ |
| Processus users   |      50052 | ecoute gRPC | flag `-port` (defaut 50052)    |
| Docker local      | non publie | interne     | atteint via `devis-user:50052` |
| Docker production | non publie | interne     | atteint via `devis-user:50052` |

## Variables d'environnement (exhaustif)

### Variables declarees dans le service (`services/env.go`)

| Variable              | Usage                    | Definie local | Definie prod |
| --------------------- | ------------------------ | ------------- | ------------ |
| `ENV`                 | convention environnement | non           | non          |
| `POSTGRES_USER`       | compat legacy            | non           | non          |
| `POSTGRES_PASSWORD`   | compat legacy            | non           | non          |
| `POSTGRES_DB`         | compat legacy            | non           | non          |
| `POSTGRES_DB_ADDRESS` | compat legacy            | non           | non          |
| `POSTGRES_DB_PORT`    | compat legacy            | non           | non          |
| `DB_HOST`             | connexion DB             | oui           | oui          |
| `DB_PORT`             | connexion DB             | oui           | oui          |
| `DB_USER`             | connexion DB             | oui           | oui          |
| `DB_PASSWORD`         | fallback secret direct   | non           | non          |
| `DB_PASSWORD_FILE`    | secret DB via fichier    | oui           | oui          |
| `DB_NAME`             | base cible               | oui           | oui          |
