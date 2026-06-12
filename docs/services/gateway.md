# Service Gateway

## Role

Facade HTTP publique du backend.

- expose l'API `/api/*`
- applique auth middleware
- appelle les services gRPC
- normalise les erreurs

## Point d'entree

- `backend/gateway/main.go`

## Routes

- auth: `controllers.AuthRoutes`
- users: `controllers.UserRoutes`
- quotes: `controllers.QuotesRoutes`
- schedules: `controllers.SchedulesRoutes`
- export: `controllers.ExportRoutes`
- templates: `controllers.TemplateRoutes`

## Dependances

Variables inter-services:

- `AUTH_SERVICE_ADDRESS`
- `USER_SERVICE_ADDRESS`
- `QUOTE_SERVICE_ADDRESS`
- `SCHEDULE_SERVICE_ADDRESS`
- `EXPORT_SERVICE_ADDRESS`
- `TEMPLATE_SERVICE_ADDRESS` (local/dev)

## Ports

| Contexte          |      Port | Direction   | Note                         |
| ----------------- | --------: | ----------- | ---------------------------- |
| Processus gateway |      8080 | ecoute HTTP | `r.Run(":8080")`             |
| Docker local      | 8080:8080 | publie      | API accessible depuis l'hote |
| Docker production | 8080:8080 | publie      | API accessible depuis l'hote |

## Variables d'environnement (vue exhaustive)

### Variables consommees par le code

| Variable                   | Usage                                | Definie local | Definie prod  |
| -------------------------- | ------------------------------------ | ------------- | ------------- |
| `AUTH_SERVICE_ADDRESS`     | client gRPC auth                     | oui           | oui           |
| `USER_SERVICE_ADDRESS`     | client gRPC users                    | oui           | oui           |
| `QUOTE_SERVICE_ADDRESS`    | client gRPC quote                    | oui           | oui           |
| `SCHEDULE_SERVICE_ADDRESS` | client gRPC schedule                 | a ajouter     | a ajouter     |
| `EXPORT_SERVICE_ADDRESS`   | client gRPC export                   | oui           | oui           |
| `TEMPLATE_SERVICE_ADDRESS` | client gRPC template                 | oui           | non           |
| `ENV`                      | cookie `secure` dans auth controller | non (compose) | non (compose) |

### Variables injectees par compose (non lues directement par le code gateway)

| Variable | Definie local | Definie prod | Note             |
| -------- | ------------- | ------------ | ---------------- |
| `TZ`     | oui           | oui          | timezone runtime |

## Middleware auth

- source: `backend/gateway/middleware/auth.go`
- extraction token via Bearer ou cookie
- introspection gRPC via auth (`IntrospectToken`)
- injection contexte (`user_id`, `email`, `role`, `account_status`, `subscription_tier`, `session_version`)

Comportement en cas de session invalidee:

- si auth renvoie `CodeSessionInvalidated (1008)`, le middleware retourne `401` avec code `SESSION_INVALIDATED`
- sinon token invalide/expire: `401` avec code `TOKEN_INVALID`

## Contrat de sortie

- JSON standard avec `success`, `message`, `code`
- cas export: `application/pdf`

## Extension cible pour les echeanciers

Le gateway devra ajouter un controleur dedie aux echeanciers afin de:

- exposer les routes `/api/schedules/*`
- mapper les codes metier du service schedule vers des statuts HTTP coherents
- deleguer l'export PDF d'echeancier au service export

L'integration cible est detaillee dans `docs/services/schedule.md` et `docs/contracts/http-gateway.md`.

## Risques connus

- cookie `secure` conditionne par `ENV=production`
- routes templates exposees alors que la cible n'est pas encore en compose prod
- routes schedules a ajouter en meme temps que `SCHEDULE_SERVICE_ADDRESS` et le deploiement du service dedie
