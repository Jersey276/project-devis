# Service Export

## Role

Produire des documents PDF (devis) a partir des donnees metier.

## Point d'entree

- `backend/export/main.go`

## Pattern de demarrage

1. lecture des variables d'adresse (quote/users/gotenberg)
2. connexion gRPC aux services quote et users
3. initialisation client Gotenberg HTTP
4. exposition gRPC sur `:50054`

## Dependances

- `QUOTE_SERVICE_ADDRESS`
- `USER_SERVICE_ADDRESS`
- `GOTENBERG_ADDRESS`

## Particularites

- taille max message gRPC: 8 MiB (recv/send)
- le gateway applique les memes limites cote client

## Integration gateway

- endpoint HTTP principal: `GET /api/export/quotes/:id`
- type de reponse: `application/pdf`
- header `Content-Disposition` enrichi (filename + filename\*)

## Risques/contraintes

- si payload PDF grossit au-dela de 8 MiB, envisager un passage en streaming gRPC
