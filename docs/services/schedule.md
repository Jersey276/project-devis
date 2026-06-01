# Service Schedule

## Role

Gerer le cycle de vie des echeanciers associes aux devis:

- creation d'un echeancier a partir d'un devis
- initialisation et persistance de la grille mensuelle de facturation
- edition unitaire des cellules avec controle metier
- calcul des ecarts par ligne et au global
- validation d'un unique echeancier par devis
- refus automatique des autres echeanciers du meme devis
- exposition des donnees necessaires a l'export PDF

Ce service est dedie au domaine echeancier afin d'eviter d'alourdir le service devis et de limiter l'impact en cas de panne ou de regression locale.

## Etat

Document de conception et de bootstrap.

Un premier squelette backend TDD est maintenant present dans le depot:

- `backend/schedule/go.mod`
- `backend/schedule/main.go`
- `backend/schedule/actions/`
- `backend/schedule/services/`
- `backend/schedule/migrations/`
- `backend/schedule/tests/`

Le service n'est pas encore branche au gateway ni au compose, et les RPC metier ne sont pas encore implementees. En revanche, le socle suivant existe deja:

- bootstrap gRPC minimal sur le port cible `50056`
- structure `actions.Server` alignee avec les autres services Go
- migrations initiales `schedules` et `schedule_cells`
- premier lot TDD backend sur:
  - validation des inputs de creation
  - validation des montants EUR
  - regles d'editabilite selon le statut

Reference tests: `docs/operations/schedule-test-matrix.md`

## Point d'entree cible

- `backend/schedule/main.go`

## Pattern de demarrage cible

1. lecture des variables d'environnement DB et inter-services
2. connexion DB
3. execution des migrations embed
4. exposition gRPC sur un port dedie
5. enregistrement des handlers RPC dans `actions/`

## Dossiers cles cibles

- `backend/schedule/actions/`
- `backend/schedule/services/`
- `backend/schedule/migrations/`
- `backend/schedule/tests/`

## Frontiere de responsabilite

Le service schedule est responsable de:

- stocker les metadonnees d'un echeancier
- stocker la matrice des montants mensuels par ligne de devis
- appliquer les regles de validation et de verrouillage metier
- calculer les etats de coherence necessaires au front

Le service schedule n'est pas responsable de:

- gerer le cycle de vie complet du devis
- recalculer ou modifier les lignes de devis
- generer directement les PDF

La generation PDF doit rester dans le service export, qui consomme les donnees consolidees exposees par le service schedule.

## Regles metier portees par le service

### Statuts d'echeancier

- `DRAFT`: brouillon editable
- `NEGOCIATE`: statut informatif, editable, utilise lorsqu'un echeancier est presente au client
- `DENIED`: echeancier refuse, conserve en lecture seule, non restaurable
- `VALID`: echeancier valide, verrouille, unique pour un devis donne

### Creation

- un echeancier est toujours rattache a un devis existant
- la creation peut etre lancee depuis le listing devis ou depuis un listing dedie echeanciers
- si la creation est lancee depuis le listing echeanciers, l'utilisateur doit choisir un devis dans la modale
- les cellules sont initialisees a `0.00`
- seules les lignes de devis actives et valides sont integrees dans la grille

### Edition

- une cellule represente le montant affecte a une ligne de devis pour un mois donne
- les montants sont en euro
- les valeurs negatives sont interdites
- les montants sont limites a 2 decimales
- les echeanciers `DENIED` et `VALID` sont non editables
- les echeanciers `DRAFT` et `NEGOCIATE` restent editables

### Validation

- un echeancier est validable uniquement si toutes les lignes de devis actives et valides sont exactement equilibrees
- aucun ecart n'est autorise a la validation
- un seul echeancier peut etre en `VALID` pour un devis donne
- lorsqu'un echeancier passe en `VALID`, tous les autres echeanciers du meme devis passent automatiquement en `DENIED`
- cette transition doit etre atomique pour eviter deux validations concurrentes

### Couleurs et controles de coherence

Pour chaque ligne de devis incluse dans l'echeancier:

- jaune si la somme planifiee est inferieure au montant de la ligne du devis
- vert si la somme planifiee correspond exactement au montant de la ligne du devis
- rouge si la somme planifiee est superieure au montant de la ligne du devis

Pour la ligne de total des colonnes:

- la reference globale est le total du devis
- jaune si le cumul des colonnes est inferieur au total du devis
- vert si le cumul des colonnes correspond exactement au total du devis
- rouge si le cumul des colonnes est superieur au total du devis

## Representation temporelle

- la granularite est mensuelle uniquement
- aucune notion de jour n'est stockee ni exposee
- le mois de debut est saisi au format annee/mois
- les colonnes du tableau representent des mois consecutifs a partir du mois de debut
- le front affiche une ligne d'annee et une ligne de mois en francais

## Modele de donnees cible

### Table `schedules`

Colonnes cibles:

- `schedule_id`
- `quote_id`
- `user_id`
- `name`
- `status`
- `start_month`
- `duration_months`
- `currency`
- `created_at`
- `updated_at`
- `validated_at` nullable

Contraintes cibles:

- `status` borne a `DRAFT`, `NEGOCIATE`, `DENIED`, `VALID`
- `duration_months > 0`
- `currency = EUR`
- au plus un echeancier `VALID` par `quote_id`

### Table `schedule_cells`

Colonnes cibles:

- `schedule_id`
- `quote_line_id`
- `month_index`
- `amount_cents`
- `updated_at`

Contraintes cibles:

- cle unique sur `schedule_id`, `quote_line_id`, `month_index`
- `month_index >= 1`
- `amount_cents >= 0`

### Lecture derivee exposee au front

Le detail d'un echeancier doit permettre de reconstruire:

