# Contrat HTTP Gateway

## Base URL

- Local: `http://localhost:8080`
- Prefix API: `/api`

## Groupes de routes

- `/api/auth/*`
- `/api/users/*`
- `/api/quotes/*`
- `/api/export/*`
- `/api/templates/*`

## Authentification

Routes protegees:

- `users`, `quotes`, `export`, `templates`

Mecanismes acceptes par le middleware:

- Header `Authorization: Bearer <token>`
- Cookie access token

## Format de reponse

Pattern courant:

```json
{
  "success": true,
  "message": "...",
  "code": 1001
}
```

Champs optionnels selon endpoint:

- `field_errors`
- `token`, `refresh_token`
- payload metier (`user`, `quotes`, `template`, etc.)

## Mapping d'erreurs

Le gateway mappe les codes metier gRPC vers des statuts HTTP via des tables par controleur.

References:

- auth: `backend/gateway/controllers/auth.go`
- users: `backend/gateway/controllers/users.go`
- quotes: `backend/gateway/controllers/quotes.go`
- templates: `backend/gateway/controllers/templates.go`
- export: `backend/gateway/controllers/export.go`

## Endpoints auth principaux

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `POST /api/auth/password/reset`
- `POST /api/auth/password/confirm-reset`
- `POST /api/auth/password/update`
- `POST /api/auth/email/verify`

## Endpoints users principaux

- `GET/PUT/DELETE /api/users/me`
- CRUD clients: `/api/users/clients`
- CRUD adresses: `/api/users/addresses`
- CRUD pays/groupes: `/api/users/countries`, `/api/users/country-groups`
- CRUD taxes: `/api/users/taxes`

## Endpoints quotes principaux

- `GET/POST /api/quotes`
- `GET/PUT/DELETE /api/quotes/:id`
- transitions: archive, restore, drop, continue
- lignes: `/api/quotes/:id/lines/*`

## Endpoints templates principaux

- `GET/POST /api/templates`
- `GET/PUT/DELETE /api/templates/:id`
- archive/restore
- lignes: `/api/templates/:id/lines/*`

## Endpoint export principal

- `GET /api/export/quotes/:id`
  - response `application/pdf`
  - header `Content-Disposition` renseigne le nom de fichier

## Notes front

Le front implemente:

- refresh/retry automatique sur 401
- exclusions de refresh pour certaines routes auth

Reference: `front/lib/api.ts`
