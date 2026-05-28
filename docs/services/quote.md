# Service Quote

## Role

Gerer le cycle de vie des devis:

- creation/mise a jour/suppression logique
- archivage/restauration
- lignes de devis
- transitions d'etat metier

## Point d'entree

- `backend/quote/main.go`

## Pattern de demarrage

1. connexion DB
2. migrations embed
3. exposition gRPC sur `:50053`

## Dossiers cles

- `backend/quote/actions/`
- `backend/quote/services/`
- `backend/quote/migrations/`

## Integration gateway

- routes: `backend/gateway/controllers/quotes.go`
- le gateway agrege:
  - liste devis
  - lignes utilisateur
  - taxes users
  - calcul de `total_ttc`

## Points d'attention

- Le calcul TTC depend de la qualite des donnees `quantity` (string parsee)
- Les codes metier sont mappes vers HTTP dans le gateway

## Ports

| Contexte          |       Port | Direction   | Note                            |
| ----------------- | ---------: | ----------- | ------------------------------- |
| Processus quote   |      50053 | ecoute gRPC | flag `-port` (defaut 50053)     |
| Docker local      | non publie | interne     | atteint via `devis-quote:50053` |
| Docker production | non publie | interne     | atteint via `devis-quote:50053` |

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
