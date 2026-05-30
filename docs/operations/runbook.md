# Runbook Exploitation

Ce runbook couvre les operations recurrentes hors pipeline CI/CD.

## 1) Verifier l'etat de la stack

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f --tail=200 devis-gateway
docker compose -f docker-compose.prod.yml logs -f --tail=200 front
```

## 2) Verifier la connectivite applicative

```bash
curl -i http://localhost:3001
curl -i http://localhost:8080
```

## 3) Ajouter une nouvelle base/service metier

1. Ajouter la creation role/db dans `backend/postgres/init.sh`.

Exemple pour le service echeancier:

```bash
create_user_and_db "devis-schedule" "schedule"
```

1. Executer l'init idempotente:

```bash
cd backend
make db-init
```

1. Ajouter le service dans les fichiers compose concernes:

- `backend/docker-compose.yml`
- `docker-compose.prod.yml`

1. Injecter les variables inter-services necessaires dans le gateway et les services dependants.

Pour le service echeancier, les ajouts cibles sont:

- `SCHEDULE_SERVICE_ADDRESS` dans le gateway
- `QUOTE_SERVICE_ADDRESS` dans le service schedule
- configuration DB dediee du service schedule

## 4) Recuperer une migration dirty

1. Corriger la migration SQL.
1. Nettoyer le flag dirty:

```bash
docker compose exec postgres psql -U devis-auth -d <db> -c \
  "UPDATE schema_migrations SET dirty=false WHERE version=<N>;"
```

1. Rebuild/restart du service:

```bash
cd backend
make rebuild-<service>
```

Exemple pour le service echeancier apres ajout au `Makefile`:

```bash
cd backend
make rebuild-schedule
```

## 5) Rollback image en prod

1. Modifier `IMAGE_TAG` dans `.env`.
2. Re-puller et redemarrer:

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

## 6) Sauvegardes minimales Postgres

Executer periodiquement:

```bash
TS=$(date +%F-%H%M)

docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U devis-auth -d auth > /var/backups/project-devis/auth-${TS}.sql

docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U devis-users -d users > /var/backups/project-devis/users-${TS}.sql

docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U devis-quote -d quote > /var/backups/project-devis/quote-${TS}.sql

docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U devis-schedule -d schedule > /var/backups/project-devis/schedule-${TS}.sql
```

## 7) Check incident rapide auth/session

1. Verifier que les cookies arrivent bien au navigateur.
2. Verifier la presence et coherence de `APP_KEY` sur auth + gateway.
3. Verifier le comportement refresh (`/api/auth/refresh`) dans les logs gateway/auth.

## 8) Check incident rapide echeancier

1. Verifier la presence de `SCHEDULE_SERVICE_ADDRESS` dans le runtime du gateway.
2. Verifier que le service schedule demarre bien et applique ses migrations au boot.
3. Verifier que la base `schedule` et le role `devis-schedule` existent dans Postgres.
4. Verifier que le service quote reste joignable depuis le service schedule.
5. Si l'export PDF echeancier echoue, verifier la chaine gateway -> export -> schedule -> quote.

## 9) Checklist de mise en service du futur service schedule

Avant ouverture fonctionnelle des echeanciers:

1. ajouter `create_user_and_db "devis-schedule" "schedule"` dans `backend/postgres/init.sh`
2. ajouter `devis-schedule` dans `backend/docker-compose.yml`
3. ajouter `devis-schedule` dans `docker-compose.prod.yml`
4. ajouter `SCHEDULE_SERVICE_ADDRESS` au gateway
5. ajouter la configuration DB du service schedule
6. ajouter le service a la chaine de sauvegarde Postgres
7. valider les routes `/api/schedules/*` et l'export PDF d'echeancier apres deploiement
