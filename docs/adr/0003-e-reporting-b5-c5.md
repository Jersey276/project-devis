# ADR 0003 - E-reporting transaction (B5) et transfrontalier B2C (C5)

- Statut: Accepte
- Date: 2026-06-22

## Contexte

La reforme francaise de la facturation electronique distingue deux obligations :
la **facturation electronique** (e-invoicing B2B domestique, deja couverte par le
depot B6 sur une Plateforme Agreee) et l'**e-reporting**, c'est-a-dire la
transmission periodique a l'administration (via la PA) des **donnees** des
operations hors champ de la facturation electronique :

- **B5** : donnees de transaction (et de paiement) des operations B2C domestiques ;
- **C5** : ventes a distance B2C intra-UE (transfrontalier), c'est-a-dire le
  perimetre du regime OSS deja gele dans le service (cf. ADR 0002).

Contrairement au depot B6 (une facture a la fois), l'e-reporting transmet un
**agregat par periode**. Il fallait l'implementer sans contracter de PA reelle et
sans changer le comportement de production.

## Decision

1. **Seam neutre commun** `pdp.Reporter` (`SubmitReport` / `FetchReportStatus`),
   3e abstraction a cote de `pdp.Client` (depot) et `pdp.Directory` (annuaire), sans
   proto ni DB. Adaptateur `NoopReporter` par defaut (inerte en prod), `MockReporter`
   pour les tests. **Un seul seam** parametre par un **type de rapport**
   (`TRANSACTION` = B5, `CROSS_BORDER_B2C` = C5) plutot que deux mecanismes separes :
   B5 et C5 ne different que par le perimetre des ventes agregees.

2. **Perimetres disjoints**, identifies sur le **snapshot gele** (inalterabilite) :
   - B5 `TRANSACTION` : `client_type='individual'` ET `client_country_code='FR'` ;
   - C5 `CROSS_BORDER_B2C` : `client_type='individual'` ET `client_country_code<>'FR'`
     ET `counts_toward_oss_threshold` — exactement l'assiette OSS d'ADR 0002.
   Les ventes hors-UE (export) et le B2B etranger sont hors perimetre (extensions
   futures).

3. **Agregat calcule en local** (`actions/reporting.go:reportingAggregate`) : net
   HT/TVA par taux de TVA et pays de destination, somme des factures `ISSUED|PAID`
   du mois civil (Europe/Paris) **moins** les avoirs du mois portant le meme
   perimetre gele. Meme logique nette que le cumul OSS ; aucun appel amont.

4. **Suivi de statut sur la ligne** `invoice_reports` (migration 000015), pas de
   journal d'evenements append-only : un agregat periodique n'a pas l'exigence
   d'inalterabilite d'une facture. Le statut reutilise le vocabulaire du cycle de
   vie B3 (NONE/DEPOSITED/RECEIVED/APPROVED/REJECTED/COLLECTED) pour une vue
   homogene. Contrainte `UNIQUE (user_id, kind, period_year, period_month)` :
   idempotence et `ON CONFLICT` pour resoumettre une periode non terminale ; une
   periode deja `APPROVED`/`COLLECTED` n'est pas resoumise.

5. **Soumission manuelle + worker periodique** : RPC `SubmitInvoiceReport`
   (declenche du front) et worker `PollReportStatuses`/`StartReportPoller` derriere
   `REPORT_POLL_INTERVAL` (vide = desactive), inerte en no-op. La periodicite legale
   n'est pas figee en dur.

6. **Adaptateur Iopole reel differe** : meme sous `PDP_PROVIDER=iopole`, le reporter
   reste no-op. La couverture e-reporting de la spec PPD sera traitee dans une
   iteration ulterieure (smoke test + contractualisation, comme B6).

## Consequences positives

- Reutilisation maximale de l'infrastructure B6/B3 (seam, vocabulaire de statut,
  planif `reconcileSteps`, gating par env) ; surface nouvelle minimale.
- Aucun changement de comportement en production (no-op par defaut).
- Assiette historiquement correcte (figee au snapshot), coherente avec l'OSS.

## Consequences negatives

- Pas encore de transmission reelle (no-op) : la conformite est cablee mais inerte
  tant qu'une PA n'est pas branchee.
- Pas de journal d'evenements sur les rapports : l'historique des transitions n'est
  pas conserve (choix assume ; le statut courant suffit pour un agregat).

## Alternatives ecartees

1. Deux mecanismes separes pour B5 et C5 : duplication inutile (meme agregation,
   meme suivi) ; un parametre `kind` suffit.
2. Exposer seulement des agregats en lecture (sans seam ni soumission) : ne
   "transmet" rien, ne prepare pas la PA.
3. Journal append-only par rapport (calque `invoice_lifecycle_events`) : surface
   supplementaire non justifiee pour une donnee non legale au sens facture.

## References

- `backend/invoice/pdp/reporter.go`, `noop.go`, `mock.go`
- `backend/invoice/actions/reporting.go`, `report.go`, `report_poll.go`
- `backend/invoice/migrations/000015_add_invoice_reports.up.sql`
- `docs/services/invoice.md`, `docs/adr/0002-oss-seuil-bascule-automatique.md`
