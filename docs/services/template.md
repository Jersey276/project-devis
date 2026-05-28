# Service Template

## Role

Gerer des templates reutilisables de devis:

- CRUD template
- archivage/restauration
- gestion des lignes de template

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
