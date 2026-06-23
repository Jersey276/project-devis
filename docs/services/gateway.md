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

| Groupe           | Contrôleur                     | Notes                                      |
| ---------------- | ------------------------------ | ------------------------------------------ |
| `/api/auth`      | `controllers.AuthRoutes`       |                                            |
| `/api/users`     | `controllers.UserRoutes`       |                                            |
| `/api/quotes`    | `controllers.QuotesRoutes`     |                                            |
| `/api/schedules` | `controllers.SchedulesRoutes`  |                                            |
| `/api/export`    | `controllers.ExportRoutes`     |                                            |
| `/api/templates` | `controllers.TemplateRoutes`   |                                            |
| `/api/invoices`  | `controllers.InvoicesRoutes`   |                                            |
| `/api/projects`  | `controllers.ProjectsRoutes`   | inclut `GET /:id/detail` (fan-out agrégé)  |
| `/api/logs`      | `controllers.AuditRoutes`      | super-admin uniquement, non audité         |

## Dependances

Variables inter-services:

- `AUTH_SERVICE_ADDRESS`
- `USER_SERVICE_ADDRESS`
- `QUOTE_SERVICE_ADDRESS`
- `SCHEDULE_SERVICE_ADDRESS`
- `EXPORT_SERVICE_ADDRESS`
- `TEMPLATE_SERVICE_ADDRESS` (local/dev)
- `INVOICE_SERVICE_ADDRESS`
- `AUDIT_SERVICE_ADDRESS`
- `PROJECT_SERVICE_ADDRESS`

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
| `SCHEDULE_SERVICE_ADDRESS` | client gRPC schedule                 | oui           | oui           |
| `EXPORT_SERVICE_ADDRESS`   | client gRPC export                   | oui           | oui           |
| `TEMPLATE_SERVICE_ADDRESS` | client gRPC template                 | oui           | non           |
| `INVOICE_SERVICE_ADDRESS`  | client gRPC invoice                  | oui           | oui           |
| `AUDIT_SERVICE_ADDRESS`    | client gRPC audit                    | oui           | oui           |
| `PROJECT_SERVICE_ADDRESS`  | client gRPC project                  | oui           | oui           |
| `ENV`                      | cookie `secure` dans auth controller | non (compose) | non (compose) |

### Variables injectees par compose (non lues directement par le code gateway)

| Variable | Definie local | Definie prod | Note             |
| -------- | ------------- | ------------ | ---------------- |
| `TZ`     | oui           | oui          | timezone runtime |

## Middleware audit

- source: `backend/gateway/middleware/audit.go`
- enregistre chaque requête HTTP (méthode, URL, durée, statuts, corps tronqué à 64 KB) dans le service audit via gRPC
- monté sur le groupe `/api` (tous les groupes métier), **sauf** le groupe `/api/logs` qui est enregistré directement sur `r` pour éviter que les consultations de logs se loggent elles-mêmes
- non-bloquant : envoi asynchrone via channel (taille 512) ; les entrées sont silencieusement abandonnées si le channel est plein

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

## Risques connus

- cookie `secure` conditionne par `ENV=production`
- routes templates exposees alors que la cible n'est pas encore en compose prod
- routes schedules a ajouter en meme temps que `SCHEDULE_SERVICE_ADDRESS` et le deploiement du service dedie