- les metadonnees de l'echeancier
- les lignes de devis incluses
- les montants par cellule
- le total par ligne
- le total par colonne
- l'etat de coherence par ligne
- l'etat de coherence global

## Dependances

### Dependances runtime

- base Postgres dediee au service schedule
- service quote pour verifier le devis cible et recuperer les lignes eligibles

### Dependances fonctionnelles indirectes

- gateway pour l'exposition HTTP
- export pour la generation du PDF d'echeancier

## Integration gateway cible

Le gateway expose les routes HTTP et mappe les erreurs metier du service schedule vers des statuts HTTP.

Groupe API cible:

- `/api/schedules`

Endpoints HTTP cibles:

- `POST /api/schedules`
- `GET /api/schedules?quote_id=:quoteId`
- `GET /api/schedules/:id`
- `PATCH /api/schedules/:id/cells`
- `POST /api/schedules/:id/validate`
- `GET /api/schedules/:id/export/pdf`

## Contrat gRPC cible

Le service doit suivre le pattern commun du backend:

- `success: bool`
- `code: int32`
- payload metier specifique

RPC attendues a minima:

- `CreateSchedule`
- `ListSchedules`
- `GetSchedule`
- `UpdateScheduleCell`
- `ValidateSchedule`
- `GetScheduleExportPayload` ou RPC equivalente pour fournir les donnees d'export

Etat actuel du squelette:

- un package gRPC minimal est present pour permettre le bootstrap et les tests du module
- le contrat `.proto` complet reste a ecrire avant l'implementation des RPC metier reelles

## Strategie d'export PDF

L'export PDF ne doit pas etre genere par le service schedule.

Strategie cible:

1. le gateway expose `GET /api/schedules/:id/export/pdf`
2. le gateway appelle le service export
3. le service export recupere les donnees d'echeancier via le service schedule
4. le service export assemble les metadonnees complementaires necessaires via quote et users si besoin
5. le PDF est produit en paysage, avec logo et mentions legales

L'export est autorise pour tous les statuts, y compris `DRAFT`, `NEGOCIATE`, `DENIED` et `VALID`.

## Ports

| Contexte           |       Port | Direction   | Note                               |
| ------------------ | ---------: | ----------- | ---------------------------------- |
| Processus schedule |      50056 | ecoute gRPC | port cible propose                 |
| Docker local       | non publie | interne     | atteint via `devis-schedule:50056` |
| Docker production  | non publie | interne     | atteint via `devis-schedule:50056` |

## Variables d'environnement cibles

### Variables declarees par le service

| Variable                | Usage                    | Definie local  | Definie prod   |
| ----------------------- | ------------------------ | -------------- | -------------- |
| `ENV`                   | convention environnement | non            | non            |
| `DB_HOST`               | connexion DB             | a ajouter      | a ajouter      |
| `DB_PORT`               | connexion DB             | a ajouter      | a ajouter      |
| `DB_USER`               | connexion DB             | a ajouter      | a ajouter      |
| `DB_PASSWORD`           | fallback secret direct   | non recommande | non recommande |
| `DB_PASSWORD_FILE`      | secret DB via fichier    | a ajouter      | a ajouter      |
| `DB_NAME`               | base cible               | a ajouter      | a ajouter      |
| `QUOTE_SERVICE_ADDRESS` | client gRPC quote        | a ajouter      | a ajouter      |

### Variables consommees par le gateway

| Variable                   | Usage                | Definie local | Definie prod |
| -------------------------- | -------------------- | ------------- | ------------ |
| `SCHEDULE_SERVICE_ADDRESS` | client gRPC schedule | a ajouter     | a ajouter    |

## Integration compose cible

### Backend local

Ajouts attendus dans `backend/docker-compose.yml`:

- service `devis-schedule`
- variable `SCHEDULE_SERVICE_ADDRESS` dans le gateway
- secret DB partage via `DB_PASSWORD_FILE`

### Initialisation Postgres

Ajout attendu dans `backend/postgres/init.sh`:

- `create_user_and_db "devis-schedule" "schedule"`

### Production

Ajouts attendus dans `docker-compose.prod.yml`:

- service `devis-schedule`
- variable `SCHEDULE_SERVICE_ADDRESS`
- configuration DB du service

## Codes metier cibles

Codes a definir dans le service, puis a mapper dans le gateway:

- ressource introuvable
- entree invalide
- devise/devis introuvable ou inaccessible
- echeancier finalise donc non modifiable
- validation refusee car ecart detecte
- validation refusee car un autre `VALID` existe deja
- erreur interne

## Strategie de test imposee

Le service schedule suit une regle TDD stricte:

1. tests ecrits avant implementation
2. tests backend metier prioritaires avant gateway et front
3. tests de concurrence pour l'unicite du statut `VALID`
4. tests d'export avant branchement final du PDF

Reference detaillee: `docs/operations/schedule-test-matrix.md`

Premier lot deja pose dans `backend/schedule/tests/`:

- `TestNewServer`
- `TestValidateCreateScheduleInput`
- `TestParseAmountEuros`
- `TestIsEditableStatus`

## Risques et points d'attention

- validation concurrente de deux echeanciers du meme devis
- derive entre lignes de devis source et photo attendue dans l'echeancier apres creation
- performance et lisibilite du rendu PDF sur de longues periodes
- volume de cellules si la duree et le nombre de lignes sont eleves
- discipline de non-couplage fort avec le service quote

## Documentation connexe a maintenir

Lors de l'implementation effective, mettre a jour au minimum:

- `docs/architecture-overview.md`
- `docs/contracts/http-gateway.md`
- `docs/contracts/grpc-services.md`
- `docs/services/gateway.md`
- `docs/services/export.md`
- `docs/operations/runbook.md` si impact exploitation
