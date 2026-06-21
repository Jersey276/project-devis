# ADR 0002 - Seuil OSS et bascule automatique de TVA

- Statut: Accepte
- Date: 2026-06-21

## Contexte

Le service `invoice` applique le regime OSS (vente a distance B2C intra-UE) en
taxant chaque ligne au taux du pays de destination. Jusqu'ici la bascule reposait
sur un opt-in manuel binaire (`users.oss_enabled`), sans tenir compte du seuil
legal de 10 000 EUR de l'art. 259 D du CGI : au-dela de ce cumul annuel de ventes
a distance B2C intra-UE (hors FR), la TVA de destination devient obligatoire,
independamment de toute option du vendeur.

Il fallait donc declencher la bascule origine -> destination automatiquement, tout
en restant coherent avec la philosophie d'inalterabilite du service (donnees gelees
au snapshot a l'emission).

## Decision

1. `oss_enabled` devient une **option d'anticipation** : l'OSS s'applique des lors
   que `oss_enabled OU cumul_annuel >= 10 000 EUR`. Le seuil prime.
2. **Assiette identifiee par un drapeau gele** `counts_toward_oss_threshold` sur
   `invoice_party_snapshots` (migration 000009), pose a l'emission quand la vente
   est B2C, intra-UE et hors FR. Ce drapeau est plus large qu'`oss_applied` : il
   couvre aussi les ventes encore facturees en TVA origine sous le seuil.
3. **Cumul calcule en local** par somme des `total_ht_cents` des factures
   `ISSUED`/`PAID` de l'annee civile (Europe/Paris) portant ce drapeau. Aucun appel
   aux services amont sur le chemin d'emission.
4. **Bascule a la facture suivante** : le cumul exclut le brouillon courant. La
   facture qui franchit le seuil reste en TVA origine ; les suivantes basculent.
5. **Avoirs deduits de l'assiette** : un avoir qui neutralise une vente dans le
   perimetre OSS (drapeau gele herite de la facture origine) reduit le cumul
   annuel. L'assiette est donc le chiffre d'affaires *net* des ventes a distance
   B2C intra-UE. Le drapeau `counts_toward_oss_threshold` est mirroir sur
   `credit_note_party_snapshots`.

## Consequences positives

- Conformite renforcee : la bascule n'est plus oubliee faute d'opt-in manuel.
- Assiette historiquement correcte (figee au moment de la vente, pas recalculee).
- Cout nul en appels amont (somme SQL locale, index partiel dedie).
- Decision metier isolee dans une fonction pure (`ossApplies`), testable.

## Consequences negatives

- Ecart connu avec la regle stricte annee N-1 (voir Alternatives ecartees).
- La facture pivot peut rester en TVA origine (simplification assumee).

## Alternatives ecartees

1. Recalculer l'appartenance UE a l'agregat (re-fetch des pays) :
   - N appels amont par emission, et dependance a l'appartenance UE *actuelle*
     plutot qu'a celle du moment de la vente.
2. Bascule immediate sur la facture pivot :
   - plus exacte juridiquement mais alourdit la logique d'emission pour un gain
     marginal au regard de la cible produit.
3. Regle annee N-1 (TVA destination sur toute l'annee N si seuil franchi en N-1) :
   - reportee ; extensible via une seconde requete sur l'annee precedente.

## Garde-fous

- Colonne `counts_toward_oss_threshold` portee a l'identique sur
  `invoice_party_snapshots` et `credit_note_party_snapshots` ; les deux legs du
  cumul (factures - avoirs) sont donc figes au moment de la vente.
- Tests : decision pure (`oss_threshold_test.go`) et cumul DB
  (`oss_threshold_integration_test.go`, garde par `INVOICE_TEST_DATABASE_URL`).

## References

- `backend/invoice/actions/oss.go`
- `backend/invoice/actions/resolve.go`
- `backend/invoice/migrations/000009_add_oss_threshold_counts_to_party_snapshots.up.sql`
- `docs/services/invoice.md`
