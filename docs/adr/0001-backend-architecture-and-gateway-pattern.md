# ADR 0001 - Architecture backend et pattern gateway

- Statut: Accepte
- Date: 2026-05-29

## Contexte

Le projet doit livrer rapidement des fonctionnalites metier heterogenes (auth, users, quote, template, export) avec une API HTTP unifiee pour le front.

## Decision

1. Conserver une architecture microservices Go avec frontale HTTP unique (`gateway`).
2. Standardiser les services metier en gRPC interne.
3. Centraliser l'auth HTTP/JWT et le mapping d'erreurs dans le gateway.
4. Conserver Postgres mutualise avec une base logique par service metier.

## Consequences positives

- Evolution independante par domaine metier.
- Contrat front stable via l'API gateway.
- Isolation logique des donnees par service.
- Pattern de demarrage homogene (connect DB + migrations + gRPC).

## Consequences negatives

- Risque de divergence entre contrats gRPC et mapping HTTP.
- Duplication potentielle de codes d'erreurs entre services.
- Besoin de discipline documentaire a chaque changement de contrat.

## Alternatives ecartees

1. Monolithe HTTP unique:
   - plus simple au debut, mais moins modulable par domaine.
2. Exposition directe des services gRPC au front:
   - augmente le couplage front/service et complexifie la securite.

## Garde-fous

- Maintenir `docs/contracts/http-gateway.md` et `docs/contracts/grpc-services.md` a jour.
- Ajouter un controle PR sur l'impact documentaire.
- Traiter les anomalies connues en fin de phase (secrets/ENV/template prod).

## References

- `backend/gateway/main.go`
- `backend/docker-compose.yml`
- `docker-compose.prod.yml`
- `docs/architecture-overview.md`
