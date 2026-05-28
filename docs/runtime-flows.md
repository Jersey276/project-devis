# Runtime Flows

Ce document detaille les flux d'execution applicatifs les plus importants.

## 1) Authentification et session

```mermaid
sequenceDiagram
  participant UI as Front
  participant GW as Gateway
  participant AU as Auth gRPC

  UI->>GW: POST /api/auth/login
  GW->>AU: Login(email, password, remember_me)
  AU-->>GW: token + refresh_token
  GW-->>UI: Set-Cookie access/refresh + 200

  UI->>GW: GET /api/users/me
  GW->>GW: Middleware AuthRequired (JWT)
  GW-->>UI: 200 user

  UI->>GW: requete protegee avec access expire
  GW-->>UI: 401
  UI->>GW: POST /api/auth/refresh
  GW->>AU: RefreshToken(refresh_token)
  AU-->>GW: nouveaux tokens
  GW-->>UI: Set-Cookie + 200
  UI->>GW: retry requete initiale
  GW-->>UI: 200
```

Notes:

- Le front coalesce les 401 concurrents via une promesse unique de refresh.
- Certaines routes auth sont exclues de la boucle refresh/retry.

## 2) Consultation des devis

```mermaid
sequenceDiagram
  participant UI as Front
  participant GW as Gateway
  participant Q as Quote gRPC
  participant U as Users gRPC

  UI->>GW: GET /api/quotes
  GW->>Q: ListQuotes(user_id)
  GW->>Q: ListUserQuoteLines(user_id)
  GW->>U: ListTaxesForUser(user_id, include_ids)
  GW->>GW: computeQuoteTotals(lines, taxes)
  GW-->>UI: quotes + total_ttc
```

Notes:

- Le gateway calcule les totaux TTC agreges pour la liste.
- Les taxes sont chargees avec `include_ids` pour couvrir les references historiques.

## 3) Export PDF d'un devis

```mermaid
sequenceDiagram
  participant UI as Front
  participant GW as Gateway
  participant EX as Export gRPC
  participant Q as Quote gRPC
  participant U as Users gRPC
  participant GT as Gotenberg

  UI->>GW: GET /api/export/quotes/:id
  GW->>EX: ExportQuote(quote_id, user_id)
  EX->>Q: fetch quote payload
  EX->>U: fetch user/client payload
  EX->>GT: render HTML -> PDF
  EX-->>GW: PDF bytes + filename
  GW-->>UI: application/pdf + Content-Disposition
```

Notes:

- Les tailles max gRPC sont alignees a 8 MiB entre gateway et export.
- Un deny-list Chromium est configure cote Gotenberg (mitigation SSRF).

## 4) Gestion des templates

```mermaid
sequenceDiagram
  participant UI as Front
  participant GW as Gateway
  participant T as Template gRPC

  UI->>GW: CRUD /api/templates
  GW->>T: RPC correspondant
  T-->>GW: success/code/payload
  GW-->>UI: mapping HTTP uniforme
```

Notes:

- Le flux est actif en local.
- Le service template n'est pas encore inclus dans la stack production actuelle.

## 5) Demarrage des services avec migrations

1. Le conteneur du service demarre.
2. Connexion DB.
3. Execution des migrations embed (`//go:embed migrations`).
4. Ecoute gRPC sur port dedie.

Ce pattern est applique par `auth`, `users`, `quote`, `template`.
