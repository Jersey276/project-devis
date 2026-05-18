# Déploiement en production

Ce document décrit le pipeline CI/CD de production et les pré-requis à
satisfaire **une seule fois** sur le serveur cible avant que le pipeline
fonctionne.

## Vue d'ensemble du pipeline

Le workflow [`.github/workflows/prod.yml`](../.github/workflows/prod.yml) se
déclenche sur toute pull request ouverte vers la branche `main`. Il enchaîne :

1. **`front-quality`** — ESLint, `npm audit --audit-level=high`, build Next.js,
   tests Cypress (mocks).
2. **`back-quality`** — `gofmt`, `go vet`, `govulncheck`, `go test` (matrix sur
   les 5 services).
3. **`e2e-integration`** — monte le backend complet via
   `backend/docker-compose.yml`, démarre le front en local et lance Cypress
   contre l'API réelle (vérifie l'interconnexion).
4. **`build-and-push`** — build les 6 images Docker (front + 5 services Go) et
   les push sur GHCR (`ghcr.io/jersey276/project-devis-<name>`) avec les tags
   `:latest` et `:<sha>`.
5. **`deploy`** — SSH sur le serveur de production, checkout du commit, `docker
   compose pull` et `up -d` en utilisant
   [`docker-compose.prod.yml`](../docker-compose.prod.yml).
6. **`auto-merge`** — si le déploiement réussit, `gh pr merge --squash
   --delete-branch` est exécuté automatiquement.

Si **n'importe quelle** étape échoue, le pipeline s'arrête et la PR n'est ni
déployée, ni mergée.

## Secrets GitHub requis

À configurer dans **Settings → Secrets and variables → Actions → Repository
secrets** :

| Secret | Description |
|---|---|
| `SSH_HOST` | IP ou hostname du serveur de production. |
| `SSH_USER` | Utilisateur SSH (doit être dans le groupe `docker`). |
| `SSH_PRIVATE_KEY` | Clé privée SSH (RSA ou ED25519) au format OpenSSH. |
| `DEPLOY_PATH` | Chemin absolu du repo cloné sur le serveur (ex: `/srv/project-devis`). |
| `POSTGRES_PASSWORD` | Mot de passe Postgres utilisé pour le job E2E en CI (sans rapport avec celui du serveur prod). |

`GITHUB_TOKEN` est fourni automatiquement par GitHub Actions ; il est utilisé
pour pousser sur GHCR et pour `gh pr merge`.

## Pré-requis serveur (à exécuter une seule fois manuellement)

### 1. Installer Docker et le plugin compose

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker "$USER"
# Se déconnecter / reconnecter pour appliquer le groupe
```

Vérifier :

```bash
docker compose version
```

### 2. Cloner le repo dans `DEPLOY_PATH`

```bash
sudo mkdir -p /srv/project-devis
sudo chown "$USER":"$USER" /srv/project-devis
git clone https://github.com/Jersey276/project-devis.git /srv/project-devis
cd /srv/project-devis
```

Le pipeline fera ensuite des `git fetch` + `git checkout <sha>` à chaque
déploiement, donc cette étape sert juste à initialiser le repo et avoir
`docker-compose.prod.yml` + `backend/Dockerfile.postgres` disponibles
localement.

### 3. Créer le secret Postgres

```bash
mkdir -p backend/secrets
openssl rand -base64 32 > backend/secrets/postgres_pswd.txt
chmod 600 backend/secrets/postgres_pswd.txt
```

> **À garder précieusement** : ce mot de passe protège les bases de données
> persistées dans le volume `postgres`. Le perdre signifie perdre l'accès aux
> données.

### 3.bis. Configurer `.env` pour le compose

`docker-compose.prod.yml` lit `IMAGE_PREFIX` depuis un fichier `.env` placé à
côté de lui :

```bash
cp .env.example .env
# Éditer .env si le prefix GHCR n'est pas ghcr.io/jersey276/project-devis
```

### 4. Login GHCR pour pouvoir `pull` les images privées

Générer un **Personal Access Token (classic)** sur GitHub avec le scope
`read:packages`, puis :

```bash
echo "ghp_xxxxxxxxxxxxxxxx" | docker login ghcr.io -u "<github-username>" --password-stdin
```

Cela écrit les credentials dans `~/.docker/config.json` et permet à `docker
compose pull` de fonctionner sans réauthentification.

> **Note** : le pipeline ré-exécute `docker login` à chaque déploiement avec
> `GITHUB_TOKEN`, donc cette étape manuelle n'est strictement nécessaire que
> pour les commandes manuelles (rollback, debug).

### 5. Donner les droits sur le dossier au user SSH

```bash
sudo chown -R "$USER":"$USER" /srv/project-devis
```

### 6. (Optionnel mais recommandé) Reverse proxy

Le compose expose les ports `3000` (front) et `8080` (gateway API) en clair sur
toutes les interfaces. Pour un usage en production réelle, placer un reverse
proxy (Caddy, Nginx, Traefik) devant pour gérer TLS et redirection.

Hors scope de ce document.

## Premier déploiement

1. Pousser le code sur une branche feature.
2. Ouvrir une pull request **vers `main`**.
3. Suivre l'exécution dans l'onglet **Actions** du repo.
4. Une fois le job `deploy` au vert, la PR est mergée automatiquement.

## Rollback manuel

Si une version pose problème :

```bash
ssh "$SSH_USER@$SSH_HOST"
cd /srv/project-devis

# Lister les SHAs disponibles dans GHCR (depuis github.com/<user>?tab=packages)
# Ou : docker images | grep project-devis

# Re-tagger localement (ou éditer docker-compose.prod.yml pour pointer sur un sha précis)
SHA=<sha-précédent>
sed -i "s|:latest|:${SHA}|g" docker-compose.prod.yml
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

Pour revenir à `:latest`, restaurer le fichier (`git checkout
docker-compose.prod.yml`) et relancer `pull`/`up`.

## Debug rapide

```bash
ssh "$SSH_USER@$SSH_HOST"
cd /srv/project-devis

docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f --tail=200 devis-gateway
docker compose -f docker-compose.prod.yml logs -f --tail=200 front

# Healthcheck rapide
curl -i http://localhost:3000
curl -i http://localhost:8080
```

## Limitations connues

- **Race condition auto-merge** : si quelqu'un push sur la PR pendant
  l'exécution du pipeline, le `gh pr merge` final échouera (la branche aura
  avancé). Relancer le pipeline depuis l'UI GitHub.
- **`govulncheck` strict** : pas de niveau de tolérance natif. Si une CVE
  apparaît côté Go, le job `back-quality` bloque et il faut soit upgrader la
  dépendance, soit ignorer manuellement la vulnérabilité (édition du
  workflow).
- **Postgres buildé sur le serveur** : l'image Postgres custom n'est pas
  poussée sur GHCR ; elle est buildée localement sur le serveur à partir de
  `backend/Dockerfile.postgres`. C'est volontaire pour éviter de pousser des
  scripts d'init contenant potentiellement des structures sensibles, mais ça
  veut dire que le serveur a besoin d'un clone du repo (pas juste du compose
  file).
