# Contrat gRPC des Services

Ce document fournit une vue contractuelle de haut niveau. Les details exacts de messages et RPC sont portes par les fichiers generes `*.pb.go` de chaque service.

## Services et ports

- Auth: `:50051`
- Users: `:50052`
- Quote: `:50053`
- Export: `:50054`
- Template: `:50055`
- Schedule: `:50056`
- Audit: `:50057`
- Project: `:50061`

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
- consentements RGPD (`AcceptConsent`, `GetConsentStatus`)

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

## Schedule service

Responsabilites:

- creation d'echeanciers rattaches a un devis
- stockage de la grille mensuelle des montants par ligne de devis
- controle des ecarts ligne par ligne et au global
- validation d'un unique echeancier par devis
- passage automatique des autres echeanciers du devis en `DENIED`

Particularites:

- depend du service quote pour verifier le devis cible et recuperer les lignes eligibles
- expose les donnees consolidees necessaires a l'export PDF d'echeancier
- applique les regles de verrouillage metier selon le statut (`DRAFT`, `NEGOCIATE`, `DENIED`, `VALID`)

## Audit service

Responsabilites:

- traçage de chaque requête HTTP du gateway (`LogActivity`)
- consultation / filtrage paginé du journal
- statistiques d'activité sur 6 mois glissants
- export CSV envoyé par email
- purge automatique des entrées de plus de 6 mois (worker dédié)

Particularites:

- `user_id` et `req_body` sont stockés NULL si vides (requêtes non authentifiées)
- utilise deux connexions DB distinctes : lecture/écriture + DELETE-only pour la purge
- dépend du service email (`:50058`) pour `ExportActivityLogs`

## Project service

Responsabilites:

- création / mise à jour / suppression de projets
- rattachement de devis à un projet (contrainte : un devis = un seul projet via `UNIQUE(quote_id)`)
- liste paginée avec filtres search / status / client

Particularites:

- `DeleteProject` s'exécute dans une transaction (supprime d'abord `project_quotes`)
- `AddQuoteToProject` retourne `AlreadyExists` (1002) si le devis appartient déjà à un autre projet
- `ListProjectQuoteIds` est utilisé par le gateway pour le fan-out vers quote / schedule / invoice

## Export service

Responsabilites:

- assembler les donnees quote + users + schedule selon le document demande
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
