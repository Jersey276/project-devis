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

## Clients

Les clients sont les tiers (B2C ou B2B) pour lesquels un prestataire cree des devis.

### Schema (table `clients`)

| Colonne | Type | Notes |
| --- | --- | --- |
| `client_id` | TEXT UNIQUE | UUID genere a la creation |
| `user_id` | TEXT FK | proprietaire (utilisateur prestataire) |
| `first_name`, `last_name` | TEXT NOT NULL | identite |
| `email`, `phone` | TEXT | optionnels |
| `company`, `siren`, `siret`, `vat` | TEXT | champs B2B uniquement |
| `client_type` | TEXT CHECK | `'individual'` ou `'business'` |
| `archived_at` | TIMESTAMP | soft-delete |

### RPCs

| RPC | Description |
| --- | --- |
| `CreateClient` | Cree un client ; genere `client_id` UUID |
| `GetClient` | Retourne un client non archive |
| `ListClients` | Pagination, recherche full-text, filtre `client_types` |
| `UpdateClient` | Remplacement complet (full-replace) |
| `ArchiveClient` | Soft-delete via `archived_at = NOW()` |

### Notes

- `client_type` vaut `'individual'` par defaut si non fourni (retrocompatibilite).
- `siret` (14 chiffres) est l'identifiant d'etablissement ; distinct de `siren` (9 chiffres, identifiant entreprise).
- `GetClient` exclut les clients archives ; pour les voir, passer `include_archived=true` a `ListClients`.
- L'update est un full-replace : omettre un champ le met a NULL cote serveur.
- L'archivage client est **definitif** (pas de RPC restore) : le client disparait de toutes les listes actives.
- L'interface expose une checkbox "Inclure les clients archives" dans la FilterSidebar du tableau des clients ; les lignes archivees sont affichees avec un badge "Archivé" et une opacite reduite.

### Adresses

- `archived_at TIMESTAMP` present sur la table `addresses` ; l'archivage est definitif (pas de restore expose).
- `ListAddresses` filtre toujours `archived_at IS NULL` — les adresses archivees sont invisibles par design (historique conserve en DB uniquement).

## Notes gateway

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

## Variables d'environnement (vue exhaustive)

### Variables declarees par le service (`services/env.go`)

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
