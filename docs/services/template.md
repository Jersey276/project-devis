# Service Template

## Role

Gerer des templates reutilisables de devis:

- CRUD template
- archivage/restauration
- gestion des lignes de template

## Archivage

| Champ DB | Comportement |
| --- | --- |
| `archived_at TIMESTAMPTZ` | NULL = actif, timestamp = archive |

- `ArchiveTemplate` : `SET archived_at = now()` — uniquement si non deja archive
- `RestoreTemplate` : `SET archived_at = NULL` — uniquement si archive
- `ListTemplates` : exclut les archives par defaut ; passer `include_archived=true` (param HTTP `?archived=true`) pour les inclure
- L'interface expose une checkbox "Inclure les templates archives" au-dessus des onglets (s'applique aux deux types : `quote_line` et `quote_document`)
- Un template archive affiche un badge "Archivé" ; l'action "Modifier" est masquee, seule "Restaurer" est proposee dans le menu

## Point d'entree

- `backend/template/main.go`

## Pattern de demarrage

1. connexion DB
2. migrations embed
3. exposition gRPC sur `:50055`

## Dossiers cles

- `backend/template/actions/`
- `backend/template/services/`
- `backend/template/migrations/`

## Integration gateway

- routes: `backend/gateway/controllers/templates.go`
- groupe API: `/api/templates`

## Etat de deploiement

- present en local (`backend/docker-compose.yml`)
- non present dans le compose production actuel

## Ports

| Contexte           |        Port | Direction   | Note                               |
| ------------------ | ----------: | ----------- | ---------------------------------- |
| Processus template |       50055 | ecoute gRPC | flag `-port` (defaut 50055)        |
| Docker local       |  non publie | interne     | atteint via `devis-template:50055` |
| Docker production  | non deploye | n/a         | service absent du compose prod     |

## Variables d'environnement (vue exhaustive)

### Variables declarees par le service (`services/env.go`)

| Variable              | Usage                    | Definie local | Definie prod |
| --------------------- | ------------------------ | ------------- | ------------ |
| `ENV`                 | convention environnement | non           | n/a          |
| `POSTGRES_USER`       | compat legacy            | non           | n/a          |
| `POSTGRES_PASSWORD`   | compat legacy            | non           | n/a          |
| `POSTGRES_DB`         | compat legacy            | non           | n/a          |
| `POSTGRES_DB_ADDRESS` | compat legacy            | non           | n/a          |
| `POSTGRES_DB_PORT`    | compat legacy            | non           | n/a          |
| `DB_HOST`             | connexion DB             | oui           | n/a          |
| `DB_PORT`             | connexion DB             | oui           | n/a          |
| `DB_USER`             | connexion DB             | oui           | n/a          |
| `DB_PASSWORD`         | fallback secret direct   | non           | n/a          |
| `DB_PASSWORD_FILE`    | secret DB via fichier    | oui           | n/a          |
| `DB_NAME`             | base cible               | oui           | n/a          |
