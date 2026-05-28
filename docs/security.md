# Securite Technique

Ce document synthese les choix de securite applicative et les limites connues.

## Authentification

- JWT HS256 signes cote auth.
- Middleware gateway verifie signature + expiration.
- Contexte utilisateur injecte dans les handlers (`user_id`, `email`).

## Session HTTP

- Cookies HTTP-only utilises pour `access_token` et `refresh_token`.
- SameSite = Lax.
- Duree access token courte (15 min), refresh token longue.

## Services interconnectes

- Communications internes gRPC en clair sur reseau Docker prive.
- Les ports publics exposes sont principalement `3001` (front) et `8080` (gateway) en production.

## Mesures existantes

- Gotenberg configure avec deny-list Chromium pour reduire les risques SSRF.
- Isolation logique des donnees via une base Postgres par service metier.
- Secret DB via Docker secrets (`*_PASSWORD_FILE`).

## Risques connus et suivi

Ces points sont identifies pour correction en fin de phase de developpement:

1. Gouvernance APP_KEY:
   - variable critique non centralisee dans les compose principaux.

2. Cookie secure en production:
   - dependance a `ENV=production` qui peut ne pas etre injecte partout.

3. Service template en prod:
   - routes exposees sans service cible deploye.

4. Contrat d'erreur auth:
   - certaines reponses 400 utilisent des codes texte heterogenes.

## Bonnes pratiques operationnelles

- Forcer TLS en entree (reverse proxy) et HSTS au niveau edge.
- Rotation reguliere des secrets.
- Backups testes periodiquement.
- Journalisation des erreurs 5xx et alerting.
- Mise a jour des dependances Go/Node selon cadence fixe.
