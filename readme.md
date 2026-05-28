# project-devis

Portail de documentation technique du projet.

## Index rapide

### Architecture

- [docs/architecture-overview.md](docs/architecture-overview.md)
- [docs/runtime-flows.md](docs/runtime-flows.md)

### Services backend

- [docs/services/gateway.md](docs/services/gateway.md)
- [docs/services/auth.md](docs/services/auth.md)
- [docs/services/users.md](docs/services/users.md)
- [docs/services/quote.md](docs/services/quote.md)
- [docs/services/template.md](docs/services/template.md)
- [docs/services/export.md](docs/services/export.md)

### Contrats

- [docs/contracts/http-gateway.md](docs/contracts/http-gateway.md)
- [docs/contracts/grpc-services.md](docs/contracts/grpc-services.md)
- [docs/ERROR_CODES.md](docs/ERROR_CODES.md)

### Operations

- [docs/DEPLOY.md](docs/DEPLOY.md)
- [docs/operations/env-and-config.md](docs/operations/env-and-config.md)
- [docs/operations/runbook.md](docs/operations/runbook.md)

### Securite

- [docs/security.md](docs/security.md)

## Principes de maintenance documentaire

1. Toute evolution de contrat API (HTTP ou gRPC) implique une mise a jour de `docs/contracts/*`.
2. Toute evolution de configuration runtime implique une mise a jour de `docs/operations/env-and-config.md`.
3. Toute evolution de routage gateway implique une mise a jour de `docs/services/gateway.md`.
4. Toute migration metier impactant l'exploitation doit etre refletee dans `docs/operations/runbook.md`.
