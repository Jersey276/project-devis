# Variables d'Environnement et Configuration

## Objectif

Centraliser les variables techniques utilisees par la plateforme.

## Compose production

Fichier: `docker-compose.prod.yml`

Variables compose (fichier `.env` a la racine):

- `IMAGE_PREFIX`: prefix GHCR des images
- `IMAGE_TAG`: tag image (latest ou SHA)
- `FRONT_PORT`: port publie du front

Variables runtime communes:

- `TZ`

Variables runtime gateway:

- `AUTH_SERVICE_ADDRESS`
- `USER_SERVICE_ADDRESS`
- `QUOTE_SERVICE_ADDRESS`
- `EXPORT_SERVICE_ADDRESS`

Variables runtime auth/users/quote:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_NAME`
- `DB_PASSWORD_FILE`

Variables runtime auth (password reset email):

- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USER`
- `SMTP_PASSWORD`
- `SMTP_FROM`
- `RESET_PASSWORD_BASE_URL`

Variables runtime export:

- `QUOTE_SERVICE_ADDRESS`
- `USER_SERVICE_ADDRESS`
- `GOTENBERG_ADDRESS`

## Compose local backend

Fichier: `backend/docker-compose.yml`

Variables supplementaires notables:

- `TEMPLATE_SERVICE_ADDRESS` dans le gateway
- service `devis-template` present

## Variables sensibles

Secret Docker configure:

- `db_password` (fichier `backend/secrets/postgres_pswd.txt`)

Variables sensibles a gouverner explicitement:

- `APP_KEY` (signature/validation JWT)

## Strategie recommandee

1. Documenter un fichier d'exemple dedie au backend (`backend/.env.example`) pour les variables applicatives.
2. Eviter les fallback implicites en production pour les secrets.
3. Conserver les adresses inter-services dans le compose, pas dans le code.
4. Aligner `ENV`/`APP_ENV` entre front et gateway pour eviter les comportements differents en prod.
