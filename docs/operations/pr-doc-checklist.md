# Checklist PR - Impact Documentation

Objectif: eviter les changements de code non refletés dans la documentation technique.

## A cocher dans chaque PR

- [ ] Le changement modifie un contrat HTTP (`/api/*`) ?
- [ ] Le changement modifie un contrat gRPC (RPC/messages/codes) ?
- [ ] Le changement ajoute/modifie des variables d'environnement ?
- [ ] Le changement ajoute/modifie un port, une dependance runtime ou un service compose ?
- [ ] Le changement modifie un flux runtime (auth, refresh, export, quote, template) ?
- [ ] Le changement modifie les mappings d'erreur ou le format des reponses ?
- [ ] Le changement impacte les operations (deploy, migrations, rollback, backup) ?
- [ ] Le changement impacte la securite (secrets, cookies, auth, SSRF, permissions) ?

## Si au moins une case est cochee

Mettre a jour au minimum les documents correspondants:

- Architecture/flux:
  - `docs/architecture-overview.md`
  - `docs/runtime-flows.md`
- Contrats:
  - `docs/contracts/http-gateway.md`
  - `docs/contracts/grpc-services.md`
  - `docs/ERROR_CODES.md` (si auth)
- Services:
  - `docs/services/*.md` (service(s) impacte(s))
- Operations/Securite:
  - `docs/operations/env-and-config.md`
  - `docs/operations/runbook.md`
  - `docs/security.md`
- ADR (si decision structurante):
  - `docs/adr/*.md`

## Bloc PR recommande

Copier/coller ce bloc dans la description de PR:

```md
### Impact documentation

- [ ] Aucun impact doc
- [ ] Impact doc traite

Documents mis a jour:

- ...
```
