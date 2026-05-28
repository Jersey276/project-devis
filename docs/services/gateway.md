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
- export: `controllers.ExportRoutes`
- templates: `controllers.TemplateRoutes`

## Dependances

Variables inter-services:

- `AUTH_SERVICE_ADDRESS`
- `USER_SERVICE_ADDRESS`
- `QUOTE_SERVICE_ADDRESS`
- `EXPORT_SERVICE_ADDRESS`
- `TEMPLATE_SERVICE_ADDRESS` (local/dev)

JWT:

- `APP_KEY`

## Ports

| Contexte          |      Port | Direction   | Note                         |
| ----------------- | --------: | ----------- | ---------------------------- |
| Processus gateway |      8080 | ecoute HTTP | `r.Run(":8080")`             |
| Docker local      | 8080:8080 | publie      | API accessible depuis l'hote |
| Docker production | 8080:8080 | publie      | API accessible depuis l'hote |

## Variables d'environnement (exhaustif)

### Variables consommees par le code

| Variable                   | Usage                                | Definie local | Definie prod  |
| -------------------------- | ------------------------------------ | ------------- | ------------- |
| `AUTH_SERVICE_ADDRESS`     | client gRPC auth                     | oui           | oui           |
| `USER_SERVICE_ADDRESS`     | client gRPC users                    | oui           | oui           |
| `QUOTE_SERVICE_ADDRESS`    | client gRPC quote                    | oui           | oui           |
| `EXPORT_SERVICE_ADDRESS`   | client gRPC export                   | oui           | oui           |
| `TEMPLATE_SERVICE_ADDRESS` | client gRPC template                 | oui           | non           |
| `APP_KEY`                  | verification JWT middleware          | non (compose) | non (compose) |
| `ENV`                      | cookie `secure` dans auth controller | non (compose) | non (compose) |

### Variables injectees par compose (non lues directement par le code gateway)

| Variable | Definie local | Definie prod | Note             |
| -------- | ------------- | ------------ | ---------------- |
| `TZ`     | oui           | oui          | timezone runtime |

## Middleware auth

- source: `backend/gateway/middleware/auth.go`
- extraction token via Bearer ou cookie
- validation HS256
- injection contexte (`user_id`, `email`)

## Contrat de sortie

- JSON standard avec `success`, `message`, `code`
- cas export: `application/pdf`

## Risques connus

- cookie `secure` conditionne par `ENV=production`
- routes templates exposees alors que la cible n'est pas encore en compose prod
