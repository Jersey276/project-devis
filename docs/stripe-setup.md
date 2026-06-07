# Stripe — Configuration et déploiement

Ce guide couvre la mise en place de Stripe pour l'intégration d'abonnement (Phase 2).

## 1. Créer les produits et prix dans le Dashboard Stripe

Sur [dashboard.stripe.com](https://dashboard.stripe.com) → **Catalogue de produits** → créer un produit pour chaque plan payant :

| Produit    | Prix    | Intervalle | Variable à récupérer |
|------------|---------|------------|----------------------|
| Pro        | 9,00 €  | Mensuel    | `price_id` → `price_xxxx` |
| Enterprise | 49,00 € | Mensuel    | `price_id` → `price_xxxx` |

Une fois les `price_id` obtenus, les renseigner en base de données :

```sql
UPDATE plans SET stripe_price_id = 'price_xxx_pro'        WHERE tier = 'pro';
UPDATE plans SET stripe_price_id = 'price_xxx_enterprise' WHERE tier = 'enterprise';
```

> Le plan Free n'a pas de `stripe_price_id` — c'est intentionnel (la colonne reste NULL).

---

## 2. En local (mode test)

### 2.1 Clés API

Depuis Dashboard → **Développeurs → Clés API** (s'assurer d'être en mode **Test**) :

- `pk_test_...` → publishable key (frontend)
- `sk_test_...` → secret key (backend)

Renseigner les variables :

**`backend/docker-compose.yml`** — section `devis-subscription` :
```yaml
STRIPE_SECRET_KEY: sk_test_...
STRIPE_WEBHOOK_SECRET: whsec_...   # généré par Stripe CLI (voir §2.2)
```

**`front/.env.local`** :
```
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...
```

### 2.2 Stripe CLI — webhooks en local

La Stripe CLI simule l'envoi de webhooks vers le gateway local, ce qui est nécessaire pour que les abonnements s'activent après un paiement.

```bash
# Installation (Windows)
winget install Stripe.StripeCLI

# Authentification
stripe login

# Rediriger les webhooks vers le gateway local
stripe listen --forward-to localhost:8080/api/webhooks/stripe
```

La CLI affiche un `whsec_...` à copier dans `STRIPE_WEBHOOK_SECRET` du docker-compose.

### 2.3 Tester avec des événements simulés

```bash
# Simuler un abonnement confirmé
stripe trigger customer.subscription.updated

# Simuler une résiliation
stripe trigger customer.subscription.deleted

# Simuler un échec de paiement
stripe trigger invoice.payment_failed
```

### 2.4 Cartes de test

| Numéro                  | Scénario                  |
|-------------------------|---------------------------|
| `4242 4242 4242 4242`   | Paiement accepté          |
| `4000 0000 0000 9995`   | Paiement refusé           |
| `4000 0025 0000 3155`   | Authentification 3DS requise |

Date d'expiration : n'importe quelle date future. CVV : 3 chiffres quelconques.

---

## 3. En production

### 3.1 Clés API live

Depuis Dashboard → **Développeurs → Clés API** (basculer en mode **Live**) :

- `pk_live_...`
- `sk_live_...`

### 3.2 Enregistrer le webhook

Dashboard → **Développeurs → Webhooks → Ajouter un endpoint** :

- **URL** : `https://ton-domaine.com/api/webhooks/stripe`
- **Événements à écouter** :
  - `customer.subscription.created`
  - `customer.subscription.updated`
  - `customer.subscription.deleted`
  - `invoice.payment_failed`

Copier le `whsec_...` généré après la création.

### 3.3 Variables d'environnement

`docker-compose.prod.yml` (ou système de secrets) — `devis-subscription` :
```yaml
STRIPE_SECRET_KEY: sk_live_...
STRIPE_WEBHOOK_SECRET: whsec_...
```

Conteneur Next.js (build arg ou env) :
```
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...
```

---

## 4. Ordre de mise en service

1. Dashboard Stripe → créer les produits Pro et Enterprise → noter les `price_id`
2. Mettre à jour la colonne `stripe_price_id` dans la table `plans` en base
3. Dashboard Stripe → créer l'endpoint webhook → noter le `whsec_`
4. Renseigner `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY`
5. Redéployer `devis-subscription` et `devis-gateway`
6. Vérifier les logs : `docker compose logs devis-subscription -f`
7. Faire un paiement test (carte `4242 4242 4242 4242`) et confirmer que le tier passe à `pro` dans la table `auth`

---

## 5. Architecture du flux de paiement

```
Frontend (Payment Element)
    │
    │  POST /api/subscriptions/payment-intent { plan_id }
    ▼
Gateway ──────────────────────────────► Subscription service
                                              │
                                              │  Stripe API: Customer + Subscription
                                              │  (payment_behavior: default_incomplete)
                                              │
                                        ◄─── client_secret
    │
    │  stripe.confirmPayment()
    ▼
Stripe
    │
    │  Webhook: customer.subscription.updated
    ▼
Gateway (POST /api/webhooks/stripe, raw body)
    │
    ▼
Subscription service
    │  UPDATE subscriptions SET status='active', stripe_subscription_id=...
    │  gRPC → Auth service: UpdateSubscriptionTier(user_id, 'pro')
    ▼
Auth service
    │  UPDATE auth SET subscription_tier='pro', session_version+1
    ▼
Prochain appel API → 401 → refresh token → nouveau JWT avec subscription_tier='pro'
```
