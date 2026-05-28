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
