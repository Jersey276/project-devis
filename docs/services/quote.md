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
- Les codes metier sont mappees vers HTTP dans le gateway
