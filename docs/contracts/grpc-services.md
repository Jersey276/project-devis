# Contrat gRPC des Services

Ce document fournit une vue contractuelle de haut niveau. Les details exacts de messages et RPC sont portes par les fichiers generes `*.pb.go` de chaque service.

## Services et ports

- Auth: `:50051`
- Users: `:50052`
- Quote: `:50053`
- Export: `:50054`
- Template: `:50055`

## Pattern de reponse

Pattern frequent cote gRPC:

- `success: bool`
- `code: int32` (code metier)
- payload specifique

## Auth service

Responsabilites:

- register/login/logout
- refresh token
- verification email / reset password
- introspection access token (`IntrospectToken`) pour controle strict de session

Dependance externe:

- appel users service pour certaines operations identite

Note session:

- le contrat d'introspection permet de detecter les tokens obsoletes via `session_version`
- en cas d'obsolescence, auth renvoie le code metier `1008` (session invalidee)

## Users service

Responsabilites:

- profil utilisateur
- clients/adresses
- pays/groupes
- taxes

## Quote service

Responsabilites:

- gestion des devis
- gestion des lignes
- transitions d'etat (archive, finalisation/continuation)

## Template service

Responsabilites:

- templates de devis
- lignes de template
- archivage/restauration

## Export service

Responsabilites:

- assembler les donnees quote + users
- produire un PDF via Gotenberg

Particularites:

- tailles max gRPC configurees a 8 MiB cote client gateway et serveur export

## Gouvernance des codes

- Les codes metier sont definis par service.
- Le gateway est responsable du mapping vers HTTP.
- Le document `docs/ERROR_CODES.md` reste la source de verite pour l'auth existant.

## Synchronisation schema

Regle de maintenance recommandee:

1. modifier proto et regeneration
2. adapter impl services
3. adapter mapping gateway
4. adapter front si contrat HTTP impacte
5. mettre a jour cette documentation
