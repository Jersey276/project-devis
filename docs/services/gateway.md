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
