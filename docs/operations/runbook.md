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
2. Executer l'init idempotente:

```bash
cd backend
make db-init
```

## 4) Recuperer une migration dirty

1. Corriger la migration SQL.
2. Nettoyer le flag dirty:

```bash
docker compose exec postgres psql -U devis-auth -d <db> -c \
  "UPDATE schema_migrations SET dirty=false WHERE version=<N>;"
```

3. Rebuild/restart du service:

```bash
cd backend
make rebuild-<service>
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
```

## 7) Check incident rapide auth/session

1. Verifier que les cookies arrivent bien au navigateur.
2. Verifier la presence et coherence de `APP_KEY` sur auth + gateway.
3. Verifier le comportement refresh (`/api/auth/refresh`) dans les logs gateway/auth.
